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

// seatTenant is a credit tenant with a monthly seat subscription whose period
// started start.
func seatTenant(t *testing.T, pool *pgxpool.Pool, start time.Time) string {
	t.Helper()

	planID, versionID := draftCreditVersion(t, pool, seats.SeatsPlanSlug)
	tenantID := newCreditLedgerTenant(t, pool, credits.BillingUnitCredit)

	_, err := pool.Exec(context.Background(), `
		INSERT INTO tenant_subscriptions (tenant_id, plan_id, version_id, status, billing_interval,
		                                  current_period_start, current_period_end)
		VALUES ($1::uuid, $2::uuid, $3::uuid, 'active', 'monthly', $4, $5)
	`, tenantID, planID, versionID, start, start.AddDate(0, 1, 0))
	require.NoError(t, err)

	return tenantID
}

func insertChatSpendLog(t *testing.T, pool *pgxpool.Pool, tenantID, userID, product string, creditsMC int64) {
	t.Helper()

	_, err := pool.Exec(context.Background(), `
		INSERT INTO spend_logs (request_id, tenant_id, user_id, model, cost_usd, status, credits_mc, rate_card_version,
		                        product, credit_source, created_at, pricing_snapshot_json)
		VALUES ($1, $2::uuid, $3::uuid, 'openai/gpt-test', 0.002, 'success', $4, 1, $5, 'components',
		        NOW() - INTERVAL '5 minutes', '{"revision":"rev-7"}'::jsonb)
	`, "seat-"+uuid.NewString(), tenantID, userID, creditsMC, product)
	require.NoError(t, err)
}

func personalRemainingMC(t *testing.T, pool *pgxpool.Pool, tenantID, userID string) int64 {
	t.Helper()

	var remaining int64

	require.NoError(t, pool.QueryRow(context.Background(), `
		SELECT COALESCE(SUM(remaining_mc), 0)::bigint FROM credit_lot_balances
		WHERE tenant_id = $1::uuid AND owner_user_id = $2::uuid
	`, tenantID, userID).Scan(&remaining))

	return remaining
}

func poolRemainingMC(t *testing.T, pool *pgxpool.Pool, tenantID string) int64 {
	t.Helper()

	var remaining int64

	require.NoError(t, pool.QueryRow(context.Background(), `
		SELECT COALESCE(SUM(remaining_mc), 0)::bigint FROM credit_lot_balances
		WHERE tenant_id = $1::uuid AND owner_user_id IS NULL
	`, tenantID).Scan(&remaining))

	return remaining
}

func grantPoolCredits(t *testing.T, store *storage.Store, tenantID string, amountMC int64) {
	t.Helper()

	_, err := store.GrantCredits(context.Background(), storage.CreditGrant{
		TenantID: tenantID, Kind: credits.EntryPurchase, Source: credits.LotSourcePurchaseAddon,
		AmountMC: amountMC, IdempotencyKey: "seat-test-pool:" + uuid.NewString(),
	})
	require.NoError(t, err)
}

func TestIntegration_SeatsFreeSeatIsGrantedOnceAndSpentBeforeThePool(t *testing.T) {
	store, pool := newCreditIntegrationStore(t)
	ctx := context.Background()
	now := time.Now().UTC()

	tenantID := seatTenant(t, pool, now.AddDate(0, 0, -3))
	member := addTenantUser(t, pool, tenantID, true)
	grantPoolCredits(t, store, tenantID, 10_000_000)

	result, err := store.RunSeatLifecycle(ctx, now)
	require.NoError(t, err)
	require.GreaterOrEqual(t, result.AutoAssigned, 1, "a member without a seat gets the free seat")
	require.Equal(t, int64(500_000), personalRemainingMC(t, pool, tenantID, member))

	_, err = store.RunSeatLifecycle(ctx, now.Add(time.Minute))
	require.NoError(t, err)
	require.Equal(t, int64(500_000), personalRemainingMC(t, pool, tenantID, member), "the free seat is granted once")

	// Chat usage spends the seat first and falls through to the pool;
	// gateway (API) usage never touches personal credits.
	insertChatSpendLog(t, pool, tenantID, member, platform.CreditProductLibreChat, 600_000)
	insertChatSpendLog(t, pool, tenantID, member, platform.CreditProductGateway, 1_000_000)
	settleAllCredits(t, store)

	require.Zero(t, personalRemainingMC(t, pool, tenantID, member))
	require.Equal(t, int64(10_000_000-100_000-1_000_000), poolRemainingMC(t, pool, tenantID))

	windowID, _ := credits.PoolWindow(now)

	allowance, err := store.SeatWindowAllowanceMC(ctx, tenantID, member, windowID)
	require.NoError(t, err)
	require.Equal(t, int64(500_000), allowance, "the seat window adds back what was settled from the seat")

	// Leaving and coming back does not restore the free seat.
	_, err = pool.Exec(ctx, `DELETE FROM tenant_seat_assignments WHERE tenant_id = $1::uuid AND user_id = $2::uuid`, tenantID, member)
	require.NoError(t, err)

	_, err = store.SetMemberSeat(ctx, tenantID, member, seats.Free, "test", "", now)
	require.ErrorIs(t, err, seats.ErrFreeSeatUsed)
}

