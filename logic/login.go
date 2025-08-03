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
	// 参数验证
	if phone == "" || password == "" {
		global.Logger.Warnf("CheckPassword: phone or password is empty, phone: %s", phone)
		return false
	}

	// 密码长度必须大于6
	if len(password) < 6 {
		global.Logger.Warnf("CheckPassword: password too short, phone: %s, password length: %d", phone, len(password))
		return false
	}

	// 对密码进行md5加密
	encryptedPassword := utils.EncryptPassword(password)

	// 首先尝试从Redis获取用户ID
	userPhoneKey := utils.MakeUserPhoneKey(phone)
	userIdBytes, err := utils.GetValueFromRedis(userPhoneKey)

	if err == nil && len(userIdBytes) > 0 {

		userId := string(userIdBytes)
		global.Logger.Infof("CheckPassword: get userId from redis success, phone: %s, userId: %s", phone, userId)

		// 从Redis获取用户信息
		userInfo, err := GetUserInfo2Redis(userId)
		if err != nil {
			if errors.Is(err, redis.Nil) {
				global.Logger.Infof("CheckPassword: user info not found in redis, phone: %s, userId: %s", phone, userId)
			} else {
				global.Logger.Errorf("CheckPassword: get user info from redis failed, phone: %s, userId: %s, err: %v", phone, userId, err)
			}
			// Redis获取失败，继续查询数据库
		} else {
			// Redis获取成功，比较密码
			global.Logger.Infof("CheckPassword: get user info from redis success, phone: %s, userId: %s", phone, userId)
			return userInfo.Password == encryptedPassword
		}
	} else if errors.Is(err, redis.Nil) {
		global.Logger.Infof("CheckPassword: userId not found in redis, phone: %s", phone)
	} else {
		global.Logger.Errorf("CheckPassword: get userId from redis failed, phone: %s, err: %v", phone, err)
	}

	// Redis查询失败，从数据库查询
	global.Logger.Infof("CheckPassword: querying user from database, phone: %s", phone)
	userInfo, err := utils.QueryUserByPhone(ctx, phone)
	if err != nil {
		global.Logger.Errorf("CheckPassword: failed to query user by phone, phone: %s, err: %v", phone, err)
		return false
	}

	if userInfo == nil {
		global.Logger.Warnf("CheckPassword: user not found in database, phone: %s", phone)
		return false
	}

	// 比较密码
	passwordMatch := userInfo.Password == encryptedPassword
	global.Logger.Infof("CheckPassword: password check result, phone: %s, match: %v", phone, passwordMatch)

	return passwordMatch
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

func GetUserInfo2DBById(ctx context.Context, userId string) (model.User, error) {
	if userId == "" {
		return model.User{}, errors.New("userId is zero")
	}

	userInfo, err := utils.QueryUserByID(ctx, userId)
	if err != nil {
		global.Logger.Errorf("userId %s, failed to query user by phone: %v", userId, err)
		return model.User{}, err
	}

	global.Logger.Infof("userId %s, user info found in db: %v", userId, userInfo)
	return *userInfo, nil
}
