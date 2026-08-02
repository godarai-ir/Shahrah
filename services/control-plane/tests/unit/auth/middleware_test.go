package auth_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestHashKey_ShouldProduceConsistentHash verifies that HashKey produces consistent hashes.
func TestHashKey_ShouldProduceConsistentHash(t *testing.T) {
	hash1 := auth.HashKey("test-api-key")
	hash2 := auth.HashKey("test-api-key")

	assert.Equal(t, hash1, hash2)
	assert.NotEmpty(t, hash1)
	assert.Len(t, hash1, 64) // SHA-256 produces 64 hex characters
}

// TestHashKey_ShouldDifferForDifferentKeys verifies that different keys produce different hashes.
func TestHashKey_ShouldDifferForDifferentKeys(t *testing.T) {
	hash1 := auth.HashKey("key-one")
	hash2 := auth.HashKey("key-two")

	assert.NotEqual(t, hash1, hash2)
}

// TestInMemoryKeyStore_GetKeyByHash_ShouldReturnKey verifies that valid keys are retrieved correctly.
func TestInMemoryKeyStore_GetKeyByHash_ShouldReturnKey(t *testing.T) {
	store := auth.NewInMemoryKeyStore()
	hash := auth.HashKey("test-key")

	store.AddKey(hash, &platform.AuthContext{
		KeyAlias: "test-key",
		Role:     "team",
	})

	key, err := store.GetKeyByHash(context.Background(), hash)
	require.NoError(t, err)
	assert.Equal(t, "test-key", key.KeyAlias)
	assert.Equal(t, "team", key.Role)
}

// TestInMemoryKeyStore_GetKeyByHash_ShouldReturnError_WhenKeyNotFound verifies that missing keys return an error.
func TestInMemoryKeyStore_GetKeyByHash_ShouldReturnError_WhenKeyNotFound(t *testing.T) {
	store := auth.NewInMemoryKeyStore()

	key, err := store.GetKeyByHash(context.Background(), "non-existent-hash")
	assert.Error(t, err)
	assert.Nil(t, key)
	assert.Equal(t, auth.ErrInvalidAPIKey, err)
}

// TestInMemoryKeyStore_BlockedKey_ShouldBeRetrieved verifies that blocked keys are correctly stored.
func TestInMemoryKeyStore_BlockedKey_ShouldBeRetrieved(t *testing.T) {
	store := auth.NewInMemoryKeyStore()
	hash := auth.HashKey("blocked-key")
	blocked := true

	store.AddKey(hash, &platform.AuthContext{
		KeyAlias: "blocked-key",
		Blocked:  blocked,
	})

	key, err := store.GetKeyByHash(context.Background(), hash)
	require.NoError(t, err)
	assert.True(t, key.Blocked)
}

// TestInMemoryKeyStore_ExpiredKey_ShouldBeRetrieved verifies that expired keys are correctly stored.
func TestInMemoryKeyStore_ExpiredKey_ShouldBeRetrieved(t *testing.T) {
	store := auth.NewInMemoryKeyStore()
	hash := auth.HashKey("expired-key")
	expiredTime := time.Now().Add(-24 * time.Hour)

	store.AddKey(hash, &platform.AuthContext{
		KeyAlias:  "expired-key",
		ExpiresAt: &expiredTime,
	})

	key, err := store.GetKeyByHash(context.Background(), hash)
	require.NoError(t, err)
	assert.True(t, time.Now().After(*key.ExpiresAt))
}

