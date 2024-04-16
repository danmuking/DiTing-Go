package service

import (
	"DiTing-Go/dal/model"
	"DiTing-Go/domain/dto"
	"DiTing-Go/domain/enum"
	"DiTing-Go/domain/vo/req"
	"DiTing-Go/global"
	"DiTing-Go/pkg/cursor"
	"DiTing-Go/pkg/resp"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"strconv"
	"strings"
	"time"
)

// CreateGroupService 创建群聊
//
//	@Summary	创建群聊
//	@Produce	json
//	@Param		name	body		string					true	"群聊名称"
//	@Success	200	{object}	resp.ResponseData	"成功"
//	@Failure	500	{object}	resp.ResponseData	"内部错误"
//	@Router		/api/group/create [post]
func CreateGroupService(c *gin.Context) {
	uid := c.GetInt64("uid")
	creatGroupReq := req.CreateGroupReq{}
	if err := c.ShouldBind(&creatGroupReq); err != nil { //ShouldBind()会自动推导
		resp.ErrorResponse(c, "参数错误")
		global.Logger.Errorf("参数错误: %v", err)
		c.Abort()
		return
	}

	tx := global.Query.Begin()
	defer func() {

	}()
	ctx := context.Background()
	// 创建房间表
	roomTx := tx.Room.WithContext(ctx)
	newRoom := model.Room{
		Type:    enum.GROUP,
		ExtJSON: "{}",
	}
	if err := roomTx.Create(&newRoom); err != nil {
		if err := tx.Rollback(); err != nil {
			global.Logger.Errorf("事务回滚失败 %s", err.Error())
			return
		}
		resp.ErrorResponse(c, "创建群聊失败")
		c.Abort()
		global.Logger.Errorf("添加房间表失败 %s", err.Error())
		return
	}

	// 查询用户头像
	user := global.Query.User
	userTx := tx.User.WithContext(ctx)
	userR, err := userTx.Where(user.ID.Eq(uid)).First()
	if err != nil {
		if err := tx.Rollback(); err != nil {
			global.Logger.Errorf("事务回滚失败 %s", err.Error())
			return
		}
		resp.ErrorResponse(c, "创建群聊失败")
		c.Abort()
		global.Logger.Errorf("查询用户表失败 %s", err.Error())
		return
	}

	// 创建群聊表
	roomGroupTx := tx.RoomGroup.WithContext(ctx)
	newRoomGroup := model.RoomGroup{
		RoomID: newRoom.ID,
		Name:   creatGroupReq.Name,
		// 默认为创建者头像
		Avatar:  userR.Avatar,
		ExtJSON: "{}",
	}
	if err := roomGroupTx.Create(&newRoomGroup); err != nil {
		if err := tx.Rollback(); err != nil {
			global.Logger.Errorf("事务回滚失败 %s", err.Error())
			return
		}
		resp.ErrorResponse(c, "创建群聊失败")
		c.Abort()
		global.Logger.Errorf("添加群聊表失败 %s", err.Error())
		return
	}

	groupMemberTx := tx.GroupMember.WithContext(ctx)
	newGroupMember := model.GroupMember{
		UID:     uid,
		GroupID: newRoomGroup.ID,
		// TODO: 1为群主,抽取为常量
		Role: 1,
	}
	if err := groupMemberTx.Create(&newGroupMember); err != nil {
		if err := tx.Rollback(); err != nil {
			global.Logger.Errorf("事务回滚失败 %s", err.Error())
			return
		}
		resp.ErrorResponse(c, "创建群聊失败")
		c.Abort()
		global.Logger.Errorf("添加群组成员表失败 %s", err.Error())
		return
	}
	// 自动发送一条消息
	messageTx := tx.Message.WithContext(ctx)
	newMessage := model.Message{
		FromUID: uid,
		RoomID:  newRoom.ID,
		Type:    enum.TextMessageType,
		Content: "欢迎加入群聊",
		Extra:   "{}",
	}
	if err := messageTx.Create(&newMessage); err != nil {
		if err := tx.Rollback(); err != nil {
			global.Logger.Errorf("事务回滚失败 %s", err.Error())
			return
		}
		resp.ErrorResponse(c, "创建群聊失败")
		c.Abort()
		global.Logger.Errorf("添加消息表失败 %s", err.Error())
		return

	}

	// 创建会话表
	contactTx := tx.Contact.WithContext(ctx)
	newContact := model.Contact{
		UID:        uid,
		RoomID:     newRoom.ID,
		ReadTime:   time.Now(),
		ActiveTime: time.Now(),
		LastMsgID:  newMessage.ID,
	}
	if err := contactTx.Create(&newContact); err != nil {
		if err := tx.Rollback(); err != nil {
			global.Logger.Errorf("事务回滚失败 %s", err.Error())
			return
		}
		resp.ErrorResponse(c, "创建群聊失败")
		c.Abort()
		global.Logger.Errorf("添加会话表失败 %s", err.Error())
		return
	}

	if err := tx.Commit(); err != nil {
		global.Logger.Errorf("事务提交失败 %s", err.Error())
		resp.ErrorResponse(c, "创建群聊失败")
		c.Abort()
		return
	}

	global.Bus.Publish(enum.NewMessageEvent, newMessage)

	resp.SuccessResponseWithMsg(c, "success")
	return
}

