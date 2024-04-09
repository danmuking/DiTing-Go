package listener

import (
	"DiTing-Go/dal/model"
	"DiTing-Go/domain/enum"
	"DiTing-Go/global"
	"DiTing-Go/pkg/utils"
	global2 "DiTing-Go/websocket/global"
	"context"
	"log"
	"sort"
	"strconv"
)

func init() {
	err := global.Bus.SubscribeAsync(enum.FriendNewEvent, FriendNewEvent, false)
	if err != nil {
		log.Fatalln("订阅事件失败", err.Error())
	}
}

// FriendNewEvent 新好友事件
func FriendNewEvent(friend model.UserFriend) {
	ctx := context.Background()
	q := global.Query
	msg := global2.Msg{
		Uid: friend.FriendUID,
	}
	room := model.Room{
		Type:    enum.GROUP,
		HotFlag: enum.NORMAL,
		ExtJSON: "{}",
	}
	tx := q.Begin()
	roomQ := tx.WithContext(ctx).Room
	// 创建房间表
	err := roomQ.Create(&room)
	if err != nil {
		log.Fatalln("创建房间失败", err.Error())
		tx.Rollback()
	}
	uids := utils.Int64Slice{friend.UID, friend.FriendUID}
	sort.Sort(uids)
	roomFriendQ := tx.WithContext(ctx).RoomFriend
	roomFriend := model.RoomFriend{
		RoomID:  room.ID,
		Uid1:    uids[0],
		Uid2:    uids[1],
		RoomKey: strconv.FormatInt(uids[0], 10) + "," + strconv.FormatInt(uids[1], 10),
	}
	// 创建私聊表
	err = roomFriendQ.Create(&roomFriend)
	if err != nil {
		log.Fatalln("创建房间失败", err.Error())
		tx.Rollback()
	}
	err = tx.Commit()
	if err != nil {
		log.Fatalln("创建房间失败", err.Error())
		return
	}
	//TODO 自动发送一条消息

	// 发送新消息事件
	global.Bus.Publish("NewMsgEvent", msg)
}
