package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/dgraph-io/ristretto"
	"go.uber.org/zap"
)

// MemoryCache implements Cache using Ristretto.
type MemoryCache struct {
	cache  *ristretto.Cache
	logger *zap.Logger
}

// NewMemoryCache creates a new in-memory cache.
func NewMemoryCache(logger *zap.Logger) (*MemoryCache, error) {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,     // number of keys to track frequency of
		MaxCost:     1 << 30, // maximum cost of cache (1GB)
		BufferItems: 64,      // number of keys per Get buffer
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create ristretto cache: %w", err)
	}
	return &MemoryCache{cache: cache, logger: logger}, nil
}

// Get retrieves a value from the cache.
func (m *MemoryCache) Get(ctx context.Context, key string) (interface{}, bool) {
	return m.cache.Get(key)
}

// Set stores a value in the cache.
func (m *MemoryCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) {
	m.cache.SetWithTTL(key, value, 1, ttl)
	m.cache.Wait()
}

// Delete removes a value from the cache.
func (m *MemoryCache) Delete(ctx context.Context, key string) {
	m.cache.Del(key)
}

// Increment increments a value in the cache.
func (m *MemoryCache) Increment(ctx context.Context, key string, value int64, ttl time.Duration) (int64, error) {
	val, found := m.cache.Get(key)
	if !found {
		m.cache.SetWithTTL(key, value, 1, ttl)
		m.cache.Wait()
		return value, nil
	}
	newVal := val.(int64) + value
	m.cache.SetWithTTL(key, newVal, 1, ttl)
	m.cache.Wait()
	return newVal, nil
}

// Close closes the cache.
func (m *MemoryCache) Close() error {
	m.cache.Close()
	return nil
}
