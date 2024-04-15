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
	"bytes"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"github.com/jinzhu/copier"
	"io"
	"log"
	"sort"
	"strconv"
)

var q *query.Query = global.Query

// Register 用户注册
//
//	@Summary	用户注册
//	@Produce	json
//	@Param		name		body		string				true	"用户名"
//	@Param		password	body		string				true	"密码"
//	@Success	200			{object}	resp.ResponseData	"成功"
//	@Failure	500			{object}	resp.ResponseData	"内部错误"
//	@Router		/api/public/register [post]
func Register(c *gin.Context) {
	user := model.User{}
	if err := c.ShouldBind(&user); err != nil { //ShouldBind()会自动推导
		resp.ErrorResponse(c, "参数错误")
		c.Abort()
		return
	}

	u := query.User
	// 用户名是否已存在
	exist, _ := u.WithContext(context.Background()).Where(u.Name.Eq(user.Name)).First()
	if exist != nil {
		resp.ErrorResponse(c, "用户名已存在")
		c.Abort()
		return
	}

	// 创建对象
	err := u.WithContext(context.Background()).Omit(u.OpenID).Create(&user)
	if err != nil {
		resp.ErrorResponse(c, "系统正忙请稍后再试~")
		c.Abort()
		return
	}
	resp.SuccessResponseWithMsg(c, "注册成功")
	return
}

// Login 用户登录
//
//	@Summary	用户登录
//	@Produce	json
//	@Param		name		body		string				true	"用户名"
//	@Param		password	body		string				true	"密码"
//	@Success	200			{object}	resp.ResponseData	"成功"
//	@Failure	500			{object}	resp.ResponseData	"内部错误"
//	@Router		/api/public/login [post]
func Login(c *gin.Context) {
	login := model.User{}
	if err := c.ShouldBind(&login); err != nil { //ShouldBind()会自动推导
		resp.ErrorResponse(c, "参数错误")
		c.Abort()
		return
	}

	u := query.User
	// 检查密码是否正确
	user, _ := u.WithContext(context.Background()).Where(u.Name.Eq(login.Name), u.Password.Eq(login.Password)).First()
	if user == nil {
		resp.ErrorResponse(c, "用户名或密码错误")
		c.Abort()
		return
	}
	//生成jwt
	token, _ := utils.GenerateToken(user.ID)
	resp.SuccessResponse(c, token)
}

// ApplyFriend 添加好友
//
//	@Summary	添加好友
//	@Produce	json
//	@Param		uid	body		int					true	"好友uid"
//	@Param		msg	body		string				true	"验证消息"
//	@Success	200	{object}	resp.ResponseData	"成功"
//	@Failure	500	{object}	resp.ResponseData	"内部错误"
//	@Router		/api/contact/add [post]
func ApplyFriend(c *gin.Context) {
	uid := c.GetInt64("uid")
	applyReq := req.UserApplyReq{}
	data, err := c.GetRawData()
	if err != nil { //ShouldBind()会自动推导
		resp.ErrorResponse(c, "参数错误")
		c.Abort()
		return
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(data))

	if err := json.Unmarshal(data, &applyReq); err != nil {
		resp.ErrorResponse(c, "参数错误")
		c.Abort()
		return
	}
	friendUid := applyReq.Uid
	//检查用户是否存在
	user, err := query.User.WithContext(context.Background()).Where(query.User.ID.Eq(friendUid)).First()
	if user == nil {
		resp.ErrorResponse(c, "用户不存在")
		c.Abort()
		return
	}
	// 检查是否已经是好友关系
	if isFriend := isFriend(c, uid, friendUid); isFriend {
		resp.ErrorResponse(c, "好友已存在")
		c.Abort()
		return
	}
	// 检查是否已经发送过好友请求
	friendApply, err := query.UserApply.WithContext(context.Background()).Where(query.UserApply.UID.Eq(uid), query.UserApply.TargetID.Eq(friendUid)).First()
	if err != nil && err.Error() != "record not found" {
		resp.ErrorResponse(c, "参数错误")
		c.Abort()
		return
	}
	if friendApply != nil {
		resp.ErrorResponse(c, "已发送过好友请求，请等待对方同意")
		c.Abort()
		return
	}
	// 检查对方是否给我们发送过好友请求，如果是，直接同意
	apply, err := query.UserApply.WithContext(context.Background()).Where(query.UserApply.UID.Eq(friendUid), query.UserApply.TargetID.Eq(uid)).First()
	if err != nil && err.Error() != "record not found" {
		resp.ErrorResponse(c, "参数错误")
		c.Abort()
		return
	}
	if apply != nil {
		Agree(c)
		return
	}
	// 发送好友请求
	err = query.UserApply.WithContext(context.Background()).Create(&model.UserApply{
		UID:        uid,
		TargetID:   friendUid,
		Msg:        applyReq.Msg,
		Status:     enum.NO,
		ReadStatus: enum.NO,
	})
	if err != nil {
		resp.ErrorResponse(c, "参数错误")
		c.Abort()
		return
	}
	// 发送好友申请事件
	global.Bus.Publish(domainEnum.FriendApplyEvent, model.UserApply{
		UID:        uid,
		TargetID:   friendUid,
		Msg:        applyReq.Msg,
		Status:     enum.NO,
		ReadStatus: enum.NO,
	})

	resp.SuccessResponseWithMsg(c, "success")
	return
}

