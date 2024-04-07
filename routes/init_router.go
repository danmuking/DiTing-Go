package routes

import (
	_ "DiTing-Go/docs"
	"DiTing-Go/pkg/middleware/jwt"
	"DiTing-Go/service"
	websocketService "DiTing-Go/websocket/service"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"log"
	"net/http"
)

// InitRouter 初始化路由
func InitRouter() {
	go initWebSocket()
	initGin()
}

// 初始化websocket
func initWebSocket() {
	http.HandleFunc("/socket", websocketService.Connect)
	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}

// 初始化gin
func initGin() {
	router := gin.Default()

	//添加swagger访问路由
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	// 不需要身份验证的路由
	apiPublic := router.Group("/api/public")
	{
		//获取标签列表
		apiPublic.POST("/register", service.Register)
		//新建标签
		apiPublic.POST("/login", service.Login)
	}

	apiContact := router.Group("/api/contact")
	apiContact.Use(jwt.JWT())
	{
		//添加好友
		apiContact.POST("/add", service.ApplyFriend)
		//删除好友
		apiContact.DELETE("/delete", service.DeleteFriend)
		//获取好友申请列表
		apiContact.GET("/getApplyList", service.GetApplyList)
		//同意好友申请
		apiContact.PUT("/agree", service.Agree)
		//获取好友列表
		apiContact.GET("/getContactList", service.GetContactList)
		//判断是否是好友
		apiContact.GET("/isFriend/:friendUid", service.IsFriend)
		//好友申请未读数量
		apiContact.GET("/unreadApplyNum", service.UnreadApplyNum)
	}

	err := router.Run(":5000")
	if err != nil {
		return
	}
}
