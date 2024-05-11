package resp

type GetUserInfoByNameResp struct {
	// 用户ID
	Uid int64 `json:"uid"`
	// 用户名
	Name string `json:"name"`
	// 头像
	Avatar string `json:"avatar"`
	// 好友状态
	Status int32 `json:"status"`
}
