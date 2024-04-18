package controller

import (
	"DiTing-Go/domain/vo/req"
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
//	@Router		/api/contact/add [post]
func ApplyFriendController(c *gin.Context) {
	uid := c.GetInt64("uid")
	applyReq := req.UserApplyReq{}
	if err := c.ShouldBind(&applyReq); err != nil {
		resp.ErrorResponse(c, "参数错误")
		c.Abort()
		return
	}
	response := service.ApplyFriendService(uid, applyReq)
	resp.ReturnResponse(c, response)
	return
}
