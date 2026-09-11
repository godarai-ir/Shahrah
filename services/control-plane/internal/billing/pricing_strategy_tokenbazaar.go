package billing

import (
	"fmt"
	"sort"

	"github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/shopspring/decimal"
)

// tokenBazaarCachedTokens splits a TokenBazaar usage record into the billable
// (uncached) prompt tokens, the cache-read tokens, and the cache-write tokens.
//
// TokenBazaar fronts several upstream families through different envelopes,
// and those envelopes disagree about whether cached tokens are already counted
// inside the prompt-token total:
//
//   - The Anthropic Messages envelope (POST /v1/messages, parsed by
//     proxy.parseAnthropicNativeUsage) reports cache_read_input_tokens and
//     cache_creation_input_tokens ALONGSIDE input_tokens, which excludes them.
//     Those quantities are billed in addition to the input dimension.
//   - The OpenAI and Gemini envelopes report cached tokens INSIDE
//     prompt_tokens, surfaced as usage.PromptTokensDetails.CachedTokens, so
//     the cached portion is subtracted from the billable input dimension
//     before being priced at the cache-read rate.
//
// Both shapes are handled because one TokenBazaar deployment serves all of
// them; in practice a single response carries one convention or the other,
// never both.
func tokenBazaarCachedTokens(usage *openai.Usage) (billableInput, cacheRead, cacheWrite int, err error) {
	inclusiveCached := 0
	if usage.PromptTokensDetails != nil {
		inclusiveCached = usage.PromptTokensDetails.CachedTokens
	}

	if inclusiveCached < 0 || inclusiveCached > usage.PromptTokens {
		return 0, 0, 0, fmt.Errorf("%w: invalid TokenBazaar cached token usage", ErrPricingUnavailable)
	}

	if usage.CacheReadInputTokens < 0 {
		return 0, 0, 0, fmt.Errorf("%w: invalid TokenBazaar cache read token usage", ErrPricingUnavailable)
	}

	cacheWrite = usage.CacheCreation5mInputTokens + usage.CacheCreation1hInputTokens
	if cacheWrite == 0 {
		cacheWrite = usage.CacheCreationInputTokens
	}

	if cacheWrite < 0 {
		return 0, 0, 0, fmt.Errorf("%w: invalid TokenBazaar cache creation token usage", ErrPricingUnavailable)
	}

	return usage.PromptTokens - inclusiveCached, inclusiveCached + usage.CacheReadInputTokens, cacheWrite, nil
}

// tokenBazaarUsageHasCachedTokens reports whether a usage record carries any
// cached-input dimension, in either of the two conventions above. When it does
// not -- the common case, and every embeddings response -- billing stays on the
// inherited standardPricingStrategy.FinalizeUsage path unchanged.
func tokenBazaarUsageHasCachedTokens(usage *openai.Usage) bool {
	if usage == nil {
		return false
	}

	if usage.PromptTokensDetails != nil && usage.PromptTokensDetails.CachedTokens != 0 {
		return true
	}

	return usage.CacheReadInputTokens != 0 ||
		usage.CacheCreationInputTokens != 0 ||
		usage.CacheCreation5mInputTokens != 0 ||
		usage.CacheCreation1hInputTokens != 0
}

