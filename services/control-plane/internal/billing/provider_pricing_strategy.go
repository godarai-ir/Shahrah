package billing

import (
	"fmt"
	"strings"
	"time"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	typedmetadata "github.com/Delpak/shahrah/services/control-plane/internal/metadata"
	"github.com/Delpak/shahrah/services/control-plane/internal/providerregistry"
)

type pricingEstimateInput struct {
	request      contractsopenai.ChatCompletionRequest
	promptTokens int
	outputTokens int
	price        priceInfo
	exchangeRate float64
}

type pricingUsageInput struct {
	model        string
	usage        *contractsopenai.Usage
	price        priceInfo
	exchangeRate float64
	conditions   map[string]string
	metadata     typedmetadata.Object
}

type pricingImageInput struct {
	pricingUsageInput
	inputTokens          int
	outputTokens         int
	totalTokens          int
	providerCostUSDTicks *int64
}

type providerPricingStrategy interface {
	Estimate(input pricingEstimateInput) (platform.UsageCostBreakdown, error)
	FinalizeUsage(input pricingUsageInput) (platform.UsageCostBreakdown, error)
	FinalizeBatch(input pricingUsageInput) (platform.UsageCostBreakdown, error)
	FinalizeImage(input pricingImageInput) (platform.UsageCostBreakdown, error)
	MeteredUsageFailsClosed() bool
}

var providerPricingStrategies = map[providerregistry.PricingStrategyKind]providerPricingStrategy{
	providerregistry.PricingStrategyStandard:     standardPricingStrategy{},
	providerregistry.PricingStrategyOpenAI:       openAIPricingStrategy{},
	providerregistry.PricingStrategyAnthropic:    anthropicPricingStrategy{},
	providerregistry.PricingStrategyXAI:          xAIPricingStrategy{},
	providerregistry.PricingStrategyDeepSeek:     deepSeekPricingStrategy{},
	providerregistry.PricingStrategyZAI:          zaiPricingStrategy{},
	providerregistry.PricingStrategyMoonshot:     moonshotPricingStrategy{},
	providerregistry.PricingStrategyGoogleGemini: googleGeminiPricingStrategy{},
	providerregistry.PricingStrategyMistral:      mistralPricingStrategy{},
}

func pricingStrategy(provider string) providerPricingStrategy {
	definition, ok := providerregistry.Lookup(provider)
	if ok {
		if strategy := providerPricingStrategies[definition.PricingStrategy]; strategy != nil {
			return strategy
		}
	}

	return providerPricingStrategies[providerregistry.PricingStrategyStandard]
}

type standardPricingStrategy struct{}

func (standardPricingStrategy) Estimate(input pricingEstimateInput) (platform.UsageCostBreakdown, error) {
	return costBreakdown(
		input.request.Model,
		input.promptTokens,
		input.outputTokens,
		input.price,
		input.exchangeRate,
	), nil
}

func (standardPricingStrategy) FinalizeUsage(input pricingUsageInput) (platform.UsageCostBreakdown, error) {
	if input.usage == nil {
		return platform.UsageCostBreakdown{}, fmt.Errorf("%w: provider usage is missing", ErrPricingUnavailable)
	}

	return costBreakdown(
		input.model,
		input.usage.PromptTokens,
		input.usage.CompletionTokens,
		input.price,
		input.exchangeRate,
	), nil
}

func (standardPricingStrategy) FinalizeBatch(pricingUsageInput) (platform.UsageCostBreakdown, error) {
	return platform.UsageCostBreakdown{}, fmt.Errorf("%w: provider does not declare batch pricing", ErrPricingUnavailable)
}

func (standardPricingStrategy) FinalizeImage(input pricingImageInput) (platform.UsageCostBreakdown, error) {
	breakdown := costBreakdown(
		input.model,
		input.inputTokens,
		input.outputTokens,
		input.price,
		input.exchangeRate,
	)
	if input.totalTokens > 0 {
		breakdown.TotalTokens = input.totalTokens
	}

	return applyReportedProviderCost(breakdown, input.providerCostUSDTicks, input.exchangeRate, input.metadata), nil
}

func (standardPricingStrategy) MeteredUsageFailsClosed() bool {
	return false
}

type anthropicPricingStrategy struct {
	standardPricingStrategy
}

