package req

type MessageBody struct {
	Content    string `json:"content" form:"content" binding:"required"`
	ReplyMsgId int64  `json:"replyMsgId" form:"replyMsgId"`
}
type MessageReq struct {
	RoomId  int64       `json:"roomId" form:"roomId" binding:"required"`
	MsgType int32       `json:"msgType" form:"msgType" binding:"required"`
	Body    MessageBody `json:"body" form:"body" binding:"required"`
}
