package admin

import (
	"context"
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/storage"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
)

// AuthStore defines the storage interface for tenant-admin auth operations.
type AuthStore interface {
	CreateTenantAdminCaptcha(ctx context.Context) (string, string, time.Time, error)
	ConsumeTenantAdminCaptcha(ctx context.Context, id string, answer string) (bool, error)
	CreateTenantAdminOTPRequest(ctx context.Context, params storage.TenantAdminOTPRequestParams) (*storage.TenantAdminOTPResponse, error)
	VerifyTenantAdminOTPRequest(ctx context.Context, params storage.VerifyTenantAdminOTPRequestParams) (*storage.TenantAdminOTPVerification, error)
	CreateTenantAdminSession(ctx context.Context, params storage.CreateTenantAdminSessionParams) (*storage.TenantAdminSessionResponse, error)
	GetTenantAdminSession(ctx context.Context, id string) (*storage.TenantAdminSessionResponse, error)
	RevokeTenantAdminSession(ctx context.Context, id string) error
	FindOrCreateTenantAdminUser(ctx context.Context, params storage.FindOrCreateTenantAdminUserParams) (*storage.TenantAdminUserResponse, error)
	GetTenantAdminUser(ctx context.Context, tenantID string, userID string) (*storage.TenantAdminUserResponse, error)
	UpdateTenantAdminUserContact(ctx context.Context, params storage.UpdateTenantAdminUserContactParams) (*storage.TenantAdminUserResponse, error)
	TouchTenantAdminLastLogin(ctx context.Context, tenantID string, userID string) (*storage.TenantAdminUserResponse, error)
}

type socialAuthStore interface {
	ProvisionSocialTenant(ctx context.Context, params storage.ProvisionSocialTenantParams) (*storage.ProvisionedTenant, error)
	CreateSocialLoginHandoff(ctx context.Context, tenantSlug, accountID string) (*storage.SocialHandoff, error)
	RedeemSocialLoginHandoff(ctx context.Context, tenantSlug, code string) (*storage.RedeemedSocialHandoff, error)
	CreateSocialLinkTicket(ctx context.Context, tenantSlug, userID string) (*storage.SocialHandoff, error)
	CompleteSocialLink(ctx context.Context, ticketCode, accountID string) (*storage.SocialHandoff, error)
}

// AuthHandlers provides HTTP handlers for tenant-admin auth API.
type AuthHandlers struct {
	store       AuthStore
	socialStore socialAuthStore
	otpNotifier TenantAdminOTPNotifier
	otpLimiter  *authRateLimiter
}

func NewAuthHandlers(store AuthStore, otpNotifier TenantAdminOTPNotifier) *AuthHandlers {
	socialStore, _ := store.(socialAuthStore) //nolint:errcheck // Optional metadata uses the zero value when absent or mismatched.

	return &AuthHandlers{
		store:       store,
		socialStore: socialStore,
		otpNotifier: otpNotifier,
		otpLimiter:  newAuthRateLimiter(),
	}
}

const (
	tenantAdminOTPIdentifierRateLimit       = 5
	tenantAdminOTPIdentifierRateLimitWindow = 15 * time.Minute
	tenantAdminOTPIPRateLimit               = 30
	tenantAdminOTPIPRateLimitWindow         = time.Minute
	tenantAdminOTPMaxRateLimitBuckets       = 10000
)

type authRateLimiter struct {
	mu      sync.Mutex
	buckets map[string]authRateLimitBucket
	now     func() time.Time
}

type authRateLimitBucket struct {
	count   int
	resetAt time.Time
}

type authRateLimitPolicy struct {
	key    string
	limit  int
	window time.Duration
}

func newAuthRateLimiter() *authRateLimiter {
	return &authRateLimiter{
		buckets: make(map[string]authRateLimitBucket),
		now:     time.Now,
	}
}

func (l *authRateLimiter) allow(policy authRateLimitPolicy) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	l.pruneExpiredLocked(now)

	bucket, exists := l.buckets[policy.key]
	if !exists || !bucket.resetAt.After(now) {
		l.buckets[policy.key] = authRateLimitBucket{
			count:   1,
			resetAt: now.Add(policy.window),
		}

		return true, 0
	}

	if bucket.count >= policy.limit {
		return false, bucket.resetAt.Sub(now)
	}

	bucket.count++
	l.buckets[policy.key] = bucket

	return true, 0
}

