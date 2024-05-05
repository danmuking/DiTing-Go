package resp

type PageResp struct {
	Cursor *string `json:"cursor" form:"cursor"`
	IsLast bool    `json:"isLast" form:"is_last"`
	Data   any     `json:"data" form:"data"`
}
