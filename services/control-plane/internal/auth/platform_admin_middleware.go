package auth

import (
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// PlatformAdminAuthMiddleware authenticates platform admin routes using either
// the configured master key or a scoped platform service account token.
func PlatformAdminAuthMiddleware(masterKey string, serviceAccounts PlatformServiceAccountStore) fiber.Handler {
	return func(c *fiber.Ctx) error {
		apiKey := ExtractPlatformAdminAPIKey(c)
		if apiKey == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "platform admin authentication required")
		}

		if strings.TrimSpace(masterKey) != "" && secureTokenEqual(apiKey, masterKey) {
			return c.Next()
		}

		if serviceAccounts == nil {
			return fiber.NewError(fiber.StatusUnauthorized, "platform admin authentication required")
		}

		credential, err := serviceAccounts.GetPlatformServiceAccountByTokenHash(c.Context(), HashKey(apiKey))
		if errors.Is(err, ErrInvalidPlatformServiceCredential) {
			return fiber.NewError(fiber.StatusUnauthorized, "platform admin authentication required")
		}

		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "platform service authentication lookup failed")
		}

		requiredScope := RequiredPlatformScope(c.Method(), c.Path())
		if credential == nil || !HasPlatformScope(credential.Scopes, requiredScope) {
			return fiber.NewError(fiber.StatusForbidden, "platform service account lacks required scope")
		}

		return c.Next()
	}
}

func RequiredPlatformScope(method string, path string) string {
	if method == fiber.MethodGet || method == fiber.MethodHead || method == fiber.MethodOptions {
		return PlatformScopeRead
	}

	if strings.Contains(path, "/tenant-subscriptions") || strings.Contains(path, "/tenant-entitlement-overrides") || strings.Contains(path, "/tenants/") {
		return PlatformScopeTenantWrite
	}

	return PlatformScopeCatalogWrite
}

func ExtractPlatformAdminAPIKey(c *fiber.Ctx) string {
	authHeader := strings.TrimSpace(c.Get("Authorization"))
	if authHeader != "" {
		if strings.HasPrefix(authHeader, "Bearer ") {
			return strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		}

		return authHeader
	}

	if key := strings.TrimSpace(c.Get("api-key")); key != "" {
		return key
	}

	return strings.TrimSpace(c.Get("x-api-key"))
}

func secureTokenEqual(left string, right string) bool {
	leftHash := sha256.Sum256([]byte(left))
	rightHash := sha256.Sum256([]byte(right))

	return subtle.ConstantTimeCompare(leftHash[:], rightHash[:]) == 1
}
