package oidc

import (
	"encoding/base64"
	"net/url"
	"strings"

	"github.com/Delpak/shahrah/services/control-plane/internal/oidc"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func (r *Routes) requestTenant(c *fiber.Ctx) (oidc.Tenant, bool) {
	current := r.tenant(c)
	if current == nil || strings.TrimSpace(current.ID) == "" {
		return oidc.Tenant{}, false
	}

	return oidc.Tenant{
		ID:   current.ID,
		Slug: current.Slug,
		Host: strings.ToLower(strings.TrimSpace(c.Hostname())),
	}, true
}

func (r *Routes) discovery(c *fiber.Ctx) error {
	tenant, ok := r.requestTenant(c)
	if !ok {
		return writeProtocolError(c, &oidc.Error{Code: oidc.CodeInvalidRequest, Description: "unknown tenant", Status: fiber.StatusNotFound})
	}

	c.Set(fiber.HeaderCacheControl, "public, max-age=300")

	return c.JSON(r.service.Discovery(tenant))
}

func (r *Routes) jwks(c *fiber.Ctx) error {
	c.Set(fiber.HeaderCacheControl, "public, max-age=300")

	return c.JSON(r.service.JWKS())
}

func (r *Routes) authorize(c *fiber.Ctx) error {
	noStore(c)

	tenant, ok := r.requestTenant(c)
	if !ok {
		return c.Status(fiber.StatusNotFound).Type("txt").SendString("Unknown tenant.")
	}

	result, err := r.service.Authorize(c.Context(), tenant, oidc.AuthorizeParams{
		ClientID:            requestParam(c, "client_id"),
		RedirectURI:         requestParam(c, "redirect_uri"),
		ResponseType:        requestParam(c, "response_type"),
		Scope:               requestParam(c, "scope"),
		State:               requestParam(c, "state"),
		Nonce:               requestParam(c, "nonce"),
		CodeChallenge:       requestParam(c, "code_challenge"),
		CodeChallengeMethod: requestParam(c, "code_challenge_method"),
		Prompt:              requestParam(c, "prompt"),
		ResponseMode:        requestParam(c, "response_mode"),
	})
	if err != nil {
		if protocolErr, isProtocol := oidc.AsError(err); isProtocol {
			// The client could not be verified, so the error must not be redirected to it.
			return c.Status(protocolErr.Status).Type("txt").SendString("Sign-in request rejected: " + protocolErr.Description)
		}

		r.logger.Error("oidc authorize failed", zap.Error(err), zap.String("tenant_id", tenant.ID))

		return c.Status(fiber.StatusInternalServerError).Type("txt").SendString("Sign-in is temporarily unavailable.")
	}

	return c.Redirect(result.RedirectURL, fiber.StatusFound)
}

func (r *Routes) token(c *fiber.Ctx) error {
	noStore(c)

	tenant, ok := r.requestTenant(c)
	if !ok {
		return writeProtocolError(c, &oidc.Error{Code: oidc.CodeInvalidRequest, Description: "unknown tenant", Status: fiber.StatusNotFound})
	}

	params := oidc.TokenParams{
		GrantType:    c.FormValue("grant_type"),
		Code:         c.FormValue("code"),
		RedirectURI:  c.FormValue("redirect_uri"),
		CodeVerifier: c.FormValue("code_verifier"),
		RefreshToken: c.FormValue("refresh_token"),
		Scope:        c.FormValue("scope"),
	}

	basicID, basicSecret, hasBasic, basicErr := basicClientCredentials(c.Get(fiber.HeaderAuthorization))
	if basicErr != nil {
		return writeProtocolError(c, &oidc.Error{Code: oidc.CodeInvalidClient, Description: "malformed client credentials", Status: fiber.StatusUnauthorized})
	}

	if hasBasic {
		if c.FormValue("client_secret") != "" {
			return writeProtocolError(c, &oidc.Error{Code: oidc.CodeInvalidRequest, Description: "use only one client authentication method", Status: fiber.StatusBadRequest})
		}

		params.ClientID = basicID
		params.ClientSecret = basicSecret
	} else {
		params.ClientID = c.FormValue("client_id")
		params.ClientSecret = c.FormValue("client_secret")
	}

	response, err := r.service.Token(c.Context(), tenant, params)
	if err != nil {
		if protocolErr, isProtocol := oidc.AsError(err); isProtocol {
			if protocolErr.Code == oidc.CodeInvalidClient && hasBasic {
				c.Set(fiber.HeaderWWWAuthenticate, `Basic realm="oidc"`)
			}

			return writeProtocolError(c, protocolErr)
		}

		r.logger.Error("oidc token request failed", zap.Error(err), zap.String("tenant_id", tenant.ID))

		return writeProtocolError(c, &oidc.Error{Code: oidc.CodeServerError, Status: fiber.StatusInternalServerError})
	}

	return c.JSON(response)
}

func (r *Routes) userInfo(c *fiber.Ctx) error {
	noStore(c)

	tenant, ok := r.requestTenant(c)
	if !ok {
		return writeProtocolError(c, &oidc.Error{Code: oidc.CodeInvalidRequest, Description: "unknown tenant", Status: fiber.StatusNotFound})
	}

	token, found := strings.CutPrefix(c.Get(fiber.HeaderAuthorization), "Bearer ")
	if !found || strings.TrimSpace(token) == "" {
		c.Set(fiber.HeaderWWWAuthenticate, `Bearer realm="oidc"`)

		return writeProtocolError(c, &oidc.Error{Code: oidc.CodeInvalidToken, Description: "bearer access token required", Status: fiber.StatusUnauthorized})
	}

	info, err := r.service.UserInfo(c.Context(), tenant, token)
	if err != nil {
		if protocolErr, isProtocol := oidc.AsError(err); isProtocol {
			c.Set(fiber.HeaderWWWAuthenticate, `Bearer realm="oidc", error="invalid_token"`)

			return writeProtocolError(c, protocolErr)
		}

		r.logger.Error("oidc userinfo failed", zap.Error(err), zap.String("tenant_id", tenant.ID))

		return writeProtocolError(c, &oidc.Error{Code: oidc.CodeServerError, Status: fiber.StatusInternalServerError})
	}

	return c.JSON(info)
}

func (r *Routes) logout(c *fiber.Ctx) error {
	noStore(c)

	tenant, ok := r.requestTenant(c)
	if !ok {
		return c.Status(fiber.StatusNotFound).Type("txt").SendString("Unknown tenant.")
	}

	redirectTo, err := r.service.Logout(c.Context(), tenant, oidc.LogoutParams{
		IDTokenHint:           requestParam(c, "id_token_hint"),
		ClientID:              requestParam(c, "client_id"),
		PostLogoutRedirectURI: requestParam(c, "post_logout_redirect_uri"),
		State:                 requestParam(c, "state"),
	})
	if err != nil {
		if protocolErr, isProtocol := oidc.AsError(err); isProtocol {
			return c.Status(protocolErr.Status).Type("txt").SendString("Sign-out request rejected: " + protocolErr.Description)
		}

		r.logger.Error("oidc logout failed", zap.Error(err), zap.String("tenant_id", tenant.ID))

		return c.Status(fiber.StatusInternalServerError).Type("txt").SendString("Sign-out is temporarily unavailable.")
	}

	if redirectTo == "" {
		return c.Type("txt").SendString("You have been signed out of this application.")
	}

	return c.Redirect(redirectTo, fiber.StatusFound)
}

func requestParam(c *fiber.Ctx, name string) string {
	if c.Method() == fiber.MethodPost {
		return c.FormValue(name)
	}

	return c.Query(name)
}

// basicClientCredentials decodes RFC 6749 section 2.3.1 client_secret_basic credentials.
func basicClientCredentials(header string) (string, string, bool, error) {
	encoded, found := strings.CutPrefix(header, "Basic ")
	if !found {
		return "", "", false, nil
	}

	raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(encoded))
	if err != nil {
		return "", "", true, err
	}

	rawID, rawSecret, found := strings.Cut(string(raw), ":")
	if !found {
		return "", "", true, url.EscapeError("missing separator")
	}

	clientID, err := url.QueryUnescape(rawID)
	if err != nil {
		return "", "", true, err
	}

	secret, err := url.QueryUnescape(rawSecret)
	if err != nil {
		return "", "", true, err
	}

	return clientID, secret, true, nil
}

func noStore(c *fiber.Ctx) {
	c.Set(fiber.HeaderCacheControl, "no-store")
	c.Set(fiber.HeaderPragma, "no-cache")
}

func writeProtocolError(c *fiber.Ctx, protocolErr *oidc.Error) error {
	body := fiber.Map{"error": protocolErr.Code}
	if protocolErr.Description != "" {
		body["error_description"] = protocolErr.Description
	}

	return c.Status(protocolErr.Status).JSON(body)
}