func TestIntegration_SeatsPurchaseAssignsAndGrantsAProSeat(t *testing.T) {
	store, pool := newCreditIntegrationStore(t)
	ctx := context.Background()
	now := time.Now().UTC()

	tenantID := seatTenant(t, pool, now.AddDate(0, 0, -10))
	member := addTenantUser(t, pool, tenantID, true)

	_, err := store.SetMemberSeat(ctx, tenantID, member, seats.Pro, "test", "", now)
	require.ErrorIs(t, err, seats.ErrSeatPurchaseRequired, "no Pro seat in stock yet")

	quote, err := store.QuoteSeats(ctx, tenantID, platform.SeatQuoteRequest{Seat: "pro", Quantity: 2}, now)
	require.NoError(t, err)
	require.Positive(t, quote.PriceIRR)
	require.Less(t, quote.PriceIRR, 2*quote.UnitPeriodPriceIRR, "a mid-period purchase is prorated")
	require.Equal(t, quote.PriceIRR+quote.TaxIRR, quote.PayableIRR)

	order, err := store.CreateSeatOrder(ctx, tenantID, platform.CreateSeatOrderRequest{
		Seat: "pro", Quantity: 2, Provider: "mock", AssignUserID: member,
	}, "https://example.test/callback", "https://example.test/return", "test", now)
	require.NoError(t, err)

	authority := "seat-auth-" + uuid.NewString()
	require.NoError(t, store.AttachSeatOrderGatewayData(ctx, order.PublicID, authority, "https://pay.example", nil))

	completed, duplicate, err := store.CompleteSeatOrderByAuthority(ctx, authority, billing.PaymentVerification{RefID: "ref-1"}, now)
	require.NoError(t, err)
	require.False(t, duplicate)
	require.Equal(t, platform.CreditPackOrderCompleted, completed.Status)

	_, duplicate, err = store.CompleteSeatOrderByAuthority(ctx, authority, billing.PaymentVerification{RefID: "ref-1"}, now)
	require.NoError(t, err)
	require.True(t, duplicate, "a repeated callback changes nothing")

	member2, err := store.GetSeatMember(ctx, tenantID, member, now)
	require.NoError(t, err)
	require.Equal(t, "pro", member2.Seat)
	require.Equal(t, int64(2_000_000), personalRemainingMC(t, pool, tenantID, member), "the full monthly quota at once")

	summary, err := store.GetSeatSummary(ctx, tenantID, now)
	require.NoError(t, err)

	for _, paid := range summary.Paid {
		if paid.Type == "pro" {
			require.Equal(t, 2, paid.Purchased)
			require.Equal(t, 1, paid.Assigned)
			require.Equal(t, 1, paid.Available)
		}
	}

	usage, err := store.GetTenantSeatUsage(ctx, tenantID)
	require.NoError(t, err)
	require.NotNil(t, usage.Limit)
	require.Equal(t, seats.FreeSeatLimit+2, *usage.Limit, "paid seats add members beyond the free seats")
}

func TestIntegration_SeatsDowngradesWaitForThePeriodEnd(t *testing.T) {
	store, pool := newCreditIntegrationStore(t)
	ctx := context.Background()
	now := time.Now().UTC()
	start := now.AddDate(0, 0, -5)

	tenantID := seatTenant(t, pool, start)
	member := addTenantUser(t, pool, tenantID, true)

	_, err := pool.Exec(ctx, `
		INSERT INTO tenant_seat_blocks (tenant_id, seat_type, quantity, paid_through)
		VALUES ($1::uuid, 'max', 1, $2), ($1::uuid, 'pro', 1, $2)
	`, tenantID, start.AddDate(0, 1, 0))
	require.NoError(t, err)

	_, err = store.SetMemberSeat(ctx, tenantID, member, seats.Max, "test", "", now)
	require.NoError(t, err)

	_, err = store.SetMemberSeat(ctx, tenantID, member, seats.Free, "test", "", now)
	require.ErrorIs(t, err, seats.ErrBackToFree)

	result, err := store.SetMemberSeat(ctx, tenantID, member, seats.Pro, "test", "", now)
	require.NoError(t, err)
	require.Equal(t, string(seats.AtPeriodEnd), result.Timing)
	require.Equal(t, "max", result.Member.Seat, "Max stays until the period ends")
	require.Equal(t, "pro", result.Member.PendingSeat)

	_, err = store.RunSeatLifecycle(ctx, start.AddDate(0, 1, 0).Add(time.Minute))
	require.NoError(t, err)

	after, err := store.GetSeatMember(ctx, tenantID, member, now)
	require.NoError(t, err)
	require.NotEqual(t, "max", after.Seat, "the downgrade applied at the period end")
}

