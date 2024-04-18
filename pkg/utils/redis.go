package utils

import (
	domainEnum "DiTing-Go/domain/enum"
	"DiTing-Go/global"
	"github.com/go-redis/redis"
	"github.com/goccy/go-json"
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

func GetString(key string, value any) error {
	valueByte, err := global.Rdb.Get(key).Result()
	if err != nil && errors.Is(err, redis.Nil) {
		return err
	} else if err != nil {
		return errors.New("redis get error")
	}
	err = json.Unmarshal([]byte(valueByte), value)
	if err != nil {
		return errors.New("json unmarshal error")
	}
	return nil
}
