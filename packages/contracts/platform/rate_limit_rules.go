package platform

import "time"

type RateLimitConflictBehavior string

const (
	RateLimitConflictFirstMatch      RateLimitConflictBehavior = "first_match"
	RateLimitConflictMostRestrictive RateLimitConflictBehavior = "most_restrictive"
	RateLimitConflictApplyAll        RateLimitConflictBehavior = "apply_all"
)

type RateLimitResponseBehavior string

const (
	RateLimitResponseReturn429 RateLimitResponseBehavior = "return_429"
)

type RateLimitRule struct {
	ID               string                    `json:"id"`
	TenantID         string                    `json:"tenant_id,omitempty"`
	RuleID           string                    `json:"rule_id"`
	Name             string                    `json:"name,omitempty"`
	Description      string                    `json:"description,omitempty"`
	Enabled          bool                      `json:"enabled"`
	Priority         int                       `json:"priority"`
	Conditions       map[string]interface{}    `json:"conditions"`
	Limit            map[string]interface{}    `json:"limit"`
	ConflictBehavior RateLimitConflictBehavior `json:"conflict_behavior"`
	ResponseBehavior RateLimitResponseBehavior `json:"response_behavior"`
	SoftLimit        bool                      `json:"soft_limit"`
	RetryAfterHeader bool                      `json:"retry_after_header"`
	EffectiveFrom    *time.Time                `json:"effective_from,omitempty"`
	ExpiresAt        *time.Time                `json:"expires_at,omitempty"`
	CreatedAt        time.Time                 `json:"created_at"`
	UpdatedAt        time.Time                 `json:"updated_at"`
}

type CreateRateLimitRuleRequest struct {
	RuleID           string                    `json:"rule_id"`
	Name             string                    `json:"name,omitempty"`
	Description      string                    `json:"description,omitempty"`
	Enabled          *bool                     `json:"enabled,omitempty"`
	Priority         *int                      `json:"priority,omitempty"`
	Conditions       map[string]interface{}    `json:"conditions"`
	Limit            map[string]interface{}    `json:"limit"`
	ConflictBehavior RateLimitConflictBehavior `json:"conflict_behavior,omitempty"`
	ResponseBehavior RateLimitResponseBehavior `json:"response_behavior,omitempty"`
	SoftLimit        *bool                     `json:"soft_limit,omitempty"`
	RetryAfterHeader *bool                     `json:"retry_after_header,omitempty"`
	EffectiveFrom    *time.Time                `json:"effective_from,omitempty"`
	ExpiresAt        *time.Time                `json:"expires_at,omitempty"`
}

type UpdateRateLimitRuleRequest struct {
	Name             *string                    `json:"name,omitempty"`
	Description      *string                    `json:"description,omitempty"`
	Enabled          *bool                      `json:"enabled,omitempty"`
	Priority         *int                       `json:"priority,omitempty"`
	Conditions       map[string]interface{}     `json:"conditions,omitempty"`
	Limit            map[string]interface{}     `json:"limit,omitempty"`
	ConflictBehavior *RateLimitConflictBehavior `json:"conflict_behavior,omitempty"`
	ResponseBehavior *RateLimitResponseBehavior `json:"response_behavior,omitempty"`
	SoftLimit        *bool                      `json:"soft_limit,omitempty"`
	RetryAfterHeader *bool                      `json:"retry_after_header,omitempty"`
	EffectiveFrom    *time.Time                 `json:"effective_from,omitempty"`
	ExpiresAt        *time.Time                 `json:"expires_at,omitempty"`
}

type UpdateRateLimitRuleStatusRequest struct {
	Enabled bool `json:"enabled"`
}

type RateLimitRuleListResponse struct {
	Data       []RateLimitRule `json:"data"`
	Total      int             `json:"total"`
	Limit      int             `json:"limit"`
	Offset     int             `json:"offset"`
	NextOffset *int            `json:"next_offset,omitempty"`
}
