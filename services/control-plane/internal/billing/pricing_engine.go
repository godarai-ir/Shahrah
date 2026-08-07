package billing

import (
	"context"
	"fmt"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	typedmetadata "github.com/Delpak/shahrah/services/control-plane/internal/metadata"
	"github.com/Delpak/shahrah/services/control-plane/internal/providerregistry"
	"github.com/shopspring/decimal"
)

// EstimateCost estimates tokens and cost for a request.
func (s *PricingEngine) EstimateCost(ctx context.Context, req openai.ChatCompletionRequest) (platform.UsageCostBreakdown, error) {
	rate := s.exchangeRate(ctx)

	promptTokens := estimatePromptTokens(req)

	outputTokens := 256
	if req.MaxTokens != nil && *req.MaxTokens > 0 {
		outputTokens = *req.MaxTokens
	}

	price, err := s.modelPrice(ctx, req.Model)
	if err != nil {
		return platform.UsageCostBreakdown{}, err
	}

	if err := validateOpenAIServiceTier(req.Model, req.ServiceTier, price); err != nil {
		return platform.UsageCostBreakdown{}, err
	}

	return pricingStrategy(price.Provider).Estimate(pricingEstimateInput{
		request:      req,
		promptTokens: promptTokens,
		outputTokens: outputTokens,
		price:        price,
		exchangeRate: rate.USDtoIRR,
	})
}

// EstimateEmbeddingCost estimates cost for an embeddings request.
func (s *PricingEngine) EstimateEmbeddingCost(ctx context.Context, req openai.EmbeddingRequest) (platform.UsageCostBreakdown, error) {
	rate := s.exchangeRate(ctx)

	promptTokens := estimateEmbeddingPromptTokens(req.Input)

	price, err := s.modelPrice(ctx, req.Model)
	if err != nil {
		return platform.UsageCostBreakdown{}, err
	}

	return costBreakdown(req.Model, promptTokens, 0, price, rate.USDtoIRR), nil
}

// FinalizeFromUsage recalculates final spend from provider usage.
func (s *PricingEngine) FinalizeFromUsage(ctx context.Context, model string, usage *openai.Usage, estimated platform.UsageCostBreakdown) (platform.UsageCostBreakdown, error) {
	if usage == nil {
		return estimated, nil
	}

	rate := s.exchangeRate(ctx)

	price, err := s.pinnedOrCurrentModelPrice(ctx, model, estimated)
	if err != nil {
		return platform.UsageCostBreakdown{}, err
	}

	breakdown, err := pricingStrategy(price.Provider).FinalizeUsage(pricingUsageInput{
		model:        model,
		usage:        usage,
		price:        price,
		exchangeRate: rate.USDtoIRR,
		conditions:   estimated.PricingConditions,
		metadata:     typedmetadata.Wrap(estimated.Metadata),
	})
	if err != nil {
		return platform.UsageCostBreakdown{}, err
	}

	breakdown.ModelRuntimeProfile = estimated.ModelRuntimeProfile
	breakdown.Metadata = mergePricingMetadata(estimated.Metadata, breakdown.Metadata)

	return breakdown, nil
}

// FinalizeBatchFromUsage uses the price revision pinned when the batch was
// reserved and applies batch token dimensions instead of synchronous rates.
func (s *PricingEngine) FinalizeBatchFromUsage(ctx context.Context, model string, usage *openai.Usage, estimated platform.UsageCostBreakdown) (platform.UsageCostBreakdown, error) {
	if usage == nil {
		return estimated, nil
	}

	rate := s.exchangeRate(ctx)

	price, err := s.pinnedOrCurrentModelPrice(ctx, model, estimated)
	if err != nil {
		return platform.UsageCostBreakdown{}, err
	}

	breakdown, err := pricingStrategy(price.Provider).FinalizeBatch(pricingUsageInput{
		model:        model,
		usage:        usage,
		price:        price,
		exchangeRate: rate.USDtoIRR,
		conditions:   estimated.PricingConditions,
		metadata:     typedmetadata.Wrap(estimated.Metadata),
	})
	if err != nil {
		return platform.UsageCostBreakdown{}, err
	}

	breakdown.ModelRuntimeProfile = estimated.ModelRuntimeProfile
	breakdown.Metadata = mergePricingMetadata(estimated.Metadata, breakdown.Metadata)

	return breakdown, nil
}

