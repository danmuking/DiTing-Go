package performance

import (
	"DiTing-Go/domain/enum"
	"DiTing-Go/domain/vo/req"
	"DiTing-Go/service"
	"DiTing-Go/utils"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// BenchmarkUserRegistration 用户注册性能基准测试
func BenchmarkUserRegistration(b *testing.B) {
	// 重置计时器
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		phone := fmt.Sprintf("13800139%03d", i%1000)
		username := fmt.Sprintf("benchuser%d", i)
		password := "123456"
		captcha := "123456"

		// 设置验证码
		setupCaptcha(phone, captcha)

		// 执行注册
		registerReq := req.UserRegisterReq{
			Username: username,
			Password: password,
			Phone:    phone,
			Captcha:  captcha,
		}

		resp, err := service.RegisterService(registerReq)
		assert.NoError(b, err)
		assert.True(b, resp.Success)

		// 清理测试数据
		cleanupUser(phone)
	}
}

// BenchmarkUserLogin 用户登录性能基准测试
func BenchmarkUserLogin(b *testing.B) {
	// 先注册一个用户用于测试
	testPhone := "13800139000"
	testUsername := "benchuser"
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
	assert.NoError(b, err)
	assert.True(b, resp.Success)

	// 重置计时器
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		loginReq := req.UserLoginReq{
			Phone:     testPhone,
			Password:  testPassword,
			LoginType: enum.LoginByPassword,
		}

		loginResp, err := service.LoginService(loginReq)
		assert.NoError(b, err)
		assert.True(b, loginResp.Success)
	}

	// 清理测试数据
	cleanupUser(testPhone)
}

// TestConcurrentUserRegistration 并发用户注册测试
func TestConcurrentUserRegistration(t *testing.T) {
	const numUsers = 10
	const numGoroutines = 5

	var wg sync.WaitGroup
	errors := make(chan error, numUsers*numGoroutines)

	startTime := time.Now()

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(routineID int) {
			defer wg.Done()

			for j := 0; j < numUsers; j++ {
				userID := routineID*numUsers + j
				phone := fmt.Sprintf("13800140%03d", userID)
				username := fmt.Sprintf("concurrentuser%d", userID)
				password := "123456"
				captcha := "123456"

				// 设置验证码
				setupCaptcha(phone, captcha)

				// 执行注册
				registerReq := req.UserRegisterReq{
					Username: username,
					Password: password,
					Phone:    phone,
					Captcha:  captcha,
				}

				resp, err := service.RegisterService(registerReq)
				if err != nil {
					errors <- err
					return
				}

				if !resp.Success {
					errors <- fmt.Errorf("注册失败: %s", resp.Message)
					return
				}

				// 清理测试数据
				cleanupUser(phone)
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	duration := time.Since(startTime)
	t.Logf("并发注册 %d 个用户，耗时: %v", numUsers*numGoroutines, duration)

	// 检查是否有错误
	for err := range errors {
		t.Errorf("并发注册错误: %v", err)
	}
}

// TestConcurrentUserLogin 并发用户登录测试
func TestConcurrentUserLogin(t *testing.T) {
	const numUsers = 10
	const numGoroutines = 5

	// 先注册一些用户
	users := make([]string, numUsers*numGoroutines)
	for i := 0; i < numUsers*numGoroutines; i++ {
		phone := fmt.Sprintf("13800141%03d", i)
		username := fmt.Sprintf("loginuser%d", i)
		password := "123456"
		captcha := "123456"

		setupCaptcha(phone, captcha)
		registerReq := req.UserRegisterReq{
			Username: username,
			Password: password,
			Phone:    phone,
			Captcha:  captcha,
		}

		resp, err := service.RegisterService(registerReq)
		assert.NoError(t, err)
		assert.True(t, resp.Success)

		users[i] = phone
	}

	var wg sync.WaitGroup
	errors := make(chan error, numUsers*numGoroutines)

	startTime := time.Now()

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(routineID int) {
			defer wg.Done()

			for j := 0; j < numUsers; j++ {
				userID := routineID*numUsers + j
				phone := users[userID]
				password := "123456"

				loginReq := req.UserLoginReq{
					Phone:     phone,
					Password:  password,
					LoginType: enum.LoginByPassword,
				}

				resp, err := service.LoginService(loginReq)
				if err != nil {
					errors <- err
					return
				}

				if !resp.Success {
					errors <- fmt.Errorf("登录失败: %s", resp.Message)
					return
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	duration := time.Since(startTime)
	t.Logf("并发登录 %d 个用户，耗时: %v", numUsers*numGoroutines, duration)

	// 检查是否有错误
	for err := range errors {
		t.Errorf("并发登录错误: %v", err)
	}

	// 清理测试数据
	for _, phone := range users {
		cleanupUser(phone)
	}
}

// TestLoadTest 负载测试
func TestLoadTest(t *testing.T) {
	const numUsers = 100
	const duration = 30 * time.Second

	startTime := time.Now()
	successCount := 0
	errorCount := 0
	var mu sync.Mutex

	// 创建用户注册任务
	for i := 0; i < numUsers; i++ {
		go func(userID int) {
			phone := fmt.Sprintf("13800142%03d", userID)
			username := fmt.Sprintf("loaduser%d", userID)
			password := "123456"
			captcha := "123456"

			setupCaptcha(phone, captcha)
			registerReq := req.UserRegisterReq{
				Username: username,
				Password: password,
				Phone:    phone,
				Captcha:  captcha,
			}

			resp, err := service.RegisterService(registerReq)
			mu.Lock()
			if err != nil || !resp.Success {
				errorCount++
			} else {
				successCount++
			}
			mu.Unlock()

			// 清理测试数据
			cleanupUser(phone)
		}(i)
	}

	// 等待指定时间
	time.Sleep(duration)

	totalTime := time.Since(startTime)
	t.Logf("负载测试结果:")
	t.Logf("  总时间: %v", totalTime)
	t.Logf("  成功请求: %d", successCount)
	t.Logf("  失败请求: %d", errorCount)
	t.Logf("  成功率: %.2f%%", float64(successCount)/float64(successCount+errorCount)*100)
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
