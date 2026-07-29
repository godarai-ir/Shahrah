package admin

import (
	"fmt"
	"html"
	"mime"
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

type tenantAdminOTPEmail struct {
	subject string
	text    string
	html    string
}

const tenantAdminEmailBrandName = "گدارAI"
const tenantAdminEmailProductURL = "https://godarai.ir"
const tenantAdminEmailSupportEmail = "support@godarai.ir"
const defaultResendFromEmail = "auth@notifications.godarai.ir"
const defaultResendFromName = tenantAdminEmailBrandName
const defaultResendReplyTo = tenantAdminEmailSupportEmail

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
		fromEmail: resolveResendFromEmail(os.Getenv("RESEND_FROM_EMAIL")),
		fromName:  defaultIfEmpty(strings.TrimSpace(os.Getenv("RESEND_FROM_NAME")), defaultResendFromName),
		replyTo:   defaultIfEmpty(strings.TrimSpace(os.Getenv("RESEND_REPLY_TO")), defaultResendReplyTo),
		client: resend.NewCustomClient(&http.Client{
			Timeout: 10 * time.Second,
		}, strings.TrimSpace(os.Getenv("RESEND_API_KEY"))),
	}
}

func resolveResendFromEmail(configuredEmail string) string {
	normalizedEmail := strings.TrimSpace(configuredEmail)
	if strings.HasSuffix(strings.ToLower(normalizedEmail), "@notifications.godarai.ir") {
		return normalizedEmail
	}

	return defaultResendFromEmail
}

