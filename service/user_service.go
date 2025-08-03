package service

import (
	"DiTing-Go/dal/model"
	domainEnum "DiTing-Go/domain/enum"
	"DiTing-Go/domain/vo/req"
	"DiTing-Go/domain/vo/resp"
	"DiTing-Go/global"
	"DiTing-Go/logic"
	pkgResp "DiTing-Go/pkg/domain/vo/resp"
	"DiTing-Go/utils"
	"DiTing-Go/utils/jwt"
	"context"
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/pkg/errors"
)

// RegisterService 用户注册
func RegisterService(userReq req.UserRegisterReq) (pkgResp.ResponseData, error) {
	ctx := context.Background()

	// 参数校验
	if err := validateRegisterRequest(userReq); err != nil {
		return pkgResp.ErrorResponseData(err.Error()), errors.New("Business Error")
	}

	// 验证验证码
	if err := validateRegisterCaptcha(userReq.Phone, userReq.Captcha); err != nil {
		return pkgResp.ErrorResponseData(err.Error()), errors.New("Business Error")
	}

	// 检查用户是否已存在
	exists, err := checkUserExists(ctx, userReq.Phone)
	if err != nil {
		global.Logger.Errorf("检查用户是否存在失败: phone=%s, err=%v", userReq.Phone, err)
		return pkgResp.ErrorResponseData("系统繁忙，请稍后再试~"), errors.New("Business Error")
	}
	if exists {
		return pkgResp.ErrorResponseData("手机号已存在"), errors.New("Business Error")
	}

	// 创建用户
	if err := createNewUser(ctx, userReq); err != nil {
		global.Logger.Errorf("创建用户失败: phone=%s, err=%v", userReq.Phone, err)
		return pkgResp.ErrorResponseData("系统繁忙，请稍后再试~"), errors.New("Business Error")
	}

	return pkgResp.SuccessResponseDataWithMsg("注册成功"), nil
}

// validateRegisterRequest 验证注册请求参数
func validateRegisterRequest(userReq req.UserRegisterReq) error {
	if userReq.Username == "" || userReq.Password == "" || userReq.Phone == "" {
		global.Logger.Infof("用户名、密码和手机号不能为空: userReq=%v", userReq)
		return errors.New("用户名、密码和手机号不能为空")
	}

	if userReq.Captcha == "" {
		global.Logger.Infof("验证码不能为空: phone=%s", userReq.Phone)
		return errors.New("验证码不能为空")
	}

	// 验证密码强度
	if len(userReq.Password) < 6 {
		return errors.New("密码长度不能少于6位")
	}

	// 验证手机号格式（简单验证）
	if len(userReq.Phone) != 11 {
		return errors.New("手机号格式不正确")
	}

	return nil
}

// validateRegisterCaptcha 验证注册验证码
func validateRegisterCaptcha(phone, captcha string) error {
	if !logic.CheckCaptchaProcess(phone, captcha) {
		global.Logger.Infof("验证码错误: phone=%s", phone)
		return errors.New("验证码错误")
	}
	return nil
}

// checkUserExists 检查用户是否已存在
func checkUserExists(ctx context.Context, phone string) (bool, error) {
	// 先检查Redis缓存
	if logic.CheckPhoneInRedis(phone) {
		global.Logger.Infof("手机号已存在: phone=%s", phone)
		return true, nil
	}

	// 检查数据库
	exists, err := logic.CheckPhoneInDB(ctx, phone)
	if err != nil {
		global.Logger.Errorf("检查用户是否存在失败: phone=%s, err=%v", phone, err)
		return false, err
	}

	if exists {
		global.Logger.Infof("手机号已存在: phone=%s", phone)
		return true, nil
	}

	return false, nil
}

// createNewUser 创建新用户
func createNewUser(ctx context.Context, userReq req.UserRegisterReq) error {
	// 对密码进行md5加密
	password := utils.EncryptPassword(userReq.Password)

	// 创建用户对象
	newUser := model.User{
		Name:     userReq.Username,
		Password: password,
		Phone:    userReq.Phone,
		IPInfo:   "{}",
	}

	// 创建用户
	if err := logic.CreateUser(ctx, newUser); err != nil {
		global.Logger.Errorf("创建用户失败: phone=%s, err=%v", userReq.Phone, err)
		return err
	}

	// 缓存用户信息到Redis
	if err := logic.SetUserInfo2Redis(newUser); err != nil {
		global.Logger.Errorf("缓存用户信息失败: userId=%d, err=%v", newUser.ID, err)
		// 不返回错误，因为用户创建成功，缓存失败不影响注册流程
	}

	// 缓存用户ID映射
	userPhoneKey := utils.MakeUserPhoneKey(userReq.Phone)
	if err := utils.SetValueToRedis(userPhoneKey, fmt.Sprintf("%d", newUser.ID), domainEnum.NotExpireTime); err != nil {
		global.Logger.Errorf("缓存用户ID映射失败: userId=%d, err=%v", newUser.ID, err)
		// 不返回错误，因为用户创建成功，缓存失败不影响注册流程
	}

	global.Logger.Infof("用户注册成功: phone=%s, userId=%d", userReq.Phone, newUser.ID)
	return nil
}

