package resp

type UserApplyResp struct {
	ID     int64  `json:"ID"`     // 用户ID
	Name   string `json:"name"`   // 用户昵称
	Avatar string `json:"avatar"` // 用户头像
	Status int32  `json:"status"` // 使用状态 1.待审批 2.已接受
}
