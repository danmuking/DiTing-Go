package resp

type PageListResp struct {
	List  interface{} `json:"dataList"`
	Total int         `json:"total"`
}
