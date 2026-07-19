package main

import (
	"testing"

	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
)

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
