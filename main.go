package main

import (
	"emotionalBeach/initialize"
	"emotionalBeach/router"

	"github.com/gin-gonic/gin"
)

func main() {
	//初始化日志
	gin.SetMode(gin.ReleaseMode)
	initialize.InitLogger()
	//初始化数据库
	initialize.InitDB()
	routers := router.Router()
	err := routers.Run(":8080")
	if err != nil {
		return
	}
}
