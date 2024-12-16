package middlewear

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func CorsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置允许的源
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		// 设置允许的方法
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		// 设置允许的请求头
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Length, Content-Type, Authorization")
		// 设置是否允许携带凭证
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		// 设置预检请求的有效期
		c.Writer.Header().Set("Access-Control-Max-Age", "3600")
		// 处理预检请求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
