package unit

import (
	"DiTing-Go/domain/enum"
	"DiTing-Go/domain/vo/req"
	"DiTing-Go/service"
	"DiTing-Go/utils"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestUserLifecycleFlow 测试用户完整生命周期流程
func TestUserLifecycleFlow(t *testing.T) {
	// 设置测试环境
	gin.SetMode(gin.TestMode)
	ctx := &gin.Context{}
	ctx.Set("uid", int64(12345))

	// 测试数据
	testPhone := "13800138000"
	testUsername := "testuser"
	testPassword := "123456"
	testCaptcha := "123456"

	// 步骤1: 用户注册
	t.Log("=== 步骤1: 用户注册测试 ===")
	registerReq := req.UserRegisterReq{
		Username: testUsername,
		Password: testPassword,
		Phone:    testPhone,
		Captcha:  testCaptcha,
	}

	// 设置验证码
	setupCaptcha(testPhone, testCaptcha)
	registerResp, err := service.RegisterService(registerReq)
	assert.NoError(t, err)
	assert.True(t, registerResp.Success)

	// 步骤2: 用户登录（密码登录）
	t.Log("=== 步骤2: 用户密码登录测试 ===")
	loginReq := req.UserLoginReq{
		Phone:     testPhone,
		Password:  testPassword,
		LoginType: enum.LoginByPassword,
	}

	loginResp, err := service.LoginService(loginReq)
	assert.NoError(t, err)
	assert.True(t, loginResp.Success)

	// 步骤3: 用户登录（验证码登录）
	t.Log("=== 步骤3: 用户验证码登录测试 ===")
	loginByCaptchaReq := req.UserLoginReq{
		Phone:     testPhone,
		Captcha:   testCaptcha,
		LoginType: enum.LoginByPhoneCaptcha,
	}

	setupCaptcha(testPhone, testCaptcha)
	loginByCaptchaResp, err := service.LoginService(loginByCaptchaReq)
	assert.NoError(t, err)
	assert.True(t, loginByCaptchaResp.Success)

	// 步骤4: 用户注销
	t.Log("=== 步骤4: 用户注销测试 ===")
	cancelReq := req.UserCancelReq{
		Captcha: testCaptcha,
	}

	setupCaptcha(testPhone, testCaptcha)
	cancelResp, err := service.CancelService(ctx, cancelReq)
	assert.NoError(t, err)
	assert.True(t, cancelResp.Success)
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