// LoginService 用户登录
func LoginService(loginReq req.UserLoginReq) (pkgResp.ResponseData, error) {
	ctx := context.Background()

	// 参数校验
	if err := validateLoginRequest(loginReq); err != nil {
		return pkgResp.ErrorResponseData(err.Error()), errors.New("Business Error")
	}

	// 验证登录凭据
	if err := validateLoginCredentials(ctx, loginReq); err != nil {
		return pkgResp.ErrorResponseData(err.Error()), errors.New("Business Error")
	}

	// 获取用户信息
	user, err := getUserInfo(loginReq.Phone)
	if err != nil {
		global.Logger.Errorf("获取用户信息失败: phone=%s, err=%v", loginReq.Phone, err)
		return pkgResp.ErrorResponseData("系统繁忙，请稍后再试~"), errors.New("Business Error")
	}

	// 生成JWT token
	token, err := jwt.GenerateToken(user.ID)
	if err != nil {
		global.Logger.Errorf("生成token失败: userId=%d, err=%v", user.ID, err)
		return pkgResp.ErrorResponseData("系统繁忙，请稍后再试~"), errors.New("Business Error")
	}

	// 构建响应
	userResp := resp.UserLoginResp{
		Token:  token,
		Uid:    user.ID,
		Name:   user.Name,
		Avatar: user.Avatar,
	}

	return pkgResp.SuccessResponseData(userResp), nil
}

// validateLoginRequest 验证登录请求参数
func validateLoginRequest(loginReq req.UserLoginReq) error {
	if loginReq.LoginType == domainEnum.LoginByPassword {
		if loginReq.Phone == "" || loginReq.Password == "" {
			global.Logger.Infof("用户名和密码不能为空: loginReq=%v", loginReq)
			return errors.New("用户名和密码不能为空")
		}
	} else {
		if loginReq.Phone == "" || loginReq.Captcha == "" {
			global.Logger.Infof("手机号和验证码不能为空: loginReq=%v", loginReq)
			return errors.New("手机号和验证码不能为空")
		}
	}
	return nil
}

// validateLoginCredentials 验证登录凭据
func validateLoginCredentials(ctx context.Context, loginReq req.UserLoginReq) error {
	if loginReq.LoginType == domainEnum.LoginByPassword {
		if !logic.CheckPassword(ctx, loginReq.Phone, loginReq.Password) {
			global.Logger.Infof("用户名或密码错误: phone=%s", loginReq.Phone)
			return errors.New("用户名或密码错误")
		}
	} else {
		if !logic.CheckCaptchaProcess(loginReq.Phone, loginReq.Captcha) {
			global.Logger.Infof("验证码错误: phone=%s", loginReq.Phone)
			return errors.New("验证码错误")
		}
	}
	return nil
}

// getUserInfo 获取用户信息，优先从Redis获取，失败则从数据库获取并缓存
func getUserInfo(phone string) (model.User, error) {
	userPhoneKey := utils.MakeUserPhoneKey(phone)

	// 从Redis获取用户ID
	userIdByte, err := utils.GetValueFromRedis(userPhoneKey)
	var userId int64
	json.Unmarshal(userIdByte, &userId)
	if err != nil && !errors.Is(err, redis.Nil) {
		global.Logger.Errorf("从Redis获取用户ID失败: userPhoneKey=%s, err=%v", userPhoneKey, err)
		return model.User{}, err
	}

	// 如果Redis中有用户ID，尝试获取用户信息
	if userId != 0 {
		user, err := logic.GetUserInfo2Redis(fmt.Sprintf("%d", userId))
		if err == nil {
			return user, nil
		}
		// Redis中用户信息不存在或出错，继续从数据库获取
		global.Logger.Infof("从Redis获取用户信息失败，尝试从数据库获取: userId=%s, err=%v", userId, err)
	}

	// 从数据库获取用户信息
	user, err := logic.GetUserInfo2DB(phone)
	if err != nil {
		global.Logger.Errorf("从数据库获取用户信息失败: phone=%s, err=%v", phone, err)
		return model.User{}, err
	}

	// 将用户信息缓存到Redis
	if err := logic.SetUserInfo2Redis(user); err != nil {
		global.Logger.Errorf("缓存用户信息到Redis失败: userId=%d, err=%v", user.ID, err)
	}

	// 缓存用户ID映射
	if err := utils.SetValueToRedis(userPhoneKey, fmt.Sprintf("%d", user.ID), domainEnum.NotExpireTime); err != nil {
		global.Logger.Errorf("缓存用户ID映射失败: userId=%d, err=%v", user.ID, err)
	}

	return user, nil
}

