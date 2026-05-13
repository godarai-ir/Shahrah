package admin

import (
	"context"
	"encoding/json"
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

func setupTraceV2TestApp(h *Handlers) *fiber.App {
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals(auth.AuthContextKey{}, &platform.AuthContext{Role: "team_admin"})
		c.Locals(tenant.ContextKey{}, &platform.Tenant{ID: "tenant_a", Slug: "a", Status: "active"})
		return c.Next()
	})
	app.Get("/api/v2/traces", h.ListRequestTracesV2)
	app.Get("/api/v2/traces/facets", h.GetTraceFacetsV2)
	app.Get("/api/v2/traces/:traceID", h.GetRequestTraceV2)
	app.Get("/api/v2/traces/:traceID/timeline", h.ListRequestTraceTimelineV2)
	app.Get("/api/v2/traces/:traceID/attempts", h.ListRequestTraceAttemptsV2)
	return app
}

func TestGetTraceFacetsV2_UsesTenantContextIsolation(t *testing.T) {
	mock := &traceStoreMock{
		getFacetsV2Fn: func(_ context.Context, params platform.ListRequestTracesV2Params) (*platform.TraceFacetsV2, error) {
			assert.Equal(t, "tenant_a", params.TenantID)
			return &platform.TraceFacetsV2{
				Providers: []platform.TraceFacetBucketV2{{Value: "openai", Count: 7}},
				Models:    []platform.TraceFacetBucketV2{{Value: "gpt-4o-mini", Count: 5}},
			}, nil
		},
	}
	app := setupTraceV2TestApp(&Handlers{traces: mock})
	req := httptest.NewRequest("GET", "/api/v2/traces/facets", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var payload platform.TraceFacetsV2Response
	err = json.NewDecoder(resp.Body).Decode(&payload)
	assert.NoError(t, err)
	assert.Len(t, payload.Data.Providers, 1)
	assert.Equal(t, "openai", payload.Data.Providers[0].Value)
}

func TestListRequestTracesV2_UsesTenantContextIsolation(t *testing.T) {
	mock := &traceStoreMock{
		listV2Fn: func(_ context.Context, params platform.ListRequestTracesV2Params) ([]platform.RequestTraceV2, int, error) {
			assert.Equal(t, "tenant_a", params.TenantID)
			assert.Equal(t, "openai", params.Provider)
			assert.Equal(t, "req_123", params.RequestID)
			assert.Equal(t, "provider_error", params.ErrorCode)
			return []platform.RequestTraceV2{{
				TraceID:    "trace_1",
				TenantID:   "tenant_a",
				Route:      "/v1/chat/completions",
				StatusCode: 200,
				StartedAt:  time.Now().UTC(),
				DurationMS: 55,
			}}, 1, nil
		},
	}
	app := setupTraceV2TestApp(&Handlers{traces: mock})

	req := httptest.NewRequest("GET", "/api/v2/traces?provider=openai&request_id=req_123&error_code=provider_error", nil)
	req.Header.Set("x-shahrah-tenant", "tenant_b")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var payload platform.ListRequestTracesV2Response
	err = json.NewDecoder(resp.Body).Decode(&payload)
	assert.NoError(t, err)
	assert.Len(t, payload.Data, 1)
	assert.Equal(t, "trace_1", payload.Data[0].TraceID)
}

func TestGetRequestTraceV2_NotFound(t *testing.T) {
	mock := &traceStoreMock{
		getV2Fn: func(_ context.Context, tenantID string, traceID string) (*platform.RequestTraceDetailV2, error) {
			assert.Equal(t, "tenant_a", tenantID)
			assert.Equal(t, "trace_missing", traceID)
			return nil, pgx.ErrNoRows
		},
	}
	app := setupTraceV2TestApp(&Handlers{traces: mock})

	req := httptest.NewRequest("GET", "/api/v2/traces/trace_missing", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

func TestListRequestTraceAttemptsV2_ReturnsTenantScopedData(t *testing.T) {
	mock := &traceStoreMock{
		listAttemptsV2Fn: func(_ context.Context, tenantID string, traceID string) ([]platform.TraceAttemptV2, error) {
			assert.Equal(t, "tenant_a", tenantID)
			assert.Equal(t, "trace_1", traceID)
			return []platform.TraceAttemptV2{{
				AttemptNo:  1,
				TraceID:    "trace_1",
				Provider:   "openai",
				Model:      "gpt-4o-mini",
				Status:     "ok",
				StartedAt:  time.Now().UTC(),
				DurationMS: 120,
			}}, nil
		},
	}
	app := setupTraceV2TestApp(&Handlers{traces: mock})

	req := httptest.NewRequest("GET", "/api/v2/traces/trace_1/attempts", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var payload platform.TraceAttemptsV2Response
	err = json.NewDecoder(resp.Body).Decode(&payload)
	assert.NoError(t, err)
	assert.Len(t, payload.Data, 1)
	assert.Equal(t, "openai", payload.Data[0].Provider)
}