// DeleteGroupService 删除群聊
//
//	@Summary	删除群聊
//	@Produce	json
//	@Param		id	body		string					true	"房间ID"
//	@Success	200	{object}	resp.ResponseData	"成功"
//	@Failure	500	{object}	resp.ResponseData	"内部错误"
//	@Router		/api/group/:id [delete]
func DeleteGroupService(c *gin.Context) {
	uid := c.GetInt64("uid")
	deleteGroupReq := req.DeleteGroupReq{}
	if err := c.ShouldBindUri(&deleteGroupReq); err != nil { //ShouldBind()会自动推导
		resp.ErrorResponse(c, "参数错误")
		global.Logger.Errorf("参数错误: %v", err)
		c.Abort()
		return
	}

	tx := global.Query.Begin()
	defer func() {

	}()
	ctx := context.Background()
	// 查询群聊id
	roomGroup := global.Query.RoomGroup
	roomGroupTx := tx.RoomGroup.WithContext(ctx)
	roomGroupR, err := roomGroupTx.Where(roomGroup.RoomID.Eq(deleteGroupReq.ID)).First()
	if err != nil {
		if err := tx.Rollback(); err != nil {
			global.Logger.Errorf("事务回滚失败 %s", err.Error())
			return
		}
		resp.ErrorResponse(c, "删除群聊失败")
		c.Abort()
		global.Logger.Errorf("查询群聊表失败 %s", err.Error())
		return

	}
	// TODO:查询用户是否在群聊中
	groupMember := global.Query.GroupMember
	groupMemberTx := tx.GroupMember.WithContext(ctx)
	// 查询用户是否是群主
	_, err = groupMemberTx.Where(groupMember.UID.Eq(uid), groupMember.GroupID.Eq(roomGroupR.ID), groupMember.Role.Eq(1)).First()
	if err != nil {
		if err := tx.Rollback(); err != nil {
			global.Logger.Errorf("事务回滚失败 %s", err.Error())
			return
		}
		resp.ErrorResponse(c, "删除群聊失败")
		c.Abort()
		global.Logger.Errorf("查询群组成员表失败 %s", err.Error())
		return
	}
	// 获取群聊成员
	groupMemberList, err := groupMemberTx.Where(groupMember.GroupID.Eq(roomGroupR.ID)).Find()
	if err != nil {
		if err := tx.Rollback(); err != nil {
			global.Logger.Errorf("事务回滚失败 %s", err.Error())
			return
		}
		resp.ErrorResponse(c, "删除群聊失败")
		c.Abort()
		global.Logger.Errorf("查询群组成员表失败 %s", err.Error())
		return
	}

	// 删除所有成员的会话表
	for _, groupMember := range groupMemberList {
		contact := global.Query.Contact
		contactTx := tx.Contact.WithContext(ctx)
		if _, err := contactTx.Where(contact.UID.Eq(groupMember.UID), contact.RoomID.Eq(roomGroupR.ID)).Delete(); err != nil {
			if err := tx.Rollback(); err != nil {
				global.Logger.Errorf("事务回滚失败 %s", err.Error())
				return
			}
			resp.ErrorResponse(c, "删除群聊失败")
			c.Abort()
			global.Logger.Errorf("删除会话表失败 %s", err.Error())
			return
		}
	}

	// 删除群聊表
	if _, err := roomGroupTx.Where(roomGroup.RoomID.Eq(roomGroupR.ID)).Delete(); err != nil {
		if err := tx.Rollback(); err != nil {
			global.Logger.Errorf("事务回滚失败 %s", err.Error())
			return
		}
		resp.ErrorResponse(c, "删除群聊失败")
		c.Abort()
		global.Logger.Errorf("删除群聊表失败 %s", err.Error())
		return
	}
	// 删除房间表
	room := global.Query.Room
	roomTx := tx.Room.WithContext(ctx)
	if _, err := roomTx.Where(room.ID.Eq(roomGroupR.ID)).Delete(); err != nil {
		if err := tx.Rollback(); err != nil {
			global.Logger.Errorf("事务回滚失败 %s", err.Error())
			return
		}
		resp.ErrorResponse(c, "删除群聊失败")
		c.Abort()
		global.Logger.Errorf("删除房间表失败 %s", err.Error())
		return
	}
	// 删除群组成员表
	if _, err := groupMemberTx.Where(groupMember.GroupID.Eq(roomGroupR.ID)).Delete(); err != nil {
		if err := tx.Rollback(); err != nil {
			global.Logger.Errorf("事务回滚失败 %s", err.Error())
			return
		}
		resp.ErrorResponse(c, "删除群聊失败")
		c.Abort()
		global.Logger.Errorf("删除群组成员表失败 %s", err.Error())
		return
	}
	// TODO:抽取为事件
	// 删除消息表
	message := global.Query.Message
	messageTx := tx.Message.WithContext(ctx)
	msg := model.Message{
		Status: 0,
	}
	if _, err := messageTx.Where(message.RoomID.Eq(roomGroupR.ID)).Updates(msg); err != nil {
		if err := tx.Rollback(); err != nil {
			global.Logger.Errorf("事务回滚失败 %s", err.Error())
			return
		}
		resp.ErrorResponse(c, "删除群聊失败")
		c.Abort()
		global.Logger.Errorf("删除消息表失败 %s", err.Error())
		return
	}
	// TODO: 删除群聊仅禁止发送新消息，不删除消息
	if err := tx.Commit(); err != nil {
		global.Logger.Errorf("事务提交失败 %s", err.Error())
		resp.ErrorResponse(c, "删除群聊失败")
		c.Abort()
		return
	}
	resp.SuccessResponseWithMsg(c, "success")
	return
}

