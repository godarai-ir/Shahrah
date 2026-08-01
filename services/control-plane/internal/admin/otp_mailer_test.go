package admin

import (
	"strings"
	"testing"
)

func TestNewResendOTPEmailSenderUsesGodarAINotificationsDefaults(t *testing.T) {
	t.Setenv("RESEND_API_KEY", "test-api-key")
	t.Setenv("RESEND_FROM_EMAIL", "noreply@godarai.ir")
	t.Setenv("RESEND_FROM_NAME", "")
	t.Setenv("RESEND_REPLY_TO", "")

	sender, ok := newResendOTPEmailSenderFromEnv().(*resendOTPEmailSender)
	if !ok {
		t.Fatal("newResendOTPEmailSenderFromEnv returned an unexpected sender type")
	}

	if sender.fromEmail != defaultResendFromEmail {
		t.Fatalf("fromEmail = %q, want %q", sender.fromEmail, defaultResendFromEmail)
	}

	if sender.fromName != defaultResendFromName {
		t.Fatalf("fromName = %q, want %q", sender.fromName, defaultResendFromName)
	}

	if sender.replyTo != defaultResendReplyTo {
		t.Fatalf("replyTo = %q, want %q", sender.replyTo, defaultResendReplyTo)
	}
}

func TestResolveResendFromEmailAllowsVerifiedNotificationsSubdomain(t *testing.T) {
	const configuredEmail = "notifications@notifications.godarai.ir"

	if got := resolveResendFromEmail(configuredEmail); got != configuredEmail {
		t.Fatalf("resolveResendFromEmail() = %q, want %q", got, configuredEmail)
	}
}

func TestRenderTenantAdminOTPEmailIsBrandedRTLAndEscaped(t *testing.T) {
	email := renderTenantAdminOTPEmail(`acme<script>`, `123<456`, 5)

	if !strings.Contains(email.subject, tenantAdminEmailBrandName) {
		t.Fatalf("subject = %q, want brand name", email.subject)
	}

	if !strings.Contains(email.text, "کد ورود یک\u200cبارمصرف شما") {
		t.Fatalf("text body did not include Persian OTP copy: %q", email.text)
	}

	if !strings.Contains(email.html, `dir="rtl"`) {
		t.Fatal("html body did not declare RTL direction")
	}

	if !strings.Contains(email.html, "کد ورود یک&#8204;بارمصرف") {
		t.Fatal("html body did not include Persian OTP heading")
	}

	if !strings.Contains(email.html, "123&lt;456") {
		t.Fatal("html body did not escape OTP code")
	}

	if strings.Contains(email.html, "<script>") {
		t.Fatal("html body included unescaped tenant content")
	}
}
