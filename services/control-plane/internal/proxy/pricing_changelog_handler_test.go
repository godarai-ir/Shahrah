package proxy

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

type changelogStoreFake struct{ err error }

func (f changelogStoreFake) ListPricingChangelog(context.Context) (*platform.PricingChangelog, error) {
	if f.err != nil {
		return nil, f.err
	}

	return &platform.PricingChangelog{
		RateCards:    []platform.PricingChangelogRateCard{{Version: 1, Label: "rc-1", Status: "published", EffectiveFrom: time.Unix(0, 0).UTC()}},
		CreditPrices: []platform.PricingChangelogPrice{{PriceTomanPerCredit: 1000, Status: "active"}},
	}, nil
}

func TestPricingChangelogHandlerIsPublicAndCacheable(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Get("/v1/pricing/changelog", NewPricingChangelogHandler(changelogStoreFake{}, nil))

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/v1/pricing/changelog", http.NoBody))
	require.NoError(t, err)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())

	require.Equal(t, fiber.StatusOK, resp.StatusCode)
	require.Equal(t, pricingChangelogCacheControl, resp.Header.Get(fiber.HeaderCacheControl))

	var changelog platform.PricingChangelog
	require.NoError(t, json.Unmarshal(body, &changelog))
	require.Equal(t, "rc-1", changelog.RateCards[0].Label)
	require.NotContains(t, strings.ToLower(string(body)), "usd", "the changelog never exposes dollar figures")
}

func TestPricingChangelogHandlerHidesStoreErrors(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Get("/v1/pricing/changelog", NewPricingChangelogHandler(changelogStoreFake{err: errors.New("pg: connection refused")}, nil))

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/v1/pricing/changelog", http.NoBody))
	require.NoError(t, err)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())

	require.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)
	require.NotContains(t, string(body), "connection refused")
}