// DeleteFriendService 删除好友
//
//	@Summary	删除好友
//	@Produce	json
//	@Param		uid	body		int					true	"好友uid"
//	@Success	200	{object}	resp.ResponseData	"成功"
//	@Failure	500	{object}	resp.ResponseData	"内部错误"
//	@Router		/api/contact/delete [delete]
func DeleteFriendService(c *gin.Context) {
	uid := c.GetInt64("uid")
	friend := req.UidReq{}
	if err := c.ShouldBind(&friend); err != nil { //ShouldBind()会自动推导
		resp.ErrorResponse(c, "参数错误")
		return
	}
	friendUid := friend.Uid
	// 检查是否为好友
	if isFriend := isFriend(c, uid, friendUid); isFriend {
		tx := global.Query.Begin()
		userFriend := global.Query.UserFriend
		userFriendTx := tx.UserFriend.WithContext(context.Background())
		userApply := global.Query.UserApply
		userApplyTx := tx.UserApply.WithContext(context.Background())
		// TODO: 抽取为异步事件
		// 事务
		// 删除好友关系
		if _, err := userFriendTx.Where(userFriend.UID.Eq(uid), userFriend.FriendUID.Eq(friendUid)).Delete(); err != nil {
			if err := tx.Rollback(); err != nil {
				log.Println("事务回滚失败", err.Error())
				return
			}
			resp.ErrorResponse(c, "删除失败")
			c.Abort()
			log.Println("删除好友失败", err.Error())
			return
		}
		if _, err := userFriendTx.Where(userFriend.UID.Eq(friendUid), userFriend.FriendUID.Eq(uid)).Delete(); err != nil {
			if err := tx.Rollback(); err != nil {
				log.Println("事务回滚失败", err.Error())
				return
			}
			resp.ErrorResponse(c, "删除失败")
			c.Abort()
			log.Println("删除好友失败", err.Error())
			return
		}
		// 删除好友申请
		if _, err := userApplyTx.Where(userApply.UID.Eq(uid), userApply.TargetID.Eq(friendUid)).Delete(); err != nil {
			if err := tx.Rollback(); err != nil {
				log.Println("事务回滚失败", err.Error())
				return
			}
			resp.ErrorResponse(c, "删除失败")
			c.Abort()
			log.Println("删除好友失败", err.Error())
			return
		}
		if _, err := userApplyTx.Where(userApply.UID.Eq(friendUid), userApply.TargetID.Eq(uid)).Delete(); err != nil {
			if err := tx.Rollback(); err != nil {
				log.Println("事务回滚失败", err.Error())
				return
			}
			resp.ErrorResponse(c, "删除失败")
			c.Abort()
			log.Println("删除好友失败", err.Error())
			return
		}

		// 删除好友房间
		roomFriend := global.Query.RoomFriend
		roomFriendTx := tx.RoomFriend.WithContext(context.Background())
		uids := utils.Int64Slice{uid, friendUid}
		sort.Sort(uids)
		roomFriendR, err := roomFriendTx.Where(roomFriend.Uid1.Eq(uids[0]), roomFriend.Uid2.Eq(uids[1])).First()
		if err != nil {
			if err := tx.Rollback(); err != nil {
				log.Println("事务回滚失败", err.Error())
				return
			}
			resp.ErrorResponse(c, "删除失败")
			c.Abort()
			log.Println("查询好友房间失败", err.Error())
			return
		}
		if _, err := roomFriendTx.Where(roomFriend.ID.Eq(roomFriendR.ID)).Delete(); err != nil {
			if err := tx.Rollback(); err != nil {
				log.Println("事务回滚失败", err.Error())
				return
			}
			resp.ErrorResponse(c, "删除失败")
			c.Abort()
			log.Println("删除好友房间失败", err.Error())
			return
		}

		// 删除房间表
		room := global.Query.Room
		roomTx := tx.Room.WithContext(context.Background())
		if _, err := roomTx.Where(room.ID.Eq(roomFriendR.RoomID)).Delete(); err != nil {
			if err := tx.Rollback(); err != nil {
				log.Println("事务回滚失败", err.Error())
				return
			}
			resp.ErrorResponse(c, "删除失败")
			c.Abort()
			log.Println("删除房间失败", err.Error())
			return
		}

		msg := global.Query.Message
		msgTx := tx.Message.WithContext(context.Background())
		// 删除消息
		if _, err := msgTx.Where(msg.RoomID.Eq(roomFriendR.RoomID)).Delete(); err != nil {
			if err := tx.Rollback(); err != nil {
				log.Println("事务回滚失败", err.Error())
				return
			}
			resp.ErrorResponse(c, "删除失败")
			c.Abort()
			log.Println("删除消息失败", err.Error())
			return
		}

		// 删除会话
		contact := global.Query.Contact
		contactTx := tx.Contact.WithContext(context.Background())
		if _, err := contactTx.Where(contact.RoomID.Eq(roomFriendR.RoomID)).Delete(); err != nil {
			if err := tx.Rollback(); err != nil {
				log.Println("事务回滚失败", err.Error())
				return
			}
			resp.ErrorResponse(c, "删除失败")
			c.Abort()
			log.Println("删除会话失败", err.Error())
			return
		}

		if err := tx.Commit(); err != nil {
			log.Println("事务提交失败", err.Error())
			resp.ErrorResponse(c, "删除失败")
			c.Abort()
			return
		}
	}
	resp.SuccessResponseWithMsg(c, "success")
	return
}

