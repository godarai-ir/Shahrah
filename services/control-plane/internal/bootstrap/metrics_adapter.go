package bootstrap

import (
	"context"
	"time"

	metricsobs "github.com/Delpak/shahrah/services/control-plane/internal/observability/metrics"
	"github.com/Delpak/shahrah/services/control-plane/internal/storage"
)

type costReconcileAdapter struct {
	store *storage.Store
}

func (a *costReconcileAdapter) ListSpendCostByTenantAndDay(
	ctx context.Context,
	day time.Time,
) ([]metricsobs.ReconcileTenantCostRow, error) {
	rows, err := a.store.ListSpendCostByTenantAndDay(ctx, day)
	if err != nil {
		return nil, err
	}

	out := make([]metricsobs.ReconcileTenantCostRow, 0, len(rows))
	for _, row := range rows {
		out = append(out, metricsobs.ReconcileTenantCostRow{TenantID: row.TenantID, CostUSD: row.CostUSD})
	}

	return out, nil
}

func (a *costReconcileAdapter) ListMetricsCostByTenantAndDay(
	ctx context.Context,
	day time.Time,
) ([]metricsobs.ReconcileTenantCostRow, error) {
	rows, err := a.store.ListMetricsCostByTenantAndDay(ctx, day)
	if err != nil {
		return nil, err
	}

	out := make([]metricsobs.ReconcileTenantCostRow, 0, len(rows))
	for _, row := range rows {
		out = append(out, metricsobs.ReconcileTenantCostRow{TenantID: row.TenantID, CostUSD: row.CostUSD})
	}

	return out, nil
}
