package admin

import (
	"fmt"
	"net/http"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/resend/resend-go/v3"
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

type resendOTPEmailSender struct {
	enabled   bool
	apiKey    string
	fromEmail string
	fromName  string
	replyTo   string
	client    *resend.Client
}

func NewOTPEmailSenderFromEnv() OTPEmailSender {
	if strings.EqualFold(strings.TrimSpace(os.Getenv("NODE_ENV")), "production") {
		return newResendOTPEmailSenderFromEnv()
	}
	return newSMTPOTPEmailSenderFromEnv()
}

func newSMTPOTPEmailSenderFromEnv() OTPEmailSender {
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

func newResendOTPEmailSenderFromEnv() OTPEmailSender {
	enabled := strings.EqualFold(strings.TrimSpace(os.Getenv("EMAIL_SENDING_ENABLED")), "true")
	if v := strings.TrimSpace(os.Getenv("EMAIL_SENDING_ENABLED")); v == "" {
		enabled = true
	}

	return &resendOTPEmailSender{
		enabled:   enabled,
		apiKey:    strings.TrimSpace(os.Getenv("RESEND_API_KEY")),
		fromEmail: strings.TrimSpace(os.Getenv("RESEND_FROM_EMAIL")),
		fromName:  defaultIfEmpty(strings.TrimSpace(os.Getenv("RESEND_FROM_NAME")), "Shahrah"),
		replyTo:   strings.TrimSpace(os.Getenv("RESEND_REPLY_TO")),
		client: resend.NewCustomClient(&http.Client{
			Timeout: 10 * time.Second,
		}, strings.TrimSpace(os.Getenv("RESEND_API_KEY"))),
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

func (s *resendOTPEmailSender) SendTenantAdminOTP(to string, tenantSlug string, otpCode string, expiresInMinutes int) error {
	if !s.enabled {
		return nil
	}

	to = strings.TrimSpace(to)
	if to == "" {
		return nil
	}
	if s.apiKey == "" || s.fromEmail == "" {
		return fmt.Errorf("RESEND_API_KEY and RESEND_FROM_EMAIL are required when NODE_ENV=production")
	}

	subject := "Shahrah OTP Code"
	text := strings.Join([]string{
		fmt.Sprintf("Your one-time code for tenant \"%s\" is: %s", tenantSlug, otpCode),
		fmt.Sprintf("This code expires in %d minutes.", expiresInMinutes),
		"If you did not request this code, ignore this email.",
	}, "\n")
	html := fmt.Sprintf(`
<div style="font-family: sans-serif; line-height: 1.6;">
  <h2>Shahrah OTP Code</h2>
  <p>Your one-time code for tenant <strong>%s</strong> is:</p>
  <p style="font-size: 24px; font-weight: bold; letter-spacing: 3px;">%s</p>
  <p>This code expires in %d minutes.</p>
  <p>If you did not request this code, ignore this email.</p>
</div>`, tenantSlug, otpCode, expiresInMinutes)

	payload := &resend.SendEmailRequest{
		From:    fmt.Sprintf("%s <%s>", s.fromName, s.fromEmail),
		To:      []string{to},
		Subject: subject,
		Text:    text,
		Html:    html,
		ReplyTo: s.replyTo,
		Tags: []resend.Tag{
			{Name: "app", Value: "control-plane"},
			{Name: "flow", Value: "tenant-admin-otp"},
			{Name: "tenant", Value: sanitizeResendTagValue(tenantSlug)},
		},
	}

	_, err := s.client.Emails.Send(payload)
	if err != nil {
		return err
	}
	return nil
}

func sanitizeResendTagValue(value string) string {
	var builder strings.Builder
	for _, r := range strings.TrimSpace(value) {
		switch {
		case r >= 'a' && r <= 'z':
			builder.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			builder.WriteRune(r)
		case r >= '0' && r <= '9':
			builder.WriteRune(r)
		case r == '-' || r == '_':
			builder.WriteRune(r)
		}
	}
	if builder.Len() == 0 {
		return "unknown"
	}
	return builder.String()
}

func defaultIfEmpty(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
