package admin

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/storage"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
)

// AuthStore defines the storage interface for tenant-admin auth operations.
type AuthStore interface {
	CreateTenantAdminCaptcha(ctx context.Context) (string, string, time.Time, error)
	ConsumeTenantAdminCaptcha(ctx context.Context, id string, answer string) (bool, error)
	CreateTenantAdminOTPRequest(ctx context.Context, params storage.TenantAdminOTPRequestParams) (map[string]interface{}, error)
	VerifyTenantAdminOTPRequest(ctx context.Context, params storage.VerifyTenantAdminOTPRequestParams) (map[string]interface{}, error)
	CreateTenantAdminSession(ctx context.Context, params storage.CreateTenantAdminSessionParams) (map[string]interface{}, error)
	GetTenantAdminSession(ctx context.Context, id string) (map[string]interface{}, error)
	RevokeTenantAdminSession(ctx context.Context, id string) error
	FindOrCreateTenantAdminUser(ctx context.Context, params storage.FindOrCreateTenantAdminUserParams) (map[string]interface{}, error)
	GetTenantAdminUser(ctx context.Context, tenantID string, userID string) (map[string]interface{}, error)
	UpdateTenantAdminUserContact(ctx context.Context, params storage.UpdateTenantAdminUserContactParams) (map[string]interface{}, error)
	TouchTenantAdminLastLogin(ctx context.Context, tenantID string, userID string) (map[string]interface{}, error)
}

// AuthHandlers provides HTTP handlers for tenant-admin auth API.
type AuthHandlers struct {
	store       AuthStore
	otpNotifier OTPEmailSender
}

func NewAuthHandlers(store AuthStore, otpNotifier OTPEmailSender) *AuthHandlers {
	return &AuthHandlers{
		store:       store,
		otpNotifier: otpNotifier,
	}
}

// CreateCaptcha handles POST /api/v1/auth/captcha/create
func (h *AuthHandlers) CreateCaptcha(c *fiber.Ctx) error {
	id, answer, expiresAt, err := h.store.CreateTenantAdminCaptcha(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create captcha")
	}

	return c.JSON(fiber.Map{
		"captcha_id":         id,
		"image_data_url":     buildCaptchaDataURL(answer),
		"expires_in_seconds": int(time.Until(expiresAt).Seconds()),
	})
}

func buildCaptchaDataURL(answer string) string {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	width := 170
	height := 56

	var chars strings.Builder
	runes := []rune(answer)
	for i, ch := range runes {
		x := 16 + i*30 + rng.Intn(5)
		y := 34 + rng.Intn(12)
		rot := -24 + rng.Intn(49)
		fontSize := 24 + rng.Intn(8)
		chars.WriteString(fmt.Sprintf(
			`<text x="%d" y="%d" transform="rotate(%d %d %d)" font-size="%d" fill="#0f172a" font-family="monospace" font-weight="700">%s</text>`,
			x,
			y,
			rot,
			x,
			y,
			fontSize,
			string(ch),
		))
	}

	var noise strings.Builder
	for i := 0; i < 10; i++ {
		x1 := rng.Intn(width)
		y1 := rng.Intn(height)
		x2 := rng.Intn(width)
		y2 := rng.Intn(height)
		opacity := 0.20 + rng.Float64()*0.35
		noise.WriteString(fmt.Sprintf(
			`<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="#334155" stroke-opacity="%.2f" stroke-width="1"/>`,
			x1, y1, x2, y2, opacity,
		))
	}
	for i := 0; i < 28; i++ {
		cx := rng.Intn(width)
		cy := rng.Intn(height)
		r := 1 + rng.Intn(2)
		opacity := 0.15 + rng.Float64()*0.35
		noise.WriteString(fmt.Sprintf(
			`<circle cx="%d" cy="%d" r="%d" fill="#475569" fill-opacity="%.2f"/>`,
			cx, cy, r, opacity,
		))
	}

	svg := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">
<rect width="%d" height="%d" rx="10" fill="#cbd5e1" />
%s
%s
</svg>`, width, height, width, height, width, height, noise.String(), chars.String())
	return "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(svg))
}

// ConsumeCaptcha handles POST /api/v1/auth/captcha/consume
func (h *AuthHandlers) ConsumeCaptcha(c *fiber.Ctx) error {
	var body struct {
		CaptchaID string `json:"captcha_id"`
		Answer    string `json:"answer"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	valid, err := h.store.ConsumeTenantAdminCaptcha(c.Context(), body.CaptchaID, body.Answer)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Failed to consume captcha")
	}

	if !valid {
		return c.JSON(fiber.Map{"ok": false, "error": "کپچا اشتباه است یا منقضی شده است."})
	}

	return c.JSON(fiber.Map{"ok": true})
}