// JoinGroupService 加入群聊
//
//	@Summary	加入群聊
//	@Produce	json
//	@Param		id	body		int					true	"房间id"
//	@Success	200	{object}	resp.ResponseData	"成功"
//	@Failure	500	{object}	resp.ResponseData	"内部错误"
//	@Router		/api/group/create [post]
func JoinGroupService(c *gin.Context) {
	uid := c.GetInt64("uid")
	joinGroupReq := req.JoinGroupReq{}
	if err := c.ShouldBind(&joinGroupReq); err != nil { //ShouldBind()会自动推导
		resp.ErrorResponse(c, "参数错误")
		global.Logger.Errorf("参数错误: %v", err)
		c.Abort()
		return
	}
	ctx := context.Background()
	// 房间是否存在
	room := global.Query.Room
	roomQ := room.WithContext(ctx)
	roomR, err := roomQ.Where(room.ID.Eq(joinGroupReq.ID)).First()
	if err != nil {
		resp.ErrorResponse(c, "加入群聊失败")
		global.Logger.Errorf("查询房间失败 %s", err)
		c.Abort()
		return
	}
	if roomR.Type != enum.GROUP {
		resp.ErrorResponse(c, "加入群聊失败")
		global.Logger.Errorf("房间类型错误 %s", err)
		c.Abort()
		return
	}
	// 是否已经加入群聊
	roomGroup := global.Query.RoomGroup
	roomGroupQ := roomGroup.WithContext(ctx)
	// 查询群聊表
	roomGroupR, err := roomGroupQ.Where(roomGroup.RoomID.Eq(roomR.ID)).First()
	if err != nil {
		resp.ErrorResponse(c, "加入群聊失败")
		global.Logger.Errorf("查询群聊失败 %s", err)
		c.Abort()
		return
	}

	groupMember := global.Query.GroupMember
	groupMemberQ := groupMember.WithContext(ctx)
	groupMemberR, err := groupMemberQ.Where(groupMember.UID.Eq(uid), groupMember.GroupID.Eq(roomGroupR.ID)).First()
	if err != nil && err.Error() != "record not found" {
		resp.ErrorResponse(c, "加入群聊失败")
		global.Logger.Errorf("查询群成员表失败 %s", err)
		c.Abort()
		return
	}
	if groupMemberR != nil {
		resp.ErrorResponse(c, "禁止重复加入群聊")
		c.Abort()
		return
	}

	// 加入群聊
	tx := global.Query.Begin()
	groupMemberTx := tx.GroupMember.WithContext(ctx)
	newGroupMember := model.GroupMember{
		UID:     uid,
		GroupID: roomGroupR.ID,
		// 普通成员
		Role: 3,
	}
	if err := groupMemberTx.Create(&newGroupMember); err != nil {
		if err := tx.Rollback(); err != nil {
			global.Logger.Errorf("事务回滚失败 %s", err.Error())
			return
		}
		resp.ErrorResponse(c, "加入群聊失败")
		c.Abort()
		global.Logger.Errorf("添加群组成员表失败 %s", err.Error())
		return
	}
	// 创建会话表
	contactTx := tx.Contact.WithContext(ctx)
	newContact := model.Contact{
		UID:    uid,
		RoomID: roomR.ID,
	}
	if err := contactTx.Create(&newContact); err != nil {
		if err := tx.Rollback(); err != nil {
			global.Logger.Errorf("事务回滚失败 %s", err.Error())
			return
		}
		resp.ErrorResponse(c, "加入群聊失败")
		c.Abort()
		global.Logger.Errorf("添加会话表失败 %s", err.Error())
		return
	}

	// 自动发送一条消息
	messageTx := tx.Message.WithContext(ctx)
	newMessage := model.Message{
		FromUID: uid,
		RoomID:  roomR.ID,
		Type:    enum.TextMessageType,
		Content: "大家好~",
		Extra:   "{}",
	}
	if err := messageTx.Create(&newMessage); err != nil {
		if err := tx.Rollback(); err != nil {
			global.Logger.Errorf("事务回滚失败 %s", err.Error())
			return
		}
		resp.ErrorResponse(c, "加入群聊失败")
		c.Abort()
		global.Logger.Errorf("添加消息表失败 %s", err.Error())
		return
	}
	if err := tx.Commit(); err != nil {
		global.Logger.Errorf("事务提交失败 %s", err.Error())
		resp.ErrorResponse(c, "加入群聊失败")
		c.Abort()
		return
	}
	global.Bus.Publish(enum.NewMessageEvent, newMessage)

	resp.SuccessResponseWithMsg(c, "success")
	return
}

