package billing

import (
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/credits"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func perMillionUSD(value string) decimal.Decimal {
	return decimal.RequireFromString(value).Div(decimal.NewFromInt(1_000_000))
}

func tokenRate(dimension, perMillion string) normalizedPricingRate {
	return normalizedPricingRate{Dimension: dimension, Unit: "token", UnitPriceUSD: perMillionUSD(perMillion)}
}

func TestCreditRateCardFromSnapshots(t *testing.T) {
	t.Parallel()

	anthropic := &AnthropicPricingSnapshotManager{now: time.Now}
	sonnetRates := []normalizedPricingRate{
		tokenRate("input_tokens", "2"),
		tokenRate("output_tokens", "10"),
		tokenRate("cache_read_input_tokens", "0.2"),
		{Dimension: "web_search_requests", Unit: "request", UnitPriceUSD: decimal.RequireFromString("0.01")},
	}
	anthropic.state.Store(&anthropicPricingSnapshot{
		Revision:    "anthropic-rev-1",
		GeneratedAt: time.Now(),
		Models: map[string]anthropicPricingModel{
			"claude-sonnet-5": {
				CanonicalID: "anthropic/claude-sonnet-5",
				Status:      "active",
				Rates: map[string]normalizedPricingRate{
					"input_tokens":  sonnetRates[0],
					"output_tokens": sonnetRates[1],
				},
				RateList: sonnetRates,
			},
			"claude-retired": {CanonicalID: "anthropic/claude-retired", Status: "retired"},
		},
	})

	staleCatalog := &CatalogPricingSnapshotManager{provider: "xai", now: time.Now}
	staleCatalog.state.Store(&anthropicPricingSnapshot{
		Revision:    "xai-stale",
		GeneratedAt: time.Now().Add(-catalogPricingMaxAge - time.Minute),
		Models:      map[string]anthropicPricingModel{"grok-4.7": {Status: "active"}},
	})

	engine := &PricingEngine{
		anthropicPricing: anthropic,
		catalogPricing:   map[string]*CatalogPricingSnapshotManager{"xai": staleCatalog},
	}

	effective := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	card, revision := engine.CreditRateCard(credits.RateCard{Version: 1, EffectiveFrom: effective})

	require.Equal(t, platform.CreditRateCardObject, card.Object)
	require.Equal(t, "rc-1", card.Version)
	require.Equal(t, platform.CreditUnit, card.Unit)
	require.Equal(t, effective, card.EffectiveFrom)
	require.NotEmpty(t, revision)
	require.Len(t, card.Models, 1, "retired and stale-provider models must be omitted")

	model := card.Models[0]
	require.Equal(t, "anthropic/claude-sonnet-5", model.Model)
	require.Equal(t, "anthropic", model.Provider)
	require.Equal(t, "4.50", model.TypicalMessageCredits)
	require.Equal(t, []platform.CreditRate{
		{Dimension: "input_tokens", Per: platform.CreditRatePer1KToken, Credits: "1.00"},
		{Dimension: "output_tokens", Per: platform.CreditRatePer1KToken, Credits: "5.00"},
		{Dimension: "cache_read_input_tokens", Per: platform.CreditRatePer1KToken, Credits: "0.10"},
		{Dimension: "web_search_requests", Per: "request", Credits: "5.00"},
	}, model.Rates)

	// The revision changes with the rate-card version and with provider revisions.
	_, nextVersion := engine.CreditRateCard(credits.RateCard{Version: 2})
	require.NotEqual(t, revision, nextVersion)
}

func TestCreditRateCardAppliesMultiplier(t *testing.T) {
	t.Parallel()

	anthropic := &AnthropicPricingSnapshotManager{now: time.Now}
	rates := []normalizedPricingRate{tokenRate("input_tokens", "4"), tokenRate("output_tokens", "20")}
	anthropic.state.Store(&anthropicPricingSnapshot{
		Revision:    "r",
		GeneratedAt: time.Now(),
		Models: map[string]anthropicPricingModel{
			"claude-opus-5-5": {
				CanonicalID: "claude-opus-5-5",
				Status:      "active",
				Rates:       map[string]normalizedPricingRate{"input_tokens": rates[0], "output_tokens": rates[1]},
				RateList:    rates,
			},
		},
	})

	engine := &PricingEngine{anthropicPricing: anthropic}
	card, _ := engine.CreditRateCard(credits.RateCard{
		Version:     3,
		Multipliers: map[string]decimal.Decimal{"anthropic/claude-opus-5-5": decimal.RequireFromString("1.1")},
	})

	require.Len(t, card.Models, 1)
	require.Equal(t, "2.20", card.Models[0].Rates[0].Credits)
	require.Equal(t, "9.90", card.Models[0].TypicalMessageCredits)
}

func TestCreditRateCardWithoutSnapshots(t *testing.T) {
	t.Parallel()

	var engine *PricingEngine

	card, revision := engine.CreditRateCard(credits.IdentityRateCard())
	require.Empty(t, card.Models)
	require.NotNil(t, card.Models)
	require.NotEmpty(t, revision)
}
