package platform

import "time"

// CreditPriceRevision is one credit list price, proposed or decided. Prices
// are toman per credit; dollar rates are toman per USD.
type CreditPriceRevision struct {
	ID                  string     `json:"id"`
	PriceTomanPerCredit int64      `json:"price_toman_per_credit"`
	PreviousPriceToman  *int64     `json:"previous_price_toman,omitempty"`
	BaseUSDToman        string     `json:"base_usd_toman"`
	ObservedAvgUSDToman *string    `json:"observed_avg_usd_toman,omitempty"`
	BandPercent         string     `json:"band_pct"`
	Status              string     `json:"status"`
	EffectiveFrom       *time.Time `json:"effective_from,omitempty"`
	PlanEffectiveFrom   *time.Time `json:"plan_effective_from,omitempty"`
	ProposedBy          string     `json:"proposed_by,omitempty"`
	DecidedBy           string     `json:"decided_by,omitempty"`
	DecidedAt           *time.Time `json:"decided_at,omitempty"`
	DecisionReason      string     `json:"decision_reason,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
}

// CreditPriceStatus is the support-admin view of the credit price policy.
type CreditPriceStatus struct {
	Active                  *CreditPriceRevision  `json:"active"`
	Proposed                *CreditPriceRevision  `json:"proposed"`
	SevenDayAverageUSDToman *string               `json:"seven_day_average_usd_toman"`
	Revisions               []CreditPriceRevision `json:"revisions"`
}

// CreditPriceDecisionRequest approves or rejects a proposed price.
type CreditPriceDecisionRequest struct {
	Reason string `json:"reason"`
}

// SetCreditPriceRequest sets the active credit list price by hand.
type SetCreditPriceRequest struct {
	PriceTomanPerCredit int64  `json:"price_toman_per_credit"`
	Reason              string `json:"reason"`
}

// PricingChangelog is the public history of rate-card and credit price changes.
type PricingChangelog struct {
	RateCards    []PricingChangelogRateCard `json:"rate_cards"`
	CreditPrices []PricingChangelogPrice    `json:"credit_prices"`
}

// PricingChangelogRateCard is one published or scheduled rate-card version.
type PricingChangelogRateCard struct {
	Version       int       `json:"version"`
	Label         string    `json:"label"`
	Status        string    `json:"status"`
	EffectiveFrom time.Time `json:"effective_from"`
	ChangelogFA   string    `json:"changelog_fa"`
	ChangelogEN   string    `json:"changelog_en"`
}

// PricingChangelogPrice is one decided credit list price.
type PricingChangelogPrice struct {
	PriceTomanPerCredit int64     `json:"price_toman_per_credit"`
	Status              string    `json:"status"`
	EffectiveFrom       time.Time `json:"effective_from"`
	PlanEffectiveFrom   time.Time `json:"plan_effective_from"`
}

// CreditRateCardVersion is a rate-card version with its model multipliers.
type CreditRateCardVersion struct {
	Version       int                     `json:"version"`
	Status        string                  `json:"status"`
	EffectiveFrom time.Time               `json:"effective_from"`
	NoticeUntil   *time.Time              `json:"notice_until,omitempty"`
	ChangelogFA   string                  `json:"changelog_fa"`
	ChangelogEN   string                  `json:"changelog_en"`
	PublishedBy   string                  `json:"published_by,omitempty"`
	PublishedAt   *time.Time              `json:"published_at,omitempty"`
	Multipliers   []CreditModelMultiplier `json:"multipliers"`
}

// CreditModelMultiplier adjusts one model's credit rate, always with a reason.
type CreditModelMultiplier struct {
	Model      string    `json:"model"`
	Multiplier string    `json:"multiplier"`
	Reason     string    `json:"reason"`
	CreatedBy  string    `json:"created_by,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// UpsertCreditMultiplierRequest sets a model multiplier on a draft rate card.
type UpsertCreditMultiplierRequest struct {
	Model      string `json:"model"`
	Multiplier string `json:"multiplier"`
	Reason     string `json:"reason"`
}

// PublishCreditRateCardRequest publishes a draft rate card with its changelog.
type PublishCreditRateCardRequest struct {
	ChangelogFA string `json:"changelog_fa"`
	ChangelogEN string `json:"changelog_en"`
}

// CreditPackVersion is one version of an add-on pack in the catalog.
type CreditPackVersion struct {
	Slug            string    `json:"slug"`
	Version         int       `json:"version"`
	Credits         int64     `json:"credits"`
	DiscountPercent string    `json:"discount_percent"`
	ValidityMonths  int       `json:"validity_months"`
	Status          string    `json:"status"`
	SortOrder       int       `json:"sort_order"`
	CreatedBy       string    `json:"created_by,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

// CreateCreditPackVersionRequest drafts a new version of a pack.
type CreateCreditPackVersionRequest struct {
	Slug            string `json:"slug"`
	Credits         int64  `json:"credits"`
	DiscountPercent string `json:"discount_percent"`
	ValidityMonths  int    `json:"validity_months"`
	SortOrder       int    `json:"sort_order"`
}

// TenantCreditOverview is the support-admin summary of a tenant's credits.
// Amounts are credits as decimal strings with two places.
type TenantCreditOverview struct {
	TenantID         string         `json:"tenant_id"`
	BillingUnit      string         `json:"billing_unit"`
	AvailableCredits string         `json:"available_credits"`
	DebtCredits      string         `json:"debt_credits"`
	LedgerSumCredits string         `json:"ledger_sum_credits"`
	Settings         CreditSettings `json:"settings"`
}

// CreditLedgerEntry is one row of a tenant's credit ledger.
type CreditLedgerEntry struct {
	EntryID          string    `json:"entry_id"`
	Kind             string    `json:"kind"`
	AmountCredits    string    `json:"amount_credits"`
	OccurredAt       time.Time `json:"occurred_at"`
	LotID            string    `json:"lot_id,omitempty"`
	RequestID        string    `json:"request_id,omitempty"`
	Model            string    `json:"model,omitempty"`
	UserID           string    `json:"user_id,omitempty"`
	TeamID           string    `json:"team_id,omitempty"`
	Product          string    `json:"product,omitempty"`
	BillingMode      string    `json:"billing_mode,omitempty"`
	CreditEquivalent *string   `json:"credit_equivalent,omitempty"`
	Reason           string    `json:"reason,omitempty"`
	Actor            string    `json:"actor,omitempty"`
}

// CreditLot is a tenant's credit lot with its remaining balance.
type CreditLot struct {
	LotID             string                 `json:"lot_id"`
	Source            string                 `json:"source"`
	AmountCredits     string                 `json:"amount_credits"`
	RemainingCredits  string                 `json:"remaining_credits"`
	ExpiresAt         *time.Time             `json:"expires_at,omitempty"`
	PurchasePriceIRR  *int64                 `json:"purchase_price_irr,omitempty"`
	PricePerCreditIRR *string                `json:"price_per_credit_irr,omitempty"`
	SourceRef         map[string]interface{} `json:"source_ref,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
}

// CreditLotAllocation is one movement into or out of a lot.
type CreditLotAllocation struct {
	EntryID       string    `json:"entry_id"`
	Kind          string    `json:"kind"`
	AmountCredits string    `json:"amount_credits"`
	OccurredAt    time.Time `json:"occurred_at"`
}

// CreditAdjustmentRequest corrects a tenant balance. AmountCredits is a signed
// decimal with at most two places; positive amounts create a lot.
type CreditAdjustmentRequest struct {
	AmountCredits  string     `json:"amount_credits"`
	Reason         string     `json:"reason"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	IdempotencyKey string     `json:"idempotency_key"`
}

// SetBillingUnitRequest moves a tenant between IRR and credit billing.
type SetBillingUnitRequest struct {
	BillingUnit string `json:"billing_unit"`
	Reason      string `json:"reason"`
}

// Credit conversion warnings reported by a dry run.
const (
	CreditConversionWarningNoSubscription   = "no_subscription"
	CreditConversionWarningNoCreditPlan     = "no_published_credit_plan"
	CreditConversionWarningPendingTopUp     = "pending_wallet_topup"
	CreditConversionWarningPendingTransfer  = "pending_bank_transfer"
	CreditConversionWarningUnsettledUsage   = "unsettled_irr_usage"
	CreditConversionWarningAlreadyConverted = "already_credit"
)

// CreditConversionPreview is the dry run of converting one tenant to credits.
type CreditConversionPreview struct {
	TenantID            string `json:"tenant_id"`
	TenantSlug          string `json:"tenant_slug"`
	BillingUnit         string `json:"billing_unit"`
	PriceTomanPerCredit int64  `json:"price_toman_per_credit"`
	WalletIRR           int64  `json:"wallet_irr"`
	// UnsettledIRR is IRR spend committed this month and not yet debited; the
	// conversion settles it first.
	UnsettledIRR    int64    `json:"unsettled_irr"`
	Credits         string   `json:"credits"`
	ResidueIRR      int64    `json:"residue_irr"`
	RulesToConvert  int      `json:"rules_to_convert"`
	PlanSlug        string   `json:"plan_slug,omitempty"`
	BillingInterval string   `json:"billing_interval,omitempty"`
	NewPlanPriceIRR *int64   `json:"new_plan_price_irr,omitempty"`
	NewPlanCredits  string   `json:"new_plan_monthly_credits,omitempty"`
	Warnings        []string `json:"warnings"`
}

// CreditConversion is one executed conversion of a tenant.
type CreditConversion struct {
	ID                  string    `json:"id"`
	TenantID            string    `json:"tenant_id"`
	Sequence            int       `json:"sequence"`
	Direction           string    `json:"direction"`
	PriceTomanPerCredit int64     `json:"price_toman_per_credit"`
	WalletIRR           int64     `json:"wallet_irr"`
	ConvertedIRR        int64     `json:"converted_irr"`
	ResidueIRR          int64     `json:"residue_irr"`
	Credits             string    `json:"credits"`
	Debt                string    `json:"debt"`
	UncollectedIRR      int64     `json:"uncollected_irr"`
	RulesConverted      int       `json:"rules_converted"`
	PlanSlug            string    `json:"plan_slug,omitempty"`
	SubscriptionID      string    `json:"subscription_id,omitempty"`
	Actor               string    `json:"actor"`
	Reason              string    `json:"reason"`
	CreatedAt           time.Time `json:"created_at"`
}

// CreditConversionRequest converts or rolls back one tenant.
type CreditConversionRequest struct {
	Reason string `json:"reason"`
}

// CreditFleetConversionRequest converts the next IRR tenants in one batch.
type CreditFleetConversionRequest struct {
	Reason string `json:"reason"`
	// Limit is the batch size, 1 to 100 (default 20).
	Limit int `json:"limit,omitempty"`
}

// CreditConversionFailure is a tenant a fleet batch could not convert.
type CreditConversionFailure struct {
	TenantID   string `json:"tenant_id"`
	TenantSlug string `json:"tenant_slug"`
	Error      string `json:"error"`
}

// CreditFleetConversionResult reports one fleet batch.
type CreditFleetConversionResult struct {
	Converted []CreditConversion        `json:"converted"`
	Failed    []CreditConversionFailure `json:"failed"`
	// Remaining is the number of IRR tenants left after the batch.
	Remaining int `json:"remaining"`
}

// TenantCreditConversionStatus is a tenant's conversion dry run and history.
type TenantCreditConversionStatus struct {
	Preview     CreditConversionPreview `json:"preview"`
	Conversions []CreditConversion      `json:"conversions"`
}
