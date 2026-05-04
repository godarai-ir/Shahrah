package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// AuthContextKey is the key type for auth context in Fiber Locals.
type AuthContextKey struct{}

// ErrInvalidAPIKey is returned when the API key is invalid.
var ErrInvalidAPIKey = errors.New("invalid API key")

// KeyStore defines the interface for API key lookup.
type KeyStore interface {
	GetKeyByHash(ctx context.Context, hash string) (*platform.AuthContext, error)
}

// KeyInfo stores information about an API key.
type KeyInfo struct {
	KeyAlias  string
	TeamID    string
	UserID    string
	Role      string
	Models    []string
	Spend     float64
	MaxBudget *float64
	RPMLimit  *int64
	TPMLimit  *int64
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
		// Skip auth for public health, onboarding, and payment gateway callbacks.
		if c.Path() == "/health" ||
			c.Path() == "/ready" ||
			strings.HasPrefix(c.Path(), "/api/v1/onboarding/") ||
			strings.HasPrefix(c.Path(), "/api/v1/wallet/topup/callback/") {
			return c.Next()
		}

		// Extract API key from headers
		apiKey := m.extractAPIKey(c)
		if apiKey == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": map[string]interface{}{
					"message": "Missing API key. Pass it using the `Authorization` header",
					"type":    "authentication_error",
				},
			})
		}

		// Check master key
		if m.masterKey != "" && apiKey == m.masterKey {
			c.Locals(AuthContextKey{}, &platform.AuthContext{
				Role: roleProxyAdmin,
			})
			return c.Next()
		}

		// Hash the key for lookup
		hash := HashKey(apiKey)

		// Look up key in store
		keyInfo, err := m.keyStore.GetKeyByHash(c.Context(), hash)
		if err != nil {
			m.logger.Debug("Key lookup failed", zap.Error(err))
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": map[string]interface{}{
					"message": "Invalid API key",
					"type":    "authentication_error",
				},
			})
		}

		// Check if key is blocked
		if keyInfo.Blocked {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": map[string]interface{}{
					"message": "API key has been blocked",
					"type":    "authorization_error",
				},
			})
		}

		// Check expiration
		if keyInfo.ExpiresAt != nil && time.Now().After(*keyInfo.ExpiresAt) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": map[string]interface{}{
					"message": "API key has expired",
					"type":    "authentication_error",
				},
			})
		}

		// Store auth info in context
		c.Locals(AuthContextKey{}, keyInfo)
		if err := CheckRouteAuthorization(c.Method(), c.Path(), keyInfo); err != nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": map[string]interface{}{
					"message": err.Error(),
					"type":    "authorization_error",
				},
			})
		}
		m.logger.Debug("Authentication successful",
			zap.String("key_alias", keyInfo.KeyAlias),
			zap.String("role", keyInfo.Role),
		)

		return c.Next()
	}
}

// GetAuthInfo retrieves auth info from Fiber context.
func GetAuthInfo(c *fiber.Ctx) *platform.AuthContext {
	val := c.Locals(AuthContextKey{})
	if val == nil {
		return nil
	}
	return val.(*platform.AuthContext)
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
