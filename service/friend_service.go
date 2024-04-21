package service

import (
	"DiTing-Go/dal/model"
	domainEnum "DiTing-Go/domain/enum"
	"DiTing-Go/domain/vo/req"
	"DiTing-Go/global"
	"DiTing-Go/pkg/enum"
	"DiTing-Go/pkg/resp"
	"DiTing-Go/pkg/utils"
	"DiTing-Go/utils/redisCache"
	"context"
	"encoding/json"
	"fmt"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/go-redsync/redsync/v4"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"sort"
	"strconv"
)

// ApplyFriendService 添加好友
// TODO：考虑软删除条件
func ApplyFriendService(uid int64, applyReq req.UserApplyReq) (resp.ResponseData, error) {
	ctx := context.Background()
	friendUid := applyReq.Uid
	user := global.Query.User
	userQ := user.WithContext(ctx)

	mutex := global.RedSync.NewMutex(domainEnum.UserLock + strconv.FormatInt(uid, 10))
	if err := mutex.LockContext(ctx); err != nil {
		global.Logger.Errorf("加锁失败 %s", err)
		return resp.ErrorResponseData("系统正忙，请稍后再试"), errors.New("Business Error")
	}
	defer func(mutex *redsync.Mutex) {
		_, err := mutex.Unlock()
		if err != nil {
			global.Logger.Errorf("解锁失败 %s", err)
		}
	}(mutex)

	//检查用户是否存在
	fun := func() (interface{}, error) {
		return userQ.Where(user.ID.Eq(friendUid)).First()
	}
	userR := model.User{}
	key := fmt.Sprintf(domainEnum.UserCacheByID, applyReq.Uid)
	err := utils.GetData(key, &userR, fun)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp.ErrorResponseData("用户不存在"), errors.New("Business Error")
		}
		global.Logger.Errorf("查询用户失败 %s", err)
		return resp.ErrorResponseData("系统正忙，请稍后再试"), errors.New("Business Error")
	}

	// 检查是否已经是好友关系
	isFriend, err := IsFriend(uid, friendUid)
	if err != nil {
		global.Logger.Errorf("查询好友失败 %s", err)
		return resp.ErrorResponseData("系统正忙，请稍后再试"), errors.New("Business Error")
	}
	// 已经是好友
	if isFriend {
		return resp.ErrorResponseData("已经是好友"), errors.New("Business Error")
	}
	// 检查是否已经发送过好友请求
	userApply := global.Query.UserApply
	userApplyQ := userApply.WithContext(ctx)
	userApplyR := model.UserApply{}
	fun = func() (interface{}, error) {
		return userApplyQ.Where(userApply.UID.Eq(uid), userApply.TargetID.Eq(friendUid)).First()
	}
	key = fmt.Sprintf(domainEnum.UserApplyCacheByUidAndFriendUid, uid, friendUid)
	err = utils.GetData(key, &userApplyR, fun)
	// 查到了
	if err == nil {
		return resp.ErrorResponseData("已发送过好友请求，请等待对方同意"), errors.New("Business Error")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		global.Logger.Errorf("查询好友请求失败 %s", err)
		return resp.ErrorResponseData("系统正忙，请稍后再试"), errors.New("Business Error")
	}

	// 检查对方是否给我们发送过好友请求，如果是，直接同意
	fun = func() (interface{}, error) {
		return userApplyQ.Where(userApply.UID.Eq(friendUid), userApply.TargetID.Eq(uid)).First()
	}
	key = fmt.Sprintf(domainEnum.UserApplyCacheByUidAndFriendUid, friendUid, uid)
	err = utils.GetData(key, &userApplyR, fun)
	// 查到了
	if err == nil {
		err := AgreeFriend(uid, friendUid)
		if err != nil {
			global.Logger.Errorf("同意好友请求失败 %s", err)
			return resp.ErrorResponseData("系统正忙，请稍后再试"), errors.New("Business Error")
		}

		return resp.SuccessResponseData(nil), nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		global.Logger.Errorf("查询好友请求失败 %s", err)
		return resp.ErrorResponseData("系统正忙，请稍后再试"), errors.New("Business Error")
	}

	// 发送好友请求
	err = userApplyQ.Create(&model.UserApply{
		UID:        uid,
		TargetID:   friendUid,
		Msg:        applyReq.Msg,
		Status:     enum.NO,
		ReadStatus: enum.NO,
	})
	if err != nil {
		global.Logger.Errorf("插入好友请求失败 %s", err)
		return resp.ErrorResponseData("系统正忙，请稍后再试"), errors.New("Business Error")
	}
	// 发送好友申请事件
	global.Bus.Publish(domainEnum.FriendApplyEvent, model.UserApply{
		UID:        uid,
		TargetID:   friendUid,
		Msg:        applyReq.Msg,
		Status:     enum.NO,
		ReadStatus: enum.NO,
	})
	return resp.SuccessResponseData(nil), nil
}

