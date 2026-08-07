package storage

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/apperrors"
	"github.com/Delpak/shahrah/services/control-plane/internal/runtimeprofile"
	"github.com/jackc/pgx/v5"
)

func (s *TenantRepository) GetTenantByID(ctx context.Context, tenantID string) (*platform.Tenant, error) {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return nil, apperrors.ErrNotFound
	}

	var tenant platform.Tenant

	err := s.pool.QueryRow(ctx, `
		SELECT tenant_id::text, slug, domain, status, created_at
		FROM tenants
		WHERE tenant_id = $1::uuid
		LIMIT 1
	`, tenantID).Scan(&tenant.ID, &tenant.Slug, &tenant.Domain, &tenant.Status, &tenant.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &tenant, nil
}

func (s *RuntimeProfileRepository) ListRuntimeAuthContexts(ctx context.Context, tenantID string) ([]platform.AuthContext, error) {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return nil, nil
	}

	rows, err := s.pool.Query(ctx, `
		SELECT vt.token_hash
		FROM verification_tokens vt
		LEFT JOIN virtual_users vu
		  ON vu.virtual_user_id = vt.virtual_user_id
		 AND vu.tenant_id = vt.tenant_id
		WHERE vt.tenant_id = $1::uuid
		  AND vt.revoked_at IS NULL
		  AND (vt.expires IS NULL OR vt.expires > NOW())
		  AND COALESCE(vt.blocked, false) = false
		  AND (vt.virtual_user_id IS NULL OR COALESCE(vu.active, false) = true)
		  AND COALESCE(vt.token_hash, '') <> ''
		ORDER BY vt.created_at ASC
	`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tokenHashes := make([]string, 0)

	for rows.Next() {
		var tokenHash string
		if err := rows.Scan(&tokenHash); err != nil {
			return nil, err
		}

		tokenHash = strings.TrimSpace(tokenHash)
		if tokenHash != "" {
			tokenHashes = append(tokenHashes, tokenHash)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	out := make([]platform.AuthContext, 0, len(tokenHashes))
	for _, tokenHash := range tokenHashes {
		authContext, err := s.auth.GetKeyByHash(ctx, tokenHash)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				continue
			}

			return nil, err
		}

		if authContext == nil || strings.TrimSpace(authContext.TokenHash) == "" {
			continue
		}

		out = append(out, *authContext)
	}

	return out, nil
}

func (s *PolicyRepository) ListEnabledTenantGuardrails(ctx context.Context, tenantID string) ([]platform.TenantGuardrail, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT guardrail_id::text, tenant_id::text, slug, name, COALESCE(description, ''), guardrail_type,
		       operation, enforcement, enabled, config_json, COALESCE(created_by, ''), COALESCE(updated_by, ''),
		       created_at, updated_at
		FROM tenant_guardrails
		WHERE tenant_id = $1::uuid AND enabled = true
		ORDER BY slug ASC
	`, strings.TrimSpace(tenantID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]platform.TenantGuardrail, 0)

	for rows.Next() {
		item, err := scanTenantGuardrail(rows)
		if err != nil {
			return nil, err
		}

		items = append(items, *item)
	}

	return items, rows.Err()
}

func (s *RuntimeProfileRepository) NextTenantRuntimeProfileVersion(ctx context.Context, tenantID string) (int64, error) {
	var version int64

	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(MAX(version), 0) + 1
		FROM tenant_runtime_profiles
		WHERE tenant_id = $1::uuid
	`, strings.TrimSpace(tenantID)).Scan(&version)

	return version, err
}

func (s *RuntimeProfileRepository) SaveTenantRuntimeProfile(ctx context.Context, stored runtimeprofile.StoredProfile) error {
	if stored.CompiledAt.IsZero() {
		stored.CompiledAt = time.Now().UTC()
	}

	if stored.SourceUpdatedAt.IsZero() {
		stored.SourceUpdatedAt = stored.CompiledAt
	}

	if strings.TrimSpace(stored.Status) == "" {
		stored.Status = runtimeprofile.ProfileStatusActive
	}

	_, err := s.pool.Exec(ctx, `
		INSERT INTO tenant_runtime_profiles (
			tenant_id, version, profile_json, checksum, compiled_at, source_updated_at, status, error
		) VALUES (
			$1::uuid, $2, $3::jsonb, $4, $5, $6, $7, NULLIF($8, '')
		)
		ON CONFLICT (tenant_id, version) DO UPDATE
		SET profile_json = EXCLUDED.profile_json,
		    checksum = EXCLUDED.checksum,
		    compiled_at = EXCLUDED.compiled_at,
		    source_updated_at = EXCLUDED.source_updated_at,
		    status = EXCLUDED.status,
		    error = EXCLUDED.error
	`, strings.TrimSpace(stored.TenantID), stored.Version, stored.ProfileJSON, stored.Checksum, stored.CompiledAt, stored.SourceUpdatedAt, stored.Status, stored.Error)

	return err
}

func (s *RuntimeProfileRepository) LoadLatestTenantRuntimeProfile(ctx context.Context, tenantID string) (*runtimeprofile.StoredProfile, error) {
	var stored runtimeprofile.StoredProfile

	err := s.pool.QueryRow(ctx, `
		SELECT tenant_id::text, version, profile_json, checksum, compiled_at, source_updated_at, status, COALESCE(error, '')
		FROM tenant_runtime_profiles
		WHERE tenant_id = $1::uuid AND status = 'active'
		ORDER BY version DESC
		LIMIT 1
	`, strings.TrimSpace(tenantID)).Scan(
		&stored.TenantID,
		&stored.Version,
		&stored.ProfileJSON,
		&stored.Checksum,
		&stored.CompiledAt,
		&stored.SourceUpdatedAt,
		&stored.Status,
		&stored.Error,
	)
	if err != nil {
		return nil, err
	}

	return &stored, nil
}
