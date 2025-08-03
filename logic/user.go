package logic

import (
	"DiTing-Go/dal/model"
	"DiTing-Go/dal/query"
	"DiTing-Go/domain/enum"
	"DiTing-Go/global"
	"context"
)

func CreateUser(ctx context.Context, userInfo model.User) error {
	user := query.User
	userQ := user.WithContext(ctx)

	userId := userInfo.ID
	// 检查用户是否存在
	userResult, err := userQ.Where(user.ID.Eq(userId)).First()
	if err != nil {
		global.Logger.Errorf("user:%v, 检查用户是否存在失败: %v", userResult, err)
		return err
	}
	// 将用户状态更新为正常
	if userResult.Status == int32(enum.UserStatusCancel) {
		global.Logger.Errorf("user:%v, 用户已注销", userResult)
		userInfo.Status = enum.UserStatusNormal
		if err := userQ.Save(&userInfo); err != nil {
			global.Logger.Errorf("user:%v, 更新用户状态失败: %v", userResult, err)
			return err
		}
	} else {
		// 创建对象
		if err := userQ.Create(&userInfo); err != nil {
			global.Logger.Errorf("user:%v, 创建用户失败: %v", userResult, err)
			return err
		}
	}
	return nil
}
