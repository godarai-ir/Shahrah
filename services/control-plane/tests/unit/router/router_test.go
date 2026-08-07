package router_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
	"github.com/Delpak/shahrah/services/control-plane/internal/config"
	llmrouter "github.com/Delpak/shahrah/services/control-plane/internal/router"
	"github.com/Delpak/shahrah/services/control-plane/internal/testcache"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestRouter_ShouldSelectDeployment verifies that the router selects a healthy deployment.
func TestRouter_ShouldSelectDeployment(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := testcache.New(t, logger)

	modelList := []config.ModelDeployment{
		{
			ModelName: "gpt-3.5-turbo",
			ProviderEndpoint: config.ProviderEndpointConfig{
				Model: "openai/gpt-3.5-turbo",
				RPM:   1000,
			},
			ModelInfo: map[string]interface{}{"id": "dep-1"},
		},
	}

	settings := config.RouterSettings{
		RoutingStrategy: "simple-shuffle",
		CooldownTime:    60,
		NumRetries:      3,
	}

	r, err := llmrouter.NewRouter(modelList, settings, memCache, nil, logger)
	require.NoError(t, err)
	assert.NotNil(t, r)

	dep, err := r.Route(context.Background(), "", "gpt-3.5-turbo")
	require.NoError(t, err)
	assert.NotNil(t, dep)
	assert.Equal(t, "gpt-3.5-turbo", dep.ModelName)
}

// TestRouter_ShouldReturnError_WhenModelNotFound verifies that the router returns an error for unknown models.
func TestRouter_ShouldReturnError_WhenModelNotFound(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := testcache.New(t, logger)

	modelList := []config.ModelDeployment{}
	settings := config.RouterSettings{}

	r, err := llmrouter.NewRouter(modelList, settings, memCache, nil, logger)
	require.NoError(t, err)

	dep, err := r.Route(context.Background(), "", "unknown-model")
	assert.Error(t, err)
	assert.Nil(t, dep)
	assert.Equal(t, llmrouter.ErrModelNotFound, err)
}

// TestRouter_ShouldExcludeCooldownedDeployment verifies that cooldowned deployments are excluded from routing.
func TestRouter_ShouldExcludeCooldownedDeployment(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := testcache.New(t, logger)

	modelList := []config.ModelDeployment{
		{
			ModelName: "test-model",
			ProviderEndpoint: config.ProviderEndpointConfig{
				Model: "openai/test-model",
				RPM:   1000,
			},
			ModelInfo: map[string]interface{}{"id": "dep-1"},
		},
	}

	settings := config.RouterSettings{
		CooldownTime: 60,
	}

	r, err := llmrouter.NewRouter(modelList, settings, memCache, nil, logger)
	require.NoError(t, err)

	// First, select the deployment
	dep, err := r.Route(context.Background(), "", "test-model")
	require.NoError(t, err)
	assert.NotNil(t, dep)

	// Mark it as failed (adds to cooldown)
	r.MarkFailed(context.Background(), dep, "test failure")

	// Should fail to route since the only deployment is in cooldown
	dep2, err := r.Route(context.Background(), "", "test-model")
	assert.Error(t, err)
	assert.Nil(t, dep2)
	assert.Equal(t, llmrouter.ErrNoHealthyDeployments, err)
}

