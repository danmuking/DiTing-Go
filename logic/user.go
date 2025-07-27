package logic

import (
	"DiTing-Go/dal/model"
	"DiTing-Go/dal/query"
	"DiTing-Go/global"
	"context"
)

func CreateUser(ctx context.Context, userInfo model.User) error {
	user := query.User
	userQ := user.WithContext(ctx)
	// 创建对象
	if err := userQ.Create(&userInfo); err != nil {
		global.Logger.Errorf("user:%v, 创建用户失败: %v", user, err)
		return err
	}
	return nil
}
