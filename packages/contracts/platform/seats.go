package platform

import "time"

// Seat types. A member without a seat is SeatUnassigned.
const (
	SeatFree       = "free"
	SeatPro        = "pro"
	SeatMax        = "max"
	SeatUnassigned = "unassigned"
)

// Seat upgrade request statuses.
const (
	SeatRequestPending  = "pending"
	SeatRequestApproved = "approved"
	SeatRequestDenied   = "denied"
	SeatRequestCanceled = "canceled"
)

// Seat order kinds.
const (
	SeatOrderPurchase = "purchase"
	SeatOrderRenewal  = "renewal"
)

// SeatProfile is the seat state compiled into a tenant's runtime profile, so
// chat admission needs no database read.
type SeatProfile struct {
	// SubscriptionStatus decides whether seat limits apply.
	SubscriptionStatus string `json:"subscription_status,omitempty"`
	// Members maps a user id to their seat type; absent means no seat.
	Members map[string]string `json:"members,omitempty"`
	// PoolLimitMC is the default personal pool limit per seat type, in
	// millicredits; absent means no limit.
	PoolLimitMC map[string]int64 `json:"pool_limit_mc,omitempty"`
}

// SeatTypeOffer is one seat type with its current price. Amounts are IRR.
type SeatTypeOffer struct {
	Type            string `json:"type"`
	MonthlyCredits  int64  `json:"monthly_credits"`
	Lifetime        bool   `json:"lifetime"`
	PoolAccess      bool   `json:"pool_access"`
	DiscountPercent string `json:"discount_percent"`
	// YearlyDiscountPercent is the extra discount of yearly billing.
	YearlyDiscountPercent string `json:"yearly_discount_percent"`
	MonthlyPriceIRR       int64  `json:"monthly_price_irr"`
	YearlyPriceIRR        int64  `json:"yearly_price_irr"`
	Active                bool   `json:"active"`
}

// SeatTypeCount is how many seats of one paid type a tenant holds and uses.
type SeatTypeCount struct {
	Type      string `json:"type"`
	Purchased int    `json:"purchased"`
	Assigned  int    `json:"assigned"`
	Available int    `json:"available"`
	// Renewing is how many seats the next renewal keeps.
	Renewing int `json:"renewing"`
}

// SeatSummary is a tenant's seat overview.
type SeatSummary struct {
	Enforced           bool   `json:"enforced"`
	SubscriptionStatus string `json:"subscription_status,omitempty"`
	// SubscriptionKind is "free" until the first paid seat purchase starts the
	// "seats" subscription. Billing interval and period are set only for "seats".
	SubscriptionKind   string          `json:"subscription_kind"`
	BillingInterval    string          `json:"billing_interval,omitempty"`
	PeriodStart        *time.Time      `json:"period_start,omitempty"`
	PeriodEnd          *time.Time      `json:"period_end,omitempty"`
	FreeLimit          *int            `json:"free_limit"`
	FreeUsed           int             `json:"free_used"`
	Unassigned         int             `json:"unassigned"`
	PendingInvitations int             `json:"pending_invitations"`
	Paid               []SeatTypeCount `json:"paid"`
	Offers             []SeatTypeOffer `json:"offers"`
	PendingRequests    int             `json:"pending_requests"`
	// RenewalDue is set while the seat subscription's renewal window is open
	// and some paid seats are not yet paid for the next period.
	RenewalDue *SeatRenewalQuote `json:"renewal_due,omitempty"`
	// RenewalOpensAt is when the running subscription's renewal window opens.
	RenewalOpensAt *time.Time `json:"renewal_opens_at,omitempty"`
	TaxPercent     string     `json:"tax_percent"`
}

// SeatMember is one tenant member with their seat and usage.
type SeatMember struct {
	UserID             string     `json:"user_id"`
	Email              string     `json:"email,omitempty"`
	Name               string     `json:"name,omitempty"`
	Role               string     `json:"role,omitempty"`
	Active             bool       `json:"active"`
	Seat               string     `json:"seat"`
	PendingSeat        string     `json:"pending_seat,omitempty"`
	PendingEffectiveAt *time.Time `json:"pending_effective_at,omitempty"`
	FreeSeatUsed       bool       `json:"free_seat_used"`
	// Personal credits are decimal strings with two places.
	PersonalGrantedCredits   string     `json:"personal_granted_credits"`
	PersonalRemainingCredits string     `json:"personal_remaining_credits"`
	PoolUsedCredits          string     `json:"pool_used_credits"`
	PoolLimitCredits         *string    `json:"pool_limit_credits"`
	AssignedAt               *time.Time `json:"assigned_at,omitempty"`
}

// SeatMemberList is a page of seat members.
type SeatMemberList struct {
	Items []SeatMember `json:"items"`
	Total int          `json:"total"`
}

