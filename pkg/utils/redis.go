package utils

import (
	domainEnum "DiTing-Go/domain/enum"
	"DiTing-Go/global"
	"github.com/go-redis/redis"
	"github.com/goccy/go-json"
	"github.com/pkg/errors"
	"gorm.io/gorm"
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
func GetString(key string) (any, error) {
	valueByte, err := global.Rdb.Get(key).Result()
	if err != nil && errors.Is(err, redis.Nil) {
		return nil, err
	} else if err != nil {
		return nil, errors.New("redis get error")
	}
	var value any
	err = json.Unmarshal([]byte(valueByte), &value)
	if err != nil {
		return nil, errors.New("json unmarshal error")
	}
	return value, nil
}

// GetData 获取数据
func GetData(cacheKey string, dbQueryFunc func() (interface{}, error)) (any, error) {
	// 1. 从缓存中获取数据
	value, err := GetString(cacheKey)
	// 查询到数据
	if err == nil {
		return value, nil
	} else if !errors.Is(err, redis.Nil) {
		return nil, err
	}
	value, err = QueryAndSet(cacheKey, dbQueryFunc)
	if err != nil {
		return nil, err
	}
	return value, nil
}

// QueryAndSet 查询数据库并设置缓存
func QueryAndSet(cacheKey string, dbQueryFunc func() (interface{}, error)) (any, error) {
	// 2. 从数据库中获取数据
	value, err := dbQueryFunc()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			global.Logger.Errorf("查询数据库失败: %v", err)
		}
		return nil, err
	}
	// 3. 将查询结果写回缓存
	if err = SetString(cacheKey, value); err != nil {
		global.Logger.Errorf("写入redis失败: %v", err)
		return nil, err
	}
	return value, err
}
