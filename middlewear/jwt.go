package middlewear

import (
	"emotionalBeach/models"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var (
	TokenExpired = errors.New("token is expired")
	jwtSecret    = []byte("emotionBeach")
)

// Claims 是一些实体（通常指的用户）的状态和额外的元数据
type Claims struct {
	UserID uint `json:"userId"`
	jwt.StandardClaims
}

// GenerateToken 根据用户的用户名和密码产生token
func GenerateToken(userId uint, iss string) (string, error) {
	// 设置token有效时间
	nowTime := time.Now()
	expireTime := nowTime.Add(14 * 24 * time.Hour) // 过期时间为14天

	claims := Claims{
		UserID: userId,
		StandardClaims: jwt.StandardClaims{
			// 过期时间
			ExpiresAt: expireTime.Unix(),
			// 指定token发行人
			Issuer: iss,
		},
	}

	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// 该方法内部生成签名字符串，再用于获取完整、已签名的token
	token, err := tokenClaims.SignedString(jwtSecret)
	return token, err
}

func JWY() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		user := c.Query("uid")
		userId, err := strconv.Atoi(user)
		if token == "" || user == "" {
			models.Error(c, http.StatusUnauthorized, "授权信息和用户ID不能为空!")
			c.Abort()
			return
		}
		if err != nil {
			models.Error(c, http.StatusUnauthorized, "您uid不合法")
			c.Abort()
			return
		}
		if token == "" {
			models.Error(c, http.StatusUnauthorized, "请登录")
			c.Abort()
			return
		} else {
			claims, err := ParseToken(token)
			if err != nil {
				models.Error(c, http.StatusUnauthorized, "token失效")
				c.Abort()
				return
			} else if time.Now().Unix() > claims.ExpiresAt {
				err = TokenExpired
				models.Error(c, http.StatusUnauthorized, "授权已过期")
				c.Abort()
				return
			}

			if claims.UserID != uint(userId) {
				models.Error(c, http.StatusUnauthorized, "您的登录身份不合法")
				c.Abort()
				return
			}

			zap.S().Info("token认证成功")
			c.Next()
		}
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
		// 要传入指针，项目中结构体都是用指针传递，节省空间。
		if claims, ok := tokenClaims.Claims.(*Claims); ok && tokenClaims.Valid {
			return claims, nil
		}
	}
	return nil, err
}
