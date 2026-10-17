package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var errSpendPurgeClickHouseUnavailable = errors.New("clickhouse is configured but unavailable")

// spendPurgeTarget names the cost columns of one ClickHouse table and the value
// that clears each of them.
type spendPurgeTarget struct {
	table      string
	timeColumn string
	columns    []spendPurgeColumn
}

type spendPurgeColumn struct {
	name string
	zero string
}

// spendPurgeTargets are the ClickHouse tables that carry spend. Rows are kept,
// so request counts, tokens, and latency survive the purge.
var spendPurgeTargets = []spendPurgeTarget{
	{
		table:      "gateway_model_events",
		timeColumn: "ts",
		columns: []spendPurgeColumn{
			{name: "cost_usd", zero: "0"},
			{name: "avoided_cost_usd", zero: "0"},
			{name: "avoided_cost_irr", zero: "0"},
			{name: "embedding_cost_usd", zero: "0"},
			{name: "credits_mc", zero: "0"},
			{name: "credits_avoided_mc", zero: "0"},
		},
	},
	{
		table:      "gateway_model_metrics_1m",
		timeColumn: "bucket_ts",
		columns: []spendPurgeColumn{
			{name: "cost_usd_sum", zero: "initializeAggregation('sumState', toDecimal64(0, 8))"},
		},
	},
	{
		table:      "request_logs_ch",
		timeColumn: "started_at",
		columns: []spendPurgeColumn{
			{name: "cost_usd", zero: "0"},
			{name: "cost_irr", zero: "0"},
		},
	},
}

// SpendHistoryPurgeRepository runs the one-time ClickHouse spend purges that
// Postgres migrations request in clickhouse_spend_purges.
type SpendHistoryPurgeRepository struct {
	pool                 *pgxpool.Pool
	clickhouse           *sql.DB
	clickhouseConfigured bool
}

func newSpendHistoryPurgeRepository(pool *pgxpool.Pool, clickhouse *sql.DB, clickhouseConfigured bool) *SpendHistoryPurgeRepository {
	return &SpendHistoryPurgeRepository{pool: pool, clickhouse: clickhouse, clickhouseConfigured: clickhouseConfigured}
}

type pendingSpendPurge struct {
	id     int64
	cutoff time.Time
}

// RunPendingSpendPurges clears the ClickHouse spend written before each pending
// purge's cutoff and marks it complete. It returns how many purges are still
// pending. A deployment without ClickHouse has nothing to purge, so its purges
// complete at once. Clearing is idempotent, so two replicas racing is harmless.
func (r *SpendHistoryPurgeRepository) RunPendingSpendPurges(ctx context.Context) (int, error) {
	pending, err := r.listPendingSpendPurges(ctx)
	if err != nil || len(pending) == 0 {
		return len(pending), err
	}

	if r.clickhouseConfigured && r.clickhouse == nil {
		return len(pending), errSpendPurgeClickHouseUnavailable
	}

	for i, purge := range pending {
		if r.clickhouseConfigured {
			if err := r.purgeClickHouseSpend(ctx, purge.cutoff); err != nil {
				return len(pending) - i, fmt.Errorf("purge clickhouse spend before %s: %w", purge.cutoff.Format(time.RFC3339), err)
			}
		}

		if _, err := r.pool.Exec(ctx, `
			UPDATE public.clickhouse_spend_purges SET completed_at = NOW()
			WHERE purge_id = $1 AND completed_at IS NULL
		`, purge.id); err != nil {
			return len(pending) - i, fmt.Errorf("complete clickhouse spend purge %d: %w", purge.id, err)
		}
	}

	return 0, nil
}

func (r *SpendHistoryPurgeRepository) listPendingSpendPurges(ctx context.Context) ([]pendingSpendPurge, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT purge_id, cutoff FROM public.clickhouse_spend_purges
		WHERE completed_at IS NULL
		ORDER BY purge_id
	`)
	if err != nil {
		return nil, fmt.Errorf("list clickhouse spend purges: %w", err)
	}
	defer rows.Close()

	var pending []pendingSpendPurge

	for rows.Next() {
		var purge pendingSpendPurge
		if err := rows.Scan(&purge.id, &purge.cutoff); err != nil {
			return nil, fmt.Errorf("scan clickhouse spend purge: %w", err)
		}

		pending = append(pending, purge)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list clickhouse spend purges: %w", err)
	}

	return pending, nil
}

// purgeClickHouseSpend queues one mutation per table. ClickHouse keeps a queued
// mutation until it finishes, so the purge does not wait for the rewrite.
func (r *SpendHistoryPurgeRepository) purgeClickHouseSpend(ctx context.Context, cutoff time.Time) error {
	for _, target := range spendPurgeTargets {
		existing, err := r.clickHouseColumns(ctx, target.table)
		if err != nil {
			return err
		}

		assignments := make([]string, 0, len(target.columns))

		for _, column := range target.columns {
			if _, ok := existing[column.name]; ok {
				assignments = append(assignments, column.name+" = "+column.zero)
			}
		}

		if len(assignments) == 0 {
			continue
		}

		// The table and column names come from spendPurgeTargets, never from input.
		query := fmt.Sprintf("ALTER TABLE %s UPDATE %s WHERE %s < ?", //nolint:gosec // identifiers are compile-time constants
			target.table, strings.Join(assignments, ", "), target.timeColumn)
		if _, err := r.clickhouse.ExecContext(ctx, query, cutoff.UTC()); err != nil {
			return fmt.Errorf("clear spend in %s: %w", target.table, err)
		}
	}

	return nil
}

func (r *SpendHistoryPurgeRepository) clickHouseColumns(ctx context.Context, table string) (map[string]struct{}, error) {
	rows, err := r.clickhouse.QueryContext(ctx, `
		SELECT name FROM system.columns
		WHERE database = currentDatabase() AND table = ?
	`, table)
	if err != nil {
		return nil, fmt.Errorf("read columns of %s: %w", table, err)
	}
	defer closeBestEffort(rows)

	columns := make(map[string]struct{})

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("read columns of %s: %w", table, err)
		}

		columns[name] = struct{}{}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read columns of %s: %w", table, err)
	}

	return columns, nil
}
