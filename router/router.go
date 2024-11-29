package router

import (
	"emotionalBeach/controller"
	"emotionalBeach/middlewear"

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
	user := router.Group("user")
	{
		user.GET("/list", middlewear.JWY(), controller.GetUsers)
		user.POST("/login_pw", controller.LoginByNameAndPassWord)
		user.POST("/new", controller.NewUser)
		user.Any("/delete", middlewear.JWY(), controller.DeleteUser)
		user.POST("/update", middlewear.JWY(), controller.UpdateUser)
	}

	return router
}
