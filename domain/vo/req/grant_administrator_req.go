package req

type GrantAdministratorReq struct {
	// 房间ID
	RoomId int64 `json:"room_id" binding:"required"`
	// 授权用户ID
	GrantUid int64 `json:"grant_uid" binding:"required"`
}
