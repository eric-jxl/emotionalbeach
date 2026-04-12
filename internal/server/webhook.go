package server

import (
	"emotionalBeach/internal/common"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type webhookMessage struct {
	Title     string   `json:"title"     binding:"required"`
	Content   string   `json:"content"`
	Receivers []string `json:"receivers"`
}

// webhookEmail godoc
// @Summary      Webhook 对外接口
// @Tags         API
// @Accept       application/json
// @Produce      application/json
// @Param        message body webhookMessage true "请求参数"
// @Security     ApiKeyAuth
// @Router       /v1/api/webhook [post]
func webhookEmail(c *gin.Context) {
	var msg webhookMessage
	if err := c.ShouldBindJSON(&msg); err != nil {
		zap.S().Warnf("webhook: invalid JSON: %v", err)
		common.Fail(c, http.StatusBadRequest, fmt.Sprintf("invalid JSON: %s", err.Error()))
		return
	}
	if len(msg.Receivers) == 0 {
		common.Fail(c, http.StatusBadRequest, "at least one receiver is required")
		return
	}
	go func() {
		if err := svc.SendEmail(msg.Title, msg.Content, msg.Receivers); err != nil {
			zap.S().Errorf("webhook: background email failed: %v", err)
		}
	}()
	zap.S().Infof("✅ webhook received, queued email for %v", msg.Receivers)
	common.Success(c, gin.H{"message": "webhook received and email task queued"})
}
