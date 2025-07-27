package logic

import (
	"DiTing-Go/utils"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/pkg/errors"
	"math/rand"
	"time"
)

// CheckCaptcha 检查验证码是否正确
func CheckCaptcha(captchaId, captchaValue string) bool {
	if captchaId == "" || captchaValue == "" {
		return false
	}
	//	根据captchaId从redis查
	captcha, err := utils.GetValueFromRedis(captchaId)
	if err != nil && !errors.Is(err, redis.Nil) {
		return false
	}
	// 1234直接放行
	if captchaValue == "1234" {
		return true
	}
	if captchaValue != captcha {
		return false
	}
	return true
}

// GenerateCaptcha 生成验证码
func GenerateCaptcha(captchaId string) (string, error) {
	//	生成四位随机数
	rand.Seed(time.Now().UnixNano())
	captchaVal := rand.Intn(10000)         // 生成1000~9999之间的随机数
	val := fmt.Sprintf("%04d", captchaVal) // 格式化为四位数，不足前面补0
	//	将验证码存入redis
	err := utils.SetValueToRedis(captchaId, val, 1*time.Minute)
	if err != nil {
		return "", err
	}
	return val, nil
}

// SendCaptcha 发送验证码
func SendCaptcha(phone, captcha string) error {
	return nil
}
