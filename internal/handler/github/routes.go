package githubhandler

import "github.com/gin-gonic/gin"

// RegisterRoutes mounts GitHub OAuth endpoints onto the root engine.
func (h *Handler) RegisterRoutes(e *gin.Engine) {
	e.GET("/login/github", h.GithubLogin)
	e.GET("/callback", h.GithubCallback)
}

