package model

import (
	"DiTing-Go/dal/model"
	"DiTing-Go/domain/enum"
)

type Message model.Message

func (msg Message) GetContactMsg() string {
	if msg.Type == enum.TextMessageType {
		return msg.Content
	} else if msg.Type == enum.ImgMessageType {
		return "[图片]"
	}
	return msg.Content
}
