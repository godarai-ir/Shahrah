package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

// AuthContextKey is the key type for auth context in Fiber Locals.
type AuthContextKey struct{}

// ErrInvalidAPIKey is returned when the API key is invalid.
var ErrInvalidAPIKey = errors.New("invalid API key")

// KeyStore defines the interface for API key lookup.
type KeyStore interface {
	GetKeyByHash(ctx context.Context, hash string) (*platform.AuthContext, error)
	ResolveTenantUserAuth(ctx context.Context, tenantID string, userID string) (*platform.AuthContext, error)
}

type keyLastUsedTracker interface {
	TouchKeyLastUsed(ctx context.Context, keyID string, usedAt time.Time) error
}

// KeyInfo stores information about an API key.
type KeyInfo struct {
	KeyAlias  string
	TeamID    string
	UserID    string
	Role      string
	Models    []string
	Spend     float64
	ExpiresAt *time.Time
	Blocked   bool
}

// AuthMiddleware validates API keys on incoming requests.
type AuthMiddleware struct {
	keyStore  KeyStore
	masterKey string
	logger    *zap.Logger
}

// NewAuthMiddleware creates a new auth middleware.
func NewAuthMiddleware(keyStore KeyStore, masterKey string, logger *zap.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		keyStore:  keyStore,
		masterKey: masterKey,
		logger:    logger,
	}
}

// Middleware returns the Fiber middleware handler.
func (m *AuthMiddleware) Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		path := c.Path()
		if shouldSkipAuth(path) {
			return c.Next()
		}

		apiKey := m.extractAPIKey(c)
		if apiKey == "" {
			return writeAuthError(c, fiber.StatusUnauthorized, "Missing API key. Pass it using the `Authorization` header", "authentication_error")
		}

		if m.masterKey != "" && apiKey == m.masterKey {
			return m.handleMasterKeyRequest(c)
		}

		return m.authenticateAPIKey(c, apiKey)
	}
}

func shouldSkipAuth(path string) bool {
	return path == "/health" ||
		path == "/ready" ||
		path == "/metrics" ||
		strings.HasPrefix(path, "/api/v1/onboarding/") ||
		strings.HasPrefix(path, "/api/v1/wallet/topup/callback/")
}

func writeAuthError(c *fiber.Ctx, status int, message, errType string) error {
	return c.Status(status).JSON(fiber.Map{
		"error": map[string]interface{}{
			"message": message,
			"type":    errType,
		},
	})
}

func (m *AuthMiddleware) handleMasterKeyRequest(c *fiber.Ctx) error {
	resolved := &platform.AuthContext{Role: roleProxyAdmin}
	// Clone request header value to avoid retaining mutable fasthttp buffer data.
	userID := strings.Clone(strings.TrimSpace(c.Get("x-shahrah-user-id")))
	if userID != "" {
		currentTenant := tenant.FromContext(c)
		if currentTenant != nil {
			userCtx, err := m.keyStore.ResolveTenantUserAuth(c.Context(), currentTenant.ID, userID)
			if err == nil && userCtx != nil {
				resolved = userCtx
			} else if err != nil {
				m.logger.Debug("failed to resolve user auth context for master key request", zap.Error(err), zap.String("tenant_id", currentTenant.ID), zap.String("user_id", userID))
			}
		}
	}

	c.Locals(AuthContextKey{}, resolved)

	return c.Next()
}

func (m *AuthMiddleware) authenticateAPIKey(c *fiber.Ctx, apiKey string) error {
	keyInfo, err := m.lookupAuthContext(c, apiKey)
	if err != nil {
		return err
	}
	if keyInfo == nil {
		// lookupAuthContext already wrote the response for handled auth failures.
		return nil
	}

	if err := m.authorizeAndStoreAuthContext(c, keyInfo); err != nil {
		return err
	}

	m.logger.Debug("Authentication successful",
		zap.String("key_alias", keyInfo.KeyAlias),
		zap.String("role", keyInfo.Role),
	)

	m.touchKeyLastUsed(c, keyInfo)

	return c.Next()
}

func (m *AuthMiddleware) lookupAuthContext(c *fiber.Ctx, apiKey string) (*platform.AuthContext, error) {
	hash := HashKey(apiKey)

	keyInfo, err := m.keyStore.GetKeyByHash(c.Context(), hash)
	if err != nil {
		if isInvalidAPIKeyLookup(err) {
			return nil, writeAuthError(c, fiber.StatusUnauthorized, "Invalid API key", "authentication_error")
		}

		m.logger.Error("API key lookup failed", zap.Error(err))

		return nil, writeAuthError(c, fiber.StatusInternalServerError, "Authentication lookup failed", "server_error")
	}

	if keyInfo == nil {
		m.logger.Error("API key lookup returned nil auth context")

		return nil, writeAuthError(c, fiber.StatusInternalServerError, "Authentication lookup failed", "server_error")
	}

	return keyInfo, nil
}

func (m *AuthMiddleware) authorizeAndStoreAuthContext(c *fiber.Ctx, keyInfo *platform.AuthContext) error {
	if keyInfo.Blocked {
		return writeAuthError(c, fiber.StatusForbidden, "API key has been blocked", "authorization_error")
	}

	if keyInfo.ExpiresAt != nil && time.Now().After(*keyInfo.ExpiresAt) {
		return writeAuthError(c, fiber.StatusUnauthorized, "API key has expired", "authentication_error")
	}

	c.Locals(AuthContextKey{}, keyInfo)

	if err := CheckRouteAuthorization(c.Method(), c.Path(), keyInfo); err != nil {
		return writeAuthError(c, fiber.StatusForbidden, err.Error(), "authorization_error")
	}

	return nil
}

func (m *AuthMiddleware) touchKeyLastUsed(c *fiber.Ctx, keyInfo *platform.AuthContext) {
	if keyInfo.KeyID == "" {
		return
	}

	tracker, ok := m.keyStore.(keyLastUsedTracker)
	if !ok {
		return
	}

	if err := tracker.TouchKeyLastUsed(c.Context(), keyInfo.KeyID, time.Now().UTC()); err != nil {
		m.logger.Debug("Failed to touch key last-used timestamp", zap.Error(err), zap.String("key_id", keyInfo.KeyID))
	}
}

func isInvalidAPIKeyLookup(err error) bool {
	return errors.Is(err, ErrInvalidAPIKey) || errors.Is(err, pgx.ErrNoRows)
}

// GetAuthInfo retrieves auth info from Fiber context.
func GetAuthInfo(c *fiber.Ctx) *platform.AuthContext {
	val := c.Locals(AuthContextKey{})
	if val == nil {
		return nil
	}

	authInfo, ok := val.(*platform.AuthContext)
	if !ok {
		return nil
	}

	return authInfo
}

func (m *AuthMiddleware) extractAPIKey(c *fiber.Ctx) string {
	// Check Authorization header (Bearer token)
	auth := c.Get("Authorization")
	if auth != "" {
		if strings.HasPrefix(auth, "Bearer ") {
			return strings.TrimPrefix(auth, "Bearer ")
		}

		return auth
	}

	// Check api-key header (OpenAI style)
	if key := c.Get("api-key"); key != "" {
		return key
	}

	// Check x-api-key header
	if key := c.Get("x-api-key"); key != "" {
		return key
	}

	return ""
}

// HashKey hashes an API key for secure storage/comparison.
func HashKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}
