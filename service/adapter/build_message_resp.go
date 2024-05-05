package adapter

import (
	"DiTing-Go/dal/model"
	"DiTing-Go/domain/vo/resp"
	pkgResp "DiTing-Go/pkg/domain/vo/resp"
)

func BuildMessageRespByMsgAndUser(pageResp *pkgResp.PageResp, msgList *[]model.Message, userMap map[int64]*model.User) *pkgResp.PageResp {
	var messageRespList []resp.MessageResp
	for i := range len(*msgList) {
		messageResp := resp.MessageResp{}
		msg := (*msgList)[i]
		msgUser := resp.MsgUser{}
		msgUser.Uid = userMap[msg.FromUID].ID
		msgUser.Username = userMap[msg.FromUID].Name
		msgUser.Avatar = userMap[msg.FromUID].Avatar
		messageResp.FromUser = msgUser

		message := resp.Msg{}
		message.ID = msg.ID
		message.RoomId = msg.RoomID
		message.Type = msg.Type
		message.Body.Content = msg.Content
		message.Body.Reply = msg.ReplyMsgID
		messageResp.Message = message

		messageResp.SendTime = msg.CreateTime.UnixNano()

		messageRespList = append(messageRespList, messageResp)
	}
	pageResp.Data = messageRespList
	return pageResp
}
