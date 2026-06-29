package admin

import (
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
