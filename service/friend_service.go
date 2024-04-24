package service

import (
	"DiTing-Go/dal"
	"DiTing-Go/dal/model"
	"DiTing-Go/domain/dto"
	domainEnum "DiTing-Go/domain/enum"
	"DiTing-Go/domain/vo/req"
	domainResp "DiTing-Go/domain/vo/resp"
	"DiTing-Go/global"
	"DiTing-Go/pkg/cursor"
	"DiTing-Go/pkg/enum"
	"DiTing-Go/pkg/resp"
	"DiTing-Go/pkg/utils"
	"DiTing-Go/utils/jsonUtils"
	"DiTing-Go/utils/redisCache"
	"context"
	"fmt"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
	"gorm.io/gen"
	"gorm.io/gorm"
	"sort"
	"strconv"
)

// ApplyFriendService 添加好友
func ApplyFriendService(uid int64, applyReq req.UserApplyReq) (resp.ResponseData, error) {
	ctx := context.Background()
	friendUid := applyReq.Uid
	user := global.Query.User
	userQ := user.WithContext(ctx)

	uids := utils.Int64Slice{uid, friendUid}
	sort.Sort(uids)
	key := fmt.Sprintf(domainEnum.UserAndFriendLock, uids[0], uids[1])
	mutex, err := utils.GetLock(key)
	if err != nil {
		return resp.ErrorResponseData("系统正忙，请稍后再试"), err
	}
	defer utils.ReleaseLock(mutex)

	//检查用户是否存在
	fun := func() (interface{}, error) {
		return userQ.Where(user.ID.Eq(friendUid)).First()
	}
	userR := model.User{}
	key = fmt.Sprintf(domainEnum.UserCacheByID, applyReq.Uid)
	err = utils.GetData(key, &userR, fun)
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
	err = jsonUtils.SendMsgSync(domainEnum.FriendApplyTopic, model.UserApply{
		UID:        uid,
		TargetID:   friendUid,
		Msg:        applyReq.Msg,
		Status:     enum.NO,
		ReadStatus: enum.NO,
	})
	if err != nil {
		return resp.ErrorResponseData("系统正忙，请稍后再试"), errors.New("Business Error")
	}
	//time.Sleep(30 * time.Second)
	return resp.SuccessResponseData(nil), nil
}

// IsFriendService 判断是否是好友
func IsFriendService(uid, friendUid int64) (resp.ResponseData, error) {
	isFriend, err := IsFriend(uid, friendUid)
	if err != nil {
		global.Logger.Errorf("判断好友关系失败 %s", err)
		return resp.ErrorResponseData("系统正忙，请稍后再试"), errors.New("Business Error")
	}
	return resp.SuccessResponseData(isFriend), nil
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

	uids := utils.Int64Slice{uid, friendUid}
	sort.Sort(uids)
	key := fmt.Sprintf(domainEnum.UserAndFriendLock, uids[0], uids[1])
	mutex, err := utils.GetLock(key)
	if err != nil {
		return resp.ErrorResponseData("系统正忙，请稍后再试"), err
	}
	defer utils.ReleaseLock(mutex)

	err = AgreeFriend(uid, friendUid)
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
	err = jsonUtils.SendMsgSync(domainEnum.NewFriendTopic, userFriends[0])
	if err != nil {
		if err := tx.Rollback(); err != nil {
			global.Logger.Errorf("事务回滚失败 %s", err.Error())
		}
		global.Logger.Errorf("发送新好友事件失败 %s", err.Error())
		return err
	}
	return nil
}

