package resp

type UserApplyResp struct {
	ApplyId int64  `json:"applyId"` // 申请ID
	Uid     int64  `json:"uid"`     // 用户ID
	Msg     string `json:"msg"`     // 申请信息
	Status  int32  `json:"status"`  // 使用状态 1.待审批 2.已接受
}
