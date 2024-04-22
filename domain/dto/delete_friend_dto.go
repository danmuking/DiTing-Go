package dto

type DeleteFriendDto struct {
	Uid       int64 `json:"uid"`
	FriendUid int64 `json:"friend_uid"`
}
