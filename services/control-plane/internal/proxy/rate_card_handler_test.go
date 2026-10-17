package proxy

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/credits"
	"github.com/gofiber/fiber/v2"
)

type fakeRateCardBuilder struct{ seen credits.RateCard }

func (f *fakeRateCardBuilder) CreditRateCard(card credits.RateCard) (platform.CreditRateCard, string) {
	f.seen = card

	return platform.CreditRateCard{
		Object:  platform.CreditRateCardObject,
		Version: card.Label(),
		Unit:    platform.CreditUnit,
		Models:  []platform.CreditRateCardModel{{Model: "anthropic/claude-sonnet-5", TypicalMessageCredits: "4.50"}},
	}, "rev-abc"
}

type rateCardResult struct {
	status int
	header http.Header
	body   []byte
}

func doRateCardRequest(t *testing.T, app *fiber.App, ifNoneMatch string) rateCardResult {
	t.Helper()

	req := httptest.NewRequest(fiber.MethodGet, "/v1/rate-card", http.NoBody)
	if ifNoneMatch != "" {
		req.Header.Set(fiber.HeaderIfNoneMatch, ifNoneMatch)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	if err := resp.Body.Close(); err != nil {
		t.Fatalf("close body: %v", err)
	}

	return rateCardResult{status: resp.StatusCode, header: resp.Header, body: body}
}

func TestRateCardHandlerServesCacheableCard(t *testing.T) {
	t.Parallel()

	builder := &fakeRateCardBuilder{}
	app := fiber.New()
	app.Get("/v1/rate-card", NewRateCardHandler(builder, staticRateCards{card: credits.RateCard{Version: 7}}))

	result := doRateCardRequest(t, app, "")
	if result.status != fiber.StatusOK {
		t.Fatalf("status = %d, want 200", result.status)
	}

	if got := result.header.Get(fiber.HeaderETag); got != `"rev-abc"` {
		t.Fatalf("ETag = %q", got)
	}

	if got := result.header.Get(fiber.HeaderCacheControl); got != rateCardCacheControl {
		t.Fatalf("Cache-Control = %q", got)
	}

	var card platform.CreditRateCard
	if err := json.Unmarshal(result.body, &card); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if card.Version != "rc-7" || builder.seen.Version != 7 {
		t.Fatalf("version = %s (builder saw %d), want rc-7", card.Version, builder.seen.Version)
	}
}

func TestRateCardHandlerNotModified(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Get("/v1/rate-card", NewRateCardHandler(&fakeRateCardBuilder{}, nil))

	for _, header := range []string{`"rev-abc"`, `W/"rev-abc"`, `"other", "rev-abc"`, "*"} {
		result := doRateCardRequest(t, app, header)
		if result.status != fiber.StatusNotModified || len(result.body) != 0 {
			t.Fatalf("If-None-Match %s: status = %d body = %q, want empty 304", header, result.status, result.body)
		}
	}

	if result := doRateCardRequest(t, app, `"stale"`); result.status != fiber.StatusOK {
		t.Fatalf("stale etag status = %d, want 200", result.status)
	}
}

func TestCreditsUsedHeaderAndTelemetry(t *testing.T) {
	t.Parallel()

	cost := platform.UsageCostBreakdown{
		TotalCostUSD: 0.009,
		Components: []platform.UsageCostComponent{
			{Dimension: "input_tokens", CostUSD: "0.004"},
			{Dimension: "output_tokens", CostUSD: "0.005"},
		},
	}

	if got := creditsUsed(nil, cost); got != "" {
		t.Fatalf("creditsUsed without surface = %q, want empty", got)
	}

	if got := creditsUsed(staticRateCards{card: credits.IdentityRateCard()}, cost); got != "4.50" {
		t.Fatalf("creditsUsed = %q, want 4.50", got)
	}

	app := fiber.New()
	app.Use(CreditSurfaceMiddleware(staticRateCards{card: credits.IdentityRateCard()}))
	app.Get("/v1/test", func(c *fiber.Ctx) error {
		setCostHeaders(c, cost)

		return c.SendStatus(fiber.StatusOK)
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/v1/test", http.NoBody))
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}

	if err := resp.Body.Close(); err != nil {
		t.Fatalf("close body: %v", err)
	}

	if got := resp.Header.Get("x-godarai-credits-used"); got != "4.50" {
		t.Fatalf("x-godarai-credits-used = %q, want 4.50", got)
	}

	if got := resp.Header.Get("x-godarai-cost-irr"); got != "" {
		t.Fatalf("x-godarai-cost-irr = %q, want absent: the IRR header was removed", got)
	}
}

func TestCostHeadersWithoutCreditSurfaceAreUnchanged(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Get("/v1/test", func(c *fiber.Ctx) error {
		setCostHeaders(c, platform.UsageCostBreakdown{TotalCostUSD: 0.009})

		return c.SendStatus(fiber.StatusOK)
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/v1/test", http.NoBody))
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}

	if err := resp.Body.Close(); err != nil {
		t.Fatalf("close body: %v", err)
	}

	if got := resp.Header.Get("x-godarai-credits-used"); got != "" {
		t.Fatalf("x-godarai-credits-used = %q with shadow disabled, want absent", got)
	}
}
