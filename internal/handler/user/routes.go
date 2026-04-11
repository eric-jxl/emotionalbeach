package userhandler

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes mounts all user endpoints onto the given router group.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/list", h.GetUsers)
	rg.GET("/condition", h.GetAppointUser)
	rg.POST("/update", h.UpdateUser)
	rg.DELETE("/delete", h.DeleteUser)
}

