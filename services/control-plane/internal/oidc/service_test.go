package oidc

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	testTenantID   = "11111111-1111-4111-8111-111111111111"
	otherTenantID  = "22222222-2222-4222-8222-222222222222"
	testUserID     = "33333333-3333-4333-8333-333333333333"
	testRedirect   = "https://chat.acme.example/oauth/openid/callback"
	testLogoutURI  = "https://chat.acme.example/login"
	testVerifier   = "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk-verifier"
	testSessionID  = "session-1"
	testOtherSlug  = "globex"
	testTenantSlug = "acme"
)

type fixture struct {
	t        *testing.T
	service  *Service
	repo     *memoryRepository
	sessions memorySessions
	clock    time.Time
	tenant   Tenant
	client   *ClientWithSecret
}

func newFixture(t *testing.T) *fixture {
	t.Helper()

	f := &fixture{
		t:      t,
		clock:  time.Date(2026, 10, 7, 9, 0, 0, 0, time.UTC),
		tenant: Tenant{ID: testTenantID, Slug: testTenantSlug, Host: "acme.godar.example"},
	}

	f.repo = newMemoryRepository(f.now)
	f.repo.setIdentity(Identity{
		UserID: testUserID, TenantID: testTenantID, Email: "sara@acme.example",
		Mobile: "+989121234567", Role: "member", Active: true,
	})
	f.sessions = memorySessions{sessions: map[string]*Session{
		testSessionID: {ID: testSessionID, TenantSlug: testTenantSlug, UserID: testUserID, CreatedAt: f.clock.Add(-time.Minute)},
	}}

	keys, err := GenerateKeySet()
	require.NoError(t, err)

	f.service, err = NewService(Dependencies{
		Repository: f.repo,
		Sessions:   f.sessions,
		Principals: memoryPrincipals{repo: f.repo},
		Keys:       keys,
		Settings:   Settings{LoginURLTemplate: "https://www.{host}/login"},
		Now:        f.now,
	})
	require.NoError(t, err)

	f.client, err = f.service.CreateClient(context.Background(), testTenantID, "admin", CreateClientRequest{
		Name:                   "LibreChat",
		RedirectURIs:           []string{testRedirect},
		PostLogoutRedirectURIs: []string{testLogoutURI},
	})
	require.NoError(t, err)

	return f
}

func (f *fixture) now() time.Time {
	return f.clock
}

func challengeFor(verifier string) string {
	digest := sha256.Sum256([]byte(verifier))

	return base64.RawURLEncoding.EncodeToString(digest[:])
}

func (f *fixture) authorizeParams() AuthorizeParams {
	return AuthorizeParams{
		ClientID:            f.client.ClientID,
		RedirectURI:         testRedirect,
		ResponseType:        ResponseTypeCode,
		Scope:               "openid profile email unknown-scope",
		State:               "state-123",
		Nonce:               "nonce-456",
		CodeChallenge:       challengeFor(testVerifier),
		CodeChallengeMethod: CodeChallengeMethodS256,
	}
}

// startLogin runs authorize and returns the login-challenge identifier.
func (f *fixture) startLogin(params AuthorizeParams) string {
	f.t.Helper()

	result, err := f.service.Authorize(context.Background(), f.tenant, params)
	require.NoError(f.t, err)

	target, err := url.Parse(result.RedirectURL)
	require.NoError(f.t, err)
	require.Equal(f.t, "www.acme.godar.example", target.Host)
	require.Equal(f.t, "/login", target.Path)

	requestID := target.Query().Get("oidc_request")
	require.NotEmpty(f.t, requestID)

	return requestID
}

// login completes authorize and accept and returns the authorization code.
func (f *fixture) login() string {
	f.t.Helper()

	requestID := f.startLogin(f.authorizeParams())

	decision, err := f.service.AcceptLogin(context.Background(), testTenantSlug, requestID, testSessionID)
	require.NoError(f.t, err)

	target, err := url.Parse(decision.RedirectTo)
	require.NoError(f.t, err)
	require.Equal(f.t, "state-123", target.Query().Get("state"))
	require.Equal(f.t, "https://acme.godar.example", target.Query().Get("iss"))

	return target.Query().Get("code")
}

func (f *fixture) exchange(code string) (*TokenResponse, error) {
	return f.service.Token(context.Background(), f.tenant, TokenParams{
		GrantType:    GrantTypeAuthorizationCode,
		ClientID:     f.client.ClientID,
		ClientSecret: f.client.ClientSecret,
		Code:         code,
		RedirectURI:  testRedirect,
		CodeVerifier: testVerifier,
	})
}

