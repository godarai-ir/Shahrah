package billing

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/shopspring/decimal"
)

func imageUsageTokenBreakdown(usage *openai.Usage, estimated platform.UsageCostBreakdown) (int, int, int) {
	if usage == nil {
		return estimated.InputTokens, estimated.OutputTokens, estimated.TotalTokens
	}

	inputTokens := usage.PromptTokens
	outputTokens := usage.CompletionTokens

	totalTokens := usage.TotalTokens
	if totalTokens <= 0 {
		totalTokens = inputTokens + outputTokens
	}

	if inputTokens == 0 && outputTokens == 0 && totalTokens > 0 {
		if estimated.InputTokens > 0 && totalTokens > estimated.InputTokens {
			inputTokens = estimated.InputTokens
			outputTokens = totalTokens - inputTokens
		} else {
			outputTokens = totalTokens
		}
	}

	if totalTokens < inputTokens+outputTokens {
		totalTokens = inputTokens + outputTokens
	}

	return inputTokens, outputTokens, totalTokens
}

func xAIImageCostBreakdown(
	model string,
	inputTokens int,
	outputTokens int,
	totalTokens int,
	price priceInfo,
	exchangeRate float64,
	metadata map[string]interface{},
) (platform.UsageCostBreakdown, error) {
	inputImages := intFromBillingMetadata(metadata["image_count"])

	outputImages := intFromBillingMetadata(metadata["n"])
	if outputImages <= 0 {
		outputImages = 1
	}

	resolution := strings.ToLower(strings.TrimSpace(stringFromBillingMetadata(metadata["resolution"])))
	if resolution == "" {
		resolution = "1k"
	}

	components := make([]platform.UsageCostComponent, 0, 2)
	inputExact := decimal.Zero

	if inputImages > 0 {
		rate, ok := selectXAIImageRate(price, "input_images", "")
		if !ok {
			return platform.UsageCostBreakdown{}, fmt.Errorf("%w: missing input image rate", ErrPricingUnavailable)
		}

		inputExact = rate.UnitPriceUSD.Mul(decimal.NewFromInt(int64(inputImages)))
		components = append(components, usageCostComponent("input_images", int64(inputImages), rate))
	}

	outputRate, ok := selectXAIImageRate(price, "output_images", resolution)
	if !ok {
		return platform.UsageCostBreakdown{}, fmt.Errorf("%w: missing %s output image rate", ErrPricingUnavailable, resolution)
	}

	outputExact := outputRate.UnitPriceUSD.Mul(decimal.NewFromInt(int64(outputImages)))
	components = append(components, usageCostComponent("output_images", int64(outputImages), outputRate))
	inputUSD, _ := inputExact.Float64()
	outputUSD, _ := outputExact.Float64()
	totalUSD := inputUSD + outputUSD

	return platform.UsageCostBreakdown{
		Model: model, Metadata: cloneUsageMetadata(metadata),
		InputTokens: inputTokens, OutputTokens: outputTokens, TotalTokens: totalTokens,
		InputCostUSD: inputUSD, OutputCostUSD: outputUSD, TotalCostUSD: totalUSD,
		ExchangeRate: exchangeRate, TotalCostIRR: totalUSD * exchangeRate,
		PricingRevision: price.Revision, ModelRuntimeProfile: price.RuntimeProfile,
		Components: components, PricingSnapshot: usagePricingSnapshot(price),
		Quantities: map[string]string{"input_images": strconv.Itoa(inputImages), "output_images": strconv.Itoa(outputImages)},
	}, nil
}

func selectXAIImageRate(price priceInfo, dimension, resolution string) (normalizedPricingRate, bool) {
	for _, rate := range price.RateList {
		if rate.Dimension != dimension {
			continue
		}

		conditionResolution := strings.ToLower(strings.TrimSpace(fmt.Sprint(rate.Condition["resolution"])))
		if resolution == "" && conditionResolution == "" {
			return rate, true
		}

		if resolution != "" && conditionResolution == resolution {
			return rate, true
		}
	}

	rate, ok := price.Rates[dimension]

	return rate, ok && (resolution == "" || len(rate.Condition) == 0)
}

