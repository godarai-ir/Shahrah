package admin

import (
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/stretchr/testify/require"
)

func TestValidatePrivatePricingOverrideAcceptsPartialOfficialRate(t *testing.T) {
	publicPrice := &platform.ModelPrice{Rates: []platform.ModelPricingRate{
		{Key: "input", Dimension: "input_tokens", Unit: "token", UnitPriceUSD: "0.000001"},
		{Key: "output", Dimension: "output_tokens", Unit: "token", UnitPriceUSD: "0.000002"},
	}}
	privatePrice := &platform.ModelPrice{Rates: []platform.ModelPricingRate{
		{Key: "output", Dimension: "output_tokens", Unit: "token", UnitPriceUSD: "0.0000015"},
	}}

	require.NoError(t, validatePrivatePricingOverride(publicPrice, privatePrice))
}

func TestValidatePrivatePricingOverrideRejectsUnknownOrMutatedRate(t *testing.T) {
	publicPrice := &platform.ModelPrice{Rates: []platform.ModelPricingRate{
		{Key: "input", Dimension: "input_tokens", Unit: "token", UnitPriceUSD: "0.000001"},
	}}
	for _, privatePrice := range []*platform.ModelPrice{
		{Rates: []platform.ModelPricingRate{{Key: "unknown", Dimension: "input_tokens", Unit: "token", UnitPriceUSD: "1"}}},
		{Rates: []platform.ModelPricingRate{{Key: "input", Dimension: "output_tokens", Unit: "token", UnitPriceUSD: "1"}}},
		{Rates: []platform.ModelPricingRate{{Key: "input", Dimension: "input_tokens", Unit: "token", UnitPriceUSD: "-1"}}},
	} {
		require.Error(t, validatePrivatePricingOverride(publicPrice, privatePrice))
	}
}
