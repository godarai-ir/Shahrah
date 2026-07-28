package proxy

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Delpak/shahrah/services/control-plane/internal/router"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestWriteRoutingDependencyUnavailable(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		handled, err := writeRoutingDependencyUnavailable(
			c,
			zap.NewNop(),
			fmt.Errorf("route failed: %w", router.ErrDependencyUnavailable),
			"openai:gpt:gpt-4o-mini",
			"tenant-1",
		)
		require.True(t, handled)

		return err
	})

	response, err := app.Test(httptest.NewRequest("GET", "/", http.NoBody))
	require.NoError(t, err)

	defer response.Body.Close()

	require.Equal(t, fiber.StatusServiceUnavailable, response.StatusCode)

	var body struct {
		Error struct {
			Type string `json:"type"`
			Code string `json:"code"`
		} `json:"error"`
	}
	require.NoError(t, json.NewDecoder(response.Body).Decode(&body))
	require.Equal(t, "dependency_unavailable", body.Error.Type)
	require.Equal(t, "dependency_unavailable", body.Error.Code)
}

func TestWriteRoutingPricingUnresolved(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		handled, err := writeRoutingDependencyUnavailable(
			c,
			zap.NewNop(),
			fmt.Errorf("route failed: %w", router.ErrDeploymentBillingMetadataUnresolved),
			"legacy-model",
			"tenant-1",
		)
		require.True(t, handled)
		return err
	})

	response, err := app.Test(httptest.NewRequest("GET", "/", http.NoBody))
	require.NoError(t, err)
	defer response.Body.Close()
	require.Equal(t, fiber.StatusUnprocessableEntity, response.StatusCode)

	var body struct {
		Error struct {
			Type string `json:"type"`
			Code string `json:"code"`
		} `json:"error"`
	}
	require.NoError(t, json.NewDecoder(response.Body).Decode(&body))
	require.Equal(t, "pricing_unresolved", body.Error.Type)
	require.Equal(t, "pricing_unresolved", body.Error.Code)
}
