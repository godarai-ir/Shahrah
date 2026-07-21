package platformcredentials

import (
	"context"
	"errors"
	"time"
)

var (
	ErrUnavailable      = errors.New("platform_credential_unavailable")
	ErrExpired          = errors.New("platform_credential_expired")
	ErrDecryptFailed    = errors.New("credential_decrypt_failed")
	ErrInvalidKEK       = errors.New("invalid_platform_credential_kek")
	ErrSnapshotTooLarge = errors.New("platform_credential_snapshot_too_large")
)

type Binding struct {
	ProviderSlug        string
	Environment         string
	ConnectionID        string
	CredentialVersionID string
	CredentialVersion   int64
	AuthMode            string
	SchemaVersion       int
}

type EncryptedCredential struct {
	Binding
	BaseURL      string
	APIVersion   string
	Organization string
	Project      string
	HTTPReferer  string
	AppTitle     string
	ClientName   string
	Ciphertext   []byte
	DataNonce    []byte
	WrappedDEK   []byte
	WrapNonce    []byte
	KEKVersion   string
	Fingerprint  string
	DisplayHint  string
	ExpiresAt    *time.Time
}

// ResolvedCredential is immutable after insertion into a Manager snapshot. It
// is an internal runtime value and must never be JSON-marshaled or logged.
type ResolvedCredential struct {
	ProviderSlug        string
	Environment         string
	ConnectionID        string
	CredentialVersionID string
	CredentialVersion   int64
	AuthMode            string
	APIKey              string
	BaseURL             string
	APIVersion          string
	Organization        string
	Project             string
	HTTPReferer         string
	AppTitle            string
	ClientName          string
	ExpiresAt           *time.Time
}

type ActiveCredentialLoader interface {
	LoadActivePlatformCredentials(ctx context.Context) ([]EncryptedCredential, error)
}

type CredentialDecrypter interface {
	Decrypt(record EncryptedCredential) (map[string]string, error)
}

type InvalidationEvent struct {
	ProviderSlug        string `json:"provider_slug"`
	Environment         string `json:"environment"`
	ConnectionID        string `json:"connection_id"`
	CredentialVersionID string `json:"credential_version_id,omitempty"`
	Action              string `json:"action"`
}

type InvalidationPublisher interface {
	PublishPlatformCredentialInvalidation(ctx context.Context, event InvalidationEvent) error
}

type Metrics interface {
	ObserveResolve(result string)
	ObserveSnapshotLoad(success bool, duration time.Duration, entries int)
}

type nopMetrics struct{}

func (nopMetrics) ObserveResolve(string)                        {}
func (nopMetrics) ObserveSnapshotLoad(bool, time.Duration, int) {}
