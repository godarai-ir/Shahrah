package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/testcache"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestMemoryCache_CheckRateLimit(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mem, err := testcache.New(t, logger)
	require.NoError(t, err)

	key := "rpm:key:test:gpt-4"
	window := 200 * time.Millisecond

	first, err := mem.CheckRateLimit(context.Background(), key, 2, window)
	require.NoError(t, err)
	assert.True(t, first.Allowed)
	assert.EqualValues(t, 1, first.Current)

	second, err := mem.CheckRateLimit(context.Background(), key, 2, window)
	require.NoError(t, err)
	assert.True(t, second.Allowed)
	assert.EqualValues(t, 2, second.Current)

	third, err := mem.CheckRateLimit(context.Background(), key, 2, window)
	require.NoError(t, err)
	assert.False(t, third.Allowed)
	assert.EqualValues(t, 2, third.Current)

	time.Sleep(window + 30*time.Millisecond)

	afterReset, err := mem.CheckRateLimit(context.Background(), key, 2, window)
	require.NoError(t, err)
	assert.True(t, afterReset.Allowed)
	assert.EqualValues(t, 1, afterReset.Current)
}

func TestMemoryCache_CheckSlidingWindowRateLimitWithWeight(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mem, err := testcache.New(t, logger)
	require.NoError(t, err)

	key := "tpm:key:test:gpt-4"
	window := 200 * time.Millisecond

	first, err := mem.CheckSlidingWindowRateLimitWithWeight(context.Background(), key, 100, 60, window)
	require.NoError(t, err)
	assert.True(t, first.Allowed)
	assert.EqualValues(t, 60, first.Current)

	second, err := mem.CheckSlidingWindowRateLimitWithWeight(context.Background(), key, 100, 50, window)
	require.NoError(t, err)
	assert.False(t, second.Allowed)
	assert.EqualValues(t, 60, second.Current)

	time.Sleep(window + 30*time.Millisecond)

	afterReset, err := mem.CheckSlidingWindowRateLimitWithWeight(context.Background(), key, 100, 50, window)
	require.NoError(t, err)
	assert.True(t, afterReset.Allowed)
	assert.EqualValues(t, 50, afterReset.Current)
}
