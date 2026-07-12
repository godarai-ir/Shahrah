package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
)

type traceStoreMock struct {
	getFn            func(ctx context.Context, tenantID string, traceID string) (*platform.RequestTrace, error)
	listSpansFn      func(ctx context.Context, tenantID string, traceID string) ([]platform.TraceSpan, error)
	listV2Fn         func(ctx context.Context, params platform.ListRequestTracesV2Params) ([]platform.RequestTraceV2, int, error)
	getV2Fn          func(ctx context.Context, tenantID string, traceID string) (*platform.RequestTraceDetailV2, error)
	listTimelineV2Fn func(ctx context.Context, tenantID string, traceID string) ([]platform.TraceSpan, error)
	listAttemptsV2Fn func(ctx context.Context, tenantID string, traceID string) ([]platform.TraceAttemptV2, error)
	getFacetsV2Fn    func(ctx context.Context, params platform.ListRequestTracesV2Params) (*platform.TraceFacetsV2, error)
}

func (m *traceStoreMock) ListRequestTraces(ctx context.Context, params platform.ListRequestTracesParams) ([]platform.RequestTrace, int, error) {
	return nil, 0, nil
}

func (m *traceStoreMock) GetRequestTrace(ctx context.Context, tenantID string, traceID string) (*platform.RequestTrace, error) {
	if m.getFn != nil {
		return m.getFn(ctx, tenantID, traceID)
	}

	return nil, pgx.ErrNoRows
}

func (m *traceStoreMock) ListTraceSpans(ctx context.Context, tenantID string, traceID string) ([]platform.TraceSpan, error) {
	if m.listSpansFn != nil {
		return m.listSpansFn(ctx, tenantID, traceID)
	}

	return nil, pgx.ErrNoRows
}

func (m *traceStoreMock) ListRequestTracesV2(ctx context.Context, params platform.ListRequestTracesV2Params) ([]platform.RequestTraceV2, int, error) {
	if m.listV2Fn != nil {
		return m.listV2Fn(ctx, params)
	}

	return []platform.RequestTraceV2{}, 0, nil
}

func (m *traceStoreMock) GetRequestTraceV2(ctx context.Context, tenantID string, traceID string) (*platform.RequestTraceDetailV2, error) {
	if m.getV2Fn != nil {
		return m.getV2Fn(ctx, tenantID, traceID)
	}

	return nil, pgx.ErrNoRows
}

func (m *traceStoreMock) ListTraceTimelineV2(ctx context.Context, tenantID string, traceID string) ([]platform.TraceSpan, error) {
	if m.listTimelineV2Fn != nil {
		return m.listTimelineV2Fn(ctx, tenantID, traceID)
	}

	return nil, pgx.ErrNoRows
}

func (m *traceStoreMock) ListTraceAttemptsV2(ctx context.Context, tenantID string, traceID string) ([]platform.TraceAttemptV2, error) {
	if m.listAttemptsV2Fn != nil {
		return m.listAttemptsV2Fn(ctx, tenantID, traceID)
	}

	return nil, pgx.ErrNoRows
}

func (m *traceStoreMock) GetTraceFacetsV2(ctx context.Context, params platform.ListRequestTracesV2Params) (*platform.TraceFacetsV2, error) {
	if m.getFacetsV2Fn != nil {
		return m.getFacetsV2Fn(ctx, params)
	}

	return &platform.TraceFacetsV2{}, nil
}

func setupTraceTestApp(h *Handlers) *fiber.App {
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals(auth.AuthContextKey{}, &platform.AuthContext{Role: "team_admin"})
		c.Locals(tenant.ContextKey{}, &platform.Tenant{ID: "tenant_a", Slug: "a", Status: "active"})

		return c.Next()
	})
	app.Get("/api/v1/traces/:traceID", h.GetRequestTrace)
	app.Get("/api/v1/traces/:traceID/spans", h.ListRequestTraceSpans)

	return app
}

func TestGetRequestTrace_UsesTenantContextIsolation(t *testing.T) {
	mock := &traceStoreMock{
		getFn: func(_ context.Context, tenantID string, traceID string) (*platform.RequestTrace, error) {
			assert.Equal(t, "tenant_a", tenantID)
			assert.Equal(t, "trace_1", traceID)

			return &platform.RequestTrace{
				TraceID:    "trace_1",
				TenantID:   "tenant_a",
				Route:      "/v1/chat/completions",
				StatusCode: 200,
				StartedAt:  time.Now().UTC(),
				DurationMS: 42,
				Sampled:    true,
			}, nil
		},
	}
	h := &Handlers{traces: mock}
	app := setupTraceTestApp(h)

	req := httptest.NewRequest("GET", "/api/v1/traces/trace_1", http.NoBody)
	// Attempted override should be ignored; only tenant context must be used.
	req.Header.Set("x-godarai-tenant", "tenant_b")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestGetRequestTrace_CrossTenantReturnsNotFound(t *testing.T) {
	mock := &traceStoreMock{
		getFn: func(_ context.Context, tenantID string, traceID string) (*platform.RequestTrace, error) {
			assert.Equal(t, "tenant_a", tenantID)
			assert.Equal(t, "foreign_trace", traceID)

			return nil, pgx.ErrNoRows
		},
	}
	h := &Handlers{traces: mock}
	app := setupTraceTestApp(h)

	req := httptest.NewRequest("GET", "/api/v1/traces/foreign_trace", http.NoBody)

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

func TestListTraceSpans_CrossTenantReturnsNotFound(t *testing.T) {
	mock := &traceStoreMock{
		listSpansFn: func(_ context.Context, tenantID string, traceID string) ([]platform.TraceSpan, error) {
			assert.Equal(t, "tenant_a", tenantID)
			assert.Equal(t, "foreign_trace", traceID)

			return nil, pgx.ErrNoRows
		},
	}
	h := &Handlers{traces: mock}
	app := setupTraceTestApp(h)

	req := httptest.NewRequest("GET", "/api/v1/traces/foreign_trace/spans", http.NoBody)

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

func TestListTraceSpans_ReturnsTenantScopedData(t *testing.T) {
	mock := &traceStoreMock{
		listSpansFn: func(_ context.Context, tenantID string, traceID string) ([]platform.TraceSpan, error) {
			assert.Equal(t, "tenant_a", tenantID)
			assert.Equal(t, "trace_1", traceID)

			return []platform.TraceSpan{{
				SpanID:       "span_1",
				TraceID:      "trace_1",
				Name:         "gateway.request",
				Kind:         "server",
				StartedAt:    time.Now().UTC(),
				DurationMS:   11,
				Status:       "ok",
				Attributes:   map[string]interface{}{"http.route": "/v1/chat/completions"},
				ParentSpanID: "",
			}}, nil
		},
	}
	h := &Handlers{traces: mock}
	app := setupTraceTestApp(h)

	req := httptest.NewRequest("GET", "/api/v1/traces/trace_1/spans", http.NoBody)

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var payload platform.TraceSpansResponse

	err = json.NewDecoder(resp.Body).Decode(&payload)
	assert.NoError(t, err)
	assert.Len(t, payload.Data, 1)
	assert.Equal(t, "trace_1", payload.Data[0].TraceID)
}
