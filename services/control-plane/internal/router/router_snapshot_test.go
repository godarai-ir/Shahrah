package router

import (
	"context"
	"sync"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
	"github.com/Delpak/shahrah/services/control-plane/internal/config"
	"github.com/Delpak/shahrah/services/control-plane/internal/testcache"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type deploymentSnapshotStoreStub struct {
	mu          sync.Mutex
	deployments []platform.Deployment
	version     int64
	calls       int
}

func (s *deploymentSnapshotStoreStub) DeploymentSnapshot(
	context.Context,
	string,
) ([]platform.Deployment, int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.calls++

	return s.deployments, s.version, nil
}

type deploymentCacheReadCounter struct {
	cache.Cache

	mu    sync.Mutex
	reads int
}

func (c *deploymentCacheReadCounter) Get(ctx context.Context, key string) (interface{}, bool) {
	if key == deploymentCacheKey("tenant_a") {
		c.mu.Lock()
		c.reads++
		c.mu.Unlock()
	}

	return c.Cache.Get(ctx, key)
}

func TestRouterUsesVersionedDeploymentSnapshotWithoutDistributedCache(t *testing.T) {
	snapshotStore := &deploymentSnapshotStoreStub{
		version: 7,
		deployments: []platform.Deployment{{
			ID:        "mock-deployment",
			TenantID:  "tenant_a",
			ModelName: "mock:loadtest:chat",
			Provider:  "mock",
			Model:     "mock/chat",
			ProviderAuth: &platform.ResolvedProviderAuth{
				Mode: string(platform.ProviderAuthModeNoCredential),
			},
		}},
	}
	baseCache := newRouterTestCache(t)
	countedCache := &deploymentCacheReadCounter{Cache: baseCache}
	rt, err := NewRouterWithDependencies(
		nil,
		config.RouterSettings{CooldownTime: 60, NumRetries: 1},
		countedCache,
		Dependencies{DeploymentSnapshots: snapshotStore},
		zap.NewNop(),
	)
	require.NoError(t, err)

	_, err = rt.Route(t.Context(), "tenant_a", "mock:loadtest:chat")
	require.NoError(t, err)

	rt.mu.RLock()
	firstSnapshot := &rt.lastKnownGoodDeployments["tenant_a"][0]
	rt.mu.RUnlock()

	_, err = rt.Route(t.Context(), "tenant_a", "mock:loadtest:chat")
	require.NoError(t, err)

	rt.mu.RLock()
	secondSnapshot := &rt.lastKnownGoodDeployments["tenant_a"][0]
	rt.mu.RUnlock()

	require.Same(t, firstSnapshot, secondSnapshot)
	require.Equal(t, 0, countedCache.reads)
	require.Equal(t, 2, snapshotStore.calls)
}

func BenchmarkRouterRouteFromRuntimeSnapshot(b *testing.B) {
	memory, err := testcache.New(b, zap.NewNop())
	require.NoError(b, err)

	snapshotStore := &deploymentSnapshotStoreStub{
		version: 7,
		deployments: []platform.Deployment{{
			ID:        "mock-deployment",
			TenantID:  "tenant_a",
			ModelName: "mock:loadtest:chat",
			Provider:  "mock",
			Model:     "mock/chat",
			ProviderAuth: &platform.ResolvedProviderAuth{
				Mode: string(platform.ProviderAuthModeNoCredential),
			},
		}},
	}
	rt, err := NewRouterWithDependencies(
		nil,
		config.RouterSettings{CooldownTime: 60, NumRetries: 1},
		memory,
		Dependencies{DeploymentSnapshots: snapshotStore},
		zap.NewNop(),
	)
	require.NoError(b, err)

	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		_, routeErr := rt.Route(ctx, "tenant_a", "mock:loadtest:chat")
		if routeErr != nil {
			b.Fatal(routeErr)
		}
	}
}
