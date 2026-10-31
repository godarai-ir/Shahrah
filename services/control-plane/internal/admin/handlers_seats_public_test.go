package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

// seatTypesStoreFake lists seat types; the other seat methods are not used.
type seatTypesStoreFake struct {
	seatStore
}

func (seatTypesStoreFake) ListSeatTypes(context.Context) ([]platform.SeatTypeOffer, error) {
	custom := int64(100_000_000)

	return []platform.SeatTypeOffer{
		{Type: "free", MonthlyCredits: 500, Active: true},
		{Type: "pro", MonthlyCredits: 2000, MonthlyPriceIRR: 22_500_000, SuggestedMonthlyPriceIRR: 21_600_000, DiscountPercent: "10.00", Active: true},
		{Type: "max", MonthlyCredits: 10000, MonthlyPriceIRR: custom, CustomMonthlyPriceIRR: &custom, Active: true},
		{Type: "team", MonthlyCredits: 1, Active: false},
	}, nil
}

func TestListPublicSeatTypesShowsActivePricesOnly(t *testing.T) {
	h := &Handlers{billingState: billingState{seats: seatTypesStoreFake{}}}
	app := fiber.New()
	app.Get("/api/v1/seat-types", h.ListPublicSeatTypes)

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/api/v1/seat-types", http.NoBody))
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	var body map[string]json.RawMessage

	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))

	var types []platform.PublicSeatType

	require.NoError(t, json.Unmarshal(body["data"], &types))
	require.Equal(t, []platform.PublicSeatType{
		{Type: "free", MonthlyCredits: 500},
		{Type: "pro", MonthlyCredits: 2000, MonthlyPriceIRR: 22_500_000},
		{Type: "max", MonthlyCredits: 10000, MonthlyPriceIRR: 100_000_000},
	}, types)
	require.NotContains(t, string(body["data"]), "discount", "discounts stay internal")
}
