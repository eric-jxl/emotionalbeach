package main

import (
	"emotionalBeach/initialize"
	"emotionalBeach/router"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	//初始化日志
	initialize.InitLogger()
	//初始化数据库
	initialize.InitDB()
	routers := router.Router()
	log.Println("程序加载中...")
	err := routers.Run(":8080")
	if err != nil {
		return
	}
}
