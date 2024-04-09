package listener

import (
	"DiTing-Go/dal/model"
	query "DiTing-Go/dal/query"
	"DiTing-Go/domain/enum"
	"DiTing-Go/global"
	"DiTing-Go/websocket/service"
	"context"
	"log"
)

func init() {
	err := global.Bus.SubscribeAsync(enum.NewMessageEvent, NewMsgEvent, false)
	if err != nil {
		log.Fatalln("订阅事件失败", err.Error())
	}
}

// NewMsgEvent 新消息事件
func NewMsgEvent(msg model.Message) {
	//TODO:修改会话表
	// 向房间中的所有用户发送消息，包括自己
	roomQ := global.Query.WithContext(context.Background()).Room
	room, _ := roomQ.Where(query.Room.ID.Eq(msg.RoomID)).First()
	// 单聊
	if room.Type == enum.PERSONAL {
		roomFriendQ := global.Query.WithContext(context.Background()).RoomFriend
		roomFriend, _ := roomFriendQ.Where(query.RoomFriend.RoomID.Eq(room.ID)).First()
		// 发送新消息事件
		service.Send(roomFriend.Uid1)
		service.Send(roomFriend.Uid2)
	} else if room.Type == enum.GROUP {
		roomGroupQ := global.Query.WithContext(context.Background()).RoomGroup
		roomGroup, _ := roomGroupQ.Where(query.RoomGroup.RoomID.Eq(room.ID)).First()
		//	查询所有群成员
		groupMemberQ := global.Query.WithContext(context.Background()).GroupMember
		groupMembers, _ := groupMemberQ.Where(query.GroupMember.GroupID.Eq(roomGroup.ID)).Find()
		// 发送新消息事件
		for _, groupMember := range groupMembers {
			service.Send(groupMember.UID)
		}
	}

}
