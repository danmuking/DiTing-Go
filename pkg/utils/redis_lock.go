package utils

import (
	"DiTing-Go/global"
	"context"
	"github.com/go-redsync/redsync/v4"
	"github.com/pkg/errors"
)

// GetLock 获取分布式锁
func GetLock(key string) (*redsync.Mutex, error) {
	ctx := context.Background()
	mutex := global.RedSync.NewMutex(key)
	if err := mutex.LockContext(ctx); err != nil {
		global.Logger.Errorf("加锁失败 %s", err)
		return nil, errors.New("Business Error")
	}
	return mutex, nil
}

// ReleaseLock 释放分布式锁, 释放锁失败不影响业务
func ReleaseLock(mutex *redsync.Mutex) {
	_, err := mutex.Unlock()
	if err != nil {
		global.Logger.Errorf("解锁失败 %s", err)
	}
}
