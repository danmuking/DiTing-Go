package req

type GetUserInfoByNameReq struct {
	// 用户名
	Name string `form:"name" binding:"required"`
}