// tokenBazaarUsageCostBreakdown prices a TokenBazaar completion that reported
// cached input tokens.
//
// The embedded standardPricingStrategy.FinalizeUsage prices only
// prompt_tokens and completion_tokens, which bills every cache-read token at
// zero on the Anthropic Messages path (where input_tokens excludes them) even
// though the catalog pins a cache_read_input_token_cost for those models.
//
// Cache reads are billed against "cache_read_tokens", the canonical catalog
// dimension every other catalog-pricing provider uses (see
// moonshotPricingDimensionsComplete / deepSeekPricingDimensionsComplete in
// catalog_pricing_snapshot.go), and fail closed when the rate is absent -- the
// same posture mistralUsageCostBreakdown and anthropicUsageCostBreakdown take,
// so a pricing gap refuses the request rather than silently mis-billing.
//
// Cache writes are billed against "cache_write_5m_tokens" /
// "cache_write_1h_tokens". apps/model-discovery/scripts/import-models-from-yaml.ts
// maps the catalog's cache_creation_input_token_cost to cache_write_5m_tokens
// for every provider slug except "openai", and
// cache_creation_input_token_cost_per_hour to cache_write_1h_tokens, so those
// dimensions do exist for TokenBazaar's Claude models (their YAMLs pin
// cache_creation_input_token_cost).
//
// Unlike cache reads this degrades instead of failing closed: TokenBazaar has
// no dimensions-complete gate (contrast moonshotPricingDimensionsComplete)
// asserting a cache-write rate is present, so a model that reports cache
// creation without pinning a rate folds those tokens into the plain input
// dimension rather than refusing the request. Cache-write tokens are input
// tokens that were also persisted, so the input rate is a conservative floor.
func tokenBazaarUsageCostBreakdown(
	model string,
	usage *openai.Usage,
	price priceInfo,
	exchangeRate float64,
) (platform.UsageCostBreakdown, error) {
	if usage == nil {
		return platform.UsageCostBreakdown{}, fmt.Errorf("%w: provider usage is missing", ErrPricingUnavailable)
	}

	billableInput, cacheRead, cacheWrite, err := tokenBazaarCachedTokens(usage)
	if err != nil {
		return platform.UsageCostBreakdown{}, err
	}

	const (
		inputDimension     = "input_tokens"
		outputDimension    = "output_tokens"
		cacheReadDimension = "cache_read_tokens"
	)

	quantities := map[string]int64{
		inputDimension:  int64(billableInput),
		outputDimension: int64(usage.CompletionTokens),
	}

	if cacheRead > 0 {
		if _, ok := price.Rates[cacheReadDimension]; !ok {
			return platform.UsageCostBreakdown{}, fmt.Errorf(
				"%w: missing TokenBazaar rate for %s",
				ErrPricingUnavailable,
				cacheReadDimension,
			)
		}

		quantities[cacheReadDimension] += int64(cacheRead)
	}

	if cacheWrite > 0 {
		quantities[tokenBazaarCacheWriteDimension(price, usage)] += int64(cacheWrite)
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

		rate, ok := price.Rates[dimension]
		if !ok {
			return platform.UsageCostBreakdown{}, fmt.Errorf(
				"%w: missing TokenBazaar rate for %s",
				ErrPricingUnavailable,
				dimension,
			)
		}

		cost := rate.UnitPriceUSD.Mul(decimal.NewFromInt(quantity))

		totalExact = totalExact.Add(cost)
		if dimension == outputDimension {
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
		totalTokens = usage.PromptTokens + usage.CompletionTokens + usage.CacheReadInputTokens + cacheWrite
	}

	return platform.UsageCostBreakdown{
		Model: model, InputTokens: usage.PromptTokens, OutputTokens: usage.CompletionTokens,
		TotalTokens: totalTokens, CacheReadInputTokens: cacheRead,
		InputCostUSD: inputUSD, OutputCostUSD: outputUSD, TotalCostUSD: totalUSD,
		ExchangeRate: exchangeRate, TotalCostIRR: totalUSD * exchangeRate,
		PricingRevision: price.Revision, ModelRuntimeProfile: price.RuntimeProfile,
		PricingSnapshot: usagePricingSnapshot(price), Components: components,
		ServerToolUse: usage.ServerToolUse,
	}, nil
}

// tokenBazaarCacheWriteDimension picks the dimension cache-creation tokens are
// billed under: the matching Anthropic-style cache-write rate when the
// snapshot carries one, and otherwise the plain input dimension. See
// tokenBazaarUsageCostBreakdown for why this degrades instead of failing
// closed.
func tokenBazaarCacheWriteDimension(price priceInfo, usage *openai.Usage) string {
	dimension := "cache_write_5m_tokens"
	if usage.CacheCreation5mInputTokens == 0 && usage.CacheCreation1hInputTokens > 0 {
		dimension = "cache_write_1h_tokens"
	}

	if _, ok := price.Rates[dimension]; ok {
		return dimension
	}

	return "input_tokens"
}
