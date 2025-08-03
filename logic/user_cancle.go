package logic

import (
	"DiTing-Go/dal/query"
	"DiTing-Go/domain/enum"
	"DiTing-Go/global"
	"DiTing-Go/utils"
	"context"
	"strconv"
)

// DeleteUserInfoFromRedis 删除用户缓存
func DeleteUserInfoFromRedis(userId string) error {
	if err := utils.DeleteValueFromRedis(userId); err != nil {
		global.Logger.Errorf("删除用户缓存失败: userId=%s, err=%v", userId, err)
		return err
	}
	return nil
}

// DeleteUserInfoFromDB 删除用户数据库,软删除
func DeleteUserInfoFromDB(ctx context.Context, userId string) error {
	user := query.User
	userQ := user.WithContext(ctx)
	userIdInt, err := strconv.ParseInt(userId, 10, 64)
	if err != nil {
		global.Logger.Errorf("转换用户ID失败: userId=%s, err=%v", userId, err)
		return err
	}
	if _, err := userQ.Where(user.ID.Eq(userIdInt)).Update(user.Status, enum.UserStatusCancel); err != nil {
		global.Logger.Errorf("删除用户数据库失败: userId=%s, err=%v", userId, err)
		return err
	}
	return nil
}
