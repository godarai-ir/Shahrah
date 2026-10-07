package oidc

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"strings"
)

const (
	signingAlgorithm = "ES256"
	es256CoordSize   = 32
)

var (
	errMalformedToken    = errors.New("malformed token")
	errUnsupportedToken  = errors.New("unsupported token algorithm or type")
	errUnknownSigningKey = errors.New("unknown signing key")
	errBadSignature      = errors.New("invalid token signature")
)

// JSONWebKey is a public key in JWKS form.
type JSONWebKey struct {
	KeyType   string `json:"kty"`
	Curve     string `json:"crv"`
	X         string `json:"x"`
	Y         string `json:"y"`
	KeyID     string `json:"kid"`
	Use       string `json:"use"`
	Algorithm string `json:"alg"`
}

// JSONWebKeySet is the JWKS document.
type JSONWebKeySet struct {
	Keys []JSONWebKey `json:"keys"`
}

type verificationKey struct {
	key *ecdsa.PublicKey
	jwk JSONWebKey
}

// KeySet signs tokens with one ES256 key and verifies tokens signed by it or by retired keys.
type KeySet struct {
	signer   *ecdsa.PrivateKey
	signerID string
	verifier map[string]verificationKey
	jwks     JSONWebKeySet
}

// NewKeySet builds a key set from a PEM private signing key and optional PEM public keys
// that remain valid for verification during rotation.
func NewKeySet(signingKeyPEM string, retiredPublicKeysPEM string) (*KeySet, error) {
	signer, err := parseSigningKey(signingKeyPEM)
	if err != nil {
		return nil, err
	}

	retired, err := parsePublicKeys(retiredPublicKeysPEM)
	if err != nil {
		return nil, err
	}

	return newKeySet(signer, retired)
}

// GenerateKeySet creates an ephemeral key set; tokens do not survive a restart.
func GenerateKeySet() (*KeySet, error) {
	signer, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generate oidc signing key: %w", err)
	}

	return newKeySet(signer, nil)
}

func newKeySet(signer *ecdsa.PrivateKey, retired []*ecdsa.PublicKey) (*KeySet, error) {
	set := &KeySet{signer: signer, verifier: make(map[string]verificationKey, len(retired)+1)}

	for _, public := range append([]*ecdsa.PublicKey{&signer.PublicKey}, retired...) {
		jwk, err := publicJWK(public)
		if err != nil {
			return nil, err
		}

		if _, exists := set.verifier[jwk.KeyID]; exists {
			continue
		}

		set.verifier[jwk.KeyID] = verificationKey{key: public, jwk: jwk}
		set.jwks.Keys = append(set.jwks.Keys, jwk)
	}

	set.signerID = set.jwks.Keys[0].KeyID

	return set, nil
}

// JWKS returns the public verification keys.
func (k *KeySet) JWKS() JSONWebKeySet {
	keys := make([]JSONWebKey, len(k.jwks.Keys))
	copy(keys, k.jwks.Keys)

	return JSONWebKeySet{Keys: keys}
}

// Sign serializes claims as a compact ES256 JWS with the given type header.
func (k *KeySet) Sign(tokenType string, claims any) (string, error) {
	header, err := json.Marshal(map[string]string{"alg": signingAlgorithm, "kid": k.signerID, "typ": tokenType})
	if err != nil {
		return "", err
	}

	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	signingInput := encodeSegment(header) + "." + encodeSegment(payload)
	digest := sha256.Sum256([]byte(signingInput))

	r, s, err := ecdsa.Sign(rand.Reader, k.signer, digest[:])
	if err != nil {
		return "", fmt.Errorf("sign oidc token: %w", err)
	}

	signature := make([]byte, 2*es256CoordSize)
	r.FillBytes(signature[:es256CoordSize])
	s.FillBytes(signature[es256CoordSize:])

	return signingInput + "." + encodeSegment(signature), nil
}

