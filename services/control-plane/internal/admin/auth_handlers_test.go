package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/storage"
	"github.com/gofiber/fiber/v2"
)

func TestCreateOTPRequestHidesUnknownLoginAccountBeforeCreatingOTP(t *testing.T) {
	store := &authHandlerTestStore{findUserErr: storage.ErrTenantAdminUserNotFound}
	handler := NewAuthHandlers(store, nil)
	app := fiber.New()
	app.Post("/auth/otp/request", handler.CreateOTPRequest)

	req := httptest.NewRequest("POST", "/auth/otp/request", bytes.NewBufferString(`{
		"tenant_slug": "tenant1",
		"purpose": "login",
		"channel": "email",
		"identifier": "missing@example.com"
	}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	t.Cleanup(func() { _ = resp.Body.Close() })

	var payload map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusOK)
	}

	if store.createOTPCalled {
		t.Fatal("CreateTenantAdminOTPRequest was called for an unknown login account")
	}

	if payload["request_id"] == "" {
		t.Fatal("generic OTP response did not include request_id")
	}

	if payload["destination_masked"] != "mi***@example.com" {
		t.Fatalf("destination_masked = %v, want mi***@example.com", payload["destination_masked"])
	}

	if _, exists := payload["debug_otp"]; exists {
		t.Fatal("debug_otp was exposed for a synthetic OTP response")
	}
}

func TestCreateOTPRequestRateLimitsRepeatedLoginAttempts(t *testing.T) {
	store := &authHandlerTestStore{createOTPResult: &storage.TenantAdminOTPResponse{
		RequestID:         "req-1",
		Channel:           "email",
		DestinationMasked: "ra***@example.com",
		ExpiresInSeconds:  300,
		OTPCode:           "123456",
	}}
	handler := NewAuthHandlers(store, nil)
	app := fiber.New()
	app.Post("/auth/otp/request", handler.CreateOTPRequest)

	var lastStatus int

	for attempt := 1; attempt <= 6; attempt++ {
		req := httptest.NewRequest("POST", "/auth/otp/request", bytes.NewBufferString(`{
			"tenant_slug": "tenant1",
			"purpose": "login",
			"channel": "email",
			"identifier": "ratelimited@example.com"
		}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Forwarded-For", "198.51.100.30")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request %d failed: %v", attempt, err)
		}

		lastStatus = resp.StatusCode
		_ = resp.Body.Close()
	}

	if lastStatus != fiber.StatusTooManyRequests {
		t.Fatalf("last status = %d, want %d", lastStatus, fiber.StatusTooManyRequests)
	}

	if store.createOTPCalls != 5 {
		t.Fatalf("CreateTenantAdminOTPRequest calls = %d, want 5", store.createOTPCalls)
	}
}

