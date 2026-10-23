package storage

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestClickHouseSpendPurgeIntegration(t *testing.T) {
	if os.Getenv("SHAHRAH_CLICKHOUSE_INTEGRATION") != "1" {
		t.Skip("set SHAHRAH_CLICKHOUSE_INTEGRATION=1 to run ClickHouse container integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	dsn, cleanup := startClickHouseTestContainer(t, ctx)
	defer cleanup()

	db, err := sql.Open("clickhouse", dsn)
	require.NoError(t, err)

	defer closeBestEffort(db)

	require.NoError(t, waitForClickHouse(ctx, db))
	require.NoError(t, applyClickHouseMigrations(ctx, db))

	store := composeObservabilityRepository(&observabilityState{
		clickhouse:             db,
		clickhouseReady:        true,
		clickhouseMetricsReady: true,
		clickhouseLogsReady:    true,
		logger:                 zap.NewNop(),
	})
	require.NoError(t, store.ensureClickHouseMetricsSchema(ctx))

	start := time.Now().UTC().Add(-2 * time.Hour).Truncate(time.Minute)
	cutoff := start.Add(time.Minute)

	insertClickHouseMetrics(t, ctx, store, "tenant-before", start)
	insertClickHouseLogs(t, ctx, db, "tenant-before", start)
	insertClickHouseMetrics(t, ctx, store, "tenant-after", cutoff.Add(time.Minute))
	insertClickHouseLogs(t, ctx, db, "tenant-after", cutoff.Add(time.Minute))

	purges := &SpendHistoryPurgeRepository{clickhouse: db, clickhouseConfigured: true}
	require.NoError(t, purges.purgeClickHouseSpend(ctx, cutoff))
	waitForClickHouseMutations(t, ctx, db)

	for _, check := range []struct {
		name  string
		query string
	}{
		{"events", `SELECT count(), toFloat64(sum(cost_usd)) FROM gateway_model_events WHERE tenant_id = ?`},
		{"aggregates", `SELECT toUInt64(countMerge(request_count)), toFloat64(sumMerge(cost_usd_sum)) FROM gateway_model_metrics_1m WHERE tenant_id = ?`},
		{"logs", `SELECT count(), toFloat64(sum(cost_usd)) FROM request_logs_ch WHERE tenant_id = ?`},
	} {
		var (
			before, after         uint64
			beforeCost, afterCost float64
		)

		require.NoError(t, db.QueryRowContext(ctx, check.query, "tenant-before").Scan(&before, &beforeCost), check.name)
		require.NoError(t, db.QueryRowContext(ctx, check.query, "tenant-after").Scan(&after, &afterCost), check.name)

		require.NotZero(t, before, "%s: rows before the cutoff are kept", check.name)
		require.Zero(t, beforeCost, "%s: spend before the cutoff is cleared", check.name)
		require.Equal(t, before, after, check.name)
		require.Positive(t, afterCost, "%s: spend after the cutoff is kept", check.name)
	}
}

func waitForClickHouseMutations(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()

	require.Eventually(t, func() bool {
		var running uint64
		if err := db.QueryRowContext(ctx, `
			SELECT count() FROM system.mutations
			WHERE database = currentDatabase() AND is_done = 0
		`).Scan(&running); err != nil {
			return false
		}

		return running == 0
	}, time.Minute, 200*time.Millisecond)
}
