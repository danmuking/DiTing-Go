package utils

import (
	"DiTing-Go/global"
	"context"
	"fmt"
	"github.com/go-redsync/redsync/v4"
	"github.com/pkg/errors"
	"time"
)

type RedSyncLock struct {
	mutex  *redsync.Mutex
	ctx    context.Context
	cancel context.CancelFunc
}

// GetLock 获取分布式锁
func GetLock(key string) (*RedSyncLock, error) {
	ctx, cancel := context.WithCancel(context.Background())
	mutex := global.RedSync.NewMutex(key)
	if err := mutex.LockContext(ctx); err != nil {
		cancel()
		global.Logger.Errorf("加锁失败 %s", err)
		return nil, errors.New("Business Error")
	}
	// 开启一个goroutine，周期性地续租锁
	go func() {
		ticker := time.NewTicker(5 * time.Second) // 按照需求调整
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				ok, err := mutex.Extend()
				//fmt.Printf("续租锁 %s\n", ok)
				if err != nil {
					global.Logger.Errorf("Failed to extend lock: %s", err)
					return
				} else if !ok {
					global.Logger.Errorf("Failed to extend lock: %s", fmt.Errorf("lock %s is not held", key))
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return &RedSyncLock{
		mutex:  mutex,
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

// ReleaseLock 释放分布式锁, 释放锁失败不影响业务
func ReleaseLock(lock *RedSyncLock) {
	lock.cancel()
	mutex := lock.mutex
	_, err := mutex.Unlock()
	if err != nil {
		global.Logger.Errorf("解锁失败 %s", err)
	}
}
