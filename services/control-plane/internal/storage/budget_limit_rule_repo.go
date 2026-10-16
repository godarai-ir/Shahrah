package storage

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/apperrors"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/credits"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// budgetLimitRuleColumns is the column list scanBudgetLimitRule reads.
const budgetLimitRuleColumns = `id::text, tenant_id::text, rule_id, COALESCE(name,''), COALESCE(description,''), enabled, priority,
		conditions_json, limit_to, unit, budget_applies_per, block_on_budget_exceed, alerts_json,
		limit_mc, on_exceed, alert_thresholds, created_at, updated_at`

// budgetRuleCreditFields are the credit enforcement columns of a rule.
type budgetRuleCreditFields struct {
	limitMC        *int64
	onExceed       platform.BudgetOnExceed
	thresholdsJSON []byte
}

// resolveBudgetRuleCreditFields applies requested credit fields over the
// current ones (zero values for a new rule). A nil field keeps the current
// value and an empty limit string clears it.
func resolveBudgetRuleCreditFields(current platform.BudgetLimitRule, limit *string, onExceed *platform.BudgetOnExceed, block *bool, thresholds []int) (budgetRuleCreditFields, error) {
	resolvedOnExceed, err := credits.ResolveOnExceed(onExceed, block, current.OnExceed)
	if err != nil {
		return budgetRuleCreditFields{}, err
	}

	fields := budgetRuleCreditFields{limitMC: current.LimitMC, onExceed: resolvedOnExceed}

	if fields.limitMC, err = optionalCreditAmountMC(limit, fields.limitMC); err != nil {
		return budgetRuleCreditFields{}, err
	}

	if thresholds == nil {
		thresholds = current.AlertThresholds
	}

	normalized, err := credits.NormalizeAlertThresholds(thresholds)
	if err != nil {
		return budgetRuleCreditFields{}, err
	}

	if fields.thresholdsJSON, err = json.Marshal(normalized); err != nil {
		return budgetRuleCreditFields{}, err
	}

	return fields, nil
}

func optionalCreditAmountMC(raw *string, current *int64) (*int64, error) {
	if raw == nil {
		return current, nil
	}

	if strings.TrimSpace(*raw) == "" {
		return nil, nil
	}

	amount, err := credits.ParseCreditAmountMC(*raw)
	if err != nil {
		return nil, err
	}

	return &amount, nil
}

func (s *PolicyRepository) CreateBudgetLimitRule(ctx context.Context, tenantID string, req platform.CreateBudgetLimitRuleRequest) (*platform.BudgetLimitRule, error) {
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	priority := 100
	if req.Priority != nil {
		priority = *req.Priority
	}

	creditFields, err := resolveBudgetRuleCreditFields(platform.BudgetLimitRule{}, req.Limit, req.OnExceed, req.BlockOnBudgetExceed, req.AlertThresholds)
	if err != nil {
		return nil, err
	}

	blockOnExceed := credits.OnExceedBlocks(creditFields.onExceed)

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
			limit_to, unit, budget_applies_per, block_on_budget_exceed, alerts_json,
			limit_mc, on_exceed, alert_thresholds, created_at, updated_at
		) VALUES (
			$1, $2::uuid, $3, NULLIF($4,''), NULLIF($5,''), $6, $7, $8, $9, $10, NULLIF($11,''), $12, $13,
			$14, $15, $16, NOW(), NOW()
		)
		RETURNING `+budgetLimitRuleColumns+`
	`, uuid.NewString(), tenantID, strings.TrimSpace(req.RuleID), strings.TrimSpace(req.Name), strings.TrimSpace(req.Description), enabled, priority, conditionsJSON, req.LimitTo, string(req.Unit), strings.TrimSpace(derefString(req.BudgetAppliesPer)), blockOnExceed, alertsJSON,
		creditFields.limitMC, string(creditFields.onExceed), creditFields.thresholdsJSON)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, apperrors.ErrNotFound
	}

	item, err := scanBudgetLimitRule(rows)
	if err == nil {
		s.invalidateTenantRuntimeProfile(ctx, tenantID, "budget_limit_rule.created")
	}

	return item, err
}

func (s *PolicyRepository) ListBudgetLimitRules(ctx context.Context, tenantID string, query string, limit int, offset int) ([]platform.BudgetLimitRule, int, error) {
	q, like, limit, offset := normalizePagedQuery(query, limit, offset)

	return listSearchResults(ctx, s.pool, `
		SELECT COUNT(1)
		FROM budget_limit_rules
		WHERE tenant_id = $1::uuid
		  AND ($2 = '' OR LOWER(rule_id) LIKE $3 OR LOWER(COALESCE(name,'')) LIKE $3 OR LOWER(COALESCE(description,'')) LIKE $3)
	`, []any{tenantID, q, like}, `
		SELECT `+budgetLimitRuleColumns+`
		FROM budget_limit_rules
		WHERE tenant_id = $1::uuid
		  AND ($2 = '' OR LOWER(rule_id) LIKE $3 OR LOWER(COALESCE(name,'')) LIKE $3 OR LOWER(COALESCE(description,'')) LIKE $3)
		ORDER BY priority ASC, updated_at DESC
		LIMIT $4 OFFSET $5
	`, []any{tenantID, q, like, limit, offset}, limit, scanBudgetLimitRule)
}

func (s *PolicyRepository) GetBudgetLimitRuleByRuleID(ctx context.Context, tenantID string, ruleID string) (*platform.BudgetLimitRule, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT `+budgetLimitRuleColumns+`
		FROM budget_limit_rules
		WHERE tenant_id = $1::uuid AND rule_id = $2
		LIMIT 1
	`, tenantID, strings.TrimSpace(ruleID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, apperrors.ErrNotFound
	}

	item, err := scanBudgetLimitRule(rows)
	if err == nil {
		s.invalidateTenantRuntimeProfile(ctx, tenantID, "budget_limit_rule.updated")
	}

	return item, err
}

