package auth

import (
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestRequiredPlatformScopeForProviderCredentialLifecycle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		method string
		path   string
		want   string
	}{
		{fiber.MethodGet, "/api/v1/platform/provider-connection-schemas/openai", PlatformScopeProviderCredentialRead},
		{fiber.MethodGet, "/api/v1/platform/provider-connections", PlatformScopeProviderCredentialRead},
		{fiber.MethodPost, "/api/v1/platform/provider-connections", PlatformScopeProviderCredentialCreateVersion},
		{fiber.MethodPost, "/api/v1/platform/provider-connections/id/credential-versions", PlatformScopeProviderCredentialCreateVersion},
		{fiber.MethodPost, "/api/v1/platform/provider-connections/id/credential-versions/1/test", PlatformScopeProviderCredentialTest},
		{fiber.MethodPost, "/api/v1/platform/provider-connections/id/credential-versions/1/activate", PlatformScopeProviderCredentialActivate},
		{fiber.MethodPost, "/api/v1/platform/provider-connections/id/rollback", PlatformScopeProviderCredentialActivate},
		{fiber.MethodPost, "/api/v1/platform/provider-connections/id/disable", PlatformScopeProviderCredentialDisable},
		{fiber.MethodPost, "/api/v1/platform/provider-connections/id/credential-versions/1/revoke", PlatformScopeProviderCredentialDisable},
		{fiber.MethodPost, "/api/v1/platform/provider-catalog/reconcile", PlatformScopeCatalogWrite},
	}
	for _, test := range tests {
		if got := RequiredPlatformScope(test.method, test.path); got != test.want {
			t.Fatalf("%s %s: got %q want %q", test.method, test.path, got, test.want)
		}
	}
}
