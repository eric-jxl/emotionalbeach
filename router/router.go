package router

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Router
// @title Gin Swagger Example API
// @version 1.0
// @description This is a sample server.
// @host localhost:8080
// @BasePath /api/v1
func Router() *gin.Engine {
	//初始化路由
	router := gin.Default()
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	////v1版本
	//v1 := router.Group("v1")
	//
	////用户模块，后续有个用户的api就放置其中
	//user := v1.Group("user")
	//{
	//	user.GET("/list")
	//}

	return router
}
