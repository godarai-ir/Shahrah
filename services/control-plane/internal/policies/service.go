// Package policies owns tenant gateway policy application boundaries.
package policies

import (
	"context"
	"errors"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

var errRepositoryRequired = errors.New("policies: repository is required")

type repository interface {
	CreateRateLimitRule(context.Context, string, platform.CreateRateLimitRuleRequest) (*platform.RateLimitRule, error)
	ListRateLimitRules(context.Context, string, string, int, int) ([]platform.RateLimitRule, int, error)
	GetRateLimitRuleByRuleID(context.Context, string, string) (*platform.RateLimitRule, error)
	UpdateRateLimitRuleByRuleID(context.Context, string, string, platform.UpdateRateLimitRuleRequest) (*platform.RateLimitRule, error)
	UpdateRateLimitRuleStatusByRuleID(context.Context, string, string, bool) (*platform.RateLimitRule, error)
	DeleteRateLimitRuleByRuleID(context.Context, string, string) error
	CreateBudgetLimitRule(context.Context, string, platform.CreateBudgetLimitRuleRequest) (*platform.BudgetLimitRule, error)
	ListBudgetLimitRules(context.Context, string, string, int, int) ([]platform.BudgetLimitRule, int, error)
	GetBudgetLimitRuleByRuleID(context.Context, string, string) (*platform.BudgetLimitRule, error)
	UpdateBudgetLimitRuleByRuleID(context.Context, string, string, platform.UpdateBudgetLimitRuleRequest) (*platform.BudgetLimitRule, error)
	UpdateBudgetLimitRuleStatusByRuleID(context.Context, string, string, bool) (*platform.BudgetLimitRule, error)
	DeleteBudgetLimitRuleByRuleID(context.Context, string, string) error
	CreateTenantGuardrail(context.Context, string, platform.CreateTenantGuardrailRequest, string) (*platform.TenantGuardrail, error)
	ListTenantGuardrails(context.Context, string, string, int, int) ([]platform.TenantGuardrail, int, error)
	GetTenantGuardrailBySlug(context.Context, string, string) (*platform.TenantGuardrail, error)
	UpdateTenantGuardrailBySlug(context.Context, string, string, platform.UpdateTenantGuardrailRequest, string) (*platform.TenantGuardrail, error)
	DeleteTenantGuardrailBySlug(context.Context, string, string) error
	ListEnabledTenantGuardrailsBySlugs(context.Context, string, []string) ([]platform.TenantGuardrail, error)
	ListEnabledTenantGuardrails(context.Context, string) ([]platform.TenantGuardrail, error)
}

// Service is the application boundary for tenant gateway policies.
type Service struct {
	repository
}

// NewService creates a policy application service.
func NewService(repo repository) (*Service, error) {
	if repo == nil {
		return nil, errRepositoryRequired
	}

	return &Service{repository: repo}, nil
}
