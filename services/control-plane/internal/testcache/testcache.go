// Package testcache provides bounded cache instances for control-plane tests.
package testcache

import (
	"testing"

	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
	"go.uber.org/zap"
)

const (
	testNumCounters = 1_000
	testMaxCost     = 10_000
	testBufferItems = 64
)

// New creates a low-memory cache and registers deterministic test cleanup.
func New(tb testing.TB, logger *zap.Logger) (*cache.MemoryCache, error) {
	tb.Helper()

	memory, err := cache.NewMemoryCacheWithConfig(logger, cache.MemoryCacheConfig{
		NumCounters: testNumCounters,
		MaxCost:     testMaxCost,
		BufferItems: testBufferItems,
	})
	if err != nil {
		return nil, err
	}

	tb.Cleanup(func() {
		if closeErr := memory.Close(); closeErr != nil {
			tb.Errorf("close test memory cache: %v", closeErr)
		}
	})

	return memory, nil
}
