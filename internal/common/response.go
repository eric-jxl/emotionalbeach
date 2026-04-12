package common

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response is the unified API envelope returned by every endpoint.
type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// Success writes a 200 OK response with code=0.
func Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{Code: http.StatusOK, Message: "success", Data: data})
}

// Fail writes a business-failure (4xx) response.
func Fail(c *gin.Context, httpStatus int, message string) {
	c.JSON(httpStatus, Response{Code: httpStatus, Message: message})
}

// ServerError writes a 500 Internal Server Error response.
func ServerError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError,
		Response{Code: http.StatusInternalServerError, Message: message})
}
