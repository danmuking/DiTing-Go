package dto

import (
	"time"
)

type ContactDto struct {
	ID int64 `json:"ID"`
	// 头像
	Avatar string `json:"avatar"`
	// 会话名称
	Name string `json:"name"`
	// 最后一条消息内容
	LastMsg string `json:"last_msg"`
	// 最后一条消息时间
	LastTime time.Time `json:"last_time"`
	// 未读消息数
	UnreadCount int32 `json:"unread_count"`
}