func (l *authRateLimiter) pruneExpiredLocked(now time.Time) {
	if len(l.buckets) <= tenantAdminOTPMaxRateLimitBuckets {
		return
	}

	for key, bucket := range l.buckets {
		if !bucket.resetAt.After(now) {
			delete(l.buckets, key)
		}
	}

	for key := range l.buckets {
		if len(l.buckets) <= tenantAdminOTPMaxRateLimitBuckets {
			break
		}

		delete(l.buckets, key)
	}
}

func (h *AuthHandlers) enforceOTPRequestRateLimit(c *fiber.Ctx, tenantSlug, channel, identifier string) error {
	policies := []authRateLimitPolicy{
		{
			key:    authRateLimitKey("identifier", tenantSlug, channel, strings.TrimSpace(strings.ToLower(identifier))),
			limit:  tenantAdminOTPIdentifierRateLimit,
			window: tenantAdminOTPIdentifierRateLimitWindow,
		},
		{
			key:    authRateLimitKey("ip", tenantSlug, authRequestClientIP(c)),
			limit:  tenantAdminOTPIPRateLimit,
			window: tenantAdminOTPIPRateLimitWindow,
		},
	}

	var retryAfter time.Duration

	for _, policy := range policies {
		ok, wait := h.otpLimiter.allow(policy)
		if !ok && wait > retryAfter {
			retryAfter = wait
		}
	}

	if retryAfter <= 0 {
		return nil
	}

	retryAfterSeconds := int((retryAfter + time.Second - time.Nanosecond) / time.Second)
	if retryAfterSeconds < 1 {
		retryAfterSeconds = 1
	}

	c.Set("Retry-After", strconv.Itoa(retryAfterSeconds))

	return fiber.NewError(fiber.StatusTooManyRequests, "Too many OTP requests. Please try again later.")
}

func authRateLimitKey(parts ...string) string {
	sum := sha256.Sum256([]byte(strings.Join(parts, "\x00")))
	return fmt.Sprintf("%x", sum[:])
}

func authRequestClientIP(c *fiber.Ctx) string {
	if value := strings.TrimSpace(strings.Split(c.Get("X-Forwarded-For"), ",")[0]); value != "" {
		return value
	}

	if value := strings.TrimSpace(c.Get("X-Real-IP")); value != "" {
		return value
	}

	if value := strings.TrimSpace(c.IP()); value != "" {
		return value
	}

	return "unknown"
}

var tenantSlugPattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`)

func normalizeTenantSlug(value string) string {
	return strings.TrimSpace(strings.ToLower(value))
}

func tenantSlugFromTrustedRequestContext(c *fiber.Ctx) string {
	if currentTenant := tenant.FromContext(c); currentTenant != nil {
		return normalizeTenantSlug(currentTenant.Slug)
	}

	return normalizeTenantSlug(c.Get("x-godarai-tenant"))
}

func resolveTenantSlugForRequest(c *fiber.Ctx, requested string) (string, error) {
	effective := tenantSlugFromTrustedRequestContext(c)
	requested = normalizeTenantSlug(requested)

	if effective == "" {
		effective = requested
	}

	if effective == "" {
		return "", fiber.NewError(fiber.StatusBadRequest, "tenant_slug is required")
	}

	if !tenantSlugPattern.MatchString(effective) {
		return "", fiber.NewError(fiber.StatusBadRequest, "tenant_slug is invalid")
	}

	if requested != "" && requested != effective {
		return "", fiber.NewError(fiber.StatusForbidden, "tenant_slug does not match request tenant")
	}

	return effective, nil
}

// CreateCaptcha handles POST /api/v1/auth/captcha/create.
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
		fmt.Fprintf(&chars, `<text x="%d" y="%d" transform="rotate(%d %d %d)" font-size="%d" fill="#0f172a" font-family="monospace" font-weight="700">%s</text>`,
			x,
			y,
			rot,
			x,
			y,
			fontSize,
			string(ch))
	}

	var noise strings.Builder

	for i := 0; i < 10; i++ {
		x1 := rng.Intn(width)
		y1 := rng.Intn(height)
		x2 := rng.Intn(width)
		y2 := rng.Intn(height)
		opacity := 0.20 + rng.Float64()*0.35
		fmt.Fprintf(&noise, `<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="#334155" stroke-opacity="%.2f" stroke-width="1"/>`,
			x1, y1, x2, y2, opacity)
	}

	for i := 0; i < 28; i++ {
		cx := rng.Intn(width)
		cy := rng.Intn(height)
		r := 1 + rng.Intn(2)
		opacity := 0.15 + rng.Float64()*0.35
		fmt.Fprintf(&noise, `<circle cx="%d" cy="%d" r="%d" fill="#475569" fill-opacity="%.2f"/>`,
			cx, cy, r, opacity)
	}

	svg := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">
<rect width="%d" height="%d" rx="10" fill="#cbd5e1" />
%s
%s
</svg>`, width, height, width, height, width, height, noise.String(), chars.String())

	return "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(svg))
}