func (s *PolicyRepository) UpdateBudgetLimitRuleByRuleID(ctx context.Context, tenantID string, ruleID string, req platform.UpdateBudgetLimitRuleRequest) (*platform.BudgetLimitRule, error) {
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

	creditFields, err := resolveBudgetRuleCreditFields(*current, req.Limit, req.OnExceed, req.BlockOnBudgetExceed, req.AlertThresholds)
	if err != nil {
		return nil, err
	}

	blockOnExceed := credits.OnExceedBlocks(creditFields.onExceed)

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
		    limit_mc = $13,
		    on_exceed = $14,
		    alert_thresholds = $15,
		    updated_at = NOW()
		WHERE tenant_id = $1::uuid AND rule_id = $2
		RETURNING `+budgetLimitRuleColumns+`
	`, tenantID, strings.TrimSpace(ruleID), name, description, enabled, priority, conditionsJSON, limitTo, string(unit), strings.TrimSpace(derefString(budgetAppliesPer)), blockOnExceed, alertsJSON,
		creditFields.limitMC, string(creditFields.onExceed), creditFields.thresholdsJSON)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, apperrors.ErrNotFound
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

func (s *PolicyRepository) UpdateBudgetLimitRuleStatusByRuleID(ctx context.Context, tenantID string, ruleID string, enabled bool) (*platform.BudgetLimitRule, error) {
	rows, err := s.pool.Query(ctx, `
		UPDATE budget_limit_rules
		SET enabled = $3, updated_at = NOW()
		WHERE tenant_id = $1::uuid AND rule_id = $2
		RETURNING `+budgetLimitRuleColumns+`
	`, tenantID, strings.TrimSpace(ruleID), enabled)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, apperrors.ErrNotFound
	}

	item, err := scanBudgetLimitRule(rows)
	if err != nil {
		return nil, err
	}

	s.invalidateTenantRuntimeProfile(ctx, tenantID, "budget_limit_rule.status_updated")

	return item, nil
}

func (s *PolicyRepository) DeleteBudgetLimitRuleByRuleID(ctx context.Context, tenantID string, ruleID string) error {
	result, err := s.pool.Exec(ctx, `
		DELETE FROM budget_limit_rules
		WHERE tenant_id = $1::uuid AND rule_id = $2
	`, tenantID, strings.TrimSpace(ruleID))
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return apperrors.ErrNotFound
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
		onExceed         string
		thresholdsBytes  []byte
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
		&item.LimitMC,
		&onExceed,
		&thresholdsBytes,
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
	item.OnExceed = platform.BudgetOnExceed(onExceed)
	item.Limit = formatOptionalCredits(item.LimitMC)

	if err := json.Unmarshal(thresholdsBytes, &item.AlertThresholds); err != nil {
		return nil, err
	}

	return &item, nil
}

func formatOptionalCredits(amountMC *int64) *string {
	if amountMC == nil {
		return nil
	}

	formatted := credits.Format(credits.Millicredits(*amountMC))

	return &formatted
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}
