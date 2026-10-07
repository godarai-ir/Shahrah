package oidc

import (
	"context"
	"errors"
	"net/url"
	"slices"
	"strings"

	"github.com/google/uuid"
)

// Authorize validates an authorization request and returns where to send the browser.
//
// Errors that cannot be trusted to reach the client (unknown client, unregistered
// redirect URI) are returned as *Error and must be shown to the user directly. Every
// other failure is delivered to the client's redirect URI, per RFC 6749 section 4.1.2.1.
func (s *Service) Authorize(ctx context.Context, tenant Tenant, params AuthorizeParams) (*AuthorizeResult, error) {
	client, err := s.activeClient(ctx, tenant, strings.TrimSpace(params.ClientID))
	if err != nil {
		return nil, err
	}

	redirectURI := strings.TrimSpace(params.RedirectURI)
	if redirectURI == "" || !slices.Contains(client.RedirectURIs, redirectURI) {
		return nil, protocolError(CodeInvalidRequest, "redirect_uri is not registered for this client")
	}

	issuer := s.Issuer(tenant)

	request, protocolErr := s.validateAuthorizeParams(params)
	if protocolErr != nil {
		return redirectWithError(redirectURI, protocolErr, params.State, issuer)
	}

	id, err := randomToken("")
	if err != nil {
		return nil, err
	}

	now := s.now()
	request.ID = id
	request.TenantID = tenant.ID
	request.TenantSlug = tenant.Slug
	request.Issuer = issuer
	request.ClientID = client.ClientID
	request.RedirectURI = redirectURI
	request.Status = RequestStatusPending
	request.CreatedAt = now
	request.ExpiresAt = now.Add(s.settings.AuthorizationRequestTTL)

	if createErr := s.requests.CreateOIDCAuthorizationRequest(ctx, *request); createErr != nil {
		return nil, createErr
	}

	loginURL, err := appendQuery(s.loginURL(tenant), url.Values{"oidc_request": {id}})
	if err != nil {
		return nil, err
	}

	return &AuthorizeResult{RedirectURL: loginURL}, nil
}

func (s *Service) validateAuthorizeParams(params AuthorizeParams) (*AuthorizationRequest, *Error) {
	if strings.TrimSpace(params.ResponseType) != ResponseTypeCode {
		return nil, protocolError(CodeUnsupportedResponse, "only response_type=code is supported")
	}

	if mode := strings.TrimSpace(params.ResponseMode); mode != "" && mode != "query" {
		return nil, protocolError(CodeInvalidRequest, "only response_mode=query is supported")
	}

	scope, err := normalizeScope(params.Scope)
	if err != nil {
		protocolErr, _ := AsError(err)

		return nil, protocolErr
	}

	challenge := strings.TrimSpace(params.CodeChallenge)
	if strings.TrimSpace(params.CodeChallengeMethod) != CodeChallengeMethodS256 || !validCodeChallenge(challenge) {
		return nil, protocolError(CodeInvalidRequest, "PKCE with code_challenge_method=S256 is required")
	}

	if len(params.State) > maxStateLength || len(params.Nonce) > maxNonceLength {
		return nil, protocolError(CodeInvalidRequest, "state or nonce is too long")
	}

	prompt := ""

	for _, value := range strings.Fields(params.Prompt) {
		switch value {
		case PromptNone, PromptLogin:
			prompt = value
		case "consent", "select_account":
		default:
			return nil, protocolError(CodeInvalidRequest, "unsupported prompt value")
		}
	}

	return &AuthorizationRequest{
		Scope:               scope,
		State:               params.State,
		Nonce:               params.Nonce,
		CodeChallenge:       challenge,
		CodeChallengeMethod: CodeChallengeMethodS256,
		Prompt:              prompt,
	}, nil
}

// GetLoginRequest returns a pending request for the tenant-admin login page.
func (s *Service) GetLoginRequest(ctx context.Context, tenantSlug, requestID string) (*LoginRequest, error) {
	request, err := s.pendingRequest(ctx, tenantSlug, requestID)
	if err != nil {
		return nil, err
	}

	client, err := s.clients.GetOIDCClient(ctx, request.ClientID)
	if err != nil {
		return nil, err
	}

	return &LoginRequest{
		ID:         request.ID,
		TenantSlug: request.TenantSlug,
		ClientID:   client.ClientID,
		ClientName: client.Name,
		Scopes:     strings.Fields(request.Scope),
		Prompt:     request.Prompt,
		ExpiresAt:  request.ExpiresAt,
	}, nil
}