// FinalizeImageUsage recalculates final spend for image endpoints.
func (s *PricingEngine) FinalizeImageUsage(ctx context.Context, model string, usage *openai.Usage, providerCostUSDTicks *int64, estimated platform.UsageCostBreakdown) (platform.UsageCostBreakdown, error) {
	if usage == nil && providerCostUSDTicks == nil {
		return estimated, nil
	}

	rate := s.exchangeRate(ctx)
	inputTokens, outputTokens, totalTokens := imageUsageTokenBreakdown(usage, estimated)

	price, err := s.pinnedOrCurrentModelPrice(ctx, model, estimated)
	if err != nil {
		return platform.UsageCostBreakdown{}, err
	}

	return pricingStrategy(price.Provider).FinalizeImage(pricingImageInput{
		pricingUsageInput: pricingUsageInput{
			model:        model,
			usage:        usage,
			price:        price,
			exchangeRate: rate.USDtoIRR,
			conditions:   estimated.PricingConditions,
			metadata:     typedmetadata.Wrap(estimated.Metadata),
		},
		inputTokens:          inputTokens,
		outputTokens:         outputTokens,
		totalTokens:          totalTokens,
		providerCostUSDTicks: providerCostUSDTicks,
	})
}

// PriceMeteredUsage prices non-token provider usage (for example speech
// characters or transcription minutes) against the same pinned public catalog
// snapshot used by token billing. Unknown legacy-provider dimensions leave the
// estimate unchanged; catalog-governed Mistral dimensions fail closed.
func (s *PricingEngine) PriceMeteredUsage(
	ctx context.Context,
	model string,
	dimension string,
	quantity string,
	estimated platform.UsageCostBreakdown,
) (platform.UsageCostBreakdown, error) {
	normalizedDimension := strings.TrimSpace(dimension)

	normalizedQuantity := strings.TrimSpace(quantity)
	if normalizedDimension == "" || normalizedQuantity == "" {
		return estimated, nil
	}

	parsedQuantity, err := decimal.NewFromString(normalizedQuantity)
	if err != nil || parsedQuantity.IsNegative() {
		return platform.UsageCostBreakdown{}, fmt.Errorf(
			"%w: invalid metered quantity",
			ErrPricingUnavailable,
		)
	}

	rate := s.exchangeRate(ctx)

	price, err := s.pinnedOrCurrentModelPrice(ctx, model, estimated)
	if err != nil {
		return platform.UsageCostBreakdown{}, err
	}

	pricingRate, ok := price.Rates[normalizedDimension]
	if !ok {
		if pricingStrategy(price.Provider).MeteredUsageFailsClosed() {
			return platform.UsageCostBreakdown{}, fmt.Errorf(
				"%w: missing %s rate for %s",
				ErrPricingUnavailable,
				price.Provider,
				normalizedDimension,
			)
		}

		return estimated, nil
	}

	component := usageCostComponentDecimal(
		normalizedDimension,
		parsedQuantity,
		pricingRate,
	)

	totalExact, err := decimal.NewFromString(component.CostUSD)
	if err != nil {
		return platform.UsageCostBreakdown{}, fmt.Errorf(
			"%w: invalid metered cost",
			ErrPricingUnavailable,
		)
	}

	totalUSD, _ := totalExact.Float64()

	out := estimated
	out.Model = model
	out.InputCostUSD = totalUSD
	out.OutputCostUSD = 0
	out.TotalCostUSD = totalUSD
	out.ExchangeRate = rate.USDtoIRR
	out.TotalCostIRR = totalUSD * rate.USDtoIRR
	out.PricingRevision = price.Revision
	out.ModelRuntimeProfile = price.RuntimeProfile
	out.Components = []platform.UsageCostComponent{component}
	out.Quantities = map[string]string{
		normalizedDimension: parsedQuantity.String(),
	}
	out.PricingSnapshot = usagePricingSnapshot(price)

	return out, nil
}

