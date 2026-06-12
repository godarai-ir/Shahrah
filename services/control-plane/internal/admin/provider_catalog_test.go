package admin

import (
	"context"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapDiscoveryModels_PreservesGatewayAndUIFields(t *testing.T) {
	inputPrice := 0.4
	outputPrice := 1.2
	contextWindow := 128000

	items := mapDiscoveryModels([]modelDiscoveryModel{
		{
			CanonicalID:      "gpt-4.1-mini",
			Name:             "GPT-4.1 Mini",
			ContextWindow:    &contextWindow,
			Capabilities:     []string{"chat", "reasoning"},
			InputModalities:  []string{"text"},
			OutputModalities: []string{"text"},
			Pricing: &modelDiscoveryPricing{
				InputPerMillionTokens:  &inputPrice,
				OutputPerMillionTokens: &outputPrice,
			},
		},
	})

	require.Len(t, items, 1)
	assert.Equal(t, "gpt-4.1-mini", items[0].Model)
	assert.Equal(t, "GPT-4.1 Mini", items[0].DisplayName)
	assert.Equal(t, []platform.ModelType{platform.ModelTypeChat}, items[0].Types)
	assert.Equal(t, &contextWindow, items[0].ContextWindow)
	assert.Equal(t, []string{"text"}, items[0].InputModalities)
	assert.Equal(t, []string{"text"}, items[0].OutputModalities)
	assert.Equal(t, []string{"chat", "reasoning"}, items[0].Features)
	require.NotNil(t, items[0].PublicPricing)
	assert.Equal(t, &inputPrice, items[0].PublicPricing.InputUSDPer1M)
	assert.Equal(t, &outputPrice, items[0].PublicPricing.OutputUSDPer1M)
}

func TestResolveProviderAccountModelSelections_ExpandsCatalogMetadata(t *testing.T) {
	t.Setenv("SHAHRAH_MODEL_DISCOVERY_URL", "")

	items, invalid, err := resolveProviderAccountModelSelections(context.Background(), "openai", []platform.UpsertTenantProviderAccountModelSelection{
		{
			Model:       "gpt-4.1-mini",
			PricingMode: platform.PricingModePublic,
		},
	})
	require.NoError(t, err)
	require.Empty(t, invalid)
	require.Len(t, items, 1)

	assert.Equal(t, "gpt-4.1-mini", items[0].Model)
	assert.True(t, items[0].Enabled)
	assert.NotEmpty(t, items[0].Types)
	assert.NotEmpty(t, items[0].InputModalities)
	assert.NotEmpty(t, items[0].OutputModalities)
	assert.NotNil(t, items[0].PublicPricing)
	assert.Equal(t, platform.PricingModePublic, items[0].PricingMode)
}
