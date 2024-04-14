package req

type CreateGroupReq struct {
	Name string `json:"name" binding:"required"`
}
