// Package webhookhandler wires the HTTP presentation layer for webhook / email notifications.
package webhookhandler

import (
	"emotionalBeach/internal/service/notification"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// message is the inbound JSON body for the webhook endpoint.
type message struct {
	Title     string   `json:"title"     binding:"required"`
	Content   string   `json:"content"`
	Receivers []string `json:"receivers"`
}

// Handler holds the injected EmailSender.
type Handler struct {
	sender *notification.EmailSender
}

// NewHandler constructs a Handler.
func NewHandler(sender *notification.EmailSender) *Handler {
	return &Handler{sender: sender}
}

// WebhookEmail godoc
// @Summary      Webhook 对外接口
// @Description  根据标题、内容、邮箱列表异步发送邮件
// @Tags         API
// @Accept       application/json
// @Produce      application/json
// @Param        message body message true "请求参数"
// @Security     ApiKeyAuth
// @Router       /v1/api/webhook [post]
func (h *Handler) WebhookEmail(c *gin.Context) {
	var msg message
	if err := c.ShouldBindJSON(&msg); err != nil {
		zap.S().Warnf("webhook: invalid JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"status":  "error",
			"message": fmt.Sprintf("invalid JSON: %s", err.Error()),
		})
		return
	}
	if len(msg.Receivers) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"status":  "error",
			"message": "at least one receiver is required",
		})
		return
	}

	// Fire-and-forget — respond immediately, send in background.
	go func() {
		if err := h.sender.Send(msg.Title, msg.Content, msg.Receivers); err != nil {
			zap.S().Errorf("webhook: background email failed: %v", err)
		}
	}()

	zap.S().Infof("✅ webhook received, queued email for %v", msg.Receivers)
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"status":  "success",
		"message": "webhook received and email task queued",
	})
}

