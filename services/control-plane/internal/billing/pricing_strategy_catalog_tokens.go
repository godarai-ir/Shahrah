package billing

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/shopspring/decimal"
)

func moonshotUsageCostBreakdown(
	model string,
	usage *openai.Usage,
	price priceInfo,
	exchangeRate float64,
) (platform.UsageCostBreakdown, error) {
	if usage == nil {
		return platform.UsageCostBreakdown{}, fmt.Errorf("%w: Moonshot usage is missing", ErrPricingUnavailable)
	}

	cached := 0
	if usage.PromptTokensDetails != nil {
		cached = usage.PromptTokensDetails.CachedTokens
	}

	if cached < 0 || cached > usage.PromptTokens {
		return platform.UsageCostBreakdown{}, fmt.Errorf("%w: invalid Moonshot cached token usage", ErrPricingUnavailable)
	}

	quantities := map[string]int64{
		"input_tokens":      int64(usage.PromptTokens - cached),
		"cache_read_tokens": int64(cached),
		"output_tokens":     int64(usage.CompletionTokens),
	}
	dimensions := []string{"input_tokens", "cache_read_tokens", "output_tokens"}
	components := make([]platform.UsageCostComponent, 0, len(dimensions))
	inputExact, outputExact := decimal.Zero, decimal.Zero

	for _, dimension := range dimensions {
		quantity := quantities[dimension]
		if quantity == 0 {
			continue
		}

		rate, ok := price.Rates[dimension]
		if !ok || len(rate.Condition) != 0 {
			return platform.UsageCostBreakdown{}, fmt.Errorf("%w: missing Moonshot rate for %s", ErrPricingUnavailable, dimension)
		}

		cost := rate.UnitPriceUSD.Mul(decimal.NewFromInt(quantity))
		if dimension == "output_tokens" {
			outputExact = outputExact.Add(cost)
		} else {
			inputExact = inputExact.Add(cost)
		}

		components = append(components, usageCostComponent(dimension, quantity, rate))
	}

	inputUSD, _ := inputExact.Float64()
	outputUSD, _ := outputExact.Float64()
	totalUSD := inputUSD + outputUSD

	totalTokens := usage.TotalTokens
	if totalTokens <= 0 {
		totalTokens = usage.PromptTokens + usage.CompletionTokens
	}

	return platform.UsageCostBreakdown{
		Model: model, InputTokens: usage.PromptTokens, OutputTokens: usage.CompletionTokens,
		TotalTokens: totalTokens, CacheReadInputTokens: cached,
		InputCostUSD: inputUSD, OutputCostUSD: outputUSD, TotalCostUSD: totalUSD,
		ExchangeRate: exchangeRate, TotalCostIRR: totalUSD * exchangeRate,
		PricingRevision: price.Revision, ModelRuntimeProfile: price.RuntimeProfile,
		PricingSnapshot: usagePricingSnapshot(price), Components: components,
	}, nil
}

func deepSeekUsageCostBreakdown(
	model string,
	usage *openai.Usage,
	price priceInfo,
	exchangeRate float64,
) (platform.UsageCostBreakdown, error) {
	if usage == nil {
		return platform.UsageCostBreakdown{}, fmt.Errorf("%w: DeepSeek usage is missing", ErrPricingUnavailable)
	}

	hit := usage.PromptCacheHitTokens

	miss := usage.PromptCacheMissTokens
	if hit < 0 || miss < 0 || hit+miss != usage.PromptTokens {
		return platform.UsageCostBreakdown{}, fmt.Errorf("%w: invalid DeepSeek cache token usage", ErrPricingUnavailable)
	}

	quantities := map[string]int64{
		"input_tokens":      int64(miss),
		"cache_read_tokens": int64(hit),
		"output_tokens":     int64(usage.CompletionTokens),
	}
	dimensions := []string{"input_tokens", "cache_read_tokens", "output_tokens"}
	components := make([]platform.UsageCostComponent, 0, len(dimensions))
	inputExact, outputExact := decimal.Zero, decimal.Zero

	for _, dimension := range dimensions {
		quantity := quantities[dimension]
		if quantity == 0 {
			continue
		}

		rate, ok := price.Rates[dimension]
		if !ok || len(rate.Condition) != 0 {
			return platform.UsageCostBreakdown{}, fmt.Errorf("%w: missing DeepSeek rate for %s", ErrPricingUnavailable, dimension)
		}

		cost := rate.UnitPriceUSD.Mul(decimal.NewFromInt(quantity))
		if dimension == "output_tokens" {
			outputExact = outputExact.Add(cost)
		} else {
			inputExact = inputExact.Add(cost)
		}

		components = append(components, usageCostComponent(dimension, quantity, rate))
	}

	inputUSD, _ := inputExact.Float64()
	outputUSD, _ := outputExact.Float64()
	totalUSD := inputUSD + outputUSD

	totalTokens := usage.TotalTokens
	if totalTokens <= 0 {
		totalTokens = usage.PromptTokens + usage.CompletionTokens
	}

	return platform.UsageCostBreakdown{
		Model: model, InputTokens: usage.PromptTokens, OutputTokens: usage.CompletionTokens,
		TotalTokens: totalTokens, CacheReadInputTokens: hit,
		InputCostUSD: inputUSD, OutputCostUSD: outputUSD, TotalCostUSD: totalUSD,
		ExchangeRate: exchangeRate, TotalCostIRR: totalUSD * exchangeRate,
		PricingRevision: price.Revision, ModelRuntimeProfile: price.RuntimeProfile,
		PricingSnapshot: usagePricingSnapshot(price), Components: components,
	}, nil
}

