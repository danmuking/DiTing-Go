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
