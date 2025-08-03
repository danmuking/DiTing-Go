package tests

import (
	"DiTing-Go/global"
	"log"
	"os"
	"testing"
)

// TestMain 在所有测试运行前执行初始化
func TestMain(m *testing.M) {
	// 设置测试环境变量
	os.Setenv("GIN_MODE", "test")
	os.Setenv("TEST_ENV", "test")

	// 初始化日志
	global.LogInit()

	// 初始化Redis
	global.RedisInit()

	// 初始化数据库
	global.DBInit()

	// 运行测试
	code := m.Run()

	// 清理资源
	cleanup()

	os.Exit(code)
}

// cleanup 清理测试资源
func cleanup() {
	// 关闭Redis连接
	if global.Rdb != nil {
		global.Rdb.Close()
	}

	// 关闭数据库连接
	if global.Query != nil {
		// 这里可以添加数据库连接关闭逻辑
	}

	log.Println("测试资源清理完成")
}
