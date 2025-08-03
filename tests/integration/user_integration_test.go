package integration

import (
	"DiTing-Go/domain/enum"
	"DiTing-Go/domain/vo/req"
	"DiTing-Go/logic"
	"DiTing-Go/service"
	"DiTing-Go/utils"
	"context"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestUserRegistrationFlow 测试用户注册流程
func TestUserRegistrationFlow(t *testing.T) {
	// 测试数据
	testPhone := "13800138001"
	testUsername := "testuser1"
	testPassword := "123456"
	testCaptcha := "123456"

	// 设置验证码
	setupCaptcha(testPhone, testCaptcha)

	// 执行注册
	registerReq := req.UserRegisterReq{
		Username: testUsername,
		Password: testPassword,
		Phone:    testPhone,
		Captcha:  testCaptcha,
	}

	resp, err := service.RegisterService(registerReq)
	assert.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Contains(t, resp.Message, "注册成功")

	// 验证用户是否真的被创建
	userExists, err := logic.CheckPhoneInDB(context.Background(), testPhone)
	assert.NoError(t, err)
	assert.True(t, userExists)

	// 清理测试数据
	cleanupUser(testPhone)
}

// TestUserLoginFlow 测试用户登录流程
func TestUserLoginFlow(t *testing.T) {
	// 先注册一个用户
	testPhone := "13800138002"
	testUsername := "testuser2"
	testPassword := "123456"
	testCaptcha := "123456"

	setupCaptcha(testPhone, testCaptcha)
	registerReq := req.UserRegisterReq{
		Username: testUsername,
		Password: testPassword,
		Phone:    testPhone,
		Captcha:  testCaptcha,
	}

	registerResp, err := service.RegisterService(registerReq)
	assert.NoError(t, err)
	assert.True(t, registerResp.Success)

	// 测试密码登录
	loginReq := req.UserLoginReq{
		Phone:     testPhone,
		Password:  testPassword,
		LoginType: enum.LoginByPassword,
	}

	loginResp, err := service.LoginService(loginReq)
	assert.NoError(t, err)
	assert.True(t, loginResp.Success)

	// 验证登录响应数据结构
	loginData, ok := loginResp.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.NotEmpty(t, loginData["token"])
	assert.NotZero(t, loginData["uid"])
	assert.Equal(t, testUsername, loginData["name"])

	// 清理测试数据
	cleanupUser(testPhone)
}

// TestUserCancelFlow 测试用户注销流程
func TestUserCancelFlow(t *testing.T) {
	// 先注册一个用户
	testPhone := "13800138003"
	testUsername := "testuser3"
	testPassword := "123456"
	testCaptcha := "123456"

	setupCaptcha(testPhone, testCaptcha)
	registerReq := req.UserRegisterReq{
		Username: testUsername,
		Password: testPassword,
		Phone:    testPhone,
		Captcha:  testCaptcha,
	}

	registerResp, err := service.RegisterService(registerReq)
	assert.NoError(t, err)
	assert.True(t, registerResp.Success)

	// 设置注销验证码
	setupCaptcha(testPhone, testCaptcha)

	// 执行注销
	ctx := &gin.Context{}
	ctx.Set("uid", int64(12345))

	cancelReq := req.UserCancelReq{
		Captcha: testCaptcha,
	}

	cancelResp, err := service.CancelService(ctx, cancelReq)
	assert.NoError(t, err)
	assert.True(t, cancelResp.Success)
	assert.Contains(t, cancelResp.Message, "注销成功")

	// 验证用户是否真的被删除
	userExists, err := logic.CheckPhoneInDB(context.Background(), testPhone)
	assert.NoError(t, err)
	assert.False(t, userExists)
}

// TestDuplicateRegistration 测试重复注册
func TestDuplicateRegistration(t *testing.T) {
	testPhone := "13800138004"
	testUsername := "testuser4"
	testPassword := "123456"
	testCaptcha := "123456"

	// 第一次注册
	setupCaptcha(testPhone, testCaptcha)
	registerReq := req.UserRegisterReq{
		Username: testUsername,
		Password: testPassword,
		Phone:    testPhone,
		Captcha:  testCaptcha,
	}

	resp1, err := service.RegisterService(registerReq)
	assert.NoError(t, err)
	assert.True(t, resp1.Success)

	// 第二次注册相同手机号
	setupCaptcha(testPhone, testCaptcha)
	resp2, err := service.RegisterService(registerReq)
	assert.NoError(t, err)
	assert.False(t, resp2.Success)
	assert.Contains(t, resp2.Message, "手机号已存在")

	// 清理测试数据
	cleanupUser(testPhone)
}

// TestLoginWithWrongCredentials 测试错误凭据登录
func TestLoginWithWrongCredentials(t *testing.T) {
	// 先注册一个用户
	testPhone := "13800138005"
	testUsername := "testuser5"
	testPassword := "123456"
	testCaptcha := "123456"

	setupCaptcha(testPhone, testCaptcha)
	registerReq := req.UserRegisterReq{
		Username: testUsername,
		Password: testPassword,
		Phone:    testPhone,
		Captcha:  testCaptcha,
	}

	resp, err := service.RegisterService(registerReq)
	assert.NoError(t, err)
	assert.True(t, resp.Success)

	// 测试错误密码
	wrongPasswordReq := req.UserLoginReq{
		Phone:     testPhone,
		Password:  "wrongpassword",
		LoginType: enum.LoginByPassword,
	}

	resp2, err := service.LoginService(wrongPasswordReq)
	assert.NoError(t, err)
	assert.False(t, resp2.Success)
	assert.Contains(t, resp2.Message, "用户名或密码错误")

	// 测试错误验证码
	wrongCaptchaReq := req.UserLoginReq{
		Phone:     testPhone,
		Captcha:   "999999",
		LoginType: enum.LoginByPhoneCaptcha,
	}

	resp3, err := service.LoginService(wrongCaptchaReq)
	assert.NoError(t, err)
	assert.False(t, resp3.Success)
	assert.Contains(t, resp3.Message, "验证码错误")

	// 清理测试数据
	cleanupUser(testPhone)
}

// TestCancelWithWrongCaptcha 测试错误验证码注销
func TestCancelWithWrongCaptcha(t *testing.T) {
	// 先注册一个用户
	testPhone := "13800138006"
	testUsername := "testuser6"
	testPassword := "123456"
	testCaptcha := "123456"

	setupCaptcha(testPhone, testCaptcha)
	registerReq := req.UserRegisterReq{
		Username: testUsername,
		Password: testPassword,
		Phone:    testPhone,
		Captcha:  testCaptcha,
	}

	resp, err := service.RegisterService(registerReq)
	assert.NoError(t, err)
	assert.True(t, resp.Success)

	// 测试错误验证码注销
	ctx := &gin.Context{}
	ctx.Set("uid", int64(12345))

	wrongCaptchaReq := req.UserCancelReq{
		Captcha: "999999",
	}

	resp2, err := service.CancelService(ctx, wrongCaptchaReq)
	assert.NoError(t, err)
	assert.False(t, resp2.Success)
	assert.Contains(t, resp2.Message, "验证码错误")

	// 清理测试数据
	cleanupUser(testPhone)
}

// TestUserRegistrationWithValidData 测试有效数据注册
func TestUserRegistrationWithValidData(t *testing.T) {
	testCases := []struct {
		name     string
		username string
		password string
		phone    string
		captcha  string
	}{
		{
			name:     "标准注册",
			username: "testuser7",
			password: "123456",
			phone:    "13800138007",
			captcha:  "123456",
		},
		{
			name:     "长用户名",
			username: "verylongusername123",
			password: "123456",
			phone:    "13800138008",
			captcha:  "123456",
		},
		{
			name:     "复杂密码",
			username: "testuser9",
			password: "P@ssw0rd123",
			phone:    "13800138009",
			captcha:  "123456",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			registerReq := req.UserRegisterReq{
				Username: tc.username,
				Password: tc.password,
				Phone:    tc.phone,
				Captcha:  tc.captcha,
			}

			setupCaptcha(tc.phone, tc.captcha)
			resp, err := service.RegisterService(registerReq)
			assert.NoError(t, err)
			assert.True(t, resp.Success)

			// 清理测试数据
			cleanupUser(tc.phone)
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

// cleanupUser 清理用户数据
func cleanupUser(phone string) {
	// 这里应该实现清理用户数据的逻辑
	// 由于涉及到数据库操作，这里只是占位符
	// 实际实现时需要根据具体的数据库操作来清理
}
