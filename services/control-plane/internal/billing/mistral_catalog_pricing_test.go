package billing

import (
	"testing"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/shopspring/decimal"
)

func TestCatalogPricingProviderRecognizesCanonicalMistralIdentity(t *testing.T) {
	for _, model := range []string{"mistral-ai:account:mistral-large-2512", "mistral/mistral-small-2603", "codestral-2508", "voxtral-mini-2602"} {
		if got := catalogPricingProvider(model); got != "mistral-ai" {
			t.Fatalf("catalogPricingProvider(%q) = %q", model, got)
		}
	}
}

func TestMistralUsagePricingHandlesBatchCacheAndTools(t *testing.T) {
	rates := []normalizedPricingRate{
		{Dimension: "input_tokens", Unit: "token", UnitPriceUSD: decimal.RequireFromString("0.000001")},
		{Dimension: "input_tokens", Unit: "token", UnitPriceUSD: decimal.RequireFromString("0.0000005"), Condition: map[string]interface{}{"api": "batches"}},
		{Dimension: "output_tokens", Unit: "token", UnitPriceUSD: decimal.RequireFromString("0.000002")},
		{Dimension: "output_tokens", Unit: "token", UnitPriceUSD: decimal.RequireFromString("0.000001"), Condition: map[string]interface{}{"api": "batches"}},
		{Dimension: "cache_read_tokens", Unit: "token", UnitPriceUSD: decimal.RequireFromString("0.0000001")},
		{Dimension: "code_execution_calls", Unit: "call", UnitPriceUSD: decimal.RequireFromString("0.03")},
	}

	price := priceInfo{Provider: "mistral-ai", RateList: rates, Rates: map[string]normalizedPricingRate{}}
	for _, rate := range rates {
		if len(rate.Condition) == 0 {
			price.Rates[rate.Dimension] = rate
		}
	}

	breakdown, err := mistralUsageCostBreakdown("mistral-large-2512", &contractsopenai.Usage{
		PromptTokens: 100, CompletionTokens: 20, TotalTokens: 120, ServiceTier: "batch",
		PromptTokensDetails: &contractsopenai.UsageTokenDetails{CachedTokens: 40},
		ServerToolUse:       map[string]int{"code_execution_calls": 2},
	}, price, 1, map[string]string{"service_tier": "batch"})
	if err != nil {
		t.Fatal(err)
	}

	if breakdown.CacheReadInputTokens != 40 || len(breakdown.Components) != 4 {
		t.Fatalf("unexpected Mistral breakdown: %#v", breakdown)
	}
}

func TestMistralPricingDimensionsFailClosedByModality(t *testing.T) {
	rate := normalizedPricingRate{UnitPriceUSD: decimal.RequireFromString("0.01")}

	profile := platform.ModelRuntimeProfile{Endpoints: []string{"ocr"}}
	if mistralPricingDimensionsComplete(profile, map[string]normalizedPricingRate{"ocr_input_pages": rate}) {
		t.Fatal("OCR pricing must require both normal and annotated page dimensions")
	}

	annotatedRate := rate

	annotatedRate.Condition = map[string]interface{}{"document_ai": true}
	if !mistralPricingDimensionsComplete(profile, map[string]normalizedPricingRate{"ocr_input_pages": rate, "ocr_annotated_pages": annotatedRate}) {
		t.Fatal("complete OCR pricing was rejected")
	}
}
