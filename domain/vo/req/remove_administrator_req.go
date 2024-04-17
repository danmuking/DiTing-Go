package req

type RemoveAdministratorReq struct {
	// 房间ID
	RoomId int64 `json:"room_id" binding:"required"`
	// 移除用户ID
	RemoveUid int64 `json:"remove_uid" binding:"required"`
}
