package dto

type MessageBaseDto struct {
	// 下载地址
	Url string `json:"url"`
	// 文件大小
	Size int64 `json:"size"`
	// 文件名
	Name string `json:"name"`
}

type ImgMessageDto struct {
	MessageBaseDto MessageBaseDto `json:"message_base_dto"`
	// 图片高度
	Height int `json:"height"`
	// 图片宽度
	Width int `json:"width"`
}
