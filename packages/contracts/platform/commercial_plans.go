package platform

import "time"

const (
	FeatureGatewayRequestsMonthly  = "gateway.requests.monthly"
	FeatureUsersMax                = "users.max"
	FeatureTeamsEnabled            = "teams.enabled"
	FeatureRateLimitRulesEnabled   = "rate_limit_rules.enabled"
	FeatureBudgetLimitRulesEnabled = "budget_limit_rules.enabled"
	FeatureTracesRetentionDays     = "traces.retention_days"
	FeatureVATEnabled              = "auth.vat.enabled"
	FeatureRBACEnabled             = "rbac.enabled"
	FeatureOrgUnitsEnabled         = "org_units.enabled"
)

type CommercialPlan struct {
	ID          string                 `json:"id"`
	Slug        string                 `json:"slug"`
	DisplayName string                 `json:"display_name"`
	Description string                 `json:"description,omitempty"`
	SortOrder   int                    `json:"sort_order"`
	Active      bool                   `json:"active"`
	Versions    []PlanVersion          `json:"versions,omitempty"`
	Prices      []PlanPrice            `json:"prices,omitempty"`
	Features    map[string]interface{} `json:"features,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

type PlanVersion struct {
	ID            string     `json:"id"`
	PlanID        string     `json:"plan_id"`
	Version       int        `json:"version"`
	Status        string     `json:"status"`
	EffectiveFrom *time.Time `json:"effective_from,omitempty"`
	PublishedAt   *time.Time `json:"published_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type FeatureDefinition struct {
	Key           string      `json:"key"`
	ValueType     string      `json:"value_type"`
	DefaultValue  interface{} `json:"default_value"`
	AllowOverride bool        `json:"allow_override"`
	Category      string      `json:"category,omitempty"`
	Description   string      `json:"description,omitempty"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}

type PlanEntitlement struct {
	ID         string      `json:"id"`
	VersionID  string      `json:"version_id"`
	FeatureKey string      `json:"feature_key"`
	Value      interface{} `json:"value"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

type PlanPrice struct {
	ID              string    `json:"id"`
	VersionID       string    `json:"version_id"`
	BillingInterval string    `json:"billing_interval"`
	Currency        string    `json:"currency"`
	AmountMinor     *float64  `json:"amount_minor,omitempty"`
	DiscountPercent float64   `json:"discount_percent"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type TenantSubscription struct {
	ID                 string                 `json:"id"`
	TenantID           string                 `json:"tenant_id"`
	PlanID             string                 `json:"plan_id"`
	PlanSlug           string                 `json:"plan_slug,omitempty"`
	PlanDisplayName    string                 `json:"plan_display_name,omitempty"`
	VersionID          string                 `json:"version_id"`
	Version            int                    `json:"version,omitempty"`
	Status             string                 `json:"status"`
	BillingInterval    string                 `json:"billing_interval"`
	CurrentPeriodStart time.Time              `json:"current_period_start"`
	CurrentPeriodEnd   time.Time              `json:"current_period_end"`
	TrialStartedAt     *time.Time             `json:"trial_started_at,omitempty"`
	TrialEndsAt        *time.Time             `json:"trial_ends_at,omitempty"`
	ConvertedAt        *time.Time             `json:"converted_at,omitempty"`
	CancelAtPeriodEnd  bool                   `json:"cancel_at_period_end"`
	PendingVersionID   string                 `json:"pending_version_id,omitempty"`
	PriceSnapshot      map[string]interface{} `json:"price_snapshot,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
}

type CommercialSignupSettings struct {
	TrialDays       int       `json:"trial_days"`
	SignupCreditIRR int64     `json:"signup_credit_irr"`
	UpdatedBy       string    `json:"updated_by,omitempty"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type UpdateCommercialSignupSettingsRequest struct {
	SignupCreditIRR *int64 `json:"signup_credit_irr"`
}

type TenantEntitlementOverride struct {
	ID         string      `json:"id"`
	TenantID   string      `json:"tenant_id"`
	FeatureKey string      `json:"feature_key"`
	Value      interface{} `json:"value"`
	Reason     string      `json:"reason,omitempty"`
	Active     bool        `json:"active"`
	ExpiresAt  *time.Time  `json:"expires_at,omitempty"`
	CreatedBy  string      `json:"created_by,omitempty"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

type EffectiveEntitlementsResponse struct {
	TenantID      string                 `json:"tenant_id"`
	Subscription  *TenantSubscription    `json:"subscription,omitempty"`
	Features      map[string]interface{} `json:"features"`
	Overrides     map[string]interface{} `json:"overrides,omitempty"`
	UsageCounters []UsageCounter         `json:"usage_counters,omitempty"`
	ResolvedAt    time.Time              `json:"resolved_at"`
}

type UsageCounter struct {
	FeatureKey  string    `json:"feature_key"`
	WindowStart time.Time `json:"window_start"`
	WindowEnd   time.Time `json:"window_end"`
	Used        float64   `json:"used"`
	Limit       *float64  `json:"limit,omitempty"`
}

type SubscriptionPaymentIntent struct {
	ID                      string                 `json:"id"`
	PublicID                string                 `json:"public_id"`
	TenantID                string                 `json:"tenant_id"`
	PlanID                  string                 `json:"plan_id"`
	PlanSlug                string                 `json:"plan_slug,omitempty"`
	VersionID               string                 `json:"version_id"`
	Version                 int                    `json:"version,omitempty"`
	BillingInterval         string                 `json:"billing_interval"`
	Provider                string                 `json:"provider"`
	Status                  string                 `json:"status"`
	Currency                string                 `json:"currency"`
	AmountMinor             float64                `json:"amount_minor"`
	DiscountPercent         float64                `json:"discount_percent"`
	PriceSnapshot           map[string]interface{} `json:"price_snapshot,omitempty"`
	Authority               string                 `json:"authority,omitempty"`
	PaymentURL              string                 `json:"payment_url,omitempty"`
	CallbackURL             string                 `json:"callback_url,omitempty"`
	ReturnURL               string                 `json:"return_url,omitempty"`
	RefID                   string                 `json:"ref_id,omitempty"`
	GatewayTxnID            string                 `json:"gateway_txn_id,omitempty"`
	CardMask                string                 `json:"card_mask,omitempty"`
	PayerPhone              string                 `json:"payer_phone,omitempty"`
	FailureReason           string                 `json:"failure_reason,omitempty"`
	Metadata                map[string]interface{} `json:"metadata,omitempty"`
	RawPayload              map[string]interface{} `json:"raw_payload,omitempty"`
	CompletedSubscriptionID string                 `json:"completed_subscription_id,omitempty"`
	CreatedAt               time.Time              `json:"created_at"`
	UpdatedAt               time.Time              `json:"updated_at"`
	CompletedAt             *time.Time             `json:"completed_at,omitempty"`
}

type SubscriptionChangeResult struct {
	Mode          string                     `json:"mode"`
	Subscription  *TenantSubscription        `json:"subscription,omitempty"`
	PaymentIntent *SubscriptionPaymentIntent `json:"payment_intent,omitempty"`
}

type CreateCommercialPlanRequest struct {
	Slug        string `json:"slug"`
	DisplayName string `json:"display_name"`
	Description string `json:"description,omitempty"`
	SortOrder   int    `json:"sort_order,omitempty"`
	Active      *bool  `json:"active,omitempty"`
}

type UpdateCommercialPlanRequest struct {
	DisplayName *string `json:"display_name,omitempty"`
	Description *string `json:"description,omitempty"`
	SortOrder   *int    `json:"sort_order,omitempty"`
	Active      *bool   `json:"active,omitempty"`
}

type CreatePlanVersionRequest struct {
	Version         int        `json:"version"`
	EffectiveFrom   *time.Time `json:"effective_from,omitempty"`
	CopyFromVersion *int       `json:"copy_from_version,omitempty"`
}

type UpsertPlanEntitlementRequest struct {
	FeatureKey string      `json:"feature_key"`
	Value      interface{} `json:"value"`
}

type UpsertPlanPriceRequest struct {
	BillingInterval string   `json:"billing_interval"`
	Currency        string   `json:"currency,omitempty"`
	AmountMinor     *float64 `json:"amount_minor,omitempty"`
	DiscountPercent float64  `json:"discount_percent,omitempty"`
}

type ChangeTenantSubscriptionRequest struct {
	PlanSlug        string                 `json:"plan_slug"`
	BillingInterval string                 `json:"billing_interval,omitempty"`
	Provider        string                 `json:"provider,omitempty"`
	ReturnURL       string                 `json:"return_url,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

type UpsertTenantSubscriptionRequest struct {
	TenantID        string                 `json:"tenant_id"`
	PlanSlug        string                 `json:"plan_slug"`
	BillingInterval string                 `json:"billing_interval,omitempty"`
	Status          string                 `json:"status,omitempty"`
	PeriodStart     *time.Time             `json:"period_start,omitempty"`
	PeriodEnd       *time.Time             `json:"period_end,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

type UpsertTenantEntitlementOverrideRequest struct {
	TenantID   string      `json:"tenant_id"`
	FeatureKey string      `json:"feature_key"`
	Value      interface{} `json:"value"`
	Reason     string      `json:"reason,omitempty"`
	ExpiresAt  *time.Time  `json:"expires_at,omitempty"`
	CreatedBy  string      `json:"created_by,omitempty"`
}
