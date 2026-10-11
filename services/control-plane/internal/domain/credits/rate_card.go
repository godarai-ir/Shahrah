package credits

import (
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

// Typical message size used for the "about N credits per message" hint.
const (
	TypicalMessageInputTokens  = 2000
	TypicalMessageOutputTokens = 500
)

// RateCard is one published credit rate-card version. Provider prices are
// versioned separately by PricingRevision; the rate card versions the per-model
// multipliers and the public changelog.
type RateCard struct {
	Version       int
	EffectiveFrom time.Time
	Multipliers   map[string]decimal.Decimal
}

// IdentityRateCard is the seeded version 1 card with no multipliers.
func IdentityRateCard() RateCard {
	return RateCard{Version: 1}
}

// Label returns the public rate-card version label.
func (r RateCard) Label() string {
	return fmt.Sprintf("rc-%d", r.Version)
}

// Multiplier returns the model multiplier, or 1 when none is configured.
func (r RateCard) Multiplier(model string) decimal.Decimal {
	if multiplier, ok := r.Multipliers[NormalizeModel(model)]; ok && multiplier.IsPositive() {
		return multiplier
	}

	return decimal.NewFromInt(1)
}

// NormalizeModel returns the multiplier lookup key for a model identifier.
func NormalizeModel(model string) string {
	return strings.ToLower(strings.TrimSpace(model))
}

// ModelKey returns the canonical "provider/model" multiplier key. It strips a
// "tenant:account:" deployment prefix and a duplicated provider prefix.
func ModelKey(provider, model string) string {
	value := NormalizeModel(model)
	if strings.Count(value, ":") == 2 {
		value = value[strings.LastIndex(value, ":")+1:]
	}

	provider = NormalizeModel(provider)
	if provider == "" {
		return value
	}

	return provider + "/" + strings.TrimPrefix(value, provider+"/")
}

// RatePer1K converts a per-token USD price into exact millicredits per 1,000 tokens.
func (r RateCard) RatePer1K(model string, unitPriceUSD decimal.Decimal) decimal.Decimal {
	return ExactFromUSD(unitPriceUSD.Mul(decimal.NewFromInt(1000)), r.Multiplier(model))
}

// RatePerUnit converts a per-unit USD price (call, image) into exact millicredits.
func (r RateCard) RatePerUnit(model string, unitPriceUSD decimal.Decimal) decimal.Decimal {
	return ExactFromUSD(unitPriceUSD, r.Multiplier(model))
}

// TypicalMessage returns the rounded charge of a typical chat message.
func (r RateCard) TypicalMessage(model string, inputPerTokenUSD, outputPerTokenUSD decimal.Decimal) Millicredits {
	usd := inputPerTokenUSD.Mul(decimal.NewFromInt(TypicalMessageInputTokens)).
		Add(outputPerTokenUSD.Mul(decimal.NewFromInt(TypicalMessageOutputTokens)))

	return RoundUp(ExactFromUSD(usd, r.Multiplier(model)))
}
