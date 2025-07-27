package req

type UserLoginReq struct {
	UserName  string `json:"username,omitempty"`            // 用户名可选
	Password  string `json:"password,omitempty"`            // 密码可选
	Phone     string `json:"phone,omitempty"`               // 手机号可选
	Captcha   string `json:"captcha,omitempty"`             // 验证码
	LoginType string `json:"login_type",binding:"required"` // 登录类型, 1: 用户名密码登录, 2: 手机号验证码登录
}
