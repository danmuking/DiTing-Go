package service

import (
	"DiTing-Go/dal/model"
	"DiTing-Go/domain/enum"
	"DiTing-Go/domain/vo/req"
	"DiTing-Go/global"
	"DiTing-Go/pkg/resp"
	"context"
	"github.com/gin-gonic/gin"
	"log"
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
			log.Println("事务回滚失败", err.Error())
			return
		}
		resp.ErrorResponse(c, "创建群聊失败")
		c.Abort()
		log.Println("添加房间表失败", err.Error())
		return
	}

	// 查询用户头像
	user := global.Query.User
	userTx := tx.User.WithContext(ctx)
	userR, err := userTx.Where(user.ID.Eq(uid)).First()
	if err != nil {
		if err := tx.Rollback(); err != nil {
			log.Println("事务回滚失败", err.Error())
			return
		}
		resp.ErrorResponse(c, "创建群聊失败")
		c.Abort()
		log.Println("查询用户表失败", err.Error())
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
			log.Println("事务回滚失败", err.Error())
			return
		}
		resp.ErrorResponse(c, "创建群聊失败")
		c.Abort()
		log.Println("添加群聊表失败", err.Error())
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
			log.Println("事务回滚失败", err.Error())
			return
		}
		resp.ErrorResponse(c, "创建群聊失败")
		c.Abort()
		log.Println("添加会话表失败", err.Error())
		return
	}
	if err := tx.Commit(); err != nil {
		log.Println("事务提交失败", err.Error())
		resp.ErrorResponse(c, "创建群聊失败")
		c.Abort()
		return
	}

	resp.SuccessResponseWithMsg(c, "success")
	return
}
