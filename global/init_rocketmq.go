package global

import (
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"github.com/spf13/viper"
)

var RocketProducer rocketmq.Producer

func init() {
	host := viper.GetString("rocketmq.host")
	group := viper.GetString("rocketmq.group")
	RocketProducer, _ = rocketmq.NewProducer(
		// 设置  nameSrvAddr
		// nameSrvAddr 是 Topic 路由注册中心
		producer.WithNameServer([]string{host}),
		// 指定发送失败时的重试时间
		producer.WithRetry(3),
		// 设置 Group
		producer.WithGroupName(group),
	)
	// 开始连接
	err := RocketProducer.Start()
	if err != nil {
		Logger.Panicf("start producer error: %s", err.Error())
	}
}
