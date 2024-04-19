package req

type AgreeFriendReq struct {
	Uid int64 `json:"uid" binding:"required"`
}