// TestRouter_WeightedSelection_ShouldPreferHigherWeight verifies weighted selection favors higher weight deployments.
func TestRouter_WeightedSelection_ShouldPreferHigherWeight(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := testcache.New(t, logger)

	modelList := []config.ModelDeployment{
		{
			ModelName: "gpt-3.5-turbo",
			ProviderEndpoint: config.ProviderEndpointConfig{
				Model: "openai/gpt-3.5-turbo",
				RPM:   1000, // High weight
			},
			ModelInfo: map[string]interface{}{"id": "dep-high"},
		},
		{
			ModelName: "gpt-3.5-turbo",
			ProviderEndpoint: config.ProviderEndpointConfig{
				Model: "openai/gpt-3.5-turbo",
				RPM:   100, // Low weight
			},
			ModelInfo: map[string]interface{}{"id": "dep-low"},
		},
	}

	settings := config.RouterSettings{
		RoutingStrategy: "simple-shuffle",
	}

	r, err := llmrouter.NewRouter(modelList, settings, memCache, nil, logger)
	require.NoError(t, err)

	// Run 1000 selections, verify distribution roughly matches weights
	highCount := 0
	lowCount := 0
	iterations := 1000

	for i := 0; i < iterations; i++ {
		dep, err := r.Route(context.Background(), "", "gpt-3.5-turbo")
		require.NoError(t, err)

		if dep.ID == "dep-high" {
			highCount++
		} else {
			lowCount++
		}
	}

	// High weight deployment should be selected more often (1000 vs 100 = 10x weight)
	// Expect > 80% to be high-weight (allowing some randomness)
	highPercent := float64(highCount) / float64(iterations)
	assert.Greater(t, highPercent, 0.80, "High-weight deployment should be selected >80% of the time")
}

func TestRouter_ShouldMatchProviderScopedWildcardBeforeGlobalWildcard(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := testcache.New(t, logger)

	modelList := []config.ModelDeployment{
		{
			ModelName: "anthropic/*",
			ProviderEndpoint: config.ProviderEndpointConfig{
				Model: "anthropic/*",
				RPM:   500,
			},
			ModelInfo: map[string]interface{}{"id": "dep-anthropic"},
		},
		{
			ModelName: "*",
			ProviderEndpoint: config.ProviderEndpointConfig{
				Model: "openai/*",
				RPM:   1000,
			},
			ModelInfo: map[string]interface{}{"id": "dep-openai"},
		},
	}

	r, err := llmrouter.NewRouter(modelList, config.RouterSettings{}, memCache, nil, logger)
	require.NoError(t, err)

	dep, err := r.Route(context.Background(), "", "anthropic/claude-3-5-sonnet-latest")
	require.NoError(t, err)
	assert.NotNil(t, dep)
	assert.Equal(t, "dep-anthropic", dep.ID)
	assert.Equal(t, "anthropic", dep.Provider)
	assert.Equal(t, "*", dep.Model)
}

type staticDeploymentStore struct {
	mu          sync.Mutex
	deployments []platform.Deployment
	err         error
	calls       int
}

func (s *staticDeploymentStore) ListDeployments(_ context.Context, _ string) ([]platform.Deployment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.calls++

	if s.err != nil {
		return nil, s.err
	}

	return append([]platform.Deployment(nil), s.deployments...), nil
}

func (s *staticDeploymentStore) ResolveProviderAccountAPIKey(_ context.Context, _ string, modelFQN string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.err != nil {
		return "", s.err
	}

	modelFQN = strings.TrimSpace(modelFQN)
	for _, dep := range s.deployments {
		if strings.TrimSpace(dep.ModelName) == modelFQN && strings.TrimSpace(dep.APIKey) != "" {
			return dep.APIKey, nil
		}
	}

	return "", errors.New("api key not found")
}

func (s *staticDeploymentStore) callCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.calls
}

func (s *staticDeploymentStore) setError(err error) {
	s.mu.Lock()
	s.err = err
	s.mu.Unlock()
}

func newRouterWithStaticDeploymentStore(
	modelList []config.ModelDeployment,
	c cache.Cache,
	store *staticDeploymentStore,
	logger *zap.Logger,
) (*llmrouter.Router, error) {
	return llmrouter.NewRouterWithDependencies(
		modelList,
		config.RouterSettings{},
		c,
		llmrouter.Dependencies{
			DeploymentStore:                 store,
			LegacyProviderAccountAPIKey:     store,
			ProviderAccountCredentialPolicy: llmrouter.ProviderAccountCredentialsFailOpen,
		},
		logger,
	)
}

