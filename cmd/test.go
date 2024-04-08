package main

import (
	"DiTing-Go/event/listener"
	"DiTing-Go/global"
)

func main() {
	listener.Init()
	global.Bus.Publish("calculator", 20, 40)
}
