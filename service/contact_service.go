package service

import (
	"DiTing-Go/dal"
	"DiTing-Go/dal/model"
	"DiTing-Go/dal/query"
	"DiTing-Go/models/vo"
	"DiTing-Go/pkg/enum"
	"DiTing-Go/pkg/resp"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"log"
)

var q *query.Query

func init() {
	dal.DB = dal.ConnectDB(MySQLDSN).Debug()
	// 设置默认DB对象
	query.SetDefault(dal.DB)
	q = query.Use(dal.DB)
}

// ApplyFriend 添加好友
//
//	@Summary	添加好友
//	@Produce	json
//	@Param		uid	body		int					true	"好友uid"
//	@Param		msg	body		string				true	"验证消息"
//	@Success	200	{object}	resp.ResponseData	"成功"
//	@Failure	500	{object}	resp.ResponseData	"内部错误"
//	@Router		/api/contact/add [post]
func ApplyFriend(c *gin.Context) {
	uid := c.GetInt64("uid")
	applyDto := model.UserApplyDto{}
	if err := c.ShouldBind(&applyDto); err != nil { //ShouldBind()会自动推导
		resp.ErrorResponse(c, "参数错误")
		return
	}
	friendUid := applyDto.Uid
	//检查用户是否存在
	user, err := query.User.WithContext(context.Background()).Where(query.User.ID.Eq(friendUid)).First()
	if user == nil {
		resp.ErrorResponse(c, "用户不存在")
		c.Abort()
		return
	}
	// 检查是否已经是好友关系
	if isFriend := isFriend(c, uid, friendUid); isFriend {
		resp.ErrorResponse(c, "好友已存在")
		c.Abort()
		return
	}
	// 检查是否已经发送过好友请求
	friendApply, err := query.UserApply.WithContext(context.Background()).Where(query.UserApply.UID.Eq(uid), query.UserApply.TargetID.Eq(friendUid)).First()
	if err != nil && err.Error() != "record not found" {
		resp.ErrorResponse(c, "参数错误")
		c.Abort()
		return
	}
	if friendApply != nil {
		resp.ErrorResponse(c, "已发送过好友请求，请等待对方同意")
		c.Abort()
		return
	}
	// 检查对方是否给我们发送过好友请求，如果是，直接同意
	apply, err := query.UserApply.WithContext(context.Background()).Where(query.UserApply.UID.Eq(friendUid), query.UserApply.TargetID.Eq(uid)).First()
	if err != nil && err.Error() != "record not found" {
		resp.ErrorResponse(c, "参数错误")
		c.Abort()
		return
	}
	if apply != nil {
		// 同意好友请求
		apply.Status = 2
		_, err := query.UserApply.WithContext(context.Background()).Where(query.UserApply.UID.Eq(friendUid), query.UserApply.TargetID.Eq(uid)).Updates(apply)
		if err != nil {
			resp.ErrorResponse(c, "参数错误")
			c.Abort()
			return
		}
		// 添加好友关系
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
		err = query.UserFriend.WithContext(context.Background()).Create(userFriends...)
		if err != nil {
			resp.ErrorResponse(c, "参数错误")
			c.Abort()
			return
		}
		resp.SuccessResponseWithMsg(c, "success")
		return
	}
	// 添加好友请求
	err = query.UserApply.WithContext(context.Background()).Create(&model.UserApply{
		UID:        uid,
		TargetID:   friendUid,
		Msg:        applyDto.Msg,
		Status:     enum.NO,
		ReadStatus: enum.NO,
	})
	if err != nil {
		resp.ErrorResponse(c, "参数错误")
		c.Abort()
		return
	}
	resp.SuccessResponseWithMsg(c, "success")
}

