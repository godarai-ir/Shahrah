package billing

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/providerregistry"
	"github.com/shopspring/decimal"
)

type openAITokenPricingPolicy struct {
	// longContextMinPromptTokens is the smallest prompt size billed at
	// long-context rates. It must match the catalog min_prompt_tokens
	// condition so boundary requests always resolve to a published rate.
	longContextMinPromptTokens   int
	supportsSpeedTierLongContext bool
}

var (
	// OpenAI bills prompts with more than 272K input tokens at long-context
	// rates, so a 272,000-token prompt still uses the short-context rates.
	openAIStandardTokenPricingPolicy = openAITokenPricingPolicy{longContextMinPromptTokens: 272001}
	xAITokenPricingPolicy            = openAITokenPricingPolicy{
		longContextMinPromptTokens:   200000,
		supportsSpeedTierLongContext: true,
	}
)

func openAIUsageCostBreakdown(
	model string,
	usage *openai.Usage,
	price priceInfo,
	exchangeRate float64,
	conditions map[string]string,
) (platform.UsageCostBreakdown, error) {
	policy := openAIStandardTokenPricingPolicy
	if definition, ok := providerregistry.Lookup(price.Provider); ok &&
		definition.PricingStrategy == providerregistry.PricingStrategyXAI {
		policy = xAITokenPricingPolicy
	}

	return openAIUsageCostBreakdownWithPolicy(
		model,
		usage,
		price,
		exchangeRate,
		conditions,
		policy,
	)
}