func TestCreateOTPRequestSendsMobileOTPThroughNotifier(t *testing.T) {
	store := &authHandlerTestStore{createOTPResult: &storage.TenantAdminOTPResponse{
		RequestID:         "req-mobile-1",
		Channel:           "mobile",
		DestinationMasked: "091****789",
		ExpiresInSeconds:  300,
		OTPCode:           "654321",
	}}
	notifier := &authHandlerTestOTPNotifier{}
	handler := NewAuthHandlers(store, notifier)
	app := fiber.New()
	app.Post("/auth/otp/request", handler.CreateOTPRequest)

	req := httptest.NewRequest("POST", "/auth/otp/request", bytes.NewBufferString(`{
		"tenant_slug": "tenant1",
		"purpose": "login",
		"channel": "mobile",
		"identifier": "09123456789"
	}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	t.Cleanup(func() { _ = resp.Body.Close() })

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusOK)
	}

	if notifier.channel != "mobile" || notifier.to != "09123456789" || notifier.otpCode != "654321" {
		t.Fatalf("notifier call = channel %q, to %q, otp %q", notifier.channel, notifier.to, notifier.otpCode)
	}
}

func TestCreateOTPRequestDoesNotExposeDebugOTPWhenNodeEnvIsEmpty(t *testing.T) {
	t.Setenv("NODE_ENV", "")
	t.Setenv("SHAHRAH_EXPOSE_DEBUG_OTP", "true")

	status, payload := performCreateOTPRequest(t)
	if status != fiber.StatusOK {
		t.Fatalf("status = %d, want %d", status, fiber.StatusOK)
	}

	if _, exists := payload["debug_otp"]; exists {
		t.Fatal("debug_otp was exposed when NODE_ENV was empty")
	}

	if _, exists := payload["otp_code"]; exists {
		t.Fatal("otp_code was exposed in the response")
	}
}

func TestCreateOTPRequestRequiresExplicitDebugOTPOptIn(t *testing.T) {
	t.Setenv("NODE_ENV", "development")
	t.Setenv("SHAHRAH_EXPOSE_DEBUG_OTP", "")

	_, payload := performCreateOTPRequest(t)
	if _, exists := payload["debug_otp"]; exists {
		t.Fatal("debug_otp was exposed without explicit opt-in")
	}
}

func TestCreateOTPRequestExposesDebugOTPOnlyWithNonProductionOptIn(t *testing.T) {
	t.Setenv("NODE_ENV", "development")
	t.Setenv("SHAHRAH_EXPOSE_DEBUG_OTP", "true")

	_, payload := performCreateOTPRequest(t)
	if got := payload["debug_otp"]; got != "123456" {
		t.Fatalf("debug_otp = %v, want 123456", got)
	}
}

func TestCreateOTPRequestNeverExposesDebugOTPInProduction(t *testing.T) {
	t.Setenv("NODE_ENV", "production")
	t.Setenv("SHAHRAH_EXPOSE_DEBUG_OTP", "true")

	_, payload := performCreateOTPRequest(t)
	if _, exists := payload["debug_otp"]; exists {
		t.Fatal("debug_otp was exposed in production")
	}
}

func performCreateOTPRequest(t *testing.T) (int, map[string]interface{}) {
	t.Helper()

	store := &authHandlerTestStore{createOTPResult: &storage.TenantAdminOTPResponse{
		RequestID:         "req-1",
		Channel:           "email",
		DestinationMasked: "u***@example.com",
		ExpiresInSeconds:  300,
		OTPCode:           "123456",
	}}
	handler := NewAuthHandlers(store, nil)
	app := fiber.New()
	app.Post("/auth/otp/request", handler.CreateOTPRequest)

	req := httptest.NewRequest("POST", "/auth/otp/request", bytes.NewBufferString(`{
		"tenant_slug": "tenant1",
		"purpose": "login",
		"channel": "email",
		"identifier": "user@example.com"
	}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	t.Cleanup(func() { _ = resp.Body.Close() })

	var payload map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	return resp.StatusCode, payload
}

type authHandlerTestStore struct {
	findUserErr     error
	createOTPCalled bool
	createOTPCalls  int
	createOTPResult *storage.TenantAdminOTPResponse
}

type authHandlerTestOTPNotifier struct {
	channel string
	to      string
	otpCode string
}

func (n *authHandlerTestOTPNotifier) SendTenantAdminOTP(channel string, to string, _ string, otpCode string, _ int) error {
	n.channel = channel
	n.to = to
	n.otpCode = otpCode

	return nil
}

func (s *authHandlerTestStore) CreateTenantAdminCaptcha(context.Context) (string, string, time.Time, error) {
	return "", "", time.Time{}, nil
}

func (s *authHandlerTestStore) ConsumeTenantAdminCaptcha(context.Context, string, string) (bool, error) {
	return true, nil
}

func (s *authHandlerTestStore) CreateTenantAdminOTPRequest(context.Context, storage.TenantAdminOTPRequestParams) (*storage.TenantAdminOTPResponse, error) {
	s.createOTPCalled = true

	s.createOTPCalls++
	if s.createOTPResult != nil {
		return s.createOTPResult, nil
	}

	return &storage.TenantAdminOTPResponse{}, nil
}

func (s *authHandlerTestStore) VerifyTenantAdminOTPRequest(context.Context, storage.VerifyTenantAdminOTPRequestParams) (*storage.TenantAdminOTPVerification, error) {
	return &storage.TenantAdminOTPVerification{}, nil
}

func (s *authHandlerTestStore) CreateTenantAdminSession(context.Context, storage.CreateTenantAdminSessionParams) (*storage.TenantAdminSessionResponse, error) {
	return &storage.TenantAdminSessionResponse{}, nil
}

func (s *authHandlerTestStore) GetTenantAdminSession(context.Context, string) (*storage.TenantAdminSessionResponse, error) {
	return nil, nil
}

func (s *authHandlerTestStore) RevokeTenantAdminSession(context.Context, string) error {
	return nil
}

func (s *authHandlerTestStore) FindOrCreateTenantAdminUser(context.Context, storage.FindOrCreateTenantAdminUserParams) (*storage.TenantAdminUserResponse, error) {
	if s.findUserErr != nil {
		return nil, s.findUserErr
	}

	return &storage.TenantAdminUserResponse{ID: "user-1"}, nil
}

func (s *authHandlerTestStore) GetTenantAdminUser(context.Context, string, string) (*storage.TenantAdminUserResponse, error) {
	return nil, nil
}

func (s *authHandlerTestStore) UpdateTenantAdminUserContact(context.Context, storage.UpdateTenantAdminUserContactParams) (*storage.TenantAdminUserResponse, error) {
	return &storage.TenantAdminUserResponse{}, nil
}

func (s *authHandlerTestStore) TouchTenantAdminLastLogin(context.Context, string, string) (*storage.TenantAdminUserResponse, error) {
	return &storage.TenantAdminUserResponse{}, nil
}
