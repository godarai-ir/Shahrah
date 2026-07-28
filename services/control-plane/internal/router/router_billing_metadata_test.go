package router

import (
	"context"
	"errors"
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
		CredentialSource:    platform.CredentialSourceUserBYOK,
		BillingMode:         platform.BillingModeExternalBilled,
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
		CredentialSource:    platform.CredentialSourceGodarManagedKey,
		BillingMode:         platform.BillingModeWalletBilled,
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

func TestRouterRejectsDynamicDeploymentWithoutBillingMetadata(t *testing.T) {
	store := deploymentStoreStub{deployments: []platform.Deployment{{
		ID:        "legacy-deployment",
		TenantID:  "tenant_a",
		ModelName: "legacy-model",
		Provider:  "openai",
		Model:     "gpt-4o-mini",
		APIKey:    "tenant-key",
	}}}
	rt, err := NewRouter(nil, config.RouterSettings{CooldownTime: 60, NumRetries: 1}, newRouterTestCache(t), store, zap.NewNop())
	require.NoError(t, err)

	_, err = rt.Route(context.Background(), "tenant_a", "legacy-model")
	require.ErrorIs(t, err, ErrDeploymentBillingMetadataUnresolved)
}

func TestRouterResolvesStaticWalletPricingAndUsesLastKnownGood(t *testing.T) {
	rt, err := NewRouter([]config.ModelDeployment{{
		ModelName: "gpt-4o-mini",
		ProviderEndpoint: config.ProviderEndpointConfig{
			Provider: "openai",
			Model:    "gpt-4o-mini",
			APIKey:   "platform-key",
		},
	}}, config.RouterSettings{CooldownTime: 60, NumRetries: 1}, newRouterTestCache(t), nil, zap.NewNop())
	require.NoError(t, err)

	price := &platform.ModelPrice{Profiles: []platform.PricingProfile{{
		Region:   "global",
		Currency: "USD",
		Rates: []platform.PricingRate{{
			Key:      "input",
			Usage:    "input_tokens",
			Unit:     "token",
			Amount:   "0.01",
			Quantity: 1000,
		}},
	}}}
	resolveCalls := 0
	rt.SetPricingResolver(func(_ context.Context, provider, model string) (*platform.ModelPrice, string, error) {
		resolveCalls++
		assert.Equal(t, "openai", provider)
		assert.Equal(t, "gpt-4o-mini", model)
		if resolveCalls > 1 {
			return nil, "", errors.New("catalog unavailable")
		}
		return price, "public", nil
	})

	first, err := rt.Route(context.Background(), "tenant_a", "gpt-4o-mini")
	require.NoError(t, err)
	require.NotNil(t, first.EffectivePricing)
	assert.Equal(t, "public", first.PricingSource)

	second, err := rt.Route(context.Background(), "tenant_a", "gpt-4o-mini")
	require.NoError(t, err)
	require.NotNil(t, second.EffectivePricing)
	assert.Equal(t, "public", second.PricingSource)
	assert.Equal(t, 2, resolveCalls)
}

func TestRouterResolvesStaticWildcardPricingForConcreteModel(t *testing.T) {
	rt, err := NewRouter([]config.ModelDeployment{{
		ModelName: "*",
		ProviderEndpoint: config.ProviderEndpointConfig{
			Provider: "openai",
			Model:    "openai/*",
			APIKey:   "platform-key",
		},
	}}, config.RouterSettings{CooldownTime: 60, NumRetries: 1}, newRouterTestCache(t), nil, zap.NewNop())
	require.NoError(t, err)

	price := &platform.ModelPrice{Profiles: []platform.PricingProfile{{
		Region:   "global",
		Currency: "USD",
		Rates: []platform.PricingRate{{
			Key:      "input",
			Usage:    "input_tokens",
			Unit:     "token",
			Amount:   "0.01",
			Quantity: 1000,
		}},
	}}}
	rt.SetPricingResolver(func(_ context.Context, provider, model string) (*platform.ModelPrice, string, error) {
		assert.Equal(t, "openai", provider)
		assert.Equal(t, "gpt-4o-mini", model)
		return price, "public", nil
	})

	deployment, err := rt.Route(context.Background(), "tenant_a", "openai/gpt-4o-mini")
	require.NoError(t, err)
	require.NotNil(t, deployment.EffectivePricing)
	assert.Equal(t, "public", deployment.PricingSource)
}