// ConsumeCaptcha handles POST /api/v1/auth/captcha/consume.
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

// CreateOTPRequest handles POST /api/v1/auth/otp/request.
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

	tenantSlug, err := resolveTenantSlugForRequest(c, body.TenantSlug)
	if err != nil {
		return err
	}

	channel := strings.TrimSpace(strings.ToLower(body.Channel))
	if channel == "" {
		channel = "email"
	}

	if rateLimitErr := h.enforceOTPRequestRateLimit(c, tenantSlug, channel, body.Identifier); rateLimitErr != nil {
		return rateLimitErr
	}

	if strings.TrimSpace(strings.ToLower(body.Purpose)) == "login" {
		if _, userErr := h.store.FindOrCreateTenantAdminUser(c.Context(), storage.FindOrCreateTenantAdminUserParams{
			TenantSlug: tenantSlug,
			Channel:    channel,
			Identifier: body.Identifier,
		}); userErr != nil {
			if errors.Is(userErr, storage.ErrTenantAdminUserNotFound) {
				return c.JSON(buildGenericTenantAdminOTPResponse(channel, body.Identifier))
			}

			return fiber.NewError(fiber.StatusInternalServerError, "Failed to find tenant user")
		}
	}

	result, err := h.store.CreateTenantAdminOTPRequest(c.Context(), storage.TenantAdminOTPRequestParams{
		TenantSlug:   tenantSlug,
		Purpose:      body.Purpose,
		Channel:      channel,
		Identifier:   body.Identifier,
		UserID:       body.UserID,
		PendingValue: body.PendingValue,
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create OTP request")
	}

	destination := result.DestinationMasked
	otpCode := result.OTPCode

	if h.otpNotifier != nil && otpCode != "" {
		if err := h.otpNotifier.SendTenantAdminOTP(channel, body.Identifier, tenantSlug, otpCode, 5); err != nil {
			log.Printf("[control-plane][auth][otp] failed tenant=%s channel=%s to=%s: %v", tenantSlug, channel, destination, err)
		}
	}

	if debugOTPExposureEnabled() && otpCode != "" {
		result.DebugOTP = otpCode
	}

	return c.JSON(result)
}

func debugOTPExposureEnabled() bool {
	nodeEnv := strings.ToLower(strings.TrimSpace(os.Getenv("NODE_ENV")))
	if nodeEnv != "development" && nodeEnv != "test" {
		return false
	}

	return strings.EqualFold(strings.TrimSpace(os.Getenv("SHAHRAH_EXPOSE_DEBUG_OTP")), "true")
}

func buildGenericTenantAdminOTPResponse(channel, identifier string) fiber.Map {
	return fiber.Map{
		"request_id":         generateTenantAdminAuthID(),
		"channel":            channel,
		"destination_masked": maskTenantAdminAuthIdentifier(channel, identifier),
		"expires_in_seconds": int(5 * 60),
	}
}

func generateTenantAdminAuthID() string {
	b := make([]byte, 16)
	_, _ = crand.Read(b)

	return hex.EncodeToString(b)
}

