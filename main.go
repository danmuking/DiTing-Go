package main

import (
	//_ "DiTing-Go/event/listener"
	"DiTing-Go/global"
	"DiTing-Go/routes"
)

// swagger 中添加header.Authorization:token 校验 token
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	// 初始化数据库连接
	global.DBInit()
	//初始化redis连接
	global.RedisInit()
	global.LogInit()

	routes.InitRouter()
}
