package middlewear

import (
	"emotionalBeach/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

var (
	jwtSecret = []byte("emotionBeach")
)

// Claims 是一些实体（通常指的用户）的状态和额外的元数据
type Claims struct {
	UserID uint `json:"userId"`
	jwt.RegisteredClaims
}

// GenerateToken 根据用户的用户名和密码产生token
func GenerateToken(userId uint, iss string) (string, error) {
	// 设置token有效时间

	claims := Claims{
		UserID: userId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			Issuer:    iss,
		},
	}

	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// 该方法内部生成签名字符串，再用于获取完整、已签名的token
	token, err := tokenClaims.SignedString(jwtSecret)
	return token, err
}

func AuthJwt() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		user := c.GetHeader("Uid")
		if token == "" || user == "" {
			models.Error(c, http.StatusUnauthorized, "授权信息和用户ID不能为空!")
			c.Abort()
			return
		}
		userId, err := strconv.Atoi(user)
		if err != nil {
			models.Error(c, http.StatusUnauthorized, "用户ID不合法")
			c.Abort()
			return
		}
		claims, errs := ParseToken(token)
		if errs != nil || claims.UserID != uint(userId) {
			models.Error(c, http.StatusUnauthorized, "token无效或用户身份不合法")
			c.Abort()
			return
		}

		zap.S().Info("token认证成功")
		c.Next()
	}
}

// ParseToken 根据传入的token值获取到Claims对象信息（进而获取其中的用户id）
func ParseToken(token string) (*Claims, error) {

	//用于解析鉴权的声明，方法内部主要是具体的解码和校验的过程，最终返回*Token
	tokenClaims, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if tokenClaims != nil {
		// 从tokenClaims中获取到Claims对象，并使用断言，将该对象转换为我们自己定义的Claims
		if claims, ok := tokenClaims.Claims.(*Claims); ok && tokenClaims.Valid {
			return claims, nil
		}
	}
	return nil, err
}
