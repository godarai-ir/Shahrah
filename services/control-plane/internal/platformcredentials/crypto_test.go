package platformcredentials

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func testSecretStore(t *testing.T) *EncryptedDBSecretStore {
	t.Helper()
	encoded, err := json.Marshal(map[string]string{
		"v1": base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{1}, 32)),
		"v2": base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{2}, 32)),
	})
	require.NoError(t, err)
	keks, err := NewEnvironmentKEKProvider("v2", string(encoded))
	require.NoError(t, err)
	store, err := NewEncryptedDBSecretStore(keks, bytes.Repeat([]byte{3}, 32))
	require.NoError(t, err)
	return store
}

func testBinding() Binding {
	return Binding{ProviderSlug: "openai", Environment: "production", ConnectionID: "connection-1", CredentialVersionID: "credential-1", CredentialVersion: 1, AuthMode: "api_key", SchemaVersion: 1}
}

func TestEncryptedDBSecretStoreRoundTripAndUniqueness(t *testing.T) {
	store := testSecretStore(t)
	secrets := map[string]string{"api_key": "sk-super-secret", "base_url": "https://api.openai.com"}
	one, err := store.Encrypt(testBinding(), secrets)
	require.NoError(t, err)
	two, err := store.Encrypt(testBinding(), secrets)
	require.NoError(t, err)

	require.NotEqual(t, one.DataNonce, two.DataNonce)
	require.NotEqual(t, one.WrapNonce, two.WrapNonce)
	require.NotEqual(t, one.WrappedDEK, two.WrappedDEK)
	require.NotEqual(t, one.Ciphertext, two.Ciphertext)
	require.Equal(t, one.Fingerprint, two.Fingerprint)
	require.True(t, store.FingerprintsEqual(one.Fingerprint, two.Fingerprint))
	require.NotContains(t, string(one.Ciphertext), secrets["api_key"])
	require.NotContains(t, string(one.WrappedDEK), secrets["api_key"])

	resolved, err := store.Decrypt(one)
	require.NoError(t, err)
	require.Equal(t, secrets, resolved)
}

func TestEncryptedDBSecretStoreRejectsTamperWrongKEKAndMalformedRecords(t *testing.T) {
	store := testSecretStore(t)
	record, err := store.Encrypt(testBinding(), map[string]string{"api_key": "sk-super-secret"})
	require.NoError(t, err)

	t.Run("aad", func(t *testing.T) {
		tampered := record
		tampered.ProviderSlug = "anthropic"
		_, err := store.Decrypt(tampered)
		require.ErrorIs(t, err, ErrDecryptFailed)
	})
	t.Run("kek version", func(t *testing.T) {
		tampered := record
		tampered.KEKVersion = "missing"
		_, err := store.Decrypt(tampered)
		require.ErrorIs(t, err, ErrDecryptFailed)
	})
	t.Run("ciphertext", func(t *testing.T) {
		tampered := record
		tampered.Ciphertext = []byte("malformed")
		_, err := store.Decrypt(tampered)
		require.ErrorIs(t, err, ErrDecryptFailed)
	})
}

func TestEnvironmentKEKProviderValidationAndVersionedReads(t *testing.T) {
	encoded := `{"old":"` + base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{7}, 32)) + `"}`
	provider, err := NewEnvironmentKEKProvider("old", encoded)
	require.NoError(t, err)
	version, key, err := provider.ActiveKey()
	require.NoError(t, err)
	require.Equal(t, "old", version)
	require.Len(t, key, 32)
	key[0] = 0
	again, err := provider.Key("old")
	require.NoError(t, err)
	require.Equal(t, byte(7), again[0], "callers must not mutate the provider key")

	_, err = NewEnvironmentKEKProvider("missing", encoded)
	require.ErrorIs(t, err, ErrInvalidKEK)
	_, err = NewEnvironmentKEKProvider("old", `{"old":"c2hvcnQ="}`)
	require.ErrorIs(t, err, ErrInvalidKEK)
}

func TestEncryptedRecordHasNoPlaintextInJSON(t *testing.T) {
	store := testSecretStore(t)
	record, err := store.Encrypt(testBinding(), map[string]string{"api_key": "unique-plaintext-value"})
	require.NoError(t, err)
	encoded, err := json.Marshal(record)
	require.NoError(t, err)
	require.False(t, strings.Contains(string(encoded), "unique-plaintext-value"))
	require.False(t, errors.Is(err, ErrDecryptFailed))
}
