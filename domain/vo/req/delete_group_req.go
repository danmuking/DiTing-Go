package req

type DeleteGroupReq struct {
	ID int64 `uri:"id" binding:"required"`
}
