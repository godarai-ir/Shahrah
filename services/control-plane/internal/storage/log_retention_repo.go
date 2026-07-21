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
	values, err := s.listRetentionPolicyValues(ctx, platform.FeatureLogsRetentionDays, logretention.ParseLogRetentionDays, "log")
	if err != nil {
		return nil, err
	}

	out := make([]logretention.LogRetentionPolicy, 0, len(values))
	for _, value := range values {
		out = append(out, logretention.LogRetentionPolicy{TenantID: value.tenantID, RetentionDays: value.days})
	}

	return out, nil
}

type retentionPolicyValue struct {
	tenantID string
	days     *int
}

func (s *Store) listRetentionPolicyValues(
	ctx context.Context,
	featureKey string,
	parse func(interface{}) (*int, bool),
	policyKind string,
) ([]retentionPolicyValue, error) {
	rows, err := s.pool.Query(ctx, traceRetentionPoliciesSQL, featureKey)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []retentionPolicyValue{}

	for rows.Next() {
		var (
			tenantID string
			raw      []byte
		)
		if err := rows.Scan(&tenantID, &raw); err != nil {
			return nil, err
		}

		days, ok := parse(parseJSONValue(raw))
		if !ok {
			s.logger.Warn(
				"invalid retention entitlement ignored",
				zap.String("policy_kind", policyKind),
				zap.String("tenant_id", tenantID),
				zap.String("feature_key", featureKey),
				zap.ByteString("value_json", raw),
			)

			continue
		}

		out = append(out, retentionPolicyValue{tenantID: tenantID, days: days})
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
