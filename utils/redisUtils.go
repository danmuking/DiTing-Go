package utils

import (
	domainEnum "DiTing-Go/domain/enum"
	"DiTing-Go/global"
	"fmt"
	"time"

	"github.com/goccy/go-json"
	"github.com/pkg/errors"
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
	if err != nil {
		global.Logger.Errorf("json marshal error: %v", err)
		return errors.New("json marshal error")
	}
	if err = global.Rdb.Set(key, valueByte, expireTime).Err(); err != nil {
		global.Logger.Errorf("key:%s, value:%s, redis set error: %v", key, valueByte, err)
		return errors.New("redis set error")
	}
	return nil
}

// GetValueFromRedis 获取字符串
func GetValueFromRedis(key string) (value []byte, err error) {
	valueByte, err := global.Rdb.Get(key).Bytes()
	if err != nil {
		return nil, err
	}
	return valueByte, nil
}

// DeleteValueFromRedis 删除字符串
func DeleteValueFromRedis(key string) error {
	if err := global.Rdb.Del(key).Err(); err != nil {
		global.Logger.Errorf("key:%s, redis delete error: %v", key, err)
		return errors.New("redis delete error")
	}
	return nil
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
