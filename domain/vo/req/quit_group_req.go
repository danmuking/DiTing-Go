package req

type QuitGroupReq struct {
	// 房间di
	RoomId int64 `json:"roomId" binding:"required"`
}
