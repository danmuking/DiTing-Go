package listener

import (
	"DiTing-Go/dal/model"
	"DiTing-Go/domain/enum"
	"DiTing-Go/global"
	pkgEnum "DiTing-Go/pkg/domain/enum"
	"DiTing-Go/utils/redisCache"
	"context"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/goccy/go-json"
	"github.com/spf13/viper"
)

func init() {
	host := viper.GetString("rocketmq.host")
	// 设置推送消费者
	rocketConsumer, _ := rocketmq.NewPushConsumer(
		//消费组
		consumer.WithGroupName(enum.RoomGroupDeleteTopic),
		// namesrv地址
		consumer.WithNameServer([]string{host}),
	)
	err := rocketConsumer.Subscribe(enum.RoomGroupDeleteTopic, consumer.MessageSelector{}, roomGroupDeleteEvent)
	if err != nil {
		global.Logger.Panicf("subscribe error: %s", err.Error())
	}
	err = rocketConsumer.Start()
	if err != nil {
		global.Logger.Panicf("start consumer error: %s", err.Error())
	}
}

func roomGroupDeleteEvent(ctx context.Context, ext ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
	for i := range ext {
		// 解码
		roomGroup := model.RoomGroup{}
		roomGroupMsgByte := ext[i].Message.Body
		err := json.Unmarshal(roomGroupMsgByte, &roomGroup)
		if err != nil {
			global.Logger.Errorf("json unmarshal error: %s", err.Error())
			return consumer.ConsumeRetryLater, nil
		}
		err = roomGroupDelete(roomGroup)
		if err != nil {
			global.Logger.Errorf("friendNew error: %s", err.Error())
			return consumer.ConsumeRetryLater, nil
		}

	}
	return consumer.ConsumeSuccess, nil
}

func roomGroupDelete(roomGroupR model.RoomGroup) error {
	ctx := context.Background()
	tx := global.Query.Begin()

	// 获取群聊成员
	groupMember := global.Query.GroupMember
	groupMemberTx := tx.GroupMember.WithContext(ctx)
	groupMemberList, err := groupMemberTx.Where(groupMember.GroupID.Eq(roomGroupR.ID)).Find()
	if err != nil {
		if err := tx.Rollback(); err != nil {
			global.Logger.Errorf("事务回滚失败 %s", err.Error())
		}
		global.Logger.Errorf("查询群组成员表失败 %s", err.Error())
		return err
	}

	// 删除所有成员的会话表
	var memberUids []int64
	for _, groupMember := range groupMemberList {
		memberUids = append(memberUids, groupMember.UID)
	}
	contact := global.Query.Contact
	contactTx := tx.Contact.WithContext(ctx)
	// 查询要删除的会话
	contactRList, err := contactTx.Where(contact.UID.In(memberUids...), contact.RoomID.Eq(roomGroupR.RoomID)).Find()
	contactIdList := make([]int64, 0)
	for _, contactR := range contactRList {
		contactIdList = append(contactIdList, contactR.ID)
	}
	if _, err := contactTx.Where(contact.ID.In(contactIdList...)).Delete(); err != nil {
		if err := tx.Rollback(); err != nil {
			global.Logger.Errorf("事务回滚失败 %s", err.Error())
		}
		global.Logger.Errorf("删除会话表失败 %s", err.Error())
		return err
	}
	// 删除缓存
	for _, contactR := range contactRList {
		redisCache.RemoveContact(*contactR)
	}

	// 删除群组成员表
	if _, err := groupMemberTx.Where(groupMember.GroupID.Eq(roomGroupR.ID)).Delete(); err != nil {
		if err := tx.Rollback(); err != nil {
			global.Logger.Errorf("事务回滚失败 %s", err.Error())
		}
		global.Logger.Errorf("删除群组成员表失败 %s", err.Error())
		return err
	}
	// TODO:删除群组成员缓存

	// 删除消息表
	message := global.Query.Message
	messageTx := tx.Message.WithContext(ctx)
	msg := model.Message{
		DeleteStatus: pkgEnum.DELETED,
	}
	if _, err := messageTx.Where(message.RoomID.Eq(roomGroupR.ID)).Updates(msg); err != nil {
		if err := tx.Rollback(); err != nil {
			global.Logger.Errorf("事务回滚失败 %s", err.Error())
		}
		global.Logger.Errorf("删除消息表失败 %s", err.Error())
		return err
	}
	// TODO：删除消息缓存

	if err := tx.Commit(); err != nil {
		global.Logger.Errorf("事务提交失败 %s", err.Error())
		return err
	}
	return nil
}
