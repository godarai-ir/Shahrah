package platformcredentials

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

// ProbeMetadata contains sanitized provider-specific credential state.
type ProbeMetadata struct {
	WIFExpiresAt   *time.Time
	WIFRefreshedAt *time.Time
	WIFScopes      []string
	Permissions    []string
	KeyDisabled    *bool
	KeyBlocked     *bool
	AccountBlocked *bool
}

// ProbeResult separates authentication from runtime readiness.
type ProbeResult struct {
	Authenticated bool
	Ready         bool
	Category      string
	Summary       string
	Metadata      ProbeMetadata
}

// CredentialProber validates a provider credential without exposing its secret.
type CredentialProber interface {
	Probe(context.Context, EncryptedCredential, map[string]string) ProbeResult
}

// CredentialProberRegistry dispatches credential probes by canonical provider.
type CredentialProberRegistry struct {
	mu       sync.RWMutex
	probers  map[string]CredentialProber
	fallback CredentialProber
}

// NewCredentialProberRegistry creates a provider-aware credential prober.
func NewCredentialProberRegistry(fallback CredentialProber) *CredentialProberRegistry {
	return &CredentialProberRegistry{
		probers:  make(map[string]CredentialProber),
		fallback: fallback,
	}
}

// Register assigns a dedicated prober to a provider.
func (r *CredentialProberRegistry) Register(provider string, prober CredentialProber) {
	if r == nil || prober == nil {
		return
	}

	provider = normalizeProvider(provider)
	if provider == "" {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.probers[provider] = prober
}

// Probe dispatches to the provider-specific prober or the configured fallback.
func (r *CredentialProberRegistry) Probe(
	ctx context.Context,
	record EncryptedCredential,
	secrets map[string]string,
) ProbeResult {
	if r == nil {
		return unsupportedProbeResult()
	}

	provider := normalizeProvider(record.ProviderSlug)

	r.mu.RLock()
	prober := r.probers[provider]
	fallback := r.fallback
	r.mu.RUnlock()

	if prober != nil {
		return prober.Probe(ctx, record, secrets)
	}

	if fallback != nil {
		return fallback.Probe(ctx, record, secrets)
	}

	return unsupportedProbeResult()
}

func unsupportedProbeResult() ProbeResult {
	return ProbeResult{
		Category: "unsupported",
		Summary:  "This provider authentication mode has no implemented non-billable validation probe.",
	}
}

// PlatformDetails converts the result into the public sanitized contract.
func (r ProbeResult) PlatformDetails() *platform.ProviderCredentialProbeDetails {
	details := &platform.ProviderCredentialProbeDetails{
		Authenticated:      r.Authenticated,
		Ready:              r.Ready,
		Permissions:        append([]string(nil), r.Metadata.Permissions...),
		CredentialDisabled: cloneBoolPointer(r.Metadata.KeyDisabled),
		CredentialBlocked:  cloneBoolPointer(r.Metadata.KeyBlocked),
		AccountBlocked:     cloneBoolPointer(r.Metadata.AccountBlocked),
	}

	if !details.Authenticated && !details.Ready && len(details.Permissions) == 0 &&
		details.CredentialDisabled == nil && details.CredentialBlocked == nil && details.AccountBlocked == nil {
		return nil
	}

	return details
}

func cloneBoolPointer(value *bool) *bool {
	if value == nil {
		return nil
	}

	cloned := *value

	return &cloned
}

func boolPointer(value bool) *bool {
	return &value
}

func normalizedPermission(permission string) string {
	return strings.ToLower(strings.TrimSpace(permission))
}
