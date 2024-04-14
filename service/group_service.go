package service

import (
	"DiTing-Go/dal/model"
	"DiTing-Go/domain/enum"
	"DiTing-Go/domain/vo/req"
	"DiTing-Go/global"
	"DiTing-Go/pkg/resp"
	"context"
	"github.com/gin-gonic/gin"
)

// CreateGroupService 创建群聊
//
//	@Summary	创建群聊
//	@Produce	json
//	@Param		name	body		string					true	"群聊名称"
//	@Success	200	{object}	resp.ResponseData	"成功"
//	@Failure	500	{object}	resp.ResponseData	"内部错误"
//	@Router		/api/group/create [post]
func CreateGroupService(c *gin.Context) {
	uid := c.GetInt64("uid")
	creatGroupReq := req.CreateGroupReq{}
	if err := c.ShouldBind(&creatGroupReq); err != nil { //ShouldBind()会自动推导
		resp.ErrorResponse(c, "参数错误")
		global.Logger.Errorf("参数错误: %v", err)
		c.Abort()
		return
	}

	tx := global.Query.Begin()
	ctx := context.Background()
	// 创建房间表
	roomTx := tx.Room.WithContext(ctx)
	newRoom := model.Room{
		Type:    enum.GROUP,
		ExtJSON: "{}",
	}
	if err := roomTx.Create(&newRoom); err != nil {
		if err := tx.Rollback(); err != nil {
			global.Logger.Errorf("事务回滚失败 %s", err.Error())
			return
		}
		resp.ErrorResponse(c, "创建群聊失败")
		c.Abort()
		global.Logger.Errorf("添加房间表失败 %s", err.Error())
		return
	}

	// 查询用户头像
	user := global.Query.User
	userTx := tx.User.WithContext(ctx)
	userR, err := userTx.Where(user.ID.Eq(uid)).First()
	if err != nil {
		if err := tx.Rollback(); err != nil {
			global.Logger.Errorf("事务回滚失败 %s", err.Error())
			return
		}
		resp.ErrorResponse(c, "创建群聊失败")
		c.Abort()
		global.Logger.Errorf("查询用户表失败 %s", err.Error())
		return
	}

	// 创建群聊表
	roomGroupTx := tx.RoomGroup.WithContext(ctx)
	newRoomGroup := model.RoomGroup{
		RoomID: newRoom.ID,
		Name:   creatGroupReq.Name,
		// 默认为创建者头像
		Avatar:  userR.Avatar,
		ExtJSON: "{}",
	}
	if err := roomGroupTx.Create(&newRoomGroup); err != nil {
		if err := tx.Rollback(); err != nil {
			global.Logger.Errorf("事务回滚失败 %s", err.Error())
			return
		}
		resp.ErrorResponse(c, "创建群聊失败")
		c.Abort()
		global.Logger.Errorf("添加群聊表失败 %s", err.Error())
		return
	}

	// 创建会话表
	contactTx := tx.Contact.WithContext(ctx)
	newContact := model.Contact{
		UID:    uid,
		RoomID: newRoom.ID,
	}
	if err := contactTx.Create(&newContact); err != nil {
		if err := tx.Rollback(); err != nil {
			global.Logger.Errorf("事务回滚失败 %s", err.Error())
			return
		}
		resp.ErrorResponse(c, "创建群聊失败")
		c.Abort()
		global.Logger.Errorf("添加会话表失败 %s", err.Error())
		return
	}
	if err := tx.Commit(); err != nil {
		global.Logger.Errorf("事务提交失败 %s", err.Error())
		resp.ErrorResponse(c, "创建群聊失败")
		c.Abort()
		return
	}
	// TODO: 添加群组成员表

	resp.SuccessResponseWithMsg(c, "success")
	return
}

// DeleteGroupService 删除群聊
//
//	@Summary	删除群聊
//	@Produce	json
//	@Param		id	body		string					true	"群聊ID"
//	@Success	200	{object}	resp.ResponseData	"成功"
//	@Failure	500	{object}	resp.ResponseData	"内部错误"
//	@Router		/api/group/:id [delete]
func DeleteGroupService(c *gin.Context) {
	uid := c.GetInt64("uid")
	deleteGroupReq := req.DeleteGroupReq{}
	if err := c.ShouldBindUri(&deleteGroupReq); err != nil { //ShouldBind()会自动推导
		resp.ErrorResponse(c, "参数错误")
		global.Logger.Errorf("参数错误: %v", err)
		c.Abort()
		return
	}

	tx := global.Query.Begin()
	ctx := context.Background()
	// TODO:查询用户是否在群聊中
	// TODO:删除所有成员的会话表
	// 删除会话表
	contact := global.Query.Contact
	contactTx := tx.Contact.WithContext(ctx)
	if _, err := contactTx.Where(contact.UID.Eq(uid), contact.RoomID.Eq(deleteGroupReq.ID)).Delete(); err != nil {
		if err := tx.Rollback(); err != nil {
			global.Logger.Errorf("事务回滚失败 %s", err.Error())
			return
		}
		resp.ErrorResponse(c, "删除群聊失败")
		c.Abort()
		global.Logger.Errorf("删除会话表失败 %s", err.Error())
		return
	}
	// 删除群聊表
	roomGroup := global.Query.RoomGroup
	roomGroupTx := tx.RoomGroup.WithContext(ctx)
	if _, err := roomGroupTx.Where(roomGroup.RoomID.Eq(deleteGroupReq.ID)).Delete(); err != nil {
		if err := tx.Rollback(); err != nil {
			global.Logger.Errorf("事务回滚失败 %s", err.Error())
			return
		}
		resp.ErrorResponse(c, "删除群聊失败")
		c.Abort()
		global.Logger.Errorf("删除群聊表失败 %s", err.Error())
		return
	}
	// 删除房间表
	room := global.Query.Room
	roomTx := tx.Room.WithContext(ctx)
	if _, err := roomTx.Where(room.ID.Eq(deleteGroupReq.ID)).Delete(); err != nil {
		if err := tx.Rollback(); err != nil {
			global.Logger.Errorf("事务回滚失败 %s", err.Error())
			return
		}
		resp.ErrorResponse(c, "删除群聊失败")
		c.Abort()
		global.Logger.Errorf("删除房间表失败 %s", err.Error())
		return
	}
	// TODO:抽取为事件
	// 删除消息表
	message := global.Query.Message
	messageTx := tx.Message.WithContext(ctx)
	msg := model.Message{
		Status: 0,
	}
	if _, err := messageTx.Where(message.RoomID.Eq(deleteGroupReq.ID)).Updates(msg); err != nil {
		if err := tx.Rollback(); err != nil {
			global.Logger.Errorf("事务回滚失败 %s", err.Error())
			return
		}
		resp.ErrorResponse(c, "删除群聊失败")
		c.Abort()
		global.Logger.Errorf("删除消息表失败 %s", err.Error())
		return
	}
	if err := tx.Commit(); err != nil {
		global.Logger.Errorf("事务提交失败 %s", err.Error())
		resp.ErrorResponse(c, "删除群聊失败")
		c.Abort()
		return
	}

	resp.SuccessResponseWithMsg(c, "success")
	return
}
