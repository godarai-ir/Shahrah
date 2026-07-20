package main

import (
	"slices"
	"testing"

	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	"github.com/gofiber/fiber/v2"
)

func TestSupportAdminProfileContainsEveryRequiredScope(t *testing.T) {
	t.Parallel()

	name, scopes, err := resolveCredentialConfig("", supportAdminProfile, "")
	if err != nil {
		t.Fatalf("resolveCredentialConfig returned error: %v", err)
	}
	if name != supportAdminProfile {
		t.Fatalf("got account name %q want %q", name, supportAdminProfile)
	}

	required := []struct {
		method string
		path   string
	}{
		{fiber.MethodGet, "/api/v1/platform/providers"},
		{fiber.MethodPost, "/api/v1/platform/providers"},
		{fiber.MethodPost, "/api/v1/platform/tenants/id/suspend"},
		{fiber.MethodGet, "/api/v1/platform/provider-connections"},
		{fiber.MethodPost, "/api/v1/platform/provider-connections"},
		{fiber.MethodPost, "/api/v1/platform/provider-connections/id/credential-versions/1/test"},
		{fiber.MethodPost, "/api/v1/platform/provider-connections/id/credential-versions/1/activate"},
		{fiber.MethodPost, "/api/v1/platform/provider-connections/id/disable"},
	}
	requiredScopes := make(map[string]struct{}, len(required))
	for _, endpoint := range required {
		scope := auth.RequiredPlatformScope(endpoint.method, endpoint.path)
		requiredScopes[scope] = struct{}{}
		if !slices.Contains(scopes, scope) {
			t.Errorf("support-admin profile is missing %q required by %s %s", scope, endpoint.method, endpoint.path)
		}
	}
	if len(scopes) != len(requiredScopes) {
		t.Errorf("support-admin profile has unexpected scopes: got %v", scopes)
	}
}

func TestResolveCredentialConfigRejectsProfileScopeOverride(t *testing.T) {
	t.Parallel()

	if _, _, err := resolveCredentialConfig("support-admin", supportAdminProfile, auth.PlatformScopeRead); err == nil {
		t.Fatal("expected --profile with --scopes to fail")
	}
}

func TestResolveCredentialConfigRequiresProfileForSupportAdminNames(t *testing.T) {
	t.Parallel()

	for _, name := range []string{"support-admin", "platform-support"} {
		if _, _, err := resolveCredentialConfig(name, "", auth.PlatformScopeRead); err == nil {
			t.Errorf("expected manually scoped %q credential to fail", name)
		}
	}
}

func TestParseScopesAllowsProviderCredentialScopes(t *testing.T) {
	t.Parallel()

	scopes, err := parseScopes(auth.PlatformScopeRead + "," + auth.PlatformScopeProviderCredentialCreateVersion)
	if err != nil {
		t.Fatalf("parseScopes returned error: %v", err)
	}

	want := []string{auth.PlatformScopeProviderCredentialCreateVersion, auth.PlatformScopeRead}
	if len(scopes) != len(want) {
		t.Fatalf("got %v want %v", scopes, want)
	}
	for index := range want {
		if scopes[index] != want[index] {
			t.Fatalf("got %v want %v", scopes, want)
		}
	}
}
