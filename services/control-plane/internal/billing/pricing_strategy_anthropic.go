package billing

import (
	"fmt"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	prommetrics "github.com/Delpak/shahrah/services/control-plane/internal/observability/prometheus"
	"github.com/shopspring/decimal"
)

func anthropicUsageCostBreakdown(
	model string,
	usage *openai.Usage,
	price priceInfo,
	exchangeRate float64,
	batch bool,
	conditions map[string]string,
) (platform.UsageCostBreakdown, error) {
	inputDimension := "input_tokens"
	outputDimension := "output_tokens"
	cache5mDimension := "cache_write_5m_tokens"
	cache1hDimension := "cache_write_1h_tokens"
	cacheReadDimension := "cache_read_tokens"

	if batch {
		inputDimension = "batch_input_tokens"
		outputDimension = "batch_output_tokens"
		cache5mDimension = "batch_cache_write_5m_tokens"
		cache1hDimension = "batch_cache_write_1h_tokens"
		cacheReadDimension = "batch_cache_read_tokens"
	}

	if strings.EqualFold(strings.TrimSpace(usage.Speed), "fast") {
		if batch {
			return platform.UsageCostBreakdown{}, fmt.Errorf(
				"%w: Anthropic fast-mode is unavailable with Batch",
				ErrPricingUnavailable,
			)
		}

		inputDimension = "speed_input_tokens"
		outputDimension = "speed_output_tokens"
		cache5mDimension = "speed_cache_write_5m_tokens"
		cache1hDimension = "speed_cache_write_1h_tokens"
		cacheReadDimension = "speed_cache_read_tokens"
	}

	if strings.EqualFold(conditions["inference_geo"], "us") {
		inputDimension = "data_residency_" + inputDimension
		outputDimension = "data_residency_" + outputDimension
		cache5mDimension = "data_residency_" + cache5mDimension
		cache1hDimension = "data_residency_" + cache1hDimension
		cacheReadDimension = "data_residency_" + cacheReadDimension
	}

	quantities := map[string]int64{
		inputDimension:     int64(usage.PromptTokens),
		outputDimension:    int64(usage.CompletionTokens),
		cache5mDimension:   int64(usage.CacheCreation5mInputTokens),
		cache1hDimension:   int64(usage.CacheCreation1hInputTokens),
		cacheReadDimension: int64(usage.CacheReadInputTokens),
	}
	if usage.CacheCreationInputTokens > 0 &&
		quantities[cache5mDimension] == 0 &&
		quantities[cache1hDimension] == 0 {
		quantities[cache5mDimension] = int64(
			usage.CacheCreationInputTokens,
		)
	}

	for tool, quantity := range usage.ServerToolUse {
		quantities[tool] = int64(quantity)
	}

	totalExact, inputExact, outputExact, components, err := accumulateUsageCostDimensions(
		quantities,
		func(dimension string) bool { return dimension == outputDimension },
		func(dimension string, quantity int64) (normalizedPricingRate, error) {
			rate, ok := price.Rates[dimension]
			if dimension == "code_execution_requests" {
				rate, ok = anthropicCodeExecutionUsageRate(price, conditions)
				if !ok {
					return normalizedPricingRate{}, fmt.Errorf(
						"%w: billable code execution requires "+
							"provider cost reconciliation",
						ErrPricingUnavailable,
					)
				}
			}

			if !ok {
				return normalizedPricingRate{}, fmt.Errorf(
					"%w: missing rate for %s",
					ErrPricingUnavailable,
					dimension,
				)
			}

			prommetrics.RecordAnthropicUsageDimension(dimension, quantity)

			return rate, nil
		},
	)
	if err != nil {
		return platform.UsageCostBreakdown{}, err
	}

	inputUSD, _ := inputExact.Float64()
	outputUSD, _ := outputExact.Float64()
	totalUSD, _ := totalExact.Float64()

	return platform.UsageCostBreakdown{
		Model:                      model,
		InputTokens:                usage.PromptTokens,
		OutputTokens:               usage.CompletionTokens,
		TotalTokens:                usage.TotalTokens,
		InputCostUSD:               inputUSD,
		OutputCostUSD:              outputUSD,
		TotalCostUSD:               totalUSD,
		ExchangeRate:               exchangeRate,
		TotalCostIRR:               totalUSD * exchangeRate,
		CacheCreation5mInputTokens: usage.CacheCreation5mInputTokens,
		CacheCreation1hInputTokens: usage.CacheCreation1hInputTokens,
		CacheReadInputTokens:       usage.CacheReadInputTokens,
		ServerToolUse:              usage.ServerToolUse,
		ServiceTier:                usage.ServiceTier,
		Speed:                      usage.Speed,
		Batch:                      batch,
		PricingConditions:          conditions,
		PricingRevision:            price.Revision,
		ModelRuntimeProfile:        price.RuntimeProfile,
		Components:                 components,
		PricingSnapshot:            usagePricingSnapshot(price),
	}, nil
}

func anthropicCodeExecutionUsageRate(
	price priceInfo,
	conditions map[string]string,
) (normalizedPricingRate, bool) {
	if conditions[platform.AnthropicCodeExecutionPricingCondition] !=
		platform.AnthropicCodeExecutionFreeWithCurrentWebTools {
		return normalizedPricingRate{}, false
	}

	hourlyRate, ok := price.Rates["code_execution_container_hours"]
	if !ok || hourlyRate.Unit != "container_hour" {
		return normalizedPricingRate{}, false
	}

	return normalizedPricingRate{
		Dimension: "code_execution_requests",
		Unit:      "request",
		Condition: map[string]interface{}{
			platform.AnthropicCodeExecutionPricingCondition: platform.AnthropicCodeExecutionFreeWithCurrentWebTools,
			"source_dimension": hourlyRate.Dimension,
		},
		UnitPriceUSD: decimal.Zero,
		Region:       hourlyRate.Region,
		SourceURL:    hourlyRate.SourceURL,
	}, true
}

const usdTicksPerDollar = 10_000_000_000.0
