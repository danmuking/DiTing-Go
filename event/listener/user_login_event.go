package listener

import (
	"DiTing-Go/domain/enum"
	"DiTing-Go/global"
	"context"
	"fmt"
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
		consumer.WithGroupName(enum.UserLoginTopic+"-test"),
		// namesrv地址
		consumer.WithNameServer([]string{host}),
	)
	go test(rocketConsumer)

}

func test(rocketConsumer rocketmq.PushConsumer) {
	// 必须先在 开始前
	err := rocketConsumer.Subscribe(enum.UserLoginTopic, consumer.MessageSelector{}, func(ctx context.Context, ext ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
		for i := range ext {
			fmt.Printf("subscribe callback:%v \n", ext[i])
		}
		return consumer.ConsumeSuccess, nil
	})
	if err != nil {
		global.Logger.Panicf("subscribe error: %s", err.Error())
	}
	err = rocketConsumer.Start()
	if err != nil {
		global.Logger.Panicf("start consumer error: %s", err.Error())
	}
}
