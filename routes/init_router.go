package routes

import (
	"DiTing-Go/controller"
	_ "DiTing-Go/docs"
	"DiTing-Go/pkg/domain/vo/resp"
	"DiTing-Go/pkg/middleware"
	"DiTing-Go/service"
	"DiTing-Go/websocket/global"
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
	http.HandleFunc("/websocket", websocketService.Connect)
	log.Fatal(http.ListenAndServe("localhost:5001", nil))
}

// 初始化gin
func initGin() {
	router := gin.Default()
	router.Use(middleware.LoggerToFile())
	router.Use(middleware.Cors())
	//添加swagger访问路由
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	// 不需要身份验证的路由
	apiPublic := router.Group("/api/public")
	{
		//获取标签列表
		apiPublic.POST("/register", controller.RegisterController)
		//新建标签
		apiPublic.POST("/login", controller.LoginController)
	}

	apiUser := router.Group("/api/user")
	apiUser.Use(middleware.JWT())
	{
		//添加好友
		apiUser.POST("/add", controller.ApplyFriendController)
		//删除好友
		apiUser.DELETE("/delete/", controller.DeleteFriendController)
		//同意好友申请
		apiUser.PUT("/agree", controller.AgreeFriendController)
		//获取好友申请列表
		apiUser.GET("/getApplyList", controller.GetUserApplyController)
		//获取好友列表
		apiUser.GET("/getFriendList", controller.GetFriendListController)
		// 判断是否是好友
		apiUser.GET("/isFriend/:friendUid", controller.IsFriendController)
		//好友申请未读数量
		apiUser.GET("/unreadApplyNum", controller.UnreadApplyNumController)
		//根据好友昵称搜索好友
		apiUser.GET("/getUserInfoByName", controller.GetUserInfoByNameController)
		// TODO:测试使用
		apiUser.GET("/test", test)
	}

	apiGroup := router.Group("/api/group")
	apiGroup.Use(middleware.JWT())
	{
		//创建群聊
		apiGroup.POST("/create", controller.CreateGroupController)
		apiGroup.DELETE("/delete/", controller.DeleteGroupController)
		apiGroup.POST("/join", controller.JoinGroupController)
		apiGroup.POST("/quit", service.QuitGroupService)
		apiGroup.GET("/getGroupMemberList", controller.GetGroupMemberListController)
		apiGroup.POST("/grantAdministrator", service.GrantAdministratorService)
		apiGroup.POST("/removeAdministrator", service.RemoveAdministratorService)
	}

	apiContact := router.Group("/api/contact")
	apiContact.Use(middleware.JWT())
	{
		apiContact.GET("getContactList", controller.GetContactListController)
		apiContact.GET("getNewContactList", controller.GetNewContactListController)
		apiContact.GET("getMessageList", service.GetContactDetailService)
		apiContact.GET("getNewMsgList", controller.GetNewMsgListController)
		apiContact.POST("userInfo/batch", controller.GetUserInfoBatchController)
	}

	apiMsg := router.Group("/api/chat")
	apiMsg.Use(middleware.JWT())
	{
		apiMsg.POST("msg", controller.SendMessageController)
	}

	apiFile := router.Group("/api/file")
	apiFile.Use(middleware.JWT())
	{
		apiFile.GET("getPreSigned", service.GetPreSigned)
	}

	err := router.Run(":5000")
	if err != nil {
		return
	}
}

// TODO:测试使用
func test(c *gin.Context) {
	msg := new(global.Msg)
	msg.Uid = 2
	websocketService.Send(msg.Uid, []byte("{\"type\":4}"))
	resp.SuccessResponse(c, nil)
}
