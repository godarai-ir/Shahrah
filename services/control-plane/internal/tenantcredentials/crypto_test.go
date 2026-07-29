package tenantcredentials

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/Delpak/shahrah/services/control-plane/internal/platformcredentials"
	"github.com/stretchr/testify/require"
)

func TestTenantCredentialAADIncludesTenantAndAccount(t *testing.T) {
	keys, err := json.Marshal(map[string]string{
		"v1": base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{1}, 32)),
	})
	require.NoError(t, err)
	keks, err := platformcredentials.NewEnvironmentKEKProvider("v1", string(keys))
	require.NoError(t, err)
	store, err := platformcredentials.NewEncryptedDBSecretStore(
		keks,
		bytes.Repeat([]byte{2}, 32),
	)
	require.NoError(t, err)

	binding := platformcredentials.Binding{
		TenantID:            "tenant-1",
		AccountID:           "account-1",
		ProviderSlug:        "anthropic",
		AuthMode:            "api_key",
		SchemaVersion:       1,
		CredentialVersionID: "credential-1",
		CredentialVersion:   1,
	}
	encrypted, err := store.Encrypt(binding, map[string]string{
		"api_key": "secret",
	})
	require.NoError(t, err)

	encrypted.AccountID = "account-2"
	_, err = store.Decrypt(encrypted)
	require.ErrorIs(t, err, platformcredentials.ErrDecryptFailed)
}
