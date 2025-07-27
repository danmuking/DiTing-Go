package service

//
//import (
//	"DiTing-Go/dal/model"
//	"DiTing-Go/domain/dto"
//	"DiTing-Go/domain/enum"
//	voResp "DiTing-Go/domain/vo/resp"
//	"DiTing-Go/global"
//	"DiTing-Go/pkg/domain/vo/resp"
//	"context"
//	"fmt"
//	"github.com/gin-gonic/gin"
//	"github.com/goccy/go-json"
//	"github.com/minio/minio-go/v7"
//	"strconv"
//	"time"
//)
//
//// GetPreSigned 签发url
//func GetPreSigned(c *gin.Context) {
//	uid := c.GetInt64("uid")
//	ctx := context.Background()
//	roomIdStr, found := c.GetQuery("roomId")
//	if !found {
//		global.Logger.Errorf("参数错误 %s", roomIdStr)
//		resp.ErrorResponse(c, "参数错误")
//		c.Abort()
//		return
//	}
//	roomId, err := strconv.ParseInt(roomIdStr, 10, 64)
//	if err != nil {
//		global.Logger.Errorf("参数错误 %s", roomId)
//		resp.ErrorResponse(c, "参数错误")
//		c.Abort()
//		return
//	}
//
//	fileName, found := c.GetQuery("fileName")
//	if !found {
//		global.Logger.Errorf("参数错误 %s", roomIdStr)
//		resp.ErrorResponse(c, "参数错误")
//		c.Abort()
//		return
//	}
//	// 构造文件名：time+uid+filename
//	// 按天创建桶
//	timeStr := time.Now().Format("2006-01-02")
//	fileName = fmt.Sprintf("%s/%d/%s", timeStr, uid, fileName)
//
//	policy := minio.NewPostPolicy()
//	// TODO:抽象为常量
//	if err := policy.SetBucket("diting"); err != nil {
//		global.Logger.Errorf("创建policy失败 %s", roomIdStr)
//		resp.ErrorResponse(c, "获取签名失败，请稍后再试")
//		c.Abort()
//		return
//	}
//	if err := policy.SetKey(fileName); err != nil {
//		global.Logger.Errorf("创建policy失败 %s", roomIdStr)
//		resp.ErrorResponse(c, "获取签名失败，请稍后再试")
//		c.Abort()
//		return
//	}
//	// 失效时间1天
//	if err := policy.SetExpires(time.Now().UTC().AddDate(0, 0, 1)); err != nil {
//		global.Logger.Errorf("创建policy失败 %s", roomIdStr)
//		resp.ErrorResponse(c, "获取签名失败，请稍后再试")
//		c.Abort()
//		return
//	}
//	url, formData, err := global.MinioClient.PresignedPostPolicy(ctx, policy)
//	if err != nil {
//		global.Logger.Errorf("创建policy失败 %s", roomIdStr)
//		resp.ErrorResponse(c, "获取签名失败，请稍后再试")
//		c.Abort()
//		return
//	}
//	preSignedResp := voResp.PreSignedResp{
//		Url:    url.String(),
//		Policy: formData,
//	}
//	tx := global.Query.Begin()
//	// 插入消息表
//	messageTx := tx.Message.WithContext(ctx)
//	base := dto.MessageBaseDto{
//		Url:  url.String(),
//		Size: -1,
//		Name: fileName,
//	}
//	extra := dto.ImgMessageDto{
//		MessageBaseDto: base,
//		// TODO: 宽高需要前端传
//		Width:  -1,
//		Height: -1,
//	}
//	jsonStr, err := json.Marshal(extra)
//	if err != nil {
//		if err := tx.Rollback(); err != nil {
//			global.Logger.Errorf("事务回滚失败 %s", err)
//		}
//		global.Logger.Errorf("json序列化失败 %s", err)
//		resp.ErrorResponse(c, "获取签名失败，请稍后再试")
//		c.Abort()
//		return
//	}
//	// TODO:抽象为常量
//	newMsg := model.Message{
//		FromUID:      uid,
//		RoomID:       roomId,
//		Content:      "[图片]",
//		DeleteStatus: 0,
//		Type:         3,
//		Extra:        string(jsonStr),
//	}
//	if err := messageTx.Create(&newMsg); err != nil {
//		if err := tx.Rollback(); err != nil {
//			global.Logger.Errorf("事务回滚失败 %s", err)
//		}
//		global.Logger.Errorf("数据库插入失败 %s", err)
//		resp.ErrorResponse(c, "获取签名失败，请稍后再试")
//		c.Abort()
//		return
//	}
//
//	global.Bus.Publish(enum.NewMessageEvent, newMsg)
//
//	resp.SuccessResponse(c, preSignedResp)
//	return
//}
