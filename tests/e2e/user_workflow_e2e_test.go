package e2e

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

// TestUserCompleteWorkflow 测试用户完整工作流
func TestUserCompleteWorkflow(t *testing.T) {
	// 设置测试环境
	gin.SetMode(gin.TestMode)

	// 测试数据
	testPhone := "13800138100"
	testUsername := "e2euser"
	testPassword := "123456"
	testCaptcha := "123456"

	// 步骤1: 用户注册
	t.Log("=== E2E测试: 用户注册 ===")
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
	t.Log("✅ 用户注册成功")

	// 步骤2: 用户登录（密码登录）
	t.Log("=== E2E测试: 用户密码登录 ===")
	loginReq := req.UserLoginReq{
		Phone:     testPhone,
		Password:  testPassword,
		LoginType: enum.LoginByPassword,
	}

	loginResp, err := service.LoginService(loginReq)
	assert.NoError(t, err)
	assert.True(t, loginResp.Success)

	// 验证登录响应
	loginData, ok := loginResp.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.NotEmpty(t, loginData["token"])
	assert.NotZero(t, loginData["uid"])
	assert.Equal(t, testUsername, loginData["name"])
	t.Logf("✅ 用户登录成功: uid=%v, name=%s", loginData["uid"], loginData["name"])

	// 步骤3: 用户登录（验证码登录）
	t.Log("=== E2E测试: 用户验证码登录 ===")
	setupCaptcha(testPhone, testCaptcha)
	loginByCaptchaReq := req.UserLoginReq{
		Phone:     testPhone,
		Captcha:   testCaptcha,
		LoginType: enum.LoginByPhoneCaptcha,
	}

	loginByCaptchaResp, err := service.LoginService(loginByCaptchaReq)
	assert.NoError(t, err)
	assert.True(t, loginByCaptchaResp.Success)
	t.Log("✅ 验证码登录成功")

	// 步骤4: 用户注销
	t.Log("=== E2E测试: 用户注销 ===")
	ctx := &gin.Context{}
	ctx.Set("uid", loginData["uid"])

	setupCaptcha(testPhone, testCaptcha)
	cancelReq := req.UserCancelReq{
		Captcha: testCaptcha,
	}

	cancelResp, err := service.CancelService(ctx, cancelReq)
	assert.NoError(t, err)
	assert.True(t, cancelResp.Success)
	t.Log("✅ 用户注销成功")

	// 步骤5: 验证用户已注销（尝试登录应该失败）
	t.Log("=== E2E测试: 验证用户已注销 ===")
	loginAfterCancelReq := req.UserLoginReq{
		Phone:     testPhone,
		Password:  testPassword,
		LoginType: enum.LoginByPassword,
	}

	loginAfterCancelResp, err := service.LoginService(loginAfterCancelReq)
	assert.NoError(t, err)
	assert.False(t, loginAfterCancelResp.Success)
	t.Log("✅ 验证用户已注销成功")
}

// TestMultipleUserWorkflow 测试多用户并发工作流
func TestMultipleUserWorkflow(t *testing.T) {
	// 测试多个用户同时注册和登录
	users := []struct {
		phone    string
		username string
		password string
	}{
		{"13800138101", "user1", "123456"},
		{"13800138102", "user2", "123456"},
		{"13800138103", "user3", "123456"},
	}

	for i, user := range users {
		t.Run(user.username, func(t *testing.T) {
			// 注册用户
			setupCaptcha(user.phone, "123456")
			registerReq := req.UserRegisterReq{
				Username: user.username,
				Password: user.password,
				Phone:    user.phone,
				Captcha:  "123456",
			}

			registerResp, err := service.RegisterService(registerReq)
			assert.NoError(t, err)
			assert.True(t, registerResp.Success)

			// 登录用户
			loginReq := req.UserLoginReq{
				Phone:     user.phone,
				Password:  user.password,
				LoginType: enum.LoginByPassword,
			}

			loginResp, err := service.LoginService(loginReq)
			assert.NoError(t, err)
			assert.True(t, loginResp.Success)

			t.Logf("✅ 用户 %d 注册和登录成功", i+1)

			// 清理测试数据
			cleanupUser(user.phone)
		})
	}
}

// TestUserErrorScenarios 测试用户错误场景
func TestUserErrorScenarios(t *testing.T) {
	t.Run("重复注册", func(t *testing.T) {
		testPhone := "13800138104"
		testUsername := "erroruser"
		testPassword := "123456"

		// 第一次注册
		setupCaptcha(testPhone, "123456")
		registerReq := req.UserRegisterReq{
			Username: testUsername,
			Password: testPassword,
			Phone:    testPhone,
			Captcha:  "123456",
		}

		resp1, err := service.RegisterService(registerReq)
		assert.NoError(t, err)
		assert.True(t, resp1.Success)

		// 第二次注册相同手机号
		setupCaptcha(testPhone, "123456")
		resp2, err := service.RegisterService(registerReq)
		assert.NoError(t, err)
		assert.False(t, resp2.Success)
		assert.Contains(t, resp2.Message, "手机号已存在")

		cleanupUser(testPhone)
	})

	t.Run("错误密码登录", func(t *testing.T) {
		testPhone := "13800138105"
		testUsername := "erroruser2"
		testPassword := "123456"

		// 注册用户
		setupCaptcha(testPhone, "123456")
		registerReq := req.UserRegisterReq{
			Username: testUsername,
			Password: testPassword,
			Phone:    testPhone,
			Captcha:  "123456",
		}

		resp, err := service.RegisterService(registerReq)
		assert.NoError(t, err)
		assert.True(t, resp.Success)

		// 使用错误密码登录
		wrongPasswordReq := req.UserLoginReq{
			Phone:     testPhone,
			Password:  "wrongpassword",
			LoginType: enum.LoginByPassword,
		}

		loginResp, err := service.LoginService(wrongPasswordReq)
		assert.NoError(t, err)
		assert.False(t, loginResp.Success)
		assert.Contains(t, loginResp.Message, "用户名或密码错误")

		cleanupUser(testPhone)
	})
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
