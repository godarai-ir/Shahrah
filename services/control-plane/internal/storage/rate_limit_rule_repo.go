package storage

import (
	"context"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *Store) CreateRateLimitRule(ctx context.Context, tenantID string, req platform.CreateRateLimitRuleRequest) (*platform.RateLimitRule, error) {
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	priority := 100
	if req.Priority != nil {
		priority = *req.Priority
	}
	conflict := platform.RateLimitConflictMostRestrictive
	if req.ConflictBehavior != "" {
		conflict = req.ConflictBehavior
	}
	response := platform.RateLimitResponseReturn429
	if req.ResponseBehavior != "" {
		response = req.ResponseBehavior
	}
	softLimit := false
	if req.SoftLimit != nil {
		softLimit = *req.SoftLimit
	}
	retryAfter := true
	if req.RetryAfterHeader != nil {
		retryAfter = *req.RetryAfterHeader
	}
	conditionsJSON, err := marshalJSONObject(req.Conditions)
	if err != nil {
		return nil, err
	}
	limitJSON, err := marshalJSONObject(req.Limit)
	if err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx, `
		INSERT INTO rate_limit_rules (
			id, tenant_id, rule_id, name, description, enabled, priority, conditions_json, limit_json,
			conflict_behavior, response_behavior, soft_limit, retry_after_header, effective_from, expires_at, created_at, updated_at
		) VALUES (
			$1, $2::uuid, $3, NULLIF($4,''), NULLIF($5,''), $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, NOW(), NOW()
		)
		RETURNING id::text, tenant_id::text, rule_id, COALESCE(name,''), COALESCE(description,''), enabled, priority,
		          conditions_json, limit_json, conflict_behavior, response_behavior, soft_limit, retry_after_header,
		          effective_from, expires_at, created_at, updated_at
	`, uuid.NewString(), tenantID, strings.TrimSpace(req.RuleID), strings.TrimSpace(req.Name), strings.TrimSpace(req.Description), enabled, priority, conditionsJSON, limitJSON, string(conflict), string(response), softLimit, retryAfter, req.EffectiveFrom, req.ExpiresAt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, pgx.ErrNoRows
	}
	return scanRateLimitRule(rows)
}

