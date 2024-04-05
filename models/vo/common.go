package vo

type PageListResponse struct {
	List  interface{} `json:"dataList"`
	Total int         `json:"total"`
}
