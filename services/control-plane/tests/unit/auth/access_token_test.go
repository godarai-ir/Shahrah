package auth_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	"github.com/Delpak/shahrah/services/control-plane/internal/oidc"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const testAccessToken = "eyJhbGciOiJFUzI1NiJ9.eyJzdWIiOiJ1c2VyIn0.signature"

type stubAccessTokens struct {
	resolved   *platform.AuthContext
	err        error
	gotTenant  string
	gotToken   string
	callsCount int
}

func (s *stubAccessTokens) AuthenticateAccessToken(_ context.Context, tenantID, token string) (*platform.AuthContext, error) {
	s.callsCount++
	s.gotTenant = tenantID
	s.gotToken = token

	return s.resolved, s.err
}

func newAccessTokenTestApp(tokens auth.AccessTokenAuthenticator, seen **platform.AuthContext) *fiber.App {
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals(tenant.ContextKey{}, &platform.Tenant{ID: "tenant-1", Slug: "acme"})
		return c.Next()
	})
	app.Use(auth.NewAuthMiddlewareWithOptions(auth.NewInMemoryKeyStore(), "test-master-key", zap.NewNop(), auth.AuthMiddlewareOptions{
		AccessTokens: tokens,
	}).Middleware())

	handler := func(c *fiber.Ctx) error {
		*seen = auth.GetAuthInfo(c)
		return c.SendStatus(fiber.StatusOK)
	}
	app.Post("/v1/chat/completions", handler)
	app.Get("/api/v1/users", handler)
	app.Get("/oidc/jwks", handler)

	return app
}

func doBearer(t *testing.T, app *fiber.App, method, path, token string) int {
	t.Helper()

	req, err := http.NewRequest(method, path, http.NoBody)
	require.NoError(t, err)

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := app.Test(req)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	return resp.StatusCode
}

func TestAccessTokenAuthenticatesGatewayRequestsForTheResolvedTenant(t *testing.T) {
	tokens := &stubAccessTokens{resolved: &platform.AuthContext{TenantID: "tenant-1", UserID: "user-1", KeyType: oidc.KeyType}}

	var seen *platform.AuthContext

	app := newAccessTokenTestApp(tokens, &seen)

	require.Equal(t, fiber.StatusOK, doBearer(t, app, http.MethodPost, "/v1/chat/completions", testAccessToken))
	require.Equal(t, "tenant-1", tokens.gotTenant)
	require.Equal(t, testAccessToken, tokens.gotToken)
	require.NotNil(t, seen)
	require.Equal(t, "user-1", seen.UserID)
}

func TestAccessTokenCannotCallManagementAPIs(t *testing.T) {
	tokens := &stubAccessTokens{resolved: &platform.AuthContext{TenantID: "tenant-1", UserID: "user-1"}}

	var seen *platform.AuthContext

	app := newAccessTokenTestApp(tokens, &seen)

	require.Equal(t, fiber.StatusForbidden, doBearer(t, app, http.MethodGet, "/api/v1/users", testAccessToken))
	require.Equal(t, fiber.StatusForbidden, doBearer(t, app, http.MethodGet, "/API/v1/users", testAccessToken))
	require.Zero(t, tokens.callsCount)
	require.Nil(t, seen)
}

func TestInvalidAccessTokenIsUnauthorizedAndOutagesAreServerErrors(t *testing.T) {
	var seen *platform.AuthContext

	invalid := &stubAccessTokens{err: &oidc.Error{Code: oidc.CodeInvalidToken, Status: http.StatusUnauthorized}}
	require.Equal(t, fiber.StatusUnauthorized, doBearer(t, newAccessTokenTestApp(invalid, &seen), http.MethodPost, "/v1/chat/completions", testAccessToken))

	outage := &stubAccessTokens{err: errors.New("database unavailable")}
	require.Equal(t, fiber.StatusInternalServerError, doBearer(t, newAccessTokenTestApp(outage, &seen), http.MethodPost, "/v1/chat/completions", testAccessToken))
	require.Nil(t, seen)
}

func TestOIDCProviderRoutesSkipAPIKeyAuthentication(t *testing.T) {
	var seen *platform.AuthContext

	app := newAccessTokenTestApp(&stubAccessTokens{}, &seen)

	require.Equal(t, fiber.StatusOK, doBearer(t, app, http.MethodGet, "/oidc/jwks", ""))
}

func TestAPIKeysStillUseTheKeyStoreWhenAccessTokensAreEnabled(t *testing.T) {
	tokens := &stubAccessTokens{}

	var seen *platform.AuthContext

	app := newAccessTokenTestApp(tokens, &seen)

	require.Equal(t, fiber.StatusUnauthorized, doBearer(t, app, http.MethodPost, "/v1/chat/completions", "sk-not-a-jwt"))
	require.Zero(t, tokens.callsCount)
}
