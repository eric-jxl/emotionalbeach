package service

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gopkg.in/gomail.v2"
)

// WebhookMessage  消息结构
type WebhookMessage struct {
	Title     string   `json:"title"`
	Content   string   `json:"content"`
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

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func isValidEmail(email string) bool {
	return emailRegex.MatchString(strings.TrimSpace(email))
}

// SendEmail 发送邮件
func SendEmail(subject, content string, receivers []string) (err error) {
	go func() {
		if len(receivers) == 0 {
			zap.S().Errorf("❌ 邮件发送失败: 收件人列表为空")
			err = errors.New("receivers is null")
			return
		}
		// 过滤并验证有效的邮箱地址
		var validReceivers []string
		for _, r := range receivers {
			if isValidEmail(r) {
				validReceivers = append(validReceivers, r)
			} else {
				zap.S().Warnf("⚠️ 无效的收件人邮箱被忽略: %s", r)
			}
		}

		if len(validReceivers) == 0 {
			zap.S().Errorf("❌ 邮件发送失败: 没有有效的收件人邮箱")
			err = errors.New("receivers is invalid")
			return
		}
		m := gomail.NewMessage()
		m.SetHeader("From", os.Getenv("MAIL_FROM"))
		m.SetHeader("To", receivers...) // 支持多个收件人
		m.SetHeader("Subject", subject)

		// 同时设置纯文本和 HTML
		plainText := htmlToPlain(content)
		m.SetBody("text/plain", plainText)
		m.AddAlternative("text/html", content)

		d := gomail.NewDialer(smtpHost, smtpPort, os.Getenv("SmtpUser"), os.Getenv("SmtpPassword"))
		if err = d.DialAndSend(m); err != nil {
			zap.S().Errorf("❌ 邮件发送失败: %v\n", err)
			return
		} else {
			zap.S().Infof("📧 已发送 HTML 邮件: [%s] 给 %v\n", subject, receivers)
		}
	}()
	return nil
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

	// 发送邮件
	errs := SendEmail(msg.Title, msg.Content, msg.Receivers)
	if errs != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errs, "code": http.StatusInternalServerError})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"status":  "success",
		"message": "Webhook received and email sent",
	})
}
