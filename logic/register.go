package logic

import (
	"DiTing-Go/domain/enum"
	"DiTing-Go/global"
	"DiTing-Go/utils"
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"gorm.io/gorm"
)

// CheckCaptchaProcess 检查验证码
func CheckCaptchaProcess(phone, captcha string) bool {
	// 检查验证码
	phoneKey := utils.MakeUserCaptchaKey(phone)
	if !CheckCaptcha(phoneKey, captcha) {
		global.Logger.Infof("验证码错误")
		return false
	}
	return true
}

// CheckCaptchaExist 检查验证码是否存在
func CheckCaptchaExist(phone string) bool {
	// 检查验证码
	phoneKey := utils.MakeUserCaptchaKey(phone)
	_, err := utils.GetValueFromRedis(phoneKey)
	if errors.Is(err, redis.Nil) {
		global.Logger.Infof("phoneKey:%s, 验证码不存在", phoneKey)
		return false
	} else if err != nil {
		global.Logger.Errorf("phoneKey:%s, 查询验证码失败: %v", phoneKey, err)
		return false
	}

	global.Logger.Infof("phoneKey:%s, 验证码存在", phoneKey)
	return true
}

// CheckPhoneInRedis 检查手机号在redis中是否存在
func CheckPhoneInRedis(phone string) bool {
	phoneKey := utils.MakeUserPhoneKey(phone)
	rst, err := utils.GetValueFromRedis(phoneKey)
	// 如果redis没查到,查数据库
	if errors.Is(err, redis.Nil) {
		return false
	}
	// 如果redis查询出错，查数据库
	if err != nil {
		global.Logger.Errorf("phoneKey:%s, 查询手机号失败: %v", phoneKey, err)
		return false
	}
	// 如果redis查到了,直接返回
	if rst != "" {
		return true
	}
	return false
}

// CheckPhoneInDB 检查手机号在db中是否存在
func CheckPhoneInDB(ctx context.Context, phone string) (bool, error) {
	// 如果redis查不到,查数据库
	user, err := utils.QueryUserByPhone(ctx, phone)
	// 数据库查询出错，返回失败
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		global.Logger.Errorf("phone:%s, 查询手机号失败: %v", phone, err)
		return true, err
	}
	//数据库查到了，返回失败
	if user != nil {
		// 将用户信息存入redis
		phoneKey := utils.MakeUserPhoneKey(phone)
		// 将phone->user.ID存入redis
		if err := utils.SetValueToRedis(phoneKey, fmt.Sprintf("%d", user.ID), enum.DefaultCacheTime); err != nil {
			global.Logger.Errorf("phone:%s, 设置phoneUid映射 redis失败: %v", phone, err)
			return true, err
		}
		global.Logger.Infof("phone:%s, 用户已存在", phone)
		return true, err
	}
	return false, nil
}
