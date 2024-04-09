package listener

import (
	"DiTing-Go/dal/model"
	"DiTing-Go/global"
	global2 "DiTing-Go/websocket/global"
	"DiTing-Go/websocket/service"
	"log"
)

func init() {
	//if err := global.Bus.SubscribeAsync("FriendApplyEvent", FriendApplyEvent, false); err != nil {
	if err := global.Bus.Subscribe("main:FriendApplyEvent", FriendApplyEvent); err != nil {
		log.Fatalln("订阅事件失败", err.Error())
	}
	//if err := global.Bus.Subscribe("test", test); err != nil {
	//	log.Fatalln("订阅事件失败", err.Error())
	//}
}

// FriendApplyEvent 好友申请事件
func FriendApplyEvent(apply model.UserApply) {
	msg := global2.Msg{
		Uid: apply.TargetID,
	}
	// 发送新消息事件
	service.Send(&msg)
}
