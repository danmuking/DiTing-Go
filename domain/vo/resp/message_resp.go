package resp

import "time"

type MessageResp struct {
	ID         int64     `json:"id"`
	Content    string    `json:"content"`
	ReplyMsgID int64     `json:"reply_msg_id"`
	GapCount   int32     `json:"gap_count"`
	Type       int32     `json:"type"`
	Extra      string    `json:"extra"`
	CreateTime time.Time `json:"create_time"`
	// 发送者信息
	UserName   string `json:"user_name"`
	UserAvatar string `json:"user_avatar"`
}
