package admin

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

type reviewWalletStoreFake struct{ walletStore }

func TestRejectPlatformBankTransferRequiresReason(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	handlers := &Handlers{billingState: billingState{wallets: reviewWalletStoreFake{}}}
	app.Post("/tenants/:tenantID/wallet/bank-transfers/:requestID/reject", handlers.RejectPlatformBankTransfer)

	req := httptest.NewRequest(http.MethodPost, "/tenants/t1/wallet/bank-transfers/r1/reject", strings.NewReader(`{"reason":"  "}`))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	resp, err := app.Test(req)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
