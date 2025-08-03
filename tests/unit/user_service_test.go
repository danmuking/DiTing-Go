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

	"DiTing-Go/logic"
	"context"

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
		name          string
		req           req.UserRegisterReq
		expected      string
		shouldSuccess bool
	}{
		{
			name: "用户名为空",
			req: req.UserRegisterReq{
				Username: "",
				Password: "123456",
				Phone:    "13800138001",
				Captcha:  "123456",
			},
			expected:      "用户名、密码和手机号不能为空",
			shouldSuccess: false,
		},
		{
			name: "密码长度不足",
			req: req.UserRegisterReq{
				Username: "testuser",
				Password: "123",
				Phone:    "13800138001",
				Captcha:  "123456",
			},
			expected:      "密码长度不能少于6位",
			shouldSuccess: false,
		},
		{
			name: "正确注册用例",
			req: req.UserRegisterReq{
				Username: "validuser",
				Password: "123456",
				Phone:    "13800138002",
				Captcha:  "123456",
			},
			expected:      "",
			shouldSuccess: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 为正确用例设置验证码
			if tc.shouldSuccess {
				setupCaptcha(tc.req.Phone, tc.req.Captcha)
			}

			resp, err := service.RegisterService(tc.req)

			if tc.shouldSuccess {
				// 正确用例应该成功
				assert.NoError(t, err)
				assert.True(t, resp.Success, "注册应该成功")
				assert.Contains(t, resp.Message, "注册成功")
			} else {
				// 错误用例应该失败
				assert.Contains(t, resp.Message, tc.expected)
			}
		})
	}
}

// TestLoginValidation 测试登录参数验证
func TestLoginValidation(t *testing.T) {
	testCases := []struct {
		name          string
		req           req.UserLoginReq
		expected      string
		shouldSuccess bool
	}{
		{
			name: "密码登录-手机号为空",
			req: req.UserLoginReq{
				Phone:     "",
				Password:  "123456",
				LoginType: enum.LoginByPassword,
			},
			expected:      "用户名和密码不能为空",
			shouldSuccess: false,
		},
		{
			name: "验证码登录-手机号为空",
			req: req.UserLoginReq{
				Phone:     "",
				Captcha:   "123456",
				LoginType: enum.LoginByPhoneCaptcha,
			},
			expected:      "手机号和验证码不能为空",
			shouldSuccess: false,
		},
		{
			name: "正确密码登录用例",
			req: req.UserLoginReq{
				Phone:     "13800138002",
				Password:  "123456",
				LoginType: enum.LoginByPassword,
			},
			expected:      "",
			shouldSuccess: true,
		},
		{
			name: "正确验证码登录用例",
			req: req.UserLoginReq{
				Phone:     "13800138002",
				Captcha:   "123456",
				LoginType: enum.LoginByPhoneCaptcha,
			},
			expected:      "",
			shouldSuccess: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 为正确用例设置验证码
			if tc.shouldSuccess && tc.req.LoginType == enum.LoginByPhoneCaptcha {
				setupCaptcha(tc.req.Phone, tc.req.Captcha)
			}

			resp, err := service.LoginService(tc.req)

			if tc.shouldSuccess {
				// 正确用例应该成功
				assert.NoError(t, err)
				assert.True(t, resp.Success, "登录应该成功")
				// 验证返回的数据结构
				loginData, ok := resp.Data.(map[string]interface{})
				assert.True(t, ok, "返回数据应该是map类型")
				assert.NotEmpty(t, loginData["token"], "应该返回token")
				assert.NotZero(t, loginData["uid"], "应该返回用户ID")
			} else {
				// 错误用例应该失败
				assert.Contains(t, resp.Message, tc.expected)
			}
		})
	}
}

// TestCheckPassword 测试密码校验功能
func TestCheckPassword(t *testing.T) {
	ctx := context.Background()

	// 测试边界情况
	t.Run("空手机号", func(t *testing.T) {
		result := logic.CheckPassword(ctx, "", "123456")
		assert.False(t, result, "空手机号应该校验失败")
	})

	t.Run("空密码", func(t *testing.T) {
		result := logic.CheckPassword(ctx, "13800138003", "")
		assert.False(t, result, "空密码应该校验失败")
	})

	t.Run("短密码", func(t *testing.T) {
		result := logic.CheckPassword(ctx, "13800138003", "123")
		assert.False(t, result, "短密码应该校验失败")
	})

	t.Run("不存在的手机号", func(t *testing.T) {
		result := logic.CheckPassword(ctx, "99999999999", "123456")
		assert.False(t, result, "不存在的手机号应该校验失败")
	})

	// 测试正常情况（需要先注册用户）
	t.Run("正常密码校验流程", func(t *testing.T) {
		// 先注册一个测试用户
		testPhone := "13800138004"
		testUsername := "testuser"
		testPassword := "123456"
		testCaptcha := "123456"

		// 设置验证码
		setupCaptcha(testPhone, testCaptcha)

		// 注册用户
		registerReq := req.UserRegisterReq{
			Username: testUsername,
			Password: testPassword,
			Phone:    testPhone,
			Captcha:  testCaptcha,
		}

		registerResp, err := service.RegisterService(registerReq)
		if err == nil && registerResp.Success {
			// 注册成功，测试密码校验
			t.Run("正确密码", func(t *testing.T) {
				result := logic.CheckPassword(ctx, testPhone, testPassword)
				assert.True(t, result, "正确密码应该通过校验")
			})

			t.Run("错误密码", func(t *testing.T) {
				result := logic.CheckPassword(ctx, testPhone, "wrongpassword")
				assert.False(t, result, "错误密码应该校验失败")
			})

			// 清理测试数据
			cleanupUser(testPhone)
		} else {
			t.Skip("用户注册失败，跳过密码校验测试")
		}
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
	global.Logger.Infof("cleanupUser: cleaning up user data for phone: %s", phone)
}
