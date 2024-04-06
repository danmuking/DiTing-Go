package service

import (
	"DiTing-Go/dal"
	"DiTing-Go/dal/model"
	"DiTing-Go/dal/query"
	"DiTing-Go/pkg/resp"
	_ "DiTing-Go/pkg/setting"
	"DiTing-Go/pkg/utils"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

var MySQLDSN = fmt.Sprintf("%s:%s@tcp(%s:%s)/diting?charset=utf8mb4&parseTime=True", viper.GetString("mysql.username"), viper.GetString("mysql.password"), viper.GetString("mysql.host"), viper.GetString("mysql.port"))

func init() {
	println("MySQLDSN: ", MySQLDSN)
	dal.DB = dal.ConnectDB(MySQLDSN).Debug()
	// 设置默认DB对象
	query.SetDefault(dal.DB)
}

// Register 用户注册
//
//	@Summary	用户注册
//	@Produce	json
//	@Param		name		body		string				true	"用户名"
//	@Param		password	body		string				true	"密码"
//	@Success	200			{object}	resp.ResponseData	"成功"
//	@Failure	500			{object}	resp.ResponseData	"内部错误"
//	@Router		/api/public/register [post]
func Register(c *gin.Context) {
	user := model.User{}
	if err := c.ShouldBind(&user); err != nil { //ShouldBind()会自动推导
		resp.ErrorResponse(c, "参数错误")
		return
	}

	u := query.User
	// 用户名是否已存在
	exist, _ := u.WithContext(context.Background()).Where(u.Name.Eq(user.Name)).First()
	if exist != nil {
		resp.ErrorResponse(c, "用户名已存在")
		return
	}

	// 创建对象
	err := u.WithContext(context.Background()).Omit(u.OpenID).Create(&user)
	if err != nil {
		resp.SuccessResponseWithMsg(c, "注册成功")
		return
	}
}

// Login 用户登录
//
//	@Summary	用户登录
//	@Produce	json
//	@Param		name		body		string				true	"用户名"
//	@Param		password	body		string				true	"密码"
//	@Success	200			{object}	resp.ResponseData	"成功"
//	@Failure	500			{object}	resp.ResponseData	"内部错误"
//	@Router		/api/public/login [post]
func Login(c *gin.Context) {
	login := model.User{}
	if err := c.ShouldBind(&login); err != nil { //ShouldBind()会自动推导
		resp.ErrorResponse(c, "参数错误")
		return
	}

	u := query.User
	// 检查密码是否正确
	user, _ := u.WithContext(context.Background()).Where(u.Name.Eq(login.Name), u.Password.Eq(login.Password)).First()
	if user == nil {
		resp.ErrorResponse(c, "用户名或密码错误")
		return
	}
	//生成jwt
	token, _ := utils.GenerateToken(user.ID)
	resp.SuccessResponse(c, token)
}