func reconcileXAIProviderCost(breakdown platform.UsageCostBreakdown, ticks int64, exchangeRate float64) (platform.UsageCostBreakdown, error) {
	if ticks <= 0 {
		return platform.UsageCostBreakdown{}, fmt.Errorf("%w: xAI response omitted cost_in_usd_ticks", ErrPricingUnavailable)
	}

	catalogTotalUSD := breakdown.TotalCostUSD
	exactUSD := decimal.NewFromInt(ticks).Div(decimal.NewFromInt(int64(usdTicksPerDollar)))
	totalUSD, _ := exactUSD.Float64()
	delta := exactUSD.Sub(decimal.NewFromFloat(catalogTotalUSD))

	breakdown = applyProviderImageCost(breakdown, totalUSD)
	if !delta.IsZero() {
		breakdown.Components = append(breakdown.Components, platform.UsageCostComponent{
			Dimension:    "provider_cost_reconciliation",
			Quantity:     "1",
			Unit:         "request",
			UnitPriceUSD: delta.String(),
			CostUSD:      delta.String(),
			Condition:    map[string]interface{}{"source": "cost_in_usd_ticks"},
		})
	}

	breakdown.ExchangeRate = exchangeRate

	breakdown.TotalCostIRR = totalUSD * exchangeRate
	if breakdown.Metadata == nil {
		breakdown.Metadata = map[string]interface{}{}
	}

	breakdown.Metadata["provider_cost_in_usd_ticks"] = ticks
	breakdown.Metadata["provider_cost_usd"] = exactUSD.String()
	breakdown.Metadata["catalog_estimated_cost_usd"] = decimal.NewFromFloat(catalogTotalUSD).String()
	breakdown.Metadata["provider_cost_reconciliation_delta_usd"] = delta.String()
	breakdown.Metadata["provider_cost_exact"] = true

	return breakdown, nil
}

func cloneUsageMetadata(input map[string]interface{}) map[string]interface{} {
	output := make(map[string]interface{}, len(input)+2)
	for key, value := range input {
		output[key] = value
	}

	return output
}

func mergePricingMetadata(base map[string]interface{}, provider map[string]interface{}) map[string]interface{} {
	output := cloneUsageMetadata(base)
	for key, value := range provider {
		output[key] = value
	}

	return output
}

func stringFromBillingMetadata(value interface{}) string {
	text, _ := value.(string) //nolint:errcheck // Optional metadata uses the zero value when absent or mismatched.
	return text
}

