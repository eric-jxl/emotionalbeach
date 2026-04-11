// Package notification provides an email sender backed by SMTP configuration.
package notification

import (
	"emotionalBeach/config"
	ebmetrics "emotionalBeach/internal/infra/metrics"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"go.uber.org/zap"
	"gopkg.in/gomail.v2"
)

const (
	smtpHost = "smtp.qq.com"
	smtpPort = 587
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// EmailSender sends HTML emails via a configured SMTP server.
// All credentials are injected from config at construction time — no env vars needed.
type EmailSender struct {
	mailFrom string
	smtpUser string
	smtpPwd  string
}

// NewEmailSender constructs an EmailSender from application config.
func NewEmailSender(cfg *config.Config) *EmailSender {
	return &EmailSender{
		mailFrom: cfg.MailConfig.MailFrom,
		smtpUser: cfg.MailConfig.SmtpUser,
		smtpPwd:  cfg.MailConfig.SmtpPassword,
	}
}

// Send delivers an HTML email to all valid receiver addresses.
func (e *EmailSender) Send(subject, htmlContent string, receivers []string) error {
	if len(receivers) == 0 {
		return errors.New("receivers list is empty")
	}
	valid := filterValid(receivers)
	if len(valid) == 0 {
		return errors.New("no valid recipient addresses")
	}
	if e.mailFrom == "" || e.smtpUser == "" || e.smtpPwd == "" {
		ebmetrics.EmailSentTotal.WithLabelValues("failure").Inc()
		return errors.New("SMTP credentials are not configured")
	}

	m := gomail.NewMessage()
	m.SetHeader("From", e.mailFrom)
	m.SetHeader("To", valid...)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", htmlToPlain(htmlContent))
	m.AddAlternative("text/html", htmlContent)

	d := gomail.NewDialer(smtpHost, smtpPort, e.smtpUser, e.smtpPwd)
	if err := d.DialAndSend(m); err != nil {
		zap.S().Errorf("❌ send email [%s] failed: %v", subject, err)
		ebmetrics.EmailSentTotal.WithLabelValues("failure").Inc()
		return fmt.Errorf("dial and send: %w", err)
	}
	ebmetrics.EmailSentTotal.WithLabelValues("success").Inc()
	ebmetrics.EmailReceiversTotal.Add(float64(len(valid)))
	zap.S().Infof("📧 email [%s] sent to %v", subject, valid)
	return nil
}

// filterValid removes blank and malformed email addresses.
func filterValid(addrs []string) []string {
	out := make([]string, 0, len(addrs))
	for _, a := range addrs {
		a = strings.TrimSpace(a)
		if emailRegex.MatchString(a) {
			out = append(out, a)
		} else {
			zap.S().Warnf("⚠️  invalid email ignored: %q", a)
		}
	}
	return out
}

func htmlToPlain(html string) string {
	return regexp.MustCompile(`<[^>]*>`).ReplaceAllString(html, "")
}

