// Package management composes feature-owned management HTTP routes.
package management

import (
	"errors"

	managementauth "github.com/Delpak/shahrah/services/control-plane/internal/management/auth"
	"github.com/Delpak/shahrah/services/control-plane/internal/management/billing"
	"github.com/Delpak/shahrah/services/control-plane/internal/management/catalog"
	"github.com/Delpak/shahrah/services/control-plane/internal/management/iam"
	"github.com/Delpak/shahrah/services/control-plane/internal/management/observability"
	managementoidc "github.com/Delpak/shahrah/services/control-plane/internal/management/oidc"
	"github.com/Delpak/shahrah/services/control-plane/internal/management/onboarding"
	"github.com/Delpak/shahrah/services/control-plane/internal/management/support"
	"github.com/gofiber/fiber/v2"
)

// Dependencies contains the handler ports consumed by management features.
type Dependencies struct {
	Auth          managementauth.Handler
	Billing       billing.Handler
	Catalog       catalog.Handler
	IAM           iam.Handler
	Observability observability.Handler
	Onboarding    onboarding.Handler
	Support       support.Handler
	// OIDC is optional; it is nil when the OIDC provider is disabled.
	OIDC *managementoidc.Routes
}

// Handler combines the non-authentication ports for focused HTTP tests.
type Handler interface {
	billing.Handler
	catalog.Handler
	iam.Handler
	observability.Handler
	onboarding.Handler
	support.Handler
}

// Routes is the complete management HTTP surface grouped by feature.
type Routes struct {
	auth          *managementauth.Routes
	billing       *billing.Routes
	catalog       *catalog.Routes
	iam           *iam.Routes
	observability *observability.Routes
	onboarding    *onboarding.Routes
	support       *support.Routes
	oidc          *managementoidc.Routes
}

// New creates a validated management route collection.
func New(deps Dependencies) (*Routes, error) {
	if deps.Auth == nil || deps.Billing == nil || deps.Catalog == nil || deps.IAM == nil ||
		deps.Observability == nil || deps.Onboarding == nil || deps.Support == nil {
		return nil, errors.New("management routes: all feature handlers are required")
	}

	return &Routes{
		auth:          managementauth.NewRoutes(deps.Auth),
		billing:       billing.NewRoutes(deps.Billing),
		catalog:       catalog.NewRoutes(deps.Catalog),
		iam:           iam.NewRoutes(deps.IAM),
		observability: observability.NewRoutes(deps.Observability),
		onboarding:    onboarding.NewRoutes(deps.Onboarding),
		support:       support.NewRoutes(deps.Support),
		oidc:          deps.OIDC,
	}, nil
}

// NewHandlerRoutes creates non-authentication feature routes for focused tests.
func NewHandlerRoutes(handler Handler) *Routes {
	return &Routes{
		billing:       billing.NewRoutes(handler),
		catalog:       catalog.NewRoutes(handler),
		iam:           iam.NewRoutes(handler),
		observability: observability.NewRoutes(handler),
		onboarding:    onboarding.NewRoutes(handler),
		support:       support.NewRoutes(handler),
	}
}

// RegisterPublic registers unauthenticated management endpoints.
func (r *Routes) RegisterPublic(router fiber.Router) {
	if r.onboarding != nil {
		r.onboarding.RegisterPublic(router)
	}

	if r.billing != nil {
		r.billing.RegisterPublic(router)
	}
}

// RegisterAuth registers authentication endpoints.
func (r *Routes) RegisterAuth(router fiber.Router) {
	if r.auth != nil {
		r.auth.Register(router)
	}

	if r.oidc != nil {
		r.oidc.RegisterAuth(router)
	}
}

// RegisterOIDCProvider registers the tenant-host OpenID Connect endpoints when enabled.
func (r *Routes) RegisterOIDCProvider(app fiber.Router) {
	if r.oidc != nil {
		r.oidc.RegisterProvider(app)
	}
}

// RegisterTenant registers authenticated tenant-management endpoints.
func (r *Routes) RegisterTenant(v1 fiber.Router, v2 fiber.Router) {
	if r.onboarding != nil {
		r.onboarding.RegisterTenant(v1)
	}

	if r.iam != nil {
		r.iam.RegisterTenant(v1)
	}

	if r.catalog != nil {
		r.catalog.RegisterTenant(v1)
	}

	if r.billing != nil {
		r.billing.RegisterTenant(v1)
	}

	if r.support != nil {
		r.support.RegisterTenant(v1)
	}

	if r.observability != nil {
		r.observability.RegisterTenant(v1, v2)
	}

	if r.oidc != nil {
		r.oidc.RegisterTenant(v1)
	}
}

// RegisterPlatform registers platform-administration endpoints.
func (r *Routes) RegisterPlatform(router fiber.Router) {
	if r.billing != nil {
		r.billing.RegisterPlatform(router)
	}

	if r.catalog != nil {
		r.catalog.RegisterPlatform(router)
	}

	if r.support != nil {
		r.support.RegisterPlatform(router)
	}
}
