package req

type GetGroupMemberListReq struct {
	// 房间ID
	ID       int64   `form:"id" binding:"required"`
	Cursor   *string `form:"cursor" binding:"required"`
	PageSize int     `form:"page_size" binding:"required"`
}
