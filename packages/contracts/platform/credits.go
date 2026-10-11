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

// CreditUsageSummary reports credit consumption for a tenant over a report
// window. In shadow mode credits are computed but not charged.
type CreditUsageSummary struct {
	Available         bool              `json:"available"`
	ShadowMode        bool              `json:"shadow_mode"`
	Preset            string            `json:"preset"`
	PeriodStart       time.Time         `json:"period_start"`
	PeriodEnd         time.Time         `json:"period_end"`
	TotalCredits      string            `json:"total_credits"`
	CacheSavedCredits string            `json:"cache_saved_credits"`
	RequestCount      int64             `json:"request_count"`
	ByModel           []CreditUsageItem `json:"by_model"`
	ByTeam            []CreditUsageItem `json:"by_team"`
	ByProduct         []CreditUsageItem `json:"by_product"`
}

// CreditUsageItem is one grouped row of a credit usage summary.
type CreditUsageItem struct {
	Key          string `json:"key"`
	Label        string `json:"label,omitempty"`
	Credits      string `json:"credits"`
	RequestCount int64  `json:"request_count"`
}