func costBreakdown(model string, inputTokens int, outputTokens int, price priceInfo, exchangeRate float64) platform.UsageCostBreakdown {
	inputRate := price.InputPerToken
	if inputRate.IsZero() && price.Input != 0 {
		inputRate = decimal.NewFromFloat(price.Input).Div(decimal.NewFromInt(1000))
	}

	outputRate := price.OutputPerToken
	if outputRate.IsZero() && price.Output != 0 {
		outputRate = decimal.NewFromFloat(price.Output).Div(decimal.NewFromInt(1000))
	}

	inputCostExact := inputRate.Mul(decimal.NewFromInt(int64(inputTokens)))
	outputCostExact := outputRate.Mul(decimal.NewFromInt(int64(outputTokens)))
	inputCost, _ := inputCostExact.Float64()
	outputCost, _ := outputCostExact.Float64()
	totalExact := inputCostExact.Add(outputCostExact)
	totalUSD, _ := totalExact.Float64()

	breakdown := platform.UsageCostBreakdown{
		Model:         model,
		InputTokens:   inputTokens,
		OutputTokens:  outputTokens,
		TotalTokens:   inputTokens + outputTokens,
		InputCostUSD:  inputCost,
		OutputCostUSD: outputCost,
		TotalCostUSD:  totalUSD,
		ExchangeRate:  exchangeRate,
		TotalCostIRR:  totalUSD * exchangeRate,
	}
	if price.Provider != "" && len(price.Rates) > 0 {
		breakdown.PricingRevision = price.Revision

		breakdown.ModelRuntimeProfile = price.RuntimeProfile
		for _, dimension := range []string{"input_tokens", "output_tokens"} {
			rate, ok := price.Rates[dimension]
			if !ok {
				continue
			}

			quantity := inputTokens
			if dimension == "output_tokens" {
				quantity = outputTokens
			}

			breakdown.Components = append(
				breakdown.Components,
				usageCostComponent(dimension, int64(quantity), rate),
			)
		}

		breakdown.PricingSnapshot = usagePricingSnapshot(price)
	}

	return breakdown
}

func anthropicTokenCostBreakdown(model string, inputTokens, outputTokens int, price priceInfo, exchangeRate float64, inputDimension, outputDimension string) (platform.UsageCostBreakdown, error) {
	inputRate, inputOK := price.Rates[inputDimension]

	outputRate, outputOK := price.Rates[outputDimension]
	if !inputOK || !outputOK {
		return platform.UsageCostBreakdown{}, fmt.Errorf("%w: conditional Anthropic token rate is missing", ErrPricingUnavailable)
	}

	inputExact := inputRate.UnitPriceUSD.Mul(decimal.NewFromInt(int64(inputTokens)))
	outputExact := outputRate.UnitPriceUSD.Mul(decimal.NewFromInt(int64(outputTokens)))
	totalExact := inputExact.Add(outputExact)
	inputUSD, _ := inputExact.Float64()
	outputUSD, _ := outputExact.Float64()
	totalUSD, _ := totalExact.Float64()

	return platform.UsageCostBreakdown{
		Model: model, InputTokens: inputTokens, OutputTokens: outputTokens,
		TotalTokens: inputTokens + outputTokens, InputCostUSD: inputUSD,
		OutputCostUSD: outputUSD, TotalCostUSD: totalUSD, ExchangeRate: exchangeRate,
		TotalCostIRR: totalUSD * exchangeRate, PricingRevision: price.Revision,
		ModelRuntimeProfile: price.RuntimeProfile,
		Components: []platform.UsageCostComponent{
			usageCostComponent(inputDimension, int64(inputTokens), inputRate),
			usageCostComponent(outputDimension, int64(outputTokens), outputRate),
		},
		PricingSnapshot: usagePricingSnapshot(price),
	}, nil
}