func (anthropicPricingStrategy) Estimate(input pricingEstimateInput) (platform.UsageCostBreakdown, error) {
	speed := strings.TrimSpace(input.request.Speed)

	geo := strings.TrimSpace(input.request.InferenceGeo)
	if !strings.EqualFold(speed, "fast") && !strings.EqualFold(geo, "us") {
		return standardPricingStrategy{}.Estimate(input)
	}

	prefix := ""
	conditions := map[string]string{}

	if strings.EqualFold(geo, "us") {
		prefix = "data_residency_"
		conditions["inference_geo"] = "us"
	}

	inputDimension := prefix + "input_tokens"

	outputDimension := prefix + "output_tokens"
	if strings.EqualFold(speed, "fast") {
		inputDimension = prefix + "speed_input_tokens"
		outputDimension = prefix + "speed_output_tokens"
		conditions["speed"] = "fast"
	}

	breakdown, err := anthropicTokenCostBreakdown(
		input.request.Model,
		input.promptTokens,
		input.outputTokens,
		input.price,
		input.exchangeRate,
		inputDimension,
		outputDimension,
	)
	if err != nil {
		return platform.UsageCostBreakdown{}, err
	}

	breakdown.PricingConditions = conditions

	return breakdown, nil
}

func (anthropicPricingStrategy) FinalizeUsage(input pricingUsageInput) (platform.UsageCostBreakdown, error) {
	return anthropicUsageCostBreakdown(
		input.model,
		input.usage,
		input.price,
		input.exchangeRate,
		false,
		input.conditions,
	)
}

func (anthropicPricingStrategy) FinalizeBatch(input pricingUsageInput) (platform.UsageCostBreakdown, error) {
	return anthropicUsageCostBreakdown(
		input.model,
		input.usage,
		input.price,
		input.exchangeRate,
		true,
		input.conditions,
	)
}

type openAIPricingStrategy struct {
	standardPricingStrategy
}

func (openAIPricingStrategy) Estimate(input pricingEstimateInput) (platform.UsageCostBreakdown, error) {
	return openAIUsageCostBreakdownWithPolicy(
		input.request.Model,
		usageFromEstimate(input),
		input.price,
		input.exchangeRate,
		map[string]string{"service_tier": input.request.ServiceTier},
		openAIStandardTokenPricingPolicy,
	)
}

func (openAIPricingStrategy) FinalizeUsage(input pricingUsageInput) (platform.UsageCostBreakdown, error) {
	return openAIUsageCostBreakdownWithPolicy(
		input.model,
		input.usage,
		input.price,
		input.exchangeRate,
		input.conditions,
		openAIStandardTokenPricingPolicy,
	)
}

func (openAIPricingStrategy) FinalizeBatch(input pricingUsageInput) (platform.UsageCostBreakdown, error) {
	usage := *input.usage
	usage.ServiceTier = "batch"

	if strings.Contains(strings.ToLower(input.model), "gpt-image") {
		return openAIImageCostBreakdown(
			input.model,
			&usage,
			usage.PromptTokens,
			usage.CompletionTokens,
			usage.TotalTokens,
			input.price,
			input.exchangeRate,
			input.metadata,
			"batch",
		)
	}

	return openAIUsageCostBreakdownWithPolicy(
		input.model,
		&usage,
		input.price,
		input.exchangeRate,
		input.conditions,
		openAIStandardTokenPricingPolicy,
	)
}

func (openAIPricingStrategy) FinalizeImage(input pricingImageInput) (platform.UsageCostBreakdown, error) {
	breakdown, err := openAIImageCostBreakdown(
		input.model,
		input.usage,
		input.inputTokens,
		input.outputTokens,
		input.totalTokens,
		input.price,
		input.exchangeRate,
		input.metadata,
		"standard",
	)
	if err != nil {
		return platform.UsageCostBreakdown{}, err
	}

	return applyReportedProviderCost(breakdown, input.providerCostUSDTicks, input.exchangeRate, input.metadata), nil
}

type xAIPricingStrategy struct {
	standardPricingStrategy
}

func (xAIPricingStrategy) FinalizeUsage(input pricingUsageInput) (platform.UsageCostBreakdown, error) {
	breakdown, err := standardPricingStrategy{}.FinalizeUsage(input)
	if err != nil {
		return platform.UsageCostBreakdown{}, err
	}

	breakdown, err = reconcileXAIProviderCost(breakdown, input.usage.CostInUSDTicks, input.exchangeRate)
	if err == nil {
		metadata := typedmetadata.Wrap(breakdown.Metadata)
		metadata.Set(billingMetadataServerSideToolsUsed, input.usage.NumServerSideToolsUsed)
	}

	return breakdown, err
}

