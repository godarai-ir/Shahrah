package platformcredentials

import (
	"context"
	"strings"
	"sync/atomic"
	"time"

	"golang.org/x/sync/singleflight"
)

type snapshot struct {
	entries  map[credentialCacheKey]ResolvedCredential
	loadedAt time.Time
}

type credentialCacheKey struct {
	providerSlug string
	environment  string
}

const maxSnapshotEntries = 1024

type Manager struct {
	loader    ActiveCredentialLoader
	decrypter CredentialDecrypter
	metrics   Metrics
	now       func() time.Time
	state     atomic.Pointer[snapshot]
	loads     singleflight.Group
}

func NewManager(loader ActiveCredentialLoader, decrypter CredentialDecrypter, metrics Metrics) *Manager {
	if metrics == nil {
		metrics = nopMetrics{}
	}
	m := &Manager{loader: loader, decrypter: decrypter, metrics: metrics, now: time.Now}
	m.state.Store(&snapshot{entries: map[credentialCacheKey]ResolvedCredential{}})
	return m
}

// Resolve is deliberately context-free: a warm resolve is one atomic pointer
// load plus one immutable map lookup. It cannot accidentally perform I/O.
func (m *Manager) Resolve(providerSlug, environment string) (ResolvedCredential, error) {
	entry, ok := m.state.Load().entries[makeCredentialCacheKey(providerSlug, environment)]
	if !ok {
		m.metrics.ObserveResolve("unavailable")
		return ResolvedCredential{}, ErrUnavailable
	}
	if entry.ExpiresAt != nil && !entry.ExpiresAt.After(m.now()) {
		m.metrics.ObserveResolve("expired")
		return ResolvedCredential{}, ErrExpired
	}
	m.metrics.ObserveResolve("hit")
	return entry, nil
}

func (m *Manager) Reload(ctx context.Context) error {
	_, err, _ := m.loads.Do("active-snapshot", func() (interface{}, error) {
		started := m.now()
		records, err := m.loader.LoadActivePlatformCredentials(ctx)
		if err != nil {
			m.metrics.ObserveSnapshotLoad(false, m.now().Sub(started), 0)
			return nil, err
		}
		if len(records) > maxSnapshotEntries {
			m.metrics.ObserveSnapshotLoad(false, m.now().Sub(started), 0)
			return nil, ErrSnapshotTooLarge
		}
		entries := make(map[credentialCacheKey]ResolvedCredential, len(records))
		for _, record := range records {
			secrets, err := m.decrypter.Decrypt(record)
			if err != nil {
				m.metrics.ObserveSnapshotLoad(false, m.now().Sub(started), 0)
				return nil, ErrDecryptFailed
			}
			secret := runtimeAuthSecret(record, secrets)
			clearSecretMap(secrets)
			if strings.TrimSpace(secret) == "" {
				m.metrics.ObserveSnapshotLoad(false, m.now().Sub(started), 0)
				return nil, ErrDecryptFailed
			}
			entries[makeCredentialCacheKey(record.ProviderSlug, record.Environment)] = ResolvedCredential{
				ProviderSlug: record.ProviderSlug, Environment: record.Environment,
				ConnectionID: record.ConnectionID, CredentialVersionID: record.CredentialVersionID,
				CredentialVersion: record.CredentialVersion, AuthMode: record.AuthMode,
				APIKey: secret, BaseURL: record.BaseURL, APIVersion: record.APIVersion,
				Organization: record.Organization, Project: record.Project,
				HTTPReferer: record.HTTPReferer, AppTitle: record.AppTitle, ClientName: record.ClientName,
				ExpiresAt: record.ExpiresAt,
			}
		}
		m.state.Store(&snapshot{entries: entries, loadedAt: m.now()})
		m.metrics.ObserveSnapshotLoad(true, m.now().Sub(started), len(entries))
		return nil, nil
	})
	return err
}

// Invalidate removes one provider/environment entry without I/O. Lifecycle
// mutations and Pub/Sub consumers call it before attempting a reload so a
// known-changed credential is never served stale when PostgreSQL is degraded.
func (m *Manager) Invalidate(providerSlug, environment string) bool {
	key := makeCredentialCacheKey(providerSlug, environment)
	for {
		current := m.state.Load()
		if _, exists := current.entries[key]; !exists {
			return false
		}
		entries := make(map[credentialCacheKey]ResolvedCredential, len(current.entries)-1)
		for existingKey, value := range current.entries {
			if existingKey != key {
				entries[existingKey] = value
			}
		}
		next := &snapshot{entries: entries, loadedAt: current.loadedAt}
		if m.state.CompareAndSwap(current, next) {
			return true
		}
	}
}

func runtimeAuthSecret(record EncryptedCredential, secrets map[string]string) string {
	switch normalizeProvider(record.ProviderSlug) {
	case "databricks":
		if record.AuthMode == "personal_access_token" {
			return secrets["access_token"]
		}
	default:
		if record.AuthMode == "api_key" {
			return secrets["api_key"]
		}
	}
	return ""
}

// Reconcile continuously repairs missed Pub/Sub events. The interval is capped
// at 30 seconds so two healthy nodes converge well inside the 60 second SLO.
func (m *Manager) Reconcile(ctx context.Context, interval time.Duration) {
	if interval <= 0 || interval > 30*time.Second {
		interval = 30 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = m.Reload(ctx)
		}
	}
}

func (m *Manager) EntryCount() int { return len(m.state.Load().entries) }
func (m *Manager) Ready() bool     { return !m.state.Load().loadedAt.IsZero() }

func makeCredentialCacheKey(providerSlug, environment string) credentialCacheKey {
	return credentialCacheKey{
		providerSlug: strings.ToLower(strings.TrimSpace(providerSlug)),
		environment:  strings.ToLower(strings.TrimSpace(environment)),
	}
}
