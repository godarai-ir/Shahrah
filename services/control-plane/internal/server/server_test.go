package server

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/admin"
	controlauth "github.com/Delpak/shahrah/services/control-plane/internal/auth"
	"github.com/Delpak/shahrah/services/control-plane/internal/config"
	tenantmw "github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
	fiberlogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
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
	defer closeResponseBody(t, resp)

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

func TestOperationalLoggerSkipsFastSuccessfulRequests(t *testing.T) {
	t.Parallel()

	core, observed := observer.New(zap.InfoLevel)
	app := fiber.New()
	app.Use(newOperationalLoggerMiddleware(config.OperationalLogSettings{
		Enabled:           true,
		QueueSize:         8,
		SlowRequestMS:     1000,
		SuccessSampleRate: 0,
	}, zap.New(core)))
	app.Get("/ok", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/ok", http.NoBody))
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	require.Equal(t, fiber.StatusNoContent, resp.StatusCode)
	time.Sleep(20 * time.Millisecond)
	assert.Equal(t, 0, observed.Len())
}

func TestOperationalLoggerKeepsErrorsAndSlowRequests(t *testing.T) {
	t.Parallel()

	core, observed := observer.New(zap.InfoLevel)
	app := fiber.New()
	app.Use(newOperationalLoggerMiddleware(config.OperationalLogSettings{
		Enabled:           true,
		QueueSize:         8,
		SlowRequestMS:     1,
		SuccessSampleRate: 0,
	}, zap.New(core)))
	app.Get("/slow", func(c *fiber.Ctx) error {
		time.Sleep(5 * time.Millisecond)
		return c.SendStatus(fiber.StatusOK)
	})
	app.Get("/fail", func(c *fiber.Ctx) error {
		return fiber.NewError(fiber.StatusServiceUnavailable, "provider unavailable")
	})

	slowResp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/slow", http.NoBody))
	defer closeResponseBody(t, slowResp)

	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, slowResp.StatusCode)

	failResp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/fail", http.NoBody))
	defer closeResponseBody(t, failResp)

	require.NoError(t, err)
	require.Equal(t, fiber.StatusServiceUnavailable, failResp.StatusCode)

	require.Eventually(t, func() bool {
		return observed.Len() >= 2
	}, time.Second, 10*time.Millisecond)

	entries := observed.All()
	messages := make([]string, 0, len(entries))

	for _, entry := range entries {
		messages = append(messages, entry.Message)
	}

	assert.Contains(t, messages, "http request")
	assert.Contains(t, messages, "http request failed")
}

func TestNewFiberAppExposesPrometheusMetrics(t *testing.T) {
	t.Parallel()

	app := NewFiberApp(&config.Config{}, Dependencies{})

	healthResp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/health", http.NoBody))
	defer closeResponseBody(t, healthResp)

	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, healthResp.StatusCode)

	metricsResp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/metrics", http.NoBody))
	defer closeResponseBody(t, metricsResp)

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

func TestReadyIncludesClickHouseSubsystemReadiness(t *testing.T) {
	t.Parallel()

	app := NewFiberApp(&config.Config{}, Dependencies{
		ClickHouseHealth: fakeClickHouseHealthReporter{
			report: map[string]interface{}{
				"connected": true,
				"traces":    false,
				"metrics":   true,
				"logs":      true,
			},
		},
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/ready", http.NoBody))
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	payload := string(body)
	assert.Contains(t, payload, `"clickhouse"`)
	assert.Contains(t, payload, `"connected":true`)
	assert.Contains(t, payload, `"traces":false`)
	assert.Contains(t, payload, `"metrics":true`)
	assert.Contains(t, payload, `"logs":true`)
}

type fakeClickHouseHealthReporter struct {
	report map[string]interface{}
}

func (r fakeClickHouseHealthReporter) ClickHouseReadinessReport() map[string]interface{} {
	return r.report
}

func TestNewFiberAppPrometheusLabelsRemainStableAcrossRequests(t *testing.T) {
	t.Parallel()

	app := NewFiberApp(&config.Config{}, Dependencies{})
	app.Post("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	for range 5 {
		resp, err := app.Test(httptest.NewRequest(fiber.MethodPost, "/test", http.NoBody))
		require.NoError(t, err)
		require.Equal(t, fiber.StatusNoContent, resp.StatusCode)
		closeResponseBody(t, resp)
	}

	metricsResp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/metrics", http.NoBody))
	defer closeResponseBody(t, metricsResp)

	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, metricsResp.StatusCode)

	body, err := io.ReadAll(metricsResp.Body)
	require.NoError(t, err)

	payload := string(body)
	assert.Contains(t, payload, `shahrah_control_plane_http_requests_total{method="POST",route="/test",status_code="204"} 5`)
	assert.NotContains(t, payload, `method="POS"`)
}

func TestNewFiberAppExposesPprofOnlyWhenEnabled(t *testing.T) {
	t.Parallel()

	disabledApp := NewFiberApp(&config.Config{}, Dependencies{})

	disabledResp, err := disabledApp.Test(httptest.NewRequest(fiber.MethodGet, "/debug/pprof/goroutine?debug=1", http.NoBody))
	defer closeResponseBody(t, disabledResp)

	require.NoError(t, err)
	require.Equal(t, fiber.StatusNotFound, disabledResp.StatusCode)

	enabledApp := NewFiberApp(&config.Config{
		Profiling: config.ProfilingSettings{EnablePprof: true},
	}, Dependencies{})

	enabledResp, err := enabledApp.Test(httptest.NewRequest(fiber.MethodGet, "/debug/pprof/goroutine?debug=1", http.NoBody))
	defer closeResponseBody(t, enabledResp)

	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, enabledResp.StatusCode)
}

