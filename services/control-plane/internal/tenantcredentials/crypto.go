package tenantcredentials

import (
	"encoding/base64"
	"errors"
	"os"
	"strings"

	"github.com/Delpak/shahrah/services/control-plane/internal/platformcredentials"
)

var ErrUnavailable = errors.New("tenant_credential_encryption_unavailable")

// NewEncryptedStoreFromEnv creates the tenant envelope-encryption store. Tenant
// keys are intentionally separate from platform credential keys.
func NewEncryptedStoreFromEnv() (*platformcredentials.EncryptedDBSecretStore, error) {
	active := strings.TrimSpace(os.Getenv("TENANT_PROVIDER_CREDENTIAL_ACTIVE_KEK_VERSION"))
	encodedKeys := strings.TrimSpace(os.Getenv("TENANT_PROVIDER_CREDENTIAL_KEKS_JSON"))
	encodedFingerprint := strings.TrimSpace(os.Getenv("TENANT_PROVIDER_CREDENTIAL_FINGERPRINT_KEY"))
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
