package req

type JoinGroupReq struct {
	RoomId int64 `json:"roomId" binding:"required"`
}
