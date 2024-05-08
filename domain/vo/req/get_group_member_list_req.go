package req

type GetGroupMemberListReq struct {
	// 房间ID
	RoomId   int64   `form:"roomId" binding:"required"`
	Cursor   *string `form:"cursor"`
	PageSize int     `form:"pageSize" binding:"required"`
}
