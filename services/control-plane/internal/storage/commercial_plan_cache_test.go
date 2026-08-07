package storage

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
)

type entitlementTestCache struct {
	mu    sync.Mutex
	items map[string]entitlementTestCacheItem
}

type entitlementTestCacheItem struct {
	value     interface{}
	expiresAt time.Time
}

func newEntitlementTestCache() *entitlementTestCache {
	return &entitlementTestCache{items: map[string]entitlementTestCacheItem{}}
}

func (c *entitlementTestCache) Get(_ context.Context, key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, ok := c.items[key]
	if !ok {
		return nil, false
	}

	if !item.expiresAt.IsZero() && time.Now().After(item.expiresAt) {
		delete(c.items, key)
		return nil, false
	}

	return item.value, true
}

func (c *entitlementTestCache) Set(_ context.Context, key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item := entitlementTestCacheItem{value: value}
	if ttl > 0 {
		item.expiresAt = time.Now().Add(ttl)
	}

	c.items[key] = item
}

func (c *entitlementTestCache) Delete(_ context.Context, key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
}

func (c *entitlementTestCache) Increment(context.Context, string, int64, time.Duration) (int64, error) {
	return 0, nil
}

func (c *entitlementTestCache) CheckRateLimit(context.Context, string, int64, time.Duration) (*cache.RateLimitResult, error) {
	return nil, nil
}

func (c *entitlementTestCache) CheckSlidingWindowRateLimit(context.Context, string, int64, time.Duration) (*cache.RateLimitResult, error) {
	return nil, nil
}

func (c *entitlementTestCache) CheckSlidingWindowRateLimitWithWeight(context.Context, string, int64, int64, time.Duration) (*cache.RateLimitResult, error) {
	return nil, nil
}

func (c *entitlementTestCache) Close() error {
	return nil
}

func TestEffectiveEntitlementsCacheOmitsUsageAndClones(t *testing.T) {
	ctx := context.Background()
	limit := 10.0
	store := &PlanRepository{
		entitlementCache:    newEntitlementTestCache(),
		entitlementCacheTTL: time.Minute,
	}

	resolved := &platform.EffectiveEntitlementsResponse{
		TenantID: "tenant-1",
		Subscription: &platform.TenantSubscription{
			TenantID:      "tenant-1",
			PriceSnapshot: map[string]interface{}{"amount": float64(100)},
			Metadata:      map[string]interface{}{"region": "ir"},
		},
		Features: map[string]interface{}{
			"requests_per_minute": float64(1000),
			"nested":              map[string]interface{}{"enabled": true},
		},
		Overrides: map[string]interface{}{"requests_per_minute": float64(1000)},
		UsageCounters: []platform.UsageCounter{{
			FeatureKey: "requests_per_minute",
			Used:       2,
			Limit:      &limit,
		}},
		ResolvedAt: time.Now().UTC(),
	}

	store.cacheEntitlements(ctx, resolved)
	resolved.Features["requests_per_minute"] = float64(1)
	resolved.Subscription.Metadata["region"] = "changed"

	cached, ok := store.getCachedEntitlements(ctx, "tenant-1")
	if !ok {
		t.Fatal("expected cached entitlements")
	}

	if len(cached.UsageCounters) != 0 {
		t.Fatalf("expected usage counters to be omitted from cache, got %d", len(cached.UsageCounters))
	}

	if cached.Features["requests_per_minute"] != float64(1000) {
		t.Fatalf("expected cached feature to keep original value, got %v", cached.Features["requests_per_minute"])
	}

	if cached.Subscription.Metadata["region"] != "ir" {
		t.Fatalf("expected cached subscription metadata to keep original value, got %v", cached.Subscription.Metadata["region"])
	}

	cached.Features["requests_per_minute"] = float64(5)

	again, ok := store.getCachedEntitlements(ctx, "tenant-1")
	if !ok {
		t.Fatal("expected cached entitlements on second read")
	}

	if again.Features["requests_per_minute"] != float64(1000) {
		t.Fatalf("expected defensive copy on read, got %v", again.Features["requests_per_minute"])
	}
}

func TestInvalidateEffectiveEntitlementsDeletesCache(t *testing.T) {
	ctx := context.Background()
	store := &PlanRepository{
		entitlementCache:    newEntitlementTestCache(),
		entitlementCacheTTL: time.Minute,
	}

	store.cacheEntitlements(ctx, &platform.EffectiveEntitlementsResponse{
		TenantID: "tenant-1",
		Features: map[string]interface{}{"enabled": true},
	})
	store.InvalidateEffectiveEntitlements(ctx, "tenant-1")

	if _, ok := store.getCachedEntitlements(ctx, "tenant-1"); ok {
		t.Fatal("expected cache entry to be invalidated")
	}
}

func TestEffectiveEntitlementsCacheHonorsTTL(t *testing.T) {
	ctx := context.Background()
	store := &PlanRepository{
		entitlementCache:    newEntitlementTestCache(),
		entitlementCacheTTL: 10 * time.Millisecond,
	}

	store.cacheEntitlements(ctx, &platform.EffectiveEntitlementsResponse{
		TenantID: "tenant-1",
		Features: map[string]interface{}{"enabled": true},
	})
	time.Sleep(20 * time.Millisecond)

	if _, ok := store.getCachedEntitlements(ctx, "tenant-1"); ok {
		t.Fatal("expected cache entry to expire")
	}
}

func TestInvalidateDeploymentCacheDeletesTenantSnapshot(t *testing.T) {
	ctx := context.Background()
	testCache := newEntitlementTestCache()
	store := &ModelRepository{sharedCache: testCache}

	testCache.Set(ctx, deploymentCacheKey("tenant-1"), []platform.Deployment{{ID: "dep-1"}}, time.Minute)
	store.invalidateDeploymentCache(ctx, "tenant-1")

	if _, ok := testCache.Get(ctx, deploymentCacheKey("tenant-1")); ok {
		t.Fatal("expected deployment cache entry to be invalidated")
	}
}

func TestInvalidateWalletCacheDeletesTenantBalance(t *testing.T) {
	ctx := context.Background()
	testCache := newEntitlementTestCache()
	store := &WalletRepository{entitlementCache: testCache}

	testCache.Set(ctx, walletCacheKey("tenant-1"), &platform.Wallet{TenantID: "tenant-1"}, time.Minute)
	store.invalidateWalletCache(ctx, "tenant-1")

	if _, ok := testCache.Get(ctx, walletCacheKey("tenant-1")); ok {
		t.Fatal("expected wallet cache entry to be invalidated")
	}
}
