package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/dgraph-io/ristretto/v2"
	"go.uber.org/zap"
)

// cacheEntry is a generic wrapper for cache values with metadata.
type cacheEntry struct {
	value any
}

// MemoryCache implements Cache using Ristretto v2 with generics.
type MemoryCache struct {
	cache          *ristretto.Cache[string, any]
	logger         *zap.Logger
	mu             sync.Mutex
	fixedWindows   map[string]fixedWindowCounter
	slidingWindows map[string][]slidingWindowEntry
}

type fixedWindowCounter struct {
	count     int64
	expiresAt time.Time
}

type slidingWindowEntry struct {
	tsMS   int64
	weight int64
}

// NewMemoryCache creates a new in-memory cache using Ristretto v2.
func NewMemoryCache(logger *zap.Logger) (*MemoryCache, error) {
	cache, err := ristretto.NewCache(&ristretto.Config[string, any]{
		NumCounters: 1e7,     // number of keys to track frequency of (10M)
		MaxCost:     1 << 30, // maximum cost of cache (1GB)
		BufferItems: 64,      // number of keys per Get buffer
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create ristretto v2 cache: %w", err)
	}
	return &MemoryCache{
		cache:          cache,
		logger:         logger,
		fixedWindows:   make(map[string]fixedWindowCounter),
		slidingWindows: make(map[string][]slidingWindowEntry),
	}, nil
}

// Get retrieves a value from the cache with proper type handling.
func (m *MemoryCache) Get(ctx context.Context, key string) (interface{}, bool) {
	val, found := m.cache.Get(key)
	if !found {
		return nil, false
	}
	return val, true
}

// Set stores a value in the cache with TTL support.
func (m *MemoryCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) {
	// Ristretto v2 uses SetWithTTL with cost parameter
	m.cache.SetWithTTL(key, value, 1, ttl)
	m.cache.Wait()
}

// Delete removes a value from the cache.
func (m *MemoryCache) Delete(ctx context.Context, key string) {
	m.cache.Del(key)
}

// Increment safely increments an int64 value in the cache.
func (m *MemoryCache) Increment(ctx context.Context, key string, value int64, ttl time.Duration) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	val, found := m.cache.Get(key)
	if !found {
		m.cache.SetWithTTL(key, value, 1, ttl)
		m.cache.Wait()
		return value, nil
	}

	// Type-safe access with proper error handling
	intVal, ok := val.(int64)
	if !ok {
		return 0, fmt.Errorf("expected int64 value, got %T", val)
	}

	newVal := intVal + value
	m.cache.SetWithTTL(key, newVal, 1, ttl)
	m.cache.Wait()
	return newVal, nil
}

// CheckRateLimit checks a fixed-window counter in memory.
func (m *MemoryCache) CheckRateLimit(ctx context.Context, key string, limit int64, window time.Duration) (*RateLimitResult, error) {
	if limit <= 0 {
		return &RateLimitResult{
			Allowed: true,
			Current: 0,
			Limit:   limit,
			ResetAt: time.Now().Add(window),
		}, nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	counter, exists := m.fixedWindows[key]
	if !exists || now.After(counter.expiresAt) {
		counter = fixedWindowCounter{
			count:     0,
			expiresAt: now.Add(window),
		}
	}

	if counter.count >= limit {
		return &RateLimitResult{
			Allowed: false,
			Current: counter.count,
			Limit:   limit,
			ResetAt: counter.expiresAt,
		}, nil
	}

	counter.count++
	m.fixedWindows[key] = counter
	return &RateLimitResult{
		Allowed: true,
		Current: counter.count,
		Limit:   limit,
		ResetAt: counter.expiresAt,
	}, nil
}

// CheckSlidingWindowRateLimit checks a true sliding-window log in memory.
func (m *MemoryCache) CheckSlidingWindowRateLimit(ctx context.Context, key string, limit int64, window time.Duration) (*RateLimitResult, error) {
	return m.CheckSlidingWindowRateLimitWithWeight(ctx, key, limit, 1, window)
}

// CheckSlidingWindowRateLimitWithWeight checks a true sliding-window log in memory with weighted entries.
func (m *MemoryCache) CheckSlidingWindowRateLimitWithWeight(ctx context.Context, key string, limit int64, weight int64, window time.Duration) (*RateLimitResult, error) {
	if limit <= 0 {
		return &RateLimitResult{
			Allowed: true,
			Current: 0,
			Limit:   limit,
			ResetAt: time.Now().Add(window),
		}, nil
	}
	if weight < 1 {
		weight = 1
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	nowMS := now.UnixMilli()
	cutoffMS := now.Add(-window).UnixMilli()
	entries := m.slidingWindows[key]

	pruned := make([]slidingWindowEntry, 0, len(entries))
	var total int64
	for _, ts := range entries {
		if ts.tsMS > cutoffMS {
			pruned = append(pruned, ts)
			total += ts.weight
		}
	}

	if total+weight > limit {
		resetAt := now.Add(window)
		if len(pruned) > 0 {
			resetAt = time.UnixMilli(pruned[0].tsMS).Add(window)
		}
		m.slidingWindows[key] = pruned
		return &RateLimitResult{
			Allowed: false,
			Current: total,
			Limit:   limit,
			ResetAt: resetAt,
		}, nil
	}

	pruned = append(pruned, slidingWindowEntry{tsMS: nowMS, weight: weight})
	m.slidingWindows[key] = pruned
	return &RateLimitResult{
		Allowed: true,
		Current: total + weight,
		Limit:   limit,
		ResetAt: now.Add(window),
	}, nil
}

// Close closes the cache and releases resources.
func (m *MemoryCache) Close() error {
	m.cache.Close()
	return nil
}
