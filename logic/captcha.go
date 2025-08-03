package logic

import (
	"DiTing-Go/global"
	"DiTing-Go/utils"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"
)

// CheckCaptcha 检查验证码是否正确
func CheckCaptcha(captchaId, captchaValue string) bool {
	if captchaId == "" || captchaValue == "" {
		return false
	}
	//	根据captchaId从redis查
	captchaByte, err := utils.GetValueFromRedis(captchaId)
	if err != nil {
		global.Logger.Errorf("get captcha from redis error: %v", err)
		return false
	}
	var captcha string
	err = json.Unmarshal(captchaByte, &captcha)
	if err != nil {
		global.Logger.Errorf("json unmarshal error: %v", err)
		return false
	}
	if captchaValue == "1234" {
		global.Logger.Infof("captcha is 1234, pass")
		return true
	}
	// 1234直接放行
	return captchaValue == captcha
}

// GenerateCaptcha 生成验证码
func GenerateCaptcha(captchaId string) (string, error) {
	//	生成四位随机数
	rand.Seed(time.Now().UnixNano())
	captchaVal := rand.Intn(10000)         // 生成1000~9999之间的随机数
	val := fmt.Sprintf("%04d", captchaVal) // 格式化为四位数，不足前面补0
	//	将验证码存入redis
	captchaKey := utils.MakeUserCaptchaKey(captchaId)
	err := utils.SetValueToRedis(captchaKey, val, 1*time.Minute)
	if err != nil {
		return "", err
	}
	return val, nil
}

// SendCaptcha 发送验证码
func SendCaptcha(phone, captcha string) error {
	return nil
}
