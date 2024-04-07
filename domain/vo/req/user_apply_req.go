package req

type UserApplyReq struct {
	Uid int64  `json:"uid"`
	Msg string `json:"msg"`
}