func openAIUsageCostBreakdownWithPolicy(
	model string,
	usage *openai.Usage,
	price priceInfo,
	exchangeRate float64,
	conditions map[string]string,
	policy openAITokenPricingPolicy,
) (platform.UsageCostBreakdown, error) {
	requestedTier := strings.TrimSpace(usage.ServiceTier)
	if requestedTier == "" {
		requestedTier = strings.TrimSpace(conditions["service_tier"])
	}

	if err := validateOpenAIServiceTier(model, requestedTier, price); err != nil {
		return platform.UsageCostBreakdown{}, err
	}

	tier, err := normalizeOpenAIServiceTier(requestedTier)
	if err != nil {
		return platform.UsageCostBreakdown{}, err
	}

	promptAudio, promptCached, promptCacheWrite, completionAudio := 0, 0, 0, 0
	if usage.PromptTokensDetails != nil {
		promptAudio = usage.PromptTokensDetails.AudioTokens
		promptCached = usage.PromptTokensDetails.CachedTokens
		promptCacheWrite = usage.PromptTokensDetails.CacheWriteTokens
	}

	if usage.CompletionTokensDetails != nil {
		completionAudio = usage.CompletionTokensDetails.AudioTokens
	}

	if promptAudio < 0 || promptCached < 0 || promptCacheWrite < 0 || completionAudio < 0 ||
		promptAudio+promptCached+promptCacheWrite > usage.PromptTokens ||
		completionAudio > usage.CompletionTokens {
		return platform.UsageCostBreakdown{}, fmt.Errorf(
			"%w: invalid OpenAI usage token detail totals",
			ErrPricingUnavailable,
		)
	}

	inputDimension, outputDimension := "input_tokens", "output_tokens"
	cacheDimension := "cache_read_tokens"
	cacheWriteDimension := "cache_write_tokens"

	switch tier {
	case "batch":
		inputDimension, outputDimension = "batch_input_tokens", "batch_output_tokens"
		cacheDimension = "batch_cache_read_tokens"
		cacheWriteDimension = "batch_cache_write_tokens"
	case "flex":
		inputDimension, outputDimension = "flex_input_tokens", "flex_output_tokens"
		cacheDimension = "flex_cache_read_tokens"
		cacheWriteDimension = "flex_cache_write_tokens"
	case "priority":
		inputDimension, outputDimension = "speed_input_tokens", "speed_output_tokens"
		cacheDimension = "speed_cache_read_tokens"
		cacheWriteDimension = "speed_cache_write_tokens"
	}

	if usage.PromptTokens >= policy.longContextMinPromptTokens {
		switch tier {
		case "priority":
			// Models that publish Fast mode long-context rates, such as
			// GPT-6 Sol and Luna, are priced at them; the others reject.
			if policy.supportsSpeedTierLongContext ||
				hasOpenAIRate(price.RateList, "speed_long_context_input_tokens", usage.PromptTokens, tier, conditions) {
				inputDimension, outputDimension = "speed_long_context_input_tokens", "speed_long_context_output_tokens"
				cacheDimension = "speed_long_context_cache_read_tokens"
				cacheWriteDimension = "speed_long_context_cache_write_tokens"
			} else {
				return platform.UsageCostBreakdown{}, newUnsupportedServiceTierError(
					price,
					model,
					requestedTier,
				)
			}
		case "batch":
			inputDimension, outputDimension = "batch_long_context_input_tokens", "batch_long_context_output_tokens"
			cacheDimension = "batch_long_context_cache_read_tokens"
			cacheWriteDimension = "batch_long_context_cache_write_tokens"
		case "flex":
			inputDimension, outputDimension = "flex_long_context_input_tokens", "flex_long_context_output_tokens"
			cacheDimension = "flex_long_context_cache_read_tokens"
			cacheWriteDimension = "flex_long_context_cache_write_tokens"
		default:
			inputDimension, outputDimension = "long_context_input_tokens", "long_context_output_tokens"
			cacheDimension = "long_context_cache_read_tokens"
			cacheWriteDimension = "long_context_cache_write_tokens"
		}
	}

	if promptCached > 0 && !hasOpenAIRate(price.RateList, cacheDimension, usage.PromptTokens, tier, conditions) {
		// Pro models publish no cached-input discount. Cached tokens are still
		// billable, at the applicable ordinary input rate.
		cacheDimension = inputDimension
	}

	if promptCacheWrite > 0 && !hasOpenAIRate(price.RateList, cacheWriteDimension, usage.PromptTokens, tier, conditions) {
		cacheWriteDimension = inputDimension
	}

	quantities := make(map[string]int64, 7)
	quantities[inputDimension] += int64(usage.PromptTokens - promptAudio - promptCached - promptCacheWrite)
	quantities[outputDimension] += int64(usage.CompletionTokens - completionAudio)
	quantities[cacheDimension] += int64(promptCached)
	quantities[cacheWriteDimension] += int64(promptCacheWrite)
	quantities["audio_input_tokens"] += int64(promptAudio)
	quantities["audio_output_tokens"] += int64(completionAudio)

	for tool, quantity := range usage.ServerToolUse {
		if quantity > 0 {
			quantities[tool] += int64(quantity)
		}
	}

	if webSearchCalls := usage.ServerToolUse["web_search_requests"]; webSearchCalls > 0 &&
		openAIWebSearchUsesFixedInputBlock(model) {
		quantities[inputDimension] += int64(webSearchCalls * 8000)
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

		selected, err := selectOpenAIRate(price.RateList, dimension, usage.PromptTokens, tier, conditions)
		if err != nil {
			return platform.UsageCostBreakdown{}, err
		}

		cost := selected.UnitPriceUSD.Mul(decimal.NewFromInt(quantity))

		totalExact = totalExact.Add(cost)
		if strings.Contains(dimension, "output") {
			outputExact = outputExact.Add(cost)
		} else {
			inputExact = inputExact.Add(cost)
		}

		components = append(components, usageCostComponent(dimension, quantity, selected))
	}

	inputUSD, _ := inputExact.Float64()
	outputUSD, _ := outputExact.Float64()
	totalUSD, _ := totalExact.Float64()

	totalTokens := usage.TotalTokens
	if totalTokens <= 0 {
		totalTokens = usage.PromptTokens + usage.CompletionTokens
	}

	applied := map[string]string{"service_tier": tier}

	for key, value := range conditions {
		if strings.TrimSpace(value) != "" {
			applied[key] = value
		}
	}

	return platform.UsageCostBreakdown{
		Model: model, InputTokens: usage.PromptTokens,
		OutputTokens: usage.CompletionTokens, TotalTokens: totalTokens,
		InputCostUSD: inputUSD, OutputCostUSD: outputUSD,
		TotalCostUSD: totalUSD, ExchangeRate: exchangeRate,
		TotalCostIRR:               totalUSD * exchangeRate,
		CacheCreation5mInputTokens: promptCacheWrite,
		CacheReadInputTokens:       promptCached,
		ServerToolUse:              usage.ServerToolUse,
		ServiceTier:                tier,
		PricingConditions:          applied, PricingRevision: price.Revision,
		ModelRuntimeProfile: price.RuntimeProfile, Components: components,
		PricingSnapshot: usagePricingSnapshot(price),
	}, nil
}

func openAIWebSearchUsesFixedInputBlock(model string) bool {
	value := strings.ToLower(strings.TrimSpace(model))
	value = strings.TrimPrefix(value, "openai/")
	value = strings.TrimPrefix(value, "openai:")

	return value == "gpt-4o-mini" ||
		strings.HasPrefix(value, "gpt-4o-mini-") ||
		value == "gpt-4.1-mini" ||
		strings.HasPrefix(value, "gpt-4.1-mini-")
}