func (s *smtpOTPEmailSender) SendTenantAdminOTP(to string, tenantSlug string, otpCode string, expiresInMinutes int) error {
	if !s.enabled {
		return nil
	}

	to = strings.TrimSpace(to)
	if to == "" {
		return nil
	}

	email := renderTenantAdminOTPEmail(tenantSlug, otpCode, expiresInMinutes)
	boundary := fmt.Sprintf("tenant-admin-otp-%d", time.Now().UnixNano())

	message := strings.Join([]string{
		fmt.Sprintf("From: %s", s.from),
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: %s", mime.BEncoding.Encode("UTF-8", email.subject)),
		"MIME-Version: 1.0",
		fmt.Sprintf("Content-Type: multipart/alternative; boundary=%q", boundary),
		"",
		fmt.Sprintf("--%s", boundary),
		`Content-Type: text/plain; charset="UTF-8"`,
		"Content-Transfer-Encoding: 8bit",
		"",
		email.text,
		fmt.Sprintf("--%s", boundary),
		`Content-Type: text/html; charset="UTF-8"`,
		"Content-Transfer-Encoding: 8bit",
		"",
		email.html,
		fmt.Sprintf("--%s--", boundary),
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

	if s.apiKey == "" {
		return fmt.Errorf("RESEND_API_KEY is required when NODE_ENV=production")
	}

	email := renderTenantAdminOTPEmail(tenantSlug, otpCode, expiresInMinutes)

	payload := &resend.SendEmailRequest{
		From:    fmt.Sprintf("%s <%s>", s.fromName, s.fromEmail),
		To:      []string{to},
		Subject: email.subject,
		Text:    email.text,
		Html:    email.html,
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

func renderTenantAdminOTPEmail(tenantSlug string, otpCode string, expiresInMinutes int) tenantAdminOTPEmail {
	tenantSlug = normalizeEmailTextLine(tenantSlug)
	otpCode = normalizeEmailTextLine(otpCode)
	workspaceURL := resolveTenantAdminWorkspaceURL(tenantSlug)
	logoURL := resolveTenantAdminEmailLogoURL(workspaceURL)
	subject := fmt.Sprintf("کد ورود یک\u200cبارمصرف %s", tenantAdminEmailBrandName)
	text := strings.Join([]string{
		fmt.Sprintf("کد ورود یک\u200cبارمصرف شما برای فضای کاری «%s»: %s", tenantSlug, otpCode),
		fmt.Sprintf("این کد تا %d دقیقه معتبر است.", expiresInMinutes),
		"برای امنیت حساب، این کد را در اختیار هیچ\u200cکس قرار ندهید.",
		fmt.Sprintf("ورود به فضای کاری: %s", workspaceURL),
		"",
		"اگر شما درخواست ورود نداده\u200cاید، این ایمیل را نادیده بگیرید.",
		"",
		fmt.Sprintf("تیم %s", tenantAdminEmailBrandName),
		tenantAdminEmailSupportEmail,
	}, "\n")

	logoHTML := fmt.Sprintf(`<div style="display: inline-block; border-radius: 16px; background: #ffffff; border: 1px solid #d6e2f0; padding: 10px 14px; color: #0238e1; font-size: 18px; font-weight: 800;">%s</div>`, tenantAdminEmailBrandName)
	if logoURL != "" {
		logoHTML = fmt.Sprintf(`<img src=%q width="52" height="52" alt=%q style="display: block; width: 52px; height: 52px; border: 0;" />`, html.EscapeString(logoURL), tenantAdminEmailBrandName)
	}

	htmlBody := fmt.Sprintf(`<!doctype html>
<html lang="fa" dir="rtl">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>%s</title>
  </head>
  <body dir="rtl" style="margin: 0; padding: 0; background: #f3f7fb; color: #0b1220; font-family: Tahoma, Arial, sans-serif;">
    <div style="display: none; max-height: 0; overflow: hidden; opacity: 0; color: transparent;">
      کد ورود شما به فضای کاری %s تا %d دقیقه معتبر است.
    </div>
    <table role="presentation" width="100%%" cellspacing="0" cellpadding="0" style="width: 100%%; border-collapse: collapse; background: #f3f7fb;">
      <tr>
        <td align="center" style="padding: 32px 16px;">
          <table role="presentation" width="640" cellspacing="0" cellpadding="0" style="width: 100%%; max-width: 640px; border-collapse: collapse;">
            <tr>
              <td style="padding: 0 0 14px; text-align: right;">%s</td>
            </tr>
            <tr>
              <td style="overflow: hidden; border: 1px solid #d6e2f0; border-radius: 20px; background: #ffffff; box-shadow: 0 18px 42px rgba(15, 23, 42, 0.08);">
                <table role="presentation" width="100%%" cellspacing="0" cellpadding="0" style="border-collapse: collapse;">
                  <tr>
                    <td style="height: 5px; background: linear-gradient(90deg, #04b8a9 0%%, #0238e1 100%%); font-size: 0; line-height: 0;">&nbsp;</td>
                  </tr>
                  <tr>
                    <td style="padding: 34px 34px 28px;">
                      <p style="margin: 0 0 10px; color: #0238e1; font-size: 13px; font-weight: 700;">ورود امن</p>
                      <h1 style="margin: 0; color: #0b1220; font-size: 25px; line-height: 1.6; font-weight: 800;">کد ورود یک&#8204;بارمصرف</h1>
                      <p style="margin: 10px 0 0; color: #475569; font-size: 15px; line-height: 1.9;">این کد فقط برای همین درخواست ورود صادر شده است.</p>
                      <div style="margin-top: 26px;">
                        <p style="margin: 0 0 14px; color: #334155; font-size: 15px; line-height: 2;">
                          از کد زیر برای ورود به فضای کاری <strong dir="ltr">%s</strong> استفاده کنید.
                        </p>
                        <div dir="ltr" style="margin: 22px 0; border-radius: 14px; border: 1px solid #d6e2f0; background: #f8fbff; padding: 18px 20px; text-align: center; color: #0b1220;
                          font-family: ui-monospace, SFMono-Regular, Consolas, 'Liberation Mono', Menlo, monospace; font-size: 30px; font-weight: 800; letter-spacing: 8px;">%s</div>
                        <div style="margin: 16px 0 0; border-radius: 12px; border: 1px solid #d6e2f0; background: #ffffff; padding: 13px 16px;">
                          <p style="margin: 0; color: #64748b; font-size: 12px; line-height: 1.7;">اعتبار کد</p>
                          <p style="margin: 3px 0 0; color: #0b1220; font-size: 14px; font-weight: 700; line-height: 1.7;">%d دقیقه</p>
                        </div>
                        <div style="margin: 18px 0 0; border-radius: 12px; border: 1px solid #d6e2f0; background: #f8fbff; padding: 14px 16px;">
                          <p style="margin: 0 0 6px; color: #64748b; font-size: 12px; font-weight: 700;">ورود به فضای کاری</p>
                          <a dir="ltr" href="%s" style="display: block; color: #0238e1; font-family: Arial, sans-serif; font-size: 13px; line-height: 1.7; text-align: left; text-decoration: none; word-break: break-all;">%s</a>
                        </div>
                        <p style="margin: 22px 0 0; border-right: 3px solid #04b8a9; padding: 10px 14px 10px 0; color: #475569; font-size: 13px; line-height: 1.9; background: #f8fbff;">
                          برای امنیت حساب، این کد را در اختیار هیچ&#8204;کس قرار ندهید. اگر شما درخواست ورود نداده&#8204;اید، این پیام را نادیده بگیرید.
                        </p>
                      </div>
                    </td>
                  </tr>
                  <tr>
                    <td style="border-top: 1px solid #e2e8f0; padding: 22px 34px 28px; background: #fbfdff;">
                      <p style="margin: 0; color: #64748b; font-size: 12px; line-height: 1.9;">این پیام به درخواست شما برای ورود به پنل کاربری گدارAI ارسال شده است.</p>
                      <p style="margin: 8px 0 0; color: #64748b; font-size: 12px; line-height: 1.9;">
                        <a href="%s" style="color: #0238e1; text-decoration: none;">%s</a>
                        <span style="padding: 0 6px; color: #cbd5e1;">|</span>
                        <a href="mailto:%s" style="color: #0238e1; text-decoration: none;">%s</a>
                      </p>
                    </td>
                  </tr>
                </table>
              </td>
            </tr>
          </table>
        </td>
      </tr>
    </table>
  </body>
</html>`,
		html.EscapeString(subject),
		html.EscapeString(tenantSlug),
		expiresInMinutes,
		logoHTML,
		html.EscapeString(tenantSlug),
		html.EscapeString(otpCode),
		expiresInMinutes,
		html.EscapeString(workspaceURL),
		html.EscapeString(workspaceURL),
		tenantAdminEmailProductURL,
		tenantAdminEmailProductURL,
		tenantAdminEmailSupportEmail,
		tenantAdminEmailSupportEmail,
	)

	return tenantAdminOTPEmail{subject: subject, text: text, html: htmlBody}
}

func resolveTenantAdminEmailLogoURL(workspaceURL string) string {
	if value := strings.TrimSpace(os.Getenv("GODARAI_EMAIL_LOGO_URL")); value != "" {
		return value
	}

	if value := strings.TrimSpace(os.Getenv("SHAHRAH_EMAIL_LOGO_URL")); value != "" {
		return value
	}

	if workspaceURL == "" {
		return ""
	}

	return strings.TrimRight(workspaceURL, "/") + "/brand/mini_logo.png"
}

func resolveTenantAdminWorkspaceURL(tenantSlug string) string {
	if value := strings.TrimSpace(os.Getenv("NEXT_PUBLIC_TENANT_ADMIN_URL")); value != "" {
		return value
	}

	if value := strings.TrimSpace(os.Getenv("SHAHRAH_TENANT_ADMIN_URL")); value != "" {
		return value
	}

	baseDomain := strings.TrimSpace(os.Getenv("SHAHRAH_TENANT_BASE_DOMAIN"))
	if baseDomain == "" {
		baseDomain = strings.TrimSpace(os.Getenv("TENANT_BASE_DOMAIN"))
	}

	if baseDomain == "" {
		baseDomain = "godarai.ir"
	}

	baseDomain = strings.TrimRight(strings.TrimPrefix(strings.TrimPrefix(baseDomain, "https://"), "http://"), "/")

	if tenantSlug == "" {
		return "https://" + baseDomain
	}

	return "https://" + tenantSlug + "." + baseDomain
}

func normalizeEmailTextLine(value string) string {
	return strings.Join(strings.Fields(strings.ReplaceAll(strings.ReplaceAll(value, "\r", " "), "\n", " ")), " ")
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
