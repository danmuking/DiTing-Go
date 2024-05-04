package dto

type ContactDto struct {
	// 会话ID
	ID int64 `json:"id"`
	// 房间ID
	RoomID int64 `json:"roomId"`
	// 头像
	Avatar string `json:"avatar"`
	// 会话名称
	Name string `json:"name"`
	// 最后一条消息内容
	LastMsg string `json:"lastMsg"`
	// 最后一条消息时间 时间戳格式
	LastTime int64 `json:"lastTime"`
	// 未读消息数
	UnreadCount int32 `json:"unreadCount"`
}
