package resp

type PageResp struct {
	Cursor *string `json:"cursor" form:"cursor"`
	IsLast bool    `json:"is_last" form:"is_last"`
	Data   any     `json:"data" form:"data"`
}
