package cursor

import (
	"DiTing-Go/dal/model"
	"fmt"
	"github.com/goccy/go-json"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"testing"
)

func TestPaginate(t *testing.T) {

	// 设置配置文件的名字
	viper.SetConfigName("config")
	// 设置配置文件的类型
	viper.SetConfigType("yaml")
	// 添加配置文件的路径，指定 config 目录下寻找
	viper.AddConfigPath("../../conf")
	_ = viper.ReadInConfig()
	// 连接到 MySQL 数据库
	var MySQLDSN = fmt.Sprintf("%s:%s@tcp(%s:%s)/diting?charset=utf8mb4&parseTime=True", viper.GetString("mysql.username"), viper.GetString("mysql.password"), viper.GetString("mysql.host"), viper.GetString("mysql.port"))
	db, err := gorm.Open(mysql.Open(MySQLDSN),
		&gorm.Config{Logger: logger.Default.LogMode(logger.Info)})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}

	cursor := "20017"
	// 分页参数
	params := PageReq{
		PageSize: 2,
		Cursor:   &cursor, // 可以从某个特定的游标开始分页
	}
	db = db.Table("user")

	var result []model.User = make([]model.User, 0)
	db.Raw("select * from user").Scan(&result)
	println(result)
	// 调用Paginate函数进行分页
	resp, err := Paginate(db, params, &result, "id", true)
	if err != nil {
		t.Fatalf("Paginate error: %v", err)
	}

	// 检查结果
	marshal, err := json.Marshal(resp)
	if err != nil {
		return
	}
	fmt.Println(string(marshal))

	// 分页参数
	params = PageReq{
		PageSize: 2,
		Cursor:   nil, // 可以从某个特定的游标开始分页
	}

	db = db.Session(&gorm.Session{NewDB: true})
	condition := []string{"name = ?", "test1"}
	result = make([]model.User, 0)
	// 调用Paginate函数进行分页
	resp, err = Paginate(db, params, &result, "id", true, condition...)
	if err != nil {
		t.Fatalf("Paginate error: %v", err)
	}

	// 检查结果
	marshal, err = json.Marshal(resp)
	if err != nil {
		return
	}
	fmt.Println(string(marshal))
}
