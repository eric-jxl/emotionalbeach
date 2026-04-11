package webhookhandler

import "github.com/gin-gonic/gin"

// RegisterRoutes mounts the webhook endpoint onto the given router group.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/webhook", h.WebhookEmail)
}

