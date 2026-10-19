package platform

import "time"

type BudgetLimitUnit string

const (
	BudgetLimitCostPerDay   BudgetLimitUnit = "cost_per_day"
	BudgetLimitCostPerWeek  BudgetLimitUnit = "cost_per_week"
	BudgetLimitCostPerMonth BudgetLimitUnit = "cost_per_month"
)

type BudgetAppliesPer string

// BudgetOnExceed is what happens to a request that would exceed a credit budget.
type BudgetOnExceed string

const (
	// BudgetOnExceedBlock refuses the request with HTTP 402.
	BudgetOnExceedBlock BudgetOnExceed = "block"
	// BudgetOnExceedAlert allows the request and records an alert.
	BudgetOnExceedAlert BudgetOnExceed = "alert"
)

const (
	BudgetAppliesPerTenant         BudgetAppliesPer = "tenant"
	BudgetAppliesPerUser           BudgetAppliesPer = "user"
	BudgetAppliesPerTeam           BudgetAppliesPer = "team"
	BudgetAppliesPerModel          BudgetAppliesPer = "model"
	BudgetAppliesPerVirtualAccount BudgetAppliesPer = "virtualaccount"
)

type BudgetLimitRule struct {
	ID                  string                 `json:"id"`
	TenantID            string                 `json:"tenant_id,omitempty"`
	RuleID              string                 `json:"rule_id"`
	Name                string                 `json:"name,omitempty"`
	Description         string                 `json:"description,omitempty"`
	Enabled             bool                   `json:"enabled"`
	Priority            int                    `json:"priority"`
	Conditions          map[string]interface{} `json:"conditions"`
	LimitTo             float64                `json:"limit_to"`
	Unit                BudgetLimitUnit        `json:"unit"`
	BudgetAppliesPer    *string                `json:"budget_applies_per,omitempty"`
	BlockOnBudgetExceed bool                   `json:"block_on_budget_exceed"`
	Alerts              map[string]interface{} `json:"alerts,omitempty"`
	// Limit is the credit limit as a decimal string with two decimals.
	Limit *string `json:"limit,omitempty"`
	// LimitMC is Limit in millicredits; credit tenants enforce it.
	LimitMC         *int64         `json:"limit_mc,omitempty"`
	OnExceed        BudgetOnExceed `json:"on_exceed,omitempty"`
	AlertThresholds []int          `json:"alert_thresholds,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

type CreateBudgetLimitRuleRequest struct {
	RuleID              string                 `json:"rule_id"`
	Name                string                 `json:"name,omitempty"`
	Description         string                 `json:"description,omitempty"`
	Enabled             *bool                  `json:"enabled,omitempty"`
	Priority            *int                   `json:"priority,omitempty"`
	Conditions          map[string]interface{} `json:"conditions"`
	LimitTo             float64                `json:"limit_to"`
	Unit                BudgetLimitUnit        `json:"unit"`
	BudgetAppliesPer    *string                `json:"budget_applies_per,omitempty"`
	BlockOnBudgetExceed *bool                  `json:"block_on_budget_exceed,omitempty"`
	Alerts              map[string]interface{} `json:"alerts,omitempty"`
	Limit               *string                `json:"limit,omitempty"`
	OnExceed            *BudgetOnExceed        `json:"on_exceed,omitempty"`
	AlertThresholds     []int                  `json:"alert_thresholds,omitempty"`
}

type UpdateBudgetLimitRuleRequest struct {
	Name                *string                `json:"name,omitempty"`
	Description         *string                `json:"description,omitempty"`
	Enabled             *bool                  `json:"enabled,omitempty"`
	Priority            *int                   `json:"priority,omitempty"`
	Conditions          map[string]interface{} `json:"conditions,omitempty"`
	LimitTo             *float64               `json:"limit_to,omitempty"`
	Unit                *BudgetLimitUnit       `json:"unit,omitempty"`
	BudgetAppliesPer    *string                `json:"budget_applies_per,omitempty"`
	BlockOnBudgetExceed *bool                  `json:"block_on_budget_exceed,omitempty"`
	Alerts              map[string]interface{} `json:"alerts,omitempty"`
	Limit               *string                `json:"limit,omitempty"`
	OnExceed            *BudgetOnExceed        `json:"on_exceed,omitempty"`
	AlertThresholds     []int                  `json:"alert_thresholds,omitempty"`
}

type UpdateBudgetLimitRuleStatusRequest struct {
	Enabled bool `json:"enabled"`
}

type BudgetLimitRuleListResponse struct {
	Data       []BudgetLimitRule `json:"data"`
	Total      int               `json:"total"`
	Limit      int               `json:"limit"`
	Offset     int               `json:"offset"`
	NextOffset *int              `json:"next_offset,omitempty"`
}
