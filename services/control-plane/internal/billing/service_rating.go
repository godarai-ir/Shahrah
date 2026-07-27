package billing

import (
	"context"
	"errors"
	"strconv"
	"strings"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/shopspring/decimal"
)

// RateDeploymentUsage applies the effective pricing carried by the selected
// deployment. Wallet-billed traffic fails closed; external-billed traffic is
// retained as explicitly unrated usage.
func (s *Service) RateDeploymentUsage(ctx context.Context, dep platform.Deployment, breakdown platform.UsageCostBreakdown) (platform.UsageCostBreakdown, error) {
	rate := s.exchangeRate(ctx)
	exchangeRate := decimal.NewFromFloat(rate.USDtoIRR)
	rated, err := RateUsage(RateInput{
		Model:         firstNonEmpty(breakdown.Model, dep.Model),
		Provider:      dep.Provider,
		Region:        dep.RegionName,
		PricingSource: dep.PricingSource,
		Pricing:       dep.EffectivePricing,
		Measures:      breakdown.UsageMeasures,
		ExchangeRate:  exchangeRate,
	})
	if err != nil {
		if strings.EqualFold(strings.TrimSpace(dep.BillingMode), platform.BillingModeExternalBilled) &&
			(errors.Is(err, ErrPricingUnresolved) || errors.Is(err, ErrRateUnresolved)) {
			breakdown.RatingStatus = "unrated"
			breakdown.PricingSource = dep.PricingSource
			breakdown.PricingSnapshot = cloneModelPrice(dep.EffectivePricing)
			breakdown.ExchangeRate = rate.USDtoIRR
			breakdown.ExchangeRateExact = exchangeRate.String()
			return breakdown, nil
		}
		return platform.UsageCostBreakdown{}, err
	}

	return mergeRatedBreakdown(breakdown, rated)
}

func (s *Service) rateSnapshotUsage(model string, measures []platform.UsageMeasure, estimated platform.UsageCostBreakdown) (platform.UsageCostBreakdown, error) {
	if estimated.PricingSnapshot == nil {
		if estimated.RatingStatus == "unrated" {
			estimated.Model = model
			estimated.UsageMeasures = cloneUsageMeasures(measures)
			return estimated, nil
		}
		return platform.UsageCostBreakdown{}, ErrPricingUnresolved
	}

	exchangeRate := decimal.NewFromFloat(estimated.ExchangeRate)
	if raw := strings.TrimSpace(estimated.ExchangeRateExact); raw != "" {
		if parsed, err := decimal.NewFromString(raw); err == nil {
			exchangeRate = parsed
		}
	}

	region := ""
	if len(estimated.AppliedRates) > 0 {
		region = estimated.AppliedRates[0].Region
	}
	rated, err := RateUsage(RateInput{
		Model:         model,
		Region:        region,
		PricingSource: estimated.PricingSource,
		Pricing:       estimated.PricingSnapshot,
		Measures:      measures,
		ExchangeRate:  exchangeRate,
	})
	if err != nil {
		return platform.UsageCostBreakdown{}, err
	}

	base := estimated
	base.Model = model
	base.InputTokens = measureInt(measures, "input", "token")
	base.OutputTokens = measureInt(measures, "output", "token")
	base.TotalTokens = base.InputTokens + base.OutputTokens
	base.UsageMeasures = cloneUsageMeasures(measures)
	return mergeRatedBreakdown(base, rated)
}

