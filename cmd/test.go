package main

import (
	"context"
	"fmt"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	goredislib "github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
	"os"
	"strconv"
	"time"
)

func main() {
	//client := redis.NewClient(&redis.Options{
	//	Addr:     "150.158.151.30:14490",
	//	Password: "550210817@", // 密码
	//	DB:       0,            // 数据库
	//	PoolSize: 20,           // 连接池大小
	//})
	client := goredislib.NewClient(&goredislib.Options{
		Addr:     "150.158.151.30:14490",
		Password: "550210817@", // 密码
		DB:       0,            // 数据库
		PoolSize: 20,           // 连接池大小
	})
	pool := goredis.NewPool(client)
	rs := redsync.New(pool)

	mutex := rs.NewMutex("test-redsync")
	ctx := context.Background()

	if err := mutex.LockContext(ctx); err != nil {
		panic(err)
	}

	if _, err := mutex.UnlockContext(ctx); err != nil {
		panic(err)
	}
}

func producerR() {
	p, _ := rocketmq.NewProducer(
		// 设置  nameSrvAddr
		// nameSrvAddr 是 Topic 路由注册中心
		producer.WithNameServer([]string{"150.158.151.30:12999"}),
		// 指定发送失败时的重试时间
		producer.WithRetry(2),
		// 设置 Group
		producer.WithGroupName("testGroup"),
	)
	// 开始连接
	err := p.Start()
	if err != nil {
		fmt.Printf("start producer error: %s", err.Error())
		os.Exit(1)
	}

	// 设置节点名称
	topic := "Topic-test"
	// 循坏发送信息 (同步发送)
	for i := 0; i < 10; i++ {
		msg := &primitive.Message{
			Topic: topic,
			Body:  []byte("Hello RocketMQ Go Client" + strconv.Itoa(i)),
		}
		// 发送信息
		res, err := p.SendSync(context.Background(), msg)
		if err != nil {
			fmt.Printf("send message error:%s\n", err)
		} else {
			fmt.Printf("send message success: result=%s\n", res.String())
		}
	}
	// 关闭生产者
	err = p.Shutdown()
	if err != nil {
		fmt.Printf("shutdown producer error:%s", err.Error())
	}
}
func consumerR() {
	// 设置推送消费者
	c, _ := rocketmq.NewPushConsumer(
		//消费组
		consumer.WithGroupName("testGroup"),
		// namesrv地址
		consumer.WithNameServer([]string{"150.158.151.30:12999"}),
	)
	// 必须先在 开始前
	err := c.Subscribe("Topic-test", consumer.MessageSelector{}, func(ctx context.Context, ext ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
		for i := range ext {
			fmt.Printf("subscribe callback:%v \n", ext[i])
		}
		return consumer.ConsumeSuccess, nil
	})
	if err != nil {
		fmt.Println(err.Error())
	}
	err = c.Start()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}
	time.Sleep(time.Hour)
	err = c.Shutdown()
	if err != nil {
		fmt.Printf("shutdown Consumer error:%s", err.Error())
	}
}
