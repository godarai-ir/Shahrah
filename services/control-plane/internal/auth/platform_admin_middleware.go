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
			setPlatformServiceAccountContext(c, PlatformServiceAccountContext{ID: "master-key"})

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

		setPlatformServiceAccountContext(c, PlatformServiceAccountContext{
			ID:   credential.CredentialID,
			Name: credential.AccountName,
		})

		return c.Next()
	}
}

func RequiredPlatformScope(method, path string) string {
	// Fiber routes case-insensitively, so scope selection must too.
	path = strings.ToLower(path)

	if strings.Contains(path, "/support/tickets") || strings.Contains(path, "/support/ticket-attachments") {
		if method == fiber.MethodGet || method == fiber.MethodHead || method == fiber.MethodOptions {
			return PlatformScopeSupportTicketRead
		}

		return PlatformScopeSupportTicketWrite
	}

	if strings.Contains(path, "/provider-connections") || strings.Contains(path, "/provider-connection-schemas") {
		if method == fiber.MethodGet || method == fiber.MethodHead || method == fiber.MethodOptions {
			return PlatformScopeProviderCredentialRead
		}

		if strings.Contains(path, "/test") {
			return PlatformScopeProviderCredentialTest
		}

		if strings.Contains(path, "/activate") || strings.Contains(path, "/rollback") {
			return PlatformScopeProviderCredentialActivate
		}

		if strings.Contains(path, "/disable") || strings.Contains(path, "/revoke") {
			return PlatformScopeProviderCredentialDisable
		}

		return PlatformScopeProviderCredentialCreateVersion
	}

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

func secureTokenEqual(left, right string) bool {
	leftHash := sha256.Sum256([]byte(left))
	rightHash := sha256.Sum256([]byte(right))

	return subtle.ConstantTimeCompare(leftHash[:], rightHash[:]) == 1
}