func openAIImageCostBreakdown(
	model string,
	usage *openai.Usage,
	inputTokens int,
	outputTokens int,
	totalTokens int,
	price priceInfo,
	exchangeRate float64,
	metadata map[string]interface{},
	serviceTier string,
) (platform.UsageCostBreakdown, error) {
	textInputTokens := inputTokens
	imageInputTokens := 0
	cachedTextTokens := 0
	cachedImageTokens := 0

	if usage != nil && usage.PromptTokensDetails != nil {
		details := usage.PromptTokensDetails
		textInputTokens = details.TextTokens
		imageInputTokens = details.ImageTokens
		cachedTextTokens = details.CachedTextTokens

		cachedImageTokens = details.CachedImageTokens
		if cachedTextTokens+cachedImageTokens == 0 && details.CachedTokens > 0 {
			if imageInputTokens > 0 && textInputTokens == 0 {
				cachedImageTokens = details.CachedTokens
			} else {
				cachedTextTokens = details.CachedTokens
			}
		}
	}

	if imageInputTokens == 0 && intFromBillingMetadata(metadata["image_count"]) > 0 {
		estimatedImageTokens := intFromBillingMetadata(metadata["estimated_image_input_tokens"])
		if estimatedImageTokens > inputTokens {
			estimatedImageTokens = inputTokens
		}

		imageInputTokens = estimatedImageTokens
		textInputTokens = inputTokens - imageInputTokens
	}

	if textInputTokens+imageInputTokens == 0 {
		textInputTokens = inputTokens
	}

	quantities := map[string]int64{
		"input_tokens":            int64(max(textInputTokens-cachedTextTokens, 0)),
		"image_input_tokens":      int64(max(imageInputTokens-cachedImageTokens, 0)),
		"cache_read_tokens":       int64(cachedTextTokens),
		"image_cache_read_tokens": int64(cachedImageTokens),
		"image_output_tokens":     int64(outputTokens),
	}
	if strings.EqualFold(serviceTier, "batch") {
		quantities = map[string]int64{
			"batch_input_tokens":            int64(max(textInputTokens-cachedTextTokens, 0)),
			"batch_image_input_tokens":      int64(max(imageInputTokens-cachedImageTokens, 0)),
			"batch_cache_read_tokens":       int64(cachedTextTokens),
			"batch_image_cache_read_tokens": int64(cachedImageTokens),
			"batch_image_output_tokens":     int64(outputTokens),
		}
	}

	dimensions := make([]string, 0, len(quantities))
	for dimension, quantity := range quantities {
		if quantity > 0 {
			dimensions = append(dimensions, dimension)
		}
	}

	sort.Strings(dimensions)

	totalExact, inputExact, outputExact := decimal.Zero, decimal.Zero, decimal.Zero

	components := make([]platform.UsageCostComponent, 0, len(dimensions))
	for _, dimension := range dimensions {
		rate, ok := price.Rates[dimension]
		if !ok {
			fallback := dimension
			switch dimension {
			case "image_input_tokens", "batch_image_input_tokens":
				fallback = "input_tokens"
			case "image_cache_read_tokens", "batch_image_cache_read_tokens":
				fallback = "cache_read_tokens"
			case "image_output_tokens", "batch_image_output_tokens":
				fallback = "output_tokens"
			}

			rate, ok = price.Rates[fallback]
		}

		if !ok {
			return platform.UsageCostBreakdown{}, fmt.Errorf(
				"%w: missing OpenAI image rate for %s",
				ErrPricingUnavailable,
				dimension,
			)
		}

		quantity := quantities[dimension]
		cost := rate.UnitPriceUSD.Mul(decimal.NewFromInt(quantity))

		totalExact = totalExact.Add(cost)
		if dimension == "image_output_tokens" || dimension == "batch_image_output_tokens" {
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

	if totalTokens <= 0 {
		totalTokens = inputTokens + outputTokens
	}

	return platform.UsageCostBreakdown{
		Model: model, InputTokens: inputTokens, OutputTokens: outputTokens,
		TotalTokens: totalTokens, InputCostUSD: inputUSD,
		OutputCostUSD: outputUSD, TotalCostUSD: totalUSD,
		ExchangeRate: exchangeRate, TotalCostIRR: totalUSD * exchangeRate,
		ServiceTier:     serviceTier,
		PricingRevision: price.Revision, ModelRuntimeProfile: price.RuntimeProfile,
		Components: components, PricingSnapshot: usagePricingSnapshot(price),
	}, nil
}

func intFromBillingMetadata(value interface{}) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	case json.Number:
		parsed, _ := typed.Int64() //nolint:errcheck // Failure is intentionally handled by the surrounding best-effort path.
		return int(parsed)
	default:
		return 0
	}
}

