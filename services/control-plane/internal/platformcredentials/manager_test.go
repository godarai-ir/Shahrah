package platformcredentials

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type fakeLoader struct {
	calls   atomic.Int64
	records []EncryptedCredential
	gate    <-chan struct{}
}

func (f *fakeLoader) LoadActivePlatformCredentials(context.Context) ([]EncryptedCredential, error) {
	f.calls.Add(1)

	if f.gate != nil {
		<-f.gate
	}

	return append([]EncryptedCredential(nil), f.records...), nil
}

type fakeDecrypter struct {
	calls   atomic.Int64
	secrets map[string]string
}

func (f *fakeDecrypter) Decrypt(EncryptedCredential) (map[string]string, error) {
	f.calls.Add(1)

	copyOfSecrets := make(map[string]string, len(f.secrets))
	for key, value := range f.secrets {
		copyOfSecrets[key] = value
	}

	return copyOfSecrets, nil
}

func activeRecord(version int64) EncryptedCredential {
	return EncryptedCredential{Binding: Binding{
		ProviderSlug: "openai", Environment: "production", ConnectionID: "connection-1",
		CredentialVersionID: "credential-version", CredentialVersion: version, AuthMode: "api_key", SchemaVersion: 1,
	}}
}

func TestManagerWarmResolvePerformsNoLoadsOrDecrypts(t *testing.T) {
	loader := &fakeLoader{records: []EncryptedCredential{activeRecord(1)}}
	decrypter := &fakeDecrypter{secrets: map[string]string{"api_key": "sk-runtime"}}
	manager := NewManager(loader, decrypter, nil)
	require.NoError(t, manager.Reload(context.Background()))
	loader.calls.Store(0)
	decrypter.calls.Store(0)

	const resolves = 10000

	var group sync.WaitGroup
	group.Add(resolves)
	errors := make(chan error, resolves)

	for range resolves {
		go func() {
			defer group.Done()

			resolved, err := manager.Resolve("openai", "production")
			if err != nil {
				errors <- err
				return
			}

			if resolved.APIKey != "sk-runtime" {
				errors <- fmt.Errorf("unexpected API key %q", resolved.APIKey)
			}
		}()
	}

	group.Wait()
	close(errors)

	for err := range errors {
		require.NoError(t, err)
	}

	require.Zero(t, loader.calls.Load())
	require.Zero(t, decrypter.calls.Load())
}

func TestManagerCollapsesConcurrentReloadsAndSwapsAtomically(t *testing.T) {
	gate := make(chan struct{})
	loader := &fakeLoader{records: []EncryptedCredential{activeRecord(2)}, gate: gate}
	decrypter := &fakeDecrypter{secrets: map[string]string{"api_key": "sk-v2"}}
	manager := NewManager(loader, decrypter, nil)

	const reloads = 50

	var group sync.WaitGroup
	group.Add(reloads)
	errors := make(chan error, reloads)

	for range reloads {
		go func() {
			defer group.Done()

			if err := manager.Reload(context.Background()); err != nil {
				errors <- err
			}
		}()
	}

	time.Sleep(20 * time.Millisecond)
	close(gate)
	group.Wait()
	close(errors)

	for err := range errors {
		require.NoError(t, err)
	}

	require.Equal(t, int64(1), loader.calls.Load())
	require.Equal(t, int64(1), decrypter.calls.Load())

	resolved, err := manager.Resolve("openai", "production")
	require.NoError(t, err)
	require.Equal(t, int64(2), resolved.CredentialVersion)
	require.Equal(t, "sk-v2", resolved.APIKey)
}

func TestManagerReloadRemovesDisabledCredentialImmediately(t *testing.T) {
	loader := &fakeLoader{records: []EncryptedCredential{activeRecord(1)}}
	decrypter := &fakeDecrypter{secrets: map[string]string{"api_key": "sk-v1"}}
	manager := NewManager(loader, decrypter, nil)
	require.NoError(t, manager.Reload(context.Background()))

	loader.records = nil

	require.NoError(t, manager.Reload(context.Background()))
	_, err := manager.Resolve("openai", "production")
	require.ErrorIs(t, err, ErrUnavailable)
}

func TestManagerInvalidationFailsClosedWithoutRepositoryIO(t *testing.T) {
	loader := &fakeLoader{records: []EncryptedCredential{activeRecord(1)}}
	decrypter := &fakeDecrypter{secrets: map[string]string{"api_key": "sk-v1"}}
	manager := NewManager(loader, decrypter, nil)
	require.NoError(t, manager.Reload(context.Background()))
	loader.calls.Store(0)
	decrypter.calls.Store(0)

	require.True(t, manager.Invalidate("openai", "production"))
	_, err := manager.Resolve("openai", "production")
	require.ErrorIs(t, err, ErrUnavailable)
	require.Zero(t, loader.calls.Load())
	require.Zero(t, decrypter.calls.Load())
}

func TestManagerFailsExpiredCredentialClosed(t *testing.T) {
	expired := time.Now().Add(-time.Second)
	record := activeRecord(1)
	record.ExpiresAt = &expired
	manager := NewManager(&fakeLoader{records: []EncryptedCredential{record}}, &fakeDecrypter{secrets: map[string]string{"api_key": "sk-old"}}, nil)
	require.NoError(t, manager.Reload(context.Background()))
	_, err := manager.Resolve("openai", "production")
	require.ErrorIs(t, err, ErrExpired)
}

func TestManagerRejectsOversizedSnapshotWithoutReplacingWarmState(t *testing.T) {
	loader := &fakeLoader{records: []EncryptedCredential{activeRecord(1)}}
	decrypter := &fakeDecrypter{secrets: map[string]string{"api_key": "sk-v1"}}
	manager := NewManager(loader, decrypter, nil)
	require.NoError(t, manager.Reload(context.Background()))

	loader.records = make([]EncryptedCredential, maxSnapshotEntries+1)

	require.ErrorIs(t, manager.Reload(context.Background()), ErrSnapshotTooLarge)

	resolved, err := manager.Resolve("openai", "production")
	require.NoError(t, err)
	require.Equal(t, "sk-v1", resolved.APIKey)
	require.Equal(t, int64(1), decrypter.calls.Load())
}

func BenchmarkPlatformCredentialResolve(b *testing.B) {
	manager := NewManager(&fakeLoader{records: []EncryptedCredential{activeRecord(1)}}, &fakeDecrypter{secrets: map[string]string{"api_key": "sk-runtime"}}, nil)
	require.NoError(b, manager.Reload(context.Background()))
	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		_, err := manager.Resolve("openai", "production")
		if err != nil {
			b.Fatal(err)
		}
	}
}
