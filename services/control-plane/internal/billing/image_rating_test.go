package billing

import (
	"context"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func imageTestInt(value int) *int {
	return &value
}

func imageTestPricing(rates ...platform.PricingRate) *platform.ModelPrice {
	return &platform.ModelPrice{
		Status: "priced",
		Profiles: []platform.PricingProfile{{
			Region:   "global",
			Currency: "USD",
			Rates:    rates,
		}},
	}
}

func imageTestRate(key, usage, unit, amount string, quantity int) platform.PricingRate {
	return platform.PricingRate{
		Key:         key,
		Usage:       usage,
		Unit:        unit,
		Amount:      amount,
		Quantity:    quantity,
		ServiceTier: "standard",
	}
}

func imageTestMeasureAmount(
	t *testing.T,
	measures []platform.UsageMeasure,
	usage string,
	unit string,
) string {
	t.Helper()
	for _, measure := range measures {
		if measure.Usage == usage && measure.Unit == unit {
			return measure.Amount
		}
	}
	t.Fatalf("missing %s/%s measure", usage, unit)
	return ""
}

func imageTestService(t *testing.T) *Service {
	t.Helper()
	t.Setenv("IMAGE_DIMENSIONAL_BILLING_MODE", imageBillingModeEnforce)
	return NewService(nil, nil, 850000, "test")
}

func imageTestDeployment(price *platform.ModelPrice) platform.Deployment {
	return platform.Deployment{
		Model:            "image-model",
		Provider:         "test",
		RegionName:       "global",
		BillingMode:      platform.BillingModeWalletBilled,
		PricingSource:    "public",
		EffectivePricing: price,
	}
}

func TestBuildImageUsageMeasuresNormalizesProviderDetails(t *testing.T) {
	price := imageTestPricing(
		imageTestRate("text-in", "input", "token", "5", 1_000_000),
		imageTestRate("image-in", "input", "image_token", "8", 1_000_000),
		imageTestRate("image-out", "output", "image_token", "30", 1_000_000),
	)
	usage := platform.ImageUsage{
		InputTokens:      50,
		OutputTokens:     50,
		TotalTokens:      100,
		InputTextTokens:  imageTestInt(10),
		InputImageTokens: imageTestInt(40),
		OutputImageCount: 1,
		Quality:          "high",
	}

	measures, err := buildImageUsageMeasures(price, "global", usage, imageTestNow())
	require.NoError(t, err)
	require.Len(t, measures, 3)
	assert.Equal(t, "10", imageTestMeasureAmount(t, measures, "input", "token"))
	assert.Equal(t, "40", imageTestMeasureAmount(t, measures, "input", "image_token"))
	assert.Equal(t, "50", imageTestMeasureAmount(t, measures, "output", "image_token"))
	for _, measure := range measures {
		assert.Equal(t, map[string]interface{}{"quality": "high"}, measure.Qualifiers)
	}
}

func TestEstimateImageUsageRatesThreeDimensionsExactly(t *testing.T) {
	svc := imageTestService(t)
	price := imageTestPricing(
		imageTestRate("text-in", "input", "token", "5", 1_000_000),
		imageTestRate("image-in", "input", "image_token", "8", 1_000_000),
		imageTestRate("image-out", "output", "image_token", "30", 1_000_000),
	)
	usage := platform.ImageUsage{
		InputTokens:      300,
		OutputTokens:     1000,
		TotalTokens:      1300,
		InputTextTokens:  imageTestInt(100),
		InputImageTokens: imageTestInt(200),
		OutputImageCount: 1,
		Source:           platform.ImageUsageSourceEstimated,
	}

	rated, err := svc.EstimateImageUsage(
		context.Background(),
		imageTestDeployment(price),
		"gpt-image",
		usage,
	)
	require.NoError(t, err)
	assert.Equal(t, "0.0321", rated.TotalCostUSDExact)
	assert.InDelta(t, 0.0321, rated.TotalCostUSD, 1e-12)
	require.Len(t, rated.AppliedRates, 3)
}

func TestEstimateImageUsageRatesPerImageCount(t *testing.T) {
	svc := imageTestService(t)
	price := imageTestPricing(
		imageTestRate("per-image", "output", "image", "0.02", 1),
	)

	rated, err := svc.EstimateImageUsage(
		context.Background(),
		imageTestDeployment(price),
		"xai-image",
		platform.ImageUsage{
			OutputTokens:     1000,
			TotalTokens:      1000,
			OutputImageCount: 2,
		},
	)
	require.NoError(t, err)
	assert.Equal(t, "0.04", rated.TotalCostUSDExact)
	require.Len(t, rated.AppliedRates, 1)
	assert.Equal(t, "image", rated.AppliedRates[0].Unit)
	assert.Equal(t, "2", rated.AppliedRates[0].UsageAmount)
}

func TestEstimateImageUsagePrefersExactResolutionOverGeneric(t *testing.T) {
	svc := imageTestService(t)
	generic := imageTestRate("generic", "output", "image", "0.134", 1)
	exact := imageTestRate("4k", "output", "image", "0.24", 1)
	exact.Resolution = "4k"
	price := imageTestPricing(generic, exact)

	rated, err := svc.EstimateImageUsage(
		context.Background(),
		imageTestDeployment(price),
		"image-model",
		platform.ImageUsage{
			OutputImageCount: 2,
			Resolution:       "4k",
		},
	)
	require.NoError(t, err)
	assert.Equal(t, "0.48", rated.TotalCostUSDExact)
	require.Len(t, rated.AppliedRates, 1)
	assert.Equal(t, "4k", rated.AppliedRates[0].RateKey)
}

func TestEstimateImageUsageDoesNotDoubleChargeOutputStrategies(t *testing.T) {
	svc := imageTestService(t)
	price := imageTestPricing(
		imageTestRate("per-image", "output", "image", "0.02", 1),
		imageTestRate("per-image-token", "output", "image_token", "30", 1_000_000),
		imageTestRate("legacy-token", "output", "token", "40", 1_000_000),
	)

	rated, err := svc.EstimateImageUsage(
		context.Background(),
		imageTestDeployment(price),
		"image-model",
		platform.ImageUsage{
			OutputTokens:     1000,
			TotalTokens:      1000,
			OutputImageCount: 2,
		},
	)
	require.NoError(t, err)
	assert.Equal(t, "0.04", rated.TotalCostUSDExact)
	require.Len(t, rated.AppliedRates, 1)
	assert.Equal(t, "per-image", rated.AppliedRates[0].RateKey)
}

func TestEstimateImageUsageSelectsQualityQualifier(t *testing.T) {
	svc := imageTestService(t)
	generic := imageTestRate("generic", "output", "image", "0.10", 1)
	high := imageTestRate("high", "output", "image", "0.20", 1)
	high.Qualifiers = map[string]interface{}{"quality": "high"}
	price := imageTestPricing(generic, high)

	rated, err := svc.EstimateImageUsage(
		context.Background(),
		imageTestDeployment(price),
		"private-image-model",
		platform.ImageUsage{
			OutputImageCount: 1,
			Quality:          "HIGH",
		},
	)
	require.NoError(t, err)
	assert.Equal(t, "0.2", rated.TotalCostUSDExact)
	require.Len(t, rated.AppliedRates, 1)
	assert.Equal(t, "high", rated.AppliedRates[0].RateKey)
}

func TestEstimateImageUsageMapsAggregateToOnlyAvailableInputUnit(t *testing.T) {
	svc := imageTestService(t)
	price := imageTestPricing(
		imageTestRate("image-in", "input", "image_token", "8", 1_000_000),
		imageTestRate("image-out", "output", "image_token", "30", 1_000_000),
	)

	rated, err := svc.EstimateImageUsage(
		context.Background(),
		imageTestDeployment(price),
		"image-model",
		platform.ImageUsage{
			InputTokens:      50,
			OutputTokens:     1,
			TotalTokens:      51,
			OutputImageCount: 1,
		},
	)
	require.NoError(t, err)
	assert.Equal(t, "50", imageTestMeasureAmount(
		t,
		rated.UsageMeasures,
		"input",
		"image_token",
	))
	assert.Empty(t, imageTestMeasureAmountOrEmpty(
		rated.UsageMeasures,
		"input",
		"token",
	))
}

func imageTestMeasureAmountOrEmpty(
	measures []platform.UsageMeasure,
	usage string,
	unit string,
) string {
	for _, measure := range measures {
		if measure.Usage == usage && measure.Unit == unit {
			return measure.Amount
		}
	}
	return ""
}

func TestFinalizeImageUsageSplitsMissingInputDetailsFromEstimate(t *testing.T) {
	svc := imageTestService(t)
	price := imageTestPricing(
		imageTestRate("text-in", "input", "token", "1", 1),
		imageTestRate("image-in", "input", "image_token", "1", 1),
		imageTestRate("image-out", "output", "image_token", "1", 1),
	)
	estimateUsage := platform.ImageUsage{
		InputTokens:      300,
		OutputTokens:     1,
		TotalTokens:      301,
		InputTextTokens:  imageTestInt(100),
		InputImageTokens: imageTestInt(200),
		OutputImageCount: 1,
	}
	estimated, err := svc.EstimateImageUsage(
		context.Background(),
		imageTestDeployment(price),
		"image-model",
		estimateUsage,
	)
	require.NoError(t, err)

	finalized, err := svc.FinalizeImageUsage(
		context.Background(),
		"image-model",
		platform.ImageUsage{
			InputTokens:      10,
			OutputTokens:     1,
			TotalTokens:      11,
			OutputImageCount: 1,
			Source:           platform.ImageUsageSourceProviderUsage,
		},
		estimated,
	)
	require.NoError(t, err)
	require.Len(t, finalized.UsageMeasures, 3)
	assert.Equal(t, "4", imageTestMeasureAmount(t, finalized.UsageMeasures, "input", "token"))
	assert.Equal(t, "6", imageTestMeasureAmount(t, finalized.UsageMeasures, "input", "image_token"))
	assert.Equal(t, "1", imageTestMeasureAmount(t, finalized.UsageMeasures, "output", "image_token"))
}

func TestFinalizeImageUsageUsesExactProviderTicksAndSingleComponent(t *testing.T) {
	svc := imageTestService(t)
	price := imageTestPricing(
		imageTestRate("per-image", "output", "image", "0.03", 1),
	)
	estimated, err := svc.EstimateImageUsage(
		context.Background(),
		imageTestDeployment(price),
		"xai-image",
		platform.ImageUsage{OutputImageCount: 1},
	)
	require.NoError(t, err)

	ticks := int64(200_000_000)
	finalized, err := svc.FinalizeImageUsage(
		context.Background(),
		"xai-image",
		platform.ImageUsage{
			OutputImageCount:     1,
			ProviderCostUSDTicks: &ticks,
			Source:               platform.ImageUsageSourceProviderReported,
		},
		estimated,
	)
	require.NoError(t, err)
	assert.Equal(t, "0.02", finalized.TotalCostUSDExact)
	assert.InDelta(t, 0.02, finalized.OutputCostUSD, 1e-12)
	assert.Zero(t, finalized.InputCostUSD)
	assert.Equal(t, "provider_reported", finalized.RatingStatus)
	require.Len(t, finalized.AppliedRates, 1)
	assert.Equal(t, "provider-reported:request", finalized.AppliedRates[0].RateKey)
	assert.Equal(t, finalized.TotalCostUSDExact, finalized.AppliedRates[0].ComponentCostUSD)
}

func TestFinalizeImageUsageFallsBackToReservedEstimate(t *testing.T) {
	svc := imageTestService(t)
	price := imageTestPricing(
		imageTestRate("image-out", "output", "image_token", "30", 1_000_000),
	)
	estimated, err := svc.EstimateImageUsage(
		context.Background(),
		imageTestDeployment(price),
		"image-model",
		platform.ImageUsage{
			OutputTokens:     1000,
			TotalTokens:      1000,
			OutputImageCount: 1,
			Source:           platform.ImageUsageSourceEstimated,
		},
	)
	require.NoError(t, err)

	finalized, err := svc.FinalizeImageUsage(
		context.Background(),
		"image-model",
		platform.ImageUsage{
			OutputImageCount: 1,
			Source:           platform.ImageUsageSourceProviderUsage,
		},
		estimated,
	)
	require.NoError(t, err)
	assert.Equal(t, "estimated_final", finalized.RatingStatus)
	assert.Equal(t, estimated.TotalCostUSDExact, finalized.TotalCostUSDExact)
	assert.Equal(t, "actual_usage_rate_unresolved", finalized.Metadata["image_billing_fallback_reason"])
}

func TestEstimateImageUsageAllowsUnratedExternalBilling(t *testing.T) {
	svc := imageTestService(t)
	deployment := imageTestDeployment(imageTestPricing(
		imageTestRate("text-only", "input", "token", "5", 1_000_000),
	))
	deployment.BillingMode = platform.BillingModeExternalBilled

	rated, err := svc.EstimateImageUsage(
		context.Background(),
		deployment,
		"external-image-model",
		platform.ImageUsage{OutputImageCount: 1},
	)
	require.NoError(t, err)
	assert.Equal(t, "unrated", rated.RatingStatus)
	assert.Zero(t, rated.TotalCostUSD)
}

func imageTestNow() time.Time {
	return time.Date(2026, time.July, 28, 0, 0, 0, 0, time.UTC)
}
