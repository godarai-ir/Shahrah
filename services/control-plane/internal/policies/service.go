// Package policies owns tenant gateway policy application boundaries.
package policies

import (
	"context"
	"errors"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

var errRepositoryRequired = errors.New("policies: repository is required")

type rateLimitRepository interface {
	CreateRateLimitRule(context.Context, string, platform.CreateRateLimitRuleRequest) (*platform.RateLimitRule, error)
	ListRateLimitRules(context.Context, string, string, int, int) ([]platform.RateLimitRule, int, error)
	GetRateLimitRuleByRuleID(context.Context, string, string) (*platform.RateLimitRule, error)
	UpdateRateLimitRuleByRuleID(context.Context, string, string, platform.UpdateRateLimitRuleRequest) (*platform.RateLimitRule, error)
	UpdateRateLimitRuleStatusByRuleID(context.Context, string, string, bool) (*platform.RateLimitRule, error)
	DeleteRateLimitRuleByRuleID(context.Context, string, string) error
}

type budgetLimitRepository interface {
	CreateBudgetLimitRule(context.Context, string, platform.CreateBudgetLimitRuleRequest) (*platform.BudgetLimitRule, error)
	ListBudgetLimitRules(context.Context, string, string, int, int) ([]platform.BudgetLimitRule, int, error)
	GetBudgetLimitRuleByRuleID(context.Context, string, string) (*platform.BudgetLimitRule, error)
	UpdateBudgetLimitRuleByRuleID(context.Context, string, string, platform.UpdateBudgetLimitRuleRequest) (*platform.BudgetLimitRule, error)
	UpdateBudgetLimitRuleStatusByRuleID(context.Context, string, string, bool) (*platform.BudgetLimitRule, error)
	DeleteBudgetLimitRuleByRuleID(context.Context, string, string) error
}

type guardrailRepository interface {
	CreateTenantGuardrail(context.Context, string, platform.CreateTenantGuardrailRequest, string) (*platform.TenantGuardrail, error)
	ListTenantGuardrails(context.Context, string, string, int, int) ([]platform.TenantGuardrail, int, error)
	GetTenantGuardrailBySlug(context.Context, string, string) (*platform.TenantGuardrail, error)
	UpdateTenantGuardrailBySlug(context.Context, string, string, platform.UpdateTenantGuardrailRequest, string) (*platform.TenantGuardrail, error)
	DeleteTenantGuardrailBySlug(context.Context, string, string) error
	ListEnabledTenantGuardrailsBySlugs(context.Context, string, []string) ([]platform.TenantGuardrail, error)
	ListEnabledTenantGuardrails(context.Context, string) ([]platform.TenantGuardrail, error)
}

type repository interface {
	rateLimitRepository
	budgetLimitRepository
	guardrailRepository
}

// Service is the application boundary for tenant gateway policies.
type Service struct {
	rateLimits rateLimitRepository
	budgets    budgetLimitRepository
	guardrails guardrailRepository
}

// NewService creates a policy application service.
func NewService(repo repository) (*Service, error) {
	if repo == nil {
		return nil, errRepositoryRequired
	}

	return &Service{rateLimits: repo, budgets: repo, guardrails: repo}, nil
}

// CreateRateLimitRule creates a tenant rate-limit rule.
func (s *Service) CreateRateLimitRule(ctx context.Context, tenantID string, request platform.CreateRateLimitRuleRequest) (*platform.RateLimitRule, error) {
	return s.rateLimits.CreateRateLimitRule(ctx, tenantID, request)
}

// ListRateLimitRules lists tenant rate-limit rules.
func (s *Service) ListRateLimitRules(ctx context.Context, tenantID, query string, limit, offset int) ([]platform.RateLimitRule, int, error) {
	return s.rateLimits.ListRateLimitRules(ctx, tenantID, query, limit, offset)
}

// GetRateLimitRuleByRuleID gets a rate-limit rule.
func (s *Service) GetRateLimitRuleByRuleID(ctx context.Context, tenantID, ruleID string) (*platform.RateLimitRule, error) {
	return s.rateLimits.GetRateLimitRuleByRuleID(ctx, tenantID, ruleID)
}

// UpdateRateLimitRuleByRuleID updates a rate-limit rule.
func (s *Service) UpdateRateLimitRuleByRuleID(ctx context.Context, tenantID, ruleID string, request platform.UpdateRateLimitRuleRequest) (*platform.RateLimitRule, error) {
	return s.rateLimits.UpdateRateLimitRuleByRuleID(ctx, tenantID, ruleID, request)
}

// UpdateRateLimitRuleStatusByRuleID changes a rate-limit rule status.
func (s *Service) UpdateRateLimitRuleStatusByRuleID(ctx context.Context, tenantID, ruleID string, enabled bool) (*platform.RateLimitRule, error) {
	return s.rateLimits.UpdateRateLimitRuleStatusByRuleID(ctx, tenantID, ruleID, enabled)
}

// DeleteRateLimitRuleByRuleID deletes a rate-limit rule.
func (s *Service) DeleteRateLimitRuleByRuleID(ctx context.Context, tenantID, ruleID string) error {
	return s.rateLimits.DeleteRateLimitRuleByRuleID(ctx, tenantID, ruleID)
}

// CreateBudgetLimitRule creates a tenant budget-limit rule.
func (s *Service) CreateBudgetLimitRule(ctx context.Context, tenantID string, request platform.CreateBudgetLimitRuleRequest) (*platform.BudgetLimitRule, error) {
	return s.budgets.CreateBudgetLimitRule(ctx, tenantID, request)
}

// ListBudgetLimitRules lists tenant budget-limit rules.
func (s *Service) ListBudgetLimitRules(ctx context.Context, tenantID, query string, limit, offset int) ([]platform.BudgetLimitRule, int, error) {
	return s.budgets.ListBudgetLimitRules(ctx, tenantID, query, limit, offset)
}

// GetBudgetLimitRuleByRuleID gets a budget-limit rule.
func (s *Service) GetBudgetLimitRuleByRuleID(ctx context.Context, tenantID, ruleID string) (*platform.BudgetLimitRule, error) {
	return s.budgets.GetBudgetLimitRuleByRuleID(ctx, tenantID, ruleID)
}

// UpdateBudgetLimitRuleByRuleID updates a budget-limit rule.
func (s *Service) UpdateBudgetLimitRuleByRuleID(ctx context.Context, tenantID, ruleID string, request platform.UpdateBudgetLimitRuleRequest) (*platform.BudgetLimitRule, error) {
	return s.budgets.UpdateBudgetLimitRuleByRuleID(ctx, tenantID, ruleID, request)
}

// UpdateBudgetLimitRuleStatusByRuleID changes a budget-limit rule status.
func (s *Service) UpdateBudgetLimitRuleStatusByRuleID(ctx context.Context, tenantID, ruleID string, enabled bool) (*platform.BudgetLimitRule, error) {
	return s.budgets.UpdateBudgetLimitRuleStatusByRuleID(ctx, tenantID, ruleID, enabled)
}

// DeleteBudgetLimitRuleByRuleID deletes a budget-limit rule.
func (s *Service) DeleteBudgetLimitRuleByRuleID(ctx context.Context, tenantID, ruleID string) error {
	return s.budgets.DeleteBudgetLimitRuleByRuleID(ctx, tenantID, ruleID)
}

// CreateTenantGuardrail creates a tenant guardrail.
func (s *Service) CreateTenantGuardrail(ctx context.Context, tenantID string, request platform.CreateTenantGuardrailRequest, actor string) (*platform.TenantGuardrail, error) {
	return s.guardrails.CreateTenantGuardrail(ctx, tenantID, request, actor)
}

// ListTenantGuardrails lists tenant guardrails.
func (s *Service) ListTenantGuardrails(ctx context.Context, tenantID, query string, limit, offset int) ([]platform.TenantGuardrail, int, error) {
	return s.guardrails.ListTenantGuardrails(ctx, tenantID, query, limit, offset)
}

// GetTenantGuardrailBySlug gets a tenant guardrail.
func (s *Service) GetTenantGuardrailBySlug(ctx context.Context, tenantID, slug string) (*platform.TenantGuardrail, error) {
	return s.guardrails.GetTenantGuardrailBySlug(ctx, tenantID, slug)
}

// UpdateTenantGuardrailBySlug updates a tenant guardrail.
func (s *Service) UpdateTenantGuardrailBySlug(ctx context.Context, tenantID, slug string, request platform.UpdateTenantGuardrailRequest, actor string) (*platform.TenantGuardrail, error) {
	return s.guardrails.UpdateTenantGuardrailBySlug(ctx, tenantID, slug, request, actor)
}

// DeleteTenantGuardrailBySlug deletes a tenant guardrail.
func (s *Service) DeleteTenantGuardrailBySlug(ctx context.Context, tenantID, slug string) error {
	return s.guardrails.DeleteTenantGuardrailBySlug(ctx, tenantID, slug)
}

// ListEnabledTenantGuardrailsBySlugs lists enabled guardrails selected by slug.
func (s *Service) ListEnabledTenantGuardrailsBySlugs(ctx context.Context, tenantID string, slugs []string) ([]platform.TenantGuardrail, error) {
	return s.guardrails.ListEnabledTenantGuardrailsBySlugs(ctx, tenantID, slugs)
}

// ListEnabledTenantGuardrails lists all enabled tenant guardrails.
func (s *Service) ListEnabledTenantGuardrails(ctx context.Context, tenantID string) ([]platform.TenantGuardrail, error) {
	return s.guardrails.ListEnabledTenantGuardrails(ctx, tenantID)
}
