package listener

import (
	"DiTing-Go/dal/model"
	"DiTing-Go/domain/enum"
	"DiTing-Go/global"
	"DiTing-Go/pkg/utils"
	"DiTing-Go/service"
	"context"
	"log"
	"sort"
	"strconv"
	"time"
)

func init() {
	err := global.Bus.SubscribeAsync(enum.FriendNewEvent, FriendNewEvent, false)
	if err != nil {
		log.Println("订阅事件失败", err.Error())
	}
}

// FriendNewEvent 新好友事件
func FriendNewEvent(friend model.UserFriend) {
	//println(FriendNewEvent)
	ctx := context.Background()
	q := global.Query
	tx := q.Begin()
	roomQ := tx.WithContext(ctx).Room
	roomFriendQ := tx.WithContext(ctx).RoomFriend
	contactQ := tx.WithContext(ctx).Contact

	room := model.Room{
		Type:    enum.PERSONAL,
		HotFlag: enum.NORMAL,
		ExtJSON: "{}",
	}

	// 创建房间表
	if err := roomQ.Create(&room); err != nil {
		if err := tx.Rollback(); err != nil {
			log.Println("事务回滚失败", err.Error())
			return
		}
		log.Println("创建房间失败", err.Error())
		return
	}
	// 排序，uid小的在前
	uids := utils.Int64Slice{friend.UID, friend.FriendUID}
	sort.Sort(uids)

	// 创建私聊表
	roomFriend := model.RoomFriend{
		RoomID:  room.ID,
		Uid1:    uids[0],
		Uid2:    uids[1],
		RoomKey: strconv.FormatInt(uids[0], 10) + "," + strconv.FormatInt(uids[1], 10),
	}
	if err := roomFriendQ.Create(&roomFriend); err != nil {
		if err := tx.Rollback(); err != nil {
			log.Println("事务回滚失败", err.Error())
			return
		}
		log.Println("创建房间失败", err.Error())
		return
	}

	// 自动发送一条消息
	newMsg := model.Message{
		RoomID:  room.ID,
		FromUID: friend.UID,
		Content: "你们已经是好友了，开始聊天吧",
		// TODO: 抽取为常量
		Status: 0,
		Type:   enum.TextMessage,
		Extra:  "{}",
	}
	if err := service.SendTextMsg(&newMsg); err != nil {
		log.Println("发送消息失败", err.Error())
		return
	}

	//创建会话表
	s, _ := time.ParseDuration("-1s")
	if err := contactQ.Create(&model.Contact{
		UID:        friend.UID,
		RoomID:     room.ID,
		LastMsgID:  newMsg.ID,
		ReadTime:   time.Now(),
		ActiveTime: time.Now(),
	}); err != nil {
		if err := tx.Rollback(); err != nil {
			log.Println("事务回滚失败", err.Error())
			return
		}
		log.Println("创建会话失败", err.Error())
		return
	}
	if err := contactQ.Create(&model.Contact{
		UID:       friend.FriendUID,
		RoomID:    room.ID,
		LastMsgID: newMsg.ID,
		// 读到时间设为1秒前
		ReadTime:   time.Now().Add(s),
		ActiveTime: time.Now(),
	}); err != nil {
		if err := tx.Rollback(); err != nil {
			log.Println("事务回滚失败", err.Error())
			return
		}
		log.Println("创建会话失败", err.Error())
		return
	}
	// 提交
	if err := tx.Commit(); err != nil {
		log.Println("事务提交失败", err.Error())
		return
	}
	// 发送新消息事件
	global.Bus.Publish(enum.NewMessageEvent, newMsg)

}
