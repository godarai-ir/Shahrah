package proxy

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Delpak/shahrah/services/control-plane/internal/billing"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

func TestWriteBillingEstimateErrorRejectsUnsupportedServiceTier(t *testing.T) {
	app := fiber.New()
	app.Post("/v1/responses", func(c *fiber.Ctx) error {
		return writeBillingEstimateError(c, &billing.UnsupportedServiceTierError{
			Provider: "openai",
			Model:    "gpt-5.4-nano",
			Tier:     "priority",
		})
	})

	req := httptest.NewRequest(
		http.MethodPost,
		"/v1/responses",
		http.NoBody,
	)
	resp, err := app.Test(req)
	require.NoError(t, err)

	defer resp.Body.Close()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var body struct {
		Error struct {
			Message string `json:"message"`
			Type    string `json:"type"`
		} `json:"error"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	require.Equal(
		t,
		"Model gpt-5.4-nano does not support service_tier \"priority\"",
		body.Error.Message,
	)
	require.Equal(t, "invalid_request_error", body.Error.Type)
}

func TestWriteBillingEstimateErrorKeepsPricingFailuresDistinct(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		message string
	}{
		{
			name:    "snapshot",
			err:     billing.ErrPricingUnavailable,
			message: "Provider pricing snapshot is unavailable or stale",
		},
		{
			name:    "model",
			err:     billing.ErrPricingModelUnavailable,
			message: "Provider pricing is unavailable for the requested model",
		},
		{
			name:    "rate",
			err:     billing.ErrPricingRateUnavailable,
			message: "Provider pricing is incomplete for the requested configuration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			app.Get("/", func(c *fiber.Ctx) error {
				return writeBillingEstimateError(c, tt.err)
			})

			resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/", http.NoBody))
			require.NoError(t, err)

			defer resp.Body.Close()

			require.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)

			var body struct {
				Error struct {
					Message string `json:"message"`
					Type    string `json:"type"`
				} `json:"error"`
			}
			require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
			require.Equal(t, tt.message, body.Error.Message)
			require.Equal(t, "pricing_unavailable", body.Error.Type)
		})
	}
}