func TestIntegration_SeatsUpgradeRequests(t *testing.T) {
	store, pool := newCreditIntegrationStore(t)
	ctx := context.Background()
	now := time.Now().UTC()

	tenantID := seatTenant(t, pool, now.AddDate(0, 0, -1))
	member := addTenantUser(t, pool, tenantID, true)

	_, err := store.SetMemberSeat(ctx, tenantID, member, seats.Free, "test", "", now)
	require.NoError(t, err)

	request, err := store.CreateSeatUpgradeRequest(ctx, tenantID, member, platform.CreateSeatUpgradeRequest{Seat: "pro", Reason: "more chat"})
	require.NoError(t, err)
	require.Equal(t, platform.SeatRequestPending, request.Status)
	require.Equal(t, "free", request.CurrentSeat)

	_, err = store.CreateSeatUpgradeRequest(ctx, tenantID, member, platform.CreateSeatUpgradeRequest{Seat: "max"})
	require.ErrorIs(t, err, seats.ErrRequestOpen, "one open request per member")

	denied, err := store.DenySeatUpgradeRequest(ctx, tenantID, request.ID, "admin", "not this month", now)
	require.NoError(t, err)
	require.Equal(t, platform.SeatRequestDenied, denied.Status)

	_, err = store.CreateSeatUpgradeRequest(ctx, tenantID, member, platform.CreateSeatUpgradeRequest{Seat: "free"})
	require.ErrorIs(t, err, seats.ErrInvalidSeatType)
}

func TestIntegration_SeatsListMembersCountsAndFilters(t *testing.T) {
	store, pool := newCreditIntegrationStore(t)
	ctx := context.Background()
	now := time.Now().UTC()

	tenantID := seatTenant(t, pool, now.AddDate(0, 0, -1))
	member := addTenantUser(t, pool, tenantID, true)
	addTenantUser(t, pool, tenantID, true)

	_, err := store.SetMemberSeat(ctx, tenantID, member, seats.Free, "test", "", now)
	require.NoError(t, err)

	all, err := store.ListSeatMembers(ctx, tenantID, "", "", 100, 0, now)
	require.NoError(t, err)
	require.Equal(t, 2, all.Total)
	require.Len(t, all.Items, 2)

	free, err := store.ListSeatMembers(ctx, tenantID, "", "free", 100, 0, now)
	require.NoError(t, err)
	require.Equal(t, 1, free.Total)
	require.Equal(t, member, free.Items[0].UserID)
}

func TestIntegration_SupportTenantListShowsSeatsAndPool(t *testing.T) {
	store, pool := newCreditIntegrationStore(t)
	ctx := context.Background()
	now := time.Now().UTC()

	tenantID := seatTenant(t, pool, now.AddDate(0, 0, -3))
	member := addTenantUser(t, pool, tenantID, true)

	_, err := pool.Exec(ctx, `
		INSERT INTO tenant_seat_blocks (tenant_id, seat_type, quantity, paid_through) VALUES ($1::uuid, 'pro', 2, $2)
	`, tenantID, now.AddDate(0, 1, 0))
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
		INSERT INTO tenant_seat_assignments (tenant_id, user_id, seat_type) VALUES ($1::uuid, $2::uuid, 'pro')
	`, tenantID, member)
	require.NoError(t, err)

	grantPoolCredits(t, store, tenantID, 3*credits.MillicreditsPerCredit)

	items, err := store.ListSupportTenants(ctx, tenantID, "", "", 10, 0)
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, 1, items[0].SeatsPro)
	require.Equal(t, 2, items[0].SeatsProPurchased)
	require.Zero(t, items[0].SeatsFree+items[0].SeatsMax+items[0].SeatsMaxPurchased)
	require.Equal(t, credits.Format(credits.Millicredits(3*credits.MillicreditsPerCredit)), items[0].PoolCredits)

	detail, err := store.GetSupportTenantSummary(ctx, tenantID)
	require.NoError(t, err)
	require.Equal(t, items[0].PoolCredits, detail.PoolCredits)
}
