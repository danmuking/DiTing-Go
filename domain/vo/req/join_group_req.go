package req

type JoinGroupReq struct {
	ID int64 `json:"id" binding:"required"`
}
