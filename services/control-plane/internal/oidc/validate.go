package oidc

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net"
	"net/url"
	"slices"
	"strings"
)

const (
	maxStateLength       = 1024
	maxNonceLength       = 512
	maxRedirectURIs      = 10
	maxRedirectURILength = 2048
	maxClientNameLength  = 120
	minVerifierLength    = 43
	maxVerifierLength    = 128
	tokenEntropyBytes    = 32
	clientIDEntropyBytes = 16
)

var knownScopes = []string{ScopeOpenID, ScopeProfile, ScopeEmail, ScopePhone, ScopeOfflineAccess}

func supportedScopes() []string {
	return slices.Clone(knownScopes)
}

// normalizeScope keeps known scopes in request order, drops unknown ones, and requires openid.
func normalizeScope(raw string) (string, error) {
	seen := make(map[string]struct{}, len(knownScopes))
	scopes := make([]string, 0, len(knownScopes))

	for _, scope := range strings.Fields(raw) {
		if !slices.Contains(knownScopes, scope) {
			continue
		}

		if _, ok := seen[scope]; ok {
			continue
		}

		seen[scope] = struct{}{}
		scopes = append(scopes, scope)
	}

	if _, ok := seen[ScopeOpenID]; !ok {
		return "", protocolError(CodeInvalidScope, "the openid scope is required")
	}

	return strings.Join(scopes, " "), nil
}

// narrowScope validates a refresh-time scope request against the originally granted scope.
func narrowScope(requested, granted string) (string, error) {
	if strings.TrimSpace(requested) == "" {
		return granted, nil
	}

	grantedScopes := strings.Fields(granted)
	for _, scope := range strings.Fields(requested) {
		if !slices.Contains(grantedScopes, scope) {
			return "", protocolError(CodeInvalidScope, "requested scope exceeds the original grant")
		}
	}

	return normalizeScope(requested)
}

func hasScope(scope, want string) bool {
	return slices.Contains(strings.Fields(scope), want)
}

// verifyPKCE checks an RFC 7636 S256 code verifier against its challenge.
func verifyPKCE(verifier, challenge string) bool {
	if len(verifier) < minVerifierLength || len(verifier) > maxVerifierLength {
		return false
	}

	for _, char := range verifier {
		if !isUnreserved(char) {
			return false
		}
	}

	digest := sha256.Sum256([]byte(verifier))
	expected := base64.RawURLEncoding.EncodeToString(digest[:])

	return subtle.ConstantTimeCompare([]byte(expected), []byte(challenge)) == 1
}

func validCodeChallenge(challenge string) bool {
	if len(challenge) != base64.RawURLEncoding.EncodedLen(sha256.Size) {
		return false
	}

	_, err := base64.RawURLEncoding.DecodeString(challenge)

	return err == nil
}

func isUnreserved(char rune) bool {
	return (char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') ||
		char == '-' || char == '.' || char == '_' || char == '~'
}

// normalizeRedirectURIs validates registered redirect URIs: absolute, no fragment, and HTTPS
// unless the host is loopback or a .localhost development name.
func normalizeRedirectURIs(raw []string, required bool) ([]string, error) {
	uris := make([]string, 0, len(raw))

	for _, value := range raw {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}

		if err := validateRedirectURI(value); err != nil {
			return nil, err
		}

		if !slices.Contains(uris, value) {
			uris = append(uris, value)
		}
	}

	if required && len(uris) == 0 {
		return nil, fmt.Errorf("%w: at least one redirect URI is required", ErrInvalidClientConfig)
	}

	if len(uris) > maxRedirectURIs {
		return nil, fmt.Errorf("%w: at most %d redirect URIs are allowed", ErrInvalidClientConfig, maxRedirectURIs)
	}

	return uris, nil
}

func validateRedirectURI(value string) error {
	if len(value) > maxRedirectURILength {
		return fmt.Errorf("%w: redirect URI is too long", ErrInvalidClientConfig)
	}

	parsed, err := url.Parse(value)
	if err != nil || !parsed.IsAbs() || parsed.Host == "" || parsed.User != nil {
		return fmt.Errorf("%w: redirect URI must be an absolute URL without credentials", ErrInvalidClientConfig)
	}

	if parsed.Fragment != "" || strings.Contains(value, "#") {
		return fmt.Errorf("%w: redirect URI must not contain a fragment", ErrInvalidClientConfig)
	}

	switch parsed.Scheme {
	case "https":
		return nil
	case "http":
		if isDevelopmentHost(parsed.Hostname()) {
			return nil
		}
	}

	return fmt.Errorf("%w: redirect URI must use https", ErrInvalidClientConfig)
}

func isDevelopmentHost(host string) bool {
	host = strings.ToLower(host)
	if host == "localhost" || strings.HasSuffix(host, ".localhost") {
		return true
	}

	ip := net.ParseIP(host)

	return ip != nil && ip.IsLoopback()
}

func normalizeClientName(raw string) (string, error) {
	name := strings.TrimSpace(raw)
	if name == "" || len([]rune(name)) > maxClientNameLength {
		return "", fmt.Errorf("%w: name must be between 1 and %d characters", ErrInvalidClientConfig, maxClientNameLength)
	}

	return name, nil
}

// randomToken returns a URL-safe opaque token with 256 bits of entropy.
func randomToken(prefix string) (string, error) {
	raw := make([]byte, tokenEntropyBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generate oidc token: %w", err)
	}

	return prefix + base64.RawURLEncoding.EncodeToString(raw), nil
}

// randomClientID returns a public client identifier with 128 bits of entropy.
func randomClientID() (string, error) {
	raw := make([]byte, clientIDEntropyBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generate oidc client id: %w", err)
	}

	return "gdc-" + hex.EncodeToString(raw), nil
}

// hashToken stores high-entropy secrets and codes as SHA-256 digests.
func hashToken(token string) string {
	digest := sha256.Sum256([]byte(token))

	return hex.EncodeToString(digest[:])
}

func secretMatches(secret, storedHash string) bool {
	if secret == "" || storedHash == "" {
		return false
	}

	return subtle.ConstantTimeCompare([]byte(hashToken(secret)), []byte(storedHash)) == 1
}

func appendQuery(rawURL string, values url.Values) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	query := parsed.Query()

	for key, items := range values {
		for _, item := range items {
			if item != "" {
				query.Add(key, item)
			}
		}
	}

	parsed.RawQuery = query.Encode()

	return parsed.String(), nil
}
