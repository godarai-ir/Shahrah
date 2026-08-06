package billing

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/shopspring/decimal"
)

// ApplyPricingOverride replaces catalog component rates with the effective
// tenant price card after routing. The catalog revision remains pinned and a
// deterministic private suffix identifies the exact override set.
func (s *PricingEngine) ApplyPricingOverride(
	_ context.Context,
	estimated platform.UsageCostBreakdown,
	effective *platform.ModelPrice,
) (platform.UsageCostBreakdown, error) {
	if effective == nil || estimated.PricingSnapshot == nil || len(estimated.Components) == 0 {
		return platform.UsageCostBreakdown{}, ErrPricingUnavailable
	}

	bySignature := make(map[string]platform.ModelPricingRate, len(effective.Rates))
	for _, rate := range effective.Rates {
		signature := pricingComponentSignature(rate.Dimension, rate.Unit, rate.Condition)
		if _, duplicate := bySignature[signature]; duplicate {
			return platform.UsageCostBreakdown{}, fmt.Errorf(
				"%w: effective private pricing contains an ambiguous rate",
				ErrPricingUnavailable,
			)
		}

		bySignature[signature] = rate
	}

	snapshot := &platform.UsagePricingSnapshot{
		Provider: estimated.PricingSnapshot.Provider, Revision: estimated.PricingSnapshot.Revision,
		GeneratedAt: estimated.PricingSnapshot.GeneratedAt,
		Rates:       make(map[string]platform.UsageCostComponent, len(estimated.PricingSnapshot.Rates)),
		RateList:    append([]platform.UsageCostComponent(nil), estimated.PricingSnapshot.RateList...),
	}
	for key, component := range estimated.PricingSnapshot.Rates {
		snapshot.Rates[key] = component
	}

	inputExact, outputExact := decimal.Zero, decimal.Zero

	for index := range estimated.Components {
		component := estimated.Components[index]

		rate, ok := bySignature[pricingComponentSignature(component.Dimension, component.Unit, component.Condition)]
		if !ok {
			return platform.UsageCostBreakdown{}, fmt.Errorf(
				"%w: effective private pricing is missing %s",
				ErrPricingUnavailable,
				component.Dimension,
			)
		}

		price, err := decimal.NewFromString(rate.UnitPriceUSD)

		quantity, quantityErr := decimal.NewFromString(component.Quantity)
		if err != nil || quantityErr != nil || price.IsNegative() || quantity.IsNegative() {
			return platform.UsageCostBreakdown{}, ErrPricingUnavailable
		}

		cost := price.Mul(quantity)
		component.UnitPriceUSD = price.String()
		component.CostUSD = cost.String()

		estimated.Components[index] = component
		if strings.Contains(component.Dimension, "output") {
			outputExact = outputExact.Add(cost)
		} else {
			inputExact = inputExact.Add(cost)
		}

		for rateIndex := range snapshot.RateList {
			item := snapshot.RateList[rateIndex]
			if pricingComponentSignature(item.Dimension, item.Unit, item.Condition) ==
				pricingComponentSignature(component.Dimension, component.Unit, component.Condition) {
				item.UnitPriceUSD = price.String()
				snapshot.RateList[rateIndex] = item
			}
		}

		if item, exists := snapshot.Rates[component.Dimension]; exists &&
			pricingComponentSignature(item.Dimension, item.Unit, item.Condition) ==
				pricingComponentSignature(component.Dimension, component.Unit, component.Condition) {
			item.UnitPriceUSD = price.String()
			snapshot.Rates[component.Dimension] = item
		}
	}

	totalExact := inputExact.Add(outputExact)
	estimated.InputCostUSD, _ = inputExact.Float64()
	estimated.OutputCostUSD, _ = outputExact.Float64()
	estimated.TotalCostUSD, _ = totalExact.Float64()
	estimated.TotalCostIRR = estimated.TotalCostUSD * estimated.ExchangeRate
	fingerprint, _ := json.Marshal(effective.Rates) //nolint:errcheck // The value is composed exclusively of JSON-safe fields.
	revision := fmt.Sprintf("%s:private:%x", snapshot.Revision, sha256.Sum256(fingerprint))
	snapshot.Revision = revision
	estimated.PricingRevision = revision
	estimated.PricingSnapshot = snapshot

	return estimated, nil
}

func pricingComponentSignature(
	dimension string,
	unit string,
	condition map[string]interface{},
) string {
	encoded, _ := json.Marshal(struct { //nolint:errcheck // The value is composed exclusively of JSON-safe fields.
		Dimension string                 `json:"dimension"`
		Unit      string                 `json:"unit"`
		Condition map[string]interface{} `json:"condition,omitempty"`
	}{
		Dimension: strings.TrimSpace(dimension),
		Unit:      strings.TrimSpace(unit),
		Condition: condition,
	})

	return string(encoded)
}