// SetMemberSeatRequest changes a member's seat. Seat "unassigned" removes it.
type SetMemberSeatRequest struct {
	Seat string `json:"seat"`
	// RequestID approves a pending upgrade request with this change.
	RequestID string `json:"request_id,omitempty"`
}

// SetMemberSeatResult reports how a seat change was applied.
type SetMemberSeatResult struct {
	Member SeatMember `json:"member"`
	// Timing is "immediate" or "period_end".
	Timing string `json:"timing"`
}

// SeatQuoteRequest prices buying paid seats for the rest of the period.
type SeatQuoteRequest struct {
	Seat     string `json:"seat"`
	Quantity int    `json:"quantity"`
	// BillingInterval picks monthly or yearly for the purchase that starts the
	// seat subscription. Later purchases follow the subscription's interval.
	BillingInterval string `json:"billing_interval,omitempty"`
}

// SeatQuote is the price of a seat purchase. Amounts are IRR.
type SeatQuote struct {
	Seat                  string    `json:"seat"`
	Quantity              int       `json:"quantity"`
	BillingInterval       string    `json:"billing_interval"`
	PeriodStart           time.Time `json:"period_start"`
	PeriodEnd             time.Time `json:"period_end"`
	RemainingDays         int       `json:"remaining_days"`
	PeriodDays            int       `json:"period_days"`
	ListPriceIRRPerCredit int64     `json:"list_price_irr_per_credit"`
	UnitPeriodPriceIRR    int64     `json:"unit_period_price_irr"`
	PriceIRR              int64     `json:"price_irr"`
	TaxPercent            string    `json:"tax_percent"`
	TaxIRR                int64     `json:"tax_irr"`
	PayableIRR            int64     `json:"payable_irr"`
	// StartsSubscription is true when paying this quote starts the seat
	// subscription: the period then begins at the purchase.
	StartsSubscription bool `json:"starts_subscription"`
}

// SeatRenewalLine is one block of seats a renewal pays for.
type SeatRenewalLine struct {
	BlockID     string    `json:"block_id"`
	Seat        string    `json:"seat"`
	Quantity    int       `json:"quantity"`
	PaidThrough time.Time `json:"paid_through"`
}

// SeatRenewalQuote is the price of carrying paid seats to the end of the
// subscription's current period. Amounts are IRR.
type SeatRenewalQuote struct {
	Lines      []SeatRenewalLine `json:"lines"`
	PeriodEnd  time.Time         `json:"period_end"`
	PriceIRR   int64             `json:"price_irr"`
	TaxPercent string            `json:"tax_percent"`
	TaxIRR     int64             `json:"tax_irr"`
	PayableIRR int64             `json:"payable_irr"`
}

// SeatRenewalPreviewMember is a member who loses a paid seat at the period end
// when the tenant keeps fewer seats than members hold them.
type SeatRenewalPreviewMember struct {
	UserID     string     `json:"user_id"`
	Email      string     `json:"email,omitempty"`
	Name       string     `json:"name,omitempty"`
	AssignedAt *time.Time `json:"assigned_at,omitempty"`
	// Fallback is "free" when the member gets a free seat after losing the paid
	// one, and "unassigned" when they are left without a seat.
	Fallback string `json:"fallback"`
}

// SeatRenewalPreviewLine is one paid seat type in a renewal preview.
type SeatRenewalPreviewLine struct {
	Seat string `json:"seat"`
	// Held is how many seats of the type the tenant holds, Renewed how many of
	// them are already paid for the next period, Keep how many stay after this
	// renewal, and Assigned how many members hold one.
	Held         int                        `json:"held"`
	Renewed      int                        `json:"renewed"`
	Keep         int                        `json:"keep"`
	Assigned     int                        `json:"assigned"`
	UnitPriceIRR int64                      `json:"unit_price_irr"`
	Losing       []SeatRenewalPreviewMember `json:"losing"`
}

// SeatRenewalPreview is what renewing a chosen number of seats of each type
// costs and which members lose their seat, without changing anything. Amounts
// are IRR.
type SeatRenewalPreview struct {
	PeriodEnd     time.Time                `json:"period_end"`
	NextPeriodEnd time.Time                `json:"next_period_end"`
	Lines         []SeatRenewalPreviewLine `json:"lines"`
	PriceIRR      int64                    `json:"price_irr"`
	TaxPercent    string                   `json:"tax_percent"`
	TaxIRR        int64                    `json:"tax_irr"`
	PayableIRR    int64                    `json:"payable_irr"`
}

