package req

type UserInfoBatchReqItem struct {
	Uid            int64 `json:"uid"`
	LastModifyTime int64 `json:"lastModifyTime"`
}
type GetUserInfoBatchReq struct {
	List []UserInfoBatchReqItem `json:"list" bind:"required"`
}