func TestNewFiberAppRegistersPromptRoutes(t *testing.T) {
	t.Parallel()

	app := NewFiberApp(&config.Config{}, Dependencies{
		AdminHandlers: admin.New(admin.Dependencies{
			Keys:           publicLoginDiscoveryStore{},
			LoginDiscovery: publicLoginDiscoveryStore{},
		}),
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/api/v1/prompts", http.NoBody))
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	require.NotEqual(t, fiber.StatusNotFound, resp.StatusCode)
}

func TestTenantAdminPATRevokeAllRouteUsesSpecificHandler(t *testing.T) {
	t.Parallel()

	app := fiber.New(fiber.Config{ErrorHandler: customErrorHandler})
	app.Use(func(c *fiber.Ctx) error {
		c.Locals(tenantmw.ContextKey{}, &platform.Tenant{ID: "tenant-1", Slug: "tenant1", Domain: "tenant1.localhost", Status: "active"})
		return c.Next()
	})
	registerTenantAdminRoutes(app, Dependencies{
		AdminHandlers: admin.New(admin.Dependencies{Keys: patRouteStore{}}),
	})

	req := httptest.NewRequest(fiber.MethodDelete, "/api/v1/personal-access-tokens/revoke/all", http.NoBody)
	req.Header.Set("x-godarai-user-id", "user-1")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.JSONEq(t, `{"revoked":7}`, string(body))
}

func TestPlatformAdminAuthMiddlewareRequiresMasterKey(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Get("/platform", controlauth.PlatformAdminAuthMiddleware("master-secret", nil), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	missing, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/platform", http.NoBody))
	defer closeResponseBody(t, missing)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, missing.StatusCode)

	request := httptest.NewRequest(fiber.MethodGet, "/platform", http.NoBody)
	request.Header.Set("Authorization", "Bearer master-secret")

	allowed, err := app.Test(request)
	defer closeResponseBody(t, allowed)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusNoContent, allowed.StatusCode)
}

type platformServiceAccountStoreStub struct {
	credential *controlauth.PlatformServiceAccountCredential
	err        error
}

func (s platformServiceAccountStoreStub) GetPlatformServiceAccountByTokenHash(context.Context, string) (*controlauth.PlatformServiceAccountCredential, error) {
	return s.credential, s.err
}

func TestPlatformAdminAuthMiddlewareEnforcesServiceAccountScopes(t *testing.T) {
	t.Parallel()

	store := platformServiceAccountStoreStub{credential: &controlauth.PlatformServiceAccountCredential{
		CredentialID: "credential-1",
		AccountName:  "support-admin",
		Scopes:       []string{controlauth.PlatformScopeRead},
	}}
	app := fiber.New()
	app.Use(controlauth.PlatformAdminAuthMiddleware("master-secret", store))
	app.Get("/api/v1/platform/features", func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusNoContent) })
	app.Post("/api/v1/platform/commercial-plans", func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusNoContent) })

	readRequest := httptest.NewRequest(fiber.MethodGet, "/api/v1/platform/features", http.NoBody)
	readRequest.Header.Set("Authorization", "Bearer service-token")

	readResponse, err := app.Test(readRequest)
	defer closeResponseBody(t, readResponse)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusNoContent, readResponse.StatusCode)

	writeRequest := httptest.NewRequest(fiber.MethodPost, "/api/v1/platform/commercial-plans", http.NoBody)
	writeRequest.Header.Set("Authorization", "Bearer service-token")

	writeResponse, err := app.Test(writeRequest)
	defer closeResponseBody(t, writeResponse)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, writeResponse.StatusCode)
}

