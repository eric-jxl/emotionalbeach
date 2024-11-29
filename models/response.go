package models

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Resp struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func RespJson(c *gin.Context, code int, message string, data interface{}) {
	resp := Resp{Code: code, Message: message, Data: data}
	c.JSON(http.StatusOK, resp)
}

func Success(c *gin.Context, data interface{}) {
	RespJson(c, http.StatusOK, "操作成功", data)
}

func Error(c *gin.Context, code int, message string) {
	RespJson(c, code, message, nil)
}

func Fatal(c *gin.Context, message string) {
	RespJson(c, http.StatusInternalServerError, message, nil)
}

func Exception() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				Error(c, http.StatusInternalServerError, "服务器错误")
			}
		}()
		c.Next()
	}
}
