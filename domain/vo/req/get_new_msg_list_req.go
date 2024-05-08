package req

type GetNewMsgListReq struct {
	// 房间ID
	RoomId int64 `json:"roomId" form:"roomId" binding:"required"`
	// 消息ID
	MsgId int64 `json:"msgId" form:"msgId" binding:"required"`
}
