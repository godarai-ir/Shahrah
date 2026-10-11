package platform

import "time"

// CreditRateCard is the public, versioned credit rate card served by GET /v1/rate-card.
// Amounts are decimal strings in credits; USD prices are intentionally not exposed.
type CreditRateCard struct {
	Object        string                `json:"object"`
	Version       string                `json:"version"`
	EffectiveFrom time.Time             `json:"effective_from"`
	Unit          string                `json:"unit"`
	Models        []CreditRateCardModel `json:"models"`
}

// CreditRateCardModel lists the credit rates for one model.
type CreditRateCardModel struct {
	Model                 string       `json:"model"`
	Provider              string       `json:"provider"`
	Rates                 []CreditRate `json:"rates"`
	TypicalMessageCredits string       `json:"typical_message_credits,omitempty"`
}

// CreditRate is the credit price of one usage dimension.
type CreditRate struct {
	Dimension string                 `json:"dimension"`
	Per       string                 `json:"per"`
	Credits   string                 `json:"credits"`
	Condition map[string]interface{} `json:"condition,omitempty"`
}

// Credit rate-card constants.
const (
	CreditRateCardObject = "rate_card"
	CreditUnit           = "credit"
	CreditRatePer1KToken = "1k_tokens"
	CreditRatePer1KChars = "1k_characters"
)
