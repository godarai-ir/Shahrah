package storage

import (
	"context"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/observability/logretention"
	"go.uber.org/zap"
)

// ListLogRetentionPolicies resolves the current effective request log retention for every tenant.
func (s *Store) ListLogRetentionPolicies(ctx context.Context) ([]logretention.LogRetentionPolicy, error) {
	rows, err := s.pool.Query(ctx, traceRetentionPoliciesSQL, platform.FeatureLogsRetentionDays)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []logretention.LogRetentionPolicy{}

	for rows.Next() {
		var (
			tenantID string
			raw      []byte
		)
		if err := rows.Scan(&tenantID, &raw); err != nil {
			return nil, err
		}

		days, ok := logretention.ParseLogRetentionDays(parseJSONValue(raw))
		if !ok {
			s.logger.Warn(
				"invalid log retention entitlement ignored",
				zap.String("tenant_id", tenantID),
				zap.String("feature_key", platform.FeatureLogsRetentionDays),
				zap.ByteString("value_json", raw),
			)

			continue
		}

		out = append(out, logretention.LogRetentionPolicy{
			TenantID:      tenantID,
			RetentionDays: days,
		})
	}

	return out, rows.Err()
}

// DeleteLogDataBefore deletes tenant-scoped request log rows older than cutoff from ClickHouse.
func (s *Store) DeleteLogDataBefore(ctx context.Context, tenantID string, cutoff time.Time) error {
	if !s.clickHouseLogsReady() {
		return ErrLogStoreUnavailable
	}

	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" || cutoff.IsZero() {
		return nil
	}

	cutoff = cutoff.UTC()
	if _, err := s.clickhouse.ExecContext(ctx, `
		ALTER TABLE request_logs_ch
		DELETE WHERE tenant_id = ? AND started_at < ?
	`, tenantID, cutoff); err != nil {
		return err
	}

	return nil
}
