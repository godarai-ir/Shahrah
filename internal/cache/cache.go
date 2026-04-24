package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Cache interface defines the contract for cache implementations.
type Cache interface {
	Get(ctx context.Context, key string) (interface{}, bool)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration)
	Delete(ctx context.Context, key string)
	Increment(ctx context.Context, key string, value int64, ttl time.Duration) (int64, error)
	Close() error
}

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

// DualCache implements Cache using both memory and Redis.
type DualCache struct {
	memory *MemoryCache
	redis  *redis.Client
	logger *zap.Logger
	mu     sync.RWMutex
}

// NewDualCache creates a new dual cache (memory + Redis).
func NewDualCache(redisHost string, redisPort int, redisPassword string, logger *zap.Logger) (*DualCache, error) {
	memory, err := NewMemoryCache(logger)
	if err != nil {
		return nil, err
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", redisHost, redisPort),
		Password: redisPassword,
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &DualCache{memory: memory, redis: rdb, logger: logger}, nil
}

// Get retrieves a value from the cache (memory first, then Redis).
func (d *DualCache) Get(ctx context.Context, key string) (interface{}, bool) {
	// Try memory cache first
	if val, found := d.memory.Get(ctx, key); found {
		return val, true
	}

	// Fallback to Redis
	val, err := d.redis.Get(ctx, key).Result()
	if err != nil {
		return nil, false
	}

	return val, true
}

// Set stores a value in both memory and Redis.
func (d *DualCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) {
	d.memory.Set(ctx, key, value, ttl)
	d.redis.Set(ctx, key, value, ttl)
}

// Delete removes a value from both memory and Redis.
func (d *DualCache) Delete(ctx context.Context, key string) {
	d.memory.Delete(ctx, key)
	d.redis.Del(ctx, key)
}

// Increment increments a value in Redis, falling back to memory.
func (d *DualCache) Increment(ctx context.Context, key string, value int64, ttl time.Duration) (int64, error) {
	// Increment in Redis
	result, err := d.redis.IncrBy(ctx, key, value).Result()
	if err != nil {
		// Fallback to memory
		return d.memory.Increment(ctx, key, value, ttl)
	}

	// Sync to memory cache
	d.memory.Set(ctx, key, result, ttl)

	return result, nil
}

// Close closes both memory and Redis caches.
func (d *DualCache) Close() error {
	d.memory.Close()
	return d.redis.Close()
}
