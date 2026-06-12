package router_test

import (
	"context"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
	"github.com/Delpak/shahrah/services/control-plane/internal/config"
	llmrouter "github.com/Delpak/shahrah/services/control-plane/internal/router"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// TestRouter_ShouldSelectDeployment verifies that the router selects a healthy deployment.
func TestRouter_ShouldSelectDeployment(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)

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
	assert.NoError(t, err)
	assert.NotNil(t, r)

	dep, err := r.Route(context.Background(), "", "gpt-3.5-turbo")
	assert.NoError(t, err)
	assert.NotNil(t, dep)
	assert.Equal(t, "gpt-3.5-turbo", dep.ModelName)
}

// TestRouter_ShouldReturnError_WhenModelNotFound verifies that the router returns an error for unknown models.
func TestRouter_ShouldReturnError_WhenModelNotFound(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)

	modelList := []config.ModelDeployment{}
	settings := config.RouterSettings{}

	r, err := llmrouter.NewRouter(modelList, settings, memCache, nil, logger)
	assert.NoError(t, err)

	dep, err := r.Route(context.Background(), "", "unknown-model")
	assert.Error(t, err)
	assert.Nil(t, dep)
	assert.Equal(t, llmrouter.ErrModelNotFound, err)
}

// TestRouter_ShouldExcludeCooldownedDeployment verifies that cooldowned deployments are excluded from routing.
func TestRouter_ShouldExcludeCooldownedDeployment(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)

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
	assert.NoError(t, err)

	// First, select the deployment
	dep, err := r.Route(context.Background(), "", "test-model")
	assert.NoError(t, err)
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
	memCache, _ := cache.NewMemoryCache(logger)

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
	assert.NoError(t, err)

	// Run 1000 selections, verify distribution roughly matches weights
	highCount := 0
	lowCount := 0
	iterations := 1000

	for i := 0; i < iterations; i++ {
		dep, err := r.Route(context.Background(), "", "gpt-3.5-turbo")
		assert.NoError(t, err)
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
	memCache, _ := cache.NewMemoryCache(logger)

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
	assert.NoError(t, err)

	dep, err := r.Route(context.Background(), "", "anthropic/claude-3-5-sonnet-latest")
	assert.NoError(t, err)
	assert.NotNil(t, dep)
	assert.Equal(t, "dep-anthropic", dep.ID)
	assert.Equal(t, "anthropic", dep.Provider)
	assert.Equal(t, "*", dep.Model)
}

type staticDeploymentStore struct {
	deployments []platform.Deployment
}

func (s *staticDeploymentStore) ListDeployments(_ context.Context, _ string) ([]platform.Deployment, error) {
	return append([]platform.Deployment(nil), s.deployments...), nil
}

func TestRouter_FQNShouldNotFallBackToGlobalWildcard(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)

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

	r, err := llmrouter.NewRouter(modelList, config.RouterSettings{}, memCache, store, logger)
	assert.NoError(t, err)

	dep, err := r.Route(context.Background(), "tenant-1", "anthropic:anthropic-sunny-zephyr:claude-3-haiku-20240307")
	assert.NoError(t, err)
	assert.NotNil(t, dep)
	assert.Equal(t, "provider-account:anthropic-1:anthropic/claude-3-haiku-20240307", dep.ID)
	assert.Equal(t, "anthropic", dep.Provider)
}

func TestRouter_FQNWithoutExactDeploymentShouldNotUseWildcard(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)

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

	r, err := llmrouter.NewRouter(modelList, config.RouterSettings{}, memCache, &staticDeploymentStore{}, logger)
	assert.NoError(t, err)

	dep, err := r.Route(context.Background(), "tenant-1", "anthropic:anthropic-sunny-zephyr:claude-3-haiku-20240307")
	assert.Error(t, err)
	assert.Nil(t, dep)
	assert.Equal(t, llmrouter.ErrModelNotFound, err)
}
