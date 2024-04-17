package main

import (
	"fmt"
	"github.com/go-redis/redis"
	"time"
)

func main() {
	Rdb := redis.NewClient(&redis.Options{
		Addr:     "150.158.151.30:14490",
		Password: "550210817@", // 密码
		DB:       0,            // 数据库
		PoolSize: 20,           // 连接池大小
	})
	// 直接执行命令获取错误
	Rdb.Set("test", "test", time.Hour)

	// 直接执行命令获取值
	value := Rdb.Get("test").Val()
	fmt.Println(value)
}
