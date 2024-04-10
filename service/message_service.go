package service

import (
	"DiTing-Go/dal/model"
	"DiTing-Go/domain/enum"
	"DiTing-Go/global"
	"DiTing-Go/pkg/resp"
	"bytes"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"io"
	"log"
)

// SendTextMsgService 发送文本消息
func SendTextMsgService(c *gin.Context) {
	// TODO：封装dto 状态不需要前端传递参数
	uid := c.GetInt64("uid")
	msg := model.Message{}
	data, err := c.GetRawData()
	if err != nil { //ShouldBind()会自动推导
		resp.ErrorResponse(c, "参数错误")
		c.Abort()
		return
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(data))
	if err := json.Unmarshal(data, &msg); err != nil {
		resp.ErrorResponse(c, "参数错误")
		c.Abort()
		log.Fatalln("参数错误", err.Error())
		return
	}
	msg.Type = enum.TextMessage
	msg.FromUID = uid
	if msg.Extra == "" {
		msg.Extra = "{}"
	}

	// 发送消息
	if err := SendTextMsg(&msg); err != nil {
		resp.ErrorResponse(c, "消息发送失败")
		c.Abort()
		return
	}

	// 发送新消息事件
	global.Bus.Publish(enum.NewMessageEvent, msg)

	// 返回成功
	resp.SuccessResponseWithMsg(c, "success")
	c.Abort()
	return
}

func SendTextMsg(msg *model.Message) error {
	ctx := context.Background()
	msgQ := global.Query.WithContext(ctx).Message
	if err := msgQ.Create(msg); err != nil {
		log.Fatalln("消息发送失败", err.Error())
		return err
	}
	return nil
}
