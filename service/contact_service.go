package service

import (
	"DiTing-Go/dal"
	"DiTing-Go/dal/model"
	"DiTing-Go/domain/dto"
	"DiTing-Go/domain/enum"
	domainModel "DiTing-Go/domain/model"
	domainResp "DiTing-Go/domain/vo/resp"
	"DiTing-Go/global"
	cursorUtils "DiTing-Go/pkg/cursor"
	"DiTing-Go/pkg/resp"
	"context"
	"github.com/gin-gonic/gin"
	"log"
	"strconv"
)

func GetContactListService(c *gin.Context) {
	uid := c.GetInt64("uid")
	// 游标翻页
	// 默认值
	var cursor *string = nil
	var pageSize int = 20
	pageRequest := cursorUtils.PageReq{
		Cursor:   cursor,
		PageSize: pageSize,
	}
	if err := c.ShouldBindQuery(&pageRequest); err != nil { //ShouldBind()会自动推导
		resp.ErrorResponse(c, "参数错误")
		c.Abort()
		return
	}
	pageResp, err := GetContactList(uid, pageRequest)
	if err != nil {
		resp.ErrorResponse(c, "查询会话列表失败")
		c.Abort()
		return
	}
	resp.SuccessResponse(c, pageResp)
	c.Abort()
	return
}

func GetContactList(uid int64, pageRequest cursorUtils.PageReq) (*cursorUtils.PageResp, error) {
	db := dal.DB
	contact := make([]model.Contact, 0)
	condition := []interface{}{"uid=?", strconv.FormatInt(uid, 10)}
	pageResp, err := cursorUtils.Paginate(db, pageRequest, &contact, "active_time", false, condition...)
	if err != nil {
		log.Printf("查询会话列表失败: %s", err.Error())
		return nil, err
	}
	contactList := pageResp.Data.(*[]model.Contact)
	contactDtoList := make([]dto.ContactDto, 0)
	for _, contact := range *contactList {
		contactDto, err := getContactDto(contact)
		if err != nil {
			log.Printf("查询会话列表失败: %s", err.Error())
			return nil, err
		}
		contactDtoList = append(contactDtoList, *contactDto)
	}
	pageResp.Data = contactDtoList
	return pageResp, nil
}

// 获取会话dto
func getContactDto(contact model.Contact) (*dto.ContactDto, error) {
	ctx := context.Background()
	room := global.Query.Room
	roomQ := room.WithContext(ctx)
	contactDto := new(dto.ContactDto)
	contactDto.ID = contact.ID
	// 查询房间类型
	roomR, err := roomQ.Where(room.ID.Eq(contact.RoomID)).First()
	if err != nil {
		return nil, err
	}
	// 如果是个人会话，名称是对方的昵称
	if roomR.Type == enum.PERSONAL {
		// 查询好友房间信息
		roomFriend := global.Query.RoomFriend
		roomFriendQ := roomFriend.WithContext(ctx)
		roomFriendR, err := roomFriendQ.Where(roomFriend.RoomID.Eq(roomR.ID)).First()
		if err != nil {
			return nil, err
		}
		var friendUid int64
		if roomFriendR.Uid1 == contact.UID {
			friendUid = roomFriendR.Uid2
		} else {
			friendUid = roomFriendR.Uid1
		}
		user := global.Query.User
		userQ := user.WithContext(ctx)
		userR, err := userQ.Where(user.ID.Eq(friendUid)).First()
		if err != nil {
			return nil, err
		}
		contactDto.Name = userR.Name
		contactDto.Avatar = user.Avatar
		contactDto.LastTime = contact.ActiveTime

		// TODO: 支持多种消息
		msg := global.Query.Message
		msgQ := msg.WithContext(ctx)
		msgR, err := msgQ.Where(msg.ID.Eq(contact.LastMsgID)).First()
		message := domainModel.Message(*msgR)
		if err != nil {
			return nil, err
		}
		contactDto.LastMsg = message.GetContactMsg()
	}
	// TODO: 返回消息未读数
	// TODO: 群聊
	return contactDto, nil
}

func GetContactDetailService(c *gin.Context) {

	roomIdString, _ := c.GetQuery("room_id")
	roomId, _ := strconv.ParseInt(roomIdString, 10, 64)
	//TODO:参数校验
	// 游标翻页
	// 默认值
	var cursor *string = nil
	var pageSize int = 20
	pageRequest := cursorUtils.PageReq{
		Cursor:   cursor,
		PageSize: pageSize,
	}
	if err := c.ShouldBindQuery(&pageRequest); err != nil { //ShouldBind()会自动推导
		resp.ErrorResponse(c, "参数错误")
		c.Abort()
		return
	}
	pageResp, err := GetContactDetail(roomId, pageRequest)
	if err != nil {
		resp.ErrorResponse(c, "系统正忙，请稍后再试")
		c.Abort()
		return
	}
	resp.SuccessResponse(c, pageResp)
	c.Abort()
	return
}

func GetContactDetail(roomID int64, pageRequest cursorUtils.PageReq) (*cursorUtils.PageResp, error) {
	// 查询消息
	db := dal.DB
	msgs := make([]model.Message, 0)
	// TODO: 抽象成常量
	condition := []interface{}{"room_id=? AND status=?", strconv.FormatInt(roomID, 10), "0"}
	pageResp, err := cursorUtils.Paginate(db, pageRequest, &msgs, "create_time", false, condition...)
	if err != nil {
		log.Printf("查询消息失败: %s", err.Error())
		return nil, err
	}
	msgList := pageResp.Data.(*[]model.Message)
	userIdMap := make(map[int64]*int64)
	for _, msg := range *msgList {
		if userIdMap[msg.FromUID] == nil {
			userIdMap[msg.FromUID] = &msg.FromUID
		}
	}
	// 转换成列表
	userIdList := make([]int64, 0)
	for _, uid := range userIdMap {
		userIdList = append(userIdList, *uid)
	}
	// 查询用户信息
	ctx := context.Background()
	user := global.Query.User
	userQ := user.WithContext(ctx)
	users, err := userQ.Where(user.ID.In(userIdList...)).Find()
	if err != nil {
		log.Printf("查询用户失败: %s", err.Error())
		return nil, err
	}
	userMap := make(map[int64]*model.User)
	for _, user := range users {
		userMap[user.ID] = user
	}

	// 拼装结果
	msgRespList := make([]domainResp.MessageResp, 0)
	for _, msg := range *msgList {
		msgResp := domainResp.MessageResp{
			ID:         msg.ID,
			Content:    msg.Content,
			ReplyMsgID: msg.ReplyMsgID,
			GapCount:   msg.GapCount,
			Type:       msg.Type,
			CreateTime: msg.CreateTime,
			UserName:   userMap[msg.FromUID].Name,
			UserAvatar: userMap[msg.FromUID].Avatar,
		}
		msgRespList = append(msgRespList, msgResp)
	}
	pageResp.Data = msgRespList
	return pageResp, nil
}
