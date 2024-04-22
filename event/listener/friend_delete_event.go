package listener

import (
	"DiTing-Go/dal/model"
	"DiTing-Go/domain/dto"
	"DiTing-Go/domain/enum"
	"DiTing-Go/global"
	pkgEnum "DiTing-Go/pkg/enum"
	"DiTing-Go/pkg/utils"
	"DiTing-Go/utils/jsonUtils"
	"DiTing-Go/utils/redisCache"
	"context"
	"fmt"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"sort"
)

func init() {
	host := viper.GetString("rocketmq.host")
	// 设置推送消费者
	rocketConsumer, _ := rocketmq.NewPushConsumer(
		//消费组
		consumer.WithGroupName(enum.DeleteFriendTopic),
		// namesrv地址
		consumer.WithNameServer([]string{host}),
	)
	err := rocketConsumer.Subscribe(enum.DeleteFriendTopic, consumer.MessageSelector{}, deleteFriendEvent)
	if err != nil {
		global.Logger.Panicf("subscribe error: %s", err.Error())
	}
	err = rocketConsumer.Start()
	if err != nil {
		global.Logger.Panicf("start consumer error: %s", err.Error())
	}
}

func deleteFriendEvent(ctx context.Context, ext ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
	for i := range ext {
		// 解码
		deleteFriendDto := dto.DeleteFriendDto{}
		if err := jsonUtils.UnmarshalMsg(&deleteFriendDto, ext[i]); err != nil {
			global.Logger.Errorf("jsonUtils unmarshal error: %s", err.Error())
			return consumer.ConsumeRetryLater, nil
		}

		if err := deleteFriend(deleteFriendDto); err != nil {
			global.Logger.Errorf("deleteFriend error: %s", err.Error())
			return consumer.ConsumeRetryLater, nil
		}
	}
	return consumer.ConsumeSuccess, nil
}

func deleteFriend(deleteFriendDto dto.DeleteFriendDto) error {

	uid := deleteFriendDto.Uid
	deleteFriendUid := deleteFriendDto.FriendUid

	// 加锁
	uids := utils.Int64Slice{uid, deleteFriendUid}
	sort.Sort(uids)
	key := fmt.Sprintf(enum.UserAndFriendLock, uids[0], uids[1])
	mutex, err := utils.GetLock(key)
	if err != nil {
		return err
	}
	defer utils.ReleaseLock(mutex)

	ctx := context.Background()
	tx := global.Query.Begin()
	userApply := global.Query.UserApply
	userApplyTx := tx.UserApply.WithContext(ctx)

	// 删除好友申请
	if _, err := userApplyTx.Where(userApply.UID.Eq(uid), userApply.TargetID.Eq(deleteFriendUid)).Delete(); err != nil {
		if err := tx.Rollback(); err != nil {
			global.Logger.Errorf("事务回滚失败 %s", err.Error())
		}
		global.Logger.Errorf("删除好友失败 %s", err.Error())
		return errors.New("Business Error")
	}
	// 删除redis缓存
	defer redisCache.RemoveUserApply(uid, deleteFriendUid)

	if _, err := userApplyTx.Where(userApply.UID.Eq(deleteFriendUid), userApply.TargetID.Eq(uid)).Delete(); err != nil {
		if err := tx.Rollback(); err != nil {
			global.Logger.Errorf("事务回滚失败 %s", err.Error())
		}
		global.Logger.Errorf("删除好友失败 %s", err.Error())
		return errors.New("Business Error")
	}
	// 删除redis缓存
	defer redisCache.RemoveUserApply(deleteFriendUid, uid)

	// 软删除好友房间
	roomFriend := global.Query.RoomFriend
	roomFriendTx := tx.RoomFriend.WithContext(ctx)
	uids = utils.Int64Slice{uid, deleteFriendUid}
	sort.Sort(uids)
	fun := func() (interface{}, error) {
		return roomFriendTx.Where(roomFriend.Uid1.Eq(uids[0]), roomFriend.Uid2.Eq(uids[1])).First()
	}
	roomFriendR := model.RoomFriend{}
	key = fmt.Sprintf(enum.RoomFriendCacheByUidAndFriendUid, uids[0], uids[1])
	if err := utils.GetData(key, &roomFriendR, fun); err != nil {
		if err := tx.Rollback(); err != nil {
			global.Logger.Errorf("事务回滚失败 %s", err.Error())
		}
		global.Logger.Errorf("查询好友房间失败 %s", err.Error())
		return errors.New("Business Error")
	}

	if _, err := roomFriendTx.Where(roomFriend.ID.Eq(roomFriendR.ID)).Update(roomFriend.DeleteStatus, pkgEnum.DELETED); err != nil {
		if err := tx.Rollback(); err != nil {
			global.Logger.Errorf("事务回滚失败 %s", err.Error())
		}
		global.Logger.Errorf("删除好友房间失败 %s", err.Error())
		return errors.New("Business Error")
	}
	// 删除redis缓存
	defer redisCache.RemoveRoomFriend(roomFriendR)

	// 软删除房间表
	room := global.Query.Room
	roomTx := tx.Room.WithContext(ctx)
	if _, err := roomTx.Where(room.ID.Eq(roomFriendR.RoomID)).Update(room.DeleteStatus, pkgEnum.DELETED); err != nil {
		if err := tx.Rollback(); err != nil {
			global.Logger.Errorf("事务回滚失败 %s", err.Error())
		}
		global.Logger.Errorf("删除房间失败 %s", err.Error())
		return errors.New("Business Error")
	}
	// 删除redis缓存
	roomR := model.Room{
		ID: roomFriendR.RoomID,
	}
	defer redisCache.RemoveRoomCache(roomR)

	// 删除消息表
	msg := global.Query.Message
	msgTx := tx.Message.WithContext(ctx)
	if _, err := msgTx.Where(msg.RoomID.Eq(roomFriendR.RoomID)).Update(msg.DeleteStatus, pkgEnum.DELETED); err != nil {
		if err := tx.Rollback(); err != nil {
			global.Logger.Errorf("事务回滚失败 %s", err.Error())
		}
		global.Logger.Errorf("删除消息失败 %s", err.Error())
		return errors.New("Business Error")
	}
	// TODO: 删除消息表缓存

	if err := tx.Commit(); err != nil {
		global.Logger.Errorf("事务提交失败 %s", err.Error())
		return errors.New("Business Error")
	}

	return nil
}
