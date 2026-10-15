package oidc

import (
	"context"
	"crypto/sha256"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

const tokenTypeBearer = "Bearer"

type accessTokenClaims struct {
	Issuer    string   `json:"iss"`
	Subject   string   `json:"sub"`
	Audience  []string `json:"aud"`
	ExpiresAt int64    `json:"exp"`
	IssuedAt  int64    `json:"iat"`
	NotBefore int64    `json:"nbf"`
	JWTID     string   `json:"jti"`
	ClientID  string   `json:"client_id"`
	Scope     string   `json:"scope"`
	TenantID  string   `json:"tid"`
	GrantID   string   `json:"gid"`
	// Product is the client's credit product label at issuance (for example librechat).
	Product string `json:"gdp,omitempty"`
}

type idTokenClaims struct {
	Issuer              string   `json:"iss"`
	Subject             string   `json:"sub"`
	Audience            string   `json:"aud"`
	AuthorizedParty     string   `json:"azp"`
	ExpiresAt           int64    `json:"exp"`
	IssuedAt            int64    `json:"iat"`
	AuthTime            int64    `json:"auth_time"`
	Nonce               string   `json:"nonce,omitempty"`
	AccessTokenHash     string   `json:"at_hash,omitempty"`
	TenantID            string   `json:"tid"`
	Tenant              string   `json:"tenant,omitempty"`
	Name                string   `json:"name,omitempty"`
	GivenName           string   `json:"given_name,omitempty"`
	FamilyName          string   `json:"family_name,omitempty"`
	PreferredUsername   string   `json:"preferred_username,omitempty"`
	Email               string   `json:"email,omitempty"`
	EmailVerified       *bool    `json:"email_verified,omitempty"`
	PhoneNumber         string   `json:"phone_number,omitempty"`
	PhoneNumberVerified *bool    `json:"phone_number_verified,omitempty"`
	Roles               []string `json:"roles,omitempty"`
}

type tokenGrant struct {
	tenant   Tenant
	issuer   string
	client   *Client
	identity *Identity
	grantID  string
	scope    string
	nonce    string
	authTime time.Time
}

// Token handles the token endpoint for the authorization-code and refresh-token grants.
func (s *Service) Token(ctx context.Context, tenant Tenant, params TokenParams) (*TokenResponse, error) {
	switch strings.TrimSpace(params.GrantType) {
	case GrantTypeAuthorizationCode:
		return s.exchangeCode(ctx, tenant, params)
	case GrantTypeRefreshToken:
		return s.refreshTokens(ctx, tenant, params)
	case "":
		return nil, protocolError(CodeInvalidRequest, "grant_type is required")
	default:
		return nil, protocolError(CodeUnsupportedGrantType, "unsupported grant_type")
	}
}

func (s *Service) exchangeCode(ctx context.Context, tenant Tenant, params TokenParams) (*TokenResponse, error) {
	client, err := s.authenticateClient(ctx, tenant, params)
	if err != nil {
		return nil, err
	}

	code := strings.TrimSpace(params.Code)
	if code == "" {
		return nil, protocolError(CodeInvalidRequest, "code is required")
	}

	request, err := s.requests.ConsumeOIDCAuthorizationCode(ctx, hashToken(code))
	if err != nil {
		return nil, s.mapCodeError(ctx, request, err)
	}

	if request.ClientID != client.ClientID || !strings.EqualFold(request.TenantID, tenant.ID) {
		return nil, protocolError(CodeInvalidGrant, "authorization code was not issued to this client")
	}

	if strings.TrimSpace(params.RedirectURI) != request.RedirectURI {
		return nil, protocolError(CodeInvalidGrant, "redirect_uri does not match the authorization request")
	}

	if !verifyPKCE(strings.TrimSpace(params.CodeVerifier), request.CodeChallenge) {
		return nil, protocolError(CodeInvalidGrant, "PKCE verification failed")
	}

	identity, err := s.activeIdentity(ctx, request.TenantID, request.UserID)
	if err != nil {
		return nil, mapIdentityError(err)
	}

	now := s.now()

	refreshToken, err := s.newRefreshToken(RefreshToken{
		GrantID:        request.GrantID,
		TenantID:       request.TenantID,
		ClientID:       client.ClientID,
		UserID:         request.UserID,
		Scope:          request.Scope,
		Nonce:          request.Nonce,
		AuthTime:       request.AuthTime,
		GrantExpiresAt: now.Add(s.settings.GrantTTL),
	})
	if err != nil {
		return nil, err
	}

	if createErr := s.refresh.CreateOIDCRefreshToken(ctx, refreshToken.record); createErr != nil {
		return nil, createErr
	}

	response, err := s.issueTokens(tokenGrant{
		tenant:   tenant,
		issuer:   request.Issuer,
		client:   client,
		identity: identity,
		grantID:  request.GrantID,
		scope:    request.Scope,
		nonce:    request.Nonce,
		authTime: request.AuthTime,
	})
	if err != nil {
		return nil, err
	}

	response.RefreshToken = refreshToken.raw

	return response, nil
}

func (s *Service) mapCodeError(ctx context.Context, request *AuthorizationRequest, err error) error {
	switch {
	case errors.Is(err, ErrCodeReused):
		// RFC 6749 section 4.1.2: revoke what the replayed code already produced.
		if request != nil && request.GrantID != "" {
			if revokeErr := s.refresh.RevokeOIDCGrant(ctx, request.GrantID); revokeErr != nil {
				return revokeErr
			}
		}

		return protocolError(CodeInvalidGrant, "authorization code is invalid or expired")
	case errors.Is(err, ErrCodeNotFound):
		return protocolError(CodeInvalidGrant, "authorization code is invalid or expired")
	default:
		return err
	}
}

func (s *Service) refreshTokens(ctx context.Context, tenant Tenant, params TokenParams) (*TokenResponse, error) {
	client, err := s.authenticateClient(ctx, tenant, params)
	if err != nil {
		return nil, err
	}

	presented := strings.TrimSpace(params.RefreshToken)
	if presented == "" {
		return nil, protocolError(CodeInvalidRequest, "refresh_token is required")
	}

	next, err := s.newRefreshToken(RefreshToken{TenantID: tenant.ID, ClientID: client.ClientID})
	if err != nil {
		return nil, err
	}

	scope := ""
	now := s.now()

	rotated, err := s.refresh.RotateOIDCRefreshToken(ctx, hashToken(presented), next.record, func(old RefreshToken) error {
		if old.ClientID != client.ClientID || !strings.EqualFold(old.TenantID, tenant.ID) {
			return ErrRefreshTokenNotFound
		}

		if !now.Before(old.ExpiresAt) || !now.Before(old.GrantExpiresAt) {
			return ErrRefreshTokenNotFound
		}

		narrowed, scopeErr := narrowScope(params.Scope, old.Scope)
		scope = narrowed

		return scopeErr
	})
	if err != nil {
		return nil, mapRefreshError(err)
	}

	identity, err := s.activeIdentity(ctx, rotated.TenantID, rotated.UserID)
	if err != nil {
		if errors.Is(err, ErrIdentityNotFound) {
			if revokeErr := s.refresh.RevokeOIDCGrant(ctx, rotated.GrantID); revokeErr != nil {
				return nil, revokeErr
			}
		}

		return nil, mapIdentityError(err)
	}

	response, err := s.issueTokens(tokenGrant{
		tenant:   tenant,
		issuer:   s.Issuer(tenant),
		client:   client,
		identity: identity,
		grantID:  rotated.GrantID,
		scope:    scope,
		authTime: rotated.AuthTime,
	})
	if err != nil {
		return nil, err
	}

	response.RefreshToken = next.raw

	return response, nil
}

func mapRefreshError(err error) error {
	if protocolErr, ok := AsError(err); ok {
		return protocolErr
	}

	if errors.Is(err, ErrRefreshTokenNotFound) || errors.Is(err, ErrRefreshTokenReused) {
		return protocolError(CodeInvalidGrant, "refresh token is invalid, expired, or revoked")
	}

	return err
}

func mapIdentityError(err error) error {
	if errors.Is(err, ErrIdentityNotFound) {
		return protocolError(CodeInvalidGrant, "the account is no longer allowed to sign in")
	}

	return err
}

type issuedRefreshToken struct {
	raw    string
	record RefreshToken
}

func (s *Service) newRefreshToken(base RefreshToken) (*issuedRefreshToken, error) {
	raw, err := randomToken("gdr_")
	if err != nil {
		return nil, err
	}

	now := s.now()
	base.ID = uuid.NewString()
	base.TokenHash = hashToken(raw)
	base.CreatedAt = now

	base.ExpiresAt = now.Add(s.settings.RefreshTokenTTL)
	if !base.GrantExpiresAt.IsZero() && base.GrantExpiresAt.Before(base.ExpiresAt) {
		base.ExpiresAt = base.GrantExpiresAt
	}

	return &issuedRefreshToken{raw: raw, record: base}, nil
}

// authenticateClient implements client_secret_basic and client_secret_post for confidential clients.
func (s *Service) authenticateClient(ctx context.Context, tenant Tenant, params TokenParams) (*Client, error) {
	client, err := s.activeClient(ctx, tenant, strings.TrimSpace(params.ClientID))
	if err != nil {
		return nil, err
	}

	if !secretMatches(params.ClientSecret, client.SecretHash) {
		return nil, protocolError(CodeInvalidClient, "client authentication failed")
	}

	return client, nil
}

func (s *Service) issueTokens(grant tokenGrant) (*TokenResponse, error) {
	now := s.now()
	accessExpiry := now.Add(s.settings.AccessTokenTTL)

	accessToken, err := s.keys.Sign(accessTokenType, accessTokenClaims{
		Issuer:    grant.issuer,
		Subject:   grant.identity.UserID,
		Audience:  []string{grant.client.ClientID},
		ExpiresAt: accessExpiry.Unix(),
		IssuedAt:  now.Unix(),
		NotBefore: now.Unix(),
		JWTID:     uuid.NewString(),
		ClientID:  grant.client.ClientID,
		Scope:     grant.scope,
		TenantID:  grant.tenant.ID,
		GrantID:   grant.grantID,
		Product:   grant.client.CreditProduct,
	})
	if err != nil {
		return nil, err
	}

	idClaims := s.identityClaims(grant, now)
	idClaims.AccessTokenHash = halfHash(accessToken)

	idToken, err := s.keys.Sign(idTokenType, idClaims)
	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken: accessToken,
		TokenType:   tokenTypeBearer,
		ExpiresIn:   int64(s.settings.AccessTokenTTL / time.Second),
		IDToken:     idToken,
		Scope:       grant.scope,
	}, nil
}