func (s *PricingEngine) modelPrice(
	ctx context.Context,
	model string,
) (priceInfo, error) {
	provider, hasProvider := providerregistry.FromModelIdentifier(model)
	if hasProvider && provider.PricingSnapshot == providerregistry.PricingSnapshotAnthropic {
		if s == nil || s.anthropicPricing == nil {
			return priceInfo{}, ErrPricingUnavailable
		}

		item, snapshot, err := s.anthropicPricing.price(model)
		if err != nil {
			return priceInfo{}, err
		}

		return priceInfo{
			Input:          decimalToPerThousand(item.Rates["input_tokens"].UnitPriceUSD),
			Output:         decimalToPerThousand(item.Rates["output_tokens"].UnitPriceUSD),
			InputPerToken:  item.Rates["input_tokens"].UnitPriceUSD,
			OutputPerToken: item.Rates["output_tokens"].UnitPriceUSD,
			Rates:          item.Rates,
			RateList:       item.RateList,
			Revision:       snapshot.Revision,
			GeneratedAt:    snapshot.GeneratedAt,
			Provider:       "anthropic",
			RuntimeProfile: item.RuntimeProfile,
		}, nil
	}

	if hasProvider && provider.PricingSnapshot == providerregistry.PricingSnapshotCatalog {
		if s == nil || s.catalogPricing == nil || s.catalogPricing[provider.Slug] == nil {
			return priceInfo{}, ErrPricingUnavailable
		}

		item, snapshot, err := s.catalogPricing[provider.Slug].price(model)
		if err != nil {
			return priceInfo{}, err
		}

		if !catalogPricingModelStatusBillable(item.Status) {
			return priceInfo{}, fmt.Errorf(
				"%w: model lifecycle is %s",
				ErrPricingModelUnavailable,
				item.Status,
			)
		}

		inputRate := item.Rates["input_tokens"].UnitPriceUSD
		outputRate := item.Rates["output_tokens"].UnitPriceUSD

		return priceInfo{
			Input:          decimalToPerThousand(inputRate),
			Output:         decimalToPerThousand(outputRate),
			InputPerToken:  inputRate,
			OutputPerToken: outputRate,
			Rates:          item.Rates,
			RateList:       item.RateList,
			Revision:       snapshot.Revision,
			GeneratedAt:    snapshot.GeneratedAt,
			Provider:       provider.Slug,
			RuntimeProfile: item.RuntimeProfile,
		}, nil
	}

	if s != nil && s.pricingCache != nil {
		return s.pricingCache.Get(ctx, model), nil
	}

	return modelPriceFromRegistry(model), nil
}

func catalogPricingModelStatusBillable(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "active", "preview":
		return true
	default:
		return false
	}
}

func (s *PricingEngine) pinnedOrCurrentModelPrice(
	ctx context.Context,
	model string,
	estimated platform.UsageCostBreakdown,
) (priceInfo, error) {
	_, hasProvider := providerregistry.FromModelIdentifier(model)
	if !hasProvider || estimated.PricingSnapshot == nil {
		return s.modelPrice(ctx, model)
	}

	return priceInfoFromUsageSnapshot(estimated.PricingSnapshot)
}

func catalogPricingProvider(model string) string {
	provider, ok := providerregistry.FromModelIdentifier(model)
	if !ok || provider.PricingSnapshot != providerregistry.PricingSnapshotCatalog {
		return ""
	}

	return provider.Slug
}

func decimalToPerThousand(value decimal.Decimal) float64 {
	out, _ := value.Mul(decimal.NewFromInt(1000)).Float64()
	return out
}

