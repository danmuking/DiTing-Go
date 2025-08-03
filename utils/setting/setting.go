package setting

import (
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

func ConfigInit() {
	// 设置配置文件的名字
	viper.SetConfigName("config")
	// 设置配置文件的类型
	viper.SetConfigType("yaml")

	// 获取当前工作目录
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current directory: %v", err)
	}

	// 尝试多个可能的配置文件路径
	configPaths := []string{
		"./conf",           // 当前目录下的conf
		"../conf",          // 上级目录下的conf
		"../../conf",       // 上上级目录下的conf
		"../../../conf",    // 上上上级目录下的conf
		"../../../../conf", // 上上上上级目录下的conf
		"conf",             // 直接conf目录
	}

	// 添加配置文件的路径
	for _, path := range configPaths {
		viper.AddConfigPath(path)
	}

	// 如果设置了PROJECT_ROOT环境变量，也添加该路径
	if projectRoot := os.Getenv("PROJECT_ROOT"); projectRoot != "" {
		viper.AddConfigPath(filepath.Join(projectRoot, "conf"))
	}

	// 尝试读取配置文件
	err = viper.ReadInConfig()
	if err != nil {
		// 如果还是找不到配置文件，尝试从项目根目录查找
		// 通过查找go.mod文件来确定项目根目录
		projectRoot := findProjectRoot(currentDir)
		if projectRoot != "" {
			viper.Reset()
			viper.SetConfigName("config")
			viper.SetConfigType("yaml")
			viper.AddConfigPath(filepath.Join(projectRoot, "conf"))
			err = viper.ReadInConfig()
		}

		if err != nil {
			log.Fatalf("Fail to parse 'conf/config.yml': %v", err)
		}
	}
}

// findProjectRoot 通过查找go.mod文件来确定项目根目录
func findProjectRoot(currentDir string) string {
	dir := currentDir
	for {
		// 检查当前目录是否有go.mod文件
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		// 向上查找父目录
		parent := filepath.Dir(dir)
		if parent == dir {
			// 已经到达根目录
			break
		}
		dir = parent
	}
	return ""
}
