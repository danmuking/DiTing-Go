package routes

import (
	_ "DiTing-Go/docs"
	"DiTing-Go/models"
	"DiTing-Go/pkg/middleware/jwt"
	"DiTing-Go/service"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"log"
	"net/http"
)

// 定义一个升级器，将普通的http连接升级为websocket连接
var upgrader = &websocket.Upgrader{
	//定义读写缓冲区大小
	WriteBufferSize: 1024,
	ReadBufferSize:  1024,
	//校验请求
	CheckOrigin: func(r *http.Request) bool {
		//如果不是get请求，返回错误
		if r.Method != "GET" {
			fmt.Println("请求方式错误")
			return false
		}
		//还可以根据其他需求定制校验规则
		return true
	},
}

// 处理websocket请求
func socketHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrade our raw HTTP connection to a websocket based one
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("Error during connection upgradation:", err)
		return
	}
	defer conn.Close()

	//连接成功后注册用户
	user := &models.User{
		Conn: conn,
		Msg:  make(chan []byte),
	}
	models.Users.Register <- user
	//得到连接后，就可以开始读写数据了
	go read(user)
	write(user)
}

func read(user *models.User) {

	//从连接中循环读取信息
	for {
		_, msg, err := user.Conn.ReadMessage()
		if err != nil {
			fmt.Println("用户退出:", user.Conn.RemoteAddr().String())
			models.Users.Unregister <- user
			break
		}
		//将读取到的信息传入websocket处理器中的broadcast中，
		models.Users.Broadcast <- msg
	}
}
func write(user *models.User) {
	for data := range user.Msg {
		err := user.Conn.WriteMessage(1, data)
		if err != nil {
			fmt.Println("写入错误")
			break
		}
	}
}

// InitRouter 初始化路由
func InitRouter() {
	go initWebSocket()
	initGin()
}

// 初始化websocket
func initWebSocket() {
	go models.Users.Run()
	http.HandleFunc("/socket", socketHandler)
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
		apiContact.GET("/getContactList", service.Login)
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
