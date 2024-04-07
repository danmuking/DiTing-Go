package resp

type UserContactResp struct {
	ID     int64  `json:"ID"`     // 用户ID
	Name   string `json:"name"`   // 用户昵称
	Avatar string `json:"avatar"` // 用户头像
}