// CreateOTPRequest handles POST /api/v1/auth/otp/request
func (h *AuthHandlers) CreateOTPRequest(c *fiber.Ctx) error {
	var body struct {
		TenantSlug   string  `json:"tenant_slug"`
		Purpose      string  `json:"purpose"`
		Channel      string  `json:"channel"`
		Identifier   string  `json:"identifier"`
		UserID       *string `json:"user_id,omitempty"`
		PendingValue *string `json:"pending_value,omitempty"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if body.TenantSlug == "" || body.Identifier == "" {
		return fiber.NewError(fiber.StatusBadRequest, "tenant_slug and identifier are required")
	}

	result, err := h.store.CreateTenantAdminOTPRequest(c.Context(), storage.TenantAdminOTPRequestParams{
		TenantSlug:   body.TenantSlug,
		Purpose:      body.Purpose,
		Channel:      body.Channel,
		Identifier:   body.Identifier,
		UserID:       body.UserID,
		PendingValue: body.PendingValue,
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create OTP request")
	}

	channel, _ := result["channel"].(string)
	destination, _ := result["destination_masked"].(string)
	otpCode, _ := result["otp_code"].(string)
	delete(result, "otp_code")

	if h.otpNotifier != nil && channel == "email" && otpCode != "" {
		if err := h.otpNotifier.SendTenantAdminOTP(body.Identifier, body.TenantSlug, otpCode, 5); err != nil {
			log.Printf("[control-plane][auth][otp-email] failed tenant=%s to=%s: %v", body.TenantSlug, destination, err)
		}
	}

	if strings.ToLower(os.Getenv("NODE_ENV")) != "production" && otpCode != "" {
		result["debug_otp"] = otpCode
	}

	return c.JSON(result)
}

// VerifyOTPRequest handles POST /api/v1/auth/otp/verify
func (h *AuthHandlers) VerifyOTPRequest(c *fiber.Ctx) error {
	var body struct {
		TenantSlug string  `json:"tenant_slug"`
		RequestID  string  `json:"request_id"`
		OTPCode    string  `json:"otp_code"`
		Purpose    string  `json:"purpose"`
		UserID     *string `json:"user_id,omitempty"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	result, err := h.store.VerifyTenantAdminOTPRequest(c.Context(), storage.VerifyTenantAdminOTPRequestParams{
		TenantSlug: body.TenantSlug,
		RequestID:  body.RequestID,
		OTPCode:    body.OTPCode,
		Purpose:    body.Purpose,
		UserID:     body.UserID,
	})
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Failed to verify OTP")
	}

	return c.JSON(result)
}

// CreateSession handles POST /api/v1/auth/session
func (h *AuthHandlers) CreateSession(c *fiber.Ctx) error {
	var body struct {
		TenantSlug string `json:"tenant_slug"`
		UserID     string `json:"user_id"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	result, err := h.store.CreateTenantAdminSession(c.Context(), storage.CreateTenantAdminSessionParams{
		TenantSlug: body.TenantSlug,
		UserID:     body.UserID,
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create session")
	}

	return c.JSON(result)
}

// GetSession handles GET /api/v1/auth/session/:id
func (h *AuthHandlers) GetSession(c *fiber.Ctx) error {
	sessionID := c.Params("id")
	if sessionID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "session ID is required")
	}

	session, err := h.store.GetTenantAdminSession(c.Context(), sessionID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Session not found")
	}
	if session == nil {
		return c.SendStatus(fiber.StatusNotFound)
	}

	return c.JSON(session)
}

// RevokeSession handles DELETE /api/v1/auth/session/:id
func (h *AuthHandlers) RevokeSession(c *fiber.Ctx) error {
	sessionID := c.Params("id")
	if sessionID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "session ID is required")
	}

	if err := h.store.RevokeTenantAdminSession(c.Context(), sessionID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to revoke session")
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// FindOrCreateUser handles POST /api/v1/auth/user/find-or-create
func (h *AuthHandlers) FindOrCreateUser(c *fiber.Ctx) error {
	var body struct {
		TenantSlug string `json:"tenant_slug"`
		Channel    string `json:"channel"`
		Identifier string `json:"identifier"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	user, err := h.store.FindOrCreateTenantAdminUser(c.Context(), storage.FindOrCreateTenantAdminUserParams{
		TenantSlug: body.TenantSlug,
		Channel:    body.Channel,
		Identifier: body.Identifier,
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to find or create user")
	}

	return c.JSON(user)
}

// GetUser handles GET /api/v1/auth/user/:id
func (h *AuthHandlers) GetUser(c *fiber.Ctx) error {
	userID := c.Params("id")
	tenantSlug := c.Query("tenant_slug")

	currentTenant := tenant.FromContext(c)
	if currentTenant != nil && tenantSlug == "" {
		tenantSlug = currentTenant.Slug
	}

	user, err := h.store.GetTenantAdminUser(c.Context(), tenantSlug, userID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "User not found")
	}
	if user == nil {
		return c.SendStatus(fiber.StatusNotFound)
	}

	return c.JSON(user)
}

// UpdateUserContact handles POST /api/v1/auth/user/:id/update-contact
func (h *AuthHandlers) UpdateUserContact(c *fiber.Ctx) error {
	userID := c.Params("id")
	var body struct {
		TenantSlug string `json:"tenant_slug"`
		Field      string `json:"field"`
		Value      string `json:"value"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	user, err := h.store.UpdateTenantAdminUserContact(c.Context(), storage.UpdateTenantAdminUserContactParams{
		TenantSlug: body.TenantSlug,
		UserID:     userID,
		Field:      body.Field,
		Value:      body.Value,
	})
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Failed to update user contact")
	}

	return c.JSON(user)
}

// TouchLastLogin handles POST /api/v1/auth/user/:id/touch-last-login
func (h *AuthHandlers) TouchLastLogin(c *fiber.Ctx) error {
	userID := c.Params("id")
	var body struct {
		TenantSlug string `json:"tenant_slug"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	user, err := h.store.TouchTenantAdminLastLogin(c.Context(), body.TenantSlug, userID)
	if err != nil || user == nil {
		return c.SendStatus(fiber.StatusNotFound)
	}

	return c.JSON(user)
}
