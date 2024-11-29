package main

import (
	"emotionalBeach/global"
	"emotionalBeach/models"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// ProtectedEndpoint
// @Summary 受保护的接口
// @Description 需要API Key认证的接口
// @Tags 受保护的接口
// @Accept json
// @Produce json
// @Success 200 {object} models.SuccessResponse
// @Failure 401 {object} models.ErrorResponse
// @Security ApiKeyAuth
// @Router /protected [get]
func ProtectedEndpoint(c *gin.Context) {
	// 受保护的接口实现
	fmt.Println("ProtectedEndpoint")
}

// @title 这里写标题
// @version 1.0
// @description 这里写描述信息
// @termsOfService http://swagger.io/terms/

// @contact.name 这里写联系人信息
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host 这里写接口服务的host
// @BasePath 这里写base path
func main() {
	err := godotenv.Load("../config/.env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	url := ginSwagger.URL("http://localhost:8080/swagger/doc.json") // The url pointing to API definition
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))
	r.GET("/protect", ProtectedEndpoint)

	// 定义一个既可以处理GET也可以处理POST请求的路由
	r.Any("/example", func(c *gin.Context) {
		switch c.Request.Method {
		case http.MethodGet:
			c.JSON(http.StatusOK, gin.H{
				"method":  "GET",
				"message": "This is a GET request",
			})
		case http.MethodPost:
			c.JSON(http.StatusOK, gin.H{
				"method":  "POST",
				"message": "This is a POST request",
			})
		default:
			c.JSON(http.StatusMethodNotAllowed, gin.H{
				"error": "Method not allowed",
			})
		}
	})

	// 创建一个新的 UserBasic 记录
	newUser := &models.UserBasic{
		Name:       "John Doe",
		PassWord:   "password123",
		Avatar:     "avatar.jpg",
		Gender:     "male",
		Phone:      "13800000000",
		Email:      "john.doe@example.com",
		Identity:   "1234567890",
		ClientIp:   "192.168.1.1",
		ClientPort: "8080",
		Salt:       "salt",
		IsLoginOut: false,
		DeviceInfo: "some device info",
	}

	// 创建记录
	newObj := global.DB.Create(newUser)
	println(newObj.RowsAffected)

	r.Run(":8080")

}
