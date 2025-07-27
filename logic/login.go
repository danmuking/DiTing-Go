package logic

import (
	"DiTing-Go/dal/model"
	"DiTing-Go/domain/enum"
	"DiTing-Go/global"
	"DiTing-Go/utils"
	"context"
	"github.com/go-redis/redis"
	"github.com/goccy/go-json"
	"github.com/pkg/errors"
	"strconv"
)

// CheckPassword 校验用户名密码是否匹配
func CheckPassword(ctx context.Context, phone, password string) bool {
	// 密码长度必须大于6
	if len(password) < 6 {
		return false
	}

	// 对密码进行md5加密
	password = utils.EncryptPassword(password)

	// 查询映射
	userPhoneKey := utils.MakeUserPhoneKey(phone)
	userId, err := utils.GetValueFromRedis(userPhoneKey)
	if err == nil && userId != "" {
		// 从redis中查询用户信息
		userInfoByte, err := utils.GetValueFromRedis(userId)
		userInfo := &model.User{}
		if err == nil && userInfoByte != "" {
			if json.Unmarshal([]byte(userInfoByte), userInfo) == nil {
				return userInfo.Password == password
			} else {
				global.Logger.Errorf("userId %s,failed to unmarshal user info: %v", userId, err)
			}
		} else {
			global.Logger.Errorf("userId %s,failed to get user info from redis: %v", userId, err)
		}
	} else if errors.Is(err, redis.Nil) {
		global.Logger.Infof("get userId from redis failed, phone: %s, userId not found", phone)
	} else {
		global.Logger.Errorf("get userId from redis failed, phone: %s, err: %v", phone, err)
	}

	// 如果redis查不到,查数据库
	userInfo, err := utils.QueryUserByPhone(ctx, phone)
	if err != nil {
		global.Logger.Errorf("phone %s,failed to query user by phone: %v", phone, err)
		return false
	}

	// 更新缓存
	if err := utils.SetValueToRedis(userPhoneKey, strconv.FormatInt(userInfo.ID, 10), enum.CacheTime); err != nil {
		global.Logger.Errorf("phone %s,failed to set userId to redis: %v", phone, err)
	}
	userInfoByte, err := json.Marshal(userInfo)
	if err != nil {
		global.Logger.Errorf("userInfo %v,failed to marshal user info: %v", userInfo, err)
	}
	if err := utils.SetValueToRedis(strconv.FormatInt(userInfo.ID, 10), string(userInfoByte), enum.CacheTime); err != nil {
		global.Logger.Errorf("userInfo %v,failed to set user info to redis: %v", userInfo, err)
	}
	return userInfo.Password == password
}
