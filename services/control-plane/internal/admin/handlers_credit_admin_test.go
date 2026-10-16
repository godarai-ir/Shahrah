package admin

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

// creditAdminStoreFake implements only what the tests call; the embedded nil
// interface panics on anything else, which would fail the test loudly.
type creditAdminStoreFake struct {
	creditAdminStore
	unitChanges int
}

func (f *creditAdminStoreFake) SetTenantBillingUnitAudited(context.Context, string, platform.SetBillingUnitRequest, platform.SupportActor) error {
	f.unitChanges++

	return nil
}

func (f *creditAdminStoreFake) GetTenantCreditOverview(_ context.Context, tenantID string) (*platform.TenantCreditOverview, error) {
	return &platform.TenantCreditOverview{TenantID: tenantID, BillingUnit: "credit"}, nil
}

func creditAdminApp(store *creditAdminStoreFake, ledgerEnabled bool) *fiber.App {
	h := &Handlers{billingState: billingState{creditAdmin: store, creditLedgerEnabled: ledgerEnabled}}
	app := fiber.New()
	app.Put("/tenants/:tenantID/credits/billing-unit", h.SetPlatformTenantBillingUnit)
	app.Post("/credits/price-revisions/:revisionID/approve", h.ApprovePlatformCreditPriceRevision)

	return app
}

func doCreditAdmin(t *testing.T, app *fiber.App, method, path, body string) int {
	t.Helper()

	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	resp, err := app.Test(req)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())

	return resp.StatusCode
}

func TestBillingUnitChangeNeedsTheCreditLedgerFlag(t *testing.T) {
	store := &creditAdminStoreFake{}
	tenantPath := "/tenants/8f14e45f-ceea-467f-a9f0-5b6b7b8c9d0e/credits/billing-unit"
	body := `{"billing_unit":"credit","reason":"pilot"}`

	require.Equal(t, fiber.StatusConflict, doCreditAdmin(t, creditAdminApp(store, false), fiber.MethodPut, tenantPath, body))
	require.Zero(t, store.unitChanges)

	require.Equal(t, fiber.StatusOK, doCreditAdmin(t, creditAdminApp(store, true), fiber.MethodPut, tenantPath, body))
	require.Equal(t, 1, store.unitChanges)

	require.Equal(t, fiber.StatusBadRequest, doCreditAdmin(t, creditAdminApp(store, true), fiber.MethodPut, "/tenants/not-a-uuid/credits/billing-unit", body))
}

func TestCreditPriceDecisionsRejectMalformedIDs(t *testing.T) {
	status := doCreditAdmin(t, creditAdminApp(&creditAdminStoreFake{}, true), fiber.MethodPost, "/credits/price-revisions/1%27%3B--/approve", `{"reason":"x"}`)
	require.Equal(t, fiber.StatusNotFound, status)
	require.NotEqual(t, http.StatusInternalServerError, status)
}
