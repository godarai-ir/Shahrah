package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/billing"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/credits"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/seats"
	"github.com/Delpak/shahrah/services/control-plane/internal/storage"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

// freeTenant is a credit tenant on the free plan, as every new tenant is.
func freeTenant(t *testing.T, pool *pgxpool.Pool) string {
	t.Helper()

	planID, versionID := draftCreditVersion(t, pool, seats.FreePlanSlug)
	tenantID := newCreditLedgerTenant(t, pool, credits.BillingUnitCredit)

	_, err := pool.Exec(context.Background(), `
		INSERT INTO tenant_subscriptions (tenant_id, plan_id, version_id, status, billing_interval,
		                                  current_period_start, current_period_end)
		VALUES ($1::uuid, $2::uuid, $3::uuid, 'active', 'monthly', NOW(), $4)
	`, tenantID, planID, versionID, seats.FreePeriodEnd)
	require.NoError(t, err)

	return tenantID
}

func currentPlanSlug(t *testing.T, pool *pgxpool.Pool, tenantID string) (string, string, time.Time) {
	t.Helper()

	var (
		slug, interval string
		periodEnd      time.Time
	)

	require.NoError(t, pool.QueryRow(context.Background(), `
		SELECT p.slug, ts.billing_interval, ts.current_period_end
		FROM tenant_subscriptions ts JOIN commercial_plans p ON p.plan_id = ts.plan_id
		WHERE ts.tenant_id = $1::uuid AND ts.status IN ('trialing', 'active', 'past_due')
	`, tenantID).Scan(&slug, &interval, &periodEnd))

	return slug, interval, periodEnd
}

func paySeatOrder(t *testing.T, store *storage.Store, tenantID string, req platform.CreateSeatOrderRequest, now time.Time) {
	t.Helper()

	ctx := context.Background()
	req.Provider = "mock"

	order, err := store.CreateSeatOrder(ctx, tenantID, req, "https://example.test/callback", "https://example.test/return", "test", now)
	require.NoError(t, err)

	authority := "seat-sub-" + uuid.NewString()
	require.NoError(t, store.AttachSeatOrderGatewayData(ctx, order.PublicID, authority, "https://pay.example", nil))

	_, duplicate, err := store.CompleteSeatOrderByAuthority(ctx, authority, billing.PaymentVerification{RefID: "ref"}, now)
	require.NoError(t, err)
	require.False(t, duplicate)
}

func TestIntegration_SeatSubscriptionFreeTenantHasFiveFreeSeats(t *testing.T) {
	store, pool := newCreditIntegrationStore(t)
	ctx := context.Background()
	now := time.Now().UTC()

	tenantID := freeTenant(t, pool)

	members := make([]string, 0, seats.FreeSeatLimit+1)
	for range seats.FreeSeatLimit + 1 {
		members = append(members, addTenantUser(t, pool, tenantID, true))
	}

	_, err := store.RunSeatLifecycle(ctx, now)
	require.NoError(t, err)

	summary, err := store.GetSeatSummary(ctx, tenantID, now)
	require.NoError(t, err)
	require.Equal(t, seats.FreePlanSlug, summary.SubscriptionKind)
	require.Nil(t, summary.PeriodEnd, "the free plan has no billing period to show")
	require.NotNil(t, summary.FreeLimit)
	require.Equal(t, seats.FreeSeatLimit, *summary.FreeLimit)
	require.Equal(t, seats.FreeSeatLimit, summary.FreeUsed)
	require.Equal(t, 1, summary.Unassigned, "the sixth member gets no free seat")

	unassigned := ""

	for _, member := range members {
		seat, seatErr := store.GetSeatMember(ctx, tenantID, member, now)
		require.NoError(t, seatErr)

		if seat.Seat == string(seats.Unassigned) {
			unassigned = member
		}
	}

	require.NotEmpty(t, unassigned)

	_, err = store.SetMemberSeat(ctx, tenantID, unassigned, seats.Free, "test", "", now)
	require.ErrorIs(t, err, seats.ErrFreeSeatsFull)

	// A paid seat needs stock, not a plan: the admin is asked to buy one.
	_, err = store.SetMemberSeat(ctx, tenantID, unassigned, seats.Pro, "test", "", now)
	require.ErrorIs(t, err, seats.ErrSeatPurchaseRequired)
}

