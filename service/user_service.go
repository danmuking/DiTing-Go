package service

import (
	"DiTing-Go/dal/model"
	domainEnum "DiTing-Go/domain/enum"
	"DiTing-Go/domain/vo/req"
	"DiTing-Go/global"
	"DiTing-Go/logic"
	pkgResp "DiTing-Go/pkg/domain/vo/resp"
	_ "DiTing-Go/pkg/setting"
	"DiTing-Go/utils"
	"context"
	"github.com/pkg/errors"
)

// RegisterService 用户注册
func RegisterService(userReq req.UserRegisterReq) (pkgResp.ResponseData, error) {
	ctx := context.Background()

	if userReq.Username == "" || userReq.Password == "" || userReq.Phone == "" {
		global.Logger.Infof("用户名、密码和手机号不能为空")
		return pkgResp.ErrorResponseData("用户名、密码和手机号不能为空"), errors.New("Business Error")
	}
	if userReq.Captcha == "" {
		global.Logger.Infof("验证码不能为空")
		return pkgResp.ErrorResponseData("验证码不能为空"), errors.New("Business Error")
	}

	// 验证验证码
	if !logic.CheckCaptchaProcess(userReq.Phone, userReq.Captcha) {
		return pkgResp.ErrorResponseData("验证码错误"), errors.New("Business Error")
	}

	// 查到手机号，返回
	if logic.CheckPhoneInRedis(userReq.Phone) {
		global.Logger.Infof("手机号已存在: %s", userReq.Phone)
		return pkgResp.ErrorResponseData("手机号已存在"), errors.New("Business Error")
	}

	// 如果redis查不到,查数据库
	rst, err := logic.CheckPhoneInDB(ctx, userReq.Phone)
	if err != nil {
		return pkgResp.ErrorResponseData("系统错误，请稍后再试"), errors.New("Business Error")
	}
	if rst {
		global.Logger.Infof("手机号已存在: %s", userReq.Phone)
		return pkgResp.ErrorResponseData("手机号已存在"), errors.New("Business Error")
	}

	// 对密码进行md5加密
	password := utils.EncryptPassword(userReq.Password)

	// 创建用户
	newUser := model.User{
		Name:     userReq.Username,
		Password: password,
		Phone:    userReq.Phone,
		IPInfo:   "{}",
	}
	// 创建对象
	if err = logic.CreateUser(ctx, newUser); err != nil {
		return pkgResp.ErrorResponseData("系统繁忙，请稍后再试~"), errors.New("Business Error")
	}
	return pkgResp.SuccessResponseDataWithMsg("注册成功"), nil
}

// LoginService 用户登录
func LoginService(loginReq req.UserLoginReq) (pkgResp.ResponseData, error) {
	ctx := context.Background()
	if loginReq.LoginType == domainEnum.LoginByPassword {
		// 参数校验
		if loginReq.Phone == "" || loginReq.Password == "" {
			global.Logger.Infof("loginReq:%v, 用户名和密码不能为空", loginReq)
			return pkgResp.ErrorResponseData("用户名和密码不能为空"), errors.New("Business Error")
		}
		// 检查手机号和密码是否正确
		if !logic.CheckPassword(ctx, loginReq.Phone, loginReq.Password) {
			global.Logger.Infof("手机号或密码错误: %s", loginReq.Phone)
			return pkgResp.ErrorResponseData("用户名或密码错误"), errors.New("Business Error")
		}
	} else if loginReq.LoginType == domainEnum.LoginByPhoneCaptcha {
		// 参数校验
		if loginReq.Phone == "" || loginReq.Captcha == "" {
			global.Logger.Infof("loginReq:%v, 手机号和验证码不能为空", loginReq)
			return pkgResp.ErrorResponseData("手机号和验证码不能为空"), errors.New("Business Error")
		}
		// 检查验证码是否正确
		if !logic.CheckCaptchaProcess(loginReq.Phone, loginReq.Captcha) {
			global.Logger.Infof("手机号:%s, 验证码错误", loginReq.Phone)
			return pkgResp.ErrorResponseData("验证码错误"), errors.New("Business Error")
		}
	}

	////生成jwt
	//token, err := utils.GenerateToken(userR.ID)
	//if err != nil {
	//	global.Logger.Errorf("生成jwt失败 %v", err)
	//	return pkgResp.ErrorResponseData("系统繁忙，请稍后再试~"), errors.New("Business Error")
	//}
	//// 发送用户登录事件
	//userByte, err := json.Marshal(userR)
	//if err != nil {
	//	global.Logger.Errorf("json序列化失败 %v", err)
	//}
	//msg := &primitive.Message{
	//	Topic: domainEnum.UserLoginTopic,
	//	Body:  userByte,
	//}
	//_, _ = global.RocketProducer.SendSync(ctx, msg)
	//userResp := resp.UserLoginResp{
	//	Token:  token,
	//	Uid:    userR.ID,
	//	Name:   userR.Name,
	//	Avatar: userR.Avatar,
	//}
	return pkgResp.SuccessResponseDataWithMsg("登录成功"), nil
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
