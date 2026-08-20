// Package sessiontoken issues and verifies short-lived private chatbot JWTs.
package sessiontoken

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var rawURL = base64.RawURLEncoding

// Claims are the minimal private bot admission claims.
type Claims struct {
	ChatbotPublicID  string `json:"bot"`
	ChatbotVersionID string `json:"ver"`
	TenantID         string `json:"tid"`
	ExpiresAt        int64  `json:"exp"`
	IssuedAt         int64  `json:"iat"`
	Nonce            string `json:"jti"`
}

// Sign creates an HS256 token with a hard one-hour maximum TTL.
func Sign(secret []byte, claims Claims, now time.Time) (string, time.Time, error) {
	if len(secret) < 32 {
		return "", time.Time{}, errors.New("session token secret must be at least 32 bytes")
	}

	expiresAt := time.Unix(claims.ExpiresAt, 0).UTC()
	if !expiresAt.After(now) || expiresAt.After(now.Add(time.Hour)) {
		return "", time.Time{}, errors.New("session token TTL must be between 1 second and 1 hour")
	}

	claims.IssuedAt = now.Unix()

	header, err := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	if err != nil {
		return "", time.Time{}, fmt.Errorf("encode token header: %w", err)
	}

	payload, err := json.Marshal(claims)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("encode token claims: %w", err)
	}

	unsigned := rawURL.EncodeToString(header) + "." + rawURL.EncodeToString(payload)
	signature := sign(secret, unsigned)

	return unsigned + "." + rawURL.EncodeToString(signature), expiresAt, nil
}

// Verify validates signature, expiry, and immutable chatbot publication audience.
func Verify(
	secret []byte,
	token, chatbotPublicID, chatbotVersionID, tenantID string,
	now time.Time,
) (*Claims, error) {
	parts := strings.Split(strings.TrimSpace(token), ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid session token")
	}

	unsigned := parts[0] + "." + parts[1]

	signature, err := rawURL.DecodeString(parts[2])
	if err != nil || !hmac.Equal(signature, sign(secret, unsigned)) {
		return nil, errors.New("invalid session token")
	}

	payload, err := rawURL.DecodeString(parts[1])
	if err != nil {
		return nil, errors.New("invalid session token")
	}

	var claims Claims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, errors.New("invalid session token")
	}

	if claims.ChatbotPublicID != chatbotPublicID ||
		claims.ChatbotVersionID != chatbotVersionID ||
		claims.TenantID != tenantID ||
		claims.ExpiresAt <= now.Unix() {
		return nil, errors.New("expired or mismatched session token")
	}

	return &claims, nil
}

func sign(secret []byte, value string) []byte {
	hash := hmac.New(sha256.New, secret)
	_, _ = hash.Write([]byte(value))

	return hash.Sum(nil)
}
