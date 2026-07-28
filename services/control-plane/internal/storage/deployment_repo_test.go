package storage

import (
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/stretchr/testify/assert"
)

func TestProviderAccountUsesNoCredential(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		provider string
		authMode string
		want     bool
	}{
		{
			name:     "internal mock overrides platform managed auth mode",
			provider: "mock",
			authMode: "api_key",
			want:     true,
		},
		{
			name:     "internal mock comparison is normalized",
			provider: " Mock ",
			authMode: "platform_managed",
			want:     true,
		},
		{
			name:     "explicit no credential remains supported",
			provider: "openai",
			authMode: "no_credential",
			want:     true,
		},
		{
			name:     "external provider keeps its configured auth mode",
			provider: "openai",
			authMode: "api_key",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, providerAccountUsesNoCredential(tt.provider, tt.authMode))
		})
	}
}

func TestCatalogPriceResolved(t *testing.T) {
	t.Parallel()

	valid := &platform.ModelPrice{Profiles: []platform.PricingProfile{{
		Region:   "global",
		Currency: "USD",
		Rates: []platform.PricingRate{{
			Key:      "input",
			Usage:    "input_tokens",
			Unit:     "token",
			Amount:   "0.01",
			Quantity: 1000,
		}},
	}}}

	assert.True(t, catalogPriceResolved(valid))
	assert.True(t, catalogPriceResolved(&platform.ModelPrice{Status: "free"}))
	assert.False(t, catalogPriceResolved(nil))
	assert.False(t, catalogPriceResolved(&platform.ModelPrice{Status: "unknown", Profiles: valid.Profiles}))
	assert.False(t, catalogPriceResolved(&platform.ModelPrice{Profiles: []platform.PricingProfile{{
		Region:   "global",
		Currency: "USD",
	}}}))
}

func TestProviderConnectionBaseURLAliasesFollowProbePrecedence(t *testing.T) {
	t.Parallel()

	config := map[string]interface{}{
		"api_base":            "https://legacy.example",
		"base_url":            "https://base.example",
		"endpoint":            "https://azure.example",
		"serving_endpoint_url": "https://databricks.example",
	}

	got := firstString(
		config,
		nil,
		"serving_endpoint_url",
		"endpoint",
		"base_url",
		"api_base",
	)

	assert.Equal(t, "https://databricks.example", got)
}
