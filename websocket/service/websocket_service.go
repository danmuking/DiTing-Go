package service

import (
	"DiTing-Go/pkg/utils"
	"DiTing-Go/websocket/global"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
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

// Connect 建立WebSocket连接
func Connect(w http.ResponseWriter, r *http.Request) {
	//先获得Http的token中的uid
	uid, err := parseJwt(r)
	if err != nil {
		log.Print("Error during connection upgradation:", err)
		return
	}
	// Upgrade our raw HTTP connection to a websocket based one
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("Error during connection upgradation:", err)
		return
	}

	//连接成功后注册用户
	// 将uid转换为string
	stringUid := strconv.FormatInt(*uid, 10)
	userChannel := global.Channels{
		Uid:         *uid,
		ChannelList: make([]*websocket.Conn, 0),
		Mu:          new(sync.RWMutex),
	}
	user := global.User{
		Uid:     *uid,
		Channel: conn,
	}
	global.UserChannelMap.SetIfAbsent(stringUid, userChannel)
	userChannel, _ = global.UserChannelMap.Get(stringUid)
	// TODO:加锁方式是否正确
	// 将连接加入到用户的channel中
	userChannel.Mu.Lock()
	userChannel.ChannelList = append(userChannel.ChannelList, conn)
	userChannel.Mu.Unlock()

	// 定时发送心跳消息
	go heatBeat(&user)
}

// Send 发送空消息代表有新消息，前端收到消息后再去后端拉取消息
func Send(msg *global.Msg) {
	uid := msg.Uid
	stringUid := strconv.FormatInt(uid, 10)
	channels, _ := global.UserChannelMap.Get(stringUid)
	for _, conn := range channels.ChannelList {
		// TODO: 消息类型抽象成枚举
		err := conn.WriteMessage(2, []byte(""))
		if err != nil {
			fmt.Println("写入错误")
			break
		}
	}
}

// 移除连接
func disConnect(user *global.User) {
	stringUid := strconv.FormatInt(user.Uid, 10)
	conn := user.Channel
	userChannel, _ := global.UserChannelMap.Get(stringUid)
	userChannel.Mu.Lock()
	for i, item := range userChannel.ChannelList {
		if item == conn {
			userChannel.ChannelList = append(userChannel.ChannelList[:i], userChannel.ChannelList[i+1:]...)
		}

	}
	err := conn.Close()
	if err != nil {
		return
	}
	userChannel.Mu.Unlock()
}

// 解析jwt
func parseJwt(r *http.Request) (*int64, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, errors.New("无权限访问")
	}
	// 按空格分割
	parts := strings.SplitN(authHeader, " ", 2)
	if !(len(parts) == 2 && parts[0] == "Bearer") {
		return nil, errors.New("无权限访问")
	}
	// parts[1]是获取到的tokenString，我们使用之前定义好的解析JWT的函数来解析它
	token, err := utils.ParseToken(parts[1])
	if err != nil {
		return nil, errors.New("无权限访问")
	}
	return &token.Uid, nil
}

// 心跳检测
func heatBeat(user *global.User) {
	conn := user.Channel
	// TODO:心跳时间从配置文件中读取
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := conn.WriteMessage(websocket.PingMessage, []byte("heartbeat"))
			if err != nil {
				log.Println(err)
				return
			}
			conn.SetReadDeadline(time.Now().Add(10 * time.Second))
			_, _, err = conn.ReadMessage()
			if err != nil {
				// TODO:下线操作
				disConnect(user)
				log.Println("heartbeat response error:", err)
				return
			}
		}
	}
}
