package controller

import (
	"DiTing-Go/domain/vo/req"
	"DiTing-Go/pkg/domain/vo/resp"
	"DiTing-Go/service"
	"github.com/gin-gonic/gin"
)

func CaptchaController(c *gin.Context) {
	captchaReq := req.CaptchaReq{}
	if err := c.ShouldBind(&captchaReq); err != nil {
		resp.ErrorResponse(c, "参数错误")
		c.Abort()
		return
	}
	response, err := service.CaptchaService(captchaReq)
	if err != nil {
		c.Abort()
		resp.ReturnErrorResponse(c, response)
		return
	}
	resp.ReturnSuccessResponse(c, response)
}
