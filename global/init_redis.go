package global

import (
	"github.com/go-redis/redis"
	"github.com/spf13/viper"
)

var Rdb *redis.Client

func init() {
	Addr := viper.GetString("redis.host")
	Password := viper.GetString("redis.password")
	Rdb = redis.NewClient(&redis.Options{
		Addr:     Addr,
		Password: Password, // 密码
		DB:       0,        // 数据库
		PoolSize: 20,       // 连接池大小
	})
}