func TestPlatformProviderCatalogReconcileRouteIsRegistered(t *testing.T) {
	t.Parallel()

	app := NewFiberApp(&config.Config{
		GeneralSettings: config.GeneralSettings{
			MasterKey: "master-secret",
		},
	}, Dependencies{
		AdminHandlers: admin.New(admin.Dependencies{}),
	})

	request := httptest.NewRequest(fiber.MethodPost, "/api/v1/platform/provider-catalog/reconcile", http.NoBody)
	request.Header.Set("Authorization", "Bearer master-secret")

	response, err := app.Test(request)
	defer closeResponseBody(t, response)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusNotImplemented, response.StatusCode)
}

func TestLoginDiscoveryRoutesArePublic(t *testing.T) {
	t.Parallel()

	store := publicLoginDiscoveryStore{}
	app := NewFiberApp(&config.Config{}, Dependencies{
		AdminHandlers:  admin.New(admin.Dependencies{Keys: store, LoginDiscovery: store}),
		TenantResolver: tenantmw.NewMiddlewareWithCache(tenantLookupAlwaysFails{}, nil, tenantmw.MiddlewareConfig{}),
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/api/v1/login-discovery/tenant/tenant1", http.NoBody))
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), `"exists":true`)
	assert.Contains(t, string(body), `"slug":"tenant1"`)
}

type tenantLookupAlwaysFails struct{}

func (tenantLookupAlwaysFails) GetTenantByHost(context.Context, string) (*platform.Tenant, error) {
	return nil, errors.New("tenant lookup should not run for public login discovery")
}

type publicLoginDiscoveryStore struct{}

type patRouteStore struct {
	publicLoginDiscoveryStore
}

func (patRouteStore) RevokePAT(context.Context, string, string, string, string) (bool, error) {
	return false, errors.New("single PAT revoke handler should not match revoke/all")
}

func (patRouteStore) RevokeAllPATs(context.Context, string, string, string) (int64, error) {
	return 7, nil
}

func (publicLoginDiscoveryStore) GetTenantBySlug(context.Context, string) (*platform.Tenant, error) {
	return &platform.Tenant{
		ID:        "tenant-1",
		Slug:      "tenant1",
		Domain:    "tenant1.localhost",
		Status:    "active",
		CreatedAt: time.Now(),
	}, nil
}

func (publicLoginDiscoveryStore) ListTenantsByAdminIdentifier(context.Context, string, string) ([]platform.Tenant, error) {
	return []platform.Tenant{}, nil
}

func (publicLoginDiscoveryStore) CreateKey(context.Context, platform.CreateKeyParams) (*platform.CreatedKey, error) {
	return nil, errors.New("not implemented")
}

func (publicLoginDiscoveryStore) ListKeys(context.Context, string) ([]platform.CreatedKey, error) {
	return nil, errors.New("not implemented")
}

func (publicLoginDiscoveryStore) ListAllowedModelFQNs(context.Context, string, string, string) ([]string, error) {
	return nil, errors.New("not implemented")
}

func (publicLoginDiscoveryStore) ListPATs(context.Context, string, string, string, int, int) ([]platform.CreatedKey, int, error) {
	return nil, 0, errors.New("not implemented")
}

func (publicLoginDiscoveryStore) PATExists(context.Context, string, string, string) (bool, error) {
	return false, errors.New("not implemented")
}

func (publicLoginDiscoveryStore) RevokePAT(context.Context, string, string, string, string) (bool, error) {
	return false, errors.New("not implemented")
}

func (publicLoginDiscoveryStore) RevokeAllPATs(context.Context, string, string, string) (int64, error) {
	return 0, errors.New("not implemented")
}

func (publicLoginDiscoveryStore) ListVATTokens(context.Context, string, string, string, int, int) ([]platform.CreatedKey, int, error) {
	return nil, 0, errors.New("not implemented")
}

func (publicLoginDiscoveryStore) UpdateVATToken(context.Context, string, string, string, platform.UpdateVATTokenRequest) (*platform.CreatedKey, error) {
	return nil, errors.New("not implemented")
}

func (publicLoginDiscoveryStore) RevokeVATToken(context.Context, string, string, string, string, string) (bool, error) {
	return false, errors.New("not implemented")
}

func (publicLoginDiscoveryStore) RotateVATToken(context.Context, string, string, string, string, string) (*platform.CreatedKey, error) {
	return nil, errors.New("not implemented")
}
