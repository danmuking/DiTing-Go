package listener

import (
	"DiTing-Go/dal/model"
	"DiTing-Go/domain/enum"
	"DiTing-Go/global"
	"DiTing-Go/websocket/service"
	"log"
)

func init() {
	if err := global.Bus.Subscribe(enum.FriendApplyEvent, FriendApplyEvent); err != nil {
		log.Println("订阅事件失败", err.Error())
	}
}

// FriendApplyEvent 好友申请事件
func FriendApplyEvent(apply model.UserApply) {
	// 发送新消息事件
	service.Send(apply.TargetID)
}
