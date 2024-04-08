package listener

import (
	"DiTing-Go/dal/model"
	global2 "DiTing-Go/websocket/global"

	//"DiTing-Go/dal/model"
	"DiTing-Go/global"
	//global2 "DiTing-Go/websocket/global"
	//"fmt"
	"log"
)

func Init() {
	if err := global.Bus.SubscribeAsync("FriendApplyEvent", FriendApplyEvent, false); err != nil {
		log.Fatalln("订阅事件失败", err.Error())
	}
	//if err := global.Bus.Subscribe("calculator", calculator); err != nil {
	//	log.Fatalln("订阅事件失败", err.Error())
	//}
}

// FriendApplyEvent 好友申请事件
func FriendApplyEvent(apply model.UserApply) {
	msg := global2.Msg{
		Uid: apply.TargetID,
	}
	// 发送新消息事件
	global.Bus.Publish("NewMsgEvent", msg)
}

//func calculator(a int, b int) {
//	fmt.Printf("%d\n", a+b)
//}
