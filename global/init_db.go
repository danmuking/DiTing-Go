package global

import (
	"DiTing-Go/dal"
	"fmt"
	"github.com/spf13/viper"
	"gorm.io/gen/examples/dal/query"
)

var MySQLDSN string
var Query *query.Query

func init() {
	MySQLDSN = fmt.Sprintf("%s:%s@tcp(%s:%s)/diting?charset=utf8mb4&parseTime=True", viper.GetString("mysql.username"), viper.GetString("mysql.password"), viper.GetString("mysql.host"), viper.GetString("mysql.port"))
	//println("MySQLDSN: ", MySQLDSN)
	dal.DB = dal.ConnectDB(MySQLDSN).Debug()
	// 设置默认DB对象
	query.SetDefault(dal.DB)
	Query = query.Use(dal.DB)
}
