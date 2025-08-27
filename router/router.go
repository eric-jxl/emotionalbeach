package router

import (
	"emotionalBeach/controller"
	_ "emotionalBeach/docs"
	"emotionalBeach/middleware"
	"emotionalBeach/templates"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"net/http"
)

func Router() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	//router.GET("/", func(c *gin.Context) {
	//	c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	//})

	router.GET("/", func(c *gin.Context) {
		data, err := templates.IndexHTML.ReadFile("index.html")
		if err != nil {
			c.String(http.StatusInternalServerError, "Error loading index.html")
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	})
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	router.Any("/login", func(c *gin.Context) {
		switch c.Request.Method {
		case http.MethodPost:
			controller.LoginByNameAndPassWord(c)
		default:
			c.JSON(http.StatusMethodNotAllowed, gin.H{"code": http.StatusMethodNotAllowed, "message": "Only  POST methods are allowed"})
		}

	})
	router.POST("/register", controller.NewUser)

	v1 := router.Group("/v1")
	//用户接口
	user := v1.Group("user").Use(middleware.AuthJwt())

	{
		user.GET("/list", controller.GetUsers)
		user.GET("/condition", controller.GetAppointUser)
		user.DELETE("/delete", controller.DeleteUser)
		user.POST("/update", controller.UpdateUser)
	}

	//好友关系
	relation := v1.Group("relation").Use(middleware.AuthJwt())
	{
		relation.POST("/list", controller.FriendList)
		relation.POST("/add", controller.AddFriendByName)
	}

	return router
}