// CreateSeatOrderRequest starts a seat purchase or renewal.
type CreateSeatOrderRequest struct {
	Kind     string `json:"kind,omitempty"`
	Seat     string `json:"seat,omitempty"`
	Quantity int    `json:"quantity,omitempty"`
	// BillingInterval is used only by the purchase that starts the seat subscription.
	BillingInterval string `json:"billing_interval,omitempty"`
	Provider        string `json:"provider"`
	ReturnURL       string `json:"return_url,omitempty"`
	// AssignUserID gives one of the bought seats to this member once paid.
	AssignUserID string `json:"assign_user_id,omitempty"`
	// RequestID approves this upgrade request once paid.
	RequestID string `json:"request_id,omitempty"`
}

// SeatOrder is a seat purchase or renewal and its payment state.
type SeatOrder struct {
	ID              string     `json:"-"`
	PublicID        string     `json:"id"`
	TenantID        string     `json:"tenant_id"`
	Kind            string     `json:"kind"`
	Seat            string     `json:"seat"`
	Quantity        int        `json:"quantity"`
	BillingInterval string     `json:"billing_interval"`
	PeriodStart     time.Time  `json:"period_start"`
	PeriodEnd       time.Time  `json:"period_end"`
	RemainingDays   int        `json:"remaining_days"`
	PeriodDays      int        `json:"period_days"`
	PriceIRR        int64      `json:"price_irr"`
	TaxPercent      string     `json:"tax_percent"`
	TaxIRR          int64      `json:"tax_irr"`
	PayableIRR      int64      `json:"payable_irr"`
	AssignUserID    string     `json:"assign_user_id,omitempty"`
	RequestID       string     `json:"request_id,omitempty"`
	RenewBlockIDs   []string   `json:"-"`
	Provider        string     `json:"provider"`
	Status          string     `json:"status"`
	Authority       string     `json:"authority,omitempty"`
	PaymentURL      string     `json:"payment_url,omitempty"`
	RefID           string     `json:"ref_id,omitempty"`
	FailureReason   string     `json:"failure_reason,omitempty"`
	CallbackURL     string     `json:"-"`
	ReturnURL       string     `json:"-"`
	CreatedAt       time.Time  `json:"created_at"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
}

// UpdateSeatRenewalRequest sets how many seats of a type the next renewal keeps.
type UpdateSeatRenewalRequest struct {
	Seat     string `json:"seat"`
	Quantity int    `json:"quantity"`
}

// SeatUpgradeRequest is a member asking for a better seat.
type SeatUpgradeRequest struct {
	ID            string     `json:"id"`
	TenantID      string     `json:"tenant_id"`
	UserID        string     `json:"user_id"`
	Email         string     `json:"email,omitempty"`
	Name          string     `json:"name,omitempty"`
	CurrentSeat   string     `json:"current_seat"`
	RequestedSeat string     `json:"requested_seat"`
	Reason        string     `json:"reason,omitempty"`
	BlockReason   string     `json:"block_reason,omitempty"`
	Status        string     `json:"status"`
	DecidedBy     string     `json:"decided_by,omitempty"`
	DecisionNote  string     `json:"decision_note,omitempty"`
	DecidedAt     *time.Time `json:"decided_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

// CreateSeatUpgradeRequest is a member's upgrade request.
type CreateSeatUpgradeRequest struct {
	Seat        string `json:"seat"`
	Reason      string `json:"reason,omitempty"`
	BlockReason string `json:"block_reason,omitempty"`
}

// DecideSeatUpgradeRequest denies a request with an optional note.
type DecideSeatUpgradeRequest struct {
	Note string `json:"note,omitempty"`
}

// SeatStatus is a member's own seat as the gateway reports it.
type SeatStatus struct {
	Enforced bool   `json:"enforced"`
	Type     string `json:"type"`
	// Credits are decimal strings with two places. MonthlyCredits is the total
	// of the member's active personal lots, the same lots RemainingCredits is
	// left of; it exceeds the seat quota when free-seat credits were kept.
	MonthlyCredits   string     `json:"monthly_credits"`
	RemainingCredits string     `json:"remaining_credits"`
	RenewsAt         *time.Time `json:"renews_at,omitempty"`
	Lifetime         bool       `json:"lifetime"`
	PoolAccess       bool       `json:"pool_access"`
	PoolLimitCredits *string    `json:"pool_limit_credits"`
	PoolUsedCredits  string     `json:"pool_used_credits"`
	// BlockReason is set when the member's next chat request would be refused.
	BlockReason        string `json:"block_reason,omitempty"`
	UpgradeRequestable bool   `json:"upgrade_requestable"`
	PendingRequest     string `json:"pending_request,omitempty"`
}

// SeatTypeUpdate changes one seat type's quota or discount (support admin).
type SeatTypeUpdate struct {
	MonthlyCredits        *int64  `json:"monthly_credits,omitempty"`
	DiscountPercent       *string `json:"discount_percent,omitempty"`
	YearlyDiscountPercent *string `json:"yearly_discount_percent,omitempty"`
	Active                *bool   `json:"active,omitempty"`
}