func hasOpenAIRate(
	rates []normalizedPricingRate,
	dimension string,
	promptTokens int,
	tier string,
	conditions map[string]string,
) bool {
	for _, rate := range rates {
		if rate.Dimension == dimension &&
			openAIRateConditionsMatch(rate.Condition, promptTokens, tier, conditions) {
			return true
		}
	}

	return false
}

func selectOpenAIRate(
	rates []normalizedPricingRate,
	dimension string,
	promptTokens int,
	tier string,
	conditions map[string]string,
) (normalizedPricingRate, error) {
	matches := make([]normalizedPricingRate, 0, 1)
	bestSpecificity := -1

	for _, rate := range rates {
		if rate.Dimension != dimension ||
			!openAIRateConditionsMatch(rate.Condition, promptTokens, tier, conditions) {
			continue
		}

		specificity := len(rate.Condition)
		if specificity > bestSpecificity {
			matches = matches[:0]
			bestSpecificity = specificity
		}

		if specificity == bestSpecificity {
			matches = append(matches, rate)
		}
	}

	if len(matches) != 1 {
		return normalizedPricingRate{}, fmt.Errorf(
			"%w: OpenAI rate selection for %s matched %d rates",
			ErrPricingRateUnavailable,
			dimension,
			len(matches),
		)
	}

	return matches[0], nil
}

func normalizeOpenAIServiceTier(value string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "default", "standard":
		return "standard", nil
	case "batch", "flex", "priority":
		return strings.ToLower(strings.TrimSpace(value)), nil
	case "fast":
		// OpenAI reports "priority" when the request uses the newer "fast" alias.
		return "priority", nil
	default:
		return "", &UnsupportedServiceTierError{
			Provider: "openai",
			Tier:     value,
		}
	}
}

func validateOpenAIServiceTier(model, requestedTier string, price priceInfo) error {
	if !strings.EqualFold(strings.TrimSpace(price.Provider), "openai") ||
		strings.TrimSpace(requestedTier) == "" {
		return nil
	}

	normalizedTier, err := normalizeOpenAIServiceTier(requestedTier)
	if err != nil {
		var tierErr *UnsupportedServiceTierError
		if errors.As(err, &tierErr) {
			tierErr.Model = openAIModelName(model, price)
		}

		return err
	}

	if price.RuntimeProfile == nil {
		return nil
	}

	if price.RuntimeProfile.Profile.AllowsServiceTier(normalizedTier) {
		return nil
	}

	return newUnsupportedServiceTierError(price, model, requestedTier)
}

func newUnsupportedServiceTierError(
	price priceInfo,
	model string,
	tier string,
) error {
	return &UnsupportedServiceTierError{
		Provider: price.Provider,
		Model:    openAIModelName(model, price),
		Tier:     tier,
	}
}

func openAIModelName(model string, price priceInfo) string {
	if price.RuntimeProfile != nil &&
		strings.TrimSpace(price.RuntimeProfile.Slug) != "" {
		return strings.TrimSpace(price.RuntimeProfile.Slug)
	}

	return normalizeCatalogPricingModelID(model, "openai")
}

func openAIRateConditionsMatch(
	rateConditions map[string]interface{},
	promptTokens int,
	tier string,
	conditions map[string]string,
) bool {
	for key, expected := range rateConditions {
		switch key {
		case "min_prompt_tokens":
			minimum, ok := numericCondition(expected)
			if !ok || int64(promptTokens) < minimum {
				return false
			}
		case "service_tier":
			if !strings.EqualFold(strings.TrimSpace(fmt.Sprint(expected)), tier) {
				return false
			}
		default:
			actual, ok := conditions[key]
			if !ok || !strings.EqualFold(strings.TrimSpace(actual), strings.TrimSpace(fmt.Sprint(expected))) {
				return false
			}
		}
	}

	return true
}

func numericCondition(value interface{}) (int64, bool) {
	switch typed := value.(type) {
	case float64:
		return int64(typed), typed >= 0 && typed == float64(int64(typed))
	case int:
		return int64(typed), typed >= 0
	case int64:
		return typed, typed >= 0
	case json.Number:
		out, err := typed.Int64()
		return out, err == nil && out >= 0
	default:
		out, err := strconv.ParseInt(strings.TrimSpace(fmt.Sprint(value)), 10, 64)
		return out, err == nil && out >= 0
	}
}
