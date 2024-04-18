package service

import (
	"DiTing-Go/dal/model"
	domainEnum "DiTing-Go/domain/enum"
	"DiTing-Go/domain/vo/req"
	"DiTing-Go/global"
	"DiTing-Go/pkg/enum"
	"DiTing-Go/pkg/resp"
	"DiTing-Go/pkg/utils"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"log"
	"strconv"
)

// ApplyFriendService 添加好友
func ApplyFriendService(uid int64, applyReq req.UserApplyReq) resp.ResponseData {
	ctx := context.Background()
	friendUid := applyReq.Uid
	user := global.Query.User
	userQ := user.WithContext(ctx)

	//检查用户是否存在
	fun := func() (interface{}, error) {
		return userQ.Where(user.ID.Eq(friendUid)).First()
	}
	userR := model.User{}
	err := utils.GetData(domainEnum.User+strconv.FormatInt(applyReq.Uid, 10), &userR, fun)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp.ErrorResponseData("用户不存在")
		}
		global.Logger.Errorf("查询用户失败 %s", err)
		return resp.ErrorResponseData("系统正忙，请稍后再试")
	}

	// 检查是否已经是好友关系
	isFriend, err := IsFriend(uid, friendUid)
	if err != nil {
		global.Logger.Errorf("查询好友失败 %s", err)
		return resp.ErrorResponseData("系统正忙，请稍后再试")
	}
	// 已经是好友
	if isFriend {
		return resp.ErrorResponseData("已经是好友")
	}
	// 检查是否已经发送过好友请求
	userApply := global.Query.UserApply
	userApplyQ := userApply.WithContext(ctx)
	userApplyR := model.UserApply{}
	fun = func() (interface{}, error) {
		return userApplyQ.Where(userApply.UID.Eq(uid), userApply.TargetID.Eq(friendUid)).First()
	}
	key := fmt.Sprintf("%s%d_%d", domainEnum.UserApply, uid, friendUid)
	err = utils.GetData(key, &userApplyR, fun)
	// 查到了
	if err == nil {
		return resp.ErrorResponseData("已发送过好友请求，请等待对方同意")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		global.Logger.Errorf("查询好友请求失败 %s", err)
		return resp.ErrorResponseData("系统正忙，请稍后再试")
	}

	// 检查对方是否给我们发送过好友请求，如果是，直接同意
	fun = func() (interface{}, error) {
		return userApplyQ.Where(userApply.UID.Eq(friendUid), userApply.TargetID.Eq(uid)).First()
	}
	key = fmt.Sprintf("%s%d_%d", domainEnum.UserApply, friendUid, uid)
	err = utils.GetData(key, &userApplyR, fun)
	// 查到了
	if err == nil {
		err := AgreeFriend(uid, friendUid)
		if err != nil {
			global.Logger.Errorf("同意好友请求失败 %s", err)
			return resp.ErrorResponseData("系统正忙，请稍后再试")
		}

		return resp.SuccessResponseData(nil)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		global.Logger.Errorf("查询好友请求失败 %s", err)
		return resp.ErrorResponseData("系统正忙，请稍后再试")
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
		return resp.ErrorResponseData("系统正忙，请稍后再试")
	}
	// 发送好友申请事件
	global.Bus.Publish(domainEnum.FriendApplyEvent, model.UserApply{
		UID:        uid,
		TargetID:   friendUid,
		Msg:        applyReq.Msg,
		Status:     enum.NO,
		ReadStatus: enum.NO,
	})
	return resp.SuccessResponseData(nil)
}

// IsFriend 判断是否是好友
func IsFriend(uid, friendUid int64) (bool, error) {
	ctx := context.Background()
	userFriend := global.Query.UserFriend
	userFriendQ := userFriend.WithContext(ctx)
	// 检查是否已经是好友关系
	userFriendR := model.UserFriend{}
	fun := func() (interface{}, error) {
		return userFriendQ.Where(userFriend.UID.Eq(uid), userFriend.FriendUID.Eq(friendUid)).First()
	}
	key := fmt.Sprintf("%s%d_%d", domainEnum.UserFriend, uid, friendUid)
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
	key := fmt.Sprintf("%s%d_%d", domainEnum.UserApply, friendUid, uid)
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
			log.Println("事务回滚失败", err.Error())
		}
		return err
	}

	var userFriends = []*model.UserFriend{
		{
			UID:       uid,
			FriendUID: friendUid,
		},
		{
			UID:       friendUid,
			FriendUID: uid,
		},
	}
	if err = userFriendTx.Create(userFriends...); err != nil {
		if err := tx.Rollback(); err != nil {
			log.Println("事务回滚失败", err.Error())
		}
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	utils.RemoveData(key)
	// 发送新好友事件
	global.Bus.Publish(domainEnum.FriendNewEvent, model.UserFriend{
		UID:       uid,
		FriendUID: friendUid,
	})
	return nil
}
