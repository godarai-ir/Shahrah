package router

import (
	"context"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
	"github.com/Delpak/shahrah/services/control-plane/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type deploymentStoreStub struct {
	deployments []platform.Deployment
}

func (s deploymentStoreStub) ListDeployments(_ context.Context, _ string) ([]platform.Deployment, error) {
	return append([]platform.Deployment(nil), s.deployments...), nil
}

func newRouterTestCache(t *testing.T) cache.Cache {
	t.Helper()

	mem, err := cache.NewMemoryCache(zap.NewNop())
	require.NoError(t, err)

	return mem
}

func TestRouterProviderAccountWithBYOKMarksExternalBilled(t *testing.T) {
	store := deploymentStoreStub{deployments: []platform.Deployment{{
		ID:                  "provider-account:acct_1:gpt-4o-mini",
		TenantID:            "tenant_a",
		ModelName:           "openai:personal-openai:gpt-4o-mini",
		Provider:            "openai",
		Model:               "gpt-4o-mini",
		APIKey:              "user-openai-key",
		ProviderAccountID:   "acct_1",
		ProviderAccountName: "personal-openai",
	}}}
	rt, err := NewRouter(nil, config.RouterSettings{CooldownTime: 60, NumRetries: 1}, newRouterTestCache(t), store, zap.NewNop())
	require.NoError(t, err)

	dep, err := rt.Route(context.Background(), "tenant_a", "openai:personal-openai:gpt-4o-mini")
	require.NoError(t, err)

	assert.Equal(t, "user-openai-key", dep.Params.APIKey)
	assert.Equal(t, platform.CredentialSourceUserBYOK, dep.CredentialSource)
	assert.Equal(t, platform.BillingModeExternalBilled, dep.BillingMode)
	assert.Equal(t, "acct_1", dep.ProviderAccountID)
	assert.Equal(t, "personal-openai", dep.ProviderAccountName)
}

func TestRouterProviderAccountWithoutBYOKUsesSharedProviderKeyAndWalletBilling(t *testing.T) {
	store := deploymentStoreStub{deployments: []platform.Deployment{{
		ID:                  "provider-account:acct_1:gpt-4o-mini",
		TenantID:            "tenant_a",
		ModelName:           "openai:shared-openai:gpt-4o-mini",
		Provider:            "openai",
		Model:               "gpt-4o-mini",
		ProviderAccountID:   "acct_1",
		ProviderAccountName: "shared-openai",
	}}}
	rt, err := NewRouter([]config.ModelDeployment{{
		ModelName: "gpt-4o-mini",
		ProviderEndpoint: config.ProviderEndpointConfig{
			Provider: "openai",
			Model:    "gpt-4o-mini",
			APIKey:   "godar-shared-openai-key",
		},
	}}, config.RouterSettings{CooldownTime: 60, NumRetries: 1}, newRouterTestCache(t), store, zap.NewNop())
	require.NoError(t, err)

	dep, err := rt.Route(context.Background(), "tenant_a", "openai:shared-openai:gpt-4o-mini")
	require.NoError(t, err)

	assert.Equal(t, "godar-shared-openai-key", dep.Params.APIKey)
	assert.Equal(t, platform.CredentialSourceGodarManagedKey, dep.CredentialSource)
	assert.Equal(t, platform.BillingModeWalletBilled, dep.BillingMode)
	assert.Equal(t, "acct_1", dep.ProviderAccountID)
	assert.Equal(t, "shared-openai", dep.ProviderAccountName)
}

func TestRouterNoCredentialMockBypassesPlatformCredentialResolution(t *testing.T) {
	store := deploymentStoreStub{deployments: []platform.Deployment{{
		ID:                  "provider-account:acct_mock:mock/chat",
		TenantID:            "tenant_a",
		ModelName:           "mock:loadtest:chat",
		Provider:            "mock",
		Model:               "mock/chat",
		CredentialSource:    platform.CredentialSourceGodarManagedKey,
		BillingMode:         platform.BillingModeExternalBilled,
		ProviderAccountID:   "acct_mock",
		ProviderAccountName: "loadtest",
		ProviderAuth: &platform.ResolvedProviderAuth{
			Mode: string(platform.ProviderAuthModeNoCredential),
		},
	}}}
	rt, err := NewRouter(nil, config.RouterSettings{CooldownTime: 60, NumRetries: 1}, newRouterTestCache(t), store, zap.NewNop())
	require.NoError(t, err)
	rt.SetPlatformCredentialResolver(nil, platform.ProviderCredentialModeStore, platform.ProviderCredentialEnvironmentProd)

	dep, err := rt.Route(context.Background(), "tenant_a", "mock:loadtest:chat")
	require.NoError(t, err)

	assert.Empty(t, dep.Params.APIKey)
	assert.Equal(t, string(platform.ProviderAuthModeNoCredential), dep.ProviderAuth.Mode)
	assert.Equal(t, platform.BillingModeExternalBilled, dep.BillingMode)
	assert.Equal(t, "acct_mock", dep.ProviderAccountID)
}

func TestRouterParsesFirstClassProviderModelsThroughRegistry(t *testing.T) {
	tests := []struct {
		qualifiedModel string
		provider       string
		officialModel  string
	}{
		{qualifiedModel: "openai/gpt-5", provider: "openai", officialModel: "gpt-5"},
		{qualifiedModel: "anthropic/claude-opus-5", provider: "anthropic", officialModel: "claude-opus-5"},
		{qualifiedModel: "xai/grok-4.5", provider: "xai", officialModel: "grok-4.5"},
		{qualifiedModel: "deepseek/deepseek-v4-flash", provider: "deepseek", officialModel: "deepseek-v4-flash"},
		{qualifiedModel: "moonshot/kimi-k3", provider: "moonshot", officialModel: "kimi-k3"},
		{qualifiedModel: "gemini/gemini-3.5-flash", provider: "google-gemini", officialModel: "gemini-3.5-flash"},
		{qualifiedModel: "mistral/mistral-large-2607", provider: "mistral-ai", officialModel: "mistral-large-2607"},
	}

	models := make([]config.ModelDeployment, 0, len(tests))
	for _, test := range tests {
		models = append(models, config.ModelDeployment{
			ModelName: test.qualifiedModel,
			ProviderEndpoint: config.ProviderEndpointConfig{
				Model:  test.qualifiedModel,
				APIKey: "provider-key",
				RPM:    1000,
			},
			ModelInfo: map[string]interface{}{"id": test.qualifiedModel},
		})
	}

	rt, err := NewRouter(
		models,
		config.RouterSettings{CooldownTime: 60, NumRetries: 1},
		newRouterTestCache(t),
		nil,
		zap.NewNop(),
	)
	require.NoError(t, err)

	for _, test := range tests {
		t.Run(test.provider, func(t *testing.T) {
			deployment, err := rt.Route(
				context.Background(),
				"tenant_a",
				test.qualifiedModel,
			)
			require.NoError(t, err)
			assert.Equal(t, test.provider, deployment.Provider)
			assert.Equal(t, test.officialModel, deployment.Model)
		})
	}
}
