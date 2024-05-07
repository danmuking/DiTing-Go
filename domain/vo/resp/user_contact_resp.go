package resp

type UserContactResp struct {
	Uid          int64 `json:"uid"`          // 用户ID
	ActiveStatus int   `json:"activeStatus"` // 用户状态
	LastOptTime  int64 `json:"lastOptTime"`  // 最后操作时间
}
