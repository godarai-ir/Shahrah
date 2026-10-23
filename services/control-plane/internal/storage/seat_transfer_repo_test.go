package storage

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	billingdomain "github.com/Delpak/shahrah/services/control-plane/internal/domain/billing"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/seats"
)

func TestSeatTransferRules(t *testing.T) {
	now := time.Date(2026, 10, 21, 12, 0, 0, 0, time.UTC)
	order := &platform.SeatOrder{
		Status: platform.CreditPackOrderAwaitingTransfer, PayableIRR: 7_664_516,
		CreatedAt: now.Add(-time.Hour), PeriodEnd: now.Add(24 * time.Hour),
	}

	require.NoError(t, checkSeatTransferSubmit(order, order.PayableIRR, now))
	require.NoError(t, checkSeatTransferApprove(order, order.PayableIRR, now))

	require.ErrorIs(t, checkSeatTransferSubmit(order, order.PayableIRR-1, now), billingdomain.ErrCreditPackTransferMismatch)
	require.ErrorIs(t, checkSeatTransferApprove(order, order.PayableIRR-1, now), seats.ErrOrderNotPayable)

	expired := *order
	expired.CreatedAt = now.Add(-seatTransferHold - time.Minute)
	require.ErrorIs(t, checkSeatTransferSubmit(&expired, expired.PayableIRR, now), seats.ErrOrderNotPayable)

	// Finance approves after the paid period ended: the seats would be worthless.
	lapsed := *order
	lapsed.PeriodEnd = now.Add(-time.Minute)
	require.ErrorIs(t, checkSeatTransferApprove(&lapsed, lapsed.PayableIRR, now), seats.ErrOrderNotPayable)

	done := *order
	done.Status = platform.CreditPackOrderCompleted
	require.ErrorIs(t, checkSeatTransferSubmit(&done, done.PayableIRR, now), seats.ErrOrderNotPayable)
	require.ErrorIs(t, checkSeatTransferApprove(&done, done.PayableIRR, now), seats.ErrOrderNotPayable)
}
