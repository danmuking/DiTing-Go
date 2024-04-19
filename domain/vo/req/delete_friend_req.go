package req

type DeleteFriendReq struct {
	Uid int64 `uri:"uid" binding:"required"`
}
