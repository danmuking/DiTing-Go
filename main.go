package main

import (
	"DiTing-Go/pkg/setting"
	"DiTing-Go/routes"
)

func main() {
	setting.InitConfig()
	routes.InitRouter()
}
