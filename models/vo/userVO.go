package vo

import "time"

type UserVo struct {
	ID           int64     `json:"ID"`            // 用户ID
	Name         string    `json:"name"`          // 用户昵称
	Avatar       string    `json:"avatar"`        // 用户头像
	Sex          int32     `json:"sex"`           // 性别 1为男性，2为女性
	ActiveStatus int32     `json:"active_status"` // 在线状态 1在线 2离线
	LastOptTime  time.Time `json:"last_opt_time"` // 最后上下线时间
}