func (s *Store) ListRateLimitRules(ctx context.Context, tenantID string, query string, limit int, offset int) ([]platform.RateLimitRule, int, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}
	q := strings.TrimSpace(strings.ToLower(query))
	like := "%" + q + "%"
	var total int
	if err := s.pool.QueryRow(ctx, `
		SELECT COUNT(1)
		FROM rate_limit_rules
		WHERE tenant_id = $1::uuid
		  AND ($2 = '' OR LOWER(rule_id) LIKE $3 OR LOWER(COALESCE(name,'')) LIKE $3 OR LOWER(COALESCE(description,'')) LIKE $3)
	`, tenantID, q, like).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, tenant_id::text, rule_id, COALESCE(name,''), COALESCE(description,''), enabled, priority,
		       conditions_json, limit_json, conflict_behavior, response_behavior, soft_limit, retry_after_header,
		       effective_from, expires_at, created_at, updated_at
		FROM rate_limit_rules
		WHERE tenant_id = $1::uuid
		  AND ($2 = '' OR LOWER(rule_id) LIKE $3 OR LOWER(COALESCE(name,'')) LIKE $3 OR LOWER(COALESCE(description,'')) LIKE $3)
		ORDER BY priority ASC, updated_at DESC
		LIMIT $4 OFFSET $5
	`, tenantID, q, like, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]platform.RateLimitRule, 0, limit)
	for rows.Next() {
		item, scanErr := scanRateLimitRule(rows)
		if scanErr != nil {
			return nil, 0, scanErr
		}
		items = append(items, *item)
	}
	return items, total, rows.Err()
}

func (s *Store) GetRateLimitRuleByRuleID(ctx context.Context, tenantID string, ruleID string) (*platform.RateLimitRule, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, tenant_id::text, rule_id, COALESCE(name,''), COALESCE(description,''), enabled, priority,
		       conditions_json, limit_json, conflict_behavior, response_behavior, soft_limit, retry_after_header,
		       effective_from, expires_at, created_at, updated_at
		FROM rate_limit_rules
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
	return scanRateLimitRule(rows)
}

func (s *Store) UpdateRateLimitRuleByRuleID(ctx context.Context, tenantID string, ruleID string, req platform.UpdateRateLimitRuleRequest) (*platform.RateLimitRule, error) {
	current, err := s.GetRateLimitRuleByRuleID(ctx, tenantID, ruleID)
	if err != nil {
		return nil, err
	}
	name := current.Name
	if req.Name != nil {
		name = strings.TrimSpace(*req.Name)
	}
	description := current.Description
	if req.Description != nil {
		description = strings.TrimSpace(*req.Description)
	}
	enabled := current.Enabled
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	priority := current.Priority
	if req.Priority != nil {
		priority = *req.Priority
	}
	conditions := current.Conditions
	if req.Conditions != nil {
		conditions = req.Conditions
	}
	limit := current.Limit
	if req.Limit != nil {
		limit = req.Limit
	}
	conflict := current.ConflictBehavior
	if req.ConflictBehavior != nil {
		conflict = *req.ConflictBehavior
	}
	response := current.ResponseBehavior
	if req.ResponseBehavior != nil {
		response = *req.ResponseBehavior
	}
	softLimit := current.SoftLimit
	if req.SoftLimit != nil {
		softLimit = *req.SoftLimit
	}
	retryAfter := current.RetryAfterHeader
	if req.RetryAfterHeader != nil {
		retryAfter = *req.RetryAfterHeader
	}
	effectiveFrom := current.EffectiveFrom
	if req.EffectiveFrom != nil {
		effectiveFrom = req.EffectiveFrom
	}
	expiresAt := current.ExpiresAt
	if req.ExpiresAt != nil {
		expiresAt = req.ExpiresAt
	}
	conditionsJSON, err := marshalJSONObject(conditions)
	if err != nil {
		return nil, err
	}
	limitJSON, err := marshalJSONObject(limit)
	if err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx, `
		UPDATE rate_limit_rules
		SET name = NULLIF($3, ''),
		    description = NULLIF($4, ''),
		    enabled = $5,
		    priority = $6,
		    conditions_json = $7,
		    limit_json = $8,
		    conflict_behavior = $9,
		    response_behavior = $10,
		    soft_limit = $11,
		    retry_after_header = $12,
		    effective_from = $13,
		    expires_at = $14,
		    updated_at = NOW()
		WHERE tenant_id = $1::uuid AND rule_id = $2
		RETURNING id::text, tenant_id::text, rule_id, COALESCE(name,''), COALESCE(description,''), enabled, priority,
		          conditions_json, limit_json, conflict_behavior, response_behavior, soft_limit, retry_after_header,
		          effective_from, expires_at, created_at, updated_at
	`, tenantID, strings.TrimSpace(ruleID), name, description, enabled, priority, conditionsJSON, limitJSON, string(conflict), string(response), softLimit, retryAfter, effectiveFrom, expiresAt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, pgx.ErrNoRows
	}
	return scanRateLimitRule(rows)
}

func (s *Store) UpdateRateLimitRuleStatusByRuleID(ctx context.Context, tenantID string, ruleID string, enabled bool) (*platform.RateLimitRule, error) {
	rows, err := s.pool.Query(ctx, `
		UPDATE rate_limit_rules
		SET enabled = $3, updated_at = NOW()
		WHERE tenant_id = $1::uuid AND rule_id = $2
		RETURNING id::text, tenant_id::text, rule_id, COALESCE(name,''), COALESCE(description,''), enabled, priority,
		          conditions_json, limit_json, conflict_behavior, response_behavior, soft_limit, retry_after_header,
		          effective_from, expires_at, created_at, updated_at
	`, tenantID, strings.TrimSpace(ruleID), enabled)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, pgx.ErrNoRows
	}
	return scanRateLimitRule(rows)
}

func (s *Store) DeleteRateLimitRuleByRuleID(ctx context.Context, tenantID string, ruleID string) error {
	result, err := s.pool.Exec(ctx, `
		DELETE FROM rate_limit_rules
		WHERE tenant_id = $1::uuid AND rule_id = $2
	`, tenantID, strings.TrimSpace(ruleID))
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func scanRateLimitRule(rows pgx.Rows) (*platform.RateLimitRule, error) {
	var item platform.RateLimitRule
	var name string
	var description string
	var conditionsBytes []byte
	var limitBytes []byte
	var conflict string
	var response string
	if err := rows.Scan(
		&item.ID,
		&item.TenantID,
		&item.RuleID,
		&name,
		&description,
		&item.Enabled,
		&item.Priority,
		&conditionsBytes,
		&limitBytes,
		&conflict,
		&response,
		&item.SoftLimit,
		&item.RetryAfterHeader,
		&item.EffectiveFrom,
		&item.ExpiresAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return nil, err
	}
	item.Name = name
	item.Description = description
	item.Conditions = parseJSONObject(conditionsBytes)
	item.Limit = parseJSONObject(limitBytes)
	item.ConflictBehavior = platform.RateLimitConflictBehavior(conflict)
	item.ResponseBehavior = platform.RateLimitResponseBehavior(response)
	return &item, nil
}
