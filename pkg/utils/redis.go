package utils

import (
	domainEnum "DiTing-Go/domain/enum"
	"DiTing-Go/global"
	"github.com/goccy/go-json"
	"github.com/pkg/errors"
)

// SetString 设置字符串
func SetString(key string, value any) error {
	userByte, err := json.Marshal(value)
	if err = global.Rdb.Set(key, userByte, domainEnum.CacheTime).Err(); err != nil {
		return errors.New("redis set error")
	}
	return nil
}
