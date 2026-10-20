package cache

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// budgetWindowOps is the window surface both cache implementations share.
type budgetWindowOps interface {
	SeedBudgetWindow(ctx context.Context, seed BudgetWindowSeed) (*BudgetWindowResult, error)
	ReserveBudget(ctx context.Context, req BudgetReservationRequest) (*BudgetWindowResult, error)
	CommitBudget(ctx context.Context, req BudgetCommitRequest) (*BudgetWindowResult, error)
	RecordBudgetSpend(ctx context.Context, req BudgetSpendRequest) (*BudgetWindowResult, error)
}

func newTestMemoryCache(t *testing.T) *MemoryCache {
	t.Helper()

	mem, err := NewMemoryCacheWithConfig(zap.NewNop(), MemoryCacheConfig{NumCounters: 1_000, MaxCost: 10_000, BufferItems: 64})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, mem.Close()) })

	return mem
}

// newTestRedisCache connects to REDIS_TEST_ADDR, so the Lua scripts are tested
// against a real Redis when one is available.
func newTestRedisCache(t *testing.T) *DualCache {
	t.Helper()

	addr := strings.TrimSpace(os.Getenv("REDIS_TEST_ADDR"))
	if addr == "" {
		t.Skip("REDIS_TEST_ADDR is not set")
	}

	client := redis.NewClient(&redis.Options{Addr: addr})
	require.NoError(t, client.FlushDB(context.Background()).Err())

	dual, err := newDualCache(client, zap.NewNop())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, dual.Close()) })

	return dual
}

func budgetWindowImplementations(t *testing.T) map[string]func(*testing.T) budgetWindowOps {
	t.Helper()

	return map[string]func(*testing.T) budgetWindowOps{
		"memory": func(t *testing.T) budgetWindowOps { return newTestMemoryCache(t) },
		"redis":  func(t *testing.T) budgetWindowOps { return newTestRedisCache(t) },
	}
}

func seedTestWindow(t *testing.T, ops budgetWindowOps, key string, allowance float64) {
	t.Helper()

	result, err := ops.SeedBudgetWindow(context.Background(), BudgetWindowSeed{
		WindowKey: key, AllowanceIRR: allowance, ExpiresAt: time.Now().UTC().Add(time.Hour), TTL: time.Hour, Version: 1,
	})
	require.NoError(t, err)
	require.True(t, result.Allowed)
}

func TestBudgetWindowForcedCommitRecordsOverage(t *testing.T) {
	for name, build := range budgetWindowImplementations(t) {
		t.Run(name, func(t *testing.T) {
			ops := build(t)
			ctx := context.Background()
			key := "credit:tenant:t1:window:202610"
			seedTestWindow(t, ops, key, 100)

			reserved, err := ops.ReserveBudget(ctx, BudgetReservationRequest{
				WindowKey: key, ReservationKey: key + ":r1", ReservationID: "r1", EstimatedIRR: 50, TTL: time.Hour,
			})
			require.NoError(t, err)
			require.True(t, reserved.Allowed)

			refused, err := ops.CommitBudget(ctx, BudgetCommitRequest{WindowKey: key, ReservationKey: key + ":r1", ActualIRR: 130})
			require.NoError(t, err)
			require.False(t, refused.Allowed, "an unforced commit beyond the allowance is refused")

			forced, err := ops.CommitBudget(ctx, BudgetCommitRequest{WindowKey: key, ReservationKey: key + ":r1", ActualIRR: 130, Force: true})
			require.NoError(t, err)
			require.True(t, forced.Allowed)
			require.InDelta(t, 130, forced.CommittedIRR, 0.001)
			require.InDelta(t, 0, forced.ReservedIRR, 0.001, "the reservation is released by the commit")
			require.InDelta(t, 30, forced.OverageIRR, 0.001)

			again, err := ops.CommitBudget(ctx, BudgetCommitRequest{WindowKey: key, ReservationKey: key + ":r1", ActualIRR: 130, Force: true})
			require.NoError(t, err)
			require.InDelta(t, 130, again.CommittedIRR, 0.001, "a repeated commit adds nothing")
		})
	}
}

func TestBudgetWindowRecordSpendNeedsAWindow(t *testing.T) {
	for name, build := range budgetWindowImplementations(t) {
		t.Run(name, func(t *testing.T) {
			ops := build(t)
			ctx := context.Background()
			key := "credit_policy:t1:rule:r:scope:s:window:202610"

			missing, err := ops.RecordBudgetSpend(ctx, BudgetSpendRequest{WindowKey: key, Amount: 10})
			require.NoError(t, err)
			require.Equal(t, "window_unavailable", missing.State)

			seedTestWindow(t, ops, key, 100)

			first, err := ops.RecordBudgetSpend(ctx, BudgetSpendRequest{WindowKey: key, Amount: 70})
			require.NoError(t, err)
			require.True(t, first.Allowed)
			require.InDelta(t, 70, first.CommittedIRR, 0.001)

			second, err := ops.RecordBudgetSpend(ctx, BudgetSpendRequest{WindowKey: key, Amount: 50})
			require.NoError(t, err)
			require.True(t, second.Allowed, "recorded spend is never refused")
			require.InDelta(t, 120, second.CommittedIRR, 0.001)
			require.InDelta(t, 20, second.OverageIRR, 0.001)
		})
	}
}

func TestBudgetWindowZeroEstimateIsAdmittedUntilExhausted(t *testing.T) {
	for name, build := range budgetWindowImplementations(t) {
		t.Run(name, func(t *testing.T) {
			ops := build(t)
			ctx := context.Background()
			key := "credit:tenant:t1:window:202610"
			seedTestWindow(t, ops, key, 100)

			reserve := func(id string) *BudgetWindowResult {
				t.Helper()

				result, err := ops.ReserveBudget(ctx, BudgetReservationRequest{
					WindowKey: key, ReservationKey: key + ":res:" + id, ReservationID: id, TTL: time.Hour, Now: time.Now().UTC(),
				})
				require.NoError(t, err)

				return result
			}

			commit := func(id string, actual float64) {
				t.Helper()

				result, err := ops.CommitBudget(ctx, BudgetCommitRequest{
					WindowKey: key, ReservationKey: key + ":res:" + id, ActualIRR: actual, Force: true, Now: time.Now().UTC(),
				})
				require.NoError(t, err)
				require.True(t, result.Allowed)
			}

			require.True(t, reserve("r1").Allowed)
			commit("r1", 99)

			require.True(t, reserve("r2").Allowed, "headroom is left")
			commit("r2", 1)

			exhausted := reserve("r3")
			require.False(t, exhausted.Allowed, "a window spent exactly to its allowance refuses")
			require.Equal(t, "exhausted", exhausted.State)
		})
	}
}
