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
