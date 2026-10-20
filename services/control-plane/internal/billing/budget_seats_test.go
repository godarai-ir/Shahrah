package billing

import (
	"context"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/credits"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/seats"
	"github.com/Delpak/shahrah/services/control-plane/internal/runtimeprofile"
	"github.com/Delpak/shahrah/services/control-plane/internal/testcache"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type staticSeatCredits struct {
	allowanceMC map[string]int64
	seatTypes   map[string]string
}

func (s staticSeatCredits) SeatWindowAllowanceMC(_ context.Context, _, userID, _ string) (int64, error) {
	return s.allowanceMC[userID], nil
}

func (s staticSeatCredits) MemberSeatType(_ context.Context, _, userID string) (string, error) {
	if seatType, ok := s.seatTypes[userID]; ok {
		return seatType, nil
	}

	return platform.SeatUnassigned, nil
}

type seatHarness struct {
	service *Service
	cache   *cache.MemoryCache
}

func newSeatHarness(t *testing.T, poolMC int64, seatProfile *platform.SeatProfile, store staticSeatCredits) seatHarness {
	t.Helper()

	mem, err := testcache.New(t, zap.NewNop())
	require.NoError(t, err)

	profile := &runtimeprofile.Profile{
		TenantID: "tenant-1",
		Tenant:   platform.Tenant{ID: "tenant-1", BillingUnit: credits.BillingUnitCredit},
		Seats:    seatProfile,
	}

	service := NewServiceWithDependencies(Dependencies{
		Cache:           mem,
		BudgetWindows:   mem,
		RuntimeProfiles: staticRuntimeProfiles{profile: profile},
		CreditPools:     &staticCreditPools{allowanceMC: poolMC},
		SeatCredits:     store,
		Config:          &ServiceConfig{SeatsEnabled: true, BudgetWindowGrace: time.Hour},
	}, 850000, "test")

	return seatHarness{service: service, cache: mem}
}

func chatAuth(userID string) *platform.AuthContext {
	return &platform.AuthContext{TenantID: "tenant-1", UserID: userID, CreditProduct: platform.CreditProductLibreChat}
}

func (h seatHarness) committedMC(t *testing.T, windowKey string) int64 {
	t.Helper()

	snapshot, err := h.cache.GetBudgetWindow(context.Background(), windowKey)
	require.NoError(t, err)

	if snapshot == nil {
		return 0
	}

	return int64(snapshot.CommittedIRR)
}

func activeSeats(members map[string]string, limits map[string]int64) *platform.SeatProfile {
	return &platform.SeatProfile{SubscriptionStatus: "active", Members: members, PoolLimitMC: limits}
}

func TestSeatChatSpendsTheSeatBeforeThePool(t *testing.T) {
	h := newSeatHarness(t, 100_000, activeSeats(map[string]string{"free-1": "free"}, nil),
		staticSeatCredits{allowanceMC: map[string]int64{"free-1": 5_000}})
	ctx := context.Background()
	windowID, _ := credits.PoolWindow(time.Now())
	seatKey := seats.WindowKey("tenant-1", "free-1", windowID)
	poolKey := credits.PoolWindowKey("tenant-1", windowID)

	first, err := h.service.ReserveBudget(ctx, chatAuth("free-1"), creditBreakdown(1), "req-1")
	require.NoError(t, err)
	require.Equal(t, "free-1", first.SeatUserID)
	require.Equal(t, seatKey, first.WindowKey, "the seat window is the primary target")
	require.NoError(t, h.service.CommitBudget(ctx, first, creditBreakdown(3)))
	require.Equal(t, int64(3_000), h.committedMC(t, seatKey))
	require.Zero(t, h.committedMC(t, poolKey), "the seat covered the whole request")

	second, err := h.service.ReserveBudget(ctx, chatAuth("free-1"), creditBreakdown(1), "req-2")
	require.NoError(t, err)
	require.NoError(t, h.service.CommitBudget(ctx, second, creditBreakdown(4)))
	require.Equal(t, int64(2_000), h.committedMC(t, poolKey), "only the overflow reaches the pool")

	_, err = h.service.ReserveBudget(ctx, chatAuth("free-1"), creditBreakdown(1), "req-3")
	require.ErrorIs(t, err, ErrBudgetExceeded)
	require.Equal(t, string(seats.BlockSeatCreditsExhausted), BudgetErrorCode(err), "a free seat has no pool access")
}

func TestSeatPaidMembersFallThroughToThePoolUnderTheirLimit(t *testing.T) {
	h := newSeatHarness(t, 100_000,
		activeSeats(map[string]string{"pro-1": "pro"}, map[string]int64{"pro": 2_000}),
		staticSeatCredits{allowanceMC: map[string]int64{"pro-1": 1_000}})
	ctx := context.Background()

	first, err := h.service.ReserveBudget(ctx, chatAuth("pro-1"), creditBreakdown(1), "req-1")
	require.NoError(t, err)
	require.NoError(t, h.service.CommitBudget(ctx, first, creditBreakdown(1)))

	fromPool, err := h.service.ReserveBudget(ctx, chatAuth("pro-1"), creditBreakdown(1), "req-2")
	require.NoError(t, err, "an exhausted Pro seat continues on the pool")
	require.Equal(t, credits.PoolWindowKey("tenant-1", fromPool.WindowID), fromPool.WindowKey)
	require.NoError(t, h.service.CommitBudget(ctx, fromPool, creditBreakdown(2)))

	_, err = h.service.ReserveBudget(ctx, chatAuth("pro-1"), creditBreakdown(1), "req-3")
	require.ErrorIs(t, err, ErrBudgetExceeded)
	require.Equal(t, string(seats.BlockPersonalLimitReached), BudgetErrorCode(err))
}

func TestSeatUnassignedMembersAreRefusedOnlyOnceEnforced(t *testing.T) {
	store := staticSeatCredits{}
	ctx := context.Background()

	enforced := newSeatHarness(t, 100_000, activeSeats(map[string]string{}, nil), store)
	_, err := enforced.service.ReserveBudget(ctx, chatAuth("new-1"), creditBreakdown(1), "req-1")
	require.ErrorIs(t, err, ErrBudgetExceeded)
	require.Equal(t, string(seats.BlockSeatUnassigned), BudgetErrorCode(err))

	later := time.Now().Add(24 * time.Hour)
	grace := activeSeats(map[string]string{}, nil)
	grace.EnforcedFrom = &later

	graceHarness := newSeatHarness(t, 100_000, grace, store)
	reservation, err := graceHarness.service.ReserveBudget(ctx, chatAuth("new-1"), creditBreakdown(1), "req-1")
	require.NoError(t, err, "before enforcement a member without a seat shares the pool")
	require.Empty(t, reservation.SeatUserID)

	trial := newSeatHarness(t, 100_000, &platform.SeatProfile{SubscriptionStatus: "trialing"}, store)
	_, err = trial.service.ReserveBudget(ctx, chatAuth("new-1"), creditBreakdown(1), "req-1")
	require.NoError(t, err, "trialing tenants share their trial pool")
}

func TestSeatAPIUsageSpendsOnlyThePool(t *testing.T) {
	h := newSeatHarness(t, 100_000, activeSeats(map[string]string{"pro-1": "pro"}, nil),
		staticSeatCredits{allowanceMC: map[string]int64{"pro-1": 50_000}})

	auth := &platform.AuthContext{TenantID: "tenant-1", UserID: "pro-1", CreditProduct: platform.CreditProductGateway}

	reservation, err := h.service.ReserveBudget(context.Background(), auth, creditBreakdown(1), "req-1")
	require.NoError(t, err)
	require.Empty(t, reservation.SeatUserID, "API keys never touch personal seat credits")
	require.Equal(t, credits.PoolWindowKey("tenant-1", reservation.WindowID), reservation.WindowKey)
}

func TestSeatProfileMissFallsBackToTheStore(t *testing.T) {
	h := newSeatHarness(t, 100_000, activeSeats(map[string]string{}, nil),
		staticSeatCredits{allowanceMC: map[string]int64{"joined-1": 2_000}, seatTypes: map[string]string{"joined-1": "free"}})

	reservation, err := h.service.ReserveBudget(context.Background(), chatAuth("joined-1"), creditBreakdown(1), "req-1")
	require.NoError(t, err, "a member who joined after the profile was compiled is looked up")
	require.Equal(t, "joined-1", reservation.SeatUserID)
}
