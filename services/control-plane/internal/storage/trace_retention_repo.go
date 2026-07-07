package storage

import (
	"context"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/observability/tracing"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

const maxClickHouseRetentionDays = 65535

const traceRetentionPoliciesSQL = `
	WITH tenant_scope AS (
		SELECT tenant_id
		FROM tenants
	),
	current_subscriptions AS (
		SELECT DISTINCT ON (ts.tenant_id)
			ts.tenant_id,
			ts.version_id
		FROM tenant_subscriptions ts
		WHERE ts.status IN ('trialing', 'active', 'past_due')
		ORDER BY ts.tenant_id, ts.updated_at DESC
	),
	base_values AS (
		SELECT t.tenant_id, fc.default_value AS value_json, 3 AS precedence
		FROM tenant_scope t
		CROSS JOIN feature_catalog fc
		WHERE fc.feature_key = $1
	),
	plan_values AS (
		SELECT t.tenant_id, pe.value_json, 2 AS precedence
		FROM tenant_scope t
		JOIN current_subscriptions cs
		  ON cs.tenant_id = t.tenant_id
		JOIN plan_entitlements pe
		  ON pe.version_id = cs.version_id
		 AND pe.feature_key = $1
	),
	override_values AS (
		SELECT t.tenant_id, o.value_json, 1 AS precedence
		FROM tenant_scope t
		JOIN tenant_entitlement_overrides o
		  ON o.tenant_id = t.tenant_id
		 AND o.feature_key = $1
		WHERE o.active = true
		  AND (o.expires_at IS NULL OR o.expires_at > NOW())
	),
	resolved AS (
		SELECT DISTINCT ON (tenant_id) tenant_id, value_json
		FROM (
			SELECT tenant_id, value_json, precedence FROM override_values
			UNION ALL
			SELECT tenant_id, value_json, precedence FROM plan_values
			UNION ALL
			SELECT tenant_id, value_json, precedence FROM base_values
		) x
		ORDER BY tenant_id, precedence
	)
	SELECT tenant_id::text, value_json
	FROM resolved
	ORDER BY tenant_id
`

const traceRetentionPolicySQL = `
	WITH tenant_scope AS (
		SELECT tenant_id
		FROM tenants
		WHERE tenant_id = $2::uuid
	),
	current_subscriptions AS (
		SELECT DISTINCT ON (ts.tenant_id)
			ts.tenant_id,
			ts.version_id
		FROM tenant_subscriptions ts
		JOIN tenant_scope t
		  ON t.tenant_id = ts.tenant_id
		WHERE ts.status IN ('trialing', 'active', 'past_due')
		ORDER BY ts.tenant_id, ts.updated_at DESC
	),
	base_values AS (
		SELECT t.tenant_id, fc.default_value AS value_json, 3 AS precedence
		FROM tenant_scope t
		CROSS JOIN feature_catalog fc
		WHERE fc.feature_key = $1
	),
	plan_values AS (
		SELECT t.tenant_id, pe.value_json, 2 AS precedence
		FROM tenant_scope t
		JOIN current_subscriptions cs
		  ON cs.tenant_id = t.tenant_id
		JOIN plan_entitlements pe
		  ON pe.version_id = cs.version_id
		 AND pe.feature_key = $1
	),
	override_values AS (
		SELECT t.tenant_id, o.value_json, 1 AS precedence
		FROM tenant_scope t
		JOIN tenant_entitlement_overrides o
		  ON o.tenant_id = t.tenant_id
		 AND o.feature_key = $1
		WHERE o.active = true
		  AND (o.expires_at IS NULL OR o.expires_at > NOW())
	),
	resolved AS (
		SELECT DISTINCT ON (tenant_id) tenant_id, value_json
		FROM (
			SELECT tenant_id, value_json, precedence FROM override_values
			UNION ALL
			SELECT tenant_id, value_json, precedence FROM plan_values
			UNION ALL
			SELECT tenant_id, value_json, precedence FROM base_values
		) x
		ORDER BY tenant_id, precedence
	)
	SELECT value_json
	FROM resolved
`

// ListTraceRetentionPolicies resolves the current effective trace retention for every subscribed tenant.
func (s *Store) ListTraceRetentionPolicies(ctx context.Context) ([]tracing.TraceRetentionPolicy, error) {
	rows, err := s.pool.Query(ctx, traceRetentionPoliciesSQL, platform.FeatureTracesRetentionDays)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []tracing.TraceRetentionPolicy{}

	for rows.Next() {
		var (
			tenantID string
			raw      []byte
		)
		if err := rows.Scan(&tenantID, &raw); err != nil {
			return nil, err
		}

		days, ok := tracing.ParseTraceRetentionDays(parseJSONValue(raw))
		if !ok {
			s.logger.Warn(
				"invalid trace retention entitlement ignored",
				zap.String("tenant_id", tenantID),
				zap.String("feature_key", platform.FeatureTracesRetentionDays),
				zap.ByteString("value_json", raw),
			)

			continue
		}

		out = append(out, tracing.TraceRetentionPolicy{
			TenantID:      tenantID,
			RetentionDays: days,
		})
	}

	return out, rows.Err()
}

// ResolveTraceRetentionDays resolves the current effective trace TTL value for a tenant.
// A nil return means unlimited retention and is stored as retention_days = 0 in ClickHouse.
func (s *Store) ResolveTraceRetentionDays(ctx context.Context, tenantID string) (*int, error) {
	var raw []byte
	if err := s.pool.QueryRow(ctx, traceRetentionPolicySQL, platform.FeatureTracesRetentionDays, tenantID).Scan(&raw); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	days, ok := tracing.ParseTraceRetentionDays(parseJSONValue(raw))
	if !ok {
		logger := s.logger
		if logger == nil {
			logger = zap.NewNop()
		}
		logger.Warn(
			"invalid trace retention entitlement ignored",
			zap.String("tenant_id", tenantID),
			zap.String("feature_key", platform.FeatureTracesRetentionDays),
			zap.ByteString("value_json", raw),
		)

		return nil, nil
	}

	return days, nil
}
