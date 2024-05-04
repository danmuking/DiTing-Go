package resp

type UserLoginResp struct {
	Token  string `json:"token"`
	Uid    int64  `json:"uid"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}
