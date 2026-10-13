package metrics

import (
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/credits"
	"github.com/shopspring/decimal"
)

// ApplyShadowCredits fills the analytics credit fields of a metric event from
// its USD cost. Analytics credits use the event's float USD cost; the ledger-grade
// value is computed from decimal cost components on the spend log.
func ApplyShadowCredits(event *platform.ModelMetricEvent, card credits.RateCard) {
	if event == nil {
		return
	}

	key := credits.ModelKey(event.ProviderName, event.ModelName)
	if event.CostUSD > 0 {
		event.CreditsMC = int64(credits.RoundUp(credits.ExactFromUSD(decimal.NewFromFloat(event.CostUSD), card.Multiplier(key))))
	}

	event.CreditsAvoidedMC = int64(credits.FromAvoidedUSD(event.AvoidedCostUSD, key, card))
	event.CreditProduct = credits.Product(event.Metadata, event.RequestType)
	event.RateCardVersion = card.Version
}