func (s *Service) identityClaims(grant tokenGrant, now time.Time) idTokenClaims {
	claims := idTokenClaims{
		Issuer:          grant.issuer,
		Subject:         grant.identity.UserID,
		Audience:        grant.client.ClientID,
		AuthorizedParty: grant.client.ClientID,
		ExpiresAt:       now.Add(s.settings.IDTokenTTL).Unix(),
		IssuedAt:        now.Unix(),
		AuthTime:        grant.authTime.Unix(),
		Nonce:           grant.nonce,
		TenantID:        grant.tenant.ID,
		Tenant:          grant.tenant.Slug,
	}

	applyProfileClaims(&claims, grant.identity, grant.scope)

	return claims
}

// applyProfileClaims releases member attributes according to the granted scopes. Contact
// details are only stored after an OTP or invitation proves control, so they are verified.
func applyProfileClaims(claims *idTokenClaims, identity *Identity, scope string) {
	verified := true

	if hasScope(scope, ScopeProfile) {
		claims.Name = firstNonEmpty(fullName(identity), identity.Name, identity.Email, identity.Mobile)
		claims.GivenName = identity.FirstName
		claims.FamilyName = identity.LastName
		claims.PreferredUsername = firstNonEmpty(identity.Email, identity.Mobile, identity.UserID)

		if identity.Role != "" {
			claims.Roles = []string{identity.Role}
		}
	}

	if hasScope(scope, ScopeEmail) && identity.Email != "" {
		claims.Email = identity.Email
		claims.EmailVerified = &verified
	}

	if hasScope(scope, ScopePhone) && identity.Mobile != "" {
		claims.PhoneNumber = identity.Mobile
		claims.PhoneNumberVerified = &verified
	}
}

// fullName joins the member's first and last name; it is empty when neither is set.
func fullName(identity *Identity) string {
	return strings.TrimSpace(identity.FirstName + " " + identity.LastName)
}

// halfHash computes the OIDC at_hash for an ES256-signed token.
func halfHash(token string) string {
	digest := sha256.Sum256([]byte(token))

	return encodeSegment(digest[:len(digest)/2])
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}

	return ""
}