func (f *fixture) refreshWith(token string) (*TokenResponse, error) {
	return f.service.Token(context.Background(), f.tenant, TokenParams{
		GrantType:    GrantTypeRefreshToken,
		ClientID:     f.client.ClientID,
		ClientSecret: f.client.ClientSecret,
		RefreshToken: token,
	})
}

func requireProtocolCode(t *testing.T, err error, code string) {
	t.Helper()

	protocolErr, ok := AsError(err)
	require.True(t, ok, "expected protocol error, got %v", err)
	require.Equal(t, code, protocolErr.Code)
}

func TestAuthorizationCodeFlowIssuesVerifiableTokens(t *testing.T) {
	t.Parallel()

	f := newFixture(t)

	tokens, err := f.exchange(f.login())
	require.NoError(t, err)
	require.Equal(t, "Bearer", tokens.TokenType)
	require.Equal(t, "openid profile email", tokens.Scope)
	require.NotEmpty(t, tokens.RefreshToken)

	var id idTokenClaims
	require.NoError(t, f.service.keys.Verify(tokens.IDToken, idTokenType, &id))
	require.Equal(t, "https://acme.godar.example", id.Issuer)
	require.Equal(t, testUserID, id.Subject)
	require.Equal(t, f.client.ClientID, id.Audience)
	require.Equal(t, "nonce-456", id.Nonce)
	require.Equal(t, "sara@acme.example", id.Email)
	require.Empty(t, id.PhoneNumber, "phone scope was not granted")
	require.Equal(t, []string{"member"}, id.Roles)
	require.Equal(t, testTenantID, id.TenantID)
	require.Equal(t, halfHash(tokens.AccessToken), id.AccessTokenHash)

	claims, err := f.service.VerifyAccessToken(tokens.AccessToken)
	require.NoError(t, err)
	require.Equal(t, testUserID, claims.Subject)
	require.Equal(t, f.client.ClientID, claims.ClientID)

	info, err := f.service.UserInfo(context.Background(), f.tenant, tokens.AccessToken)
	require.NoError(t, err)
	require.Equal(t, testUserID, info["sub"])
	require.Equal(t, "sara@acme.example", info["email"])
	require.NotContains(t, info, "phone_number")
}

func TestAccessTokensAreNotAcceptedAsIDTokensAndViceVersa(t *testing.T) {
	t.Parallel()

	f := newFixture(t)

	tokens, err := f.exchange(f.login())
	require.NoError(t, err)

	_, err = f.service.VerifyAccessToken(tokens.IDToken)
	requireProtocolCode(t, err, CodeInvalidToken)
}

func TestGatewayAuthenticationIsTenantBound(t *testing.T) {
	t.Parallel()

	f := newFixture(t)

	tokens, err := f.exchange(f.login())
	require.NoError(t, err)

	resolved, err := f.service.AuthenticateAccessToken(context.Background(), testTenantID, tokens.AccessToken)
	require.NoError(t, err)
	require.Equal(t, testUserID, resolved.UserID)
	require.Equal(t, KeyType, resolved.KeyType)
	require.NotEmpty(t, resolved.KeyID, "grant id scopes caches and spend attribution")
	require.True(t, resolved.EnforceModelAllowlist)

	_, err = f.service.AuthenticateAccessToken(context.Background(), otherTenantID, tokens.AccessToken)
	requireProtocolCode(t, err, CodeInvalidToken)
}

func TestGatewayAuthenticationStopsForDeactivatedMembers(t *testing.T) {
	t.Parallel()

	f := newFixture(t)

	tokens, err := f.exchange(f.login())
	require.NoError(t, err)

	f.repo.setIdentity(Identity{UserID: testUserID, TenantID: testTenantID, Active: false})

	_, err = f.service.AuthenticateAccessToken(context.Background(), testTenantID, tokens.AccessToken)
	requireProtocolCode(t, err, CodeInvalidToken)

	_, err = f.refreshWith(tokens.RefreshToken)
	requireProtocolCode(t, err, CodeInvalidGrant)
}

func TestExpiredAccessTokenIsRejected(t *testing.T) {
	t.Parallel()

	f := newFixture(t)

	tokens, err := f.exchange(f.login())
	require.NoError(t, err)

	f.clock = f.clock.Add(11 * time.Minute)

	_, err = f.service.AuthenticateAccessToken(context.Background(), testTenantID, tokens.AccessToken)
	requireProtocolCode(t, err, CodeInvalidToken)
}

