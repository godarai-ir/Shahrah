package storage

import (
	"context"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *Store) CreateBudgetLimitRule(ctx context.Context, tenantID string, req platform.CreateBudgetLimitRuleRequest) (*platform.BudgetLimitRule, error) {
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	priority := 100
	if req.Priority != nil {
		priority = *req.Priority
	}

	blockOnExceed := true
	if req.BlockOnBudgetExceed != nil {
		blockOnExceed = *req.BlockOnBudgetExceed
	}

	conditionsJSON, err := marshalJSONObject(req.Conditions)
	if err != nil {
		return nil, err
	}

	alertsJSON, err := marshalJSONObject(req.Alerts)
	if err != nil {
		return nil, err
	}

	rows, err := s.pool.Query(ctx, `
		INSERT INTO budget_limit_rules (
			id, tenant_id, rule_id, name, description, enabled, priority, conditions_json,
			limit_to, unit, budget_applies_per, block_on_budget_exceed, alerts_json, created_at, updated_at
		) VALUES (
			$1, $2::uuid, $3, NULLIF($4,''), NULLIF($5,''), $6, $7, $8, $9, $10, NULLIF($11,''), $12, $13, NOW(), NOW()
		)
		RETURNING id::text, tenant_id::text, rule_id, COALESCE(name,''), COALESCE(description,''), enabled, priority,
		          conditions_json, limit_to, unit, budget_applies_per, block_on_budget_exceed, alerts_json, created_at, updated_at
	`, uuid.NewString(), tenantID, strings.TrimSpace(req.RuleID), strings.TrimSpace(req.Name), strings.TrimSpace(req.Description), enabled, priority, conditionsJSON, req.LimitTo, string(req.Unit), strings.TrimSpace(derefString(req.BudgetAppliesPer)), blockOnExceed, alertsJSON)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, pgx.ErrNoRows
	}

	item, err := scanBudgetLimitRule(rows)
	if err == nil {
		s.invalidateTenantRuntimeProfile(ctx, tenantID, "budget_limit_rule.created")
	}
	return item, err
}

func (s *Store) ListBudgetLimitRules(ctx context.Context, tenantID string, query string, limit int, offset int) ([]platform.BudgetLimitRule, int, error) {
	q, like, limit, offset := normalizePagedQuery(query, limit, offset)

	return listSearchResults(ctx, s.pool, `
		SELECT COUNT(1)
		FROM budget_limit_rules
		WHERE tenant_id = $1::uuid
		  AND ($2 = '' OR LOWER(rule_id) LIKE $3 OR LOWER(COALESCE(name,'')) LIKE $3 OR LOWER(COALESCE(description,'')) LIKE $3)
	`, []any{tenantID, q, like}, `
		SELECT id::text, tenant_id::text, rule_id, COALESCE(name,''), COALESCE(description,''), enabled, priority,
		       conditions_json, limit_to, unit, budget_applies_per, block_on_budget_exceed, alerts_json, created_at, updated_at
		FROM budget_limit_rules
		WHERE tenant_id = $1::uuid
		  AND ($2 = '' OR LOWER(rule_id) LIKE $3 OR LOWER(COALESCE(name,'')) LIKE $3 OR LOWER(COALESCE(description,'')) LIKE $3)
		ORDER BY priority ASC, updated_at DESC
		LIMIT $4 OFFSET $5
	`, []any{tenantID, q, like, limit, offset}, limit, scanBudgetLimitRule)
}