// QuitGroupService 退出群聊
//
//	@Summary	退出群聊
//	@Produce	json
//	@Param		id	body		int					true	"房间id"
//	@Success	200	{object}	resp.ResponseData	"成功"
//	@Failure	500	{object}	resp.ResponseData	"内部错误"
//	@Router		/api/group/create [post]
func QuitGroupService(c *gin.Context) {
	uid := c.GetInt64("uid")
	quitGroupReq := req.QuitGroupReq{}
	if err := c.ShouldBind(&quitGroupReq); err != nil {
		resp.ErrorResponse(c, "参数错误")
		global.Logger.Errorf("参数错误: %v", err)
		c.Abort()
		return
	}

	ctx := context.Background()
	tx := global.Query.Begin()
	// 群聊是否存在
	room := global.Query.Room
	roomTx := tx.Room.WithContext(ctx)
	_, err := roomTx.Where(room.ID.Eq(quitGroupReq.ID)).First()
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			global.Logger.Errorf("查询房间失败 %s", err)
		}
		resp.ErrorResponse(c, "群聊不存在")
		c.Abort()
		return
	}
	// 用户是否在群聊中
	groupMember := global.Query.GroupMember
	groupMemberTx := tx.GroupMember.WithContext(ctx)
	_, err = groupMemberTx.Where(groupMember.UID.Eq(uid), groupMember.GroupID.Eq(quitGroupReq.ID)).First()
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			resp.ErrorResponse(c, "未加入群聊")
			c.Abort()
			return
		}
		resp.ErrorResponse(c, "退出群聊失败")
		global.Logger.Errorf("查询群组成员表失败 %s", err)
		c.Abort()
		return
	}
	// 删除会话表
	contact := global.Query.Contact
	contactTx := tx.Contact.WithContext(ctx)
	if _, err := contactTx.Where(contact.UID.Eq(uid), contact.RoomID.Eq(quitGroupReq.ID)).Delete(); err != nil {
		resp.ErrorResponse(c, "退出群聊失败")
		global.Logger.Errorf("删除会话表失败 %s", err)
		c.Abort()
		return
	}
	// 删除群组成员表
	// 查询群组
	roomGroup := global.Query.RoomGroup
	roomGroupTx := tx.RoomGroup.WithContext(ctx)
	roomGroupR, err := roomGroupTx.Where(roomGroup.RoomID.Eq(quitGroupReq.ID)).First()
	if err != nil {
		resp.ErrorResponse(c, "退出群聊失败")
		global.Logger.Errorf("查询群聊失败 %s", err)
		c.Abort()
		return
	}
	if _, err := groupMemberTx.Where(groupMember.UID.Eq(uid), groupMember.GroupID.Eq(roomGroupR.ID)).Delete(); err != nil {
		resp.ErrorResponse(c, "退出群聊失败")
		global.Logger.Errorf("删除群组成员表失败 %s", err)
		c.Abort()
		return
	}

	if err := tx.Commit(); err != nil {
		global.Logger.Errorf("事务提交失败 %s", err)
		resp.ErrorResponse(c, "退出群聊失败")
		c.Abort()
		return
	}
	resp.SuccessResponseWithMsg(c, "success")
	return
}