// CancelService 注销账户
func CancelService(ctx *gin.Context, req req.UserCancelReq) (pkgResp.ResponseData, error) {
	// 获取用户信息
	userId, exists := ctx.Get("uid")
	if !exists {
		global.Logger.Errorf("用户id不存在")
		return pkgResp.ErrorResponseData("系统繁忙，请稍后再试~"), errors.New("Business Error")
	}

	userIdNum, ok := userId.(int64)
	userIdStr := fmt.Sprintf("%d", userIdNum)

	if !ok {
		global.Logger.Errorf("获取用户ID失败: userId=%v", userId)
		return pkgResp.ErrorResponseData("系统繁忙，请稍后再试~"), errors.New("Business Error")
	}

	userInfo, err := logic.GetUserInfo2DBById(ctx, userIdStr)
	if err != nil {
		global.Logger.Errorf("获取用户信息失败: userId=%s, err=%v", userId, err)
		return pkgResp.ErrorResponseData("系统繁忙，请稍后再试~"), errors.New("Business Error")
	}

	phone := userInfo.Phone
	captcha := req.Captcha

	// 检查验证码是否正确
	if !logic.CheckCaptchaProcess(phone, captcha) {
		global.Logger.Infof("验证码错误: phone=%s", phone)
		return pkgResp.ErrorResponseData("验证码错误"), errors.New("验证码错误")
	}

	// 检查用户是否存在
	userExists, err := checkUserExists(ctx, phone)
	if err != nil {
		global.Logger.Errorf("检查用户是否存在失败: phone=%s, err=%v", phone, err)
		return pkgResp.ErrorResponseData("系统繁忙，请稍后再试~"), errors.New("Business Error")
	}
	if !userExists {
		global.Logger.Errorf("用户不存在: phone=%s", phone)
		return pkgResp.ErrorResponseData("用户不存在"), errors.New("Business Error")
	}

	// 删除用户缓存
	if err := logic.DeleteUserInfoFromRedis(userIdStr); err != nil {
		global.Logger.Errorf("删除用户缓存失败: phone=%s, err=%v", phone, err)
		return pkgResp.ErrorResponseData("系统繁忙，请稍后再试~"), errors.New("Business Error")
	}

	// 删除用户ID映射
	userPhoneKey := utils.MakeUserPhoneKey(phone)
	if err := utils.DeleteValueFromRedis(userPhoneKey); err != nil {
		global.Logger.Errorf("删除用户ID映射失败: phone=%s, err=%v", phone, err)
		return pkgResp.ErrorResponseData("系统繁忙，请稍后再试~"), errors.New("Business Error")
	}

	// 删除用户数据库
	if err := logic.DeleteUserInfoFromDB(ctx, userIdStr); err != nil {
		global.Logger.Errorf("删除用户数据库失败: phone=%s, err=%v", phone, err)
		return pkgResp.ErrorResponseData("系统繁忙，请稍后再试~"), errors.New("Business Error")
	}

	return pkgResp.SuccessResponseDataWithMsg("注销成功"), nil
}

//
//func GetUserInfoByNameService(uid int64, name string) (pkgResp.ResponseData, error) {
//	ctx := context.Background()
//	user := global.Query.User
//	userQ := user.WithContext(ctx)
//	userRList, err := userQ.Where(user.Name.Like(name + "%")).Limit(5).Find()
//	if err != nil {
//		global.Logger.Errorf("查询用户数据失败: %v", err)
//		return pkgResp.ErrorResponseData("系统繁忙，请稍后再试~"), errors.New("Business Error")
//	}
//	uidList := make([]int64, 0)
//	for _, userR := range userRList {
//		uidList = append(uidList, userR.ID)
//	}
//	//	搜索好友关系
//	userApply := global.Query.UserApply
//	userApplyQ := userApply.WithContext(ctx)
//	applyList, err := userApplyQ.Where(userApply.UID.Eq(uid), userApply.TargetID.In(uidList...)).Find()
//	if err != nil {
//		global.Logger.Errorf("查询好友关系失败: %v", err)
//		return pkgResp.ErrorResponseData("系统繁忙，请稍后再试~"), errors.New("Business Error")
//	}
//	//	查询好友关系
//	userFriend := global.Query.UserFriend
//	userFriendQ := userFriend.WithContext(ctx)
//	friendList, err := userFriendQ.Where(userFriend.UID.Eq(uid), userFriend.FriendUID.In(uidList...)).Find()
//	if err != nil {
//		global.Logger.Errorf("查询好友关系失败: %v", err)
//		return pkgResp.ErrorResponseData("系统繁忙，请稍后再试~"), errors.New("Business Error")
//	}
//	userRespList := adapter.BuildUserInfoByNameResp(userRList, applyList, friendList)
//	return pkgResp.SuccessResponseData(userRespList), nil
//}
