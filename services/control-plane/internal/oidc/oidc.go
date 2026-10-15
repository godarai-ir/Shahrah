// Package oidc implements the tenant-scoped OpenID Connect provider.
//
// Godar is the identity source of truth. Each tenant host is its own issuer, so
// a client registered for one tenant can never obtain tokens for another. The
// provider supports the authorization-code flow with mandatory PKCE (S256),
// rotating refresh tokens, userinfo, and RP-initiated logout. Login itself is
// delegated to the tenant-admin application through a login challenge: the
// authorize endpoint records the request and redirects to the login page, and
// the tenant-admin server accepts the challenge for an authenticated session.
package oidc

import (
	"errors"
	"net/http"
	"time"
)

// Supported protocol values.
const (
	ScopeOpenID        = "openid"
	ScopeProfile       = "profile"
	ScopeEmail         = "email"
	ScopePhone         = "phone"
	ScopeOfflineAccess = "offline_access"

	ResponseTypeCode        = "code"
	CodeChallengeMethodS256 = "S256"

	GrantTypeAuthorizationCode = "authorization_code"
	GrantTypeRefreshToken      = "refresh_token"

	PromptNone  = "none"
	PromptLogin = "login"

	// KeyType identifies gateway auth contexts created from OIDC access tokens.
	KeyType = "oidc"

	accessTokenType = "at+jwt"
	idTokenType     = "JWT"
)

// OAuth 2.0 and OpenID Connect error codes returned to clients.
const (
	CodeInvalidRequest       = "invalid_request"
	CodeInvalidClient        = "invalid_client"
	CodeInvalidGrant         = "invalid_grant"
	CodeUnauthorizedClient   = "unauthorized_client"
	CodeUnsupportedGrantType = "unsupported_grant_type"
	CodeUnsupportedResponse  = "unsupported_response_type"
	CodeInvalidScope         = "invalid_scope"
	CodeAccessDenied         = "access_denied"
	CodeLoginRequired        = "login_required"
	CodeInvalidToken         = "invalid_token"
	CodeServerError          = "server_error"
)

// Error is a protocol error with a stable OAuth error code and a client-safe description.
type Error struct {
	Code        string
	Description string
	Status      int
}

func (e *Error) Error() string {
	if e.Description == "" {
		return e.Code
	}

	return e.Code + ": " + e.Description
}

func protocolError(code, description string) *Error {
	status := http.StatusBadRequest

	switch code {
	case CodeInvalidClient, CodeInvalidToken:
		status = http.StatusUnauthorized
	case CodeAccessDenied:
		status = http.StatusForbidden
	case CodeServerError:
		status = http.StatusInternalServerError
	}

	return &Error{Code: code, Description: description, Status: status}
}

// AsError returns the protocol error carried by err, if any.
func AsError(err error) (*Error, bool) {
	var protocolErr *Error
	if errors.As(err, &protocolErr) {
		return protocolErr, true
	}

	return nil, false
}

// IsProtocolError reports whether err is a client-facing protocol error.
func IsProtocolError(err error) bool {
	var protocolErr *Error

	return errors.As(err, &protocolErr)
}

// Repository and validation errors.
var (
	// ErrClientNotFound indicates an unknown or foreign-tenant client.
	ErrClientNotFound = errors.New("oidc client not found")
	// ErrClientIDTaken indicates a duplicate client identifier.
	ErrClientIDTaken = errors.New("oidc client id is already taken")
	// ErrInvalidClientConfig indicates an invalid client registration.
	ErrInvalidClientConfig = errors.New("invalid oidc client configuration")
	// ErrRequestNotFound indicates an unknown authorization request.
	ErrRequestNotFound = errors.New("oidc authorization request not found")
	// ErrRequestNotPending indicates an expired or already completed authorization request.
	ErrRequestNotPending = errors.New("oidc authorization request is no longer pending")
	// ErrSessionInvalid indicates a missing, expired, or foreign-tenant login session.
	ErrSessionInvalid = errors.New("oidc login session is invalid")
	// ErrCodeNotFound indicates an unknown or expired authorization code.
	ErrCodeNotFound = errors.New("oidc authorization code not found")
	// ErrCodeReused indicates a replayed authorization code.
	ErrCodeReused = errors.New("oidc authorization code was already used")
	// ErrRefreshTokenNotFound indicates an unknown, expired, or revoked refresh token.
	ErrRefreshTokenNotFound = errors.New("oidc refresh token not found")
	// ErrRefreshTokenReused indicates a replayed refresh token.
	ErrRefreshTokenReused = errors.New("oidc refresh token was already used")
	// ErrIdentityNotFound indicates a missing or deactivated tenant member.
	ErrIdentityNotFound = errors.New("oidc identity not found")
	// ErrDisabled indicates that no signing key is configured.
	ErrDisabled = errors.New("oidc provider is disabled")
)

// Settings holds the provider's immutable configuration.
type Settings struct {
	// IssuerScheme is the URL scheme used to build each tenant issuer from its host.
	IssuerScheme string
	// LoginURLTemplate is the tenant-admin login page; {host} and {slug} are replaced per tenant.
	LoginURLTemplate        string
	AccessTokenTTL          time.Duration
	IDTokenTTL              time.Duration
	RefreshTokenTTL         time.Duration
	GrantTTL                time.Duration
	AuthorizationRequestTTL time.Duration
	CodeTTL                 time.Duration
}

