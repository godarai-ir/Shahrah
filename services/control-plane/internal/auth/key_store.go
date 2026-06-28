package auth

import (
	"context"
	"sync"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

// InMemoryKeyStore is a simple in-memory key store for development/testing.
// In production, this would be backed by a PostgreSQL database.
type InMemoryKeyStore struct {
	mu   sync.RWMutex
	keys map[string]*platform.AuthContext
}

// NewInMemoryKeyStore creates a new in-memory key store.
func NewInMemoryKeyStore() *InMemoryKeyStore {
	return &InMemoryKeyStore{
		keys: make(map[string]*platform.AuthContext),
	}
}

// AddKey adds a key to the store (for testing/admin purposes).
func (s *InMemoryKeyStore) AddKey(hash string, info *platform.AuthContext) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.keys[hash] = info
}

// GetKeyByHash retrieves a key by its hash.
func (s *InMemoryKeyStore) GetKeyByHash(ctx context.Context, hash string) (*platform.AuthContext, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key, found := s.keys[hash]
	if !found {
		return nil, ErrInvalidAPIKey
	}

	return key, nil
}

func (s *InMemoryKeyStore) ResolveTenantUserAuth(ctx context.Context, tenantID, userID string) (*platform.AuthContext, error) {
	return nil, ErrInvalidAPIKey
}

func (s *InMemoryKeyStore) TouchKeyLastUsed(ctx context.Context, keyID string, usedAt time.Time) error {
	return nil
}

func (s *InMemoryKeyStore) ListPATs(ctx context.Context, tenantID, userID, query string, limit, offset int) ([]platform.CreatedKey, int, error) {
	return []platform.CreatedKey{}, 0, nil
}

func (s *InMemoryKeyStore) PATExists(ctx context.Context, tenantID, userID, name string) (bool, error) {
	return false, nil
}

func (s *InMemoryKeyStore) RevokePAT(ctx context.Context, tenantID, userID, keyID, revokedByUserID string) (bool, error) {
	return false, nil
}

func (s *InMemoryKeyStore) RevokeAllPATs(ctx context.Context, tenantID, userID, revokedByUserID string) (int64, error) {
	return 0, nil
}

// GenerateTestKeys populates the store with test keys.
func (s *InMemoryKeyStore) GenerateTestKeys() {
	// Admin key
	adminExpires := time.Now().Add(365 * 24 * time.Hour)
	s.AddKey(HashKey("sk-admin-12345"), &platform.AuthContext{
		KeyAlias:      "admin-key",
		Role:          "proxy_admin",
		AllowedModels: []string{"*"},
		ExpiresAt:     &adminExpires,
	})

	// Team key with limited models
	teamExpires := time.Now().Add(30 * 24 * time.Hour)
	s.AddKey(HashKey("sk-team-key-67890"), &platform.AuthContext{
		KeyAlias:      "team-key",
		TeamID:        "team-1",
		Role:          "team",
		AllowedModels: []string{"gpt-3.5-turbo", "gpt-4"},
		ExpiresAt:     &teamExpires,
	})

	// Blocked key
	blocked := true
	s.AddKey(HashKey("sk-blocked-key"), &platform.AuthContext{
		KeyAlias: "blocked-key",
		Role:     "internal_user",
		Blocked:  blocked,
	})

	// Expired key
	expiredTime := time.Now().Add(-24 * time.Hour)
	s.AddKey(HashKey("sk-expired-key"), &platform.AuthContext{
		KeyAlias:  "expired-key",
		Role:      "internal_user",
		ExpiresAt: &expiredTime,
	})
}
