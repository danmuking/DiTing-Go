package main

import (
	"DiTing-Go/routes"
)

// swagger 中添加header.Authorization:token 校验 token
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	//global.InitDB()
	routes.InitRouter()
}
