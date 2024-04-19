package listener

import (
	"DiTing-Go/dal/model"
	query "DiTing-Go/dal/query"
	"DiTing-Go/domain/enum"
	"DiTing-Go/global"
	"DiTing-Go/pkg/utils"
	"DiTing-Go/websocket/service"
	"context"
	"log"
	"strconv"
	"time"
)

func init() {
	err := global.Bus.SubscribeAsync(enum.NewMessageEvent, NewMsgEvent, false)
	if err != nil {
		log.Println("订阅事件失败", err.Error())
	}
	err = global.Bus.SubscribeAsync(enum.NewMessageEvent, UpdateContactEvent, false)
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
	} else if roomR.Type == enum.GROUP {
		// 查询所有群成员
		roomGroup := global.Query.RoomGroup
		roomGroupQ := roomGroup.WithContext(ctx)
		roomGroupR, err := roomGroupQ.Where(roomGroup.RoomID.Eq(roomR.ID)).First()
		if err != nil {
			global.Logger.Errorf("查询群聊失败 %s", err)
			return
		}
		groupMember := global.Query.GroupMember
		groupMemberQ := groupMember.WithContext(ctx)
		groupMembers, _ := groupMemberQ.Where(groupMember.GroupID.Eq(roomGroupR.ID)).Find()
		for _, groupMember := range groupMembers {
			uids = append(uids, groupMember.UID)
		}
	}
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
	fun := func() (interface{}, error) {
		return roomQ.Where(query.Room.ID.Eq(msg.RoomID)).First()
	}
	room := model.Room{}
	err := utils.GetData(enum.Room+strconv.FormatInt(msg.RoomID, 10), &room, fun)
	if err != nil {
		global.Logger.Errorf("查询房间失败 %s", err)
		return
	}
	// 单聊
	if room.Type == enum.PERSONAL {
		roomFriendQ := global.Query.WithContext(context.Background()).RoomFriend
		roomFriendR := model.RoomFriend{}
		fun = func() (interface{}, error) {
			return roomFriendQ.Where(query.RoomFriend.RoomID.Eq(room.ID)).First()
		}
		err = utils.GetData(enum.RoomFriend+strconv.FormatInt(room.ID, 10), &roomFriendR, fun)
		if err != nil {
			global.Logger.Errorf("查询好友房间失败 %s", err)
			return
		}
		// 发送新消息事件
		service.Send(roomFriendR.Uid1)
		service.Send(roomFriendR.Uid2)
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
	return

}
