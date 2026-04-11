package relationhandler

import "github.com/gin-gonic/gin"

// RegisterRoutes mounts all relation endpoints onto the given router group.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/list", h.FriendList)
	rg.POST("/add", h.AddFriendByName)
}