// DefaultSettings returns conservative token lifetimes.
func DefaultSettings() Settings {
	return Settings{
		IssuerScheme:            "https",
		LoginURLTemplate:        "https://{host}/login",
		AccessTokenTTL:          10 * time.Minute,
		IDTokenTTL:              10 * time.Minute,
		RefreshTokenTTL:         14 * 24 * time.Hour,
		GrantTTL:                30 * 24 * time.Hour,
		AuthorizationRequestTTL: 10 * time.Minute,
		CodeTTL:                 time.Minute,
	}
}

// Tenant identifies the issuer a request was made against.
type Tenant struct {
	ID   string
	Slug string
	// Host is the request host (including a non-default port) the issuer is derived from.
	Host string
}

// Client is a registered relying party.
type Client struct {
	ID                     string    `json:"id"`
	TenantID               string    `json:"tenant_id"`
	ClientID               string    `json:"client_id"`
	Name                   string    `json:"name"`
	SecretHash             string    `json:"-"`
	RedirectURIs           []string  `json:"redirect_uris"`
	PostLogoutRedirectURIs []string  `json:"post_logout_redirect_uris"`
	Active                 bool      `json:"active"`
	CreditProduct          string    `json:"credit_product"`
	CreatedBy              string    `json:"created_by,omitempty"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
	SecretRotatedAt        time.Time `json:"secret_rotated_at"`
}

// ClientWithSecret is returned once, when a client secret is created or rotated.
type ClientWithSecret struct {
	Client
	ClientSecret string `json:"client_secret"`
}

// CreateClientRequest registers a relying party.
type CreateClientRequest struct {
	Name                   string   `json:"name"`
	RedirectURIs           []string `json:"redirect_uris"`
	PostLogoutRedirectURIs []string `json:"post_logout_redirect_uris"`
	CreditProduct          string   `json:"credit_product,omitempty"`
}

// UpdateClientRequest changes mutable client fields; nil fields are unchanged.
type UpdateClientRequest struct {
	Name                   *string   `json:"name"`
	RedirectURIs           *[]string `json:"redirect_uris"`
	PostLogoutRedirectURIs *[]string `json:"post_logout_redirect_uris"`
	Active                 *bool     `json:"active"`
	CreditProduct          *string   `json:"credit_product"`
}

// AuthorizationRequest is a pending or completed authorization-code request.
type AuthorizationRequest struct {
	ID                  string
	TenantID            string
	TenantSlug          string
	Issuer              string
	ClientID            string
	RedirectURI         string
	Scope               string
	State               string
	Nonce               string
	CodeChallenge       string
	CodeChallengeMethod string
	Prompt              string
	Status              string
	UserID              string
	GrantID             string
	AuthTime            time.Time
	CreatedAt           time.Time
	ExpiresAt           time.Time
}

// Authorization request statuses.
const (
	RequestStatusPending  = "pending"
	RequestStatusAccepted = "accepted"
	RequestStatusRejected = "rejected"
)

// AuthorizeParams are the raw authorize endpoint parameters.
type AuthorizeParams struct {
	ClientID            string
	RedirectURI         string
	ResponseType        string
	Scope               string
	State               string
	Nonce               string
	CodeChallenge       string
	CodeChallengeMethod string
	Prompt              string
	ResponseMode        string
}

// AuthorizeResult tells the transport where to send the browser.
type AuthorizeResult struct {
	RedirectURL string
}

// LoginRequest is the login-page view of a pending authorization request.
type LoginRequest struct {
	ID         string    `json:"id"`
	TenantSlug string    `json:"tenant_slug"`
	ClientID   string    `json:"client_id"`
	ClientName string    `json:"client_name"`
	Scopes     []string  `json:"scopes"`
	Prompt     string    `json:"prompt,omitempty"`
	ExpiresAt  time.Time `json:"expires_at"`
}

// LoginDecision is returned after a login request is accepted or rejected.
type LoginDecision struct {
	RedirectTo string `json:"redirect_to"`
}

// TokenParams are the raw token endpoint parameters after client authentication is extracted.
type TokenParams struct {
	GrantType    string
	ClientID     string
	ClientSecret string
	Code         string
	RedirectURI  string
	CodeVerifier string
	RefreshToken string
	Scope        string
}

// TokenResponse is the token endpoint success body.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
	Scope        string `json:"scope"`
}

// RefreshToken is a persisted refresh token record.
type RefreshToken struct {
	ID             string
	GrantID        string
	TenantID       string
	ClientID       string
	UserID         string
	TokenHash      string
	Scope          string
	Nonce          string
	AuthTime       time.Time
	CreatedAt      time.Time
	ExpiresAt      time.Time
	GrantExpiresAt time.Time
}

// Identity is the tenant member a token is issued for.
type Identity struct {
	UserID   string
	TenantID string
	Email    string
	Mobile   string
	Name     string
	// FirstName and LastName are the member's given and family names; both are optional.
	FirstName string
	LastName  string
	Role      string
	Active    bool
}

// Session is an authenticated tenant-admin login session.
type Session struct {
	ID         string
	TenantSlug string
	UserID     string
	CreatedAt  time.Time
}

// AccessTokenClaims are the verified claims of a gateway access token.
type AccessTokenClaims struct {
	Issuer    string
	Subject   string
	TenantID  string
	ClientID  string
	GrantID   string
	Scope     string
	Product   string
	ExpiresAt time.Time
}

// LogoutParams are the RP-initiated logout parameters.
type LogoutParams struct {
	IDTokenHint           string
	ClientID              string
	PostLogoutRedirectURI string
	State                 string
}
