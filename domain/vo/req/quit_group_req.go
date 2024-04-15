package req

type QuitGroupReq struct {
	// 房间di
	ID int64 `json:"id" binding:"required"`
}
