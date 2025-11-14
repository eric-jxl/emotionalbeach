package service

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gopkg.in/gomail.v2"
)

// WebhookMessage 消息结构
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

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func isValidEmail(email string) bool {
	return emailRegex.MatchString(strings.TrimSpace(email))
}

// htmlToPlain 将 HTML 内容转换为纯文本
func htmlToPlain(html string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(html, "")
}

// sendEmailSync 同步发送邮件，并返回结果
// 移除了不必要的 done channel 和外部 err，逻辑更清晰
func sendEmailSync(subject, content string, receivers []string) error {
	// 1. 参数预检查
	if len(receivers) == 0 {
		zap.S().Errorf("❌ 邮件发送失败: 收件人列表为空")
		return errors.New("receivers list is empty")
	}

	// 2. 过滤并验证有效的邮箱地址 (关键修复点!)
	var validReceivers []string
	for _, r := range receivers {
		if isValidEmail(r) {
			validReceivers = append(validReceivers, r)
		} else {
			zap.S().Warnf("⚠️ 无效或空的收件人邮箱被忽略: '%s'", r)
		}
	}

	if len(validReceivers) == 0 {
		zap.S().Errorf("❌ 邮件发送失败: 没有找到任何有效的收件人邮箱")
		return errors.New("no valid recipient addresses found after filtering")
	}

	// 3. 创建邮件消息
	m := gomail.NewMessage()
	fromAddr := os.Getenv("MAIL_FROM")
	if fromAddr == "" {
		zap.S().Errorf("❌ 邮件发送失败: MAIL_FROM 环境变量未设置")
		return errors.New("MAIL_FROM environment variable is not set")
	}
	m.SetHeader("From", fromAddr)

	// ✅ 安全: 使用经过滤的 validReceivers
	m.SetHeader("To", validReceivers...)

	m.SetHeader("Subject", subject)

	// 设置纯文本和 HTML 正文
	plainText := htmlToPlain(content)
	m.SetBody("text/plain", plainText)
	m.AddAlternative("text/html", content)

	// 4. 配置 SMTP 拨号器
	smtpUser := os.Getenv("SmtpUser")
	smtpPassword := os.Getenv("SmtpPassword")
	if smtpUser == "" || smtpPassword == "" {
		zap.S().Errorf("❌ 邮件发送失败: SmtpUser 或 SmtpPassword 环境变量未设置")
		return errors.New("SMTP credentials are missing")
	}

	d := gomail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPassword)

	// 5. 发送邮件
	err := d.DialAndSend(m)
	if err != nil {
		zap.S().Errorf("❌ 邮件发送失败: %v", err)
		return fmt.Errorf("failed to dial and send email: %w", err)
	}

	// 6. 记录成功日志
	zap.S().Infof("📧 已成功发送邮件: [%s] 给 %v", subject, validReceivers)
	return nil
}

// WebhookEmail 处理 webhook 请求
func WebhookEmail(c *gin.Context) {
	var msg WebhookMessage

	// 1. 解析 JSON 请求体
	if err := c.ShouldBindJSON(&msg); err != nil { // 推荐使用 ShouldBindJSON
		zap.S().Warnf("❌ 无效的 JSON 请求: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"status":  "error",
			"message": fmt.Sprintf("Invalid JSON: %v", err.Error()),
		})
		return
	}

	// 2. 基本字段校验
	if msg.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"status":  "error",
			"message": "Title is required",
		})
		return
	}
	if msg.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"status":  "error",
			"message": "Content is required",
		})
		return
	}
	if len(msg.Receivers) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"status":  "error",
			"message": "At least one receiver is required",
		})
		return
	}

	// 3. 启动协程异步发送邮件
	go func() {
		// 使用新的同步函数，并捕获其返回的错误
		err := sendEmailSync(msg.Title, msg.Content, msg.Receivers)
		if err != nil {
			// 即使在协程中，我们也记录错误，以便排查问题
			zap.S().Errorf("📧 协程内邮件发送最终失败: %v", err)
		}
	}()

	zap.S().Infof("✅ Webhook received, queued email for %v", msg.Receivers)
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"status":  "success",
		"message": "Webhook received and email task queued",
	})
}
