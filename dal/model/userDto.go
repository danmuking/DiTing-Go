package model

type Uid struct {
	Uid int64 `json:"uid"`
}

type UserApplyDto struct {
	Uid int64  `json:"uid"`
	Msg string `json:"msg"`
}
