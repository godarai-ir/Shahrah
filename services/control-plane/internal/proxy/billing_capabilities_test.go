package proxy

import (
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBillingCapabilitiesValidateFailClosedDependencies(t *testing.T) {
	tests := []struct {
		name         string
		capabilities BillingCapabilities
		errorText    string
	}{
		{
			name: "deployment pricing",
			capabilities: BillingCapabilities{
				RequireDeploymentPricing: true,
			},
			errorText: "deployment pricing capability is required",
		},
		{
			name: "metered usage",
			capabilities: BillingCapabilities{
				MeteredUsagePolicy: CapabilityFailClosed,
			},
			errorText: "metered usage pricing capability is required",
		},
		{
			name: "budget reservations",
			capabilities: BillingCapabilities{
				BudgetReservationPolicy: CapabilityFailClosed,
			},
			errorText: "budget reservations capability is required",
		},
		{
			name: "batch finalization",
			capabilities: BillingCapabilities{
				BatchUsagePolicy: CapabilityFailClosed,
			},
			errorText: "batch usage finalization capability is required",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.capabilities.Validate()
			require.Error(t, err)
			assert.Contains(t, err.Error(), test.errorText)
		})
	}
}

func TestRequestPipelineMeteredUsagePolicies(t *testing.T) {
	estimated := platform.UsageCostBreakdown{Model: "fallback", TotalCostIRR: 100}

	failOpen := NewRequestPipeline(nil, nil, &testBudgeter{}, BillingCapabilities{
		MeteredUsagePolicy: CapabilityFailOpen,
	}, nil)
	actual, err := failOpen.priceMeteredUsage(t.Context(), "mistral-small", "input_tokens", "1", estimated)
	require.NoError(t, err)
	assert.Equal(t, estimated, actual)

	failClosed := NewRequestPipeline(nil, nil, &testBudgeter{}, BillingCapabilities{
		MeteredUsagePolicy: CapabilityFailClosed,
	}, nil)
	_, err = failClosed.priceMeteredUsage(t.Context(), "mistral-small", "input_tokens", "1", estimated)
	assert.ErrorIs(t, err, errMeteredUsagePricingUnavailable)

	pricer := &testBudgeter{}
	disabled := NewRequestPipeline(nil, nil, pricer, BillingCapabilities{
		MeteredUsage:       pricer,
		MeteredUsagePolicy: CapabilityDisabled,
	}, nil)
	actual, err = disabled.priceMeteredUsage(t.Context(), "mistral-small", "input_tokens", "1", estimated)
	require.NoError(t, err)
	assert.Equal(t, estimated, actual)

	explicit := NewRequestPipeline(nil, nil, pricer, BillingCapabilities{
		MeteredUsage:       pricer,
		MeteredUsagePolicy: CapabilityFailClosed,
	}, nil)
	actual, err = explicit.priceMeteredUsage(t.Context(), "mistral-small", "input_tokens", "1", estimated)
	require.NoError(t, err)
	assert.NotEmpty(t, actual.PricingRevision)
}

func TestRequestPipelineRejectsMissingFailClosedReservationManager(t *testing.T) {
	pipeline := NewRequestPipeline(nil, nil, &testBudgeter{}, BillingCapabilities{
		BudgetReservationPolicy: CapabilityFailClosed,
	}, nil)

	reservation, err := pipeline.Reserve(t.Context(), BudgetRequest{
		Auth:      &platform.AuthContext{TenantID: "tenant-1"},
		Estimated: platform.UsageCostBreakdown{TotalCostIRR: 100},
		RequestID: "request-1",
	})

	assert.Nil(t, reservation)
	assert.ErrorIs(t, err, errBudgetReservationUnavailable)
}
