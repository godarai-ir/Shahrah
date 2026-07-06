package storage

import (
	"context"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

func TestListDeploymentsReturnsCachedDefensiveCopy(t *testing.T) {
	ctx := context.Background()
	store := &Store{
		hotPathCacheTTL: time.Minute,
		hotPathCache:    make(map[string]hotPathCacheEntry),
	}
	key := deploymentCacheKey("tenant-1")
	store.cacheDeployments(key, []platform.Deployment{{
		ID:        "dep-1",
		TenantID:  "tenant-1",
		ModelName: "gpt-4o",
		ModelInfo: map[string]interface{}{
			"family": "gpt",
		},
	}})

	first, err := store.ListDeployments(ctx, "tenant-1")
	if err != nil {
		t.Fatalf("expected cached deployment list, got error: %v", err)
	}
	if len(first) != 1 || first[0].ModelName != "gpt-4o" {
		t.Fatalf("unexpected cached deployments: %#v", first)
	}
	first[0].ModelName = "mutated"
	first[0].ModelInfo["family"] = "changed"

	second, err := store.ListDeployments(ctx, "tenant-1")
	if err != nil {
		t.Fatalf("expected cached deployment list on second read, got error: %v", err)
	}
	if second[0].ModelName != "gpt-4o" {
		t.Fatalf("expected cached deployment copy to preserve model name, got %q", second[0].ModelName)
	}
	if second[0].ModelInfo["family"] != "gpt" {
		t.Fatalf("expected cached deployment copy to preserve model info, got %#v", second[0].ModelInfo)
	}
}

func TestListVirtualModelsReturnsCachedDefensiveCopy(t *testing.T) {
	ctx := context.Background()
	store := &Store{
		hotPathCacheTTL: time.Minute,
		hotPathCache:    make(map[string]hotPathCacheEntry),
	}
	key := virtualModelsCacheKey("tenant-1")
	store.cacheVirtualModels(key, []platform.VirtualModel{{
		ID:           "vm-1",
		TenantID:     "tenant-1",
		Alias:        "smart",
		Enabled:      true,
		ExposedModes: []string{"chat"},
		Targets: []platform.VirtualModelTarget{{
			ID:       "target-1",
			Provider: "openai",
			Model:    "gpt-4o",
			Policy: platform.VirtualModelPolicy{
				Retry:    platform.VirtualModelRetryPolicy{StatusCodes: []int{429}},
				Fallback: platform.VirtualModelFallbackPolicy{StatusCodes: []int{500}},
			},
		}},
		Access: platform.TenantProviderAccountAccessControl{
			Users: []platform.AccessPrincipal{{Type: platform.PrincipalTypeEveryone}},
		},
	}})

	first, err := store.ListVirtualModels(ctx, "tenant-1")
	if err != nil {
		t.Fatalf("expected cached virtual model list, got error: %v", err)
	}
	first[0].Alias = "mutated"
	first[0].ExposedModes[0] = "responses"
	first[0].Targets[0].Policy.Retry.StatusCodes[0] = 400
	first[0].Access.Users[0].Type = platform.PrincipalTypeUser

	second, err := store.ListVirtualModels(ctx, "tenant-1")
	if err != nil {
		t.Fatalf("expected cached virtual model list on second read, got error: %v", err)
	}
	if second[0].Alias != "smart" {
		t.Fatalf("expected cached virtual model alias to be preserved, got %q", second[0].Alias)
	}
	if second[0].ExposedModes[0] != "chat" {
		t.Fatalf("expected cached exposed modes to be preserved, got %#v", second[0].ExposedModes)
	}
	if second[0].Targets[0].Policy.Retry.StatusCodes[0] != 429 {
		t.Fatalf("expected cached target policy to be preserved, got %#v", second[0].Targets[0].Policy.Retry.StatusCodes)
	}
	if second[0].Access.Users[0].Type != platform.PrincipalTypeEveryone {
		t.Fatalf("expected cached access to be preserved, got %#v", second[0].Access.Users)
	}
}

func TestHotPathCacheHonorsTTL(t *testing.T) {
	store := &Store{
		hotPathCacheTTL: 10 * time.Millisecond,
		hotPathCache:    make(map[string]hotPathCacheEntry),
	}
	key := deploymentCacheKey("tenant-1")
	store.cacheDeployments(key, []platform.Deployment{{ID: "dep-1"}})

	if _, ok := store.getCachedDeployments(key); !ok {
		t.Fatal("expected deployment cache hit before TTL expiry")
	}
	time.Sleep(20 * time.Millisecond)
	if _, ok := store.getCachedDeployments(key); ok {
		t.Fatal("expected deployment cache miss after TTL expiry")
	}
}

func TestInvalidateVirtualModelsCacheDeletesTenantSnapshot(t *testing.T) {
	ctx := context.Background()
	store := &Store{
		hotPathCacheTTL: time.Minute,
		hotPathCache:    make(map[string]hotPathCacheEntry),
	}
	key := virtualModelsCacheKey("tenant-1")
	store.cacheVirtualModels(key, []platform.VirtualModel{{ID: "vm-1", TenantID: "tenant-1"}})

	store.invalidateVirtualModelsCache(ctx, "tenant-1")

	if _, ok := store.getCachedVirtualModels(key); ok {
		t.Fatal("expected virtual model cache entry to be invalidated")
	}
}

func TestInvalidateDeploymentCacheDeletesLocalTenantSnapshot(t *testing.T) {
	ctx := context.Background()
	store := &Store{
		hotPathCacheTTL: time.Minute,
		hotPathCache:    make(map[string]hotPathCacheEntry),
	}
	key := deploymentCacheKey("tenant-1")
	store.cacheDeployments(key, []platform.Deployment{{ID: "dep-1", TenantID: "tenant-1"}})

	store.invalidateDeploymentCache(ctx, "tenant-1")

	if _, ok := store.getCachedDeployments(key); ok {
		t.Fatal("expected local deployment cache entry to be invalidated")
	}
}
