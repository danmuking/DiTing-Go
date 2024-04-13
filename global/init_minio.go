package global

import (
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/spf13/viper"
)

var MinioClient *minio.Client

func init() {
	accessKey := viper.GetString("minio.accessKey")
	accessSecret := viper.GetString("minio.accessSecret")
	endPoint := viper.GetString("minio.endPoint")
	useSSL := viper.GetBool("minio.useSSL")
	// 初始化minio客户端
	minioClient, err := minio.New(endPoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, accessSecret, ""),
		Secure: useSSL,
	})
	MinioClient = minioClient
	if err != nil {
		Logger.Fatal("minio client create fail, err %+v", err)
	}
}