func maskTenantAdminAuthIdentifier(channel, identifier string) string {
	if channel == "email" {
		parts := strings.SplitN(strings.ToLower(identifier), "@", 2)
		local := parts[0]

		domain := ""
		if len(parts) > 1 {
			domain = parts[1]
		}

		if local == "" {
			return "***@" + domain
		}

		if len(local) <= 2 {
			return string(local[0]) + "***@" + domain
		}

		return local[:2] + "***@" + domain
	}

	compact := strings.ReplaceAll(identifier, " ", "")
	if len(compact) <= 4 {
		return "***"
	}

	return compact[:3] + "****" + compact[len(compact)-3:]
}

// VerifyOTPRequest handles POST /api/v1/auth/otp/verify.
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

	tenantSlug, err := resolveTenantSlugForRequest(c, body.TenantSlug)
	if err != nil {
		return err
	}

	result, err := h.store.VerifyTenantAdminOTPRequest(c.Context(), storage.VerifyTenantAdminOTPRequestParams{
		TenantSlug: tenantSlug,
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

// CreateSession handles POST /api/v1/auth/session.
func (h *AuthHandlers) CreateSession(c *fiber.Ctx) error {
	var body struct {
		TenantSlug string `json:"tenant_slug"`
		UserID     string `json:"user_id"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	tenantSlug, err := resolveTenantSlugForRequest(c, body.TenantSlug)
	if err != nil {
		return err
	}

	result, err := h.store.CreateTenantAdminSession(c.Context(), storage.CreateTenantAdminSessionParams{
		TenantSlug: tenantSlug,
		UserID:     body.UserID,
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create session")
	}

	return c.JSON(result)
}

// GetSession handles GET /api/v1/auth/session/:id.
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

// RevokeSession handles DELETE /api/v1/auth/session/:id.
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

// FindOrCreateUser handles POST /api/v1/auth/user/find-or-create.
func (h *AuthHandlers) FindOrCreateUser(c *fiber.Ctx) error {
	var body struct {
		TenantSlug string `json:"tenant_slug"`
		Channel    string `json:"channel"`
		Identifier string `json:"identifier"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	tenantSlug, err := resolveTenantSlugForRequest(c, body.TenantSlug)
	if err != nil {
		return err
	}

	user, err := h.store.FindOrCreateTenantAdminUser(c.Context(), storage.FindOrCreateTenantAdminUserParams{
		TenantSlug: tenantSlug,
		Channel:    body.Channel,
		Identifier: body.Identifier,
	})
	if err != nil {
		if errors.Is(err, storage.ErrTenantAdminUserNotFound) {
			return fiber.NewError(fiber.StatusForbidden, "You do not have access to this tenant.")
		}

		return fiber.NewError(fiber.StatusInternalServerError, "Failed to find or create user")
	}

	return c.JSON(user)
}

// GetUser handles GET /api/v1/auth/user/:id.
func (h *AuthHandlers) GetUser(c *fiber.Ctx) error {
	userID := c.Params("id")

	tenantSlug, err := resolveTenantSlugForRequest(c, c.Query("tenant_slug"))
	if err != nil {
		return err
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

// UpdateUserContact handles POST /api/v1/auth/user/:id/update-contact.
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

	tenantSlug, err := resolveTenantSlugForRequest(c, body.TenantSlug)
	if err != nil {
		return err
	}

	user, err := h.store.UpdateTenantAdminUserContact(c.Context(), storage.UpdateTenantAdminUserContactParams{
		TenantSlug: tenantSlug,
		UserID:     userID,
		Field:      body.Field,
		Value:      body.Value,
	})
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Failed to update user contact")
	}

	return c.JSON(user)
}

// TouchLastLogin handles POST /api/v1/auth/user/:id/touch-last-login.
func (h *AuthHandlers) TouchLastLogin(c *fiber.Ctx) error {
	userID := c.Params("id")

	var body struct {
		TenantSlug string `json:"tenant_slug"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	tenantSlug, err := resolveTenantSlugForRequest(c, body.TenantSlug)
	if err != nil {
		return err
	}

	user, err := h.store.TouchTenantAdminLastLogin(c.Context(), tenantSlug, userID)
	if err != nil || user == nil {
		return c.SendStatus(fiber.StatusNotFound)
	}

	return c.JSON(user)
}
