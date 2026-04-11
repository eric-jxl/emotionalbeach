package healthhandler

import "github.com/gin-gonic/gin"

// RegisterRoutes mounts the health-check endpoint onto the root engine.
func (h *Handler) RegisterRoutes(e *gin.Engine) {
	e.GET("/health", h.HealthCheck)
}

