package auth

import (
	"strings"

	"github.com/Delpak/shahrah/services/control-plane/internal/oidc"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func looksLikeJWT(token string) bool {
	return oidc.LooksLikeJWT(token)
}

// authenticateAccessToken accepts an OIDC access token issued by this tenant's issuer.
// Access tokens are a gateway credential only; management APIs keep requiring API keys.
func (m *AuthMiddleware) authenticateAccessToken(c *fiber.Ctx, token string) error {
	// Fiber routes are case-insensitive, so compare the lowered path.
	if strings.HasPrefix(strings.ToLower(normalizeRoute(c.Path())), "/api/") {
		return writeAuthError(c, fiber.StatusForbidden, "Access tokens cannot call management APIs", "authorization_error")
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil || strings.TrimSpace(currentTenant.ID) == "" {
		return writeAuthError(c, fiber.StatusForbidden, "Tenant context is required for access-token authentication", "authorization_error")
	}

	resolved, err := m.accessTokens.AuthenticateAccessToken(c.Context(), currentTenant.ID, token)
	if err != nil {
		if oidc.IsProtocolError(err) {
			return writeAuthError(c, fiber.StatusUnauthorized, "Invalid or expired access token", "authentication_error")
		}

		m.logger.Error("access token authentication failed", zap.Error(err), zap.String("tenant_id", currentTenant.ID))

		return writeAuthError(c, fiber.StatusInternalServerError, "Authentication lookup failed", "server_error")
	}

	c.Locals(AuthContextKey{}, resolved)

	if err := CheckRouteAuthorization(c.Method(), c.Path(), resolved); err != nil {
		return writeAuthError(c, fiber.StatusForbidden, err.Error(), "authorization_error")
	}

	return c.Next()
}
