// Package oidc owns the OpenID Connect provider and client-management HTTP routes.
package oidc

import (
	"context"
	"errors"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/oidc"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// Provider is the OIDC application surface consumed by these routes.
type Provider interface {
	Discovery(tenant oidc.Tenant) map[string]any
	JWKS() oidc.JSONWebKeySet
	Authorize(ctx context.Context, tenant oidc.Tenant, params oidc.AuthorizeParams) (*oidc.AuthorizeResult, error)
	Token(ctx context.Context, tenant oidc.Tenant, params oidc.TokenParams) (*oidc.TokenResponse, error)
	UserInfo(ctx context.Context, tenant oidc.Tenant, token string) (map[string]any, error)
	Logout(ctx context.Context, tenant oidc.Tenant, params oidc.LogoutParams) (string, error)
}

// LoginBridge is the trusted login-challenge surface used by the tenant-admin server.
type LoginBridge interface {
	GetLoginRequest(ctx context.Context, tenantSlug, requestID string) (*oidc.LoginRequest, error)
	AcceptLogin(ctx context.Context, tenantSlug, requestID, sessionID string) (*oidc.LoginDecision, error)
	RejectLogin(ctx context.Context, tenantSlug, requestID string) (*oidc.LoginDecision, error)
}

// ClientManager is the tenant client-registration surface.
type ClientManager interface {
	ListClients(ctx context.Context, tenantID string) ([]oidc.Client, error)
	CreateClient(ctx context.Context, tenantID, actor string, request oidc.CreateClientRequest) (*oidc.ClientWithSecret, error)
	UpdateClient(ctx context.Context, tenantID, id string, request oidc.UpdateClientRequest) (*oidc.Client, error)
	RotateClientSecret(ctx context.Context, tenantID, id string) (*oidc.ClientWithSecret, error)
	DeleteClient(ctx context.Context, tenantID, id string) error
}

// Service combines every surface the routes need.
type Service interface {
	Provider
	LoginBridge
	ClientManager
}

// TenantResolver reads the tenant resolved by the tenant middleware.
type TenantResolver func(c *fiber.Ctx) *platform.Tenant

// AuthInfoResolver reads the authenticated principal.
type AuthInfoResolver func(c *fiber.Ctx) *platform.AuthContext

// Routes owns OIDC route registration.
type Routes struct {
	service     Service
	tenant      TenantResolver
	authInfo    AuthInfoResolver
	isPrivilege func(role string) bool
	logger      *zap.Logger
}

// RoutesDependencies wires the routes to the request context helpers.
type RoutesDependencies struct {
	Service  Service
	Tenant   TenantResolver
	AuthInfo AuthInfoResolver
	// IsProxyAdmin reports whether a role may drive the trusted login bridge.
	IsProxyAdmin func(role string) bool
	Logger       *zap.Logger
}

// NewRoutes creates validated OIDC routes.
func NewRoutes(deps RoutesDependencies) (*Routes, error) {
	if deps.Service == nil || deps.Tenant == nil || deps.AuthInfo == nil || deps.IsProxyAdmin == nil {
		return nil, errors.New("oidc routes: service and request-context helpers are required")
	}

	logger := deps.Logger
	if logger == nil {
		logger = zap.NewNop()
	}

	return &Routes{
		service:     deps.Service,
		tenant:      deps.Tenant,
		authInfo:    deps.AuthInfo,
		isPrivilege: deps.IsProxyAdmin,
		logger:      logger,
	}, nil
}

// RegisterProvider registers the protocol endpoints on the tenant host. They run after tenant
// resolution and are exempt from API-key authentication.
func (r *Routes) RegisterProvider(app fiber.Router) {
	app.Get("/.well-known/openid-configuration", r.discovery)
	app.Get("/oidc/jwks", r.jwks)
	app.Get("/oidc/authorize", r.authorize)
	app.Post("/oidc/authorize", r.authorize)
	app.Post("/oidc/token", r.token)
	app.Get("/oidc/userinfo", r.userInfo)
	app.Post("/oidc/userinfo", r.userInfo)
	app.Get("/oidc/logout", r.logout)
	app.Post("/oidc/logout", r.logout)
}

// RegisterAuth registers the login bridge under the trusted /api/v1/auth group.
func (r *Routes) RegisterAuth(router fiber.Router) {
	router.Get("/oidc/requests/:id", r.getLoginRequest)
	router.Post("/oidc/requests/:id/accept", r.acceptLogin)
	router.Post("/oidc/requests/:id/reject", r.rejectLogin)
}

// RegisterTenant registers tenant client management under /api/v1.
func (r *Routes) RegisterTenant(router fiber.Router) {
	router.Get("/oidc-clients", r.listClients)
	router.Post("/oidc-clients", r.createClient)
	router.Patch("/oidc-clients/:id", r.updateClient)
	router.Post("/oidc-clients/:id/rotate-secret", r.rotateClientSecret)
	router.Delete("/oidc-clients/:id", r.deleteClient)
}
