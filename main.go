package main

import (
	"emotionalBeach/initialize"
	"emotionalBeach/router"
	"flag"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var filepath *string

func init() {
	filepath = flag.String("e", "env", "数据库配置文件(.env)路径")
	flag.Parse()
}
func main() {
	gin.SetMode(gin.ReleaseMode)
	//初始化日志
	initialize.InitLogger()
	//初始化数据库
	initialize.InitDB(*filepath)
	routers := router.Router()
	zap.L().Info("程序加载中...")
	err := routers.Run(":8080")
	if err != nil {
		return
	}
}
