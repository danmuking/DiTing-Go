package controller

import (
	"DiTing-Go/domain/vo/req"
	cursorUtils "DiTing-Go/pkg/cursor"
	"DiTing-Go/pkg/resp"
	"DiTing-Go/service"
	"github.com/gin-gonic/gin"
)

// ApplyFriendController 添加好友
//
//	@Summary	添加好友
//	@Produce	json
//	@Param		uid	body		int					true	"好友uid"
//	@Param		msg	body		string				true	"验证消息"
//	@Success	200	{object}	resp.ResponseData	"成功"
//	@Failure	500	{object}	resp.ResponseData	"内部错误"
//	@Router		/api/user/add [post]
func ApplyFriendController(c *gin.Context) {
	uid := c.GetInt64("uid")
	applyReq := req.UserApplyReq{}
	if err := c.ShouldBind(&applyReq); err != nil {
		resp.ErrorResponse(c, "参数错误")
		c.Abort()
		return
	}
	response, err := service.ApplyFriendService(uid, applyReq)
	if err != nil {
		c.Abort()
		resp.ReturnErrorResponse(c, response)
		return
	}
	resp.ReturnSuccessResponse(c, response)
	return
}

// DeleteFriendController 删除好友
//
//	@Summary	删除好友
//	@Produce	json
//	@Param		uid	body		int					true	"好友uid"
//	@Success	200	{object}	resp.ResponseData	"成功"
//	@Failure	500	{object}	resp.ResponseData	"内部错误"
//	@Router		/api/user/delete/:uid [delete]
func DeleteFriendController(c *gin.Context) {
	uid := c.GetInt64("uid")
	deleteFriendReq := req.DeleteFriendReq{}
	if err := c.ShouldBindUri(&deleteFriendReq); err != nil {
		resp.ErrorResponse(c, "参数错误")
		return
	}
	response, err := service.DeleteFriendService(uid, deleteFriendReq)
	if err != nil {
		c.Abort()
		resp.ReturnErrorResponse(c, response)
		return
	}
	resp.ReturnSuccessResponse(c, response)
}

// AgreeFriendController 同意好友申请
//
//	@Summary	同意好友申请
//	@Produce	json
//	@Param		uid	body		int					true	"好友uid"
//	@Success	200	{object}	resp.ResponseData	"成功"
//	@Failure	500	{object}	resp.ResponseData	"内部错误"
//	@Router		/api/user/agree [put]
func AgreeFriendController(c *gin.Context) {
	uid := c.GetInt64("uid")
	agreeFriendReq := req.AgreeFriendReq{}
	if err := c.ShouldBind(&agreeFriendReq); err != nil { //ShouldBind()会自动推导
		resp.ErrorResponse(c, "参数错误")
		c.Abort()
		return
	}
	response, err := service.AgreeFriendService(uid, agreeFriendReq.Uid)
	if err != nil {
		c.Abort()
		resp.ReturnErrorResponse(c, response)
		return
	}
	resp.ReturnSuccessResponse(c, response)
}

// GetUserApplyController 同意好友申请
//
//	@Summary	同意好友申请
//	@Produce	json
//	@Param		uid	body		int					true	"好友uid"
//	@Success	200	{object}	resp.ResponseData	"成功"
//	@Failure	500	{object}	resp.ResponseData	"内部错误"
//	@Router		/api/user/getApplyList [get]
func GetUserApplyController(c *gin.Context) {
	uid := c.GetInt64("uid")
	pageRequest := cursorUtils.PageReq{}
	if err := c.ShouldBindQuery(&pageRequest); err != nil { //ShouldBind()会自动推导
		resp.ErrorResponse(c, "参数错误")
		return
	}
	response, err := service.GetUserApplyService(uid, pageRequest)
	if err != nil {
		c.Abort()
		resp.ReturnErrorResponse(c, response)
		return
	}

	resp.ReturnSuccessResponse(c, response)
}

// IsFriendController 是否为好友关系
//
//	@Summary	是否为好友关系
//	@Produce	json
//	@Param		uid	body		int					true	"好友uid"
//	@Success	200	{object}	resp.ResponseData	"成功"
//	@Failure	500	{object}	resp.ResponseData	"内部错误"
//	@Router		/api/user/isFriend/:friendUid [get]
func IsFriendController(c *gin.Context) {
	uid := c.GetInt64("uid")
	isFriendReq := req.IsFriendReq{}
	if err := c.ShouldBindUri(&isFriendReq); err != nil {
		resp.ErrorResponse(c, "参数错误")
		return
	}
	response, err := service.IsFriendService(uid, isFriendReq.FriendUid)
	if err != nil {
		c.Abort()
		resp.ReturnErrorResponse(c, response)
		return
	}
	resp.ReturnSuccessResponse(c, response)
}

// UnreadApplyNumController 好友申请未读数量
//
//	@Summary	好友申请未读数量
//	@Success	200			{object}	resp.ResponseData	"成功"
//	@Failure	500			{object}	resp.ResponseData	"内部错误"
//	@Router		/api/user/unreadApplyNum [get]
func UnreadApplyNumController(c *gin.Context) {
	uid := c.GetInt64("uid")

	response, err := service.UnreadApplyNumService(uid)
	if err != nil {
		c.Abort()
		resp.ReturnErrorResponse(c, response)
		return
	}
	resp.ReturnSuccessResponse(c, response)
}
