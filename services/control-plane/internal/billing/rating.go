package billing

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/shopspring/decimal"
)

var (
	ErrPricingUnresolved   = errors.New("pricing unresolved")
	ErrRateUnresolved      = errors.New("pricing rate unresolved")
	ErrUnsupportedCurrency = errors.New("unsupported pricing currency")
)

// RateInput contains every value needed to deterministically rate usage.
type RateInput struct {
	Model         string
	Provider      string
	Region        string
	PricingSource string
	Pricing       *platform.ModelPrice
	Measures      []platform.UsageMeasure
	ExchangeRate  decimal.Decimal
	At            time.Time
}

// RateUsage selects exact catalog rates and calculates an auditable result.
func RateUsage(input RateInput) (platform.RatedUsage, error) {
	result := platform.RatedUsage{
		Model:           strings.TrimSpace(input.Model),
		Provider:        strings.TrimSpace(input.Provider),
		Region:          strings.TrimSpace(input.Region),
		PricingSource:   strings.TrimSpace(input.PricingSource),
		Measures:        cloneUsageMeasures(input.Measures),
		TotalCostUSD:    "0",
		ExchangeRate:    input.ExchangeRate.String(),
		TotalCostIRR:    "0",
		PricingSnapshot: cloneModelPrice(input.Pricing),
	}
	if input.At.IsZero() {
		input.At = time.Now().UTC()
	}

	if input.Pricing == nil || strings.EqualFold(strings.TrimSpace(input.Pricing.Status), "unknown") {
		result.Status = "unrated"
		return result, ErrPricingUnresolved
	}
	if strings.EqualFold(strings.TrimSpace(input.Pricing.Status), "free") {
		result.Status = "free"
		result.SnapshotHash = pricingSnapshotHash(input.Pricing)
		return result, nil
	}

	profile, err := selectPricingProfile(input.Pricing.Profiles, input.Region, input.At)
	if err != nil {
		result.Status = "unrated"
		return result, err
	}
	if !strings.EqualFold(strings.TrimSpace(profile.Currency), "USD") {
		result.Status = "unrated"
		return result, fmt.Errorf("%w: %s", ErrUnsupportedCurrency, profile.Currency)
	}

	totalUSD := decimal.Zero
	for _, measure := range input.Measures {
		usageAmount, parseErr := decimal.NewFromString(strings.TrimSpace(measure.Amount))
		if parseErr != nil || usageAmount.IsNegative() {
			return result, fmt.Errorf("invalid usage amount %q", measure.Amount)
		}
		if usageAmount.IsZero() {
			continue
		}

		rate, selectErr := selectPricingRate(profile.Rates, measure, usageAmount)
		if selectErr != nil {
			result.Status = "unrated"
			return result, selectErr
		}

		rateAmount, parseErr := decimal.NewFromString(strings.TrimSpace(rate.Amount))
		if parseErr != nil || rateAmount.IsNegative() || rate.Quantity <= 0 {
			return result, fmt.Errorf("invalid pricing rate %q", rate.Key)
		}

		componentCost := usageAmount.Div(decimal.NewFromInt(int64(rate.Quantity))).Mul(rateAmount)
		totalUSD = totalUSD.Add(componentCost)
		result.Components = append(result.Components, platform.AppliedRate{
			ProfileKey:       pricingProfileKey(profile),
			RateKey:          rate.Key,
			Usage:            rate.Usage,
			Unit:             rate.Unit,
			RateAmount:       rateAmount.String(),
			RateQuantity:     rate.Quantity,
			UsageAmount:      usageAmount.String(),
			ComponentCostUSD: componentCost.String(),
			Region:           profile.Region,
			Currency:         profile.Currency,
			Resolution:       rate.Resolution,
			ServiceTier:      rate.ServiceTier,
			PricingSource:    input.PricingSource,
			Qualifiers:       cloneInterfaceMap(rate.Qualifiers),
		})
	}

	result.Status = "rated"
	result.Region = profile.Region
	result.TotalCostUSD = totalUSD.String()
	result.TotalCostIRR = totalUSD.Mul(input.ExchangeRate).String()
	result.SnapshotHash = pricingSnapshotHash(input.Pricing)

	return result, nil
}

func selectPricingProfile(profiles []platform.PricingProfile, region string, at time.Time) (platform.PricingProfile, error) {
	active := make([]platform.PricingProfile, 0, len(profiles))
	for _, profile := range profiles {
		if profileActiveAt(profile, at) {
			active = append(active, profile)
		}
	}
	if len(active) == 0 {
		return platform.PricingProfile{}, ErrPricingUnresolved
	}

	region = normalizeDimension(region)
	exact := make([]platform.PricingProfile, 0, len(active))
	for _, profile := range active {
		if region != "" && normalizeDimension(profile.Region) == region {
			exact = append(exact, profile)
		}
	}
	if len(exact) == 1 {
		return exact[0], nil
	}
	if len(exact) > 1 {
		return platform.PricingProfile{}, fmt.Errorf("%w: ambiguous profiles for region %q", ErrPricingUnresolved, region)
	}
	global := make([]platform.PricingProfile, 0, len(active))
	for _, profile := range active {
		if isGlobalRegion(profile.Region) {
			global = append(global, profile)
		}
	}
	if len(global) == 1 {
		return global[0], nil
	}
	if len(global) > 1 {
		return platform.PricingProfile{}, fmt.Errorf("%w: ambiguous global profiles", ErrPricingUnresolved)
	}
	if len(active) == 1 {
		return active[0], nil
	}

	return platform.PricingProfile{}, fmt.Errorf("%w: region %q", ErrPricingUnresolved, region)
}