func TestRouter_FQNShouldNotFallBackToGlobalWildcard(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := testcache.New(t, logger)

	modelList := []config.ModelDeployment{
		{
			ModelName: "*",
			ProviderEndpoint: config.ProviderEndpointConfig{
				Model: "openai/*",
				RPM:   1000,
			},
			ModelInfo: map[string]interface{}{"id": "dep-openai"},
		},
	}

	store := &staticDeploymentStore{
		deployments: []platform.Deployment{
			{
				ID:        "provider-account:anthropic-1:anthropic/claude-3-haiku-20240307",
				TenantID:  "tenant-1",
				ModelName: "anthropic:anthropic-sunny-zephyr:claude-3-haiku-20240307",
				Provider:  "anthropic",
				Model:     "anthropic/claude-3-haiku-20240307",
				APIKey:    "sk-ant-test",
			},
		},
	}

	r, err := newRouterWithStaticDeploymentStore(modelList, memCache, store, logger)
	require.NoError(t, err)

	dep, err := r.Route(context.Background(), "tenant-1", "anthropic:anthropic-sunny-zephyr:claude-3-haiku-20240307")
	require.NoError(t, err)
	assert.NotNil(t, dep)
	assert.Equal(t, "provider-account:anthropic-1:anthropic/claude-3-haiku-20240307", dep.ID)
	assert.Equal(t, "anthropic", dep.Provider)
}

func TestRouter_FQNWithoutExactDeploymentShouldNotUseWildcard(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := testcache.New(t, logger)

	modelList := []config.ModelDeployment{
		{
			ModelName: "*",
			ProviderEndpoint: config.ProviderEndpointConfig{
				Model: "openai/*",
				RPM:   1000,
			},
			ModelInfo: map[string]interface{}{"id": "dep-openai"},
		},
	}

	r, err := newRouterWithStaticDeploymentStore(modelList, memCache, &staticDeploymentStore{}, logger)
	require.NoError(t, err)

	dep, err := r.Route(context.Background(), "tenant-1", "anthropic:anthropic-sunny-zephyr:claude-3-haiku-20240307")
	assert.Error(t, err)
	assert.Nil(t, dep)
	assert.Equal(t, llmrouter.ErrModelNotFound, err)
}

func TestRouter_FQNProviderAccountWithoutBYOKUsesSharedProviderKey(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := testcache.New(t, logger)

	modelList := []config.ModelDeployment{
		{
			ModelName: "*",
			ProviderEndpoint: config.ProviderEndpointConfig{
				Model:  "openai/*",
				APIKey: "shared-openai-key",
				RPM:    1000,
			},
			ModelInfo: map[string]interface{}{"id": "dep-openai"},
		},
	}

	store := &staticDeploymentStore{
		deployments: []platform.Deployment{
			{
				ID:        "provider-account:openai-1:gpt-5-nano",
				TenantID:  "tenant-1",
				ModelName: "openai:openai-daring-lotus:gpt-5-nano",
				Provider:  "openai",
				Model:     "gpt-5-nano",
			},
		},
	}

	r, err := newRouterWithStaticDeploymentStore(modelList, memCache, store, logger)
	require.NoError(t, err)

	dep, err := r.Route(context.Background(), "tenant-1", "openai:openai-daring-lotus:gpt-5-nano")
	require.NoError(t, err)
	assert.NotNil(t, dep)
	assert.Equal(t, "provider-account:openai-1:gpt-5-nano", dep.ID)
	assert.Equal(t, "openai", dep.Provider)
	assert.Equal(t, "shared-openai-key", dep.Params.APIKey)
}

func TestRouter_CanonicalizesProviderPrefixedModelForProviderAccountDeployment(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := testcache.New(t, logger)

	store := &staticDeploymentStore{
		deployments: []platform.Deployment{
			{
				ID:        "provider-account:openai-1:openai/gpt-4o-mini",
				TenantID:  "tenant-1",
				ModelName: "openai:openai-daring-lotus:gpt-4o-mini",
				Provider:  "openai",
				Model:     "openai/gpt-4o-mini",
				APIKey:    "tenant-openai-key",
			},
		},
	}

	r, err := newRouterWithStaticDeploymentStore(nil, memCache, store, logger)
	require.NoError(t, err)

	dep, err := r.Route(context.Background(), "tenant-1", "openai:openai-daring-lotus:gpt-4o-mini")
	require.NoError(t, err)
	assert.NotNil(t, dep)
	assert.Equal(t, "gpt-4o-mini", dep.Model)
	assert.Equal(t, "gpt-4o-mini", dep.Params.Model)
}