func (xAIPricingStrategy) FinalizeBatch(input pricingUsageInput) (platform.UsageCostBreakdown, error) {
	var (
		breakdown platform.UsageCostBreakdown
		err       error
	)

	model := strings.ToLower(input.model)
	if strings.Contains(model, "gpt-image") {
		breakdown, err = openAIImageCostBreakdown(
			input.model,
			input.usage,
			input.usage.PromptTokens,
			input.usage.CompletionTokens,
			input.usage.TotalTokens,
			input.price,
			input.exchangeRate,
			input.metadata,
			"batch",
		)
	} else if strings.Contains(model, "grok-imagine-image") {
		breakdown, err = xAIImageCostBreakdown(
			input.model,
			input.usage.PromptTokens,
			input.usage.CompletionTokens,
			input.usage.TotalTokens,
			input.price,
			input.exchangeRate,
			input.metadata,
		)
		breakdown.Batch = true
	} else {
		usage := *input.usage
		usage.ServiceTier = "batch"
		breakdown, err = openAIUsageCostBreakdownWithPolicy(
			input.model,
			&usage,
			input.price,
			input.exchangeRate,
			input.conditions,
			xAITokenPricingPolicy,
		)
	}

	if err != nil {
		return platform.UsageCostBreakdown{}, err
	}

	return reconcileXAIProviderCost(breakdown, input.usage.CostInUSDTicks, input.exchangeRate)
}

func (xAIPricingStrategy) FinalizeImage(input pricingImageInput) (platform.UsageCostBreakdown, error) {
	breakdown, err := xAIImageCostBreakdown(
		input.model,
		input.inputTokens,
		input.outputTokens,
		input.totalTokens,
		input.price,
		input.exchangeRate,
		input.metadata,
	)
	if err != nil {
		return platform.UsageCostBreakdown{}, err
	}

	ticks := int64(0)
	if input.providerCostUSDTicks != nil {
		ticks = *input.providerCostUSDTicks
	} else if input.usage != nil {
		ticks = input.usage.CostInUSDTicks
	}

	return reconcileXAIProviderCost(breakdown, ticks, input.exchangeRate)
}

type deepSeekPricingStrategy struct {
	standardPricingStrategy
}

func (deepSeekPricingStrategy) Estimate(input pricingEstimateInput) (platform.UsageCostBreakdown, error) {
	usage := &contractsopenai.Usage{
		PromptTokens:          input.promptTokens,
		CompletionTokens:      input.outputTokens,
		TotalTokens:           input.promptTokens + input.outputTokens,
		PromptCacheMissTokens: input.promptTokens,
	}
	conditions := deepSeekPricingConditionsAt(time.Now().UTC(), input.price.RateList)

	return deepSeekUsageCostBreakdown(
		input.request.Model,
		usage,
		input.price,
		input.exchangeRate,
		conditions,
	)
}

func (deepSeekPricingStrategy) FinalizeUsage(input pricingUsageInput) (platform.UsageCostBreakdown, error) {
	return deepSeekUsageCostBreakdown(
		input.model,
		input.usage,
		input.price,
		input.exchangeRate,
		input.conditions,
	)
}

type moonshotPricingStrategy struct {
	standardPricingStrategy
}

type zaiPricingStrategy struct {
	standardPricingStrategy
}

func (zaiPricingStrategy) FinalizeUsage(input pricingUsageInput) (platform.UsageCostBreakdown, error) {
	return moonshotUsageCostBreakdown(input.model, input.usage, input.price, input.exchangeRate)
}

func (moonshotPricingStrategy) FinalizeUsage(input pricingUsageInput) (platform.UsageCostBreakdown, error) {
	return moonshotUsageCostBreakdown(input.model, input.usage, input.price, input.exchangeRate)
}

type googleGeminiPricingStrategy struct {
	standardPricingStrategy
}