// DeleteFriend 删除好友
//
//	@Summary	删除好友
//	@Produce	json
//	@Param		uid	body		int					true	"好友uid"
//	@Success	200	{object}	resp.ResponseData	"成功"
//	@Failure	500	{object}	resp.ResponseData	"内部错误"
//	@Router		/api/contact/delete [delete]
func DeleteFriend(c *gin.Context) {
	uid := c.GetInt64("uid")
	friend := model.Uid{}
	if err := c.ShouldBind(&friend); err != nil { //ShouldBind()会自动推导
		resp.ErrorResponse(c, "参数错误")
		return
	}
	friendUid := friend.Uid
	// 检查是否为好友
	if isFriend := isFriend(c, uid, friendUid); isFriend {
		// 事务
		err := q.Transaction(func(tx *query.Query) error {
			// 删除好友关系
			if _, err := tx.UserFriend.WithContext(context.Background()).Where(query.UserFriend.UID.Eq(uid), query.UserFriend.FriendUID.Eq(friendUid)).Delete(); err != nil {
				return err
			}
			if _, err := tx.UserFriend.WithContext(context.Background()).Where(query.UserFriend.UID.Eq(friendUid), query.UserFriend.FriendUID.Eq(uid)).Delete(); err != nil {
				return err
			}
			// 删除好友申请
			if _, err := tx.UserApply.WithContext(context.Background()).Where(query.UserApply.UID.Eq(uid), query.UserApply.TargetID.Eq(friendUid)).Delete(); err != nil {
				return err
			}
			if _, err := tx.UserApply.WithContext(context.Background()).Where(query.UserApply.UID.Eq(friendUid), query.UserApply.TargetID.Eq(uid)).Delete(); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			resp.ErrorResponse(c, "删除失败")
			c.Abort()
			return
		}
	}
	resp.SuccessResponseWithMsg(c, "success")
	return
}

func isFriend(c *gin.Context, uid, friendUid int64) bool {
	// 检查是否已经是好友关系
	friend, err := query.UserFriend.WithContext(context.Background()).Where(query.UserFriend.UID.Eq(uid), query.UserFriend.FriendUID.Eq(friendUid)).First()
	if err != nil && err.Error() != "record not found" {
		resp.ErrorResponse(c, "参数错误")
		c.Abort()
	}
	if friend == nil {
		return false
	}
	return true
}

// GetApplyList 获取用户好友申请列表
//
//	@Summary	获取用户好友申请列表
//	@Success	200	{object}	resp.ResponseData	"成功"
//	@Failure	500	{object}	resp.ResponseData	"内部错误"
//	@Router		/api/contact/getApplyList [get]
func GetApplyList(c *gin.Context) {

	ctx := context.Background()

	uid := c.GetInt64("uid")

	ua := query.UserApply
	// 获取 UserApply 表中 TargetID 等于 uid(登录用户ID)的用户ID集合
	// select uid form user_apply where target_id = ?
	userApplyIDs, err := ua.WithContext(ctx).Select(ua.UID).Where(ua.TargetID.Eq(uid)).Find()
	if err != nil {
		// todo 添加日志系统
		log.Printf("DB excete Sql happen [ERROR], err msg is : %v", err)
		resp.ErrorResponse(c, "系统繁忙，亲稍后再试")
		return
	}

	var uids = make([]int64, 0)
	var n int
	n = len(userApplyIDs)
	for i := 0; i < n; i++ {
		uids = append(uids, userApplyIDs[i].UID)
	}

	// 根据 uids 集合查询 User 表sex
	// select id , name , avatar , sex , active_status , last_opt_time form user where status = 0 and id in (...)
	u := query.User
	users, err := u.WithContext(ctx).Select(u.ID, u.Name, u.Avatar, u.Sex, u.ActiveStatus, u.LastOptTime).Where(u.ID.In(uids...), u.Status.Eq(0)).Find()
	if err != nil {
		// todo 添加日志系统
		log.Printf("DB excete Sql happen [ERROR], err msg is : %v", err)
		resp.ErrorResponse(c, "系统繁忙，亲稍后再试")
		c.Abort()
		return
	}

	var usersVO = make([]vo.UserVo, 0)

	// 数据转换
	_ = copier.Copy(&usersVO, &users)

	resp.SuccessResponse(c, usersVO)
}
