package req

// IsFriendReq 是否是好友请求
type IsFriendReq struct {
	FriendUid int64 `uri:"friendUid" binding:"required"`
}
