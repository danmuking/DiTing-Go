package req

type CaptchaReq struct {
	// 手机号
	Phone string `json:"phone" binding:"required"`
}
