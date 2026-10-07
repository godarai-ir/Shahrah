package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/apiheaders"
	"github.com/Delpak/shahrah/services/control-plane/internal/runtimeprofile"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

// AuthContextKey is the key type for auth context in Fiber Locals.
type AuthContextKey struct{}

type delegatedRequestContextKey struct{}

// MarkDelegatedRequest records that the request was authenticated through the
// master-key delegated-user path.
func MarkDelegatedRequest(c *fiber.Ctx) {
	c.Locals(delegatedRequestContextKey{}, true)
}

// IsDelegatedRequest reports whether the trusted delegated-user path authenticated the request.
func IsDelegatedRequest(c *fiber.Ctx) bool {
	delegated, ok := c.Locals(delegatedRequestContextKey{}).(bool)

	return ok && delegated
}

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
	keyStore        KeyStore
	runtimeProfiles RuntimeProfileLookup
	masterKey       string
	logger          *zap.Logger
}

// RuntimeProfileLookup resolves compiled authentication state.
type RuntimeProfileLookup interface {
	GetAuthByTokenHash(ctx context.Context, tenantID string, tokenHash string) (*platform.AuthContext, error)
}

// NewAuthMiddleware creates a new auth middleware.
func NewAuthMiddleware(keyStore KeyStore, masterKey string, logger *zap.Logger) *AuthMiddleware {
	return NewAuthMiddlewareWithRuntimeProfiles(keyStore, masterKey, logger, nil)
}

