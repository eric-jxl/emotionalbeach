package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func AssetsCacheMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/assets/") {
			// 设置缓存时间为 1 小时 (你可以根据需要调整)
			// Cache-Control 是现代浏览器推荐的主要缓存控制头
			c.Header("Cache-Control", "public, max-age=3600") // 3600 秒 = 1 小时
			// Expires 是较老的头，但为了兼容性也可以加上
			expiresTime := time.Now().Add(1 * time.Hour)
			c.Header("Expires", expiresTime.UTC().Format(http.TimeFormat))
		}
	}
}
