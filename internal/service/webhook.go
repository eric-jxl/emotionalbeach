package service

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"regexp"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gopkg.in/gomail.v2"
)

// WebhookMessage  消息结构
type WebhookMessage struct {
	Title     string   `json:"title"`
	Content   string   `json:"content"`
	Url       string   `json:"url"`
	Receivers []string `json:"receivers"` // 支持多个收件人
}

// 邮件配置
const (
	smtpHost = "smtp.qq.com"
	smtpPort = 587
)

func htmlToPlain(html string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(html, "")
}

// SendEmail 发送邮件
func SendEmail(subject, content string, receivers []string) {
	go func() {
		m := gomail.NewMessage()
		m.SetHeader("From", os.Getenv("MAIL_FROM"))
		m.SetHeader("To", receivers...) // 支持多个收件人
		m.SetHeader("Subject", subject)

		// 同时设置纯文本和 HTML
		plainText := htmlToPlain(content)
		m.SetBody("text/plain", plainText)
		m.AddAlternative("text/html", content)

		d := gomail.NewDialer(smtpHost, smtpPort, os.Getenv("SmtpUser"), os.Getenv("SmtpPassword"))
		if err := d.DialAndSend(m); err != nil {
			zap.S().Errorf("❌ 邮件发送失败: %v\n", err)
		} else {
			zap.S().Infof("📧 已发送 HTML 邮件: [%s] 给 %v\n", subject, receivers)
		}
	}()
}

func WebhookEmail(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot read body"})
		return
	}

	// 解析 JSON
	var msg WebhookMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	if len(msg.Receivers) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "receivers cannot be empty"})
		return
	}

	// 发送邮件
	SendEmail(msg.Title, msg.Content, msg.Receivers)

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"status":  "success",
		"message": "Webhook received and email sent",
	})
}
