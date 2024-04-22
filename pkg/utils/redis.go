package utils

import (
	domainEnum "DiTing-Go/domain/enum"
	"DiTing-Go/global"
	"github.com/go-redis/redis"
	"github.com/goccy/go-json"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
)

// SetString 设置字符串
func SetString(key string, value any) error {
	valueByte, err := json.Marshal(value)
	if err = global.Rdb.Set(key, valueByte, domainEnum.CacheTime).Err(); err != nil {
		return errors.New("redis set error")
	}
	return nil
}

// GetString 获取字符串
func GetString(key string, value any) error {
	valueByte, err := global.Rdb.Get(key).Result()
	if err != nil && errors.Is(err, redis.Nil) {
		return err
	} else if err != nil {
		return errors.New("redis get error")
	}
	err = json.Unmarshal([]byte(valueByte), value)
	if err != nil {
		return errors.New("jsonUtils unmarshal error")
	}
	return nil
}

// GetData 获取数据
func GetData(cacheKey string, value any, dbQueryFunc func() (interface{}, error)) error {
	// 1. 从缓存中获取数据
	err := GetString(cacheKey, value)
	// 查询到数据
	if err == nil {
		return nil
	} else if !errors.Is(err, redis.Nil) {
		return err
	}
	err = QueryAndSet(cacheKey, value, dbQueryFunc)
	if err != nil {
		return err
	}
	return nil
}

// QueryAndSet 查询数据库并设置缓存
func QueryAndSet(cacheKey string, value any, dbQueryFunc func() (interface{}, error)) error {
	// 2. 从数据库中获取数据
	result, err := dbQueryFunc()
	if err != nil {
		return err
	}
	err = copier.Copy(value, result)
	if err != nil {
		global.Logger.Errorf("拷贝数据失败: %v", err)
		return err
	}
	// 3. 将查询结果写回缓存
	if err = SetString(cacheKey, result); err != nil {
		global.Logger.Errorf("写入redis失败: %v", err)
		return err
	}
	return err
}

func RemoveData(key string) {
	global.Rdb.Del(key)
}
