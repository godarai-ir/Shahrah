package billing

import (
	"errors"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestRateUsageSelectsExactVideoResolution(t *testing.T) {
	rated, err := RateUsage(RateInput{
		Model: "video-model", Provider: "provider", Region: "us-west-2", Pricing: videoPricing(),
		Measures:     []platform.UsageMeasure{{Usage: "output", Unit: "second", Amount: "12", Resolution: "720p", ServiceTier: "standard"}},
		ExchangeRate: decimal.NewFromInt(850000),
	})
	require.NoError(t, err)
	require.Equal(t, "rated", rated.Status)
	require.Equal(t, "18", rated.TotalCostUSD)
	require.Equal(t, "15300000", rated.TotalCostIRR)
	require.Equal(t, "video-720p", rated.Components[0].RateKey)
	require.Equal(t, "us-west-2|USD||", rated.Components[0].ProfileKey)
}

func TestRateUsageFallsBackToGenericResolution(t *testing.T) {
	pricing := &platform.ModelPrice{Status: "priced", Profiles: []platform.PricingProfile{{
		Region: "global", Currency: "USD", Rates: []platform.PricingRate{{
			Key: "video-generic", Usage: "output", Unit: "second", Amount: "0.08", Quantity: 1, ServiceTier: "standard",
		}},
	}}}
	rated, err := RateUsage(RateInput{
		Pricing: pricing, Region: "eu-west-1", ExchangeRate: decimal.NewFromInt(1),
		Measures: []platform.UsageMeasure{{Usage: "output", Unit: "second", Amount: "10", Resolution: "1080p"}},
	})
	require.NoError(t, err)
	require.Equal(t, "0.8", rated.TotalCostUSD)
}

func TestRateUsageFallsBackToGenericServiceTier(t *testing.T) {
	pricing := &platform.ModelPrice{Status: "priced", Profiles: []platform.PricingProfile{{
		Region: "global", Currency: "USD", Rates: []platform.PricingRate{
			{Key: "generic", Usage: "output", Unit: "second", Amount: "1", Quantity: 1},
			{Key: "standard", Usage: "output", Unit: "second", Amount: "0.5", Quantity: 1, ServiceTier: "standard"},
		},
	}}}
	rated, err := RateUsage(RateInput{
		Pricing: pricing, ExchangeRate: decimal.NewFromInt(1),
		Measures: []platform.UsageMeasure{{Usage: "output", Unit: "second", Amount: "2", ServiceTier: "priority"}},
	})
	require.NoError(t, err)
	require.Equal(t, "generic", rated.Components[0].RateKey)
	require.Equal(t, "2", rated.TotalCostUSD)
}

func TestRateUsageDoesNotChooseArbitraryRegion(t *testing.T) {
	pricing := videoPricing()
	pricing.Profiles = append(pricing.Profiles, platform.PricingProfile{Region: "eu-west-1", Currency: "USD"})
	_, err := RateUsage(RateInput{
		Pricing: pricing, Region: "ap-south-1", ExchangeRate: decimal.NewFromInt(1),
		Measures: []platform.UsageMeasure{{Usage: "output", Unit: "second", Amount: "1", Resolution: "540p"}},
	})
	require.ErrorIs(t, err, ErrPricingUnresolved)
}

func TestRateUsageRejectsAmbiguousProfilesInsteadOfChoosingFirst(t *testing.T) {
	pricing := videoPricing()
	pricing.Profiles = append(pricing.Profiles, pricing.Profiles[0])
	_, err := RateUsage(RateInput{
		Pricing: pricing, Region: "us-west-2", ExchangeRate: decimal.NewFromInt(1),
		Measures: []platform.UsageMeasure{{Usage: "output", Unit: "second", Amount: "1", Resolution: "540p"}},
	})
	require.ErrorIs(t, err, ErrPricingUnresolved)
}

func TestRateUsageRejectsMissingResolutionRate(t *testing.T) {
	_, err := RateUsage(RateInput{
		Pricing: videoPricing(), Region: "us-west-2", ExchangeRate: decimal.NewFromInt(1),
		Measures: []platform.UsageMeasure{{Usage: "output", Unit: "second", Amount: "1", Resolution: "1080p"}},
	})
	require.True(t, errors.Is(err, ErrRateUnresolved))
}

func TestRateUsageFreePricing(t *testing.T) {
	rated, err := RateUsage(RateInput{
		Pricing:      &platform.ModelPrice{Status: "free"},
		Measures:     []platform.UsageMeasure{{Usage: "output", Unit: "second", Amount: "10"}},
		ExchangeRate: decimal.NewFromInt(850000),
	})
	require.NoError(t, err)
	require.Equal(t, "free", rated.Status)
	require.Equal(t, "0", rated.TotalCostIRR)
}

func videoPricing() *platform.ModelPrice {
	return &platform.ModelPrice{Status: "priced", Profiles: []platform.PricingProfile{{
		Region: "us-west-2", Currency: "USD", Rates: []platform.PricingRate{
			{Key: "video-540p", Usage: "output", Unit: "second", Amount: "0.75", Quantity: 1, ServiceTier: "standard", Resolution: "540p"},
			{Key: "video-720p", Usage: "output", Unit: "second", Amount: "1.5", Quantity: 1, ServiceTier: "standard", Resolution: "720p"},
		},
	}}}
}
