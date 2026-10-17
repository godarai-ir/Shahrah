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
	// SideUsage prices the calls GodarAI makes on a request's behalf.
	SideUsage *CreditSideUsagePrices `json:"side_usage,omitempty"`
}

// CreditSideUsagePrices lists fixed per-call credit prices of side usage.
type CreditSideUsagePrices struct {
	GuardrailJudgeCall     string `json:"guardrail_judge_call"`
	SemanticCacheEmbedding string `json:"semantic_cache_embedding"`
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

// CreditUsageSummary reports a tenant's credit consumption over a report
// window, from the model events in ClickHouse.
type CreditUsageSummary struct {
	Available         bool              `json:"available"`
	Preset            string            `json:"preset"`
	PeriodStart       time.Time         `json:"period_start"`
	PeriodEnd         time.Time         `json:"period_end"`
	TotalCredits      string            `json:"total_credits"`
	CacheSavedCredits string            `json:"cache_saved_credits"`
	RequestCount      int64             `json:"request_count"`
	ByModel           []CreditUsageItem `json:"by_model"`
	ByTeam            []CreditUsageItem `json:"by_team"`
	ByUser            []CreditUsageItem `json:"by_user"`
	ByProduct         []CreditUsageItem `json:"by_product"`
}

// CreditUsageItem is one grouped row of a credit usage summary.
type CreditUsageItem struct {
	Key          string `json:"key"`
	Label        string `json:"label,omitempty"`
	Credits      string `json:"credits"`
	RequestCount int64  `json:"request_count"`
}

// Credit forecast statuses.
const (
	// CreditForecastOK has enough recent usage for a depletion estimate.
	CreditForecastOK = "ok"
	// CreditForecastUncertain estimates from less than two days of usage.
	CreditForecastUncertain = "uncertain"
	// CreditForecastNoUsage has no usage in the window, so no estimate.
	CreditForecastNoUsage = "no_usage"
	// CreditForecastLimitedByUsage means spending runs the credits out first.
	CreditForecastLimitedByUsage = "usage"
	// CreditForecastLimitedByExpiry means the credits expire before they are spent.
	CreditForecastLimitedByExpiry = "expiry"
)

// TenantCreditBalance is a tenant's spendable credits, by lot, with a
// depletion forecast. Amounts are credit strings with two decimals.
type TenantCreditBalance struct {
	AvailableCredits string         `json:"available_credits"`
	DebtCredits      string         `json:"debt_credits"`
	Lots             []CreditLot    `json:"lots"`
	Forecast         CreditForecast `json:"forecast"`
	Settings         CreditSettings `json:"settings"`
}

// CreditForecast estimates when a tenant's credits run out from its recent
// consumption.
type CreditForecast struct {
	Status              string     `json:"status"`
	WindowDays          int        `json:"window_days"`
	ConsumedCredits     string     `json:"consumed_credits"`
	DailyAverageCredits string     `json:"daily_average_credits"`
	DepletesAt          *time.Time `json:"depletes_at,omitempty"`
	// LimitedBy is "usage" when spending runs the credits out first and
	// "expiry" when they expire before they are spent.
	LimitedBy string `json:"limited_by,omitempty"`
}

// BudgetStatusObject is the object name of a GET /v1/budget/me response.
const BudgetStatusObject = "budget_status"

// BudgetStatus reports the caller's credit budgets. Amounts are credit strings
// with two decimals.
type BudgetStatus struct {
	Object string             `json:"object"`
	Unit   string             `json:"unit"`
	Rules  []BudgetRuleStatus `json:"rules"`
	// TenantBalance is shown only to callers allowed to read billing.
	TenantBalance *BudgetBalanceStatus `json:"tenant_balance,omitempty"`
}

// BudgetRuleStatus is one budget rule that applies to the caller.
type BudgetRuleStatus struct {
	RuleID    string    `json:"rule_id"`
	Name      string    `json:"name,omitempty"`
	Scope     string    `json:"scope"`
	Limit     string    `json:"limit"`
	Spent     string    `json:"spent"`
	Remaining string    `json:"remaining"`
	WindowEnd time.Time `json:"window_end"`
	OnExceed  string    `json:"on_exceed"`
}

// BudgetBalanceStatus is what the tenant can still spend this month.
type BudgetBalanceStatus struct {
	Remaining string    `json:"remaining"`
	Reserved  string    `json:"reserved"`
	WindowEnd time.Time `json:"window_end"`
}
