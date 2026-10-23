package platform

import "time"

// Credit pack order statuses.
const (
	CreditPackOrderPending          = "PENDING"
	CreditPackOrderAwaitingTransfer = "AWAITING_TRANSFER"
	CreditPackOrderCompleted        = "COMPLETED"
	CreditPackOrderFailed           = "FAILED"
	CreditPackOrderCanceled         = "CANCELED"
)

// CreditPackProviderBankTransfer pays an order by bank transfer instead of a gateway.
const CreditPackProviderBankTransfer = "bank_transfer"

// CreditPack is a purchasable add-on pack with its current price. Amounts are IRR.
type CreditPack struct {
	Slug            string `json:"slug"`
	Version         int    `json:"version"`
	Credits         int64  `json:"credits"`
	DiscountPercent string `json:"discount_percent"`
	ValidityMonths  int    `json:"validity_months"`
	PriceIRR        int64  `json:"price_irr"`
	TaxPercent      string `json:"tax_percent"`
	TaxIRR          int64  `json:"tax_irr"`
	PayableIRR      int64  `json:"payable_irr"`
}

// CreateCreditPackOrderRequest starts a pack purchase.
type CreateCreditPackOrderRequest struct {
	PackSlug  string `json:"pack_slug"`
	Provider  string `json:"provider"`
	ReturnURL string `json:"return_url,omitempty"`
}

// CreditPackOrder is a pack purchase and its payment state.
type CreditPackOrder struct {
	ID             string     `json:"-"`
	PublicID       string     `json:"id"`
	TenantID       string     `json:"tenant_id"`
	PackSlug       string     `json:"pack_slug"`
	PackVersion    int        `json:"pack_version"`
	Credits        int64      `json:"credits"`
	ValidityMonths int        `json:"validity_months"`
	PriceIRR       int64      `json:"price_irr"`
	TaxPercent     string     `json:"tax_percent"`
	TaxIRR         int64      `json:"tax_irr"`
	PayableIRR     int64      `json:"payable_irr"`
	Provider       string     `json:"provider"`
	Status         string     `json:"status"`
	Authority      string     `json:"authority,omitempty"`
	PaymentURL     string     `json:"payment_url,omitempty"`
	RefID          string     `json:"ref_id,omitempty"`
	FailureReason  string     `json:"failure_reason,omitempty"`
	CallbackURL    string     `json:"-"`
	ReturnURL      string     `json:"-"`
	CreditsLotID   string     `json:"-"`
	CreatedAt      time.Time  `json:"created_at"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
}

// CreditSettings are a tenant's credit billing preferences.
type CreditSettings struct {
	BillingUnit    string `json:"billing_unit"`
	OverageAllowed bool   `json:"overage_allowed"`
	OverageEnabled bool   `json:"overage_enabled"`
	// OverageCapCredits is a decimal string with two places, or nil for no cap.
	OverageCapCredits *string `json:"overage_cap_credits"`
	// SeatUpgradeRequestsEnabled lets members ask for a better seat.
	SeatUpgradeRequestsEnabled bool `json:"seat_upgrade_requests_enabled"`
	// SeatPoolLimit*Credits cap how much pool credit a Pro or Max member may use
	// per month after their seat credits; nil means no cap.
	SeatPoolLimitProCredits *string `json:"seat_pool_limit_pro_credits"`
	SeatPoolLimitMaxCredits *string `json:"seat_pool_limit_max_credits"`
}

// UpdateCreditSettingsRequest changes a tenant's credit billing preferences.
// Omitted fields keep their value; an empty cap string removes the cap.
type UpdateCreditSettingsRequest struct {
	OverageEnabled    *bool   `json:"overage_enabled,omitempty"`
	OverageCapCredits *string `json:"overage_cap_credits,omitempty"`
	// Seat settings follow the same rule: an empty limit string removes it.
	SeatUpgradeRequestsEnabled *bool   `json:"seat_upgrade_requests_enabled,omitempty"`
	SeatPoolLimitProCredits    *string `json:"seat_pool_limit_pro_credits,omitempty"`
	SeatPoolLimitMaxCredits    *string `json:"seat_pool_limit_max_credits,omitempty"`
}
