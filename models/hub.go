package models

// 初始化处理中心，以便调用
var Users = &Hub{
	userList:   make(map[*User]bool),
	Register:   make(chan *User),
	Unregister: make(chan *User),
	Broadcast:  make(chan []byte),
}

type Hub struct {
	//用户列表，保存所有用户
	userList map[*User]bool
	//注册chan，用户注册时添加到chan中
	Register chan *User
	//注销chan，用户退出时添加到chan中，再从map中删除
	Unregister chan *User
	//广播消息，将消息广播给所有连接
	Broadcast chan []byte
}

// 处理中心处理获取到的信息
func (h *Hub) Run() {
	for {
		select {
		//从注册chan中取数据
		case user := <-h.Register:
			//取到数据后将数据添加到用户列表中
			h.userList[user] = true
		case user := <-h.Unregister:
			//从注销列表中取数据，判断用户列表中是否存在这个用户，存在就删掉
			if _, ok := h.userList[user]; ok {
				delete(h.userList, user)
			}
		case data := <-h.Broadcast:
			//从广播chan中取消息，然后遍历给每个用户，发送到用户的msg中
			for u := range h.userList {
				select {
				case u.Msg <- data:
				default:
					delete(h.userList, u)
					close(u.Msg)
				}
			}
		}
	}
}
