package server

import (
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGatewayBackpressureRejectsWhenFull(t *testing.T) {
	t.Setenv("GATEWAY_MAX_IN_FLIGHT_REQUESTS", "1")

	app := fiber.New()
	app.Use(newGatewayBackpressureMiddleware())

	entered := make(chan struct{})
	release := make(chan struct{})
	firstDone := make(chan error, 1)

	app.Get("/protected", func(c *fiber.Ctx) error {
		select {
		case <-entered:
		default:
			close(entered)
		}
		<-release
		return c.SendStatus(fiber.StatusNoContent)
	})

	go func() {
		req, err := http.NewRequest(http.MethodGet, "http://example.com/protected", http.NoBody)
		if err != nil {
			firstDone <- err
			return
		}
		resp, err := app.Test(req, -1)
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
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

	req, err := http.NewRequest(http.MethodGet, "http://example.com/protected", http.NoBody)
	require.NoError(t, err)
	resp, err := app.Test(req, -1)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)

	close(release)
	assert.NoError(t, <-firstDone)
}

func TestGatewayBackpressureExemptsHealthAndMetrics(t *testing.T) {
	t.Setenv("GATEWAY_MAX_IN_FLIGHT_REQUESTS", "1")

	app := fiber.New()
	app.Use(newGatewayBackpressureMiddleware())

	entered := make(chan struct{})
	release := make(chan struct{})
	firstDone := make(chan error, 1)

	app.Get("/protected", func(c *fiber.Ctx) error {
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
		req, err := http.NewRequest(http.MethodGet, "http://example.com/protected", http.NoBody)
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
		defer closeResponseBody(t, resp)

		require.NoError(t, err)
		assert.Equal(t, fiber.StatusNoContent, resp.StatusCode)
	}

	close(release)
	assert.NoError(t, <-firstDone)
}

func TestGatewayBackpressureReleasesAfterHandlerError(t *testing.T) {
	t.Setenv("GATEWAY_MAX_IN_FLIGHT_REQUESTS", "1")

	app := fiber.New()
	app.Use(newGatewayBackpressureMiddleware())
	app.Get("/fail", func(c *fiber.Ctx) error {
		return fiber.NewError(fiber.StatusInternalServerError, "boom")
	})
	app.Get("/ok", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	failReq, err := http.NewRequest(http.MethodGet, "http://example.com/fail", http.NoBody)
	require.NoError(t, err)
	failResp, err := app.Test(failReq, -1)
	defer closeResponseBody(t, failResp)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, failResp.StatusCode)

	okReq, err := http.NewRequest(http.MethodGet, "http://example.com/ok", http.NoBody)
	require.NoError(t, err)
	okResp, err := app.Test(okReq, -1)
	defer closeResponseBody(t, okResp)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusNoContent, okResp.StatusCode)
}
