package setting

import (
	"github.com/spf13/viper"
	"log"
)

func InitConfig() {
	// 设置配置文件的名字
	viper.SetConfigName("config")
	// 设置配置文件的类型
	viper.SetConfigType("yaml")
	// 添加配置文件的路径，指定 config 目录下寻找
	viper.AddConfigPath("./conf")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Fail to parse 'conf/config.yml': %v", err)
	}
}
