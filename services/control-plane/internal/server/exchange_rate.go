package server

import (
	"math"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

const (
	publicExchangeRateFreshFor = 15 * time.Minute
	publicExchangeRateMax       = 1_000_000_000
	publicExchangeRateCache     = "public, max-age=60, stale-while-revalidate=600"
)

type currentExchangeRateResponse struct {
	Base          string    `json:"base"`
	Quote         string    `json:"quote"`
	Rate          float64   `json:"rate"`
	Source        string    `json:"source"`
	MarketSymbol  string    `json:"marketSymbol"`
	PricingBasis  string    `json:"pricingBasis"`
	FetchedAt     time.Time `json:"fetchedAt"`
	Status        string    `json:"status"`
}

func currentExchangeRateHandler(reader ExchangeRateReader) fiber.Handler {
	return currentExchangeRateHandlerWithClock(reader, time.Now)
}

func currentExchangeRateHandlerWithClock(reader ExchangeRateReader, now func() time.Time) fiber.Handler {
	return func(c *fiber.Ctx) error {
		rate, err := reader.GetLatestExchangeRate(c.Context())
		if err != nil || rate == nil || !isPublicExchangeRateValid(rate.USDtoIRR, rate.FetchedAt) {
			return fiber.NewError(fiber.StatusServiceUnavailable, "current exchange rate is unavailable")
		}

		symbol, ok := nobitexMarketSymbol(rate.Source)
		if !ok {
			return fiber.NewError(fiber.StatusServiceUnavailable, "current exchange rate is unavailable")
		}

		status := "fresh"
		if now().UTC().Sub(rate.FetchedAt.UTC()) > publicExchangeRateFreshFor {
			status = "stale"
		}

		c.Set(fiber.HeaderCacheControl, publicExchangeRateCache)
		return c.JSON(currentExchangeRateResponse{
			Base:         "USD",
			Quote:        "IRR",
			Rate:         rate.USDtoIRR,
			Source:       rate.Source,
			MarketSymbol: symbol,
			PricingBasis: "USDT_AS_USD_PROXY",
			FetchedAt:    rate.FetchedAt.UTC(),
			Status:       status,
		})
	}
}

func isPublicExchangeRateValid(rate float64, fetchedAt time.Time) bool {
	return rate > 0 && rate <= publicExchangeRateMax && !math.IsNaN(rate) && !math.IsInf(rate, 0) && !fetchedAt.IsZero()
}

func nobitexMarketSymbol(source string) (string, bool) {
	provider, symbol, ok := strings.Cut(strings.TrimSpace(source), ":")
	if !ok || !strings.EqualFold(strings.TrimSpace(provider), "nobitex") {
		return "", false
	}

	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	if !strings.HasPrefix(symbol, "USDT") || len(symbol) <= len("USDT") {
		return "", false
	}

	return symbol, true
}