func TestRouter_CachesTenantDeployments(t *testing.T) {
	t.Setenv("ROUTER_DEPLOYMENT_CACHE_TTL_SECONDS", "30")

	logger, _ := zap.NewDevelopment()
	memCache, _ := testcache.New(t, logger)
	store := &staticDeploymentStore{
		deployments: []platform.Deployment{
			{
				ID:        "provider-account:openai-1:gpt-4o-mini",
				TenantID:  "tenant-1",
				ModelName: "openai:openai-main:gpt-4o-mini",
				Provider:  "openai",
				Model:     "gpt-4o-mini",
				APIKey:    "tenant-openai-key",
			},
		},
	}

	r, err := newRouterWithStaticDeploymentStore(nil, memCache, store, logger)
	require.NoError(t, err)

	for i := 0; i < 2; i++ {
		dep, err := r.Route(context.Background(), "tenant-1", "openai:openai-main:gpt-4o-mini")
		require.NoError(t, err)
		assert.NotNil(t, dep)
		assert.Equal(t, "tenant-openai-key", dep.Params.APIKey)
	}

	assert.Equal(t, 1, store.callCount())

	cached, ok := memCache.Get(context.Background(), "deployments:tenant:tenant-1")
	require.True(t, ok)
	assert.NotContains(t, cached, "tenant-openai-key")
}

func TestRouter_ReloadsTenantDeploymentsAfterCacheInvalidation(t *testing.T) {
	t.Setenv("ROUTER_DEPLOYMENT_CACHE_TTL_SECONDS", "30")

	logger, _ := zap.NewDevelopment()
	memCache, _ := testcache.New(t, logger)
	store := &staticDeploymentStore{
		deployments: []platform.Deployment{
			{
				ID:        "provider-account:openai-1:gpt-4o-mini",
				TenantID:  "tenant-1",
				ModelName: "openai:openai-main:gpt-4o-mini",
				Provider:  "openai",
				Model:     "gpt-4o-mini",
				APIKey:    "tenant-openai-key",
			},
		},
	}

	r, err := newRouterWithStaticDeploymentStore(nil, memCache, store, logger)
	require.NoError(t, err)

	dep, err := r.Route(context.Background(), "tenant-1", "openai:openai-main:gpt-4o-mini")
	require.NoError(t, err)
	assert.NotNil(t, dep)

	memCache.Delete(context.Background(), "deployments:tenant:tenant-1")

	dep, err = r.Route(context.Background(), "tenant-1", "openai:openai-main:gpt-4o-mini")
	require.NoError(t, err)
	assert.NotNil(t, dep)

	assert.Equal(t, 2, store.callCount())
}

func TestRouter_ReloadsTenantDeploymentsAfterTTLExpiry(t *testing.T) {
	t.Setenv("ROUTER_DEPLOYMENT_CACHE_TTL_SECONDS", "1")

	logger, _ := zap.NewDevelopment()
	memCache, _ := testcache.New(t, logger)
	store := &staticDeploymentStore{
		deployments: []platform.Deployment{
			{
				ID:        "provider-account:openai-1:gpt-4o-mini",
				TenantID:  "tenant-1",
				ModelName: "openai:openai-main:gpt-4o-mini",
				Provider:  "openai",
				Model:     "gpt-4o-mini",
				APIKey:    "tenant-openai-key",
			},
		},
	}

	r, err := newRouterWithStaticDeploymentStore(nil, memCache, store, logger)
	require.NoError(t, err)

	dep, err := r.Route(context.Background(), "tenant-1", "openai:openai-main:gpt-4o-mini")
	require.NoError(t, err)
	assert.NotNil(t, dep)

	time.Sleep(1100 * time.Millisecond)

	dep, err = r.Route(context.Background(), "tenant-1", "openai:openai-main:gpt-4o-mini")
	require.NoError(t, err)
	assert.NotNil(t, dep)

	assert.Equal(t, 2, store.callCount())
}

