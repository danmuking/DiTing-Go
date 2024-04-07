package domain

import "github.com/gorilla/websocket"

// User 定义一个简单的用户结构体
type User struct {
	Conn *websocket.Conn
	Msg  chan []byte
}
