package oidc

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func pemKeys(t *testing.T) (string, string) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	private, err := x509.MarshalPKCS8PrivateKey(key)
	require.NoError(t, err)

	public, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	require.NoError(t, err)

	return string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: private})),
		string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: public}))
}

func TestKeySetSignsAndVerifiesWithStableKeyID(t *testing.T) {
	t.Parallel()

	privatePEM, _ := pemKeys(t)

	keys, err := NewKeySet(privatePEM, "")
	require.NoError(t, err)

	token, err := keys.Sign(accessTokenType, map[string]any{"sub": "user"})
	require.NoError(t, err)
	require.True(t, LooksLikeJWT(token))

	var claims map[string]any
	require.NoError(t, keys.Verify(token, accessTokenType, &claims))
	require.Equal(t, "user", claims["sub"])

	again, err := NewKeySet(privatePEM, "")
	require.NoError(t, err)
	require.Equal(t, keys.JWKS().Keys[0].KeyID, again.JWKS().Keys[0].KeyID)
}

func TestKeySetRejectsTamperingAndForeignKeys(t *testing.T) {
	t.Parallel()

	keys, err := GenerateKeySet()
	require.NoError(t, err)

	token, err := keys.Sign(accessTokenType, map[string]any{"sub": "user"})
	require.NoError(t, err)

	parts := strings.Split(token, ".")
	forged := parts[0] + "." + encodeSegment([]byte(`{"sub":"admin"}`)) + "." + parts[2]

	var claims map[string]any
	require.Error(t, keys.Verify(forged, accessTokenType, &claims))
	require.Error(t, keys.Verify(token, idTokenType, &claims), "type header is enforced")

	unsigned := encodeSegment([]byte(`{"alg":"none","typ":"at+jwt"}`)) + "." + parts[1] + "."
	require.Error(t, keys.Verify(unsigned, accessTokenType, &claims))

	other, err := GenerateKeySet()
	require.NoError(t, err)
	require.Error(t, other.Verify(token, accessTokenType, &claims))
}

func TestRetiredKeysStillVerifyDuringRotation(t *testing.T) {
	t.Parallel()

	oldPrivate, oldPublic := pemKeys(t)
	newPrivate, _ := pemKeys(t)

	oldKeys, err := NewKeySet(oldPrivate, "")
	require.NoError(t, err)

	token, err := oldKeys.Sign(accessTokenType, map[string]any{"sub": "user"})
	require.NoError(t, err)

	rotated, err := NewKeySet(newPrivate, oldPublic)
	require.NoError(t, err)
	require.Len(t, rotated.JWKS().Keys, 2)

	var claims map[string]any
	require.NoError(t, rotated.Verify(token, accessTokenType, &claims))
}

func TestNewKeySetRejectsUnsupportedKeys(t *testing.T) {
	t.Parallel()

	_, err := NewKeySet("not a pem", "")
	require.Error(t, err)

	key, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	require.NoError(t, err)

	raw, err := x509.MarshalECPrivateKey(key)
	require.NoError(t, err)

	_, err = NewKeySet(string(pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: raw})), "")
	require.Error(t, err)
}
