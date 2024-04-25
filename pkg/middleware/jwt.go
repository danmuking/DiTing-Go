package middleware

import (
	"DiTing-Go/pkg/domain/vo/resp"
	"DiTing-Go/pkg/utils"
	"github.com/gin-gonic/gin"
	"strings"
)

// JWT jwt中间件
func JWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 客户端携带Token有三种方式 1.放在请求头 2.放在请求体 3.放在URI
		// 这里假设Token放在Header的Authorization中，并使用Bearer开头
		// 这里的具体实现方式要依据你的实际业务情况决定
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			resp.ErrorResponse(c, "无权限访问")
			c.Abort()
			return
		}
		// 按空格分割
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			resp.ErrorResponse(c, "无权限访问")
			c.Abort()
			return
		}
		// parts[1]是获取到的tokenString，我们使用之前定义好的解析JWT的函数来解析它
		token, err := utils.ParseToken(parts[1])
		if err != nil {
			resp.ErrorResponse(c, "无权限访问")
			c.Abort()
			return
		}
		//把解析出来的token存储到请求的上下文c上,方便后续的处理函数获取
		c.Set("uid", token.Uid)
		c.Next()
	}
}