func (googleGeminiPricingStrategy) Estimate(input pricingEstimateInput) (platform.UsageCostBreakdown, error) {
	tier, err := googleGeminiServiceTier(input.request.ServiceTier)
	if err != nil {
		return platform.UsageCostBreakdown{}, err
	}

	if input.price.RuntimeProfile != nil && !input.price.RuntimeProfile.Profile.AllowsServiceTier(tier) {
		return platform.UsageCostBreakdown{}, fmt.Errorf(
			"%w: model does not support Gemini service tier %s",
			ErrPricingUnavailable,
			tier,
		)
	}

	usage := usageFromEstimate(input)
	usage.ServiceTier = tier

	return googleGeminiUsageCostBreakdown(
		input.request.Model,
		usage,
		input.price,
		input.exchangeRate,
		map[string]string{"service_tier": tier},
	)
}

func (googleGeminiPricingStrategy) FinalizeUsage(input pricingUsageInput) (platform.UsageCostBreakdown, error) {
	return googleGeminiUsageCostBreakdown(input.model, input.usage, input.price, input.exchangeRate, input.conditions)
}

func (googleGeminiPricingStrategy) FinalizeBatch(input pricingUsageInput) (platform.UsageCostBreakdown, error) {
	usage := *input.usage
	usage.ServiceTier = "batch"
	conditions := clonePricingConditions(input.conditions)
	conditions["service_tier"] = "batch"

	return googleGeminiUsageCostBreakdown(
		input.model,
		&usage,
		input.price,
		input.exchangeRate,
		conditions,
	)
}

func (googleGeminiPricingStrategy) FinalizeImage(input pricingImageInput) (platform.UsageCostBreakdown, error) {
	breakdown, err := googleGeminiImageCostBreakdown(
		input.model,
		input.inputTokens,
		input.outputTokens,
		input.totalTokens,
		input.price,
		input.exchangeRate,
		input.metadata,
	)
	if err != nil {
		return platform.UsageCostBreakdown{}, err
	}

	return applyReportedProviderCost(breakdown, input.providerCostUSDTicks, input.exchangeRate, input.metadata), nil
}

type mistralPricingStrategy struct {
	standardPricingStrategy
}

func (mistralPricingStrategy) Estimate(input pricingEstimateInput) (platform.UsageCostBreakdown, error) {
	return mistralUsageCostBreakdown(
		input.request.Model,
		usageFromEstimate(input),
		input.price,
		input.exchangeRate,
		map[string]string{"service_tier": input.request.ServiceTier},
	)
}

func (mistralPricingStrategy) FinalizeUsage(input pricingUsageInput) (platform.UsageCostBreakdown, error) {
	return mistralUsageCostBreakdown(input.model, input.usage, input.price, input.exchangeRate, input.conditions)
}

func (mistralPricingStrategy) FinalizeBatch(input pricingUsageInput) (platform.UsageCostBreakdown, error) {
	usage := *input.usage
	usage.ServiceTier = "batch"

	return mistralUsageCostBreakdown(input.model, &usage, input.price, input.exchangeRate, input.conditions)
}

func (mistralPricingStrategy) MeteredUsageFailsClosed() bool {
	return true
}

func usageFromEstimate(input pricingEstimateInput) *contractsopenai.Usage {
	return &contractsopenai.Usage{
		PromptTokens:     input.promptTokens,
		CompletionTokens: input.outputTokens,
		TotalTokens:      input.promptTokens + input.outputTokens,
		ServiceTier:      input.request.ServiceTier,
	}
}

func clonePricingConditions(input map[string]string) map[string]string {
	output := make(map[string]string, len(input)+1)
	for key, value := range input {
		output[key] = value
	}

	return output
}

func applyReportedProviderCost(
	breakdown platform.UsageCostBreakdown,
	providerCostUSDTicks *int64,
	exchangeRate float64,
	metadata typedmetadata.Object,
) platform.UsageCostBreakdown {
	if providerCostUSDTicks == nil || *providerCostUSDTicks < 0 {
		return breakdown
	}

	totalUSD := float64(*providerCostUSDTicks) / usdTicksPerDollar
	breakdown = applyProviderImageCost(breakdown, totalUSD)
	breakdown.ExchangeRate = exchangeRate

	breakdown.TotalCostIRR = totalUSD * exchangeRate
	if breakdown.Metadata == nil {
		breakdown.Metadata = metadata.CloneValues()
	}

	resultMetadata := typedmetadata.Wrap(breakdown.Metadata)
	resultMetadata.Set(billingMetadataProviderCostUSDTicks, *providerCostUSDTicks)
	resultMetadata.Set(billingMetadataProviderCostExact, true)

	return breakdown
}
