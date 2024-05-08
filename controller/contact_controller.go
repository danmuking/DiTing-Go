package controller

import (
	"DiTing-Go/domain/vo/req"
	pkgReq "DiTing-Go/pkg/domain/vo/req"
	"DiTing-Go/pkg/domain/vo/resp"
	"DiTing-Go/service"
	"github.com/gin-gonic/gin"
)

func GetUserInfoBatchController(c *gin.Context) {
	getUserInfoBatchReq := req.GetUserInfoBatchReq{}
	if err := c.ShouldBind(&getUserInfoBatchReq); err != nil {
		resp.ErrorResponse(c, "参数错误")
		c.Abort()
		return
	}
	response, err := service.GetUserInfoBatchService(getUserInfoBatchReq)
	if err != nil {
		c.Abort()
		resp.ReturnErrorResponse(c, response)
		return
	}
	resp.ReturnSuccessResponse(c, response)
	return
}

func GetContactListController(c *gin.Context) {
	uid := c.GetInt64("uid")
	// 游标翻页
	// 默认值
	var cursor *string = nil
	var pageSize int = 20
	pageRequest := pkgReq.PageReq{
		Cursor:   cursor,
		PageSize: pageSize,
	}
	if err := c.ShouldBindQuery(&pageRequest); err != nil { //ShouldBind()会自动推导
		resp.ErrorResponse(c, "参数错误")
		c.Abort()
		return
	}
	response, err := service.GetContactListService(uid, pageRequest)
	if err != nil {
		c.Abort()
		resp.ReturnErrorResponse(c, response)
		return
	}
	resp.ReturnSuccessResponse(c, response)
}

func GetNewContactListController(c *gin.Context) {
	uid := c.GetInt64("uid")

	getNewContentListReq := req.GetNewContentListReq{}
	if err := c.ShouldBindQuery(&getNewContentListReq); err != nil { //ShouldBind()会自动推导
		resp.ErrorResponse(c, "参数错误")
		c.Abort()
		return
	}
	response, err := service.GetNewContactListService(uid, getNewContentListReq)
	if err != nil {
		c.Abort()
		resp.ReturnErrorResponse(c, response)
		return
	}
	resp.ReturnSuccessResponse(c, response)
}

func GetNewMsgListController(c *gin.Context) {

	getNewMsgListReq := req.GetNewMsgListReq{}
	if err := c.ShouldBindQuery(&getNewMsgListReq); err != nil { //ShouldBind()会自动推导
		resp.ErrorResponse(c, "参数错误")
		c.Abort()
		return
	}
	response, err := service.GetNewMsgService(getNewMsgListReq.MsgId, getNewMsgListReq.RoomId)
	if err != nil {
		c.Abort()
		resp.ReturnErrorResponse(c, response)
		return
	}
	resp.ReturnSuccessResponse(c, response)
}