// IsFriend 是否为好友关系
//
//	@Summary	是否为好友关系
//	@Produce	json
//	@Param		uid	body		int					true	"好友uid"
//	@Success	200	{object}	resp.ResponseData	"成功"
//	@Failure	500	{object}	resp.ResponseData	"内部错误"
//	@Router		/api/contact/isFriend/:friendUid [get]
func IsFriend(c *gin.Context) {
	uid := c.GetInt64("uid")
	friendUid, _ := strconv.ParseInt(c.Param("friendUid"), 10, 64)
	resp.SuccessResponse(c, isFriend(c, uid, int64(friendUid)))
}

func isFriend(c *gin.Context, uid, friendUid int64) bool {
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
	// 获取好友列表
	userFriend := query.UserFriend
	// 获取 UserFriend 表中 uid = uid 的好友的uid组成的集合
	// select friend_uid from user_friend where uid = ?
	friendIDs, err := userFriend.WithContext(ctx).Select(userFriend.FriendUID).Where(userFriend.UID.Eq(uid)).Find()
	// 将friendIDs转换为切片
	// TODO 实现游标翻页
	friendIDsSlice := make([]int64, 0)
	for _, id := range friendIDs {
		friendIDsSlice = append(friendIDsSlice, id.FriendUID)
	}
	if err != nil {
		// todo 添加日志系统
		log.Printf("SQL查询错误, 错误信息为 : %v", err)
		resp.ErrorResponse(c, "出现错误，未能获取联系人列表")
		return
	}

	// 获取好友信息
	users := query.User
	// select id , name , avatar from user where id in (...) and status = 0
	friendList, err := users.WithContext(ctx).Select(users.ID, users.Name, users.Avatar).Where(users.ID.In(friendIDsSlice...), users.Status.Eq(0)).Find()
	if err != nil {
		// todo 添加日志系统
		log.Printf("SQL查询错误, 错误信息为 : %v", err)
		resp.ErrorResponse(c, "出现错误，未能获取联系人信息")
		return
	}

	// 数据转换
	friendListVO := make([]resp2.UserContactResp, 0)
	_ = copier.Copy(&friendListVO, &friendList)
	resp.SuccessResponse(c, friendListVO)
}
