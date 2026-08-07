package router

import (
	"context"
	"errors"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type providerAccountAuthResolverStub struct {
	auth *platform.ResolvedProviderAuth
	err  error
}

func (s providerAccountAuthResolverStub) ResolveProviderAccountAuth(
	context.Context,
	string,
	string,
) (*platform.ResolvedProviderAuth, error) {
	return s.auth, s.err
}

func TestNewRouterWithDependenciesRequiresDeclaredProviderAccountResolver(t *testing.T) {
	_, err := NewRouterWithDependencies(
		nil,
		config.RouterSettings{},
		newRouterTestCache(t),
		Dependencies{
			ProviderAccountCredentialPolicy:   ProviderAccountCredentialsFailOpen,
			RequireProviderAccountCredentials: true,
		},
		zap.NewNop(),
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "provider-account credential resolver is required")
}

func TestNewRouterWithDependenciesRejectsMissingFailClosedResolver(t *testing.T) {
	_, err := NewRouterWithDependencies(
		nil,
		config.RouterSettings{},
		newRouterTestCache(t),
		Dependencies{
			ProviderAccountCredentialPolicy: ProviderAccountCredentialsFailClosed,
		},
		zap.NewNop(),
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "provider-account credential resolver is required")
}

func TestRouterProviderAccountCredentialPolicies(t *testing.T) {
	const modelFQN = "openai:primary-openai:gpt-4.1"

	store := deploymentStoreStub{deployments: []platform.Deployment{{
		ID:                  "provider-account:acct_1:gpt-4.1",
		TenantID:            "tenant_a",
		ModelName:           modelFQN,
		Provider:            "openai",
		Model:               "gpt-4.1",
		ProviderAccountID:   "acct_1",
		ProviderAccountName: "primary-openai",
	}}}

	t.Run("explicit resolver", func(t *testing.T) {
		resolver := providerAccountAuthResolverStub{auth: &platform.ResolvedProviderAuth{
			Mode:   string(platform.ProviderAuthModeAPIKey),
			APIKey: "tenant-key",
		}}
		router, err := NewRouterWithDependencies(
			nil,
			config.RouterSettings{},
			newRouterTestCache(t),
			Dependencies{
				DeploymentStore:                 store,
				ProviderAccountAuthResolver:     resolver,
				ProviderAccountCredentialPolicy: ProviderAccountCredentialsFailClosed,
			},
			zap.NewNop(),
		)
		require.NoError(t, err)

		deployment, err := router.Route(t.Context(), "tenant_a", modelFQN)
		require.NoError(t, err)
		require.NotNil(t, deployment)
		assert.Equal(t, "tenant-key", deployment.Params.APIKey)
		assert.Equal(t, platform.CredentialSourceUserBYOK, deployment.CredentialSource)
	})

	t.Run("fail closed", func(t *testing.T) {
		router, err := NewRouterWithDependencies(
			nil,
			config.RouterSettings{},
			newRouterTestCache(t),
			Dependencies{
				DeploymentStore: store,
				ProviderAccountAuthResolver: providerAccountAuthResolverStub{
					err: errors.New("credential backend unavailable"),
				},
				ProviderAccountCredentialPolicy: ProviderAccountCredentialsFailClosed,
			},
			zap.NewNop(),
		)
		require.NoError(t, err)

		deployment, err := router.Route(t.Context(), "tenant_a", modelFQN)
		assert.Nil(t, deployment)
		assert.ErrorIs(t, err, ErrDependencyUnavailable)
	})
}
