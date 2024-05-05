package resp

type MsgUser struct {
	Uid      int64  `json:"uid"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
}
type Msg struct {
	ID     int64    `json:"id"`
	RoomId int64    `json:"rooId"`
	Type   int32    `json:"type"`
	Body   TextBody `json:"body"`
}
type TextBody struct {
	Content string `json:"content"`
	Reply   int64  `json:"reply"`
}
type MessageResp struct {
	FromUser MsgUser `json:"fromUser"`
	Message  Msg     `json:"message"`
	SendTime int64   `json:"sendTime"`
}
