package redisCache

import (
	"DiTing-Go/dal/model"
	"DiTing-Go/domain/enum"
	"DiTing-Go/pkg/utils"
	"fmt"
)

// RemoveRoomCache 移除房间缓存
func RemoveRoomCache(room model.Room) {
	utils.RemoveData(fmt.Sprintf(enum.RoomCacheByID, room.ID))
}

// RemoveRoomFriend 移除房间好友缓存
func RemoveRoomFriend(roomFriend model.RoomFriend) {
	utils.RemoveData(fmt.Sprintf(enum.RoomFriendCacheByRoomID, roomFriend.RoomID))
	utils.RemoveData(fmt.Sprintf(enum.RoomFriendCacheByUidAndFriendUid, roomFriend.Uid1, roomFriend.Uid2))
}

// RemoveUserCache 移除用户缓存
func RemoveUserCache(user model.User) {
	utils.RemoveData(fmt.Sprintf(enum.UserCacheByID, user.ID))
	utils.RemoveData(fmt.Sprintf(enum.UserCacheByName, user.Name))
}

// RemoveUserFriend 移除用户好友缓存
func RemoveUserFriend(uid, friendUid int64) {
	utils.RemoveData(fmt.Sprintf(enum.UserFriendCacheByUidAndFriendUid, uid, friendUid))
	utils.RemoveData(fmt.Sprintf(enum.UserFriendCacheByUidAndFriendUid, friendUid, uid))
}

// RemoveUserApply 移除用户好友申请缓存
func RemoveUserApply(uid, friendUid int64) {
	utils.RemoveData(fmt.Sprintf(enum.UserApplyCacheByUidAndFriendUid, uid, friendUid))
	utils.RemoveData(fmt.Sprintf(enum.UserApplyCacheByUidAndFriendUid, friendUid, uid))
}

// RemoveContact 移除会话缓存
func RemoveContact(contact model.Contact) {
	utils.RemoveData(fmt.Sprintf(enum.ContactCacheById, contact.ID))
}