// IsFriend 判断是否是好友
func IsFriend(uid, friendUid int64) (bool, error) {
	ctx := context.Background()
	userFriend := global.Query.UserFriend
	userFriendQ := userFriend.WithContext(ctx)
	// 检查是否已经是好友关系
	userFriendR := model.UserFriend{}
	fun := func() (interface{}, error) {
		return userFriendQ.Where(userFriend.UID.Eq(uid), userFriend.FriendUID.Eq(friendUid), userFriend.DeleteStatus.Eq(enum.NORMAL)).First()
	}
	key := fmt.Sprintf(domainEnum.UserFriendCacheByUidAndFriendUid, uid, friendUid)
	err := utils.GetData(key, &userFriendR, fun)
	if err != nil {
		// 没查到，不是好友
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		global.Logger.Errorf("查询好友失败 %s", err)
		return false, err
	}

	return true, nil
}

func AgreeFriendService(uid, friendUid int64) (resp.ResponseData, error) {

	ctx := context.Background()
	mutex := global.RedSync.NewMutex(domainEnum.UserLock + strconv.FormatInt(uid, 10))
	if err := mutex.LockContext(ctx); err != nil {
		global.Logger.Errorf("加锁失败 %s", err)
		return resp.ErrorResponseData("系统正忙，请稍后再试"), errors.New("Business Error")
	}
	defer func(mutex *redsync.Mutex) {
		_, err := mutex.Unlock()
		if err != nil {
			global.Logger.Errorf("解锁失败 %s", err)
		}
	}(mutex)

	err := AgreeFriend(uid, friendUid)
	if err != nil {
		global.Logger.Errorf("同意好友请求失败 %s", err)
		return resp.ErrorResponseData("系统正忙，请稍后再试"), errors.New("Business Error")
	}
	return resp.SuccessResponseData(nil), nil
}

// AgreeFriend 同意好友请求
func AgreeFriend(uid, friendUid int64) error {
	ctx := context.Background()
	userApply := global.Query.UserApply
	userApplyQ := userApply.WithContext(ctx)

	// 检查是否存在好友申请且状态为待审批
	fun := func() (interface{}, error) {
		return userApplyQ.Where(userApply.UID.Eq(friendUid), userApply.TargetID.Eq(uid)).First()
	}
	userApplyR := model.UserApply{}
	key := fmt.Sprintf(domainEnum.UserApplyCacheByUidAndFriendUid, friendUid, uid)
	err := utils.GetData(key, &userApplyR, fun)
	if err != nil {
		return err
	}
	// 好友申请状态不是待审批
	if userApplyR.Status != enum.NO {
		return errors.New("error status")
	}
	// 同意好友请求
	userApplyR.Status = enum.YES
	// 事务
	tx := q.Begin()
	userApplyTx := tx.UserApply.WithContext(context.Background())
	userFriendTx := tx.UserFriend.WithContext(context.Background())
	if _, err = userApplyTx.Where(userApply.UID.Eq(friendUid), userApply.TargetID.Eq(uid)).Updates(userApplyR); err != nil {
		if err := tx.Rollback(); err != nil {
			global.Logger.Errorf("事务回滚失败 %s", err.Error())
		}
		return err
	}
	defer utils.RemoveData(key)

	var userFriends = []*model.UserFriend{
		{
			UID:          uid,
			FriendUID:    friendUid,
			DeleteStatus: enum.NORMAL,
		},
		{
			UID:          friendUid,
			FriendUID:    uid,
			DeleteStatus: enum.NORMAL,
		},
	}
	// 检查是否存在软删除状态的好友关系
	userFriend := global.Query.UserFriend
	fun = func() (interface{}, error) {
		return userFriendTx.Where(userFriend.UID.Eq(uid), userFriend.FriendUID.Eq(friendUid), userFriend.DeleteStatus.Eq(enum.DELETED)).First()
	}
	userFriendR := model.UserFriend{}
	key = fmt.Sprintf(domainEnum.UserFriendCacheByUidAndFriendUid, uid, friendUid)
	err = utils.GetData(key, &userFriendR, fun)
	// err
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		global.Logger.Errorf("更新好友关系失败 %s", err.Error())
		return err
	}
	// 查到了,更新状态
	if err == nil {
		_, err := userFriendTx.Where(userFriend.UID.Eq(uid), userFriend.FriendUID.Eq(friendUid)).Update(userFriend.DeleteStatus, enum.NORMAL)
		_, err = userFriendTx.Where(userFriend.UID.Eq(friendUid), userFriend.FriendUID.Eq(uid)).Update(userFriend.DeleteStatus, enum.NORMAL)
		if err != nil {
			if err := tx.Rollback(); err != nil {
				global.Logger.Errorf("事务回滚失败 %s", err.Error())
			}
			global.Logger.Errorf("更新好友关系失败 %s", err.Error())
			return err
		}
		// 删除redis缓存
		key = fmt.Sprintf(domainEnum.UserFriendCacheByUidAndFriendUid, uid, friendUid)
		defer utils.RemoveData(key)
		key = fmt.Sprintf(domainEnum.UserFriendCacheByUidAndFriendUid, friendUid, uid)
		defer utils.RemoveData(key)
	} else {
		// 没查到，创建新的好友关系
		if err = userFriendTx.Create(userFriends...); err != nil {
			if err := tx.Rollback(); err != nil {
				global.Logger.Errorf("事务回滚失败 %s", err)
			}
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}

	// 发送新好友事件
	userFriendMsg, err := json.Marshal(*userFriends[0])
	if err != nil {
		global.Logger.Errorf("json序列化失败 %v", err)
	}
	msg := &primitive.Message{
		Topic: domainEnum.NewFriendTopic,
		Body:  userFriendMsg,
	}
	_, err = global.RocketProducer.SendSync(ctx, msg)
	if err != nil {
		global.Logger.Errorf("发送新好友事件失败 %s", err)
	}
	return nil
}