func googleGeminiImageCostBreakdown(
	model string,
	inputTokens int,
	outputTokens int,
	totalTokens int,
	price priceInfo,
	exchangeRate float64,
	metadata map[string]interface{},
) (platform.UsageCostBreakdown, error) {
	quantities := make(map[string]int64)

	if inputTokens > 0 {
		if _, ok := price.Rates["input_tokens"]; ok {
			quantities["input_tokens"] = int64(inputTokens)
		}
	}

	if outputTokens > 0 {
		if _, ok := price.Rates["image_output_tokens"]; ok {
			quantities["image_output_tokens"] = int64(outputTokens)
		} else if _, ok := price.Rates["output_tokens"]; ok {
			quantities["output_tokens"] = int64(outputTokens)
		}
	}

	imageCount := int64(1)

	if raw, ok := metadata["n"]; ok {
		switch value := raw.(type) {
		case int:
			imageCount = int64(value)
		case float64:
			imageCount = int64(value)
		}
	}

	if imageCount < 1 {
		imageCount = 1
	}

	if _, hasTokenRate := quantities["image_output_tokens"]; !hasTokenRate {
		dimension := googleGeminiImageDimension(price, metadata)
		if dimension != "" {
			quantities[dimension] = imageCount
			delete(quantities, "output_tokens")
		}
	}

	if len(quantities) == 0 {
		return platform.UsageCostBreakdown{}, fmt.Errorf(
			"%w: Gemini image pricing is unavailable",
			ErrPricingUnavailable,
		)
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

		rate, ok := price.Rates[dimension]
		if !ok {
			return platform.UsageCostBreakdown{}, fmt.Errorf(
				"%w: missing Gemini image rate for %s",
				ErrPricingUnavailable,
				dimension,
			)
		}

		cost := rate.UnitPriceUSD.Mul(decimal.NewFromInt(quantity))

		totalExact = totalExact.Add(cost)
		if dimension == "input_tokens" || strings.HasPrefix(dimension, "input_") {
			inputExact = inputExact.Add(cost)
		} else {
			outputExact = outputExact.Add(cost)
		}

		components = append(
			components,
			usageCostComponent(dimension, quantity, rate),
		)
	}

	inputUSD, _ := inputExact.Float64()
	outputUSD, _ := outputExact.Float64()
	totalUSD, _ := totalExact.Float64()

	if totalTokens <= 0 {
		totalTokens = inputTokens + outputTokens
	}

	return platform.UsageCostBreakdown{
		Model:               model,
		InputTokens:         inputTokens,
		OutputTokens:        outputTokens,
		TotalTokens:         totalTokens,
		InputCostUSD:        inputUSD,
		OutputCostUSD:       outputUSD,
		TotalCostUSD:        totalUSD,
		ExchangeRate:        exchangeRate,
		TotalCostIRR:        totalUSD * exchangeRate,
		PricingRevision:     price.Revision,
		ModelRuntimeProfile: price.RuntimeProfile,
		Components:          components,
		PricingSnapshot:     usagePricingSnapshot(price),
	}, nil
}

func googleGeminiImageDimension(
	price priceInfo,
	metadata map[string]interface{},
) string {
	size, _ := metadata["size"].(string) //nolint:errcheck // Optional metadata uses the zero value when absent or mismatched.

	size = strings.ToLower(strings.TrimSpace(size))
	switch {
	case strings.Contains(size, "4096") || strings.Contains(size, "4k"):
		if _, ok := price.Rates["output_images_4k"]; ok {
			return "output_images_4k"
		}
	case strings.Contains(size, "2048") || strings.Contains(size, "2k"):
		if _, ok := price.Rates["output_images_2k"]; ok {
			return "output_images_2k"
		}

		if _, ok := price.Rates["output_images_1k_2k"]; ok {
			return "output_images_1k_2k"
		}
	case strings.Contains(size, "512") || strings.Contains(size, "0.5k"):
		if _, ok := price.Rates["output_images_0_5k"]; ok {
			return "output_images_0_5k"
		}
	default:
		if _, ok := price.Rates["output_images_1k"]; ok {
			return "output_images_1k"
		}

		if _, ok := price.Rates["output_images_1k_2k"]; ok {
			return "output_images_1k_2k"
		}
	}

	if _, ok := price.Rates["output_images"]; ok {
		return "output_images"
	}

	return ""
}

func applyProviderImageCost(breakdown platform.UsageCostBreakdown, totalUSD float64) platform.UsageCostBreakdown {
	breakdown.TotalCostUSD = totalUSD
	if breakdown.InputCostUSD+breakdown.OutputCostUSD > 0 {
		pricedTotal := breakdown.InputCostUSD + breakdown.OutputCostUSD
		breakdown.InputCostUSD = totalUSD * (breakdown.InputCostUSD / pricedTotal)
		breakdown.OutputCostUSD = totalUSD - breakdown.InputCostUSD

		return breakdown
	}

	breakdown.InputCostUSD = 0
	breakdown.OutputCostUSD = totalUSD

	return breakdown
}
