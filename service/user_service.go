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
	"DiTing-Go/pkg/enum"
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

// IsFriend 是否为好友关系
//
//	@Summary	是否为好友关系
//	@Produce	json
//	@Param		uid	body		int					true	"好友uid"
//	@Success	200	{object}	resp.ResponseData	"成功"
//	@Failure	500	{object}	resp.ResponseData	"内部错误"
//	@Router		/api/contact/isFriend/:friendUid [get]
func IsFriends(c *gin.Context) {
	uid := c.GetInt64("uid")
	friendUid, _ := strconv.ParseInt(c.Param("friendUid"), 10, 64)
	resp.SuccessResponse(c, isFriends(c, uid, int64(friendUid)))
}

func isFriends(c *gin.Context, uid, friendUid int64) bool {
	// 检查是否已经是好友关系
	friend, err := query.UserFriend.WithContext(context.Background()).Where(query.UserFriend.UID.Eq(uid), query.UserFriend.FriendUID.Eq(friendUid)).First()
	if err != nil && err.Error() != "record not found" {
		resp.ErrorResponse(c, "参数错误")
		c.Abort()
	}
	if friend == nil {
		return false
	}
	return true
}

// Agree 同意好友申请
//
//	@Summary	同意好友申请
//	@Produce	json
//	@Param		uid	body		int					true	"好友uid"
//	@Success	200	{object}	resp.ResponseData	"成功"
//	@Failure	500	{object}	resp.ResponseData	"内部错误"
//	@Router		/api/contact/delete [put]
func Agree(c *gin.Context) {
	ctx := context.Background()
	userApply := query.UserApply
	userApplyQ := userApply.WithContext(ctx)
	uid := c.GetInt64("uid")

	friend := req.UidReq{}
	if err := c.ShouldBind(&friend); err != nil { //ShouldBind()会自动推导
		resp.ErrorResponse(c, "参数错误")
		c.Abort()
		return
	}

	friendUid := friend.Uid
	// 检查是否存在好友申请且状态为待审批
	apply, err := userApplyQ.Where(userApply.UID.Eq(friendUid), userApply.TargetID.Eq(uid), userApply.Status.Eq(enum.NO)).First()
	if err != nil && err.Error() != "record not found" {
		resp.ErrorResponse(c, "参数错误")
		c.Abort()
	}
	if apply == nil {
		resp.ErrorResponse(c, "好友申请不存在")
		c.Abort()
		return
	}

	// 同意好友请求
	apply.Status = enum.YES
	// 事务
	tx := q.Begin()
	userApplyTx := tx.UserApply.WithContext(context.Background())
	userFriendTx := tx.UserFriend.WithContext(context.Background())
	if _, err = userApplyTx.Where(userApply.UID.Eq(friendUid), userApply.TargetID.Eq(uid)).Updates(apply); err != nil {
		if err := tx.Rollback(); err != nil {
			log.Println("事务回滚失败", err.Error())
		}
		resp.ErrorResponse(c, "系统正忙请稍后再试~")
		c.Abort()
		return
	}
	var userFriends = []*model.UserFriend{
		{
			UID:       uid,
			FriendUID: friendUid,
		},
		{
			UID:       friendUid,
			FriendUID: uid,
		},
	}
	if err = userFriendTx.Create(userFriends...); err != nil {
		if err := tx.Rollback(); err != nil {
			log.Println("事务回滚失败", err.Error())
		}
		resp.ErrorResponse(c, "系统正忙请稍后再试~")
		c.Abort()
		return
	}
	if err := tx.Commit(); err != nil {
		resp.ErrorResponse(c, "系统正忙请稍后再试~")
		c.Abort()
		log.Println("事务提交失败", err.Error())
		return
	}
	// 发送新好友事件
	global.Bus.Publish(domainEnum.FriendNewEvent, model.UserFriend{
		UID:       uid,
		FriendUID: friendUid,
	})
	resp.SuccessResponseWithMsg(c, "success")
	return
}

