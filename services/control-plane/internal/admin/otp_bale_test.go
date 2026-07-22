package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBaleSafirOTPSenderSendsDocumentedOTPPayload(t *testing.T) {
	t.Helper()

	var received baleSafirOTPRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if got := r.Header.Get("api-access-key"); got != "safir-secret" {
			t.Errorf("api-access-key = %q, want safir-secret", got)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", got)
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"message_id":"BvQjaR.fIKt7kH.EXTddgYduJ2"}`))
	}))
	t.Cleanup(server.Close)

	sender := &baleSafirOTPSender{
		enabled:      true,
		apiAccessKey: "safir-secret",
		botID:        123456789,
		endpoint:     server.URL,
		client:       server.Client(),
		requestID:    func() string { return "request-123" },
	}

	if err := sender.SendTenantAdminOTP("0912-345-6789", "tenant1", "654321", 5); err != nil {
		t.Fatalf("SendTenantAdminOTP returned error: %v", err)
	}

	if received.RequestID != "request-123" {
		t.Fatalf("request_id = %q, want request-123", received.RequestID)
	}
	if received.BotID != 123456789 {
		t.Fatalf("bot_id = %d, want 123456789", received.BotID)
	}
	if received.PhoneNumber != "989123456789" {
		t.Fatalf("phone_number = %q, want 989123456789", received.PhoneNumber)
	}
	if received.MessageData.OTPMessage.OTP != "654321" {
		t.Fatalf("otp = %q, want 654321", received.MessageData.OTPMessage.OTP)
	}
}

func TestBaleSafirOTPSenderReturnsProviderError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"message_id":"",
			"error_data":[{"phone_number":"989123456789","code":17,"description":"NotBaleUser"}]
		}`))
	}))
	t.Cleanup(server.Close)

	sender := &baleSafirOTPSender{
		enabled:      true,
		apiAccessKey: "safir-secret",
		botID:        123456789,
		endpoint:     server.URL,
		client:       server.Client(),
		requestID:    func() string { return "request-123" },
	}

	err := sender.SendTenantAdminOTP("+989123456789", "tenant1", "654321", 5)
	if err == nil || !strings.Contains(err.Error(), "code 17") {
		t.Fatalf("error = %v, want provider error code 17", err)
	}
	if strings.Contains(err.Error(), "989123456789") {
		t.Fatal("provider error leaked the destination phone number")
	}
}

func TestNormalizeBalePhoneNumber(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "local", input: "09123456789", want: "989123456789"},
		{name: "local without zero", input: "9123456789", want: "989123456789"},
		{name: "country code", input: "989123456789", want: "989123456789"},
		{name: "plus country code", input: "+989123456789", want: "989123456789"},
		{name: "international prefix", input: "00989123456789", want: "989123456789"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeBalePhoneNumber(tt.input)
			if err != nil {
				t.Fatalf("normalizeBalePhoneNumber(%q): %v", tt.input, err)
			}
			if got != tt.want {
				t.Fatalf("normalizeBalePhoneNumber(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNormalizeBalePhoneNumberRejectsInvalidNumber(t *testing.T) {
	if _, err := normalizeBalePhoneNumber("0912345"); err == nil {
		t.Fatal("normalizeBalePhoneNumber accepted an invalid number")
	}
}
