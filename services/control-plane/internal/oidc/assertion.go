package oidc

import (
	"context"
	"crypto"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

// maxUserAssertionLifetime bounds exp-iat, matching the longest TTL LibreChat signs.
const maxUserAssertionLifetime = time.Hour

// UserAssertionKeyType identifies gateway auth contexts created from LibreChat user assertions.
const UserAssertionKeyType = "user_assertion"

// UserAssertionSettings configures verification of the short-lived JWT LibreChat signs to name
// the signed-in member on each gateway request.
type UserAssertionSettings struct {
	// PublicKeyPEM is LibreChat's Ed25519 or RSA public key.
	PublicKeyPEM string
	// KeyID is the kid header LibreChat signs with; it routes a bearer JWT to this verifier.
	KeyID    string
	Issuer   string
	Audience string
}

// UserAssertionVerifier checks LibreChat user assertions.
type UserAssertionVerifier struct {
	key      crypto.PublicKey
	alg      string
	keyID    string
	issuer   string
	audience string
}

type userAssertionClaims struct {
	Issuer    string `json:"iss"`
	Audience  string `json:"aud"`
	Subject   string `json:"sub"`
	TenantID  string `json:"tid"`
	IssuedAt  int64  `json:"iat"`
	NotBefore int64  `json:"nbf"`
	ExpiresAt int64  `json:"exp"`
}

// NewUserAssertionVerifier parses the public key. Every setting is required.
func NewUserAssertionVerifier(settings UserAssertionSettings) (*UserAssertionVerifier, error) {
	keyID := strings.TrimSpace(settings.KeyID)
	issuer := strings.TrimSpace(settings.Issuer)
	audience := strings.TrimSpace(settings.Audience)

	if keyID == "" || issuer == "" || audience == "" {
		return nil, errors.New("user assertion key id, issuer, and audience are required")
	}

	block, _ := pem.Decode([]byte(strings.TrimSpace(settings.PublicKeyPEM)))
	if block == nil {
		return nil, errors.New("user assertion public key must be PEM encoded")
	}

	parsed, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, errors.New("user assertion public key must be a PEM public key")
	}

	verifier := &UserAssertionVerifier{key: parsed, keyID: keyID, issuer: issuer, audience: audience}

	switch parsed.(type) {
	case ed25519.PublicKey:
		verifier.alg = "EdDSA"
	case *rsa.PublicKey:
		verifier.alg = "RS256"
	default:
		return nil, errors.New("user assertion public key must be Ed25519 or RSA")
	}

	return verifier, nil
}

// matches reports whether a bearer JWT names this verifier's key. The header is not trusted
// beyond routing: verify checks the signature with the configured key and algorithm.
func (v *UserAssertionVerifier) matches(token string) bool {
	var header struct {
		KeyID string `json:"kid"`
	}

	head, _, found := strings.Cut(token, ".")

	return found && decodeSegment(head, &header) == nil && header.KeyID == v.keyID
}

func (v *UserAssertionVerifier) verify(token string, now time.Time) (*userAssertionClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errMalformedToken
	}

	var header struct {
		Algorithm string `json:"alg"`
	}

	if err := decodeSegment(parts[0], &header); err != nil {
		return nil, errMalformedToken
	}

	if header.Algorithm != v.alg {
		return nil, errUnsupportedToken
	}

	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, errBadSignature
	}

	if !v.validSignature([]byte(parts[0]+"."+parts[1]), signature) {
		return nil, errBadSignature
	}

	var claims userAssertionClaims
	if err := decodeSegment(parts[1], &claims); err != nil {
		return nil, errMalformedToken
	}

	if err := v.validClaims(claims, now); err != nil {
		return nil, err
	}

	return &claims, nil
}

func (v *UserAssertionVerifier) validSignature(signingInput, signature []byte) bool {
	switch key := v.key.(type) {
	case ed25519.PublicKey:
		return ed25519.Verify(key, signingInput, signature)
	case *rsa.PublicKey:
		digest := sha256.Sum256(signingInput)

		return rsa.VerifyPKCS1v15(key, crypto.SHA256, digest[:], signature) == nil
	default:
		return false
	}
}

func (v *UserAssertionVerifier) validClaims(claims userAssertionClaims, now time.Time) error {
	if claims.Issuer != v.issuer || claims.Audience != v.audience ||
		strings.TrimSpace(claims.Subject) == "" || strings.TrimSpace(claims.TenantID) == "" {
		return errors.New("user assertion claims are invalid")
	}

	expiresAt := time.Unix(claims.ExpiresAt, 0)
	issuedAt := time.Unix(claims.IssuedAt, 0)

	if claims.ExpiresAt == 0 || claims.IssuedAt == 0 || expiresAt.Sub(issuedAt) > maxUserAssertionLifetime {
		return errors.New("user assertion lifetime is invalid")
	}

	if now.After(expiresAt.Add(allowedClockSkew)) {
		return errors.New("user assertion has expired")
	}

	notBefore := issuedAt
	if claims.NotBefore != 0 {
		notBefore = time.Unix(claims.NotBefore, 0)
	}

	if now.Add(allowedClockSkew).Before(notBefore) {
		return errors.New("user assertion is not yet valid")
	}

	return nil
}

// authenticateUserAssertion resolves the member a LibreChat assertion names. Its usage is
// LibreChat usage, so a seated member's chat spends their seat credits.
func (s *Service) authenticateUserAssertion(ctx context.Context, tenantID, token string) (*platform.AuthContext, error) {
	claims, err := s.assertions.verify(token, s.now())
	if err != nil {
		return nil, protocolError(CodeInvalidToken, "user assertion is invalid")
	}

	if !strings.EqualFold(claims.TenantID, strings.TrimSpace(tenantID)) {
		return nil, protocolError(CodeInvalidToken, "user assertion was issued for another tenant")
	}

	resolved, err := s.principals.ResolveOIDCUserAuth(ctx, claims.TenantID, claims.Subject)
	if err != nil {
		if errors.Is(err, ErrIdentityNotFound) {
			return nil, protocolError(CodeInvalidToken, "the account is no longer allowed to use this assertion")
		}

		return nil, err
	}

	if resolved == nil || !strings.EqualFold(resolved.UserID, claims.Subject) || !strings.EqualFold(resolved.TenantID, claims.TenantID) {
		return nil, protocolError(CodeInvalidToken, "the account is no longer allowed to use this assertion")
	}

	// A stable per-member key id keeps response-cache scopes and audit rows per member.
	resolved.KeyID = UserAssertionKeyType + ":" + resolved.UserID
	resolved.KeyAlias = "librechat"
	resolved.KeyType = UserAssertionKeyType
	resolved.CreditProduct = platform.CreditProductLibreChat

	return resolved, nil
}
