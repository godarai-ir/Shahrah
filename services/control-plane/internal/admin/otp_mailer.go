package admin

import (
	"fmt"
	"net/smtp"
	"os"
	"strconv"
	"strings"
)

type OTPEmailSender interface {
	SendTenantAdminOTP(to string, tenantSlug string, otpCode string, expiresInMinutes int) error
}

type smtpOTPEmailSender struct {
	enabled bool
	host    string
	port    int
	from    string
	user    string
	pass    string
}

func NewSMTPOTPEmailSenderFromEnv() OTPEmailSender {
	enabled := strings.EqualFold(strings.TrimSpace(os.Getenv("EMAIL_SENDING_ENABLED")), "true")
	if v := strings.TrimSpace(os.Getenv("EMAIL_SENDING_ENABLED")); v == "" {
		enabled = true
	}

	port := 1025
	if raw := strings.TrimSpace(os.Getenv("SMTP_PORT")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			port = parsed
		}
	}

	return &smtpOTPEmailSender{
		enabled: enabled,
		host:    defaultIfEmpty(strings.TrimSpace(os.Getenv("SMTP_HOST")), "127.0.0.1"),
		port:    port,
		from:    defaultIfEmpty(strings.TrimSpace(os.Getenv("SMTP_FROM")), "no-reply@shahrah.local"),
		user:    strings.TrimSpace(os.Getenv("SMTP_USER")),
		pass:    strings.TrimSpace(os.Getenv("SMTP_PASS")),
	}
}

func (s *smtpOTPEmailSender) SendTenantAdminOTP(to string, tenantSlug string, otpCode string, expiresInMinutes int) error {
	if !s.enabled {
		return nil
	}

	to = strings.TrimSpace(to)
	if to == "" {
		return nil
	}

	subject := "Shahrah OTP Code"
	body := strings.Join([]string{
		fmt.Sprintf("Your one-time code for tenant \"%s\" is: %s", tenantSlug, otpCode),
		fmt.Sprintf("This code expires in %d minutes.", expiresInMinutes),
		"If you did not request this code, ignore this email.",
	}, "\r\n")

	message := strings.Join([]string{
		fmt.Sprintf("From: %s", s.from),
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		`Content-Type: text/plain; charset="UTF-8"`,
		"",
		body,
	}, "\r\n")

	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	var auth smtp.Auth
	if s.user != "" && s.pass != "" {
		auth = smtp.PlainAuth("", s.user, s.pass, s.host)
	}

	return smtp.SendMail(addr, auth, s.from, []string{to}, []byte(message))
}

func defaultIfEmpty(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
