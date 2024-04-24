package service

import (
	"DiTing-Go/dal"
	"DiTing-Go/dal/model"
	"DiTing-Go/dal/query"
	domainEnum "DiTing-Go/domain/enum"
	"DiTing-Go/domain/vo/req"
	resp2 "DiTing-Go/domain/vo/resp"
	"DiTing-Go/global"
	cursorUtils "DiTing-Go/pkg/cursor"
	"DiTing-Go/pkg/resp"
	_ "DiTing-Go/pkg/setting"
	"DiTing-Go/pkg/utils"
	"context"
	"fmt"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"log"
	"strconv"
)

var q *query.Query = global.Query

// RegisterService 用户注册
func RegisterService(userReq req.UserRegisterReq) (resp.ResponseData, error) {
	ctx := context.Background()
	user := global.Query.User
	userQ := user.WithContext(ctx)
	fun := func() (interface{}, error) {
		return userQ.Where(user.Name.Eq(userReq.Name)).First()
	}
	userR := model.User{}
	key := fmt.Sprintf(domainEnum.UserCacheByName, userReq.Name)
	err := utils.GetData(key, &userR, fun)
	// 查到了
	if err == nil {
		return resp.ErrorResponseData("用户已存在"), errors.New("Business Error")
	}
	// 有error
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		global.Logger.Errorf("查询数据失败: %v", err)
		return resp.ErrorResponseData("系统繁忙，请稍后再试~"), errors.New("Business Error")
	}
	// 创建用户
	newUser := model.User{
		Name:     userReq.Name,
		Password: userReq.Password,
		IPInfo:   "{}",
	}
	// 创建对象
	if err := userQ.Omit(user.OpenID).Create(&newUser); err != nil {
		return resp.ErrorResponseData("系统繁忙，请稍后再试~"), errors.New("Business Error")
	}
	return resp.SuccessResponseDataWithMsg("success"), nil
}

// LoginService 用户登录
func LoginService(loginReq req.UserLoginReq) (resp.ResponseData, error) {
	ctx := context.Background()
	user := query.User
	userQ := user.WithContext(ctx)
	// 查数据库
	// 检查密码是否正确
	fun := func() (interface{}, error) {
		return userQ.Where(user.Name.Eq(loginReq.Name), user.Password.Eq(loginReq.Password)).First()
	}
	userR := model.User{}
	key := fmt.Sprintf(domainEnum.UserCacheByName, loginReq.Name)
	err := utils.GetData(key, &userR, fun)
	if err != nil {
		global.Logger.Errorf("查询数据失败: %v", err)
		return resp.ErrorResponseData("系统繁忙，请稍后再试~"), errors.New("Business Error")
	}
	//生成jwt
	token, err := utils.GenerateToken(userR.ID)
	if err != nil {
		global.Logger.Errorf("生成jwt失败 %v", err)
		return resp.ErrorResponseData("系统繁忙，请稍后再试~"), errors.New("Business Error")
	}
	// 发送用户登录事件
	userByte, err := json.Marshal(userR)
	if err != nil {
		global.Logger.Errorf("json序列化失败 %v", err)
	}
	msg := &primitive.Message{
		Topic: domainEnum.UserLoginTopic,
		Body:  userByte,
	}
	_, _ = global.RocketProducer.SendSync(ctx, msg)
	return resp.SuccessResponseData(token), nil
}

// GetFriendList 获取好友列表
//
//	@Summary	获取好友列表
//	@Produce	json
//	@Success	200	{object}	resp.ResponseData	"成功"
//	@Failure	500	{object}	resp.ResponseData	"内部错误"
//	@Router		/api/contact/getContactList [get]
func GetFriendList(c *gin.Context) {
	ctx := context.Background()
	uid := c.GetInt64("uid")

	// 从请求中获取cursor和pagesize两个参数
	cursor := c.Query("cursor")
	pagesize, _ := strconv.Atoi(c.Query("pagesize"))

	// 如果pagesize没有提供或者为0，我们设置一个默认值
	if pagesize == 0 {
		log.Println("pagesize is 0, set it to 20")
		pagesize = 20
	}

	pageRequest := cursorUtils.PageReq{
		Cursor:   &cursor,
		PageSize: pagesize,
	}

	if err := c.ShouldBindQuery(&pageRequest); err != nil {
		resp.ErrorResponse(c, "参数错误")
		return
	}

	// 获取 UserFriend 表中 uid = uid 的好友的uid组成的集合
	db := dal.DB
	userFriend := make([]model.UserFriend, 0)
	condition := []interface{}{"uid=?", strconv.FormatInt(uid, 10)}
	pageResp, err := cursorUtils.Paginate(db, pageRequest, &userFriend, "create_time", false, condition...)
	if err != nil {
		// todo 添加日志系统
		log.Printf("DB excete Sql happen [ERROR], err msg is : %v", err)
		resp.ErrorResponse(c, "系统繁忙，亲稍后再试")
		return
	}
	uids := make([]int64, 0)

	for _, friend := range userFriend {
		uids = append(uids, friend.FriendUID)
	}

	// 获取好友信息
	users := query.User
	// select id , name , avatar from user where id in (...) and status = 0
	friendList, err := users.WithContext(ctx).Select(users.ID, users.Name, users.Avatar).Where(users.ID.In(uids...)).Find()
	if err != nil {
		// todo 添加日志系统
		log.Printf("SQL查询错误, 错误信息为 : %v", err)
		resp.ErrorResponse(c, "出现错误，未能获取联系人信息")
		return
	}

	// 数据转换
	friendListVO := make([]resp2.UserContactResp, 0)
	_ = copier.Copy(&friendListVO, &friendList)
	pageResp.Data = friendListVO
	resp.SuccessResponse(c, pageResp)
}