func TestAuthorizationCodeIsSingleUseAndReplayRevokesTheGrant(t *testing.T) {
	t.Parallel()

	f := newFixture(t)
	code := f.login()

	tokens, err := f.exchange(code)
	require.NoError(t, err)

	_, err = f.exchange(code)
	requireProtocolCode(t, err, CodeInvalidGrant)

	_, err = f.refreshWith(tokens.RefreshToken)
	requireProtocolCode(t, err, CodeInvalidGrant)
}

func TestCodeExchangeRequiresMatchingVerifierRedirectAndSecret(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		mutate func(*TokenParams)
		code   string
	}{
		"wrong verifier": {mutate: func(p *TokenParams) { p.CodeVerifier = strings.Repeat("a", 50) }, code: CodeInvalidGrant},
		"wrong redirect": {mutate: func(p *TokenParams) { p.RedirectURI = "https://chat.acme.example/other" }, code: CodeInvalidGrant},
		"wrong secret":   {mutate: func(p *TokenParams) { p.ClientSecret = "gds_wrong" }, code: CodeInvalidClient},
		"missing secret": {mutate: func(p *TokenParams) { p.ClientSecret = "" }, code: CodeInvalidClient},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			f := newFixture(t)
			params := TokenParams{
				GrantType:    GrantTypeAuthorizationCode,
				ClientID:     f.client.ClientID,
				ClientSecret: f.client.ClientSecret,
				Code:         f.login(),
				RedirectURI:  testRedirect,
				CodeVerifier: testVerifier,
			}
			tc.mutate(&params)

			_, err := f.service.Token(context.Background(), f.tenant, params)
			requireProtocolCode(t, err, tc.code)
		})
	}
}

func TestExpiredAuthorizationCodeIsRejected(t *testing.T) {
	t.Parallel()

	f := newFixture(t)
	code := f.login()
	f.clock = f.clock.Add(2 * time.Minute)

	_, err := f.exchange(code)
	requireProtocolCode(t, err, CodeInvalidGrant)
}

func TestRefreshRotatesAndReplayRevokesTheWholeGrant(t *testing.T) {
	t.Parallel()

	f := newFixture(t)

	first, err := f.exchange(f.login())
	require.NoError(t, err)

	second, err := f.refreshWith(first.RefreshToken)
	require.NoError(t, err)
	require.NotEqual(t, first.RefreshToken, second.RefreshToken)
	require.Equal(t, first.Scope, second.Scope)

	claims, err := f.service.VerifyAccessToken(second.AccessToken)
	require.NoError(t, err)
	require.Equal(t, 1, f.repo.activeRefreshTokens(claims.GrantID))

	_, err = f.refreshWith(first.RefreshToken)
	requireProtocolCode(t, err, CodeInvalidGrant)
	require.Zero(t, f.repo.activeRefreshTokens(claims.GrantID))

	_, err = f.refreshWith(second.RefreshToken)
	requireProtocolCode(t, err, CodeInvalidGrant)
}

func TestRefreshCannotWidenScopeOrCrossClients(t *testing.T) {
	t.Parallel()

	f := newFixture(t)

	tokens, err := f.exchange(f.login())
	require.NoError(t, err)

	_, err = f.service.Token(context.Background(), f.tenant, TokenParams{
		GrantType: GrantTypeRefreshToken, ClientID: f.client.ClientID, ClientSecret: f.client.ClientSecret,
		RefreshToken: tokens.RefreshToken, Scope: "openid phone",
	})
	requireProtocolCode(t, err, CodeInvalidScope)

	other, err := f.service.CreateClient(context.Background(), testTenantID, "admin", CreateClientRequest{
		Name: "Other", RedirectURIs: []string{"https://other.acme.example/cb"},
	})
	require.NoError(t, err)

	_, err = f.service.Token(context.Background(), f.tenant, TokenParams{
		GrantType: GrantTypeRefreshToken, ClientID: other.ClientID, ClientSecret: other.ClientSecret,
		RefreshToken: tokens.RefreshToken,
	})
	requireProtocolCode(t, err, CodeInvalidGrant)

	narrowed, err := f.service.Token(context.Background(), f.tenant, TokenParams{
		GrantType: GrantTypeRefreshToken, ClientID: f.client.ClientID, ClientSecret: f.client.ClientSecret,
		RefreshToken: tokens.RefreshToken, Scope: "openid",
	})
	require.NoError(t, err, "a rejected attempt must not consume the refresh token")
	require.Equal(t, "openid", narrowed.Scope)
}

