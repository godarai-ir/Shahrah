package tenantcredentials

import (
	"encoding/base64"
	"errors"

	"github.com/Delpak/shahrah/services/control-plane/internal/platformcredentials"
)

var ErrUnavailable = errors.New("tenant_credential_encryption_unavailable")

// NewEncryptedStore creates the tenant envelope-encryption store from injected secrets.
func NewEncryptedStore(active, encodedKeys, encodedFingerprint string) (*platformcredentials.EncryptedDBSecretStore, error) {
	if active == "" || encodedKeys == "" || encodedFingerprint == "" {
		return nil, ErrUnavailable
	}

	keks, err := platformcredentials.NewEnvironmentKEKProvider(active, encodedKeys)
	if err != nil {
		return nil, err
	}

	fingerprintKey, err := base64.StdEncoding.DecodeString(encodedFingerprint)
	if err != nil || len(fingerprintKey) < 32 {
		return nil, platformcredentials.ErrInvalidKEK
	}

	return platformcredentials.NewEncryptedDBSecretStore(keks, fingerprintKey)
}
