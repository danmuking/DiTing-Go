package req

type GetNewContentListReq struct {
	Timestamp int64 `json:"timestamp" form:"timestamp" binding:"required"`
}
