package unit

import (
	"DiTing-Go/domain/enum"
	"DiTing-Go/domain/vo/req"
	"DiTing-Go/global"
	"DiTing-Go/service"
	"DiTing-Go/utils"
	"DiTing-Go/utils/setting"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// 初始化函数，在包加载时执行
func init() {
	// 设置测试环境变量
	gin.SetMode(gin.TestMode)

	// 初始化配置
	setting.ConfigInit()

	// 初始化简单的测试日志
	global.Logger = logrus.New()
	global.Logger.SetOutput(gin.DefaultWriter)
	global.Logger.SetLevel(logrus.InfoLevel)

	// 初始化Redis
	global.RedisInit()

	// 初始化数据库
	global.DBInit()
}

// TestRegisterValidation 测试注册参数验证
func TestRegisterValidation(t *testing.T) {
	testCases := []struct {
		name     string
		req      req.UserRegisterReq
		expected string
	}{
		{
			name: "用户名为空",
			req: req.UserRegisterReq{
				Username: "",
				Password: "123456",
				Phone:    "13800138001",
				Captcha:  "123456",
			},
			expected: "用户名、密码和手机号不能为空",
		},
		{
			name: "密码长度不足",
			req: req.UserRegisterReq{
				Username: "testuser",
				Password: "123",
				Phone:    "13800138001",
				Captcha:  "123456",
			},
			expected: "密码长度不能少于6位",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := service.RegisterService(tc.req)
			assert.NoError(t, err)
			assert.False(t, resp.Success)
			assert.Contains(t, resp.Message, tc.expected)
		})
	}
}

// TestLoginValidation 测试登录参数验证
func TestLoginValidation(t *testing.T) {
	testCases := []struct {
		name     string
		req      req.UserLoginReq
		expected string
	}{
		{
			name: "密码登录-手机号为空",
			req: req.UserLoginReq{
				Phone:     "",
				Password:  "123456",
				LoginType: enum.LoginByPassword,
			},
			expected: "用户名和密码不能为空",
		},
		{
			name: "验证码登录-手机号为空",
			req: req.UserLoginReq{
				Phone:     "",
				Captcha:   "123456",
				LoginType: enum.LoginByPhoneCaptcha,
			},
			expected: "手机号和验证码不能为空",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := service.LoginService(tc.req)
			assert.NoError(t, err)
			assert.False(t, resp.Success)
			assert.Contains(t, resp.Message, tc.expected)
		})
	}
}

// setupCaptcha 设置验证码
func setupCaptcha(phone, captcha string) {
	captchaKey := utils.MakeUserCaptchaKey(phone)
	err := utils.SetValueToRedis(captchaKey, captcha, 5*time.Minute)
	if err != nil {
		panic(err)
	}
}
