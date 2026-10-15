package oidc

import "testing"

func TestApplyProfileClaimsReleasesNames(t *testing.T) {
	t.Parallel()

	identity := &Identity{
		UserID: "u1", Email: "ali@acme.example", Name: "alias",
		FirstName: "علی", LastName: "رضایی",
	}

	var claims idTokenClaims

	applyProfileClaims(&claims, identity, "openid profile email")

	if claims.GivenName != "علی" || claims.FamilyName != "رضایی" {
		t.Fatalf("given/family = %q/%q", claims.GivenName, claims.FamilyName)
	}

	if claims.Name != "علی رضایی" {
		t.Fatalf("name = %q, want the full name", claims.Name)
	}
}

func TestApplyProfileClaimsFallsBackWithoutNames(t *testing.T) {
	t.Parallel()

	var claims idTokenClaims

	applyProfileClaims(&claims, &Identity{UserID: "u1", Email: "ali@acme.example"}, "openid profile")

	if claims.GivenName != "" || claims.FamilyName != "" {
		t.Fatalf("unexpected names %q/%q", claims.GivenName, claims.FamilyName)
	}

	if claims.Name != "ali@acme.example" {
		t.Fatalf("name = %q, want the email", claims.Name)
	}
}

func TestApplyProfileClaimsHidesNamesWithoutProfileScope(t *testing.T) {
	t.Parallel()

	var claims idTokenClaims

	applyProfileClaims(&claims, &Identity{UserID: "u1", FirstName: "علی", LastName: "رضایی"}, "openid email")

	if claims.GivenName != "" || claims.FamilyName != "" || claims.Name != "" {
		t.Fatalf("names leaked without profile scope: %+v", claims)
	}
}
