package service

import (
	ebmetrics "emotionalBeach/internal/infra"
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

// SendEmail delivers an HTML email to all valid receiver addresses.
func (s *Service) SendEmail(subject, htmlContent string, receivers []string) error {
	if len(receivers) == 0 {
		return errors.New("receivers list is empty")
	}
	valid := filterValidEmails(receivers)
	if len(valid) == 0 {
		return errors.New("no valid recipient addresses")
	}
	if s.mailFrom == "" || s.smtpUser == "" || s.smtpPwd == "" {
		ebmetrics.EmailSentTotal.WithLabelValues("failure").Inc()
		return errors.New("SMTP credentials are not configured")
	}

	m := gomail.NewMessage()
	m.SetHeader("From", s.mailFrom)
	m.SetHeader("To", valid...)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", htmlToPlainText(htmlContent))
	m.AddAlternative("text/html", htmlContent)

	d := gomail.NewDialer(smtpHost, smtpPort, s.smtpUser, s.smtpPwd)
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

func filterValidEmails(addrs []string) []string {
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

func htmlToPlainText(html string) string {
	return regexp.MustCompile(`<[^>]*>`).ReplaceAllString(html, "")
}

