package dto

import "time"

type GetGroupMemberDto struct {
	// 用户ID
	UID int64 `json:"uid" gorm:"column:id"`
	// 用户名
	Name string `json:"name"`
	// 头像
	Avatar string `json:"avatar"`
	// 用户状态
	ActiveStatus int32 `json:"activeStatus"`
	// 最后活跃时间
	LastOptTime time.Time `json:"lastOptTime"`
}
