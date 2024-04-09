package listener

import (
	"DiTing-Go/domain/enum"
	"DiTing-Go/global"
	global2 "DiTing-Go/websocket/global"
	"DiTing-Go/websocket/service"
	"log"
)

func init() {
	err := global.Bus.SubscribeAsync(enum.NewMessageEvent, NewMsgEvent, false)
	if err != nil {
		log.Fatalln("订阅事件失败", err.Error())
	}
}

// NewMsgEvent 新消息事件
func NewMsgEvent(msg global2.Msg) {
	service.Send(&msg)
}
