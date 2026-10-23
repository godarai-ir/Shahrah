package billing

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/credits"
)

// creditRatePlaces is the maximum number of decimals shown for one rate.
const creditRatePlaces = 6

// CreditRateCard builds the public credit rate card from the loaded provider
// pricing snapshots. It never fetches: providers whose snapshot is missing or
// stale are omitted until their background refresh succeeds. The returned
// revision identifies the rate-card version plus every provider revision and is
// suitable as an HTTP ETag.
func (s *PricingEngine) CreditRateCard(card credits.RateCard) (platform.CreditRateCard, string) {
	out := platform.CreditRateCard{
		Object:        platform.CreditRateCardObject,
		Version:       card.Label(),
		EffectiveFrom: card.EffectiveFrom,
		Unit:          platform.CreditUnit,
		Models:        []platform.CreditRateCardModel{},
		SideUsage: &platform.CreditSideUsagePrices{
			GuardrailJudgeCall:     credits.Format(credits.Millicredits(credits.GuardrailJudgeCallMC)),
			SemanticCacheEmbedding: credits.Format(credits.Millicredits(credits.SemanticCacheEmbeddingMC)),
		},
	}

	// Side-usage prices are part of the card, so a price change changes the ETag.
	sideRevision := fmt.Sprintf("side=%d/%d", credits.GuardrailJudgeCallMC, credits.SemanticCacheEmbeddingMC)

	if s == nil {
		return out, creditRateCardRevision(card, []string{sideRevision})
	}

	snapshots := s.currentPricingSnapshots(time.Now())
	revisions := make([]string, 0, len(snapshots)+1)
	revisions = append(revisions, sideRevision)

	for provider, snapshot := range snapshots {
		revisions = append(revisions, provider+"="+snapshot.Revision)

		for _, model := range snapshot.Models {
			if !catalogPricingModelStatusBillable(model.Status) {
				continue
			}

			out.Models = append(out.Models, creditRateCardModel(card, provider, model))
		}
	}

	sort.Slice(out.Models, func(i, j int) bool { return out.Models[i].Model < out.Models[j].Model })

	return out, creditRateCardRevision(card, revisions)
}

// currentPricingSnapshots returns the loaded, current snapshot of every pricing provider.
func (s *PricingEngine) currentPricingSnapshots(now time.Time) map[string]*anthropicPricingSnapshot {
	out := map[string]*anthropicPricingSnapshot{}

	if s.anthropicPricing != nil {
		if snapshot := s.anthropicPricing.state.Load(); snapshot != nil && anthropicPricingSnapshotCurrent(snapshot, now) {
			out["anthropic"] = snapshot
		}
	}

	for provider, manager := range s.catalogPricing {
		if manager == nil {
			continue
		}

		if snapshot := manager.state.Load(); snapshot != nil && catalogPricingSnapshotCurrent(snapshot, now) {
			out[provider] = snapshot
		}
	}

	return out
}

func creditRateCardModel(card credits.RateCard, provider string, model anthropicPricingModel) platform.CreditRateCardModel {
	id := model.CanonicalID
	if strings.TrimSpace(id) == "" {
		id = model.Slug
	}

	key := credits.ModelKey(provider, id)
	rates := model.RateList

	if len(rates) == 0 {
		rates = sortedPricingRates(model.Rates)
	}

	out := platform.CreditRateCardModel{
		Model:    key,
		Provider: provider,
		Rates:    make([]platform.CreditRate, 0, len(rates)),
	}

	for _, rate := range rates {
		out.Rates = append(out.Rates, creditRate(card, key, rate))
	}

	input, hasInput := model.Rates["input_tokens"]
	output, hasOutput := model.Rates["output_tokens"]

	if hasInput && hasOutput {
		out.TypicalMessageCredits = credits.Format(card.TypicalMessage(key, input.UnitPriceUSD, output.UnitPriceUSD))
	}

	return out
}

func creditRate(card credits.RateCard, key string, rate normalizedPricingRate) platform.CreditRate {
	unit := strings.ToLower(strings.TrimSpace(rate.Unit))
	out := platform.CreditRate{
		Dimension: rate.Dimension,
		Condition: cloneInterfaceMap(rate.Condition),
	}

	switch unit {
	case "token":
		out.Per = platform.CreditRatePer1KToken
		out.Credits = credits.FormatExact(card.RatePer1K(key, rate.UnitPriceUSD), creditRatePlaces)
	case "character":
		out.Per = platform.CreditRatePer1KChars
		out.Credits = credits.FormatExact(card.RatePer1K(key, rate.UnitPriceUSD), creditRatePlaces)
	default:
		out.Per = unit
		out.Credits = credits.FormatExact(card.RatePerUnit(key, rate.UnitPriceUSD), creditRatePlaces)
	}

	if len(out.Condition) == 0 {
		out.Condition = nil
	}

	return out
}

func sortedPricingRates(rates map[string]normalizedPricingRate) []normalizedPricingRate {
	keys := make([]string, 0, len(rates))
	for key := range rates {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	out := make([]normalizedPricingRate, 0, len(keys))
	for _, key := range keys {
		out = append(out, rates[key])
	}

	return out
}

func creditRateCardRevision(card credits.RateCard, providerRevisions []string) string {
	sort.Strings(providerRevisions)

	hash := sha256.Sum256([]byte(card.Label() + "|" + strings.Join(providerRevisions, ",")))

	return hex.EncodeToString(hash[:16])
}
