package oidc

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/oidc"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

type stubService struct {
	tokenParams    oidc.TokenParams
	tokenErr       error
	authorizeErr   error
	authorizeURL   string
	acceptedSlug   string
	acceptedID     string
	acceptedSessID string
	createdTenant  string
}

func (s *stubService) Discovery(tenant oidc.Tenant) map[string]any {
	return map[string]any{"issuer": "https://" + tenant.Host}
}

func (s *stubService) JWKS() oidc.JSONWebKeySet { return oidc.JSONWebKeySet{} }

func (s *stubService) Authorize(context.Context, oidc.Tenant, oidc.AuthorizeParams) (*oidc.AuthorizeResult, error) {
	if s.authorizeErr != nil {
		return nil, s.authorizeErr
	}

	return &oidc.AuthorizeResult{RedirectURL: s.authorizeURL}, nil
}

func (s *stubService) Token(_ context.Context, _ oidc.Tenant, params oidc.TokenParams) (*oidc.TokenResponse, error) {
	s.tokenParams = params
	if s.tokenErr != nil {
		return nil, s.tokenErr
	}

	return &oidc.TokenResponse{AccessToken: "at", TokenType: "Bearer", ExpiresIn: 600}, nil
}

func (s *stubService) UserInfo(context.Context, oidc.Tenant, string) (map[string]any, error) {
	return map[string]any{"sub": "user"}, nil
}

func (s *stubService) Logout(context.Context, oidc.Tenant, oidc.LogoutParams) (string, error) {
	return "", nil
}

func (s *stubService) GetLoginRequest(context.Context, string, string) (*oidc.LoginRequest, error) {
	return &oidc.LoginRequest{ID: "req"}, nil
}

func (s *stubService) AcceptLogin(_ context.Context, slug, id, session string) (*oidc.LoginDecision, error) {
	s.acceptedSlug, s.acceptedID, s.acceptedSessID = slug, id, session

	return &oidc.LoginDecision{RedirectTo: "https://chat.example/cb?code=x"}, nil
}

func (s *stubService) RejectLogin(context.Context, string, string) (*oidc.LoginDecision, error) {
	return &oidc.LoginDecision{RedirectTo: "https://chat.example/cb?error=access_denied"}, nil
}

func (s *stubService) ListClients(context.Context, string) ([]oidc.Client, error) {
	return []oidc.Client{}, nil
}

func (s *stubService) CreateClient(_ context.Context, tenantID, _ string, request oidc.CreateClientRequest) (*oidc.ClientWithSecret, error) {
	s.createdTenant = tenantID

	if len(request.RedirectURIs) == 0 {
		return nil, oidc.ErrInvalidClientConfig
	}

	return &oidc.ClientWithSecret{Client: oidc.Client{ClientID: "gdc-1"}, ClientSecret: "gds_secret"}, nil
}

func (s *stubService) UpdateClient(context.Context, string, string, oidc.UpdateClientRequest) (*oidc.Client, error) {
	return nil, oidc.ErrClientNotFound
}

func (s *stubService) RotateClientSecret(context.Context, string, string) (*oidc.ClientWithSecret, error) {
	return nil, oidc.ErrClientNotFound
}

func (s *stubService) DeleteClient(context.Context, string, string) error { return nil }

func newTestApp(t *testing.T, service *stubService, role string) *fiber.App {
	t.Helper()

	routes, err := NewRoutes(RoutesDependencies{
		Service: service,
		Tenant: func(*fiber.Ctx) *platform.Tenant {
			return &platform.Tenant{ID: "tenant-1", Slug: "acme"}
		},
		AuthInfo: func(*fiber.Ctx) *platform.AuthContext {
			return &platform.AuthContext{Role: role, UserID: "admin-1"}
		},
		IsProxyAdmin: func(value string) bool { return value == "proxy_admin" },
	})
	require.NoError(t, err)

	app := fiber.New()
	routes.RegisterProvider(app)
	routes.RegisterAuth(app.Group("/api/v1/auth"))
	routes.RegisterTenant(app.Group("/api/v1"))

	return app
}

type result struct {
	StatusCode int
	Header     http.Header
}

func send(t *testing.T, app *fiber.App, req *http.Request) (result, map[string]any) {
	t.Helper()

	resp, err := app.Test(req)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	raw, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	body := map[string]any{}
	if strings.HasPrefix(resp.Header.Get(fiber.HeaderContentType), fiber.MIMEApplicationJSON) {
		require.NoError(t, json.Unmarshal(raw, &body))
	}

	return result{StatusCode: resp.StatusCode, Header: resp.Header}, body
}

func tokenRequest(form url.Values) *http.Request {
	req, _ := http.NewRequest(http.MethodPost, "http://acme.godar.example/oidc/token", strings.NewReader(form.Encode()))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationForm)

	return req
}

