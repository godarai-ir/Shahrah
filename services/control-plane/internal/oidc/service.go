package oidc

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

type clientRepository interface {
	GetOIDCClient(ctx context.Context, clientID string) (*Client, error)
	ListOIDCClients(ctx context.Context, tenantID string) ([]Client, error)
	CreateOIDCClient(ctx context.Context, client Client) (*Client, error)
	UpdateOIDCClient(ctx context.Context, tenantID, id string, update UpdateClientRequest) (*Client, error)
	RotateOIDCClientSecret(ctx context.Context, tenantID, id, secretHash string) (*Client, error)
	DeleteOIDCClient(ctx context.Context, tenantID, id string) error
}

type requestRepository interface {
	CreateOIDCAuthorizationRequest(ctx context.Context, request AuthorizationRequest) error
	GetOIDCAuthorizationRequest(ctx context.Context, id string) (*AuthorizationRequest, error)
	AcceptOIDCAuthorizationRequest(ctx context.Context, accepted AuthorizationRequest, codeHash string, codeExpiresAt time.Time) error
	RejectOIDCAuthorizationRequest(ctx context.Context, id string) error
	ConsumeOIDCAuthorizationCode(ctx context.Context, codeHash string) (*AuthorizationRequest, error)
}

type refreshTokenRepository interface {
	CreateOIDCRefreshToken(ctx context.Context, token RefreshToken) error
	// RotateOIDCRefreshToken atomically marks the presented token used and stores next, which
	// inherits the grant fields of the presented token. check runs inside the transaction
	// against the presented token; an error from it aborts the rotation.
	RotateOIDCRefreshToken(ctx context.Context, presentedHash string, next RefreshToken, check func(RefreshToken) error) (*RefreshToken, error)
	RevokeOIDCGrant(ctx context.Context, grantID string) error
	RevokeOIDCUserClientGrants(ctx context.Context, tenantID, userID, clientID string) error
}

type identityRepository interface {
	GetOIDCIdentity(ctx context.Context, tenantID, userID string) (*Identity, error)
}

// SessionLookup resolves tenant-admin login sessions owned by the tenant-admin feature.
type SessionLookup interface {
	LookupSession(ctx context.Context, sessionID string) (*Session, error)
}

// PrincipalResolver builds the gateway auth context, including the model allowlist, for a member.
type PrincipalResolver interface {
	ResolveOIDCUserAuth(ctx context.Context, tenantID, userID string) (*platform.AuthContext, error)
}

// Repository is the persistence required by the provider.
type Repository interface {
	clientRepository
	requestRepository
	refreshTokenRepository
	identityRepository
}

// Dependencies wires the provider's collaborators.
type Dependencies struct {
	Repository Repository
	Sessions   SessionLookup
	Principals PrincipalResolver
	Keys       *KeySet
	Settings   Settings
	// Now overrides the clock in tests.
	Now func() time.Time
}

// Service is the OIDC provider application boundary.
type Service struct {
	clients    clientRepository
	requests   requestRepository
	refresh    refreshTokenRepository
	identities identityRepository
	sessions   SessionLookup
	principals PrincipalResolver
	keys       *KeySet
	settings   Settings
	now        func() time.Time
}

// NewService validates dependencies and creates the provider.
func NewService(deps Dependencies) (*Service, error) {
	if deps.Repository == nil || deps.Sessions == nil || deps.Principals == nil {
		return nil, errors.New("oidc: repository, session lookup, and principal resolver are required")
	}

	if deps.Keys == nil {
		return nil, ErrDisabled
	}

	settings := deps.Settings
	defaults := DefaultSettings()

	settings.IssuerScheme = strings.ToLower(strings.TrimSpace(settings.IssuerScheme))
	if settings.IssuerScheme != "http" && settings.IssuerScheme != "https" {
		settings.IssuerScheme = defaults.IssuerScheme
	}

	if strings.TrimSpace(settings.LoginURLTemplate) == "" {
		settings.LoginURLTemplate = defaults.LoginURLTemplate
	}

	applyDurationDefault(&settings.AccessTokenTTL, defaults.AccessTokenTTL)
	applyDurationDefault(&settings.IDTokenTTL, defaults.IDTokenTTL)
	applyDurationDefault(&settings.RefreshTokenTTL, defaults.RefreshTokenTTL)
	applyDurationDefault(&settings.GrantTTL, defaults.GrantTTL)
	applyDurationDefault(&settings.AuthorizationRequestTTL, defaults.AuthorizationRequestTTL)
	applyDurationDefault(&settings.CodeTTL, defaults.CodeTTL)

	now := deps.Now
	if now == nil {
		now = time.Now
	}

	return &Service{
		clients:    deps.Repository,
		requests:   deps.Repository,
		refresh:    deps.Repository,
		identities: deps.Repository,
		sessions:   deps.Sessions,
		principals: deps.Principals,
		keys:       deps.Keys,
		settings:   settings,
		now:        func() time.Time { return now().UTC() },
	}, nil
}

func applyDurationDefault(value *time.Duration, fallback time.Duration) {
	if *value <= 0 {
		*value = fallback
	}
}

// Issuer returns the issuer URL for a tenant host.
func (s *Service) Issuer(tenant Tenant) string {
	return s.settings.IssuerScheme + "://" + strings.TrimRight(strings.TrimSpace(tenant.Host), "/")
}

// JWKS returns the provider's public signing keys.
func (s *Service) JWKS() JSONWebKeySet {
	return s.keys.JWKS()
}

// Discovery returns the OpenID Provider metadata for a tenant.
func (s *Service) Discovery(tenant Tenant) map[string]any {
	issuer := s.Issuer(tenant)

	return map[string]any{
		"issuer":                                issuer,
		"authorization_endpoint":                issuer + "/oidc/authorize",
		"token_endpoint":                        issuer + "/oidc/token",
		"userinfo_endpoint":                     issuer + "/oidc/userinfo",
		"jwks_uri":                              issuer + "/oidc/jwks",
		"end_session_endpoint":                  issuer + "/oidc/logout",
		"response_types_supported":              []string{ResponseTypeCode},
		"response_modes_supported":              []string{"query"},
		"grant_types_supported":                 []string{GrantTypeAuthorizationCode, GrantTypeRefreshToken},
		"subject_types_supported":               []string{"public"},
		"id_token_signing_alg_values_supported": []string{signingAlgorithm},
		"scopes_supported":                      supportedScopes(),
		"token_endpoint_auth_methods_supported": []string{"client_secret_basic", "client_secret_post"},
		"code_challenge_methods_supported":      []string{CodeChallengeMethodS256},
		"claims_supported": []string{
			"sub", "iss", "aud", "exp", "iat", "auth_time", "nonce", "azp",
			"name", "preferred_username", "email", "email_verified",
			"phone_number", "phone_number_verified", "roles", "tid", "tenant",
		},
		"request_parameter_supported":      false,
		"request_uri_parameter_supported":  false,
		"claims_parameter_supported":       false,
		"frontchannel_logout_supported":    false,
		"backchannel_logout_supported":     false,
		"require_request_uri_registration": false,
	}
}
