package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Without ClickHouse there is no spend to clear, so a pending purge completes
// at once and the worker stops asking.
func TestIntegration_ClickHouseSpendPurgeCompletesWithoutClickHouse(t *testing.T) {
	store, pool := newCreditIntegrationStore(t)
	ctx := context.Background()

	var purgeID int64
	require.NoError(t, pool.QueryRow(ctx, `
		INSERT INTO public.clickhouse_spend_purges (reason, cutoff) VALUES ('integration_test', $1)
		RETURNING purge_id
	`, time.Now().UTC()).Scan(&purgeID))
	t.Cleanup(func() {
		_, err := pool.Exec(context.Background(), `DELETE FROM public.clickhouse_spend_purges WHERE purge_id = $1`, purgeID)
		require.NoError(t, err)
	})

	pending, err := store.RunPendingSpendPurges(ctx)
	require.NoError(t, err)
	require.Zero(t, pending)

	var completed bool
	require.NoError(t, pool.QueryRow(ctx, `
		SELECT completed_at IS NOT NULL FROM public.clickhouse_spend_purges WHERE purge_id = $1
	`, purgeID).Scan(&completed))
	require.True(t, completed)

	pending, err = store.RunPendingSpendPurges(ctx)
	require.NoError(t, err)
	require.Zero(t, pending, "a completed purge is not run again")
}