func usageCostComponent(
	dimension string,
	quantity int64,
	rate normalizedPricingRate,
) platform.UsageCostComponent {
	return usageCostComponentDecimal(
		dimension,
		decimal.NewFromInt(quantity),
		rate,
	)
}

func usageCostComponentDecimal(
	dimension string,
	quantity decimal.Decimal,
	rate normalizedPricingRate,
) platform.UsageCostComponent {
	cost := rate.UnitPriceUSD.Mul(quantity)

	return platform.UsageCostComponent{
		Dimension:    dimension,
		Quantity:     quantity.String(),
		Unit:         rate.Unit,
		UnitPriceUSD: rate.UnitPriceUSD.String(),
		CostUSD:      cost.String(),
		Condition:    cloneInterfaceMap(rate.Condition),
	}
}

func usagePricingSnapshot(price priceInfo) *platform.UsagePricingSnapshot {
	rates := make(map[string]platform.UsageCostComponent, len(price.Rates))
	for dimension, rate := range price.Rates {
		rates[dimension] = usageCostComponent(dimension, 0, rate)
	}

	rateList := make([]platform.UsageCostComponent, 0, len(price.RateList))
	for _, rate := range price.RateList {
		rateList = append(rateList, usageCostComponent(rate.Dimension, 0, rate))
	}

	return &platform.UsagePricingSnapshot{
		Provider:    price.Provider,
		Revision:    price.Revision,
		GeneratedAt: price.GeneratedAt,
		Rates:       rates,
		RateList:    rateList,
	}
}

func priceInfoFromUsageSnapshot(
	snapshot *platform.UsagePricingSnapshot,
) (priceInfo, error) {
	if snapshot == nil || strings.TrimSpace(snapshot.Provider) == "" ||
		snapshot.Revision == "" || snapshot.GeneratedAt.IsZero() {
		return priceInfo{}, ErrPricingUnavailable
	}

	rates := make(
		map[string]normalizedPricingRate,
		len(snapshot.Rates),
	)
	for dimension, component := range snapshot.Rates {
		value, err := decimal.NewFromString(component.UnitPriceUSD)
		if err != nil || component.Unit == "" {
			return priceInfo{}, ErrPricingUnavailable
		}

		rates[dimension] = normalizedPricingRate{
			Dimension:    dimension,
			Unit:         component.Unit,
			UnitPriceUSD: value,
			Condition:    cloneInterfaceMap(component.Condition),
		}
	}

	rateList := make([]normalizedPricingRate, 0, len(snapshot.RateList))
	for _, component := range snapshot.RateList {
		value, err := decimal.NewFromString(component.UnitPriceUSD)
		if err != nil || component.Dimension == "" || component.Unit == "" {
			return priceInfo{}, ErrPricingUnavailable
		}

		rateList = append(rateList, normalizedPricingRate{
			Dimension:    component.Dimension,
			Unit:         component.Unit,
			UnitPriceUSD: value,
			Condition:    cloneInterfaceMap(component.Condition),
		})
	}

	input := rates["input_tokens"]
	output := rates["output_tokens"]

	return priceInfo{
		Input:          decimalToPerThousand(input.UnitPriceUSD),
		Output:         decimalToPerThousand(output.UnitPriceUSD),
		InputPerToken:  input.UnitPriceUSD,
		OutputPerToken: output.UnitPriceUSD,
		Rates:          rates,
		RateList:       rateList,
		Revision:       snapshot.Revision,
		GeneratedAt:    snapshot.GeneratedAt,
		Provider:       snapshot.Provider,
	}, nil
}

func (s *PricingEngine) exchangeRate(ctx context.Context) *platform.ExchangeRate {
	if s == nil {
		return nil
	}

	if s.exchangeRateCache != nil {
		if rate := s.exchangeRateCache.Get(ctx); rate != nil {
			return rate
		}
	}

	return &s.defaultRate
}
