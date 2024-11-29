package main

import (
	"emotionalBeach/controller"
	"emotionalBeach/initialize"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title 情感沙滩
// @version 3.0
// @description 情感沙滩
// @termsOfService http://swagger.io/terms/

// @contact.name 这里写联系人信息
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath
func main() {
	gin.SetMode(gin.TestMode)
	initialize.InitDB()
	r := gin.Default()
	url := ginSwagger.URL("http://localhost:8080/test/docs/swagger.json") // The url pointing to API definition
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))
	r.GET("/test/user", controller.GetUsers)
	// 创建一个新的 UserBasic 记录
	//newUser := &models.UserBasic{
	//	Name:       "John Doe",
	//	PassWord:   "password123",
	//	Avatar:     "avatar.jpg",
	//	Gender:     "male",
	//	Phone:      "13800000000",
	//	Email:      "john.doe@example.com",
	//	Identity:   "1234567890",
	//	ClientIp:   "192.168.1.1",
	//	ClientPort: "8080",
	//	Salt:       "salt",
	//	IsLoginOut: false,
	//	DeviceInfo: "some device info",
	//}
	//
	//// 创建记录
	//newObj := global.DB.Create(newUser)

	r.Run(":8080")

}