// Verify checks the signature and type header of a compact JWS and decodes its claims.
// It does not validate time or audience claims.
func (k *KeySet) Verify(token, expectedType string, claims any) error {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return errMalformedToken
	}

	var header struct {
		Algorithm string `json:"alg"`
		KeyID     string `json:"kid"`
		Type      string `json:"typ"`
	}

	if err := decodeSegment(parts[0], &header); err != nil {
		return errMalformedToken
	}

	if header.Algorithm != signingAlgorithm || !strings.EqualFold(header.Type, expectedType) {
		return errUnsupportedToken
	}

	key, ok := k.verifier[header.KeyID]
	if !ok {
		return errUnknownSigningKey
	}

	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || len(signature) != 2*es256CoordSize {
		return errBadSignature
	}

	digest := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	r := new(big.Int).SetBytes(signature[:es256CoordSize])

	s := new(big.Int).SetBytes(signature[es256CoordSize:])
	if !ecdsa.Verify(key.key, digest[:], r, s) {
		return errBadSignature
	}

	if err := decodeSegment(parts[1], claims); err != nil {
		return errMalformedToken
	}

	return nil
}

// LooksLikeJWT reports whether a bearer credential has compact JWS shape.
func LooksLikeJWT(token string) bool {
	return strings.HasPrefix(token, "eyJ") && strings.Count(token, ".") == 2
}

func parseSigningKey(raw string) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(strings.TrimSpace(raw)))
	if block == nil {
		return nil, errors.New("oidc signing key must be PEM encoded")
	}

	if key, err := x509.ParseECPrivateKey(block.Bytes); err == nil {
		return requireP256(key)
	}

	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, errors.New("oidc signing key must be an EC P-256 private key")
	}

	key, ok := parsed.(*ecdsa.PrivateKey)
	if !ok {
		return nil, errors.New("oidc signing key must be an EC P-256 private key")
	}

	return requireP256(key)
}

func requireP256(key *ecdsa.PrivateKey) (*ecdsa.PrivateKey, error) {
	if key.Curve != elliptic.P256() {
		return nil, errors.New("oidc signing key must use curve P-256")
	}

	return key, nil
}

func parsePublicKeys(raw string) ([]*ecdsa.PublicKey, error) {
	rest := []byte(strings.TrimSpace(raw))

	var keys []*ecdsa.PublicKey

	for len(rest) > 0 {
		var block *pem.Block

		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}

		parsed, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, errors.New("oidc retired verification key must be a PEM public key")
		}

		key, ok := parsed.(*ecdsa.PublicKey)
		if !ok || key.Curve != elliptic.P256() {
			return nil, errors.New("oidc retired verification key must be an EC P-256 public key")
		}

		keys = append(keys, key)
	}

	return keys, nil
}

// publicJWK renders a P-256 key and derives its RFC 7638 thumbprint as the key ID.
func publicJWK(key *ecdsa.PublicKey) (JSONWebKey, error) {
	exchange, err := key.ECDH()
	if err != nil {
		return JSONWebKey{}, fmt.Errorf("encode oidc public key: %w", err)
	}

	point := exchange.Bytes()
	if len(point) != 1+2*es256CoordSize {
		return JSONWebKey{}, errors.New("encode oidc public key: unexpected point size")
	}

	x := encodeSegment(point[1 : 1+es256CoordSize])
	y := encodeSegment(point[1+es256CoordSize:])
	thumbprint := sha256.Sum256([]byte(`{"crv":"P-256","kty":"EC","x":"` + x + `","y":"` + y + `"}`))

	return JSONWebKey{
		KeyType:   "EC",
		Curve:     "P-256",
		X:         x,
		Y:         y,
		KeyID:     encodeSegment(thumbprint[:]),
		Use:       "sig",
		Algorithm: signingAlgorithm,
	}, nil
}

func encodeSegment(raw []byte) string {
	return base64.RawURLEncoding.EncodeToString(raw)
}

func decodeSegment(segment string, target any) error {
	raw, err := base64.RawURLEncoding.DecodeString(segment)
	if err != nil {
		return err
	}

	return json.Unmarshal(raw, target)
}
