package main

import (
	"emotionalBeach/initialize"
	"emotionalBeach/router"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	//初始化日志
	initialize.InitLogger()
	//初始化数据库
	initialize.InitDB()
	routers := router.Router()
	zap.L().Info("程序加载中...")
	err := routers.Run(":8080")
	if err != nil {
		return
	}
}