func TestTokenEndpointDecodesBasicClientCredentials(t *testing.T) {
	t.Parallel()

	service := &stubService{}
	app := newTestApp(t, service, "")

	req := tokenRequest(url.Values{"grant_type": {"authorization_code"}, "code": {"c"}, "code_verifier": {"v"}})
	credentials := url.QueryEscape("gdc-1") + ":" + url.QueryEscape("se:cr/et")
	req.Header.Set(fiber.HeaderAuthorization, "Basic "+base64.StdEncoding.EncodeToString([]byte(credentials)))

	resp, body := send(t, app, req)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "no-store", resp.Header.Get(fiber.HeaderCacheControl))
	require.Equal(t, "at", body["access_token"])
	require.Equal(t, "gdc-1", service.tokenParams.ClientID)
	require.Equal(t, "se:cr/et", service.tokenParams.ClientSecret)
	require.Equal(t, "v", service.tokenParams.CodeVerifier)
}

func TestTokenEndpointRejectsTwoAuthenticationMethods(t *testing.T) {
	t.Parallel()

	app := newTestApp(t, &stubService{}, "")

	req := tokenRequest(url.Values{"grant_type": {"authorization_code"}, "client_secret": {"x"}})
	req.Header.Set(fiber.HeaderAuthorization, "Basic "+base64.StdEncoding.EncodeToString([]byte("a:b")))

	resp, body := send(t, app, req)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	require.Equal(t, oidc.CodeInvalidRequest, body["error"])
}

func TestTokenEndpointMapsProtocolAndInternalErrors(t *testing.T) {
	t.Parallel()

	protocol := &stubService{tokenErr: &oidc.Error{Code: oidc.CodeInvalidGrant, Description: "expired", Status: http.StatusBadRequest}}
	resp, body := send(t, newTestApp(t, protocol, ""), tokenRequest(url.Values{"grant_type": {"refresh_token"}}))
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	require.Equal(t, oidc.CodeInvalidGrant, body["error"])
	require.Equal(t, "expired", body["error_description"])

	internal := &stubService{tokenErr: context.DeadlineExceeded}
	resp, body = send(t, newTestApp(t, internal, ""), tokenRequest(url.Values{"grant_type": {"refresh_token"}}))
	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	require.Equal(t, oidc.CodeServerError, body["error"])
	require.NotContains(t, body, "error_description", "internal failures must not leak detail")
}

func TestAuthorizeRedirectsOrShowsUnverifiableErrors(t *testing.T) {
	t.Parallel()

	ok := &stubService{authorizeURL: "https://www.acme.godar.example/login?oidc_request=r"}
	req, _ := http.NewRequest(http.MethodGet, "http://acme.godar.example/oidc/authorize?client_id=gdc-1", http.NoBody)
	resp, _ := send(t, newTestApp(t, ok, ""), req)
	require.Equal(t, http.StatusFound, resp.StatusCode)
	require.Equal(t, ok.authorizeURL, resp.Header.Get(fiber.HeaderLocation))

	bad := &stubService{authorizeErr: &oidc.Error{Code: oidc.CodeInvalidClient, Description: "unknown client", Status: http.StatusUnauthorized}}
	req, _ = http.NewRequest(http.MethodGet, "http://acme.godar.example/oidc/authorize?client_id=x", http.NoBody)
	resp, _ = send(t, newTestApp(t, bad, ""), req)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	require.Empty(t, resp.Header.Get(fiber.HeaderLocation))
}

func TestDiscoveryUsesTheRequestHost(t *testing.T) {
	t.Parallel()

	req, _ := http.NewRequest(http.MethodGet, "http://acme.godar.example:4000/.well-known/openid-configuration", http.NoBody)
	resp, body := send(t, newTestApp(t, &stubService{}, ""), req)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "https://acme.godar.example:4000", body["issuer"])
}

func TestLoginBridgeRequiresATrustedCaller(t *testing.T) {
	t.Parallel()

	body := `{"tenant_slug":"acme","session_id":"s-1"}`

	untrusted := &stubService{}
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/oidc/requests/r-1/accept", strings.NewReader(body))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	resp, _ := send(t, newTestApp(t, untrusted, "tenant_admin"), req)
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
	require.Empty(t, untrusted.acceptedID)

	trusted := &stubService{}
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/auth/oidc/requests/r-1/accept", strings.NewReader(body))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	resp, decision := send(t, newTestApp(t, trusted, "proxy_admin"), req)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "https://chat.example/cb?code=x", decision["redirect_to"])
	require.Equal(t, "acme", trusted.acceptedSlug)
	require.Equal(t, "r-1", trusted.acceptedID)
	require.Equal(t, "s-1", trusted.acceptedSessID)
}

func TestClientManagementIsTenantScopedAndMapsErrors(t *testing.T) {
	t.Parallel()

	service := &stubService{}
	app := newTestApp(t, service, "tenant_admin")

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/oidc-clients", strings.NewReader(`{"name":"LibreChat","redirect_uris":["https://chat.example/cb"]}`))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	resp, body := send(t, app, req)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.Equal(t, "gds_secret", body["client_secret"])
	require.Equal(t, "tenant-1", service.createdTenant)

	req, _ = http.NewRequest(http.MethodPost, "/api/v1/oidc-clients", strings.NewReader(`{"name":"LibreChat"}`))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	resp, _ = send(t, app, req)
	require.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)

	req, _ = http.NewRequest(http.MethodPost, "/api/v1/oidc-clients/missing/rotate-secret", http.NoBody)
	resp, body = send(t, app, req)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
	require.Equal(t, "oidc_client_not_found", body["error"].(map[string]any)["code"])
}
