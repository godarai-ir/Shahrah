package storage

import (
	"context"
	"errors"
	"fmt"
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

func (s *ObservabilityRepository) GetTenantRequestLoggingSettings(ctx context.Context, tenantID string) (*platform.TenantRequestLoggingSettings, error) {
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

func (s *ObservabilityRepository) UpsertTenantRequestLoggingSettings(ctx context.Context, tenantID, loggingMode, updatedByUserID string) (*platform.TenantRequestLoggingSettings, error) {
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

	s.invalidateTenantRequestLoggingMode(ctx, tenantID)

	return &out, nil
}

func (s *ObservabilityRepository) ResolveTenantRequestLoggingMode(ctx context.Context, tenantID string) (string, bool, error) {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return "", false, nil
	}

	if mode, ok, found := s.getCachedTenantRequestLoggingMode(ctx, tenantID); found {
		return mode, ok, nil
	}

	value, err, _ := s.requestLoggingModeGroup.Do(requestLoggingModeCacheKey(tenantID), func() (interface{}, error) {
		if mode, ok, found := s.getCachedTenantRequestLoggingMode(ctx, tenantID); found {
			return cachedRequestLoggingMode{Mode: mode, OK: ok}, nil
		}

		mode, ok, resolveErr := s.resolveTenantRequestLoggingModeFromDB(ctx, tenantID)
		if resolveErr != nil {
			return nil, resolveErr
		}

		s.cacheTenantRequestLoggingMode(ctx, tenantID, mode, ok)

		return cachedRequestLoggingMode{Mode: mode, OK: ok}, nil
	})
	if err != nil {
		return "", false, err
	}

	cached, ok := value.(cachedRequestLoggingMode)
	if !ok {
		return "", false, fmt.Errorf("unexpected request logging mode cache value for tenant %s", tenantID)
	}

	return cached.Mode, cached.OK, nil
}

type cachedRequestLoggingMode struct {
	Mode string
	OK   bool
}

func (s *ObservabilityRepository) resolveTenantRequestLoggingModeFromDB(ctx context.Context, tenantID string) (string, bool, error) {
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

func (s *ObservabilityRepository) getCachedTenantRequestLoggingMode(ctx context.Context, tenantID string) (string, bool, bool) {
	if s.entitlementCache == nil {
		return "", false, false
	}

	value, found := s.entitlementCache.Get(ctx, requestLoggingModeCacheKey(tenantID))
	if !found {
		return "", false, false
	}

	raw, ok := value.(string)
	if !ok {
		return "", false, false
	}

	if raw == "__unset__" {
		return "", false, true
	}

	return NormalizeRequestLoggingMode(raw), true, true
}

func (s *ObservabilityRepository) cacheTenantRequestLoggingMode(ctx context.Context, tenantID string, mode string, ok bool) {
	if s.entitlementCache == nil {
		return
	}

	value := "__unset__"
	if ok {
		value = NormalizeRequestLoggingMode(mode)
	}

	s.entitlementCache.Set(ctx, requestLoggingModeCacheKey(tenantID), value, s.requestLoggingModeTTL)
}

func (s *ObservabilityRepository) invalidateTenantRequestLoggingMode(ctx context.Context, tenantID string) {
	if s.entitlementCache != nil && strings.TrimSpace(tenantID) != "" {
		s.entitlementCache.Delete(ctx, requestLoggingModeCacheKey(tenantID))
	}
}

func requestLoggingModeCacheKey(tenantID string) string {
	return "request-logging-mode:" + strings.TrimSpace(tenantID)
}
