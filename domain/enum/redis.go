package enum

import "time"

const (
	Project    = "diting:"
	User       = Project + "User"
	UserFriend = Project + "userFriend:"
	UserApply  = Project + "userApply:"
	RoomFriend = Project + "roomFriend:"
	Contact    = Project + "contact:"
	Room       = Project + "room:"
)
const (
	// 房间缓存
	RoomCacheByID = Room + "%d"

	// 好友房间缓存
	RoomFriendCacheByRoomID          = RoomFriend + "%d"
	RoomFriendCacheByUidAndFriendUid = RoomFriend + "%d_%d"

	// phoneUid映射
	PhoneUidMap      = User + "PhoneUid:" + "%s"
	UserCacheByID    = User + "%d"
	UserCacheByName  = User + "%s"
	UserCacheByPhone = User + "Phone:" + "%s"
	UserCaptcha      = User + "Captcha:" + "%s"

	// 用户好友缓存
	UserFriendCacheByUidAndFriendUid = UserFriend + "%d_%d"

	// 好友申请缓存
	UserApplyCacheByUidAndFriendUid = UserApply + "%d_%d"

	// 会话缓存
	ContactCacheById = Contact + "%d"
)

const (
	DefaultCacheTime = 7 * 24 * time.Hour
	NotExpireTime    = 10 * 365 * 24 * time.Hour
)