func TestIntegration_SeatSubscriptionFirstPaidSeatStartsIt(t *testing.T) {
	store, pool := newCreditIntegrationStore(t)
	ctx := context.Background()
	now := time.Now().UTC()

	tenantID := freeTenant(t, pool)
	member := addTenantUser(t, pool, tenantID, true)

	quote, err := store.QuoteSeats(ctx, tenantID, platform.SeatQuoteRequest{Seat: "pro", Quantity: 1, BillingInterval: "yearly"}, now)
	require.NoError(t, err)
	require.True(t, quote.StartsSubscription)
	require.Equal(t, "yearly", quote.BillingInterval)
	require.Equal(t, quote.UnitPeriodPriceIRR, quote.PriceIRR, "the first purchase pays a whole period")
	require.WithinDuration(t, now.AddDate(1, 0, 0), quote.PeriodEnd, time.Second)

	paySeatOrder(t, store, tenantID, platform.CreateSeatOrderRequest{
		Seat: "pro", Quantity: 1, BillingInterval: "yearly", AssignUserID: member,
	}, now)

	slug, interval, periodEnd := currentPlanSlug(t, pool, tenantID)
	require.Equal(t, seats.SeatsPlanSlug, slug)
	require.Equal(t, "yearly", interval)
	require.WithinDuration(t, now.AddDate(1, 0, 0), periodEnd, time.Second)

	seat, err := store.GetSeatMember(ctx, tenantID, member, now)
	require.NoError(t, err)
	require.Equal(t, "pro", seat.Seat)
	require.Equal(t, int64(2_000_000), personalRemainingMC(t, pool, tenantID, member))

	// Later seats follow the running subscription: prorated, same interval.
	later := now.AddDate(0, 3, 0)

	prorated, err := store.QuoteSeats(ctx, tenantID, platform.SeatQuoteRequest{Seat: "max", Quantity: 1}, later)
	require.NoError(t, err)
	require.False(t, prorated.StartsSubscription)
	require.Less(t, prorated.PriceIRR, prorated.UnitPeriodPriceIRR)
	require.Equal(t, periodEnd.UTC(), prorated.PeriodEnd.UTC())

	_, err = store.QuoteSeats(ctx, tenantID, platform.SeatQuoteRequest{Seat: "max", Quantity: 1, BillingInterval: "monthly"}, later)
	require.ErrorIs(t, err, seats.ErrIntervalLocked)
}

func TestIntegration_SeatSubscriptionEndsBackOnTheFreePlan(t *testing.T) {
	store, pool := newCreditIntegrationStore(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)
	start := now.AddDate(0, -1, -1)
	end := start.AddDate(0, 1, 0)

	tenantID := seatTenant(t, pool, start)
	member := addTenantUser(t, pool, tenantID, true)

	_, err := pool.Exec(ctx, `
		INSERT INTO tenant_seat_blocks (tenant_id, seat_type, quantity, paid_through) VALUES ($1::uuid, 'pro', 1, $2)
	`, tenantID, end)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
		INSERT INTO tenant_seat_assignments (tenant_id, user_id, seat_type) VALUES ($1::uuid, $2::uuid, 'pro')
	`, tenantID, member)
	require.NoError(t, err)

	_, err = store.ProcessSubscriptionTransitions(ctx, now, 1000)
	require.NoError(t, err)

	slug, _, periodEnd := currentPlanSlug(t, pool, tenantID)
	require.Equal(t, seats.FreePlanSlug, slug, "an unrenewed seat subscription goes back to free")
	require.Equal(t, seats.FreePeriodEnd, periodEnd.UTC())

	_, err = store.RunSeatLifecycle(ctx, now)
	require.NoError(t, err)

	seat, err := store.GetSeatMember(ctx, tenantID, member, now)
	require.NoError(t, err)
	require.Equal(t, string(seats.Unassigned), seat.Seat, "only the paid seat goes; it never falls back to free")

	summary, err := store.GetSeatSummary(ctx, tenantID, now)
	require.NoError(t, err)
	require.Equal(t, seats.FreePlanSlug, summary.SubscriptionKind)
	require.True(t, summary.Enforced, "the free plan is active, so the API and free seats keep working")
}

func TestIntegration_SeatSubscriptionRenewalRollsThePeriod(t *testing.T) {
	store, pool := newCreditIntegrationStore(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)
	start := now.AddDate(0, -1, 3)
	end := start.AddDate(0, 1, 0)

	tenantID := seatTenant(t, pool, start)

	_, err := pool.Exec(ctx, `
		INSERT INTO tenant_seat_blocks (tenant_id, seat_type, quantity, paid_through) VALUES ($1::uuid, 'max', 2, $2)
	`, tenantID, end)
	require.NoError(t, err)

	early := end.Add(-seats.RenewalWindow - time.Hour)
	_, err = store.QuoteSeatRenewal(ctx, tenantID, early)
	require.ErrorIs(t, err, seats.ErrRenewalNotDue, "renewal opens a week before the period ends")

	renewal, err := store.QuoteSeatRenewal(ctx, tenantID, now)
	require.NoError(t, err)
	require.Equal(t, end.AddDate(0, 1, 0).UTC(), renewal.PeriodEnd.UTC())
	require.Len(t, renewal.Lines, 1)
	require.Equal(t, 2, renewal.Lines[0].Quantity)

	price, err := store.QuoteSeats(ctx, tenantID, platform.SeatQuoteRequest{Seat: "max", Quantity: 1}, now)
	require.NoError(t, err)
	require.Equal(t, 2*price.UnitPeriodPriceIRR, renewal.PriceIRR, "a renewal pays the whole next period")

	paySeatOrder(t, store, tenantID, platform.CreateSeatOrderRequest{Kind: platform.SeatOrderRenewal}, now)

	_, err = store.QuoteSeatRenewal(ctx, tenantID, now)
	require.ErrorIs(t, err, seats.ErrRenewalNotDue, "nothing is left to renew")

	_, err = store.ProcessSubscriptionTransitions(ctx, end.Add(time.Minute), 1000)
	require.NoError(t, err)

	slug, _, periodEnd := currentPlanSlug(t, pool, tenantID)
	require.Equal(t, seats.SeatsPlanSlug, slug)
	require.Equal(t, end.AddDate(0, 1, 0).UTC(), periodEnd.UTC())
}