func mistralUsageCostBreakdown(model string, usage *openai.Usage, price priceInfo, exchangeRate float64, conditions map[string]string) (platform.UsageCostBreakdown, error) {
	if usage == nil {
		return platform.UsageCostBreakdown{}, fmt.Errorf("%w: Mistral usage is missing", ErrPricingUnavailable)
	}

	cached := 0
	if usage.PromptTokensDetails != nil {
		cached = usage.PromptTokensDetails.CachedTokens
	}

	if cached < 0 || cached > usage.PromptTokens {
		return platform.UsageCostBreakdown{}, fmt.Errorf("%w: invalid Mistral cached token usage", ErrPricingUnavailable)
	}

	inputDimension, outputDimension, cacheDimension := "input_tokens", "output_tokens", "cache_read_tokens"

	if strings.EqualFold(strings.TrimSpace(usage.ServiceTier), "batch") || strings.EqualFold(strings.TrimSpace(conditions["service_tier"]), "batch") {
		if conditions == nil {
			conditions = map[string]string{}
		}

		conditions["api"] = "batches"
	}

	quantities := map[string]int64{
		inputDimension:  int64(usage.PromptTokens - cached),
		outputDimension: int64(usage.CompletionTokens),
	}
	if cached > 0 {
		if _, ok := price.Rates[cacheDimension]; !ok {
			return platform.UsageCostBreakdown{}, fmt.Errorf("%w: missing Mistral rate for %s", ErrPricingUnavailable, cacheDimension)
		}

		quantities[cacheDimension] = int64(cached)
	}

	for dimension, quantity := range usage.ServerToolUse {
		if quantity > 0 {
			quantities[dimension] += int64(quantity)
		}
	}

	dimensions := make([]string, 0, len(quantities))
	for dimension := range quantities {
		dimensions = append(dimensions, dimension)
	}

	sort.Strings(dimensions)

	totalExact, inputExact, outputExact := decimal.Zero, decimal.Zero, decimal.Zero

	components := make([]platform.UsageCostComponent, 0, len(dimensions))
	for _, dimension := range dimensions {
		quantity := quantities[dimension]
		if quantity == 0 {
			continue
		}

		rate, rateErr := selectOpenAIRate(price.RateList, dimension, usage.PromptTokens, "", conditions)
		if rateErr != nil {
			return platform.UsageCostBreakdown{}, fmt.Errorf("%w: missing or ambiguous Mistral rate for %s", ErrPricingUnavailable, dimension)
		}

		cost := rate.UnitPriceUSD.Mul(decimal.NewFromInt(quantity))

		totalExact = totalExact.Add(cost)
		if dimension == outputDimension || strings.Contains(dimension, "output") {
			outputExact = outputExact.Add(cost)
		} else {
			inputExact = inputExact.Add(cost)
		}

		components = append(components, usageCostComponent(dimension, quantity, rate))
	}

	inputUSD, _ := inputExact.Float64()
	outputUSD, _ := outputExact.Float64()
	totalUSD, _ := totalExact.Float64()

	totalTokens := usage.TotalTokens
	if totalTokens <= 0 {
		totalTokens = usage.PromptTokens + usage.CompletionTokens
	}

	return platform.UsageCostBreakdown{
		Model: model, InputTokens: usage.PromptTokens, OutputTokens: usage.CompletionTokens,
		TotalTokens: totalTokens, CacheReadInputTokens: cached,
		InputCostUSD: inputUSD, OutputCostUSD: outputUSD, TotalCostUSD: totalUSD,
		ExchangeRate: exchangeRate, TotalCostIRR: totalUSD * exchangeRate,
		PricingRevision: price.Revision, ModelRuntimeProfile: price.RuntimeProfile,
		PricingSnapshot: usagePricingSnapshot(price), Components: components,
		ServerToolUse: usage.ServerToolUse,
	}, nil
}
