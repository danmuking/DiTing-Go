package listener

import (
	"DiTing-Go/dal/model"
	"DiTing-Go/domain/enum"
	"DiTing-Go/global"
	pkgEnum "DiTing-Go/pkg/enum"
	"DiTing-Go/pkg/utils"
	"DiTing-Go/service"
	"DiTing-Go/utils/redisCache"
	"context"
	"fmt"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/goccy/go-json"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"gorm.io/gorm"
	"sort"
	"strconv"
	"time"
)

func init() {
	host := viper.GetString("rocketmq.host")
	// 设置推送消费者
	rocketConsumer, _ := rocketmq.NewPushConsumer(
		//消费组
		consumer.WithGroupName(enum.UserLoginTopic),
		// namesrv地址
		consumer.WithNameServer([]string{host}),
	)
	err := rocketConsumer.Subscribe(enum.NewFriendTopic, consumer.MessageSelector{}, friendNewEvent)
	if err != nil {
		global.Logger.Panicf("subscribe error: %s", err.Error())
	}
	err = rocketConsumer.Start()
	if err != nil {
		global.Logger.Panicf("start consumer error: %s", err.Error())
	}
}

func friendNewEvent(ctx context.Context, ext ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
	for i := range ext {
		// 解码
		userFriend := model.UserFriend{}
		userFriendMsgByte := ext[i].Message.Body
		err := json.Unmarshal(userFriendMsgByte, &userFriend)
		if err != nil {
			global.Logger.Errorf("json unmarshal error: %s", err.Error())
			return consumer.ConsumeRetryLater, nil
		}
		err = friendNew(userFriend)
		if err != nil {
			global.Logger.Errorf("friendNew error: %s", err.Error())
			return consumer.ConsumeRetryLater, nil
		}

	}
	return consumer.ConsumeSuccess, nil
}

func friendNew(userFriend model.UserFriend) error {
	ctx := context.Background()

	uid := userFriend.UID
	friendUid := userFriend.FriendUID
	uids := utils.Int64Slice{uid, friendUid}
	sort.Sort(uids)
	key := fmt.Sprintf(enum.UserAndFriendLock, uids[0], uids[1])
	mutex, err := utils.GetLock(key)
	if err != nil {
		return err
	}
	defer utils.ReleaseLock(mutex)

	q := global.Query
	tx := q.Begin()
	roomQ := tx.WithContext(ctx).Room
	roomFriendQ := tx.WithContext(ctx).RoomFriend
	contactQ := tx.WithContext(ctx).Contact

	// 创建房间表
	room := model.Room{
		Type:    enum.PERSONAL,
		HotFlag: enum.NORMAL,
		ExtJSON: "{}",
	}
	if err := roomQ.Create(&room); err != nil {
		if err := tx.Rollback(); err != nil {
			global.Logger.Errorf("事务回滚失败 %s", err.Error())
			return err
		}
		global.Logger.Errorf("创建房间失败 %s", err.Error())
		return err
	}

	// 排序，uid小的在前
	uids = utils.Int64Slice{userFriend.UID, userFriend.FriendUID}
	sort.Sort(uids)

	//检查是否有软删除状态的记录
	roomFriend := global.Query.RoomFriend
	fun := func() (interface{}, error) {
		return roomFriendQ.Where(roomFriend.Uid1.Eq(uids[0]), roomFriend.Uid2.Eq(uids[1]), roomFriend.DeleteStatus.Eq(pkgEnum.DELETED)).First()
	}
	roomFriedR := model.RoomFriend{}
	key = fmt.Sprintf(enum.RoomFriendCacheByUidAndFriendUid, uids[0], uids[1])
	err = utils.GetData(key, &roomFriedR, fun)
	// err
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		global.Logger.Errorf("查询数据失败: %v", err)
		if err := tx.Rollback(); err != nil {
			global.Logger.Errorf("事务回滚失败 %s", err.Error())
			return err
		}
		return err
	}
	// 查到了
	if err == nil {
		roomFriedR.RoomID = room.ID
		roomFriedR.DeleteStatus = pkgEnum.NORMAL
		if _, err := roomFriendQ.Select(roomFriend.RoomID, roomFriend.DeleteStatus).Where(roomFriend.ID.Eq(roomFriedR.ID)).Updates(roomFriedR); err != nil {
			if err := tx.Rollback(); err != nil {
				global.Logger.Errorf("事务回滚失败 %s", err.Error())
				return err
			}
			global.Logger.Errorf("更新房间失败 %s", err.Error())
			return err
		}
		roomFriendR := model.RoomFriend{
			Uid1: uids[0],
			Uid2: uids[1],
		}
		defer redisCache.RemoveRoomFriend(roomFriendR)
	} else {
		// 创建私聊表
		newRoomFriend := model.RoomFriend{
			RoomID:  room.ID,
			Uid1:    uids[0],
			Uid2:    uids[1],
			RoomKey: strconv.FormatInt(uids[0], 10) + "," + strconv.FormatInt(uids[1], 10),
		}
		if err := roomFriendQ.Create(&newRoomFriend); err != nil {
			if err := tx.Rollback(); err != nil {
				global.Logger.Errorf("事务回滚失败 %s", err.Error())
				return err
			}
			global.Logger.Errorf("创建房间失败 %s", err.Error())
			return err
		}
	}

	// 自动发送一条消息
	newMsg := model.Message{
		RoomID:       room.ID,
		FromUID:      userFriend.UID,
		Content:      "你们已经是好友了，开始聊天吧",
		DeleteStatus: pkgEnum.NORMAL,
		Type:         enum.TextMessage,
		Extra:        "{}",
	}
	if err := service.SendTextMsg(&newMsg); err != nil {
		if err := tx.Rollback(); err != nil {
			global.Logger.Errorf("事务回滚失败 %s", err.Error())
			return err
		}
		global.Logger.Errorf("发送消息失败 %s", err.Error())
		return err
	}

	//创建会话表
	s, _ := time.ParseDuration("-1s")
	if err := contactQ.Create(&model.Contact{
		UID:        userFriend.UID,
		RoomID:     room.ID,
		LastMsgID:  newMsg.ID,
		ReadTime:   time.Now(),
		ActiveTime: time.Now(),
	}); err != nil {
		if err := tx.Rollback(); err != nil {
			global.Logger.Errorf("事务回滚失败 %s", err.Error())
			return err
		}
		global.Logger.Errorf("创建会话失败 %s", err.Error())
		return err
	}
	if err := contactQ.Create(&model.Contact{
		UID:       userFriend.FriendUID,
		RoomID:    room.ID,
		LastMsgID: newMsg.ID,
		// 读到时间设为1秒前
		ReadTime:   time.Now().Add(s),
		ActiveTime: time.Now(),
	}); err != nil {
		if err := tx.Rollback(); err != nil {
			global.Logger.Errorf("事务回滚失败 %s", err.Error())
			return err
		}
		global.Logger.Errorf("创建会话失败 %s", err.Error())
		return err
	}
	// 提交
	if err := tx.Commit(); err != nil {
		global.Logger.Errorf("事务提交失败 %s", err.Error())
		return err
	}
	// 发送新消息事件
	newMsgByte, _ := json.Marshal(newMsg)
	msg := &primitive.Message{
		Topic: enum.NewMessageTopic,
		Body:  newMsgByte,
	}
	_, _ = global.RocketProducer.SendSync(ctx, msg)
	return nil
}
