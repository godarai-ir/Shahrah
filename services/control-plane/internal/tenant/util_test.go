package tenant

import "testing"

func TestValidateTenantSlugRejectsInfrastructureHosts(t *testing.T) {
	t.Parallel()

	for _, slug := range []string{"admin", "auth"} {
		slug := slug
		t.Run(slug, func(t *testing.T) {
			t.Parallel()

			if ValidateTenantSlug(slug) {
				t.Fatalf("ValidateTenantSlug(%q) = true, want false", slug)
			}
		})
	}
}

func TestValidateTenantSlugAllowsOrdinaryTenant(t *testing.T) {
	t.Parallel()

	if !ValidateTenantSlug("acme-ai") {
		t.Fatal("ValidateTenantSlug(\"acme-ai\") = false, want true")
	}
}
