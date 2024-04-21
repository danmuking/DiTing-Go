package listener

import (
	"DiTing-Go/dal/model"
	query "DiTing-Go/dal/query"
	"DiTing-Go/domain/enum"
	"DiTing-Go/global"
	"DiTing-Go/pkg/utils"
	"DiTing-Go/websocket/service"
	"context"
	"encoding/json"
	"fmt"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/spf13/viper"
	"time"
)

func init() {
	host := viper.GetString("rocketmq.host")
	// 设置推送消费者
	rocketSendMsgConsumer, _ := rocketmq.NewPushConsumer(
		//消费组
		consumer.WithGroupName(enum.NewMessageTopic+"-send-message"),
		// namesrv地址
		consumer.WithNameServer([]string{host}),
	)
	err := rocketSendMsgConsumer.Subscribe(enum.NewMessageTopic, consumer.MessageSelector{}, UpdateContactEvent)
	if err != nil {
		global.Logger.Panicf("subscribe error: %s", err.Error())
	}
	err = rocketSendMsgConsumer.Start()
	if err != nil {
		global.Logger.Panicf("start consumer error: %s", err.Error())
	}

	// 设置推送消费者
	rocketUpdateContactConsumer, _ := rocketmq.NewPushConsumer(
		//消费组
		consumer.WithGroupName(enum.NewMessageTopic+"-update-contact"),
		// namesrv地址
		consumer.WithNameServer([]string{host}),
	)
	err = rocketUpdateContactConsumer.Subscribe(enum.NewMessageTopic, consumer.MessageSelector{}, SendMsgEvent)
	if err != nil {
		global.Logger.Panicf("subscribe error: %s", err.Error())
	}
	err = rocketUpdateContactConsumer.Start()
	if err != nil {
		global.Logger.Panicf("start consumer error: %s", err.Error())
	}
}

func UpdateContactEvent(ctx context.Context, ext ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
	for i := range ext {
		// 解码
		msg := model.Message{}
		msgByte := ext[i].Message.Body
		err := json.Unmarshal(msgByte, &msg)
		if err != nil {
			global.Logger.Errorf("json unmarshal error: %s", err.Error())
			return consumer.ConsumeRetryLater, nil
		}
		err = updateContact(msg)
		if err != nil {
			global.Logger.Errorf("更新会话失败 %s", err)
			return consumer.ConsumeRetryLater, nil
		}
	}
	return consumer.ConsumeSuccess, nil
}
func updateContact(msg model.Message) error {
	// 更新会话表
	ctx := context.Background()
	room := global.Query.Room
	roomQ := room.WithContext(ctx)
	roomR, err := roomQ.Where(room.ID.Eq(msg.RoomID)).First()
	if err != nil {
		global.Logger.Errorf("查询房间失败 %s", err)
		return err
	}
	var uids []int64
	if roomR.Type == enum.PERSONAL {
		roomFriend := global.Query.RoomFriend
		roomFriendQ := roomFriend.WithContext(ctx)
		roomFriendR, err := roomFriendQ.Where(roomFriend.RoomID.Eq(roomR.ID)).First()
		if err != nil {
			global.Logger.Errorf("查询好友房间失败 %s", err)
			return err
		}
		uids = []int64{roomFriendR.Uid1, roomFriendR.Uid2}
	} else if roomR.Type == enum.GROUP {
		// 查询所有群成员
		roomGroup := global.Query.RoomGroup
		roomGroupQ := roomGroup.WithContext(ctx)
		roomGroupR, err := roomGroupQ.Where(roomGroup.RoomID.Eq(roomR.ID)).First()
		if err != nil {
			global.Logger.Errorf("查询群聊失败 %s", err)
			return err
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
		return err
	}
	return nil
}

// SendMsgEvent 新消息事件
func SendMsgEvent(ctx context.Context, ext ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
	for i := range ext {
		// 解码
		msg := model.Message{}
		msgByte := ext[i].Message.Body
		err := json.Unmarshal(msgByte, &msg)
		if err != nil {
			global.Logger.Errorf("json unmarshal error: %s", err.Error())
			return consumer.ConsumeRetryLater, nil
		}
		err = sendMsg(msg)
		if err != nil {
			global.Logger.Errorf("发送消息失败 %s", err)
			return consumer.ConsumeRetryLater, nil
		}
	}
	return consumer.ConsumeSuccess, nil
}

// sendMsg 发送消息
func sendMsg(msg model.Message) error {
	// 向房间中的所有用户发送消息，包括自己
	roomQ := global.Query.WithContext(context.Background()).Room
	fun := func() (interface{}, error) {
		return roomQ.Where(query.Room.ID.Eq(msg.RoomID)).First()
	}
	room := model.Room{}
	key := fmt.Sprintf(enum.RoomCacheByID, msg.RoomID)
	err := utils.GetData(key, &room, fun)
	if err != nil {
		global.Logger.Errorf("查询房间失败 %s", err)
		return err
	}
	// 单聊
	if room.Type == enum.PERSONAL {
		roomFriendQ := global.Query.WithContext(context.Background()).RoomFriend
		roomFriendR := model.RoomFriend{}
		fun = func() (interface{}, error) {
			return roomFriendQ.Where(query.RoomFriend.RoomID.Eq(room.ID)).First()
		}
		key := fmt.Sprintf(enum.RoomFriendCacheByRoomID, room.ID)
		err = utils.GetData(key, &roomFriendR, fun)
		if err != nil {
			global.Logger.Errorf("查询好友房间失败 %s", err)
			return err
		}
		// 发送新消息事件
		service.Send(roomFriendR.Uid1)
		service.Send(roomFriendR.Uid2)
		//	TODO:群聊
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
	return nil

}
