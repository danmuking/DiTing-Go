package controller

import (
	"DiTing-Go/domain/vo/req"
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