func mergeRatedBreakdown(base platform.UsageCostBreakdown, rated platform.RatedUsage) (platform.UsageCostBreakdown, error) {
	totalUSD, err := decimal.NewFromString(rated.TotalCostUSD)
	if err != nil {
		return platform.UsageCostBreakdown{}, err
	}
	exchangeRate, err := decimal.NewFromString(rated.ExchangeRate)
	if err != nil {
		return platform.UsageCostBreakdown{}, err
	}
	totalIRR, err := decimal.NewFromString(rated.TotalCostIRR)
	if err != nil {
		return platform.UsageCostBreakdown{}, err
	}

	base.TotalCostUSD, _ = totalUSD.Float64()
	base.ExchangeRate, _ = exchangeRate.Float64()
	base.TotalCostIRR, _ = totalIRR.Float64()
	base.TotalCostUSDExact = totalUSD.String()
	base.ExchangeRateExact = exchangeRate.String()
	base.TotalCostIRRExact = totalIRR.String()
	base.UsageMeasures = cloneUsageMeasures(rated.Measures)
	base.AppliedRates = append([]platform.AppliedRate(nil), rated.Components...)
	base.RatingStatus = rated.Status
	base.PricingSource = rated.PricingSource
	base.PricingSnapshot = cloneModelPrice(rated.PricingSnapshot)
	base.PricingSnapshotHash = rated.SnapshotHash
	base.InputCostUSD = 0
	base.OutputCostUSD = 0
	for _, component := range rated.Components {
		cost, parseErr := decimal.NewFromString(component.ComponentCostUSD)
		if parseErr != nil {
			return platform.UsageCostBreakdown{}, parseErr
		}
		costFloat, _ := cost.Float64()
		switch strings.ToLower(strings.TrimSpace(component.Usage)) {
		case "input", "cache_read", "cache_write":
			base.InputCostUSD += costFloat
		default:
			base.OutputCostUSD += costFloat
		}
	}
	return base, nil
}

// UsageCostBreakdownFromRated bridges exact rated usage to existing budget and
// wallet projections without repeating decimal conversion in async workers.
func UsageCostBreakdownFromRated(rated platform.RatedUsage, model string) (platform.UsageCostBreakdown, error) {
	base := platform.UsageCostBreakdown{
		Model:         model,
		UsageMeasures: cloneUsageMeasures(rated.Measures),
		InputTokens:   measureInt(rated.Measures, "input", "token"),
		OutputTokens:  measureInt(rated.Measures, "output", "token"),
	}
	base.TotalTokens = base.InputTokens + base.OutputTokens
	return mergeRatedBreakdown(base, rated)
}

func tokenUsageMeasures(inputTokens, outputTokens int) []platform.UsageMeasure {
	measures := make([]platform.UsageMeasure, 0, 2)
	if inputTokens > 0 {
		measures = append(measures, platform.UsageMeasure{
			Usage: "input", Unit: "token", Amount: strconv.Itoa(inputTokens), ServiceTier: "standard",
		})
	}
	if outputTokens > 0 {
		measures = append(measures, platform.UsageMeasure{
			Usage: "output", Unit: "token", Amount: strconv.Itoa(outputTokens), ServiceTier: "standard",
		})
	}
	return measures
}

func measureInt(measures []platform.UsageMeasure, usage, unit string) int {
	total := decimal.Zero
	for _, measure := range measures {
		if strings.EqualFold(measure.Usage, usage) && strings.EqualFold(measure.Unit, unit) {
			if amount, err := decimal.NewFromString(measure.Amount); err == nil {
				total = total.Add(amount)
			}
		}
	}
	return int(total.IntPart())
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func estimatePromptTokens(req interface{}) int {
	switch v := req.(type) {
	case contractsopenai.ChatCompletionRequest:
		textTokens := 0
		imageTokens := 0
		fileTokens := 0
		for _, msg := range v.Messages {
			textTokens += EstimateTokens(contractsopenai.ChatContentText(msg.Content))
			imageTokens += contractsopenai.ChatContentImageCount(msg.Content) * 1024
			fileTokens += contractsopenai.ChatContentFileCount(msg.Content) * 4096
		}
		if textTokens > 0 || imageTokens > 0 || fileTokens > 0 {
			if textTokens < 100 {
				textTokens = 100
			}
			return textTokens + imageTokens + fileTokens
		}
	case *contractsopenai.ChatCompletionRequest:
		if v != nil {
			return estimatePromptTokens(*v)
		}
	}
	return 100
}
