package admin

import (
	"bytes"
	"context"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/storage"
	"github.com/gofiber/fiber/v2"
)

func TestCreateOTPRequestRejectsUnknownLoginAccountBeforeCreatingOTP(t *testing.T) {
	store := &authHandlerTestStore{findUserErr: errors.New("tenant member not found")}
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

	if resp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusForbidden)
	}
	if store.createOTPCalled {
		t.Fatal("CreateTenantAdminOTPRequest was called for an unknown login account")
	}
}

type authHandlerTestStore struct {
	findUserErr     error
	createOTPCalled bool
}

func (s *authHandlerTestStore) CreateTenantAdminCaptcha(context.Context) (string, string, time.Time, error) {
	return "", "", time.Time{}, nil
}

func (s *authHandlerTestStore) ConsumeTenantAdminCaptcha(context.Context, string, string) (bool, error) {
	return true, nil
}

func (s *authHandlerTestStore) CreateTenantAdminOTPRequest(context.Context, storage.TenantAdminOTPRequestParams) (map[string]interface{}, error) {
	s.createOTPCalled = true
	return map[string]interface{}{}, nil
}

func (s *authHandlerTestStore) VerifyTenantAdminOTPRequest(context.Context, storage.VerifyTenantAdminOTPRequestParams) (map[string]interface{}, error) {
	return map[string]interface{}{}, nil
}

func (s *authHandlerTestStore) CreateTenantAdminSession(context.Context, storage.CreateTenantAdminSessionParams) (map[string]interface{}, error) {
	return map[string]interface{}{}, nil
}

func (s *authHandlerTestStore) GetTenantAdminSession(context.Context, string) (map[string]interface{}, error) {
	return nil, nil
}

func (s *authHandlerTestStore) RevokeTenantAdminSession(context.Context, string) error {
	return nil
}

func (s *authHandlerTestStore) FindOrCreateTenantAdminUser(context.Context, storage.FindOrCreateTenantAdminUserParams) (map[string]interface{}, error) {
	if s.findUserErr != nil {
		return nil, s.findUserErr
	}
	return map[string]interface{}{"id": "user-1"}, nil
}

func (s *authHandlerTestStore) GetTenantAdminUser(context.Context, string, string) (map[string]interface{}, error) {
	return nil, nil
}

func (s *authHandlerTestStore) UpdateTenantAdminUserContact(context.Context, storage.UpdateTenantAdminUserContactParams) (map[string]interface{}, error) {
	return map[string]interface{}{}, nil
}

func (s *authHandlerTestStore) TouchTenantAdminLastLogin(context.Context, string, string) (map[string]interface{}, error) {
	return map[string]interface{}{}, nil
}