func TestAuthorizeRejectsUnverifiableClientsWithoutRedirecting(t *testing.T) {
	t.Parallel()

	f := newFixture(t)

	params := f.authorizeParams()
	params.RedirectURI = "https://attacker.example/cb"
	_, err := f.service.Authorize(context.Background(), f.tenant, params)
	requireProtocolCode(t, err, CodeInvalidRequest)

	params = f.authorizeParams()
	params.ClientID = "gdc-unknown"
	_, err = f.service.Authorize(context.Background(), f.tenant, params)
	requireProtocolCode(t, err, CodeInvalidClient)

	foreign := Tenant{ID: otherTenantID, Slug: testOtherSlug, Host: "globex.godar.example"}
	_, err = f.service.Authorize(context.Background(), foreign, f.authorizeParams())
	requireProtocolCode(t, err, CodeInvalidClient)
}

func TestAuthorizeRedirectsProtocolErrorsToTheClient(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		mutate func(*AuthorizeParams)
		code   string
	}{
		"missing pkce":   {mutate: func(p *AuthorizeParams) { p.CodeChallenge = "" }, code: CodeInvalidRequest},
		"plain pkce":     {mutate: func(p *AuthorizeParams) { p.CodeChallengeMethod = "plain" }, code: CodeInvalidRequest},
		"missing openid": {mutate: func(p *AuthorizeParams) { p.Scope = "profile" }, code: CodeInvalidScope},
		"implicit flow":  {mutate: func(p *AuthorizeParams) { p.ResponseType = "token" }, code: CodeUnsupportedResponse},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			f := newFixture(t)
			params := f.authorizeParams()
			tc.mutate(&params)

			result, err := f.service.Authorize(context.Background(), f.tenant, params)
			require.NoError(t, err)

			target, err := url.Parse(result.RedirectURL)
			require.NoError(t, err)
			require.True(t, strings.HasPrefix(result.RedirectURL, testRedirect))
			require.Equal(t, tc.code, target.Query().Get("error"))
			require.Equal(t, "state-123", target.Query().Get("state"))
		})
	}
}

func TestLoginBridgeEnforcesTenantSessionAndFreshness(t *testing.T) {
	t.Parallel()

	f := newFixture(t)
	requestID := f.startLogin(f.authorizeParams())

	_, err := f.service.GetLoginRequest(context.Background(), testOtherSlug, requestID)
	require.ErrorIs(t, err, ErrRequestNotFound)

	view, err := f.service.GetLoginRequest(context.Background(), testTenantSlug, requestID)
	require.NoError(t, err)
	require.Equal(t, "LibreChat", view.ClientName)

	f.sessions.sessions["foreign"] = &Session{ID: "foreign", TenantSlug: testOtherSlug, UserID: testUserID, CreatedAt: f.clock}
	_, err = f.service.AcceptLogin(context.Background(), testTenantSlug, requestID, "foreign")
	require.ErrorIs(t, err, ErrSessionInvalid)

	_, err = f.service.AcceptLogin(context.Background(), testTenantSlug, requestID, "missing")
	require.ErrorIs(t, err, ErrSessionInvalid)

	fresh := f.authorizeParams()
	fresh.Prompt = PromptLogin
	freshID := f.startLogin(fresh)
	_, err = f.service.AcceptLogin(context.Background(), testTenantSlug, freshID, testSessionID)
	require.ErrorIs(t, err, ErrSessionInvalid, "prompt=login needs a session created after the request")

	_, err = f.service.AcceptLogin(context.Background(), testTenantSlug, requestID, testSessionID)
	require.NoError(t, err)

	_, err = f.service.AcceptLogin(context.Background(), testTenantSlug, requestID, testSessionID)
	require.ErrorIs(t, err, ErrRequestNotPending)
}

func TestExpiredLoginRequestCannotBeAccepted(t *testing.T) {
	t.Parallel()

	f := newFixture(t)
	requestID := f.startLogin(f.authorizeParams())
	f.clock = f.clock.Add(11 * time.Minute)

	_, err := f.service.AcceptLogin(context.Background(), testTenantSlug, requestID, testSessionID)
	require.ErrorIs(t, err, ErrRequestNotPending)
}