func TestAuthMiddleware_ShouldReturnServerError_WhenKeyLookupFailsInternally(t *testing.T) {
	app := fiber.New()
	app.Use(auth.NewAuthMiddleware(&failingKeyStore{err: errors.New("database lookup failed")}, "", zap.NewNop()).Middleware())
	app.Get("/v1/models", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req, err := http.NewRequest(http.MethodGet, "/v1/models", http.NoBody)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer sk-test")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)

	defer resp.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var payload struct {
		Error struct {
			Message string `json:"message"`
			Type    string `json:"type"`
		} `json:"error"`
	}
	require.NoError(t, json.Unmarshal(body, &payload))
	assert.Equal(t, "Authentication lookup failed", payload.Error.Message)
	assert.Equal(t, "server_error", payload.Error.Type)
}

func TestAuthMiddleware_ShouldReturnInvalidAPIKey_WhenKeyIsNotFound(t *testing.T) {
	app := fiber.New()
	app.Use(auth.NewAuthMiddleware(&failingKeyStore{err: auth.ErrInvalidAPIKey}, "", zap.NewNop()).Middleware())
	app.Get("/v1/models", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req, err := http.NewRequest(http.MethodGet, "/v1/models", http.NoBody)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer sk-test")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)

	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var payload struct {
		Error struct {
			Message string `json:"message"`
			Type    string `json:"type"`
		} `json:"error"`
	}
	require.NoError(t, json.Unmarshal(body, &payload))
	assert.Equal(t, "Invalid API key", payload.Error.Message)
	assert.Equal(t, "authentication_error", payload.Error.Type)
}

func TestAuthMiddleware_MasterKeyDelegationFailsClosedWhenUserLookupFails(t *testing.T) {
	app := newMasterKeyDelegationTestApp(&delegatedKeyStore{resolveErr: errors.New("database lookup failed")})

	resp := performMasterKeyDelegationRequest(t, app, "/api/v1/users", "user-1")
	defer closeResponseBody(t, resp)

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestAuthMiddleware_MasterKeyDelegationFailsClosedWhenUserIsNotFound(t *testing.T) {
	app := newMasterKeyDelegationTestApp(&delegatedKeyStore{resolveErr: pgx.ErrNoRows})

	resp := performMasterKeyDelegationRequest(t, app, "/api/v1/users", "missing-user")
	defer closeResponseBody(t, resp)

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestAuthMiddleware_MasterKeyDelegationAppliesResolvedUserAuthorization(t *testing.T) {
	app := newMasterKeyDelegationTestApp(&delegatedKeyStore{resolved: &platform.AuthContext{
		TenantID: "tenant-1",
		UserID:   "user-1",
		Role:     "member",
	}})

	resp := performMasterKeyDelegationRequest(t, app, "/api/v1/users", "user-1")
	defer closeResponseBody(t, resp)

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestAuthMiddleware_MasterKeyDelegationAllowsResolvedUserOnAuthorizedRoute(t *testing.T) {
	app := newMasterKeyDelegationTestApp(&delegatedKeyStore{resolved: &platform.AuthContext{
		TenantID: "tenant-1",
		UserID:   "user-1",
		Role:     "member",
	}})

	resp := performMasterKeyDelegationRequest(t, app, "/api/v1/me/permissions", "user-1")
	defer closeResponseBody(t, resp)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAuthMiddleware_MasterKeyDelegationAcceptsLegacyUserHeader(t *testing.T) {
	app := newMasterKeyDelegationTestApp(&delegatedKeyStore{resolved: &platform.AuthContext{
		TenantID: "tenant-1",
		UserID:   "user-1",
		Role:     "member",
	}})

	req, err := http.NewRequest(http.MethodGet, "/api/v1/me/permissions", http.NoBody)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer test-master-key")
	req.Header.Set("x-shahrah-user-id", "user-1")

	resp, err := app.Test(req)

	require.NoError(t, err)
	defer closeResponseBody(t, resp)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAuthMiddleware_MasterKeyDelegationRequiresTenantContext(t *testing.T) {
	app := fiber.New()
	app.Use(auth.NewAuthMiddleware(&delegatedKeyStore{resolved: &platform.AuthContext{
		TenantID: "tenant-1",
		UserID:   "user-1",
		Role:     "member",
	}}, "test-master-key", zap.NewNop()).Middleware())
	app.Get("/api/v1/me/permissions", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	resp := performMasterKeyDelegationRequest(t, app, "/api/v1/me/permissions", "user-1")
	defer closeResponseBody(t, resp)

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestAuthMiddleware_MasterKeyDelegationRejectsMismatchedResolvedIdentity(t *testing.T) {
	app := newMasterKeyDelegationTestApp(&delegatedKeyStore{resolved: &platform.AuthContext{
		TenantID: "tenant-2",
		UserID:   "user-2",
		Role:     "tenant_admin",
	}})

	resp := performMasterKeyDelegationRequest(t, app, "/api/v1/users", "user-1")
	defer closeResponseBody(t, resp)

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestAuthMiddleware_MasterKeyWithoutDelegatedUserPreservesBootstrapAdmin(t *testing.T) {
	app := newMasterKeyDelegationTestApp(&delegatedKeyStore{})

	resp := performMasterKeyDelegationRequest(t, app, "/api/v1/users", "")
	defer closeResponseBody(t, resp)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

type delegatedKeyStore struct {
	resolved   *platform.AuthContext
	resolveErr error
}

func (s *delegatedKeyStore) GetKeyByHash(context.Context, string) (*platform.AuthContext, error) {
	return nil, auth.ErrInvalidAPIKey
}

func (s *delegatedKeyStore) ResolveTenantUserAuth(context.Context, string, string) (*platform.AuthContext, error) {
	return s.resolved, s.resolveErr
}

func newMasterKeyDelegationTestApp(store auth.KeyStore) *fiber.App {
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals(tenant.ContextKey{}, &platform.Tenant{ID: "tenant-1", Slug: "acme"})
		return c.Next()
	})
	app.Use(auth.NewAuthMiddleware(store, "test-master-key", zap.NewNop()).Middleware())
	app.Get("/api/v1/users", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})
	app.Get("/api/v1/me/permissions", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	return app
}

func performMasterKeyDelegationRequest(t *testing.T, app *fiber.App, path string, userID string) *http.Response {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, path, http.NoBody)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer test-master-key")

	if userID != "" {
		req.Header.Set("x-godarai-user-id", userID)
	}

	resp, err := app.Test(req)
	require.NoError(t, err)

	return resp
}

type failingKeyStore struct {
	err error
}

func (s *failingKeyStore) GetKeyByHash(ctx context.Context, hash string) (*platform.AuthContext, error) {
	return nil, s.err
}

func (s *failingKeyStore) ResolveTenantUserAuth(ctx context.Context, tenantID string, userID string) (*platform.AuthContext, error) {
	return nil, s.err
}
