package global

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func RespJson(c *gin.Context, code int, message string, data interface{}) {
	c.JSON(http.StatusOK, response{Code: code, Message: message, Data: data})
}

func Success(c *gin.Context, data interface{}) {
	RespJson(c, http.StatusOK, "Success", data)
}

func Error(c *gin.Context, code int, message string) {
	RespJson(c, code, message, nil)
}

func Fatal(c *gin.Context, message string) {
	RespJson(c, http.StatusInternalServerError, message, nil)
}
