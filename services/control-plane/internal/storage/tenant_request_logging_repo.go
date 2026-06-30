package storage

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/jackc/pgx/v5"
)

const (
	RequestLoggingModeNone            = "none"
	RequestLoggingModeMetadataOnly    = "metadata-only"
	RequestLoggingModeRedactedContent = "redacted-content"
	RequestLoggingModeFullContent     = "full-content"
)

func NormalizeRequestLoggingMode(raw string) string {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case RequestLoggingModeNone:
		return RequestLoggingModeNone
	case RequestLoggingModeRedactedContent:
		return RequestLoggingModeRedactedContent
	case RequestLoggingModeFullContent:
		return RequestLoggingModeFullContent
	default:
		return RequestLoggingModeMetadataOnly
	}
}

func (s *Store) GetTenantRequestLoggingSettings(ctx context.Context, tenantID string) (*platform.TenantRequestLoggingSettings, error) {
	var out platform.TenantRequestLoggingSettings

	err := s.pool.QueryRow(ctx, `
		SELECT tenant_id::text, logging_mode, updated_at, COALESCE(updated_by_user_id::text, '')
		FROM tenant_request_logging_settings
		WHERE tenant_id = $1
		LIMIT 1
	`, tenantID).Scan(&out.TenantID, &out.LoggingMode, &out.UpdatedAt, &out.UpdatedByUserID)
	if err != nil {
		return nil, err
	}

	out.LoggingMode = NormalizeRequestLoggingMode(out.LoggingMode)

	return &out, nil
}

func (s *Store) UpsertTenantRequestLoggingSettings(ctx context.Context, tenantID, loggingMode, updatedByUserID string) (*platform.TenantRequestLoggingSettings, error) {
	var out platform.TenantRequestLoggingSettings

	err := s.pool.QueryRow(ctx, `
		INSERT INTO tenant_request_logging_settings (tenant_id, logging_mode, updated_by_user_id, updated_at)
		VALUES ($1, $2, NULLIF($3, '')::uuid, NOW())
		ON CONFLICT (tenant_id) DO UPDATE
		SET logging_mode = EXCLUDED.logging_mode,
		    updated_by_user_id = EXCLUDED.updated_by_user_id,
		    updated_at = NOW()
		RETURNING tenant_id::text, logging_mode, updated_at, COALESCE(updated_by_user_id::text, '')
	`, tenantID, NormalizeRequestLoggingMode(loggingMode), updatedByUserID).Scan(&out.TenantID, &out.LoggingMode, &out.UpdatedAt, &out.UpdatedByUserID)
	if err != nil {
		return nil, err
	}

	out.LoggingMode = NormalizeRequestLoggingMode(out.LoggingMode)

	return &out, nil
}

func (s *Store) ResolveTenantRequestLoggingMode(ctx context.Context, tenantID string) (string, bool, error) {
	var (
		mode      string
		updatedAt time.Time
	)

	err := s.pool.QueryRow(ctx, `
		SELECT logging_mode, updated_at
		FROM tenant_request_logging_settings
		WHERE tenant_id = $1
		LIMIT 1
	`, tenantID).Scan(&mode, &updatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", false, nil
		}

		return "", false, err
	}

	_ = updatedAt

	return NormalizeRequestLoggingMode(mode), true, nil
}
