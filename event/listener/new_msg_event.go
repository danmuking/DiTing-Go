package listener

import (
	"DiTing-Go/dal/model"
	query "DiTing-Go/dal/query"
	"DiTing-Go/domain/enum"
	"DiTing-Go/global"
	"DiTing-Go/websocket/service"
	"context"
	"log"
	"time"
)

func init() {
	err := global.Bus.Subscribe(enum.NewMessageEvent, NewMsgEvent)
	if err != nil {
		log.Println("订阅事件失败", err.Error())
	}
	err = global.Bus.Subscribe(enum.NewMessageEvent, UpdateContactEvent)
	if err != nil {
		log.Println("订阅事件失败", err.Error())
	}
}

func UpdateContactEvent(msg model.Message) {
	// 更新会话表
	ctx := context.Background()
	room := global.Query.Room
	roomQ := room.WithContext(ctx)
	roomR, err := roomQ.Where(room.ID.Eq(msg.RoomID)).First()
	if err != nil {
		global.Logger.Errorf("查询房间失败 %s", err)
		return
	}
	var uids []int64
	if roomR.Type == enum.PERSONAL {
		roomFriend := global.Query.RoomFriend
		roomFriendQ := roomFriend.WithContext(ctx)
		roomFriendR, err := roomFriendQ.Where(roomFriend.RoomID.Eq(roomR.ID)).First()
		if err != nil {
			global.Logger.Errorf("查询好友房间失败 %s", err)
			return
		}
		uids = []int64{roomFriendR.Uid1, roomFriendR.Uid2}
	}
	//TODO:群聊
	//更新会话表
	update := model.Contact{
		LastMsgID:  msg.ID,
		UpdateTime: time.Now(),
		ActiveTime: time.Now(),
	}
	contact := global.Query.Contact
	contactQ := contact.WithContext(ctx)
	_, err = contactQ.Where(contact.UID.In(uids...), contact.RoomID.Eq(msg.RoomID)).Updates(&update)
	if err != nil {
		global.Logger.Errorf("更新会话失败 %s", err)
		return
	}
}

// NewMsgEvent 新消息事件
func NewMsgEvent(msg model.Message) {
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
