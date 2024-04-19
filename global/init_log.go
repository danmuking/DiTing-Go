package global

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"io"
	"os"
	"path"
)

var Logger *logrus.Logger

func init() {
	logFilePath := viper.GetString("log.log_file_path")
	logFileName := viper.GetString("log.log_file_name")
	//日志文件
	fileName := path.Join(logFilePath, logFileName)
	//写入文件
	src, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("err", err)
	}

	//实例化
	Logger = logrus.New()
	writers := []io.Writer{
		src,
		os.Stdout}
	fileAndStdoutWriter := io.MultiWriter(writers...)
	if err == nil {
		Logger.SetOutput(fileAndStdoutWriter)
	} else {
		Logger.Info("failed to log to file.")
	}

	//设置日志级别
	Logger.SetLevel(logrus.DebugLevel)
	Logger.SetReportCaller(true)
	//设置日志格式
	Logger.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})
}
