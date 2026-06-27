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

	"github.com/gofiber/fiber/v2"
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
	assert.NoError(t, err)
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
	assert.NoError(t, err)
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
	assert.NoError(t, err)
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

type failingKeyStore struct {
	err error
}

func (s *failingKeyStore) GetKeyByHash(ctx context.Context, hash string) (*platform.AuthContext, error) {
	return nil, s.err
}

func (s *failingKeyStore) ResolveTenantUserAuth(ctx context.Context, tenantID string, userID string) (*platform.AuthContext, error) {
	return nil, s.err
}