// DeleteFriendService 删除好友
// 只删除好友关系和会话,其他耗时操作异步处理
func DeleteFriendService(uid int64, deleteFriendReq req.DeleteFriendReq) (resp.ResponseData, error) {
	ctx := context.Background()

	deleteFriendUid := deleteFriendReq.Uid

	uids := utils.Int64Slice{uid, deleteFriendUid}
	sort.Sort(uids)
	key := fmt.Sprintf(domainEnum.UserAndFriendLock, uids[0], uids[1])
	mutex, err := utils.GetLock(key)
	if err != nil {
		return resp.ErrorResponseData("系统正忙，请稍后再试"), err
	}
	defer utils.ReleaseLock(mutex)

	// 判断是否为好友
	isFriend, err := IsFriend(uid, deleteFriendUid)
	if err != nil {
		global.Logger.Errorf("查询好友关系失败 %s", err)
		return resp.ErrorResponseData("系统正忙，请稍后再试"), errors.New("Business Error")
	}
	if !isFriend {
		return resp.ErrorResponseData("删除好友失败"), errors.New("Business Error")
	}

	tx := global.Query.Begin()
	// 事务
	// 软删除好友关系
	userFriend := global.Query.UserFriend
	userFriendTx := tx.UserFriend.WithContext(ctx)
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

	// 删除会话
	roomFriend := global.Query.RoomFriend
	roomFriendTx := tx.RoomFriend.WithContext(ctx)
	uids = utils.Int64Slice{uid, deleteFriendUid}
	sort.Sort(uids)
	fun := func() (interface{}, error) {
		return roomFriendTx.Where(roomFriend.Uid1.Eq(uids[0]), roomFriend.Uid2.Eq(uids[1])).First()
	}
	roomFriendR := model.RoomFriend{}
	key = fmt.Sprintf(domainEnum.RoomFriendCacheByUidAndFriendUid, uids[0], uids[1])
	if err := utils.GetData(key, &roomFriendR, fun); err != nil {
		if err := tx.Rollback(); err != nil {
			global.Logger.Errorf("事务回滚失败 %s", err.Error())
		}
		global.Logger.Errorf("查询好友房间失败 %s", err.Error())
		return resp.ErrorResponseData("删除好友失败"), errors.New("Business Error")
	}

	contact := global.Query.Contact
	contactTx := tx.Contact.WithContext(ctx)
	resultInfo, err := contactTx.Where(contact.RoomID.Eq(roomFriendR.RoomID)).Delete()
	if err != nil {
		if err := tx.Rollback(); err != nil {
			global.Logger.Errorf("事务回滚失败 %s", err.Error())
		}
		global.Logger.Errorf("删除会话失败 %s", err.Error())
		return resp.ErrorResponseData("删除好友失败"), errors.New("Business Error")
	}
	if resultInfo.RowsAffected == 0 {
		if err := tx.Rollback(); err != nil {
			global.Logger.Errorf("事务回滚失败 %s", err.Error())
		}
		global.Logger.Errorf("会话不存在 %d", roomFriendR.RoomID)
		return resp.ErrorResponseData("删除好友失败"), errors.New("Business Error")
	}
	// TODO:删除缓存

	// 发送消息
	DeleteFriendDto := dto.DeleteFriendDto{
		Uid:       uid,
		FriendUid: deleteFriendUid,
	}
	err = jsonUtils.SendMsgSync(domainEnum.DeleteFriendTopic, DeleteFriendDto)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			global.Logger.Errorf("事务回滚失败 %s", err.Error())
		}
		global.Logger.Errorf("发送删除好友事件失败 %s", err.Error())
		return resp.ErrorResponseData("系统正忙，请稍后再试"), errors.New("Business Error")
	}

	if err := tx.Commit(); err != nil {
		global.Logger.Errorf("事务提交失败 %s", err.Error())
		return resp.ErrorResponseData("删除好友失败"), errors.New("Business Error")
	}

	return resp.SuccessResponseData(nil), nil
}

// GetUserApplyService 获取好友申请列表
func GetUserApplyService(uid int64, pageReq cursor.PageReq) (resp.ResponseData, error) {
	ctx := context.Background()
	// 获取 UserApply 表中 TargetID 等于 uid(登录用户ID)的用户ID集合，采用游标分页
	db := dal.DB
	userApplys := make([]model.UserApply, 0)
	condition := []interface{}{"target_id=?", strconv.FormatInt(uid, 10)}
	pageResp, err := cursor.Paginate(db, pageReq, &userApplys, "create_time", false, condition...)
	if err != nil {
		global.Logger.Errorf("查询好友申请表失败 %s", err)
		return resp.ErrorResponseData("系统正忙，请稍后再试"), errors.New("Business Error")
	}

	uids := make([]int64, 0)
	n := len(userApplys)
	for i := 0; i < n; i++ {
		uids = append(uids, userApplys[i].UID)
	}

	user := global.Query.User
	// 根据 uids 集合查询 User 表
	users, err := user.WithContext(ctx).Select(user.ID, user.Name, user.Avatar).Where(user.ID.In(uids...)).Find()
	if err != nil {
		global.Logger.Errorf("查询用户表失败 %s", err)
		return resp.ErrorResponseData("系统正忙，请稍后再试"), errors.New("Business Error")
	}

	if len(users) != len(userApplys) {
		global.Logger.Errorf("用户表和好友申请表数据不匹配")
		return resp.ErrorResponseData("系统正忙，请稍后再试"), errors.New("Business Error")
	}
	var usersVO = make([]domainResp.UserApplyResp, 0)
	// 数据转换
	for i := 0; i < len(users); i++ {
		var userVO domainResp.UserApplyResp
		_ = copier.Copy(&userVO, &users[i])
		userVO.Msg = userApplys[i].Msg
		userVO.Status = userApplys[i].Status
		usersVO = append(usersVO, userVO)
	}
	pageResp.Data = usersVO

	userApply := global.Query.UserApply
	userApplyQ := global.Query.UserApply.WithContext(ctx)
	// 更新已读状态
	_, err = userApplyQ.Where(userApply.TargetID.Eq(uid), userApply.ReadStatus.Eq(enum.NO)).Update(userApply.ReadStatus, enum.YES)
	if err != nil {
		global.Logger.Errorf("更新好友申请表失败 %s", err)
		return resp.ErrorResponseData("系统正忙，请稍后再试"), errors.New("Business Error")
	}

	return resp.SuccessResponseData(pageResp), nil
}

func UnreadApplyNumService(uid int64) (resp.ResponseData, error) {
	ctx := context.Background()
	userApply := global.Query.UserApply
	userApplyQ := userApply.WithContext(ctx)

	// TODO 直接count
	// 获取 UserApply 表中 TargetID 等于 uid(登录用户ID)的用户ID集合
	subQuery := userApplyQ.Where(userApply.TargetID.Eq(uid), userApply.ReadStatus.Eq(enum.NO)).Limit(99)
	num, err := gen.Table(subQuery.As("t")).Count()
	if err != nil {
		global.Logger.Errorf("查询好友申请表失败 %s", err)
		return resp.ErrorResponseData("系统正忙，请稍后再试"), errors.New("Business Error")
	}
	return resp.SuccessResponseData(num), nil
}