// AcceptLogin completes a pending request for the member who owns sessionID and returns
// the client redirect carrying a single-use authorization code.
func (s *Service) AcceptLogin(ctx context.Context, tenantSlug, requestID, sessionID string) (*LoginDecision, error) {
	request, err := s.pendingRequest(ctx, tenantSlug, requestID)
	if err != nil {
		return nil, err
	}

	session, err := s.sessions.LookupSession(ctx, strings.TrimSpace(sessionID))
	if err != nil {
		return nil, err
	}

	if session == nil || !strings.EqualFold(session.TenantSlug, request.TenantSlug) || session.UserID == "" {
		return nil, ErrSessionInvalid
	}

	if request.Prompt == PromptLogin && session.CreatedAt.Before(request.CreatedAt) {
		return nil, ErrSessionInvalid
	}

	if _, err = s.activeIdentity(ctx, request.TenantID, session.UserID); err != nil {
		if errors.Is(err, ErrIdentityNotFound) {
			return s.rejectRequest(ctx, request, protocolError(CodeAccessDenied, "the account is not allowed to sign in"))
		}

		return nil, err
	}

	code, err := randomToken("gdc_")
	if err != nil {
		return nil, err
	}

	now := s.now()
	accepted := *request
	accepted.Status = RequestStatusAccepted
	accepted.UserID = session.UserID
	accepted.GrantID = uuid.NewString()

	accepted.AuthTime = session.CreatedAt.UTC()
	if accepted.AuthTime.IsZero() || accepted.AuthTime.After(now) {
		accepted.AuthTime = now
	}

	if acceptErr := s.requests.AcceptOIDCAuthorizationRequest(ctx, accepted, hashToken(code), now.Add(s.settings.CodeTTL)); acceptErr != nil {
		return nil, acceptErr
	}

	redirectTo, err := appendQuery(request.RedirectURI, url.Values{
		"code":  {code},
		"state": {request.State},
		"iss":   {request.Issuer},
	})
	if err != nil {
		return nil, err
	}

	return &LoginDecision{RedirectTo: redirectTo}, nil
}

// RejectLogin ends a pending request with access_denied, or login_required for prompt=none.
func (s *Service) RejectLogin(ctx context.Context, tenantSlug, requestID string) (*LoginDecision, error) {
	request, err := s.pendingRequest(ctx, tenantSlug, requestID)
	if err != nil {
		return nil, err
	}

	reason := protocolError(CodeAccessDenied, "the user declined to sign in")
	if request.Prompt == PromptNone {
		reason = protocolError(CodeLoginRequired, "the user is not signed in")
	}

	return s.rejectRequest(ctx, request, reason)
}

func (s *Service) rejectRequest(ctx context.Context, request *AuthorizationRequest, reason *Error) (*LoginDecision, error) {
	if err := s.requests.RejectOIDCAuthorizationRequest(ctx, request.ID); err != nil {
		return nil, err
	}

	result, err := redirectWithError(request.RedirectURI, reason, request.State, request.Issuer)
	if err != nil {
		return nil, err
	}

	return &LoginDecision{RedirectTo: result.RedirectURL}, nil
}

func (s *Service) pendingRequest(ctx context.Context, tenantSlug, requestID string) (*AuthorizationRequest, error) {
	requestID = strings.TrimSpace(requestID)
	if requestID == "" {
		return nil, ErrRequestNotFound
	}

	request, err := s.requests.GetOIDCAuthorizationRequest(ctx, requestID)
	if err != nil {
		return nil, err
	}

	if !strings.EqualFold(request.TenantSlug, strings.TrimSpace(tenantSlug)) {
		return nil, ErrRequestNotFound
	}

	if request.Status != RequestStatusPending || !s.now().Before(request.ExpiresAt) {
		return nil, ErrRequestNotPending
	}

	return request, nil
}

func (s *Service) activeClient(ctx context.Context, tenant Tenant, clientID string) (*Client, error) {
	if clientID == "" {
		return nil, protocolError(CodeInvalidClient, "unknown client")
	}

	client, err := s.clients.GetOIDCClient(ctx, clientID)
	if err != nil {
		if errors.Is(err, ErrClientNotFound) {
			return nil, protocolError(CodeInvalidClient, "unknown client")
		}

		return nil, err
	}

	if !client.Active || !strings.EqualFold(client.TenantID, tenant.ID) {
		return nil, protocolError(CodeInvalidClient, "unknown client")
	}

	return client, nil
}

func (s *Service) activeIdentity(ctx context.Context, tenantID, userID string) (*Identity, error) {
	identity, err := s.identities.GetOIDCIdentity(ctx, tenantID, userID)
	if err != nil {
		return nil, err
	}

	if identity == nil || !identity.Active {
		return nil, ErrIdentityNotFound
	}

	return identity, nil
}

func (s *Service) loginURL(tenant Tenant) string {
	return strings.NewReplacer("{host}", tenant.Host, "{slug}", tenant.Slug).Replace(s.settings.LoginURLTemplate)
}

func redirectWithError(redirectURI string, reason *Error, state, issuer string) (*AuthorizeResult, error) {
	target, err := appendQuery(redirectURI, url.Values{
		"error":             {reason.Code},
		"error_description": {reason.Description},
		"state":             {state},
		"iss":               {issuer},
	})
	if err != nil {
		return nil, err
	}

	return &AuthorizeResult{RedirectURL: target}, nil
}
