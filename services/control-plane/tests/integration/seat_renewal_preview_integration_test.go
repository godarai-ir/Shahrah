package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/seats"
	"github.com/stretchr/testify/require"
)

// The renewal preview tells the admin who loses a seat; the seat worker is what
// takes it away. They must agree, so this renews fewer seats than members hold
// them and compares the preview with what the worker did.
func TestIntegration_SeatRenewalPreviewMatchesTheWorker(t *testing.T) {
	store, pool := newCreditIntegrationStore(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)
	start := now.AddDate(0, -1, 3)
	end := start.AddDate(0, 1, 0)

	tenantID := seatTenant(t, pool, start)

	_, err := pool.Exec(ctx, `
		INSERT INTO tenant_seat_blocks (tenant_id, seat_type, quantity, paid_through) VALUES ($1::uuid, 'pro', 3, $2)
	`, tenantID, end)
	require.NoError(t, err)

	// Three members, the last assigned most recently.
	members := make([]string, 0, 3)

	for i := range 3 {
		member := addTenantUser(t, pool, tenantID, true)
		members = append(members, member)

		_, err = pool.Exec(ctx, `
			INSERT INTO tenant_seat_assignments (tenant_id, user_id, seat_type, assigned_at) VALUES ($1::uuid, $2::uuid, 'pro', $3)
		`, tenantID, member, now.Add(-time.Duration(3-i)*24*time.Hour))
		require.NoError(t, err)
	}

	preview, err := store.PreviewSeatRenewal(ctx, tenantID, map[seats.Type]int{seats.Pro: 2}, now)
	require.NoError(t, err)

	var pro platform.SeatRenewalPreviewLine

	for _, line := range preview.Lines {
		if line.Seat == string(seats.Pro) {
			pro = line
		}
	}

	require.Equal(t, 3, pro.Held)
	require.Equal(t, 3, pro.Assigned)
	require.Len(t, pro.Losing, 1)
	require.Equal(t, members[2], pro.Losing[0].UserID, "the latest assigned member loses the seat")

	nextEnd, err := seats.PeriodEnd(end, "monthly")
	require.NoError(t, err)
	require.Equal(t, nextEnd.UTC(), preview.NextPeriodEnd.UTC())
	require.Equal(t, 2*pro.UnitPriceIRR, preview.PriceIRR, "only the two kept seats are paid")

	// Renew two of the three seats and let the period end.
	_, err = store.UpdateSeatRenewal(ctx, tenantID, platform.UpdateSeatRenewalRequest{Seat: "pro", Quantity: 2}, now)
	require.NoError(t, err)

	paySeatOrder(t, store, tenantID, platform.CreateSeatOrderRequest{Kind: platform.SeatOrderRenewal}, now)

	after := end.Add(time.Minute)
	_, err = store.ProcessSubscriptionTransitions(ctx, after, 1000)
	require.NoError(t, err)

	_, err = store.RunSeatLifecycle(ctx, after)
	require.NoError(t, err)

	for _, kept := range members[:2] {
		seat, seatErr := store.GetSeatMember(ctx, tenantID, kept, after)
		require.NoError(t, seatErr)
		require.Equal(t, string(seats.Pro), seat.Seat)
	}

	lost, err := store.GetSeatMember(ctx, tenantID, members[2], after)
	require.NoError(t, err)
	require.Equal(t, string(seats.Unassigned), lost.Seat, "the worker took the seat the preview named")

	// The next pass gives the free seat the preview promised.
	_, err = store.RunSeatLifecycle(ctx, after.Add(time.Minute))
	require.NoError(t, err)

	lost, err = store.GetSeatMember(ctx, tenantID, members[2], after)
	require.NoError(t, err)
	require.Equal(t, pro.Losing[0].Fallback, lost.Seat, "what the member holds afterwards is what the preview said")
}
