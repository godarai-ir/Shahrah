package admin

import (
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLegacyDeploymentBillingMetadata(t *testing.T) {
	t.Parallel()

	t.Run("tenant api key is BYOK and externally billed", func(t *testing.T) {
		credentialSource, billingMode, err := legacyDeploymentBillingMetadata("openai", "tenant-key")
		require.NoError(t, err)
		assert.Equal(t, platform.CredentialSourceUserBYOK, credentialSource)
		assert.Equal(t, platform.BillingModeExternalBilled, billingMode)
	})

	t.Run("mock needs no credential and is externally billed", func(t *testing.T) {
		credentialSource, billingMode, err := legacyDeploymentBillingMetadata(" Mock ", "")
		require.NoError(t, err)
		assert.Equal(t, platform.CredentialSourceGodarManagedKey, credentialSource)
		assert.Equal(t, platform.BillingModeExternalBilled, billingMode)
	})

	t.Run("non mock without credential is rejected", func(t *testing.T) {
		credentialSource, billingMode, err := legacyDeploymentBillingMetadata("anthropic", " ")
		require.Error(t, err)
		assert.Empty(t, credentialSource)
		assert.Empty(t, billingMode)
	})
}

func TestValidateCreateVirtualModelRequestAllowsWeighted(t *testing.T) {
	req := platform.CreateVirtualModelRequest{
		Alias:           "smart-router",
		RoutingStrategy: "weighted",
		Targets: []platform.VirtualModelTarget{
			{Order: 1, Weight: 60, Provider: "openai", Model: "gpt-4o", Policy: platform.VirtualModelPolicy{}},
			{Order: 2, Weight: 40, Provider: "anthropic", Model: "claude-sonnet", Policy: platform.VirtualModelPolicy{}},
		},
	}

	require.NoError(t, validateCreateVirtualModelRequest(req))
}

func TestValidateCreateVirtualModelRequestRejectsInvalidWeightSum(t *testing.T) {
	req := platform.CreateVirtualModelRequest{
		Alias:           "smart-router",
		RoutingStrategy: "weighted",
		Targets: []platform.VirtualModelTarget{
			{Order: 1, Weight: 70, Provider: "openai", Model: "gpt-4o", Policy: platform.VirtualModelPolicy{}},
			{Order: 2, Weight: 20, Provider: "anthropic", Model: "claude-sonnet", Policy: platform.VirtualModelPolicy{}},
		},
	}

	err := validateCreateVirtualModelRequest(req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "sum to 100")
}

func TestValidateUpdateVirtualModelRequestRejectsUnknownStrategy(t *testing.T) {
	strategy := "latency_based"
	err := validateUpdateVirtualModelRequest(platform.UpdateVirtualModelRequest{
		RoutingStrategy: &strategy,
	}, "priority")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "priority, weighted")
}
