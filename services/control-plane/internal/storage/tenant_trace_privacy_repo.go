package storage

import (
	"context"
	"errors"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/jackc/pgx/v5"
)

func (s *Store) GetTenantTracePrivacySettings(ctx context.Context, tenantID string) (*platform.TenantTracePrivacySettings, error) {
	var out platform.TenantTracePrivacySettings

	err := s.pool.QueryRow(ctx, `
		SELECT tenant_id::text, capture_policy, updated_at, COALESCE(updated_by_user_id::text, '')
		FROM tenant_trace_privacy_settings
		WHERE tenant_id = $1
		LIMIT 1
	`, tenantID).Scan(&out.TenantID, &out.CapturePolicy, &out.UpdatedAt, &out.UpdatedByUserID)
	if err != nil {
		return nil, err
	}

	return &out, nil
}

func (s *Store) UpsertTenantTracePrivacySettings(ctx context.Context, tenantID, capturePolicy, updatedByUserID string) (*platform.TenantTracePrivacySettings, error) {
	var out platform.TenantTracePrivacySettings

	err := s.pool.QueryRow(ctx, `
		INSERT INTO tenant_trace_privacy_settings (tenant_id, capture_policy, updated_by_user_id, updated_at)
		VALUES ($1, $2, NULLIF($3, '')::uuid, NOW())
		ON CONFLICT (tenant_id) DO UPDATE
		SET capture_policy = EXCLUDED.capture_policy,
		    updated_by_user_id = EXCLUDED.updated_by_user_id,
		    updated_at = NOW()
		RETURNING tenant_id::text, capture_policy, updated_at, COALESCE(updated_by_user_id::text, '')
	`, tenantID, capturePolicy, updatedByUserID).Scan(&out.TenantID, &out.CapturePolicy, &out.UpdatedAt, &out.UpdatedByUserID)
	if err != nil {
		return nil, err
	}

	return &out, nil
}

func (s *Store) ResolveTenantTraceCapturePolicy(ctx context.Context, tenantID string) (string, bool, error) {
	var (
		policy    string
		updatedAt time.Time
	)

	err := s.pool.QueryRow(ctx, `
		SELECT capture_policy, updated_at
		FROM tenant_trace_privacy_settings
		WHERE tenant_id = $1
		LIMIT 1
	`, tenantID).Scan(&policy, &updatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", false, nil
		}

		return "", false, err
	}

	_ = updatedAt

	return policy, true, nil
}
