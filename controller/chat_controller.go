package controller

//
//import (
//	"DiTing-Go/domain/vo/req"
//	"DiTing-Go/pkg/domain/vo/resp"
//	"DiTing-Go/service"
//	"github.com/gin-gonic/gin"
//)
//
//func SendMessageController(c *gin.Context) {
//	uid := c.GetInt64("uid")
//	messageReq := req.MessageReq{}
//	if err := c.ShouldBind(&messageReq); err != nil { //ShouldBind()会自动推导
//		resp.ErrorResponse(c, "参数错误")
//		c.Abort()
//		return
//	}
//	response, err := service.SendTextMsgService(uid, messageReq)
//	if err != nil {
//		c.Abort()
//		resp.ReturnErrorResponse(c, response)
//		return
//	}
//	resp.ReturnSuccessResponse(c, response)
//}
