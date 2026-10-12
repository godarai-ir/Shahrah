package oidc

import (
	"context"
	"errors"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

// allowedClockSkew tolerates small clock differences between replicas.
const allowedClockSkew = 30 * time.Second

// VerifyAccessToken validates an access token's signature, type, and lifetime.
func (s *Service) VerifyAccessToken(token string) (*AccessTokenClaims, error) {
	var claims accessTokenClaims
	if err := s.keys.Verify(strings.TrimSpace(token), accessTokenType, &claims); err != nil {
		return nil, protocolError(CodeInvalidToken, "access token is invalid")
	}

	now := s.now()
	if claims.ExpiresAt == 0 || now.After(time.Unix(claims.ExpiresAt, 0).Add(allowedClockSkew)) {
		return nil, protocolError(CodeInvalidToken, "access token has expired")
	}

	if claims.NotBefore != 0 && now.Add(allowedClockSkew).Before(time.Unix(claims.NotBefore, 0)) {
		return nil, protocolError(CodeInvalidToken, "access token is not yet valid")
	}

	if claims.Subject == "" || claims.TenantID == "" || claims.ClientID == "" {
		return nil, protocolError(CodeInvalidToken, "access token is invalid")
	}

	return &AccessTokenClaims{
		Issuer:    claims.Issuer,
		Subject:   claims.Subject,
		TenantID:  claims.TenantID,
		ClientID:  claims.ClientID,
		GrantID:   claims.GrantID,
		Scope:     claims.Scope,
		Product:   claims.Product,
		ExpiresAt: time.Unix(claims.ExpiresAt, 0).UTC(),
	}, nil
}

// AuthenticateAccessToken resolves the gateway auth context for an access token presented to
// tenantID. Membership, role, and the model allowlist are read on every call, so deactivating
// a member or changing their models takes effect without waiting for the token to expire.
func (s *Service) AuthenticateAccessToken(ctx context.Context, tenantID, token string) (*platform.AuthContext, error) {
	claims, err := s.VerifyAccessToken(token)
	if err != nil {
		return nil, err
	}

	if !strings.EqualFold(claims.TenantID, strings.TrimSpace(tenantID)) {
		return nil, protocolError(CodeInvalidToken, "access token was issued for another tenant")
	}

	resolved, err := s.principals.ResolveOIDCUserAuth(ctx, claims.TenantID, claims.Subject)
	if err != nil {
		if errors.Is(err, ErrIdentityNotFound) {
			return nil, protocolError(CodeInvalidToken, "the account is no longer allowed to use this token")
		}

		return nil, err
	}

	if resolved == nil || !strings.EqualFold(resolved.UserID, claims.Subject) || !strings.EqualFold(resolved.TenantID, claims.TenantID) {
		return nil, protocolError(CodeInvalidToken, "the account is no longer allowed to use this token")
	}

	resolved.KeyID = claims.GrantID
	resolved.KeyAlias = "oidc:" + claims.ClientID
	resolved.KeyType = KeyType
	resolved.CreditProduct = accessTokenProduct(claims.Product)

	return resolved, nil
}

// UserInfo returns the claims released to the access token's scopes.
func (s *Service) UserInfo(ctx context.Context, tenant Tenant, token string) (map[string]any, error) {
	claims, err := s.VerifyAccessToken(token)
	if err != nil {
		return nil, err
	}

	if !strings.EqualFold(claims.TenantID, tenant.ID) {
		return nil, protocolError(CodeInvalidToken, "access token was issued for another tenant")
	}

	identity, err := s.activeIdentity(ctx, claims.TenantID, claims.Subject)
	if err != nil {
		if errors.Is(err, ErrIdentityNotFound) {
			return nil, protocolError(CodeInvalidToken, "the account is no longer allowed to sign in")
		}

		return nil, err
	}

	var profile idTokenClaims
	applyProfileClaims(&profile, identity, claims.Scope)

	info := map[string]any{"sub": identity.UserID, "tid": identity.TenantID, "tenant": tenant.Slug}
	addClaim(info, "name", profile.Name)
	addClaim(info, "preferred_username", profile.PreferredUsername)
	addClaim(info, "email", profile.Email)
	addClaim(info, "phone_number", profile.PhoneNumber)

	if profile.EmailVerified != nil {
		info["email_verified"] = *profile.EmailVerified
	}

	if profile.PhoneNumberVerified != nil {
		info["phone_number_verified"] = *profile.PhoneNumberVerified
	}

	if len(profile.Roles) > 0 {
		info["roles"] = profile.Roles
	}

	return info, nil
}

func addClaim(claims map[string]any, name, value string) {
	if value != "" {
		claims[name] = value
	}
}

// Logout implements RP-initiated logout. It revokes the client's refresh tokens for the
// member named by id_token_hint and returns the post-logout redirect, or "" when the
// client did not ask for one. The tenant-admin session itself is left untouched.
func (s *Service) Logout(ctx context.Context, tenant Tenant, params LogoutParams) (string, error) {
	clientID := strings.TrimSpace(params.ClientID)
	subject := ""

	if hint := strings.TrimSpace(params.IDTokenHint); hint != "" {
		var claims idTokenClaims
		if err := s.keys.Verify(hint, idTokenType, &claims); err != nil || !strings.EqualFold(claims.TenantID, tenant.ID) {
			return "", protocolError(CodeInvalidRequest, "id_token_hint is invalid")
		}

		if clientID != "" && clientID != claims.Audience {
			return "", protocolError(CodeInvalidRequest, "client_id does not match id_token_hint")
		}

		clientID = claims.Audience
		subject = claims.Subject
	}

	redirectURI := strings.TrimSpace(params.PostLogoutRedirectURI)
	if clientID == "" {
		if redirectURI != "" {
			return "", protocolError(CodeInvalidRequest, "post_logout_redirect_uri requires id_token_hint or client_id")
		}

		return "", nil
	}

	client, err := s.activeClient(ctx, tenant, clientID)
	if err != nil {
		return "", err
	}

	if redirectURI != "" && !slices.Contains(client.PostLogoutRedirectURIs, redirectURI) {
		return "", protocolError(CodeInvalidRequest, "post_logout_redirect_uri is not registered for this client")
	}

	if subject != "" {
		if err := s.refresh.RevokeOIDCUserClientGrants(ctx, tenant.ID, subject, client.ClientID); err != nil {
			return "", err
		}
	}

	if redirectURI == "" {
		return "", nil
	}

	return appendQuery(redirectURI, url.Values{"state": {params.State}})
}

// accessTokenProduct is the trusted product label for gateway usage made with an
// access token. Tokens issued before clients carried a product use the default.
func accessTokenProduct(product string) string {
	switch product {
	case platform.CreditProductLibreChat, platform.CreditProductOIDCApp:
		return product
	default:
		return platform.CreditProductOIDCApp
	}
}