func TestRouter_UsesLastKnownGoodDeploymentsWhenStoreFails(t *testing.T) {
	t.Setenv("ROUTER_DEPLOYMENT_CACHE_TTL_SECONDS", "0")

	logger, _ := zap.NewDevelopment()
	memCache, _ := testcache.New(t, logger)
	store := &staticDeploymentStore{
		deployments: []platform.Deployment{
			{
				ID:        "provider-account:openai-1:gpt-4o-mini",
				TenantID:  "tenant-1",
				ModelName: "openai:openai-main:gpt-4o-mini",
				Provider:  "openai",
				Model:     "gpt-4o-mini",
				APIKey:    "tenant-openai-key",
			},
		},
	}

	r, err := newRouterWithStaticDeploymentStore(nil, memCache, store, logger)
	require.NoError(t, err)

	first, err := r.Route(context.Background(), "tenant-1", "openai:openai-main:gpt-4o-mini")
	require.NoError(t, err)
	require.NotNil(t, first)

	store.setError(errors.New("database unavailable"))

	fallback, err := r.Route(context.Background(), "tenant-1", "openai:openai-main:gpt-4o-mini")
	require.NoError(t, err)
	require.NotNil(t, fallback)
	assert.Equal(t, first.ID, fallback.ID)
	assert.Equal(t, "tenant-openai-key", fallback.Params.APIKey)
}

func TestRouter_UsesResolvedLastKnownGoodRouteAfterSanitizedCacheExpires(t *testing.T) {
	t.Setenv("ROUTER_DEPLOYMENT_CACHE_TTL_SECONDS", "30")

	logger, _ := zap.NewDevelopment()
	memCache, _ := testcache.New(t, logger)
	deployment := platform.Deployment{
		ID:        "provider-account:openai-1:gpt-4o-mini",
		TenantID:  "tenant-1",
		ModelName: "openai:openai-main:gpt-4o-mini",
		Provider:  "openai",
		Model:     "gpt-4o-mini",
		APIKey:    "tenant-openai-key",
	}
	store := &staticDeploymentStore{deployments: []platform.Deployment{deployment}}

	sanitized := deployment
	sanitized.APIKey = ""
	raw, err := json.Marshal([]platform.Deployment{sanitized})
	require.NoError(t, err)
	memCache.Set(context.Background(), "deployments:tenant:tenant-1", string(raw), 30*time.Second)

	r, err := newRouterWithStaticDeploymentStore(nil, memCache, store, logger)
	require.NoError(t, err)

	first, err := r.Route(context.Background(), "tenant-1", deployment.ModelName)
	require.NoError(t, err)
	require.Equal(t, "tenant-openai-key", first.Params.APIKey)

	memCache.Delete(context.Background(), "deployments:tenant:tenant-1")
	store.setError(errors.New("database unavailable"))

	fallback, err := r.Route(context.Background(), "tenant-1", deployment.ModelName)
	require.NoError(t, err)
	require.NotNil(t, fallback)
	assert.Equal(t, deployment.ID, fallback.ID)
	assert.Equal(t, "tenant-openai-key", fallback.Params.APIKey)
}

func TestRouter_ReturnsDependencyUnavailableWithoutLastKnownGoodSnapshot(t *testing.T) {
	t.Setenv("ROUTER_DEPLOYMENT_CACHE_TTL_SECONDS", "0")

	logger, _ := zap.NewDevelopment()
	memCache, _ := testcache.New(t, logger)
	store := &staticDeploymentStore{err: errors.New("database unavailable")}

	r, err := newRouterWithStaticDeploymentStore(nil, memCache, store, logger)
	require.NoError(t, err)

	dep, err := r.Route(context.Background(), "tenant-1", "openai:openai-main:gpt-4o-mini")
	assert.Nil(t, dep)
	assert.ErrorIs(t, err, llmrouter.ErrDependencyUnavailable)
}
