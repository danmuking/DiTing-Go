package service

import (
	"DiTing-Go/dal/model"
	"DiTing-Go/dal/query"
	_ "DiTing-Go/pkg/setting"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"gorm.io/gen/examples/dal"
	"net/http"
)

var MySQLDSN = fmt.Sprintf("%s:%s@tcp(%s:%s)/diting?charset=utf8mb4&parseTime=True", viper.GetString("mysql.username"), viper.GetString("mysql.password"), viper.GetString("mysql.host"), viper.GetString("mysql.port"))

func init() {
	dal.DB = dal.ConnectDB(MySQLDSN).Debug()
}

// Register 用户注册
func Register(c *gin.Context) {
	user := model.User{}
	if err := c.ShouldBind(&user); err != nil { //ShouldBind()会自动推导
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 设置默认DB对象
	query.SetDefault(dal.DB)
	err := query.User.WithContext(context.Background()).Create(&user)
	if err != nil {
		fmt.Printf("create book fail, err:%v\n", err)
		return
	}
}

// Login 用户登录
func Login(c *gin.Context) {

}
