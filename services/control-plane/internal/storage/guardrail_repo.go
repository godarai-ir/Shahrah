package storage

import (
	"context"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *PolicyRepository) CreateTenantGuardrail(ctx context.Context, tenantID string, req platform.CreateTenantGuardrailRequest, actor string) (*platform.TenantGuardrail, error) {
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	configJSON, err := marshalJSONObject(req.Config)
	if err != nil {
		return nil, err
	}

	rows, err := s.pool.Query(ctx, `
		INSERT INTO tenant_guardrails (
			guardrail_id, tenant_id, slug, name, description, guardrail_type, operation, enforcement,
			enabled, config_json, created_by, updated_by, created_at, updated_at
		) VALUES (
			$1, $2::uuid, $3, $4, NULLIF($5, ''), $6, $7, $8, $9, $10, NULLIF($11, ''), NULLIF($11, ''), NOW(), NOW()
		)
		RETURNING guardrail_id::text, tenant_id::text, slug, name, COALESCE(description, ''), guardrail_type,
		          operation, enforcement, enabled, config_json, COALESCE(created_by, ''), COALESCE(updated_by, ''),
		          created_at, updated_at
	`, uuid.NewString(), tenantID, strings.TrimSpace(req.Slug), strings.TrimSpace(req.Name), strings.TrimSpace(req.Description), strings.TrimSpace(req.Type), string(req.Operation), string(req.Enforcement), enabled, configJSON, strings.TrimSpace(actor))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, pgx.ErrNoRows
	}

	item, err := scanTenantGuardrail(rows)
	if err == nil {
		s.invalidateTenantRuntimeProfile(ctx, tenantID, "guardrail.created")
	}

	return item, err
}

func (s *PolicyRepository) ListTenantGuardrails(ctx context.Context, tenantID string, query string, limit int, offset int) ([]platform.TenantGuardrail, int, error) {
	q, like, limit, offset := normalizePagedQuery(query, limit, offset)

	return listSearchResults(ctx, s.pool, `
		SELECT COUNT(1)
		FROM tenant_guardrails
		WHERE tenant_id = $1::uuid
		  AND ($2 = '' OR LOWER(slug) LIKE $3 OR LOWER(name) LIKE $3 OR LOWER(COALESCE(description, '')) LIKE $3)
	`, []any{tenantID, q, like}, `
		SELECT guardrail_id::text, tenant_id::text, slug, name, COALESCE(description, ''), guardrail_type,
		       operation, enforcement, enabled, config_json, COALESCE(created_by, ''), COALESCE(updated_by, ''),
		       created_at, updated_at
		FROM tenant_guardrails
		WHERE tenant_id = $1::uuid
		  AND ($2 = '' OR LOWER(slug) LIKE $3 OR LOWER(name) LIKE $3 OR LOWER(COALESCE(description, '')) LIKE $3)
		ORDER BY updated_at DESC
		LIMIT $4 OFFSET $5
	`, []any{tenantID, q, like, limit, offset}, limit, scanTenantGuardrail)
}

