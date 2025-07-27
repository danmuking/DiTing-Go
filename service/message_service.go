package service

//
//import (
//	"DiTing-Go/dal/model"
//	"DiTing-Go/domain/enum"
//	"DiTing-Go/domain/vo/req"
//	domainResp "DiTing-Go/domain/vo/resp"
//	"DiTing-Go/global"
//	pkgEnum "DiTing-Go/pkg/domain/enum"
//	"DiTing-Go/pkg/domain/vo/resp"
//	"context"
//	"github.com/apache/rocketmq-client-go/v2/primitive"
//	"github.com/goccy/go-json"
//	"log"
//	"time"
//)
//
//// SendTextMsgService 发送文本消息
//func SendTextMsgService(uid int64, msgReq req.MessageReq) (resp.ResponseData, error) {
//	ctx := context.Background()
//	user := global.Query.User
//	userQ := user.WithContext(ctx)
//	userR, err := userQ.Where(user.ID.Eq(uid)).First()
//	if err != nil {
//		log.Println("查询用户失败", err)
//		return resp.ErrorResponseData("消息发送失败"), err
//	}
//
//	msg := model.Message{}
//	msg.Type = msgReq.MsgType
//	msg.FromUID = uid
//	msg.RoomID = msgReq.RoomId
//	msg.Content = msgReq.Body.Content
//	if msg.Extra == "" {
//		msg.Extra = "{}"
//	}
//
//	// 发送消息
//	if err := SendTextMsg(&msg); err != nil {
//		return resp.ErrorResponseData("消息发送失败"), err
//	}
//	// 发送新消息事件
//	newMsgByte, _ := json.Marshal(msg)
//	rMsg := &primitive.Message{
//		Topic: enum.NewMessageTopic,
//		Body:  newMsgByte,
//	}
//	_, _ = global.RocketProducer.SendSync(ctx, rMsg)
//
//	msgResp := domainResp.MessageResp{
//		FromUser: domainResp.MsgUser{
//			Uid:      uid,
//			Username: userR.Name,
//			Avatar:   userR.Avatar,
//		},
//		SendTime: msg.CreateTime.UnixMilli(),
//		Message: domainResp.Msg{
//			ID:     msg.ID,
//			RoomId: msg.RoomID,
//			Type:   msg.Type,
//			Body: domainResp.TextBody{
//				Content: msg.Content,
//				Reply:   msg.ReplyMsgID,
//			},
//		},
//	}
//
//	// 返回成功
//	return resp.SuccessResponseData(msgResp), nil
//}
//
//func SendTextMsg(msg *model.Message) error {
//	msg.CreateTime = time.Now()
//	msg.DeleteStatus = pkgEnum.NORMAL
//	ctx := context.Background()
//	msgQ := global.Query.WithContext(ctx).Message
//	if err := msgQ.Create(msg); err != nil {
//		log.Println("消息发送失败", err.Error())
//		return err
//	}
//	return nil
//}
