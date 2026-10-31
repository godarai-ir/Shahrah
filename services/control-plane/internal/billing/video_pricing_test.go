package billing

import (
	"context"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestVideoPricingRequiresAnUnconditionalPerSecondRate(t *testing.T) {
	t.Parallel()

	rate := normalizedPricingRate{UnitPriceUSD: decimal.RequireFromString("0.05")}
	require.True(t, videoPricingDimensionsComplete(map[string]normalizedPricingRate{VideoOutputSecondsDimension: rate}))
	require.False(t, videoPricingDimensionsComplete(map[string]normalizedPricingRate{"output_images": rate}))

	free := normalizedPricingRate{UnitPriceUSD: decimal.Zero}
	require.False(t, videoPricingDimensionsComplete(map[string]normalizedPricingRate{VideoOutputSecondsDimension: free}))

	conditional := rate
	conditional.Condition = map[string]interface{}{"resolution": "1080p"}
	require.False(t, videoPricingDimensionsComplete(map[string]normalizedPricingRate{VideoOutputSecondsDimension: conditional}))
}

func TestPriceVideoUsageRejectsUnknownDuration(t *testing.T) {
	t.Parallel()

	_, err := (&PricingEngine{}).PriceVideoUsage(context.Background(), "xai:acct:grok-imagine-video", 0, nil, platform.UsageCostBreakdown{})
	require.ErrorIs(t, err, ErrPricingUnavailable)
}

func TestXAIMeteredUsageFailsClosed(t *testing.T) {
	t.Parallel()

	require.True(t, xAIPricingStrategy{}.MeteredUsageFailsClosed())
}

func TestReconcileVideoChargeUsesProviderTicks(t *testing.T) {
	t.Parallel()

	catalog := platform.UsageCostBreakdown{TotalCostUSD: 0.40}

	settled, err := reconcileXAIProviderCost(catalog, 1_800_000_000, 1)
	require.NoError(t, err)
	require.InDelta(t, 0.18, settled.TotalCostUSD, 1e-12)
}
