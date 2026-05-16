package storage

import "testing"

func TestIsValidEmailAddress(t *testing.T) {
	tests := []struct {
		name  string
		email string
		valid bool
	}{
		{name: "valid simple", email: "user@example.com", valid: true},
		{name: "valid with uppercase and spaces", email: "  User.Name+tag@Example.IR  ", valid: true},
		{name: "missing tld", email: "admin@tenant", valid: false},
		{name: "missing domain", email: "admin@", valid: false},
		{name: "missing local part", email: "@tenant.com", valid: false},
		{name: "contains space", email: "admin tenant.com", valid: false},
		{name: "domain ends dot", email: "admin@tenant.", valid: false},
		{name: "domain starts dot label", email: "admin@.com", valid: false},
		{name: "empty", email: "", valid: false},
		{name: "whitespace only", email: "   ", valid: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidEmailAddress(tt.email)
			if got != tt.valid {
				t.Fatalf("IsValidEmailAddress(%q) = %v, want %v", tt.email, got, tt.valid)
			}
		})
	}
}
