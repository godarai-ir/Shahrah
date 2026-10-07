package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// OIDCSettings configures the tenant-scoped OpenID Connect provider.
type OIDCSettings struct {
	// Enabled turns on the provider routes and gateway access-token authentication.
	Enabled bool
	// SigningKeyPEM is the EC P-256 private key that signs tokens.
	SigningKeyPEM string
	// RetiredPublicKeysPEM holds earlier public keys still accepted during a rotation.
	RetiredPublicKeysPEM string
	// EphemeralSigningKey generates a throwaway key; only honored outside production.
	EphemeralSigningKey bool
	// LoadError reports a signing key file that could not be read.
	LoadError               error
	IssuerScheme            string
	LoginURLTemplate        string
	AccessTokenTTL          time.Duration
	IDTokenTTL              time.Duration
	RefreshTokenTTL         time.Duration
	GrantTTL                time.Duration
	AuthorizationRequestTTL time.Duration
	CodeTTL                 time.Duration
}

// LoadOIDCSettings reads OIDC provider settings once at the configuration edge.
func LoadOIDCSettings() OIDCSettings {
	settings := OIDCSettings{
		Enabled:                 booleanEnv("OIDC_ENABLED", false),
		SigningKeyPEM:           strings.TrimSpace(os.Getenv("OIDC_SIGNING_KEY_PEM")),
		RetiredPublicKeysPEM:    strings.TrimSpace(os.Getenv("OIDC_RETIRED_PUBLIC_KEYS_PEM")),
		IssuerScheme:            strings.ToLower(stringEnv("OIDC_ISSUER_SCHEME", "https")),
		LoginURLTemplate:        stringEnv("OIDC_LOGIN_URL_TEMPLATE", "https://{host}/login"),
		AccessTokenTTL:          positiveDurationEnv("OIDC_ACCESS_TOKEN_TTL", 10*time.Minute),
		IDTokenTTL:              positiveDurationEnv("OIDC_ID_TOKEN_TTL", 10*time.Minute),
		RefreshTokenTTL:         positiveDurationEnv("OIDC_REFRESH_TOKEN_TTL", 14*24*time.Hour),
		GrantTTL:                positiveDurationEnv("OIDC_GRANT_TTL", 30*24*time.Hour),
		AuthorizationRequestTTL: positiveDurationEnv("OIDC_AUTHORIZATION_REQUEST_TTL", 10*time.Minute),
		CodeTTL:                 positiveDurationEnv("OIDC_CODE_TTL", time.Minute),
	}

	if settings.SigningKeyPEM == "" {
		if path := strings.TrimSpace(os.Getenv("OIDC_SIGNING_KEY_FILE")); path != "" {
			raw, err := os.ReadFile(path) //nolint:gosec // The path is operator-supplied deployment configuration.
			if err != nil {
				settings.LoadError = fmt.Errorf("read OIDC_SIGNING_KEY_FILE: %w", err)
			}

			settings.SigningKeyPEM = strings.TrimSpace(string(raw))
		}
	}

	nodeEnvironment := strings.ToLower(strings.TrimSpace(os.Getenv("NODE_ENV")))
	if nodeEnvironment == "development" || nodeEnvironment == "test" {
		settings.EphemeralSigningKey = booleanEnv("OIDC_EPHEMERAL_SIGNING_KEY", false)
	}

	return settings
}
