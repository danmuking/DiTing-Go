package dto

import (
	"gorm.io/gen/field"
	"time"
)

type ContactDto struct {
	ID int64 `json:"ID"`
	// 头像
	Avatar field.String `json:"avatar"`
	// 会话名称
	Name string `json:"name"`
	// 最后一条消息内容
	LastMsg string `json:"last_msg"`
	// 最后一条消息时间
	LastTime time.Time `json:"last_time"`
}
