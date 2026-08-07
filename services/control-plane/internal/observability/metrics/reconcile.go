// Package metrics includes cost reconciliation helpers for model usage accounting.
package metrics

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"time"

	"go.uber.org/zap"
)

// CostReconcileStore exposes the data needed for cost reconciliation.
type CostReconcileStore interface {
	ListSpendCostByTenantAndDay(ctx context.Context, day time.Time) ([]ReconcileTenantCostRow, error)
	ListMetricsCostByTenantAndDay(ctx context.Context, day time.Time) ([]ReconcileTenantCostRow, error)
}

// ReconcileTenantCostRow is a per-tenant daily cost row.
type ReconcileTenantCostRow struct {
	TenantID string
	CostUSD  float64
}

// ReconcileConfig controls the reconciliation worker.
type ReconcileConfig struct {
	Enabled              bool
	Interval             time.Duration
	ThresholdUSD         float64
	ReportPath           string
	RunForPreviousDayUTC bool
}

// DefaultReconcileConfig returns reconciliation defaults without reading process-global state.
func DefaultReconcileConfig() ReconcileConfig {
	return ReconcileConfig{
		Enabled:              false,
		Interval:             1 * time.Hour,
		ThresholdUSD:         0.01,
		ReportPath:           "./tmp/model-metrics-reconcile.json",
		RunForPreviousDayUTC: true,
	}
}

// CostReconciler compares recorded spend against model metrics.
type CostReconciler struct {
	cfg    ReconcileConfig
	store  CostReconcileStore
	logger *zap.Logger
	stopCh chan struct{}
}

// NewCostReconciler creates a reconciliation worker.
func NewCostReconciler(cfg ReconcileConfig, store CostReconcileStore, logger *zap.Logger) *CostReconciler {
	return &CostReconciler{
		cfg:    cfg,
		store:  store,
		logger: logger,
		stopCh: make(chan struct{}),
	}
}

// Start launches the reconciliation loop if enabled.
func (r *CostReconciler) Start() {
	if !r.cfg.Enabled {
		return
	}
	go r.loop()
}

// Stop stops the reconciliation loop.
func (r *CostReconciler) Stop() { close(r.stopCh) }

func (r *CostReconciler) loop() {
	ticker := time.NewTicker(r.cfg.Interval)
	defer ticker.Stop()

	if err := r.runOnce(context.Background()); err != nil {
		r.logger.Warn("initial metrics reconcile failed", zap.Error(err))
	}

	for {
		select {
		case <-r.stopCh:
			return
		case <-ticker.C:
			if err := r.runOnce(context.Background()); err != nil {
				r.logger.Warn("scheduled metrics reconcile failed", zap.Error(err))
			}
		}
	}
}

func (r *CostReconciler) runOnce(ctx context.Context) error {
	day := time.Now().UTC()
	if r.cfg.RunForPreviousDayUTC {
		day = day.Add(-24 * time.Hour)
	}

	spendRows, err := r.store.ListSpendCostByTenantAndDay(ctx, day)
	if err != nil {
		r.logger.Warn("reconcile spend query failed", zap.Error(err))
		return err
	}

	metricRows, err := r.store.ListMetricsCostByTenantAndDay(ctx, day)
	if err != nil {
		r.logger.Warn("reconcile metrics query failed", zap.Error(err))
		return err
	}

	spendByTenant := make(map[string]float64, len(spendRows))
	metricsByTenant := make(map[string]float64, len(metricRows))
	tenantSet := map[string]struct{}{}

	for _, r := range spendRows {
		spendByTenant[r.TenantID] = r.CostUSD
		tenantSet[r.TenantID] = struct{}{}
	}

	for _, r := range metricRows {
		metricsByTenant[r.TenantID] = r.CostUSD
		tenantSet[r.TenantID] = struct{}{}
	}

	type diff struct {
		TenantID        string  `json:"tenant_id"`
		SpendCostUSD    float64 `json:"spend_cost_usd"`
		MetricsCostUSD  float64 `json:"metrics_cost_usd"`
		DiffUSD         float64 `json:"diff_usd"`
		WithinThreshold bool    `json:"within_threshold"`
	}

	out := struct {
		GeneratedAt  string  `json:"generated_at"`
		DayUTC       string  `json:"day_utc"`
		ThresholdUSD float64 `json:"threshold_usd"`
		Rows         []diff  `json:"rows"`
	}{
		GeneratedAt:  time.Now().UTC().Format(time.RFC3339),
		DayUTC:       time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.UTC).Format("2006-01-02"),
		ThresholdUSD: r.cfg.ThresholdUSD,
		Rows:         make([]diff, 0, len(tenantSet)),
	}
	for tenantID := range tenantSet {
		spend := spendByTenant[tenantID]
		metric := metricsByTenant[tenantID]

		delta := spend - metric
		if delta < 0 {
			delta = -delta
		}

		out.Rows = append(out.Rows, diff{
			TenantID: tenantID, SpendCostUSD: spend, MetricsCostUSD: metric, DiffUSD: delta, WithinThreshold: delta <= r.cfg.ThresholdUSD,
		})
	}

	sort.Slice(out.Rows, func(i, j int) bool { return out.Rows[i].DiffUSD > out.Rows[j].DiffUSD })

	if err := os.MkdirAll(filepath.Dir(r.cfg.ReportPath), 0o755); err != nil {
		return err
	}

	b, marshalErr := json.MarshalIndent(out, "", "  ")
	if marshalErr != nil {
		return marshalErr
	}

	return os.WriteFile(r.cfg.ReportPath, b, 0o644)
}
