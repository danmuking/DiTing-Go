package req

type PageReq struct {
	Cursor   *string `json:"cursor" form:"cursor"`
	PageSize int     `json:"pageSize" form:"pageSize" binding:"required"`
}