func pricingProfileKey(profile platform.PricingProfile) string {
	return strings.Join([]string{
		firstNonEmptyString(normalizeDimension(profile.Region), "global"),
		strings.ToUpper(strings.TrimSpace(profile.Currency)),
		strings.TrimSpace(profile.EffectiveFrom),
		strings.TrimSpace(profile.EffectiveTo),
	}, "|")
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func selectPricingRate(rates []platform.PricingRate, measure platform.UsageMeasure, amount decimal.Decimal) (platform.PricingRate, error) {
	type candidate struct {
		rate  platform.PricingRate
		score int
	}
	candidates := make([]candidate, 0)
	for _, rate := range rates {
		if !strings.EqualFold(strings.TrimSpace(rate.Usage), strings.TrimSpace(measure.Usage)) ||
			!strings.EqualFold(strings.TrimSpace(rate.Unit), strings.TrimSpace(measure.Unit)) {
			continue
		}
		if !tierContains(rate.Tier, amount) || !qualifiersMatch(rate.Qualifiers, measure.Qualifiers) {
			continue
		}

		score := 0
		rateResolution := normalizeDimension(rate.Resolution)
		measureResolution := normalizeDimension(measure.Resolution)
		if rateResolution != "" {
			if rateResolution != measureResolution {
				continue
			}
			score += 8
		}

		rateTier := normalizeDimension(rate.ServiceTier)
		measureTier := normalizeDimension(measure.ServiceTier)
		if measureTier == "" {
			measureTier = "standard"
		}
		if rateTier != "" {
			if rateTier != measureTier {
				continue
			}
			score += 4
		}
		if rate.Tier != nil {
			score += 2
		}
		score += len(rate.Qualifiers)
		candidates = append(candidates, candidate{rate: rate, score: score})
	}

	if len(candidates) == 0 {
		return platform.PricingRate{}, fmt.Errorf(
			"%w: usage=%s unit=%s resolution=%s service_tier=%s",
			ErrRateUnresolved, measure.Usage, measure.Unit, measure.Resolution, measure.ServiceTier,
		)
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].score == candidates[j].score {
			return candidates[i].rate.Key < candidates[j].rate.Key
		}
		return candidates[i].score > candidates[j].score
	})
	if len(candidates) > 1 && candidates[0].score == candidates[1].score &&
		candidates[0].rate.Key != candidates[1].rate.Key {
		return platform.PricingRate{}, fmt.Errorf("%w: ambiguous rates %q and %q", ErrRateUnresolved, candidates[0].rate.Key, candidates[1].rate.Key)
	}

	return candidates[0].rate, nil
}

func profileActiveAt(profile platform.PricingProfile, at time.Time) bool {
	if raw := strings.TrimSpace(profile.EffectiveFrom); raw != "" {
		from, err := time.Parse(time.RFC3339, raw)
		if err != nil || at.Before(from) {
			return false
		}
	}
	if raw := strings.TrimSpace(profile.EffectiveTo); raw != "" {
		to, err := time.Parse(time.RFC3339, raw)
		if err != nil || !at.Before(to) {
			return false
		}
	}
	return true
}

func tierContains(tier *platform.PricingTier, amount decimal.Decimal) bool {
	if tier == nil {
		return true
	}
	if tier.From != nil && amount.LessThan(decimal.NewFromInt(int64(*tier.From))) {
		return false
	}
	if tier.To != nil && !amount.LessThan(decimal.NewFromInt(int64(*tier.To))) {
		return false
	}
	return true
}

func qualifiersMatch(rate, measure map[string]interface{}) bool {
	for key, expected := range rate {
		actual, ok := measure[key]
		if !ok || !reflect.DeepEqual(normalizeQualifier(expected), normalizeQualifier(actual)) {
			return false
		}
	}
	return true
}

func normalizeQualifier(value interface{}) interface{} {
	switch typed := value.(type) {
	case json.Number:
		return typed.String()
	case float64:
		return decimal.NewFromFloat(typed).String()
	case float32:
		return decimal.NewFromFloat32(typed).String()
	default:
		return typed
	}
}

func normalizeDimension(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func isGlobalRegion(region string) bool {
	switch normalizeDimension(region) {
	case "", "*", "global":
		return true
	default:
		return false
	}
}

func pricingSnapshotHash(price *platform.ModelPrice) string {
	if price == nil {
		return ""
	}
	raw, err := json.Marshal(price)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func cloneModelPrice(in *platform.ModelPrice) *platform.ModelPrice {
	if in == nil {
		return nil
	}
	raw, err := json.Marshal(in)
	if err != nil {
		return nil
	}
	var out platform.ModelPrice
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil
	}
	return &out
}

func cloneUsageMeasures(in []platform.UsageMeasure) []platform.UsageMeasure {
	out := make([]platform.UsageMeasure, len(in))
	for i := range in {
		out[i] = in[i]
		out[i].Qualifiers = cloneInterfaceMap(in[i].Qualifiers)
	}
	return out
}
