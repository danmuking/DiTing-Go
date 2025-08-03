package req

type UserCancelReq struct {
	Captcha string `json:"captcha" binding:"required"`
}
