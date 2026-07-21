package server

import (
	"net/http"
	"sync"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type backpressureMetricsStub struct {
	mu          sync.Mutex
	inFlight    int
	rejected    int
	maxInFlight int
}

func (m *backpressureMetricsStub) IncBackpressureInFlight() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.inFlight++
	if m.inFlight > m.maxInFlight {
		m.maxInFlight = m.inFlight
	}
}

func (m *backpressureMetricsStub) DecBackpressureInFlight() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.inFlight--
}

func (m *backpressureMetricsStub) IncBackpressureRejected() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.rejected++
}

func (m *backpressureMetricsStub) snapshot() (int, int, int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.inFlight, m.rejected, m.maxInFlight
}

func TestGatewayBackpressureRejectsWhenFull(t *testing.T) {
	t.Setenv("GATEWAY_MAX_IN_FLIGHT_REQUESTS", "1")

	app := fiber.New()
	metrics := &backpressureMetricsStub{}
	app.Use(newGatewayBackpressureMiddleware(metrics))

	entered := make(chan struct{})
	release := make(chan struct{})
	firstDone := make(chan error, 1)

	app.Get("/v1/protected", func(c *fiber.Ctx) error {
		select {
		case <-entered:
		default:
			close(entered)
		}

		<-release

		return c.SendStatus(fiber.StatusNoContent)
	})

	go func() {
		req, err := http.NewRequest(http.MethodGet, "http://example.com/v1/protected", http.NoBody)
		if err != nil {
			firstDone <- err
			return
		}

		resp, err := app.Test(req, -1)
		if resp != nil && resp.Body != nil {
			if closeErr := resp.Body.Close(); err == nil {
				err = closeErr
			}
		}

		if err != nil {
			firstDone <- err
			return
		}

		if resp.StatusCode != fiber.StatusNoContent {
			firstDone <- fiber.NewError(resp.StatusCode)
			return
		}

		firstDone <- nil
	}()

	<-entered

	req, err := http.NewRequest(http.MethodGet, "http://example.com/v1/protected", http.NoBody)
	require.NoError(t, err)

	resp, err := app.Test(req, -1)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)

	close(release)
	assert.NoError(t, <-firstDone)

	inFlight, rejected, maxInFlight := metrics.snapshot()
	assert.Equal(t, 0, inFlight)
	assert.Equal(t, 1, rejected)
	assert.Equal(t, 1, maxInFlight)
}

func TestGatewayBackpressureExemptsHealthAndMetrics(t *testing.T) {
	t.Setenv("GATEWAY_MAX_IN_FLIGHT_REQUESTS", "1")

	app := fiber.New()
	app.Use(newGatewayBackpressureMiddleware())

	entered := make(chan struct{})
	release := make(chan struct{})
	firstDone := make(chan error, 1)

	app.Get("/v1/protected", func(c *fiber.Ctx) error {
		close(entered)
		<-release

		return c.SendStatus(fiber.StatusNoContent)
	})
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})
	app.Get("/metrics", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	go func() {
		req, err := http.NewRequest(http.MethodGet, "http://example.com/v1/protected", http.NoBody)
		if err != nil {
			firstDone <- err
			return
		}

		resp, err := app.Test(req, -1)
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}

		firstDone <- err
	}()

	<-entered

	for _, path := range []string{"/health", "/metrics"} {
		req, err := http.NewRequest(http.MethodGet, "http://example.com"+path, http.NoBody)
		require.NoError(t, err)

		resp, err := app.Test(req, -1)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusNoContent, resp.StatusCode)
		closeResponseBody(t, resp)
	}

	close(release)
	assert.NoError(t, <-firstDone)
}

func TestGatewayBackpressureReleasesAfterHandlerError(t *testing.T) {
	t.Setenv("GATEWAY_MAX_IN_FLIGHT_REQUESTS", "1")

	app := fiber.New()
	app.Use(newGatewayBackpressureMiddleware())
	app.Get("/v1/fail", func(c *fiber.Ctx) error {
		return fiber.NewError(fiber.StatusInternalServerError, "boom")
	})
	app.Get("/v1/ok", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	failReq, err := http.NewRequest(http.MethodGet, "http://example.com/v1/fail", http.NoBody)
	require.NoError(t, err)

	failResp, err := app.Test(failReq, -1)
	defer closeResponseBody(t, failResp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, failResp.StatusCode)

	okReq, err := http.NewRequest(http.MethodGet, "http://example.com/v1/ok", http.NoBody)
	require.NoError(t, err)

	okResp, err := app.Test(okReq, -1)
	defer closeResponseBody(t, okResp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusNoContent, okResp.StatusCode)
}

func TestGatewayBackpressureDoesNotThrottleAdminAuth(t *testing.T) {
	t.Setenv("GATEWAY_MAX_IN_FLIGHT_REQUESTS", "1")

	app := fiber.New()
	app.Use(newGatewayBackpressureMiddleware())

	entered := make(chan struct{})
	release := make(chan struct{})
	firstDone := make(chan error, 1)

	app.Get("/v1/protected", func(c *fiber.Ctx) error {
		close(entered)
		<-release

		return c.SendStatus(fiber.StatusNoContent)
	})
	app.Get("/api/v1/auth/session/:id", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	go func() {
		req, err := http.NewRequest(http.MethodGet, "http://example.com/v1/protected", http.NoBody)
		if err != nil {
			firstDone <- err
			return
		}

		resp, err := app.Test(req, -1)
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}

		firstDone <- err
	}()

	<-entered

	req, err := http.NewRequest(http.MethodGet, "http://example.com/api/v1/auth/session/sess-1", http.NoBody)
	require.NoError(t, err)

	resp, err := app.Test(req, -1)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusNoContent, resp.StatusCode)

	close(release)
	assert.NoError(t, <-firstDone)
}
