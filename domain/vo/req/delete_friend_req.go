package req

type DeleteFriendReq struct {
	Uid int64 `json:"uid" binding:"required"`
}
