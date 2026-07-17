package platformcredentials

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

const fingerprintDomain = "godarai/platform-provider-credential/fingerprint/v1\x00"

type KeyEncryptionKeyProvider interface {
	ActiveKey() (version string, key []byte, err error)
	Key(version string) ([]byte, error)
}

type EnvironmentKEKProvider struct {
	active string
	keys   map[string][]byte
}

func NewEnvironmentKEKProviderFromEnv() (*EnvironmentKEKProvider, error) {
	return NewEnvironmentKEKProvider(
		os.Getenv("PLATFORM_PROVIDER_CREDENTIAL_ACTIVE_KEK_VERSION"),
		os.Getenv("PLATFORM_PROVIDER_CREDENTIAL_KEKS_JSON"),
	)
}

func NewEnvironmentKEKProvider(active, encodedJSON string) (*EnvironmentKEKProvider, error) {
	active = strings.TrimSpace(active)
	var encoded map[string]string
	if active == "" || json.Unmarshal([]byte(encodedJSON), &encoded) != nil || len(encoded) == 0 {
		return nil, ErrInvalidKEK
	}
	keys := make(map[string][]byte, len(encoded))
	for version, value := range encoded {
		key, err := base64.StdEncoding.DecodeString(value)
		if err != nil || len(key) != 32 || strings.TrimSpace(version) == "" {
			return nil, ErrInvalidKEK
		}
		keys[version] = append([]byte(nil), key...)
	}
	if _, ok := keys[active]; !ok {
		return nil, ErrInvalidKEK
	}
	return &EnvironmentKEKProvider{active: active, keys: keys}, nil
}

func (p *EnvironmentKEKProvider) ActiveKey() (string, []byte, error) {
	key, err := p.Key(p.active)
	return p.active, key, err
}

func (p *EnvironmentKEKProvider) Key(version string) ([]byte, error) {
	key, ok := p.keys[version]
	if !ok {
		return nil, ErrInvalidKEK
	}
	return append([]byte(nil), key...), nil
}

type EncryptedDBSecretStore struct {
	keks           KeyEncryptionKeyProvider
	fingerprintKey []byte
	random         io.Reader
}

func NewEncryptedDBSecretStore(keks KeyEncryptionKeyProvider, fingerprintKey []byte) (*EncryptedDBSecretStore, error) {
	if keks == nil || len(fingerprintKey) < 32 {
		return nil, ErrInvalidKEK
	}
	return &EncryptedDBSecretStore{keks: keks, fingerprintKey: append([]byte(nil), fingerprintKey...), random: rand.Reader}, nil
}

func FingerprintKeyFromEnv() ([]byte, error) {
	key, err := base64.StdEncoding.DecodeString(os.Getenv("PLATFORM_PROVIDER_CREDENTIAL_FINGERPRINT_KEY"))
	if err != nil || len(key) < 32 {
		return nil, ErrInvalidKEK
	}
	return key, nil
}

func (s *EncryptedDBSecretStore) Encrypt(binding Binding, secrets map[string]string) (EncryptedCredential, error) {
	plaintext, err := marshalSecrets(secrets)
	if err != nil {
		return EncryptedCredential{}, ErrDecryptFailed
	}
	defer clear(plaintext)

	dek := make([]byte, 32)
	if _, err := io.ReadFull(s.random, dek); err != nil {
		return EncryptedCredential{}, ErrDecryptFailed
	}
	defer clear(dek)
	dataCipher, dataNonce, err := seal(dek, plaintext, additionalData(binding), s.random)
	if err != nil {
		return EncryptedCredential{}, ErrDecryptFailed
	}

	kekVersion, kek, err := s.keks.ActiveKey()
	if err != nil {
		return EncryptedCredential{}, ErrInvalidKEK
	}
	defer clear(kek)
	wrappedDEK, wrapNonce, err := seal(kek, dek, additionalData(binding), s.random)
	if err != nil {
		return EncryptedCredential{}, ErrDecryptFailed
	}

	mac := hmac.New(sha256.New, s.fingerprintKey)
	_, _ = mac.Write([]byte(fingerprintDomain))
	_, _ = mac.Write(plaintext)
	return EncryptedCredential{
		Binding:     binding,
		Ciphertext:  dataCipher,
		DataNonce:   dataNonce,
		WrappedDEK:  wrappedDEK,
		WrapNonce:   wrapNonce,
		KEKVersion:  kekVersion,
		Fingerprint: base64.RawURLEncoding.EncodeToString(mac.Sum(nil)),
		DisplayHint: displayHint(secrets),
	}, nil
}

func (s *EncryptedDBSecretStore) Decrypt(record EncryptedCredential) (map[string]string, error) {
	kek, err := s.keks.Key(record.KEKVersion)
	if err != nil {
		return nil, ErrDecryptFailed
	}
	defer clear(kek)
	dek, err := open(kek, record.WrappedDEK, record.WrapNonce, additionalData(record.Binding))
	if err != nil || len(dek) != 32 {
		return nil, ErrDecryptFailed
	}
	defer clear(dek)
	plaintext, err := open(dek, record.Ciphertext, record.DataNonce, additionalData(record.Binding))
	if err != nil {
		return nil, ErrDecryptFailed
	}
	defer clear(plaintext)
	var secrets map[string]string
	if json.Unmarshal(plaintext, &secrets) != nil {
		return nil, ErrDecryptFailed
	}
	return secrets, nil
}

func (s *EncryptedDBSecretStore) FingerprintsEqual(left, right string) bool {
	return subtle.ConstantTimeCompare([]byte(left), []byte(right)) == 1
}

func seal(key, plaintext, aad []byte, random io.Reader) ([]byte, []byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(random, nonce); err != nil {
		return nil, nil, err
	}
	return gcm.Seal(nil, nonce, plaintext, aad), nonce, nil
}

func open(key, ciphertext, nonce, aad []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil || len(nonce) != gcm.NonceSize() {
		return nil, ErrDecryptFailed
	}
	return gcm.Open(nil, nonce, ciphertext, aad)
}

func additionalData(binding Binding) []byte {
	return []byte(fmt.Sprintf("provider=%s\nenvironment=%s\nconnection=%s\ncredential_version_id=%s\ncredential_version=%d\nauth_mode=%s\nschema_version=%d",
		binding.ProviderSlug, binding.Environment, binding.ConnectionID, binding.CredentialVersionID,
		binding.CredentialVersion, binding.AuthMode, binding.SchemaVersion))
}

func marshalSecrets(secrets map[string]string) ([]byte, error) {
	if len(secrets) == 0 {
		return nil, errorsNew("empty credential")
	}
	// encoding/json sorts map keys; reject empty keys/values before creating the
	// canonical plaintext used for both encryption and keyed fingerprinting.
	for key, value := range secrets {
		if strings.TrimSpace(key) == "" || strings.TrimSpace(value) == "" {
			return nil, errorsNew("empty credential field")
		}
	}
	return json.Marshal(secrets)
}

func displayHint(secrets map[string]string) string {
	keys := make([]string, 0, len(secrets))
	for key := range secrets {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	value := secrets[keys[0]]
	runes := []rune(value)
	if len(runes) > 4 {
		return "configured ••••" + string(runes[len(runes)-4:])
	}
	return "configured"
}

// avoids leaking lower-level validation text through the package's exported
// stable errors while keeping marshalSecrets independently testable.
func errorsNew(message string) error { return fmt.Errorf("%s", message) }
