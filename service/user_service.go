package service

import (
	"DiTing-Go/dal/model"
	"DiTing-Go/dal/query"
	"DiTing-Go/domain/enum"
	"DiTing-Go/domain/vo/req"
	"DiTing-Go/domain/vo/resp"
	"DiTing-Go/global"
	pkgResp "DiTing-Go/pkg/domain/vo/resp"
	_ "DiTing-Go/pkg/setting"
	"DiTing-Go/pkg/utils"
	"DiTing-Go/service/adapter"
	"context"
	"fmt"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/goccy/go-json"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// RegisterService 用户注册
func RegisterService(userReq req.UserRegisterReq) (pkgResp.ResponseData, error) {
	ctx := context.Background()
	user := global.Query.User
	userQ := user.WithContext(ctx)
	fun := func() (interface{}, error) {
		return userQ.Where(user.Name.Eq(userReq.Username)).First()
	}
	userR := model.User{}
	key := fmt.Sprintf(enum.UserCacheByName, userReq.Username)
	err := utils.GetData(key, &userR, fun)
	// 查到了
	if err == nil {
		return pkgResp.ErrorResponseData("用户名已存在"), errors.New("Business Error")
	}
	// 有error
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		global.Logger.Errorf("查询数据失败: %v", err)
		return pkgResp.ErrorResponseData("系统繁忙，请稍后再试~"), errors.New("Business Error")
	}
	// 创建用户
	newUser := model.User{
		Name:     userReq.Username,
		Password: userReq.Password,
		IPInfo:   "{}",
	}
	// 创建对象
	if err := userQ.Omit(user.OpenID).Create(&newUser); err != nil {
		return pkgResp.ErrorResponseData("系统繁忙，请稍后再试~"), errors.New("Business Error")
	}
	return pkgResp.SuccessResponseDataWithMsg("注册成功"), nil
}

// LoginService 用户登录
func LoginService(loginReq req.UserLoginReq) (pkgResp.ResponseData, error) {
	ctx := context.Background()
	user := query.User
	userQ := user.WithContext(ctx)
	// 查数据库
	// 检查密码是否正确
	fun := func() (interface{}, error) {
		return userQ.Where(user.Name.Eq(loginReq.UserName), user.Password.Eq(loginReq.Password)).First()
	}
	userR := model.User{}
	key := fmt.Sprintf(enum.UserCacheByName, loginReq.UserName)
	err := utils.GetData(key, &userR, fun)
	if err != nil {
		global.Logger.Errorf("查询数据失败: %v", err)
		return pkgResp.ErrorResponseData("用户名或密码错误"), errors.New("Business Error")
	}
	//生成jwt
	token, err := utils.GenerateToken(userR.ID)
	if err != nil {
		global.Logger.Errorf("生成jwt失败 %v", err)
		return pkgResp.ErrorResponseData("系统繁忙，请稍后再试~"), errors.New("Business Error")
	}
	// 发送用户登录事件
	userByte, err := json.Marshal(userR)
	if err != nil {
		global.Logger.Errorf("json序列化失败 %v", err)
	}
	msg := &primitive.Message{
		Topic: enum.UserLoginTopic,
		Body:  userByte,
	}
	_, _ = global.RocketProducer.SendSync(ctx, msg)
	userResp := resp.UserLoginResp{
		Token:  token,
		Uid:    userR.ID,
		Name:   userR.Name,
		Avatar: userR.Avatar,
	}
	return pkgResp.SuccessResponseData(userResp), nil
}

func GetUserInfoByNameService(uid int64, name string) (pkgResp.ResponseData, error) {
	ctx := context.Background()
	user := global.Query.User
	userQ := user.WithContext(ctx)
	userRList, err := userQ.Where(user.Name.Like(name + "%")).Limit(5).Find()
	if err != nil {
		global.Logger.Errorf("查询用户数据失败: %v", err)
		return pkgResp.ErrorResponseData("系统繁忙，请稍后再试~"), errors.New("Business Error")
	}
	uidList := make([]int64, 0)
	for _, userR := range userRList {
		uidList = append(uidList, userR.ID)
	}
	//	搜索好友关系
	userApply := global.Query.UserApply
	userApplyQ := userApply.WithContext(ctx)
	applyList, err := userApplyQ.Where(userApply.UID.Eq(uid), userApply.TargetID.In(uidList...)).Find()
	if err != nil {
		global.Logger.Errorf("查询好友关系失败: %v", err)
		return pkgResp.ErrorResponseData("系统繁忙，请稍后再试~"), errors.New("Business Error")
	}
	//	查询好友关系
	userFriend := global.Query.UserFriend
	userFriendQ := userFriend.WithContext(ctx)
	friendList, err := userFriendQ.Where(userFriend.UID.Eq(uid), userFriend.FriendUID.In(uidList...)).Find()
	if err != nil {
		global.Logger.Errorf("查询好友关系失败: %v", err)
		return pkgResp.ErrorResponseData("系统繁忙，请稍后再试~"), errors.New("Business Error")
	}
	userRespList := adapter.BuildUserInfoByNameResp(userRList, applyList, friendList)
	return pkgResp.SuccessResponseData(userRespList), nil
}
