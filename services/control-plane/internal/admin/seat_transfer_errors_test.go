package admin

import (
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	billingdomain "github.com/Delpak/shahrah/services/control-plane/internal/domain/billing"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/seats"
)

func TestSeatTransferErrorsAreCoded(t *testing.T) {
	for _, tc := range []struct {
		err    error
		status int
		code   string
	}{
		{seats.ErrOrderNotFound, fiber.StatusNotFound, "seat_order_not_found"},
		{seats.ErrOrderNotPayable, fiber.StatusConflict, "seat_order_not_payable"},
		{billingdomain.ErrBankTransferAlreadyPending, fiber.StatusConflict, "bank_transfer_already_pending"},
	} {
		status, code := creditPackErrorStatus(tc.err)
		require.Equal(t, tc.status, status)
		require.Equal(t, tc.code, code)
	}
}
