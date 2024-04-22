package jsonUtils

import (
	"DiTing-Go/global"
	"context"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/goccy/go-json"
	"github.com/pkg/errors"
)

// UnmarshalMsg 解析消息队列msg
func UnmarshalMsg(item any, msg *primitive.MessageExt) error {
	byteStr := msg.Message.Body
	err := json.Unmarshal(byteStr, item)
	if err != nil {
		global.Logger.Errorf("jsonUtils unmarshal error: %s", err.Error())
		return errors.New("Business Error")
	}
	return nil
}

// Marshal 编码
func Marshal(item any) ([]byte, error) {
	byteStr, err := json.Marshal(item)
	if err != nil {
		global.Logger.Errorf("json序列化失败 %v", err)
		return nil, err
	}
	return byteStr, nil
}

// SendMsgSync 发送消息
func SendMsgSync(topic string, item any) error {
	ctx := context.Background()
	byteStr, err := Marshal(item)
	if err != nil {
		return err
	}
	msg := &primitive.Message{
		Topic: topic,
		Body:  byteStr,
	}
	_, err = global.RocketProducer.SendSync(ctx, msg)
	return err
}
