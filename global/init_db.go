package global

import (
	"DiTing-Go/dal"
	"DiTing-Go/dal/query"
	"fmt"
	"github.com/spf13/viper"
)

var MySQLDSN string
var Query *query.Query

func DBInit() {
	MySQLDSN = fmt.Sprintf("%s:%s@tcp(%s:%s)/DiTing?charset=utf8mb4&parseTime=True&loc=Local", viper.GetString("mysql.username"), viper.GetString("mysql.password"), viper.GetString("mysql.host"), viper.GetString("mysql.port"))
	dal.DB = dal.ConnectDB(MySQLDSN).Debug()
	//dal.DB = dal.ConnectDB(MySQLDSN)
	// 设置默认DB对象
	query.SetDefault(dal.DB)
	Query = query.Use(dal.DB)
}