// DeleteFriendService 删除好友
func DeleteFriendService(uid int64, deleteFriendReq req.DeleteFriendReq) (resp.ResponseData, error) {
	ctx := context.Background()

	mutex := global.RedSync.NewMutex(domainEnum.UserLock + strconv.FormatInt(uid, 10))
	if err := mutex.LockContext(ctx); err != nil {
		global.Logger.Errorf("加锁失败 %s", err)
		return resp.ErrorResponseData("系统正忙，请稍后再试"), errors.New("Business Error")
	}
	defer func(mutex *redsync.Mutex) {
		_, err := mutex.Unlock()
		if err != nil {
			global.Logger.Errorf("解锁失败 %s", err)
		}
	}(mutex)

	deleteFriendUid := deleteFriendReq.Uid
	isFriend, err := IsFriend(uid, deleteFriendUid)
	if err != nil {
		global.Logger.Errorf("查询好友关系失败 %s", err)
		return resp.ErrorResponseData("系统正忙，请稍后再试"), errors.New("Business Error")
	}
	if !isFriend {
		return resp.ErrorResponseData("删除好友失败"), errors.New("Business Error")
	}
	tx := global.Query.Begin()
	userFriend := global.Query.UserFriend
	userFriendTx := tx.UserFriend.WithContext(ctx)
	userApply := global.Query.UserApply
	userApplyTx := tx.UserApply.WithContext(ctx)
	// 事务
	// 软删除好友关系
	if _, err := userFriendTx.Where(userFriend.UID.Eq(uid), userFriend.FriendUID.Eq(deleteFriendUid)).Update(userFriend.DeleteStatus, enum.DELETED); err != nil {
		if err := tx.Rollback(); err != nil {
			global.Logger.Errorf("事务回滚失败 %s", err.Error())
		}
		global.Logger.Errorf("删除好友失败 %s", err.Error())
		return resp.ErrorResponseData("删除好友失败"), errors.New("Business Error")
	}
	// 删除redis缓存
	defer redisCache.RemoveUserFriend(uid, deleteFriendUid)

	if _, err := userFriendTx.Where(userFriend.UID.Eq(deleteFriendUid), userFriend.FriendUID.Eq(uid)).Update(userFriend.DeleteStatus, enum.DELETED); err != nil {
		if err := tx.Rollback(); err != nil {
			global.Logger.Errorf("事务回滚失败 %s", err.Error())
		}
		global.Logger.Errorf("删除好友失败 %s", err.Error())
		return resp.ErrorResponseData("删除好友失败"), errors.New("Business Error")
	}
	// 删除redis缓存
	defer redisCache.RemoveUserFriend(deleteFriendUid, uid)

	// 删除好友申请
	if _, err := userApplyTx.Where(userApply.UID.Eq(uid), userApply.TargetID.Eq(deleteFriendUid)).Delete(); err != nil {
		if err := tx.Rollback(); err != nil {
			global.Logger.Errorf("事务回滚失败 %s", err.Error())
		}
		global.Logger.Errorf("删除好友失败 %s", err.Error())
		return resp.ErrorResponseData("删除好友失败"), errors.New("Business Error")
	}
	// 删除redis缓存
	defer redisCache.RemoveUserApply(uid, deleteFriendUid)

	if _, err := userApplyTx.Where(userApply.UID.Eq(deleteFriendUid), userApply.TargetID.Eq(uid)).Delete(); err != nil {
		if err := tx.Rollback(); err != nil {
			global.Logger.Errorf("事务回滚失败 %s", err.Error())
		}
		global.Logger.Errorf("删除好友失败 %s", err.Error())
		return resp.ErrorResponseData("删除好友失败"), errors.New("Business Error")
	}
	// 删除redis缓存
	defer redisCache.RemoveUserApply(deleteFriendUid, uid)

	// 软删除好友房间
	roomFriend := global.Query.RoomFriend
	roomFriendTx := tx.RoomFriend.WithContext(ctx)
	uids := utils.Int64Slice{uid, deleteFriendUid}
	sort.Sort(uids)
	fun := func() (interface{}, error) {
		return roomFriendTx.Where(roomFriend.Uid1.Eq(uids[0]), roomFriend.Uid2.Eq(uids[1])).First()
	}
	roomFriendR := model.RoomFriend{}
	key := fmt.Sprintf(domainEnum.RoomFriendCacheByUidAndFriendUid, uids[0], uids[1])
	err = utils.GetData(key, &roomFriendR, fun)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			global.Logger.Errorf("事务回滚失败 %s", err.Error())
		}
		global.Logger.Errorf("查询好友房间失败 %s", err.Error())
		return resp.ErrorResponseData("删除好友失败"), errors.New("Business Error")
	}

	if _, err := roomFriendTx.Where(roomFriend.ID.Eq(roomFriendR.ID)).Update(roomFriend.DeleteStatus, enum.DELETED); err != nil {
		if err := tx.Rollback(); err != nil {
			global.Logger.Errorf("事务回滚失败 %s", err.Error())
		}
		global.Logger.Errorf("删除好友房间失败 %s", err.Error())
		return resp.ErrorResponseData("删除好友失败"), errors.New("Business Error")
	}
	// 删除redis缓存
	defer redisCache.RemoveRoomFriend(roomFriendR)

	// 软删除房间表
	room := global.Query.Room
	roomTx := tx.Room.WithContext(ctx)
	if _, err := roomTx.Where(room.ID.Eq(roomFriendR.RoomID)).Update(room.DeleteStatus, enum.DELETED); err != nil {
		if err := tx.Rollback(); err != nil {
			global.Logger.Errorf("事务回滚失败 %s", err.Error())
		}
		global.Logger.Errorf("删除房间失败 %s", err.Error())
		return resp.ErrorResponseData("删除好友失败"), errors.New("Business Error")
	}
	// 删除redis缓存
	roomR := model.Room{
		ID: roomFriendR.RoomID,
	}
	defer redisCache.RemoveRoomCache(roomR)

	// 删除消息表
	msg := global.Query.Message
	msgTx := tx.Message.WithContext(ctx)
	if _, err := msgTx.Where(msg.RoomID.Eq(roomFriendR.RoomID)).Update(msg.DeleteStatus, enum.DELETED); err != nil {
		if err := tx.Rollback(); err != nil {
			global.Logger.Errorf("事务回滚失败 %s", err.Error())
		}
		global.Logger.Errorf("删除消息失败 %s", err.Error())
		return resp.ErrorResponseData("删除好友失败"), errors.New("Business Error")
	}
	// TODO: 删除消息表缓存

	// 删除会话
	contact := global.Query.Contact
	contactTx := tx.Contact.WithContext(ctx)
	if _, err := contactTx.Where(contact.RoomID.Eq(roomFriendR.RoomID)).Delete(); err != nil {
		if err := tx.Rollback(); err != nil {
			global.Logger.Errorf("事务回滚失败 %s", err.Error())
		}
		global.Logger.Errorf("删除会话失败 %s", err.Error())
		return resp.ErrorResponseData("删除好友失败"), errors.New("Business Error")
	}
	// TODO:删除缓存

	if err := tx.Commit(); err != nil {
		global.Logger.Errorf("事务提交失败 %s", err.Error())
		return resp.ErrorResponseData("删除好友失败"), errors.New("Business Error")
	}

	return resp.SuccessResponseData(nil), nil
}
