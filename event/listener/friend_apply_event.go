package listener

import (
	"DiTing-Go/dal/model"
	"DiTing-Go/domain/enum"
	"DiTing-Go/global"
	"DiTing-Go/utils/jsonUtils"
	"DiTing-Go/websocket/service"
	"context"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/spf13/viper"
)

func init() {
	host := viper.GetString("rocketmq.host")
	// 设置推送消费者
	rocketConsumer, _ := rocketmq.NewPushConsumer(
		//消费组
		consumer.WithGroupName(enum.FriendApplyTopic),
		// namesrv地址
		consumer.WithNameServer([]string{host}),
	)
	err := rocketConsumer.Subscribe(enum.FriendApplyTopic, consumer.MessageSelector{}, friendApplyEvent)
	if err != nil {
		global.Logger.Panicf("subscribe error: %s", err.Error())
	}
	err = rocketConsumer.Start()
	if err != nil {
		global.Logger.Panicf("start consumer error: %s", err.Error())
	}
}

// FriendApplyEvent 好友申请事件
func friendApplyEvent(ctx context.Context, ext ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
	for i := range ext {
		// 解码
		userApplyR := model.UserApply{}
		if err := jsonUtils.UnmarshalMsg(&userApplyR, ext[i]); err != nil {
			global.Logger.Errorf("jsonUtils unmarshal error: %s", err.Error())
			return consumer.ConsumeRetryLater, nil
		}

		if err := friendApply(userApplyR); err != nil {
			global.Logger.Errorf("friendApply error: %s", err.Error())
			return consumer.ConsumeRetryLater, nil
		}
	}
	return consumer.ConsumeSuccess, nil
}

func friendApply(apply model.UserApply) error {
	// 发送新消息事件
	service.Send(apply.TargetID)
	return nil
}
