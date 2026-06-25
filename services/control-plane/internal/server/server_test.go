package server

import (
	"bytes"
	"context"
	"errors"
	"io"
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

func TestNewFiberAppRegistersPromptRoutes(t *testing.T) {
	t.Parallel()

	app := NewFiberApp(&config.Config{}, Dependencies{
		AdminHandlers: admin.New(publicLoginDiscoveryStore{}, nil, nil),
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/api/v1/prompts", nil))
	require.NoError(t, err)
	require.NotEqual(t, fiber.StatusNotFound, resp.StatusCode)
}

func TestPlatformAdminAuthMiddlewareRequiresMasterKey(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	app.Get("/platform", platformAdminAuthMiddleware("master-secret", nil), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	missing, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/platform", nil))
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, missing.StatusCode)

	request := httptest.NewRequest(fiber.MethodGet, "/platform", nil)
	request.Header.Set("Authorization", "Bearer master-secret")
	allowed, err := app.Test(request)
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
	app.Use(platformAdminAuthMiddleware("master-secret", store))
	app.Get("/api/v1/platform/features", func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusNoContent) })
	app.Post("/api/v1/platform/commercial-plans", func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusNoContent) })

	readRequest := httptest.NewRequest(fiber.MethodGet, "/api/v1/platform/features", nil)
	readRequest.Header.Set("Authorization", "Bearer service-token")
	readResponse, err := app.Test(readRequest)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusNoContent, readResponse.StatusCode)

	writeRequest := httptest.NewRequest(fiber.MethodPost, "/api/v1/platform/commercial-plans", nil)
	writeRequest.Header.Set("Authorization", "Bearer service-token")
	writeResponse, err := app.Test(writeRequest)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, writeResponse.StatusCode)
}

func TestLoginDiscoveryRoutesArePublic(t *testing.T) {
	t.Parallel()

	store := publicLoginDiscoveryStore{}
	app := NewFiberApp(&config.Config{}, Dependencies{
		AdminHandlers:  admin.New(store, nil, nil),
		TenantResolver: tenantmw.NewMiddleware(tenantLookupAlwaysFails{}, "", false),
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/api/v1/login-discovery/tenant/tenant1", nil))
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
