package global

import (
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
)

var RedSync *redsync.Redsync

func init() {
	client := Rdb
	pool := goredis.NewPool(client)
	RedSync = redsync.New(pool)
}
