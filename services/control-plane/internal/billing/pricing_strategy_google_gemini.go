package billing

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/shopspring/decimal"
)

func googleGeminiServiceTier(value string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "default", "standard":
		return "standard", nil
	case "flex":
		return "flex", nil
	case "priority":
		return "priority", nil
	default:
		return "", fmt.Errorf(
			"%w: unsupported Gemini service tier %q",
			ErrPricingUnavailable,
			value,
		)
	}
}

func googleGeminiUsageCostBreakdown(
	model string,
	usage *openai.Usage,
	price priceInfo,
	exchangeRate float64,
	conditions map[string]string,
) (platform.UsageCostBreakdown, error) {
	tierValue := usage.ServiceTier
	if strings.TrimSpace(tierValue) == "" {
		tierValue = conditions["service_tier"]
	}

	tier, err := googleGeminiServiceTier(tierValue)
	if err != nil {
		return platform.UsageCostBreakdown{}, err
	}

	promptAudio := 0
	promptCached := 0

	if usage.PromptTokensDetails != nil {
		promptAudio = usage.PromptTokensDetails.AudioTokens
		promptCached = usage.PromptTokensDetails.CachedTokens
	}

	completionAudio := 0
	completionImage := 0

	if usage.CompletionTokensDetails != nil {
		completionAudio = usage.CompletionTokensDetails.AudioTokens
		completionImage = usage.CompletionTokensDetails.ImageTokens
	}

	if promptAudio < 0 || promptCached < 0 || completionAudio < 0 ||
		promptAudio+promptCached > usage.PromptTokens ||
		completionAudio+completionImage > usage.CompletionTokens {
		return platform.UsageCostBreakdown{}, fmt.Errorf(
			"%w: invalid Gemini token detail totals",
			ErrPricingUnavailable,
		)
	}
	// The OpenAI-compatible usage schema does not identify the modality of
	// cached tokens. Refuse an ambiguous mixed audio/cache bill instead of
	// silently applying the cheaper text cache rate.
	if promptAudio > 0 && promptCached > 0 {
		return platform.UsageCostBreakdown{}, fmt.Errorf(
			"%w: Gemini audio cache modality is ambiguous",
			ErrPricingUnavailable,
		)
	}

	prefix := ""
	if tier != "standard" {
		prefix = tier + "_"
	}

	if usage.PromptTokens > 200000 {
		prefix += "long_context_"
	}

	inputDimension := prefix + "input_tokens"
	audioInputDimension := prefix + "audio_input_tokens"

	textInputTokens := usage.PromptTokens - promptAudio - promptCached
	if promptAudio > 0 {
		if _, hasAudioRate := price.Rates[audioInputDimension]; !hasAudioRate {
			// A missing audio-specific dimension means the official table uses
			// one published input rate for every supported input modality.
			textInputTokens += promptAudio
			promptAudio = 0
		}
	}

	quantities := map[string]int64{
		inputDimension:                 int64(textInputTokens),
		prefix + "output_tokens":       int64(usage.CompletionTokens - completionAudio - completionImage),
		audioInputDimension:            int64(promptAudio),
		prefix + "audio_output_tokens": int64(completionAudio),
		prefix + "image_output_tokens": int64(completionImage),
		prefix + "cache_read_tokens":   int64(promptCached),
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

	totalExact := decimal.Zero
	inputExact := decimal.Zero
	outputExact := decimal.Zero

	components := make([]platform.UsageCostComponent, 0, len(dimensions))
	for _, dimension := range dimensions {
		quantity := quantities[dimension]
		if quantity == 0 {
			continue
		}

		rate, ok := price.Rates[dimension]
		if !ok {
			return platform.UsageCostBreakdown{}, fmt.Errorf(
				"%w: missing Gemini rate for %s",
				ErrPricingUnavailable,
				dimension,
			)
		}

		cost := rate.UnitPriceUSD.Mul(decimal.NewFromInt(quantity))

		totalExact = totalExact.Add(cost)
		if strings.Contains(dimension, "output_") {
			outputExact = outputExact.Add(cost)
		} else {
			inputExact = inputExact.Add(cost)
		}

		components = append(
			components,
			usageCostComponent(dimension, quantity, rate),
		)
	}

	inputUSD, _ := inputExact.Float64()
	outputUSD, _ := outputExact.Float64()
	totalUSD, _ := totalExact.Float64()

	appliedConditions := make(map[string]string, len(conditions)+1)
	for key, value := range conditions {
		appliedConditions[key] = value
	}

	appliedConditions["service_tier"] = tier

	totalTokens := usage.TotalTokens
	if totalTokens <= 0 {
		totalTokens = usage.PromptTokens + usage.CompletionTokens
	}

	return platform.UsageCostBreakdown{
		Model:                model,
		InputTokens:          usage.PromptTokens,
		OutputTokens:         usage.CompletionTokens,
		TotalTokens:          totalTokens,
		InputCostUSD:         inputUSD,
		OutputCostUSD:        outputUSD,
		TotalCostUSD:         totalUSD,
		ExchangeRate:         exchangeRate,
		TotalCostIRR:         totalUSD * exchangeRate,
		CacheReadInputTokens: promptCached,
		ServiceTier:          tier,
		PricingConditions:    appliedConditions,
		PricingRevision:      price.Revision,
		ModelRuntimeProfile:  price.RuntimeProfile,
		Components:           components,
		PricingSnapshot:      usagePricingSnapshot(price),
	}, nil
}
