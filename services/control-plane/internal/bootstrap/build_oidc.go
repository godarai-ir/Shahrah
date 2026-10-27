package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	managementoidc "github.com/Delpak/shahrah/services/control-plane/internal/management/oidc"
	"github.com/Delpak/shahrah/services/control-plane/internal/oidc"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenantadmin"
)

// buildOIDC constructs the OIDC provider when it is enabled. A disabled provider leaves
// every route and the gateway's access-token branch unregistered.
func (b *builder) buildOIDC() error {
	settings := b.cfg.OIDC
	if !settings.Enabled {
		return nil
	}

	if settings.LoadError != nil {
		return settings.LoadError
	}

	keys, err := b.oidcKeySet()
	if err != nil {
		return err
	}

	assertions, err := b.userAssertionVerifier()
	if err != nil {
		return err
	}

	service, err := oidc.NewService(oidc.Dependencies{
		Repository: b.store.OIDCRepository,
		Sessions:   tenantAdminSessionLookup{sessions: b.tenantAdminApplication},
		Principals: b.store.AuthRepository,
		Keys:       keys,
		Assertions: assertions,
		Settings: oidc.Settings{
			IssuerScheme:            settings.IssuerScheme,
			LoginURLTemplate:        settings.LoginURLTemplate,
			AccessTokenTTL:          settings.AccessTokenTTL,
			IDTokenTTL:              settings.IDTokenTTL,
			RefreshTokenTTL:         settings.RefreshTokenTTL,
			GrantTTL:                settings.GrantTTL,
			AuthorizationRequestTTL: settings.AuthorizationRequestTTL,
			CodeTTL:                 settings.CodeTTL,
		},
	})
	if err != nil {
		return fmt.Errorf("initialize oidc provider: %w", err)
	}

	b.oidcApplication = service

	return nil
}

func (b *builder) oidcKeySet() (*oidc.KeySet, error) {
	settings := b.cfg.OIDC

	switch {
	case settings.SigningKeyPEM != "":
		keys, err := oidc.NewKeySet(settings.SigningKeyPEM, settings.RetiredPublicKeysPEM)
		if err != nil {
			return nil, fmt.Errorf("load oidc signing key: %w", err)
		}

		return keys, nil
	case settings.EphemeralSigningKey:
		b.logger.Warn("OIDC is using an ephemeral signing key; issued tokens will not survive a restart")

		return oidc.GenerateKeySet()
	default:
		return nil, errors.New("OIDC_ENABLED requires OIDC_SIGNING_KEY_PEM or OIDC_SIGNING_KEY_FILE")
	}
}

// userAssertionVerifier is nil when no LibreChat verification key is configured.
func (b *builder) userAssertionVerifier() (*oidc.UserAssertionVerifier, error) {
	settings := b.cfg.OIDC.UserAssertion
	if !settings.Enabled() {
		return nil, nil
	}

	verifier, err := oidc.NewUserAssertionVerifier(oidc.UserAssertionSettings{
		PublicKeyPEM: settings.PublicKeyPEM,
		KeyID:        settings.KeyID,
		Issuer:       settings.Issuer,
		Audience:     settings.Audience,
	})
	if err != nil {
		return nil, fmt.Errorf("load user assertion verifier: %w", err)
	}

	return verifier, nil
}

func (b *builder) oidcAccessTokens() auth.AccessTokenAuthenticator {
	if b.oidcApplication == nil {
		return nil
	}

	return b.oidcApplication
}

func (b *builder) oidcRoutes() (*managementoidc.Routes, error) {
	if b.oidcApplication == nil {
		return nil, nil
	}

	return managementoidc.NewRoutes(managementoidc.RoutesDependencies{
		Service:      b.oidcApplication,
		Tenant:       tenant.FromContext,
		AuthInfo:     auth.GetAuthInfo,
		IsProxyAdmin: auth.IsProxyAdmin,
		Logger:       b.logger,
	})
}

type tenantAdminSessionReader interface {
	GetTenantAdminSession(ctx context.Context, sessionID string) (*tenantadmin.TenantAdminSessionResponse, error)
}

// tenantAdminSessionLookup adapts tenant-admin sessions to the provider's session port.
type tenantAdminSessionLookup struct {
	sessions tenantAdminSessionReader
}

func (l tenantAdminSessionLookup) LookupSession(ctx context.Context, sessionID string) (*oidc.Session, error) {
	if sessionID == "" {
		return nil, nil
	}

	session, err := l.sessions.GetTenantAdminSession(ctx, sessionID)
	if err != nil || session == nil {
		return nil, err
	}

	createdAt, err := time.Parse(time.RFC3339, session.CreatedAtISO)
	if err != nil {
		return nil, fmt.Errorf("parse tenant-admin session time: %w", err)
	}

	return &oidc.Session{
		ID:         session.ID,
		TenantSlug: session.TenantSlug,
		UserID:     session.UserID,
		CreatedAt:  createdAt.UTC(),
	}, nil
}
