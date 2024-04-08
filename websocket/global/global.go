package global

import (
	"github.com/gorilla/websocket"
	cmap "github.com/orcaman/concurrent-map/v2"
	"sync"
)

type Channels struct {
	Uid         int64
	ChannelList []*websocket.Conn
	Mu          *sync.RWMutex
}
type User struct {
	Uid     int64
	Channel *websocket.Conn
}
type Msg struct {
	Uid int64
}

// UserChannelMap 用户和channel的映射
var UserChannelMap = cmap.New[*Channels]()
