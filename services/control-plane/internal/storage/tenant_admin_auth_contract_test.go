package storage

import (
	"encoding/json"
	"testing"
)

func TestTenantAdminOTPResponseDoesNotExposeInternalCode(t *testing.T) {
	payload, err := json.Marshal(TenantAdminOTPResponse{
		RequestID:         "request-1",
		Channel:           "email",
		DestinationMasked: "us***@example.com",
		ExpiresInSeconds:  300,
		OTPCode:           "123456",
	})
	if err != nil {
		t.Fatalf("marshal OTP response: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("unmarshal OTP response: %v", err)
	}

	if _, exposed := decoded["otp_code"]; exposed {
		t.Fatal("internal OTP code was exposed")
	}

	if decoded["request_id"] != "request-1" || decoded["expires_in_seconds"] != float64(300) {
		t.Fatalf("unexpected OTP response: %s", payload)
	}
}

func TestTenantAdminUserResponsePreservesLegacyWireFields(t *testing.T) {
	payload, err := json.Marshal(TenantAdminUserResponse{
		ID:           "user-1",
		TenantSlug:   "tenant-1",
		CreatedAtISO: "2026-08-07T00:00:00Z",
		UpdatedAtISO: "2026-08-07T00:01:00Z",
	})
	if err != nil {
		t.Fatalf("marshal user response: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("unmarshal user response: %v", err)
	}

	if decoded["created_at_iso"] != "2026-08-07T00:00:00Z" || decoded["updated_at_iso"] != "2026-08-07T00:01:00Z" {
		t.Fatalf("legacy timestamp fields changed: %s", payload)
	}
}
