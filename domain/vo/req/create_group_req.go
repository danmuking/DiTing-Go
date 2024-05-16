package req

type CreateGroupReq struct {
	UidList []int64 `json:"uidList" binding:"required"`
}