func TestRejectLoginReturnsLoginRequiredForPromptNone(t *testing.T) {
	t.Parallel()

	f := newFixture(t)
	params := f.authorizeParams()
	params.Prompt = PromptNone

	decision, err := f.service.RejectLogin(context.Background(), testTenantSlug, f.startLogin(params))
	require.NoError(t, err)

	target, err := url.Parse(decision.RedirectTo)
	require.NoError(t, err)
	require.Equal(t, CodeLoginRequired, target.Query().Get("error"))
}

func TestDeactivatedMemberCannotCompleteLogin(t *testing.T) {
	t.Parallel()

	f := newFixture(t)
	f.repo.setIdentity(Identity{UserID: testUserID, TenantID: testTenantID, Active: false})

	decision, err := f.service.AcceptLogin(context.Background(), testTenantSlug, f.startLogin(f.authorizeParams()), testSessionID)
	require.NoError(t, err)

	target, err := url.Parse(decision.RedirectTo)
	require.NoError(t, err)
	require.Equal(t, CodeAccessDenied, target.Query().Get("error"))
	require.Empty(t, target.Query().Get("code"))
}

func TestInactiveClientCannotExchangeOrAuthorize(t *testing.T) {
	t.Parallel()

	f := newFixture(t)
	code := f.login()

	inactive := false
	_, err := f.service.UpdateClient(context.Background(), testTenantID, f.client.ID, UpdateClientRequest{Active: &inactive})
	require.NoError(t, err)

	_, err = f.exchange(code)
	requireProtocolCode(t, err, CodeInvalidClient)

	_, err = f.service.Authorize(context.Background(), f.tenant, f.authorizeParams())
	requireProtocolCode(t, err, CodeInvalidClient)
}

func TestRotatedClientSecretReplacesTheOldOne(t *testing.T) {
	t.Parallel()

	f := newFixture(t)
	code := f.login()
	oldSecret := f.client.ClientSecret

	rotated, err := f.service.RotateClientSecret(context.Background(), testTenantID, f.client.ID)
	require.NoError(t, err)
	require.NotEqual(t, oldSecret, rotated.ClientSecret)

	f.client.ClientSecret = oldSecret
	_, err = f.exchange(code)
	requireProtocolCode(t, err, CodeInvalidClient)
}

func TestLogoutRevokesTheClientGrantAndValidatesRedirect(t *testing.T) {
	t.Parallel()

	f := newFixture(t)

	tokens, err := f.exchange(f.login())
	require.NoError(t, err)

	_, err = f.service.Logout(context.Background(), f.tenant, LogoutParams{
		IDTokenHint: tokens.IDToken, PostLogoutRedirectURI: "https://attacker.example/",
	})
	requireProtocolCode(t, err, CodeInvalidRequest)

	redirectTo, err := f.service.Logout(context.Background(), f.tenant, LogoutParams{
		IDTokenHint: tokens.IDToken, PostLogoutRedirectURI: testLogoutURI, State: "bye",
	})
	require.NoError(t, err)
	require.Equal(t, testLogoutURI+"?state=bye", redirectTo)

	_, err = f.refreshWith(tokens.RefreshToken)
	requireProtocolCode(t, err, CodeInvalidGrant)
}

func TestClientRegistrationValidatesRedirectURIs(t *testing.T) {
	t.Parallel()

	f := newFixture(t)

	for _, uri := range []string{
		"http://chat.acme.example/cb",
		"https://chat.acme.example/cb#fragment",
		"/relative",
		"https://user:pass@chat.acme.example/cb",
	} {
		_, err := f.service.CreateClient(context.Background(), testTenantID, "admin", CreateClientRequest{
			Name: "Bad", RedirectURIs: []string{uri},
		})
		require.ErrorIs(t, err, ErrInvalidClientConfig, uri)
	}

	created, err := f.service.CreateClient(context.Background(), testTenantID, "admin", CreateClientRequest{
		Name: "Local", RedirectURIs: []string{"http://localhost:3080/oauth/openid/callback"},
	})
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(created.ClientID, "gdc-"))
	require.True(t, strings.HasPrefix(created.ClientSecret, "gds_"))
	require.NotContains(t, created.SecretHash, created.ClientSecret)
}

func TestDiscoveryAdvertisesTenantIssuer(t *testing.T) {
	t.Parallel()

	f := newFixture(t)
	metadata := f.service.Discovery(f.tenant)

	require.Equal(t, "https://acme.godar.example", metadata["issuer"])
	require.Equal(t, "https://acme.godar.example/oidc/token", metadata["token_endpoint"])
	require.Equal(t, []string{CodeChallengeMethodS256}, metadata["code_challenge_methods_supported"])
}
