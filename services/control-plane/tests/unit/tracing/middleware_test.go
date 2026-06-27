package tracing_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/apiheaders"
	"github.com/Delpak/shahrah/services/control-plane/internal/observability/tracing"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMiddlewarePropagatesTraceID(t *testing.T) {
	app := fiber.New()
	q := tracing.NewQueue(16)
	s := tracing.NewSampler(1.0, 1500)
	app.Use(tracing.Middleware(q, s, 1500))
	app.Get("/x", func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})

	req := httptest.NewRequest("GET", "/x", http.NoBody)
	req.Header.Set(apiheaders.HeaderTraceID, "trace_abc")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "trace_abc", resp.Header.Get(apiheaders.HeaderTraceID))
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

	req := httptest.NewRequest("GET", "/x", http.NoBody)

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	assert.NoError(t, err)
	assert.NotEmpty(t, resp.Header.Get(apiheaders.HeaderTraceID))

	req = httptest.NewRequest("GET", "/x", http.NoBody)
	req.Header.Set(apiheaders.LegacyHeaderTraceID, "trace_legacy")

	resp, err = app.Test(req)
	defer closeResponseBody(t, resp)

	assert.NoError(t, err)
	assert.Equal(t, "trace_legacy", resp.Header.Get(apiheaders.HeaderTraceID))
}

func TestMiddleware_DropsRootAndStepWhenUnsampled(t *testing.T) {
	app := fiber.New()
	q := tracing.NewQueue(16)
	s := tracing.NewSampler(0.0, 1500)
	app.Use(tracing.Middleware(q, s, 1500))
	app.Get("/x", func(c *fiber.Ctx) error {
		start := time.Now().Add(-5 * time.Millisecond)
		tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "quota.check", "internal", start, fiber.StatusOK, map[string]interface{}{"estimated_tokens": 42})

		return c.SendStatus(200)
	})

	req := httptest.NewRequest("GET", "/x", http.NoBody)

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, 0, q.Len())
}

func TestMiddleware_EnqueuesRootAndStepsForErrorRequest(t *testing.T) {
	app := fiber.New()
	q := tracing.NewQueue(16)
	s := tracing.NewSampler(0.0, 1500)
	app.Use(tracing.Middleware(q, s, 1500))
	app.Get("/x", func(c *fiber.Ctx) error {
		start := time.Now().Add(-5 * time.Millisecond)
		tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "provider.call", "client", start, fiber.StatusInternalServerError, map[string]interface{}{"provider": "openai"})

		return c.SendStatus(fiber.StatusInternalServerError)
	})

	req := httptest.NewRequest("GET", "/x", http.NoBody)

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	assert.Equal(t, 2, q.Len())

	batch := q.DequeueBatch(10)
	require.Len(t, batch, 2)
	assert.Equal(t, "gateway.request", batch[0].Name)
	assert.Equal(t, "server", batch[0].Kind)
	assert.Equal(t, batch[0].SpanID, batch[1].ParentSpanID)
	assert.Equal(t, "provider.call", batch[1].Name)
}

func TestRecordStepSpan_DefaultsParentToRootSpan(t *testing.T) {
	app := fiber.New()
	q := tracing.NewQueue(16)
	s := tracing.NewSampler(1.0, 1500)
	app.Use(tracing.Middleware(q, s, 1500))
	app.Get("/x", func(c *fiber.Ctx) error {
		start := time.Now().Add(-5 * time.Millisecond)
		tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "router.select_model", "internal", start, fiber.StatusOK, nil)

		return c.SendStatus(200)
	})

	req := httptest.NewRequest("GET", "/x", http.NoBody)

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	batch := q.DequeueBatch(10)
	require.Len(t, batch, 2)
	root := batch[0]
	step := batch[1]
	assert.Equal(t, root.SpanID, step.ParentSpanID)
}
