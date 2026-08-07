package admin

import (
	"bytes"
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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type metricsStoreMock struct {
	fn       func(ctx context.Context, tenantID string, req platform.MetricsQueryRequest) ([]map[string]interface{}, error)
	facetsFn func(ctx context.Context, tenantID string) (*platform.ModelMetricsFacets, error)
}

func (m *metricsStoreMock) QueryModelMetrics(ctx context.Context, tenantID string, req platform.MetricsQueryRequest) ([]map[string]interface{}, error) {
	if m.fn != nil {
		return m.fn(ctx, tenantID, req)
	}

	return []map[string]interface{}{}, nil
}

func (m *metricsStoreMock) ListModelMetricsFacets(ctx context.Context, tenantID string) (*platform.ModelMetricsFacets, error) {
	if m.facetsFn != nil {
		return m.facetsFn(ctx, tenantID)
	}

	return &platform.ModelMetricsFacets{}, nil
}

func setupMetricsTestApp(h *Handlers) *fiber.App {
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals(auth.AuthContextKey{}, &platform.AuthContext{Role: "team_admin"})
		c.Locals(tenant.ContextKey{}, &platform.Tenant{ID: "tenant_a", Slug: "a", Status: "active"})

		return c.Next()
	})
	app.Post("/api/v1/metrics/query", h.QueryModelMetrics)
	app.Get("/api/v1/metrics/facets", h.ListModelMetricsFacets)

	return app
}

func TestQueryModelMetrics_UsesTenantContext(t *testing.T) {
	mock := &metricsStoreMock{
		fn: func(_ context.Context, tenantID string, req platform.MetricsQueryRequest) ([]map[string]interface{}, error) {
			assert.Equal(t, "tenant_a", tenantID)
			assert.Equal(t, "modelMetrics", req.Datasource)
			assert.Equal(t, "distribution", req.Type)

			return []map[string]interface{}{{"countModelName": 1}}, nil
		},
	}
	h := &Handlers{observabilityState: observabilityState{metrics: mock}}
	app := setupMetricsTestApp(h)

	body := map[string]interface{}{
		"startTs":    time.Now().Add(-time.Hour).UTC().Format(time.RFC3339),
		"endTs":      time.Now().UTC().Format(time.RFC3339),
		"datasource": "modelMetrics",
		"type":       "distribution",
		"aggregations": []map[string]string{
			{"type": "count", "column": "modelName"},
		},
	}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/v1/metrics/query", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-godarai-tenant", "tenant_b")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestQueryModelMetrics_InvalidDatasource(t *testing.T) {
	h := &Handlers{observabilityState: observabilityState{metrics: &metricsStoreMock{}}}
	app := setupMetricsTestApp(h)
	body := map[string]interface{}{
		"startTs":    time.Now().Add(-time.Hour).UTC().Format(time.RFC3339),
		"endTs":      time.Now().UTC().Format(time.RFC3339),
		"datasource": "other",
		"type":       "distribution",
	}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/v1/metrics/query", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestListModelMetricsFacets_UsesTenantContext(t *testing.T) {
	mock := &metricsStoreMock{
		facetsFn: func(_ context.Context, tenantID string) (*platform.ModelMetricsFacets, error) {
			assert.Equal(t, "tenant_a", tenantID)

			return &platform.ModelMetricsFacets{
				Models:    []string{"historical-model"},
				Providers: []string{"historical-provider"},
			}, nil
		},
	}
	h := &Handlers{observabilityState: observabilityState{metrics: mock}}
	app := setupMetricsTestApp(h)

	req := httptest.NewRequest("GET", "/api/v1/metrics/facets", http.NoBody)

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var body platform.ModelMetricsFacets
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.Equal(t, []string{"historical-model"}, body.Models)
	assert.Equal(t, []string{"historical-provider"}, body.Providers)
}
