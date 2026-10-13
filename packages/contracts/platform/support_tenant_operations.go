package platform

import "time"

type SupportTenantSummary struct {
	ID                 string     `json:"id"`
	Slug               string     `json:"slug"`
	Domain             string     `json:"domain"`
	Status             string     `json:"status"`
	CreatedAt          time.Time  `json:"created_at"`
	PlanSlug           string     `json:"plan_slug,omitempty"`
	PlanDisplayName    string     `json:"plan_display_name,omitempty"`
	SubscriptionStatus string     `json:"subscription_status,omitempty"`
	PeriodEnd          *time.Time `json:"period_end,omitempty"`
	WalletBalanceMinor int64      `json:"wallet_balance_minor"`
	WalletPendingMinor int64      `json:"wallet_pending_minor"`
}

type SupportTenantListResponse struct {
	Data       []SupportTenantSummary `json:"data"`
	NextCursor string                 `json:"next_cursor,omitempty"`
}

type SupportTenantDetail struct {
	Tenant       SupportTenantSummary           `json:"tenant"`
	Owner        *TenantOwnerContact            `json:"owner,omitempty"`
	Subscription *TenantSubscription            `json:"subscription,omitempty"`
	Entitlements *EffectiveEntitlementsResponse `json:"entitlements,omitempty"`
	Wallet       *Wallet                        `json:"wallet,omitempty"`
}

// TenantOwnerContact is the primary active tenant_admin account, selected by earliest creation time.
type TenantOwnerContact struct {
	UserID string `json:"user_id"`
	Name   string `json:"name,omitempty"`
	Email  string `json:"email,omitempty"`
	Phone  string `json:"phone,omitempty"`
}

type TenantEvent struct {
	ID        string                 `json:"id"`
	TenantID  string                 `json:"tenant_id"`
	Type      string                 `json:"type"`
	Severity  string                 `json:"severity"`
	ActorID   string                 `json:"actor_id,omitempty"`
	ActorRole string                 `json:"actor_role,omitempty"`
	Reason    string                 `json:"reason,omitempty"`
	RequestID string                 `json:"request_id,omitempty"`
	Payload   map[string]interface{} `json:"payload,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
}

type WalletAdjustmentRequest struct {
	Direction         string `json:"direction"`
	AmountMinor       int64  `json:"amount_minor"`
	Reason            string `json:"reason"`
	ExternalReference string `json:"external_reference,omitempty"`
	IdempotencyKey    string `json:"idempotency_key"`
}

type WalletAdjustmentResult struct {
	Transaction   Transaction `json:"transaction"`
	BalanceBefore int64       `json:"balance_before"`
	BalanceAfter  int64       `json:"balance_after"`
	Idempotent    bool        `json:"idempotent"`
}

type TenantStatusChangeRequest struct {
	Reason string `json:"reason"`
}

type SubscriptionOperatorActionRequest struct {
	Reason string `json:"reason"`
}

type SupportActor struct {
	ID        string
	Role      string
	RequestID string
}
