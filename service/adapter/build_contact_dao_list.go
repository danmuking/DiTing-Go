package adapter

import (
	"DiTing-Go/dal/model"
	"DiTing-Go/domain/dto"
	"DiTing-Go/domain/enum"
	cmap "github.com/orcaman/concurrent-map/v2"
	"strconv"
)

type RoomDto struct {
	ID     int64
	Avatar string
	Name   string
	Type   int
}

func BuildContactDaoList(contactList []model.Contact, userList []*model.User, messageList []*model.Message, roomList []*model.Room, roomFriendList []*model.RoomFriend, roomGroupList []*model.RoomGroup, countMap cmap.ConcurrentMap[string, int64]) []dto.ContactDto {
	contactDtoList := make([]dto.ContactDto, 0)

	userMap := make(map[int64]*model.User)
	for _, user := range userList {
		userMap[user.ID] = user
	}

	msgMap := make(map[int64]*model.Message)
	for _, msg := range messageList {
		msgMap[msg.ID] = msg
	}

	roomFriendMap := make(map[int64]*model.RoomFriend)
	roomGroupMap := make(map[int64]*model.RoomGroup)
	for _, roomFriend := range roomFriendList {
		roomFriendMap[roomFriend.RoomID] = roomFriend
	}
	for _, roomGroup := range roomGroupList {
		roomGroupMap[roomGroup.RoomID] = roomGroup
	}

	roomMap := make(map[int64]RoomDto)
	for _, room := range roomList {
		roomDto := RoomDto{}
		roomDto.ID = room.ID
		if room.Type == enum.PERSONAL {
			userId := roomFriendMap[room.ID].Uid1
			if userId == contactList[0].UID {
				userId = roomFriendMap[room.ID].Uid2
			}
			user := userMap[userId]
			roomDto.Avatar = user.Avatar
			roomDto.Name = user.Name
			roomDto.Type = enum.PERSONAL
		} else {
			roomGroup := roomGroupMap[room.ID]
			roomDto.Avatar = roomGroup.Avatar
			roomDto.Name = roomGroup.Name
			roomDto.Type = enum.GROUP
		}
		roomMap[room.ID] = roomDto
	}
	for _, contact := range contactList {
		contactDto := dto.ContactDto{}
		contactDto.ID = contact.ID
		contactDto.RoomID = contact.RoomID
		contactDto.Avatar = roomMap[contact.RoomID].Avatar
		contactDto.Name = roomMap[contact.RoomID].Name
		contactDto.LastMsg = msgMap[contact.LastMsgID].Content
		contactDto.LastTime = contact.ActiveTime.UnixMilli()
		//TODO：统计未读消息数
		unreadCount, _ := countMap.Get(strconv.FormatInt(contact.RoomID, 10))
		contactDto.UnreadCount = int32(unreadCount)
		contactDto.Type = roomMap[contact.RoomID].Type
		contactDtoList = append(contactDtoList, contactDto)
	}
	return contactDtoList
}
