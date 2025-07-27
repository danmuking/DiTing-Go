package req

type UserRegisterReq struct {
	// 用户名
	Username string `json:"username" binding:"required"`
	// 密码
	Password string `json:"password" binding:"required"`
	// 手机号
	Phone string `json:"phone" binding:"required"`
	// 验证码
	Captcha string `json:"captcha" binding:"required"` // 如果需要验证码，可以取消注释
}
