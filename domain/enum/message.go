package enum

import "github.com/gorilla/websocket"

const (
	NewMessage  = websocket.TextMessage
	TextMessage = 1
)

const (
	TextMessageType = 1
	ImgMessageType  = 3
)
