package global

import (
	goredislib "github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
	"github.com/spf13/viper"
)

var RedSync *redsync.Redsync

func init() {
	Addr := viper.GetString("redis.host")
	Password := viper.GetString("redis.password")
	client := goredislib.NewClient(&goredislib.Options{
		Addr:     Addr,
		Password: Password, // 密码
		DB:       0,        // 数据库
		PoolSize: 20,       // 连接池大小
	})
	pool := goredis.NewPool(client)
	RedSync = redsync.New(pool)
}
