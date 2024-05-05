package req

type GetMessageListReq struct {
	RoomId   int64   `json:"roomId" form:"roomId" binding:"required"`
	Cursor   *string `json:"cursor" form:"cursor" binding:"required"`
	PageSize int     `json:"pageSize" form:"pageSize" binding:"required"`
}
