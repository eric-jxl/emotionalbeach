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
	router.POST("/login", controller.LoginByNameAndPassWord)
	router.POST("/register", controller.NewUser)

	v1 := router.Group("v1")
	//用户接口
	user := v1.Group("user").Use(middlewear.JWY())

	{
		user.GET("/list", controller.GetUsers)
		user.Any("/delete", controller.DeleteUser)
		user.POST("/update", controller.UpdateUser)
	}

	//好友关系
	relation := v1.Group("relation").Use(middlewear.JWY())
	{
		relation.POST("/list", controller.FriendList)
		relation.POST("/add", controller.AddFriendByName)
	}

	return router
}
