package adapter

import (
	"DiTing-Go/dal/model"
	"DiTing-Go/domain/vo/resp"
)

func BuildUserInfoByNameResp(userList []*model.User, userApply []*model.UserApply, userFriend []*model.UserFriend) []resp.GetUserInfoByNameResp {
	userApplyMap := make(map[int64]*model.UserApply)
	for _, apply := range userApply {
		userApplyMap[apply.TargetID] = apply
	}
	userFriendMap := make(map[int64]*model.UserFriend)
	for _, friend := range userFriend {
		userFriendMap[friend.FriendUID] = friend
	}
	getUserInfoByNameRespList := make([]resp.GetUserInfoByNameResp, 0)
	for _, user := range userList {
		getUserInfoByNameResp := resp.GetUserInfoByNameResp{
			Uid:    user.ID,
			Name:   user.Name,
			Avatar: user.Avatar,
		}

		//TODO:抽象为常量
		if userFriendMap[user.ID] != nil {
			getUserInfoByNameResp.Status = 3
		} else if userApplyMap[user.ID] != nil {
			getUserInfoByNameResp.Status = 2
		} else {
			getUserInfoByNameResp.Status = 1
		}
		getUserInfoByNameRespList = append(getUserInfoByNameRespList, getUserInfoByNameResp)
	}
	return getUserInfoByNameRespList
}
