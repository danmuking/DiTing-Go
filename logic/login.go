package logic

import (
	"DiTing-Go/dal/model"
	"DiTing-Go/domain/enum"
	"DiTing-Go/global"
	"DiTing-Go/utils"
	"context"
	"strconv"

	"github.com/go-redis/redis"
	"github.com/goccy/go-json"
	"github.com/pkg/errors"
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
		global.Logger.Infof("get userId from redis success, phone: %s, userId: %s", phone, userId)
		userInfo, err := GetUserInfo2Redis(userId)
		if errors.Is(err, redis.Nil) {
			global.Logger.Errorf("get user info from redis failed, phone: %s, userId: %s, err: %v", phone, userId, err)
		} else if err != nil {
			global.Logger.Errorf("get user info from redis failed, phone: %s, userId: %s, err: %v", phone, userId, err)
			return false
		} else {
			global.Logger.Infof("get user info from redis success, phone: %s, userId: %s, userInfo: %v", phone, userId, userInfo)
			return userInfo.Password == password
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

	return userInfo.Password == password
}

func SetUserInfo2Redis(userInfo model.User) error {
	userInfoByte, err := json.Marshal(userInfo)
	if err != nil {
		global.Logger.Errorf("userInfo %v,failed to marshal user info: %v", userInfo, err)
		return err
	}
	if err := utils.SetValueToRedis(strconv.FormatInt(userInfo.ID, 10), string(userInfoByte), enum.DefaultCacheTime); err != nil {
		global.Logger.Errorf("userInfo %v,failed to set user info to redis: %v", userInfo, err)
		return err
	}
	return nil
}

func GetUserInfo2Redis(userId string) (model.User, error) {
	if userId == "" || userId == "0" {
		return model.User{}, errors.New("userId is zero")
	}
	userInfoByte, err := utils.GetValueFromRedis(userId)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			global.Logger.Infof("userId %s, user info not found in redis", userId)
			return model.User{}, err
		}
		global.Logger.Errorf("userId %s, failed to get user info from redis: %v", userId, err)
		return model.User{}, err
	}
	userInfo := model.User{}
	if err := json.Unmarshal([]byte(userInfoByte), &userInfo); err != nil {
		global.Logger.Errorf("userId %s, failed to unmarshal user info: %v", userId, err)
		return model.User{}, err
	}

	global.Logger.Infof("userId %s, user info found in redis: %v", userId, userInfo)
	return userInfo, nil
}

func GetUserInfo2DB(phone string) (model.User, error) {
	if phone == "" {
		return model.User{}, errors.New("phone is empty")
	}

	ctx := context.Background()
	userInfo, err := utils.QueryUserByPhone(ctx, phone)
	if err != nil {
		global.Logger.Errorf("phone %s, failed to query user by phone: %v", phone, err)
		return model.User{}, err
	}

	global.Logger.Infof("phone %s, user info found in db: %v", phone, userInfo)
	return *userInfo, nil
}
