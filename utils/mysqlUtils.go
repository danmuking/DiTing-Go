package utils

import (
	"DiTing-Go/dal/model"
	"DiTing-Go/global"
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm"
)

// QueryUserByPhone 根据手机号查询用户
func QueryUserByPhone(ctx context.Context, phone string) (*model.User, error) {
	user := global.Query.User
	userQ := user.WithContext(ctx)
	rst, err := userQ.Where(user.Phone.Eq(phone)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			global.Logger.Errorf("user not found with phone: %s", phone)
			return nil, gorm.ErrRecordNotFound
		}
		return nil, fmt.Errorf("query user by phone error: %w", err)
	}
	return rst, nil
}
