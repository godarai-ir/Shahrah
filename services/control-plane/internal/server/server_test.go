package server

import (
	"bytes"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Delpak/shahrah/services/control-plane/internal/config"
	"github.com/gofiber/fiber/v2"
	fiberlogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPLoggerRedactsSensitiveHeadersAndOmitsBodies(t *testing.T) {
	t.Parallel()

	var logBuf bytes.Buffer
	app := fiber.New()
	app.Use(fiberlogger.New(buildHTTPLoggerConfig(&logBuf)))
	app.Post("/test", func(c *fiber.Ctx) error {
		c.Set("Set-Cookie", "session=super-secret")
		return c.SendString("super-secret-response")
	})

	req := httptest.NewRequest(fiber.MethodPost, "/test", strings.NewReader(`{"apiKey":"super-secret-request"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer super-secret-token")
	req.Header.Set("X-API-Key", "super-secret-api-key")

	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	logLine := logBuf.String()
	assert.NotContains(t, logLine, "super-secret-request")
	assert.NotContains(t, logLine, "super-secret-response")
	assert.NotContains(t, logLine, "super-secret-token")
	assert.NotContains(t, logLine, "super-secret-api-key")
	assert.NotContains(t, logLine, "session=super-secret")
	assert.Contains(t, logLine, "Authorization=[REDACTED]")
	assert.Contains(t, logLine, "X-Api-Key=[REDACTED]")
	assert.Contains(t, logLine, "Set-Cookie=[REDACTED]")
	assert.Contains(t, logLine, "Content-Type=application/json")
}

func TestFormatHeaderLogValueEscapesAndRedacts(t *testing.T) {
	t.Parallel()

	out := formatHeaderLogValue(map[string][]string{
		"Authorization": {"Bearer secret"},
		"X-Test":        {"line1\nline2", `quoted"value`},
	})

	assert.Contains(t, out, "Authorization=[REDACTED]")
	assert.Contains(t, out, `X-Test=line1\\nline2`)
	assert.Contains(t, out, `X-Test=quoted\\\"value`)
	assert.NotContains(t, out, "Bearer secret")
}

func TestNewFiberAppExposesPrometheusMetrics(t *testing.T) {
	t.Parallel()

	app := NewFiberApp(&config.Config{}, Dependencies{})

	healthResp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/health", nil))
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, healthResp.StatusCode)

	metricsResp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/metrics", nil))
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, metricsResp.StatusCode)

	body, err := io.ReadAll(metricsResp.Body)
	require.NoError(t, err)
	payload := string(body)
	assert.Contains(t, payload, "shahrah_control_plane_http_requests_total")
	assert.Contains(t, payload, `shahrah_control_plane_http_requests_total{method="GET",route="/health",status_code="200"} 1`)
	assert.Contains(t, payload, "shahrah_control_plane_http_request_duration_seconds")
	assert.Contains(t, payload, "shahrah_control_plane_http_requests_in_flight")
}
