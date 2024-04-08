package main

import (
	"DiTing-Go/global"
)

func main() {
	global.Bus.Publish("calculator", 20, 40)
}
