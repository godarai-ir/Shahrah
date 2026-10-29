package admin

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/apperrors"
	billingdomain "github.com/Delpak/shahrah/services/control-plane/internal/domain/billing"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

// creditAdminStoreFake implements only what the tests call; the embedded nil
// interface panics on anything else, which would fail the test loudly.
type creditAdminStoreFake struct {
	creditAdminStore
	conversions int
	rollbacks   int
	fleetRuns   int
	setPrice    int64
	err         error
}

func (f *creditAdminStoreFake) SetCreditListPrice(_ context.Context, priceToman int64, _ platform.SupportActor, reason string, _ time.Time) (*platform.CreditPriceRevision, error) {
	if f.err != nil {
		return nil, f.err
	}

	f.setPrice = priceToman

	return &platform.CreditPriceRevision{PriceTomanPerCredit: priceToman, Status: "active", DecisionReason: reason}, nil
}

func (f *creditAdminStoreFake) ConvertTenantToCredits(_ context.Context, tenantID, reason string, _ platform.SupportActor) (*platform.CreditConversion, error) {
	if f.err != nil {
		return nil, f.err
	}

	f.conversions++

	return &platform.CreditConversion{TenantID: tenantID, Direction: "to_credit", Reason: reason}, nil
}

func (f *creditAdminStoreFake) RollbackCreditConversion(_ context.Context, tenantID, reason string, _ platform.SupportActor) (*platform.CreditConversion, error) {
	f.rollbacks++

	return &platform.CreditConversion{TenantID: tenantID, Direction: "to_irr", Reason: reason}, nil
}

func (f *creditAdminStoreFake) ConvertCreditFleet(context.Context, platform.CreditFleetConversionRequest, platform.SupportActor) (*platform.CreditFleetConversionResult, error) {
	f.fleetRuns++

	return &platform.CreditFleetConversionResult{}, nil
}

func (f *creditAdminStoreFake) GetTenantCreditOverview(_ context.Context, tenantID string) (*platform.TenantCreditOverview, error) {
	return &platform.TenantCreditOverview{TenantID: tenantID}, nil
}

func creditAdminApp(store *creditAdminStoreFake) *fiber.App {
	return creditAdminAppWithPacks(store, nil)
}

func creditAdminAppWithPacks(store *creditAdminStoreFake, packs *creditPackStoreFake) *fiber.App {
	h := &Handlers{billingState: billingState{creditAdmin: store}}
	if packs != nil {
		h.creditPacks = packs
	}

	app := fiber.New()
	app.Put("/tenants/:tenantID/credits/overage", h.UpdatePlatformTenantCreditOverage)
	app.Post("/tenants/:tenantID/credits/conversion", h.ConvertPlatformTenantToCredits)
	app.Post("/tenants/:tenantID/credits/conversion/rollback", h.RollbackPlatformTenantCredits)
	app.Post("/credits/conversions/fleet", h.ConvertPlatformCreditFleet)
	app.Post("/credits/price-revisions", h.SetPlatformCreditListPrice)
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

func TestConversionRoutesReachTheStore(t *testing.T) {
	store := &creditAdminStoreFake{}
	tenantPath := "/tenants/8f14e45f-ceea-467f-a9f0-5b6b7b8c9d0e/credits/conversion"
	body := `{"reason":"fleet cutover"}`

	require.Equal(t, fiber.StatusOK, doCreditAdmin(t, creditAdminApp(store), fiber.MethodPost, tenantPath, body))
	require.Equal(t, fiber.StatusOK, doCreditAdmin(t, creditAdminApp(store), fiber.MethodPost, "/credits/conversions/fleet", body))
	require.Equal(t, fiber.StatusOK, doCreditAdmin(t, creditAdminApp(store), fiber.MethodPost, tenantPath+"/rollback", body))
	require.Equal(t, 1, store.conversions)
	require.Equal(t, 1, store.fleetRuns)
	require.Equal(t, 1, store.rollbacks)

	require.Equal(t, fiber.StatusBadRequest, doCreditAdmin(t, creditAdminApp(store), fiber.MethodPost, "/tenants/not-a-uuid/credits/conversion", body))
}

func TestConversionErrorsAreCoded(t *testing.T) {
	tenantPath := "/tenants/8f14e45f-ceea-467f-a9f0-5b6b7b8c9d0e/credits/conversion"

	cases := map[error]int{
		apperrors.ErrNotFound:                        fiber.StatusNotFound,
		billingdomain.ErrCreditConversionNothingToDo: fiber.StatusConflict,
		billingdomain.ErrDecisionReasonRequired:      fiber.StatusBadRequest,
		billingdomain.ErrInsufficientWalletBalance:   fiber.StatusConflict,
	}

	for err, want := range cases {
		store := &creditAdminStoreFake{err: err}
		require.Equal(t, want, doCreditAdmin(t, creditAdminApp(store), fiber.MethodPost, tenantPath, `{"reason":"x"}`), err.Error())
	}
}

func TestCreditPriceDecisionsRejectMalformedIDs(t *testing.T) {
	status := doCreditAdmin(t, creditAdminApp(&creditAdminStoreFake{}), fiber.MethodPost, "/credits/price-revisions/1%27%3B--/approve", `{"reason":"x"}`)
	require.Equal(t, fiber.StatusNotFound, status)
	require.NotEqual(t, http.StatusInternalServerError, status)
}

func TestSupportSetsTheListPriceByHand(t *testing.T) {
	store := &creditAdminStoreFake{}

	status := doCreditAdmin(t, creditAdminApp(store), fiber.MethodPost, "/credits/price-revisions", `{"price_toman_per_credit":1125,"reason":"round seat prices"}`)
	require.Equal(t, fiber.StatusCreated, status)
	require.Equal(t, int64(1125), store.setPrice)

	invalid := &creditAdminStoreFake{err: billingdomain.ErrInvalidCreditPrice}
	require.Equal(t, fiber.StatusBadRequest, doCreditAdmin(t, creditAdminApp(invalid), fiber.MethodPost, "/credits/price-revisions", `{"price_toman_per_credit":0,"reason":"x"}`))
}

func TestSupportOverageUpdateChangesOnlyOverage(t *testing.T) {
	packs := &creditPackStoreFake{}
	app := creditAdminAppWithPacks(&creditAdminStoreFake{}, packs)

	status := doCreditAdmin(t, app, http.MethodPut, "/tenants/7f0c2f3e-56a4-4b8e-9a59-8f0e0e0f0f0f/credits/overage",
		`{"overage_enabled":true,"overage_cap_credits":"500","seat_upgrade_requests_enabled":false}`)
	require.Equal(t, http.StatusOK, status)
	require.NotNil(t, packs.settingsReq)
	require.True(t, *packs.settingsReq.OverageEnabled)
	require.Equal(t, "500", *packs.settingsReq.OverageCapCredits)
	require.Nil(t, packs.settingsReq.SeatUpgradeRequestsEnabled)

	require.Equal(t, http.StatusBadRequest, doCreditAdmin(t, app, http.MethodPut, "/tenants/not-a-uuid/credits/overage", `{"overage_enabled":true}`))
}
