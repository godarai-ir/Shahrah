package storage

import (
	"context"
	"database/sql"
	"time"
)

type ReconcileTenantCostRow struct {
	TenantID string
	CostUSD  float64
}

func (s *Store) ListSpendCostByTenantAndDay(ctx context.Context, day time.Time) ([]ReconcileTenantCostRow, error) {
	start := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)
	rows, err := s.pool.Query(ctx, `
		SELECT tenant_id::text, COALESCE(SUM(cost_usd), 0) AS cost_usd
		FROM spend_logs
		WHERE completion_end_time >= $1 AND completion_end_time < $2
		GROUP BY tenant_id
	`, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]ReconcileTenantCostRow, 0, 64)
	for rows.Next() {
		var r ReconcileTenantCostRow
		if err := rows.Scan(&r.TenantID, &r.CostUSD); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Store) ListMetricsCostByTenantAndDay(ctx context.Context, day time.Time) ([]ReconcileTenantCostRow, error) {
	if !s.clickhouseReady || s.clickhouse == nil {
		return nil, ErrMetricsStoreUnavailable
	}
	start := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)
	rows, err := s.clickhouse.QueryContext(ctx, `
		SELECT tenant_id, toFloat64(sum(cost_usd)) AS cost_usd
		FROM gateway_model_events
		WHERE ts >= ? AND ts < ?
		GROUP BY tenant_id
	`, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]ReconcileTenantCostRow, 0, 64)
	for rows.Next() {
		var r ReconcileTenantCostRow
		var cost sql.NullFloat64
		if err := rows.Scan(&r.TenantID, &cost); err != nil {
			return nil, err
		}
		if cost.Valid {
			r.CostUSD = cost.Float64
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
