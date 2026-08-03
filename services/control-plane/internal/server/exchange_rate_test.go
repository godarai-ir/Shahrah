package server

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/config"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

type exchangeRateReaderStub struct {
	rate *platform.ExchangeRate
	err  error
}

func (s exchangeRateReaderStub) GetLatestExchangeRate(context.Context) (*platform.ExchangeRate, error) {
	return s.rate, s.err
}

func TestCurrentExchangeRateEndpointReturnsPublicFreshNobitexQuote(t *testing.T) {
	now := time.Now().UTC()
	app := NewFiberApp(&config.Config{}, Dependencies{
		ExchangeRates: exchangeRateReaderStub{rate: &platform.ExchangeRate{
			ID:        "rate-1",
			USDtoIRR:  1_883_500,
			Source:    "nobitex:USDTIRT",
			FetchedAt: now.Add(-5 * time.Minute),
		}},
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/api/v1/exchange-rates/current", nil))
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
	require.Equal(t, publicExchangeRateCache, resp.Header.Get("Cache-Control"))

	var body currentExchangeRateResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	require.Equal(t, "USD", body.Base)
	require.Equal(t, "IRR", body.Quote)
	require.Equal(t, 1_883_500.0, body.Rate)
	require.Equal(t, "nobitex:USDTIRT", body.Source)
	require.Equal(t, "USDTIRT", body.MarketSymbol)
	require.Equal(t, "USDT_AS_USD_PROXY", body.PricingBasis)
	require.Equal(t, "fresh", body.Status)
}

func TestCurrentExchangeRateHandlerMarksOldSnapshotStale(t *testing.T) {
	now := time.Date(2026, time.August, 3, 12, 0, 0, 0, time.UTC)
	app := fiber.New()
	app.Get("/", currentExchangeRateHandlerWithClock(exchangeRateReaderStub{rate: &platform.ExchangeRate{
		USDtoIRR:  1_883_500,
		Source:    "nobitex:USDTIRT",
		FetchedAt: now.Add(-16 * time.Minute),
	}}, func() time.Time { return now }))

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	var body currentExchangeRateResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	require.Equal(t, "stale", body.Status)
}

func TestCurrentExchangeRateEndpointRejectsUnavailableSnapshots(t *testing.T) {
	now := time.Now().UTC()
	tests := []struct {
		name   string
		reader ExchangeRateReader
	}{
		{name: "read failure", reader: exchangeRateReaderStub{err: errors.New("database unavailable")}},
		{name: "missing", reader: exchangeRateReaderStub{}},
		{name: "seed fallback", reader: exchangeRateReaderStub{rate: &platform.ExchangeRate{USDtoIRR: 1_850_000, Source: "seed", FetchedAt: now}}},
		{name: "invalid symbol", reader: exchangeRateReaderStub{rate: &platform.ExchangeRate{USDtoIRR: 1_850_000, Source: "nobitex:BTCIRT", FetchedAt: now}}},
		{name: "zero", reader: exchangeRateReaderStub{rate: &platform.ExchangeRate{Source: "nobitex:USDTIRT", FetchedAt: now}}},
		{name: "not a number", reader: exchangeRateReaderStub{rate: &platform.ExchangeRate{USDtoIRR: math.NaN(), Source: "nobitex:USDTIRT", FetchedAt: now}}},
		{name: "unbounded", reader: exchangeRateReaderStub{rate: &platform.ExchangeRate{USDtoIRR: publicExchangeRateMax + 1, Source: "nobitex:USDTIRT", FetchedAt: now}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := NewFiberApp(&config.Config{}, Dependencies{ExchangeRates: tt.reader})
			resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/api/v1/exchange-rates/current", nil))
			require.NoError(t, err)
			require.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)
		})
	}
}
