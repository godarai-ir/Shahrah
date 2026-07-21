package admin

import (
	"strings"
	"testing"
)

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