func (s *PolicyRepository) GetTenantGuardrailBySlug(ctx context.Context, tenantID string, slug string) (*platform.TenantGuardrail, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT guardrail_id::text, tenant_id::text, slug, name, COALESCE(description, ''), guardrail_type,
		       operation, enforcement, enabled, config_json, COALESCE(created_by, ''), COALESCE(updated_by, ''),
		       created_at, updated_at
		FROM tenant_guardrails
		WHERE tenant_id = $1::uuid AND slug = $2
		LIMIT 1
	`, tenantID, strings.TrimSpace(slug))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, pgx.ErrNoRows
	}

	item, err := scanTenantGuardrail(rows)
	if err == nil {
		s.invalidateTenantRuntimeProfile(ctx, tenantID, "guardrail.updated")
	}

	return item, err
}

func (s *PolicyRepository) UpdateTenantGuardrailBySlug(ctx context.Context, tenantID string, slug string, req platform.UpdateTenantGuardrailRequest, actor string) (*platform.TenantGuardrail, error) {
	current, err := s.GetTenantGuardrailBySlug(ctx, tenantID, slug)
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

	guardrailType := current.Type
	if req.Type != nil {
		guardrailType = strings.TrimSpace(*req.Type)
	}

	operation := current.Operation
	if req.Operation != nil {
		operation = *req.Operation
	}

	enforcement := current.Enforcement
	if req.Enforcement != nil {
		enforcement = *req.Enforcement
	}

	enabled := current.Enabled
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	config := current.Config
	if req.Config != nil {
		config = req.Config
	}

	configJSON, err := marshalJSONObject(config)
	if err != nil {
		return nil, err
	}

	rows, err := s.pool.Query(ctx, `
		UPDATE tenant_guardrails
		SET name = $3,
		    description = NULLIF($4, ''),
		    guardrail_type = $5,
		    operation = $6,
		    enforcement = $7,
		    enabled = $8,
		    config_json = $9,
		    updated_by = NULLIF($10, ''),
		    updated_at = NOW()
		WHERE tenant_id = $1::uuid AND slug = $2
		RETURNING guardrail_id::text, tenant_id::text, slug, name, COALESCE(description, ''), guardrail_type,
		          operation, enforcement, enabled, config_json, COALESCE(created_by, ''), COALESCE(updated_by, ''),
		          created_at, updated_at
	`, tenantID, strings.TrimSpace(slug), name, description, guardrailType, string(operation), string(enforcement), enabled, configJSON, strings.TrimSpace(actor))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, pgx.ErrNoRows
	}

	return scanTenantGuardrail(rows)
}

func (s *PolicyRepository) DeleteTenantGuardrailBySlug(ctx context.Context, tenantID string, slug string) error {
	result, err := s.pool.Exec(ctx, `
		DELETE FROM tenant_guardrails
		WHERE tenant_id = $1::uuid AND slug = $2
	`, tenantID, strings.TrimSpace(slug))
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	s.invalidateTenantRuntimeProfile(ctx, tenantID, "guardrail.deleted")

	return nil
}

func (s *PolicyRepository) ListEnabledTenantGuardrailsBySlugs(ctx context.Context, tenantID string, slugs []string) ([]platform.TenantGuardrail, error) {
	if len(slugs) == 0 {
		return nil, nil
	}

	clean := make([]string, 0, len(slugs))
	seen := map[string]bool{}

	for _, slug := range slugs {
		slug = strings.TrimSpace(slug)
		if slug == "" || seen[slug] {
			continue
		}

		seen[slug] = true
		clean = append(clean, slug)
	}

	rows, err := s.pool.Query(ctx, `
		SELECT guardrail_id::text, tenant_id::text, slug, name, COALESCE(description, ''), guardrail_type,
		       operation, enforcement, enabled, config_json, COALESCE(created_by, ''), COALESCE(updated_by, ''),
		       created_at, updated_at
		FROM tenant_guardrails
		WHERE tenant_id = $1::uuid AND enabled = true AND slug = ANY($2::text[])
	`, tenantID, clean)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []platform.TenantGuardrail{}

	for rows.Next() {
		item, scanErr := scanTenantGuardrail(rows)
		if scanErr != nil {
			return nil, scanErr
		}

		items = append(items, *item)
	}

	return items, rows.Err()
}

func scanTenantGuardrail(rows pgx.Rows) (*platform.TenantGuardrail, error) {
	var (
		item        platform.TenantGuardrail
		configBytes []byte
		operation   string
		enforcement string
	)
	if err := rows.Scan(
		&item.ID,
		&item.TenantID,
		&item.Slug,
		&item.Name,
		&item.Description,
		&item.Type,
		&operation,
		&enforcement,
		&item.Enabled,
		&configBytes,
		&item.CreatedBy,
		&item.UpdatedBy,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return nil, err
	}

	item.Operation = platform.GuardrailOperation(operation)
	item.Enforcement = platform.GuardrailEnforcement(enforcement)
	item.Config = parseJSONObject(configBytes)
	item.Selector = "guardrails/" + item.Slug

	return &item, nil
}