// GetGroupMemberListService 退出群聊
//
//	@Summary	退出群聊
//	@Produce	json
//	@Param		id	body		int					true	"房间id"
//	@Success	200	{object}	resp.ResponseData	"成功"
//	@Failure	500	{object}	resp.ResponseData	"内部错误"
//	@Router		/api/group/getGroupMemberList [get]
func GetGroupMemberListService(c *gin.Context) {
	uid := c.GetInt64("uid")
	getGroupMemberListReq := req.GetGroupMemberListReq{}
	if err := c.ShouldBindQuery(&getGroupMemberListReq); err != nil {
		resp.ErrorResponse(c, "参数错误")
		global.Logger.Errorf("参数错误: %v", err)
		c.Abort()
		return
	}
	ctx := context.Background()

	// 查询房间表
	room := global.Query.Room
	roomQ := room.WithContext(ctx)
	roomR, err := roomQ.Where(room.ID.Eq(getGroupMemberListReq.ID)).First()
	if err != nil {
		resp.ErrorResponse(c, "查询群聊失败")
		global.Logger.Errorf("查询房间失败 %s", err)
		c.Abort()
		return
	}
	// 查询群聊表
	roomGroup := global.Query.RoomGroup
	roomGroupQ := roomGroup.WithContext(ctx)
	roomGroupR, err := roomGroupQ.Where(roomGroup.RoomID.Eq(roomR.ID)).First()
	if err != nil {
		resp.ErrorResponse(c, "查询群聊失败")
		global.Logger.Errorf("查询群聊失败 %s", err)
		c.Abort()
		return
	}

	// 查询用户是否在群聊中
	groupMember := global.Query.GroupMember
	groupMemberQ := groupMember.WithContext(ctx)
	_, err = groupMemberQ.Where(groupMember.UID.Eq(uid), groupMember.GroupID.Eq(roomGroupR.ID)).First()
	if err != nil {
		resp.ErrorResponse(c, "查询群聊失败")
		global.Logger.Errorf("查询群组成员表失败 %s", err)
		c.Abort()
		return
	}

	// 分页查询
	// 默认值
	// 划分游标，status_activateTime
	var userR []dto.GetGroupMemberDto
	status, activeTime := cursorSplit(getGroupMemberListReq.Cursor)
	// 查询群组成员表,联表游标翻页
	user := global.Query.User
	userQ := user.WithContext(ctx)
	if err := userQ.Select(user.Name, user.Avatar, user.ActiveStatus, user.LastOptTime).LeftJoin(groupMemberQ, user.ID.EqCol(groupMember.UID)).Where(groupMember.GroupID.Eq(roomGroupR.ID), user.ActiveStatus.Eq(int32(status)), user.LastOptTime.Gt(activeTime)).Limit(getGroupMemberListReq.PageSize).Scan(&userR); err != nil {
		resp.ErrorResponse(c, "查询群聊失败")
		global.Logger.Errorf("查询群组成员表失败 %s", err)
		c.Abort()
		return
	}
	// 不足，用不在线的补充
	if len(userR) < getGroupMemberListReq.PageSize && status == 1 {
		var add []dto.GetGroupMemberDto
		if err := userQ.Select(user.Name, user.Avatar, user.ActiveStatus, user.LastOptTime).LeftJoin(groupMemberQ, user.ID.EqCol(groupMember.UID)).Where(groupMember.GroupID.Eq(roomGroupR.ID), user.ActiveStatus.Eq(2)).Limit(getGroupMemberListReq.PageSize - len(userR)).Scan(&add); err != nil {
			resp.ErrorResponse(c, "查询群聊失败")
			global.Logger.Errorf("查询群组成员表失败 %s", err)
			c.Abort()
			return
		}
		userR = append(userR, add...)
	}

	newCursor := genCursor(userR)
	resp.SuccessResponse(c, cursor.PageResp{
		Cursor: &newCursor,
		IsLast: len(userR) < getGroupMemberListReq.PageSize,
		Data:   userR,
	})
}

// 分割游标
func cursorSplit(cursor *string) (int, time.Time) {
	if cursor == nil {
		return 1, time.Time{}
	}
	// TODO： 抽取为常量
	lines := strings.Split(*cursor, "_")
	status, err := strconv.Atoi(lines[0])
	if err != nil {
		return 1, time.Time{}
	}
	timeUnix, err := strconv.ParseInt(lines[1], 10, 64)
	if err != nil {
		return 1, time.Time{}
	}
	activeTime := time.Unix(timeUnix, timeUnix%1000000000)
	return status, activeTime
}

func genCursor(users []dto.GetGroupMemberDto) string {
	if users == nil || len(users) == 0 {
		return fmt.Sprintf("%d_%d", 2, time.Now().Unix())
	}
	status := users[len(users)-1].ActiveStatus
	activeTime := users[len(users)-1].LastOptTime
	newCursor := fmt.Sprintf("%d_%d", status, activeTime.Unix())
	return newCursor
}
