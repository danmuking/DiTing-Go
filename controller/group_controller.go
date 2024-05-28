package controller

import (
	"DiTing-Go/domain/vo/req"
	"DiTing-Go/global"
	"DiTing-Go/pkg/domain/vo/resp"
	"DiTing-Go/service"
	"github.com/gin-gonic/gin"
)

// CreateGroupController 创建群聊
//
//	@Summary	创建群聊
//	@Produce	json
//	@Param		uidList		body		array				true	"用户id列表"
//	@Success	200			{object}	resp.ResponseData	"成功"
//	@Failure	500			{object}	resp.ResponseData	"内部错误"
//	@Router		/api/group/create [post]
func CreateGroupController(c *gin.Context) {
	uid := c.GetInt64("uid")
	creatGroupReq := req.CreateGroupReq{}
	if err := c.ShouldBind(&creatGroupReq); err != nil { //ShouldBind()会自动推导
		resp.ErrorResponse(c, "参数错误")
		global.Logger.Errorf("参数错误: %v", err)
		c.Abort()
		return
	}
	response, err := service.CreateGroupService(uid, creatGroupReq.UidList)
	if err != nil {
		c.Abort()
		resp.ReturnErrorResponse(c, response)
		return
	}
	resp.ReturnSuccessResponse(c, response)
}

// DeleteGroupController 解散群聊
//
//	@Summary	解散群聊
//	@Produce	json
//	@Param		roomId		body		int				true	"房间id"
//	@Success	200			{object}	resp.ResponseData	"成功"
//	@Failure	500			{object}	resp.ResponseData	"内部错误"
//	@Router		/api/group/delete/ [post]
func DeleteGroupController(c *gin.Context) {
	uid := c.GetInt64("uid")
	deleteGroupReq := req.DeleteGroupReq{}
	if err := c.ShouldBind(&deleteGroupReq); err != nil { //ShouldBind()会自动推导
		resp.ErrorResponse(c, "参数错误")
		global.Logger.Errorf("参数错误: %v", err)
		c.Abort()
		return
	}
	response, err := service.DeleteGroupService(uid, deleteGroupReq.RoomId)
	if err != nil {
		c.Abort()
		resp.ReturnErrorResponse(c, response)
		return
	}
	resp.ReturnSuccessResponse(c, response)
}

// JoinGroupController 加入群聊
//
//	@Summary	加入群聊
//	@Produce	json
//	@Param		id	body		int					true	"房间id"
//	@Success	200	{object}	resp.ResponseData	"成功"
//	@Failure	500	{object}	resp.ResponseData	"内部错误"
//	@Router		/api/group/join [post]
func JoinGroupController(c *gin.Context) {
	uid := c.GetInt64("uid")
	joinGroupReq := req.JoinGroupReq{}
	if err := c.ShouldBind(&joinGroupReq); err != nil { //ShouldBind()会自动推导
		resp.ErrorResponse(c, "参数错误")
		global.Logger.Errorf("参数错误: %v", err)
		c.Abort()
		return
	}
	response, err := service.JoinGroupService(uid, joinGroupReq.RoomId)
	if err != nil {
		c.Abort()
		resp.ReturnErrorResponse(c, response)
		return
	}
	resp.ReturnSuccessResponse(c, response)
}
