package tracing_test

import (
	"net/http/httptest"
	"testing"

	"github.com/Delpak/shahrah/services/control-plane/internal/observability/tracing"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestMiddlewarePropagatesTraceID(t *testing.T) {
	app := fiber.New()
	q := tracing.NewQueue(16)
	s := tracing.NewSampler(1.0, 1500)
	app.Use(tracing.Middleware(q, s, 1500))
	app.Get("/x", func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})

	req := httptest.NewRequest("GET", "/x", nil)
	req.Header.Set("x-shahrah-trace-id", "trace_abc")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "trace_abc", resp.Header.Get("x-shahrah-trace-id"))
	assert.Equal(t, 1, q.Len())
}

func TestMiddlewareGeneratesTraceIDWhenMissing(t *testing.T) {
	app := fiber.New()
	q := tracing.NewQueue(16)
	s := tracing.NewSampler(1.0, 1500)
	app.Use(tracing.Middleware(q, s, 1500))
	app.Get("/x", func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})

	req := httptest.NewRequest("GET", "/x", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.Header.Get("x-shahrah-trace-id"))
}
