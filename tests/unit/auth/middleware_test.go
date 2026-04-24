package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/Delpak/shahrah/internal/auth"

	"github.com/stretchr/testify/assert"
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

	store.AddKey(hash, &auth.KeyInfo{
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

	store.AddKey(hash, &auth.KeyInfo{
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

	store.AddKey(hash, &auth.KeyInfo{
		KeyAlias:  "expired-key",
		ExpiresAt: &expiredTime,
	})

	key, err := store.GetKeyByHash(context.Background(), hash)
	assert.NoError(t, err)
	assert.True(t, time.Now().After(*key.ExpiresAt))
}

