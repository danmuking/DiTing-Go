package utils

import (
	domainEnum "DiTing-Go/domain/enum"
	"DiTing-Go/global"
	"fmt"
	"github.com/goccy/go-json"
	"github.com/pkg/errors"
	"time"
)

// MakeUserPhoneKey 构造用户手机号
func MakeUserPhoneKey(phone string) string {
	return fmt.Sprintf(domainEnum.PhoneUidMap, phone)
}

// MakeUserCaptchaKey 构造验证码key
func MakeUserCaptchaKey(phone string) string {
	return fmt.Sprintf(domainEnum.UserCaptcha, phone)
}

// SetValueToRedis 设置字符串
func SetValueToRedis(key string, value string, expireTime time.Duration) error {
	valueByte, err := json.Marshal(value)
	if err = global.Rdb.Set(key, valueByte, expireTime).Err(); err != nil {
		global.Logger.Errorf("key:%s, value:%s, redis set error: %v", key, valueByte, err)
		return errors.New("redis set error")
	}
	return nil
}

// GetValueFromRedis 获取字符串
func GetValueFromRedis(key string) (value string, err error) {
	valueByte, err := global.Rdb.Get(key).Result()
	if err != nil {
		return "", err
	}
	return valueByte, nil
}

//// GetData 获取数据
//func GetData(cacheKey string, value any, dbQueryFunc func() (interface{}, error)) error {
//	// 1. 从缓存中获取数据
//	err := GetString(cacheKey, value)
//	// 查询到数据
//	if err == nil {
//		return nil
//	} else if !errors.Is(err, redis.Nil) {
//		return err
//	}
//	err = QueryAndSet(cacheKey, value, dbQueryFunc)
//	if err != nil {
//		return err
//	}
//	return nil
//}

//// QueryAndSet 查询数据库并设置缓存
//func QueryAndSet(cacheKey string, value any, dbQueryFunc func() (interface{}, error)) error {
//	// 2. 从数据库中获取数据
//	result, err := dbQueryFunc()
//	if err != nil {
//		return err
//	}
//	err = copier.Copy(value, result)
//	if err != nil {
//		global.Logger.Errorf("拷贝数据失败: %v", err)
//		return err
//	}
//	// 3. 将查询结果写回缓存
//	if err = SetString(cacheKey, result); err != nil {
//		global.Logger.Errorf("写入redis失败: %v", err)
//		return err
//	}
//	return err
//}
//
//func RemoveData(key string) {
//	global.Rdb.Del(key)
//}
