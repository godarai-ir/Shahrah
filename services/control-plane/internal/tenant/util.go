package tenant

import (
	"net"
	"strings"
)

var reservedTenantSlugs = map[string]struct{}{
	"admin":     {},
	"api":       {},
	"app":       {},
	"billing":   {},
	"default":   {},
	"help":      {},
	"localhost": {},
	"login":     {},
	"root":      {},
	"shahrah":   {},
	"signup":    {},
	"support":   {},
	"system":    {},
	"www":       {},
}

// ExtractSubdomain extracts the tenant slug from a hostname.
// Examples:
//
//	"acme.sharah.ir" -> "acme"
//	"my-company.sharah.ir" -> "my-company"
//	"localhost" -> ""
//	"127.0.0.1" -> ""
func ExtractSubdomain(host string) string {
	// Remove port if present
	if idx := strings.LastIndex(host, ":"); idx != -1 {
		// Only treat as port if it looks like a numeric port (not part of IPv6)
		rest := host[idx+1:]
		if isNumericPort(rest) {
			host = host[:idx]
		}
	}

	// Handle IP addresses
	if ip := net.ParseIP(host); ip != nil {
		return ""
	}

	// Split by dots
	parts := strings.Split(host, ".")
	if len(parts) < 3 {
		return "" // No subdomain (e.g., "localhost", "127.0.0.1")
	}

	return parts[0] // First part is the subdomain
}

// ValidateTenantSlug checks if a slug is valid.
// Valid slugs: 1-63 chars, lowercase alphanumeric and hyphens,
// must not start or end with hyphen.
func ValidateTenantSlug(slug string) bool {
	if len(slug) < 1 || len(slug) > 63 {
		return false
	}

	if _, reserved := reservedTenantSlugs[slug]; reserved {
		return false
	}

	for _, c := range slug {
		if (c < 'a' || c > 'z') && (c < '0' || c > '9') && c != '-' {
			return false
		}
	}

	return slug[0] != '-' && slug[len(slug)-1] != '-'
}

func isNumericPort(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}

	return len(s) > 0
}
