package resp

type GetUserInfoBatchResp struct {
	Uid         int64  `json:"uid"`
	Username    string `json:"name"`
	Avatar      string `json:"avatar"`
	NeedRefresh bool   `json:"needRefresh"`
}