// NewAuthMiddlewareWithRuntimeProfiles creates fully wired auth middleware.
func NewAuthMiddlewareWithRuntimeProfiles(
	keyStore KeyStore,
	masterKey string,
	logger *zap.Logger,
	runtimeProfiles RuntimeProfileLookup,
) *AuthMiddleware {
	return &AuthMiddleware{
		keyStore:        keyStore,
		runtimeProfiles: runtimeProfiles,
		masterKey:       masterKey,
		logger:          logger,
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
	// Clone request header value to avoid retaining mutable fasthttp buffer data.
	userID := strings.Clone(apiheaders.GetUserID(c))
	if userID == "" {
		c.Locals(AuthContextKey{}, &platform.AuthContext{Role: roleProxyAdmin})
		return c.Next()
	}

	c.Request().Header.Set(apiheaders.HeaderUserID, userID)

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil || strings.TrimSpace(currentTenant.ID) == "" {
		return writeAuthError(c, fiber.StatusForbidden, "Tenant context is required for delegated authentication", "authorization_error")
	}

	resolved, err := m.keyStore.ResolveTenantUserAuth(c.Context(), currentTenant.ID, userID)
	if err != nil {
		m.logger.Debug("failed to resolve user auth context for master key request", zap.Error(err), zap.String("tenant_id", currentTenant.ID), zap.String("user_id", userID))

		if errors.Is(err, pgx.ErrNoRows) {
			return writeAuthError(c, fiber.StatusForbidden, "Delegated user is not authorized for this tenant", "authorization_error")
		}

		return writeAuthError(c, fiber.StatusInternalServerError, "Delegated authentication lookup failed", "server_error")
	}

	if resolved == nil {
		return writeAuthError(c, fiber.StatusForbidden, "Delegated user is not authorized for this tenant", "authorization_error")
	}

	if !strings.EqualFold(strings.TrimSpace(resolved.UserID), userID) ||
		!strings.EqualFold(strings.TrimSpace(resolved.TenantID), strings.TrimSpace(currentTenant.ID)) {
		m.logger.Warn("delegated authentication lookup returned a mismatched identity", zap.String("tenant_id", currentTenant.ID), zap.String("user_id", userID))
		return writeAuthError(c, fiber.StatusForbidden, "Delegated identity does not match the requested tenant and user", "authorization_error")
	}

	c.Locals(AuthContextKey{}, resolved)
	MarkDelegatedRequest(c)

	if err := CheckRouteAuthorization(c.Method(), c.Path(), resolved); err != nil {
		return writeAuthError(c, fiber.StatusForbidden, err.Error(), "authorization_error")
	}

	return c.Next()
}

// authRejection describes an authentication or authorization failure that must
// terminate the request. Helpers return it instead of writing the response so a
// rejection can never be mistaken for "continue" and fall through to c.Next().
type authRejection struct {
	status  int
	message string
	errType string
}

func rejectAuth(status int, message, errType string) *authRejection {
	return &authRejection{status: status, message: message, errType: errType}
}

func (r *authRejection) write(c *fiber.Ctx) error {
	return writeAuthError(c, r.status, r.message, r.errType)
}

func (m *AuthMiddleware) authenticateAPIKey(c *fiber.Ctx, apiKey string) error {
	keyInfo, rejection := m.lookupAuthContext(c, apiKey)
	if rejection != nil {
		return rejection.write(c)
	}

	if rejection := m.authorizeAndStoreAuthContext(c, keyInfo); rejection != nil {
		return rejection.write(c)
	}

	m.logger.Debug("Authentication successful",
		zap.String("key_alias", keyInfo.KeyAlias),
		zap.String("role", keyInfo.Role),
	)

	m.touchKeyLastUsed(c, keyInfo)

	return c.Next()
}

// lookupAuthContext resolves the auth context for an API key. It returns either a
// non-nil auth context or a non-nil rejection, never both and never neither.
func (m *AuthMiddleware) lookupAuthContext(c *fiber.Ctx, apiKey string) (*platform.AuthContext, *authRejection) {
	hash := HashKey(apiKey)

	keyInfo, handled, rejection := m.lookupRuntimeProfileAuthContext(c, hash)
	if rejection != nil {
		return nil, rejection
	}

	if !handled {
		var err error

		keyInfo, err = m.keyStore.GetKeyByHash(c.Context(), hash)
		if err != nil {
			if isInvalidAPIKeyLookup(err) {
				return nil, rejectAuth(fiber.StatusUnauthorized, "Invalid API key", "authentication_error")
			}

			m.logger.Error("API key lookup failed", zap.Error(err))

			return nil, rejectAuth(fiber.StatusInternalServerError, "Authentication lookup failed", "server_error")
		}
	}

	if keyInfo == nil {
		m.logger.Error("API key lookup returned nil auth context")

		return nil, rejectAuth(fiber.StatusInternalServerError, "Authentication lookup failed", "server_error")
	}

	return keyInfo, nil
}

func (m *AuthMiddleware) lookupRuntimeProfileAuthContext(c *fiber.Ctx, hash string) (*platform.AuthContext, bool, *authRejection) {
	if m.runtimeProfiles == nil {
		return nil, false, nil
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil || strings.TrimSpace(currentTenant.ID) == "" {
		return nil, false, nil
	}

	keyInfo, err := m.runtimeProfiles.GetAuthByTokenHash(c.Context(), currentTenant.ID, hash)
	if err == nil {
		return keyInfo, true, nil
	}

	if errors.Is(err, runtimeprofile.ErrAuthNotFound) {
		return nil, true, rejectAuth(fiber.StatusUnauthorized, "Invalid API key", "authentication_error")
	}

	m.logger.Warn("runtime profile auth lookup unavailable", zap.Error(err), zap.String("tenant_id", currentTenant.ID))

	return nil, true, rejectAuth(fiber.StatusServiceUnavailable, "Tenant runtime profile unavailable", "server_error")
}

func (m *AuthMiddleware) authorizeAndStoreAuthContext(c *fiber.Ctx, keyInfo *platform.AuthContext) *authRejection {
	if keyInfo.Blocked {
		return rejectAuth(fiber.StatusForbidden, "API key has been blocked", "authorization_error")
	}

	if keyInfo.ExpiresAt != nil && time.Now().After(*keyInfo.ExpiresAt) {
		return rejectAuth(fiber.StatusUnauthorized, "API key has expired", "authentication_error")
	}

	c.Locals(AuthContextKey{}, keyInfo)

	if err := CheckRouteAuthorization(c.Method(), c.Path(), keyInfo); err != nil {
		return rejectAuth(fiber.StatusForbidden, err.Error(), "authorization_error")
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
