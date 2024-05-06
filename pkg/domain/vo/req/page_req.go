package req

type PageReq struct {
	Cursor   *string `json:"cursor" form:"cursor"`
	PageSize int     `json:"page_size" form:"page_size" binding:"required"`
}
