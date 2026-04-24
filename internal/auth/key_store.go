package auth

import (
	"context"
	"sync"
	"time"
)

// InMemoryKeyStore is a simple in-memory key store for development/testing.
// In production, this would be backed by a PostgreSQL database.
type InMemoryKeyStore struct {
	mu   sync.RWMutex
	keys map[string]*KeyInfo
}

// NewInMemoryKeyStore creates a new in-memory key store.
func NewInMemoryKeyStore() *InMemoryKeyStore {
	return &InMemoryKeyStore{
		keys: make(map[string]*KeyInfo),
	}
}

// AddKey adds a key to the store (for testing/admin purposes).
func (s *InMemoryKeyStore) AddKey(hash string, info *KeyInfo) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.keys[hash] = info
}

// GetKeyByHash retrieves a key by its hash.
func (s *InMemoryKeyStore) GetKeyByHash(ctx context.Context, hash string) (*KeyInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key, found := s.keys[hash]
	if !found {
		return nil, ErrInvalidAPIKey
	}

	return key, nil
}

// GenerateTestKeys populates the store with test keys.
func (s *InMemoryKeyStore) GenerateTestKeys() {
	// Admin key
	adminExpires := time.Now().Add(365 * 24 * time.Hour)
	s.AddKey(HashKey("sk-admin-12345"), &KeyInfo{
		KeyAlias:  "admin-key",
		Role:      "proxy_admin",
		Models:    []string{"*"},
		ExpiresAt: &adminExpires,
	})

	// Team key with limited models
	teamExpires := time.Now().Add(30 * 24 * time.Hour)
	maxBudget := 100.0
	rpmLimit := int64(100)
	tpmLimit := int64(100000)
	s.AddKey(HashKey("sk-team-key-67890"), &KeyInfo{
		KeyAlias:  "team-key",
		TeamID:    "team-1",
		Role:      "team",
		Models:    []string{"gpt-3.5-turbo", "gpt-4"},
		MaxBudget: &maxBudget,
		RPMLimit:  &rpmLimit,
		TPMLimit:  &tpmLimit,
		ExpiresAt: &teamExpires,
	})

	// Blocked key
	blocked := true
	s.AddKey(HashKey("sk-blocked-key"), &KeyInfo{
		KeyAlias: "blocked-key",
		Role:     "internal_user",
		Blocked:  blocked,
	})

	// Expired key
	expiredTime := time.Now().Add(-24 * time.Hour)
	s.AddKey(HashKey("sk-expired-key"), &KeyInfo{
		KeyAlias:  "expired-key",
		Role:      "internal_user",
		ExpiresAt: &expiredTime,
	})
}
