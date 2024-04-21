package enum

import "time"

const (
	CacheTime  = 7 * 24 * time.Hour
	Project    = "diting:"
	User       = Project + "user:"
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

	// 用户缓存
	UserCacheByID   = User + "%d"
	UserCacheByName = User + "%s"

	// 用户好友缓存
	UserFriendCacheByUidAndFriendUid = UserFriend + "%d_%d"

	// 好友申请缓存
	UserApplyCacheByUidAndFriendUid = UserApply + "%d_%d"

	// 会话缓存
	ContactCacheById = Contact + "%d"
)