func (s *Store) GetBudgetLimitRuleByRuleID(ctx context.Context, tenantID string, ruleID string) (*platform.BudgetLimitRule, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, tenant_id::text, rule_id, COALESCE(name,''), COALESCE(description,''), enabled, priority,
		       conditions_json, limit_to, unit, budget_applies_per, block_on_budget_exceed, alerts_json, created_at, updated_at
		FROM budget_limit_rules
		WHERE tenant_id = $1::uuid AND rule_id = $2
		LIMIT 1
	`, tenantID, strings.TrimSpace(ruleID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, pgx.ErrNoRows
	}

	item, err := scanBudgetLimitRule(rows)
	if err == nil {
		s.invalidateTenantRuntimeProfile(ctx, tenantID, "budget_limit_rule.updated")
	}
	return item, err
}

func (s *Store) UpdateBudgetLimitRuleByRuleID(ctx context.Context, tenantID string, ruleID string, req platform.UpdateBudgetLimitRuleRequest) (*platform.BudgetLimitRule, error) {
	current, err := s.GetBudgetLimitRuleByRuleID(ctx, tenantID, ruleID)
	if err != nil {
		return nil, err
	}

	name := optionalTrimmedString(req.Name, current.Name)
	description := optionalTrimmedString(req.Description, current.Description)
	enabled := optionalBool(req.Enabled, current.Enabled)
	priority := optionalInt(req.Priority, current.Priority)

	conditions := current.Conditions
	if req.Conditions != nil {
		conditions = req.Conditions
	}

	limitTo := current.LimitTo
	if req.LimitTo != nil {
		limitTo = *req.LimitTo
	}

	unit := optionalBudgetLimitUnit(req.Unit, current.Unit)
	budgetAppliesPer := optionalStringPtr(req.BudgetAppliesPer, current.BudgetAppliesPer)
	blockOnExceed := optionalBool(req.BlockOnBudgetExceed, current.BlockOnBudgetExceed)

	alerts := current.Alerts
	if req.Alerts != nil {
		alerts = req.Alerts
	}

	conditionsJSON, err := marshalJSONObject(conditions)
	if err != nil {
		return nil, err
	}

	alertsJSON, err := marshalJSONObject(alerts)
	if err != nil {
		return nil, err
	}

	rows, err := s.pool.Query(ctx, `
		UPDATE budget_limit_rules
		SET name = NULLIF($3, ''),
		    description = NULLIF($4, ''),
		    enabled = $5,
		    priority = $6,
		    conditions_json = $7,
		    limit_to = $8,
		    unit = $9,
		    budget_applies_per = NULLIF($10, ''),
		    block_on_budget_exceed = $11,
		    alerts_json = $12,
		    updated_at = NOW()
		WHERE tenant_id = $1::uuid AND rule_id = $2
		RETURNING id::text, tenant_id::text, rule_id, COALESCE(name,''), COALESCE(description,''), enabled, priority,
		          conditions_json, limit_to, unit, budget_applies_per, block_on_budget_exceed, alerts_json, created_at, updated_at
	`, tenantID, strings.TrimSpace(ruleID), name, description, enabled, priority, conditionsJSON, limitTo, string(unit), strings.TrimSpace(derefString(budgetAppliesPer)), blockOnExceed, alertsJSON)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, pgx.ErrNoRows
	}

	item, err := scanBudgetLimitRule(rows)
	if err == nil {
		s.invalidateTenantRuntimeProfile(ctx, tenantID, "budget_limit_rule.status_updated")
	}
	return item, err
}

func optionalTrimmedString(value *string, fallback string) string {
	if value == nil {
		return fallback
	}

	return strings.TrimSpace(*value)
}

func optionalBool(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}

	return *value
}

func optionalInt(value *int, fallback int) int {
	if value == nil {
		return fallback
	}

	return *value
}

func optionalStringPtr(value, fallback *string) *string {
	if value == nil {
		return fallback
	}

	return value
}

func optionalBudgetLimitUnit(value *platform.BudgetLimitUnit, fallback platform.BudgetLimitUnit) platform.BudgetLimitUnit {
	if value == nil {
		return fallback
	}

	return *value
}

func (s *Store) UpdateBudgetLimitRuleStatusByRuleID(ctx context.Context, tenantID string, ruleID string, enabled bool) (*platform.BudgetLimitRule, error) {
	rows, err := s.pool.Query(ctx, `
		UPDATE budget_limit_rules
		SET enabled = $3, updated_at = NOW()
		WHERE tenant_id = $1::uuid AND rule_id = $2
		RETURNING id::text, tenant_id::text, rule_id, COALESCE(name,''), COALESCE(description,''), enabled, priority,
		          conditions_json, limit_to, unit, budget_applies_per, block_on_budget_exceed, alerts_json, created_at, updated_at
	`, tenantID, strings.TrimSpace(ruleID), enabled)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, pgx.ErrNoRows
	}

	item, err := scanBudgetLimitRule(rows)
	if err != nil {
		return nil, err
	}
	s.invalidateTenantRuntimeProfile(ctx, tenantID, "budget_limit_rule.status_updated")

	return item, nil
}

func (s *Store) DeleteBudgetLimitRuleByRuleID(ctx context.Context, tenantID string, ruleID string) error {
	result, err := s.pool.Exec(ctx, `
		DELETE FROM budget_limit_rules
		WHERE tenant_id = $1::uuid AND rule_id = $2
	`, tenantID, strings.TrimSpace(ruleID))
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	s.invalidateTenantRuntimeProfile(ctx, tenantID, "budget_limit_rule.deleted")

	return nil
}

func scanBudgetLimitRule(rows pgx.Rows) (*platform.BudgetLimitRule, error) {
	var (
		item             platform.BudgetLimitRule
		name             string
		description      string
		conditionsBytes  []byte
		alertsBytes      []byte
		budgetAppliesPer *string
		unit             string
	)
	if err := rows.Scan(
		&item.ID,
		&item.TenantID,
		&item.RuleID,
		&name,
		&description,
		&item.Enabled,
		&item.Priority,
		&conditionsBytes,
		&item.LimitTo,
		&unit,
		&budgetAppliesPer,
		&item.BlockOnBudgetExceed,
		&alertsBytes,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return nil, err
	}

	item.Name = name
	item.Description = description
	item.Conditions = parseJSONObject(conditionsBytes)
	item.Alerts = parseJSONObject(alertsBytes)
	item.Unit = platform.BudgetLimitUnit(unit)
	item.BudgetAppliesPer = budgetAppliesPer

	return &item, nil
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}