// GetApplyList 获取用户好友申请列表
//
//		@Summary	获取用户好友申请列表
//	    @Produce 	json
//		@Security ApiKeyAuth
//		@Param		last_id	query	int					true	"last_id"
//		@Param		limit	query	int					true	"limit"
//		@Success	200	{object}	resp.ResponseData	"成功"
//		@Failure	500	{object}	resp.ResponseData	"内部错误"
//		@Router		/api/contact/getApplyList [get]
func GetApplyList(c *gin.Context) {

	ctx := context.Background()

	uid := c.GetInt64("uid")
	var n int

	u := query.User

	var uids = make([]int64, 0)
	var usersVO = make([]resp2.UserApplyResp, 0)

	// 默认值
	var cursor *string = nil
	var pageSize int = 20
	//pageSize, _ = strconv.Atoi(c.Query("page_size"))
	pageRequest := cursorUtils.PageReq{
		Cursor:   cursor,
		PageSize: pageSize,
	}
	if err := c.ShouldBindQuery(&pageRequest); err != nil { //ShouldBind()会自动推导
		resp.ErrorResponse(c, "参数错误")
		return
	}

	// 获取 UserApply 表中 TargetID 等于 uid(登录用户ID)的用户ID集合，采用游标分页
	db := dal.DB
	userApplys := make([]model.UserApply, 0)
	condition := []interface{}{"target_id=?", strconv.FormatInt(uid, 10)}
	pageResp, err := cursorUtils.Paginate(db, pageRequest, &userApplys, "create_time", false, condition...)
	if err != nil {
		// todo 添加日志系统
		log.Printf("DB excete Sql happen [ERROR], err msg is : %v", err)
		resp.ErrorResponse(c, "系统繁忙，亲稍后再试")
		return
	}

	n = len(userApplys)
	for i := 0; i < n; i++ {
		uids = append(uids, userApplys[i].UID)
	}

	// 根据 uids 集合查询 User 表
	users, err := u.WithContext(ctx).Select(u.ID, u.Name, u.Avatar).Where(u.ID.In(uids...)).Find()
	if err != nil {
		// todo 添加日志系统
		log.Printf("DB excete Sql happen [ERROR], err msg is : %v", err)
		resp.ErrorResponse(c, "系统繁忙，亲稍后再试")
		return
	}

	if len(users) != len(userApplys) {
		log.Printf("DB excete Sql happen [ERROR], err msg is : %v", err)
		resp.ErrorResponse(c, "系统繁忙，亲稍后再试")
	}
	// 数据转换
	for i := 0; i < len(users); i++ {
		var userVO resp2.UserApplyResp
		_ = copier.Copy(&userVO, &users[i])
		userVO.Msg = userApplys[i].Msg
		userVO.Status = userApplys[i].Status
		usersVO = append(usersVO, userVO)
	}
	pageResp.Data = usersVO

	resp.SuccessResponse(c, pageResp)
}

// UnreadApplyNum 好友申请未读数量
//
//	@Summary	好友申请未读数量
//	@Success	200			{object}	resp.ResponseData	"成功"
//	@Failure	500			{object}	resp.ResponseData	"内部错误"
//	@Router		/api/public/login [post]
func UnreadApplyNum(c *gin.Context) {
	ctx := context.Background()

	uid := c.GetInt64("uid")

	ua := query.UserApply

	// 获取 UserApply 表中 TargetID 等于 uid(登录用户ID)的用户ID集合
	num, err := ua.WithContext(ctx).Where(ua.TargetID.Eq(uid), ua.ReadStatus.Eq(enum.NO), ua.Status.Eq(enum.NO)).Count()
	if err != nil {
		log.Printf("DB excete Sql happen [ERROR], err msg is : %v", err)
		resp.ErrorResponse(c, "系统繁忙，亲稍后再试")
		c.Abort()
		return
	}
	resp.SuccessResponse(c, num)
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
