package enum

const (
	Lock              = "lock:"
	UserLock          = Lock + "diting-user:"
	UserAndFriendLock = UserLock + "%d_%d"
)
