package config

import (
	"os"
	"strings"
)

// CredentialSettings is the immutable bootstrap snapshot for encrypted credentials.
type CredentialSettings struct {
	PlatformMode           string
	PlatformActiveVersion  string
	PlatformKEKsJSON       string
	PlatformFingerprintKey string
	TenantActiveVersion    string
	TenantKEKsJSON         string
	TenantFingerprintKey   string
}

// LoadCredentialSettings reads credential encryption settings at the configuration edge.
func LoadCredentialSettings() CredentialSettings {
	return CredentialSettings{
		PlatformMode:           strings.TrimSpace(os.Getenv("PLATFORM_PROVIDER_CREDENTIAL_MODE")),
		PlatformActiveVersion:  strings.TrimSpace(os.Getenv("PLATFORM_PROVIDER_CREDENTIAL_ACTIVE_KEK_VERSION")),
		PlatformKEKsJSON:       strings.TrimSpace(os.Getenv("PLATFORM_PROVIDER_CREDENTIAL_KEKS_JSON")),
		PlatformFingerprintKey: strings.TrimSpace(os.Getenv("PLATFORM_PROVIDER_CREDENTIAL_FINGERPRINT_KEY")),
		TenantActiveVersion:    strings.TrimSpace(os.Getenv("TENANT_PROVIDER_CREDENTIAL_ACTIVE_KEK_VERSION")),
		TenantKEKsJSON:         strings.TrimSpace(os.Getenv("TENANT_PROVIDER_CREDENTIAL_KEKS_JSON")),
		TenantFingerprintKey:   strings.TrimSpace(os.Getenv("TENANT_PROVIDER_CREDENTIAL_FINGERPRINT_KEY")),
	}
}
