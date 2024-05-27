package req

type DeleteGroupReq struct {
	RoomId int64 `json:"roomId" binding:"required"`
}
