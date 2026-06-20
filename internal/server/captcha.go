package server

import (
	"emotionalBeach/config"
	"emotionalBeach/internal/common"
	"net/http"

	"github.com/gin-gonic/gin"
)

// captchaPublicConfig returns runtime captcha settings without hardcoding values in frontend assets.
// Requirement: captcha identity/scene parameters are managed in server config and fetched at runtime.
func captchaPublicConfig(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if cfg == nil {
			common.ServerError(c, "captcha config unavailable")
			return
		}

		prefix := cfg.Server.ESACaptchaPrefix
		sceneID := cfg.Server.ESACaptchaSceneID
		if prefix == "" || sceneID == "" {
			// Explicitly indicate captcha is disabled/misconfigured so frontend can block login.
			common.Fail(c, http.StatusNotFound, "captcha not configured")
			return
		}

		// Avoid proxy/browser caching for security-sensitive runtime config.
		c.Header("Cache-Control", "no-store")
		c.Header("Pragma", "no-cache")
		c.Header("X-Content-Type-Options", "nosniff")
		common.Success(c, gin.H{
			"region":  cfg.Server.CaptchaRegion(),
			"prefix":  prefix,
			"sceneId": sceneID,
		})
	}
}

