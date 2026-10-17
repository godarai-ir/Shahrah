package runtimeprofile

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"go.uber.org/zap"
)

func TestManagerReturnsL1ProfileWithoutL2OrStore(t *testing.T) {
	manager := NewManager(nil, nil, nil, zap.NewNop(), Config{StaleGrace: time.Minute, VersionCheckInterval: time.Minute})
	manager.setL1("tenant-1", &Profile{TenantID: "tenant-1", Version: 4})

	profile, err := manager.Get(context.Background(), "tenant-1")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}

	if profile.Version != 4 {
		t.Fatalf("expected L1 version 4, got %d", profile.Version)
	}
}

func TestManagerSnapshotIsBorrowedWhileGetRemainsIsolated(t *testing.T) {
	manager := NewManager(nil, nil, nil, zap.NewNop(), Config{StaleGrace: time.Minute, VersionCheckInterval: time.Minute})
	manager.setL1("tenant-1", &Profile{
		TenantID:        "tenant-1",
		Version:         4,
		ProviderAPIKeys: map[string]string{"deployment-1": "secret"},
	})

	first, err := manager.GetSnapshot(context.Background(), "tenant-1")
	if err != nil {
		t.Fatalf("GetSnapshot returned error: %v", err)
	}

	second, err := manager.GetSnapshot(context.Background(), "tenant-1")
	if err != nil {
		t.Fatalf("GetSnapshot returned error: %v", err)
	}

	if first != second {
		t.Fatal("expected hot-path snapshot reads to reuse the immutable profile")
	}

	isolated, err := manager.Get(context.Background(), "tenant-1")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}

	isolated.ProviderAPIKeys["deployment-1"] = "changed"
	if first.ProviderAPIKeys["deployment-1"] != "secret" {
		t.Fatal("expected compatibility Get to return an isolated clone")
	}
}

func TestManagerVersionCheckDoesNotDeleteNewerConcurrentSnapshot(t *testing.T) {
	cache := &blockingVersionCache{
		started: make(chan struct{}),
		release: make(chan struct{}),
	}
	manager := NewManager(nil, nil, cache, zap.NewNop(), Config{
		StaleGrace:           time.Minute,
		VersionCheckInterval: time.Hour,
	})
	manager.setL1("tenant-1", &Profile{TenantID: "tenant-1", Version: 1})

	manager.mu.Lock()
	entry := manager.l1["tenant-1"]
	entry.loadedAt = time.Now().Add(-2 * time.Hour)
	manager.l1["tenant-1"] = entry
	manager.mu.Unlock()

	type lookupResult struct {
		profile *Profile
		ok      bool
	}

	result := make(chan lookupResult, 1)

	go func() {
		profile, ok := manager.getL1(t.Context(), "tenant-1")
		result <- lookupResult{profile: profile, ok: ok}
	}()

	<-cache.started
	manager.setL1("tenant-1", &Profile{TenantID: "tenant-1", Version: 2})
	close(cache.release)

	loaded := <-result
	if !loaded.ok || loaded.profile == nil || loaded.profile.Version != 2 {
		t.Fatalf("expected concurrent version 2 snapshot to survive, got profile=%+v ok=%t", loaded.profile, loaded.ok)
	}
}

func BenchmarkManagerProfileReads(b *testing.B) {
	manager := NewManager(nil, nil, nil, zap.NewNop(), Config{StaleGrace: time.Hour, VersionCheckInterval: time.Hour})
	manager.setL1("tenant-1", benchmarkProfile())

	ctx := context.Background()

	b.Run("isolated_clone", func(b *testing.B) {
		b.ReportAllocs()

		for b.Loop() {
			if _, err := manager.Get(ctx, "tenant-1"); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("immutable_snapshot", func(b *testing.B) {
		b.ReportAllocs()

		for b.Loop() {
			if _, err := manager.GetSnapshot(ctx, "tenant-1"); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func benchmarkProfile() *Profile {
	profile := &Profile{
		TenantID:        "tenant-1",
		Version:         1,
		AuthByTokenHash: make(map[string]platform.AuthContext, 100),
		VirtualModels:   make(map[string]platform.VirtualModel, 20),
		Deployments:     make([]platform.Deployment, 20),
		Guardrails:      make(map[string]platform.TenantGuardrail, 10),
		RateLimitRules:  make([]platform.RateLimitRule, 10),
		BudgetRules:     make([]platform.BudgetLimitRule, 10),
		ProviderAPIKeys: make(map[string]string, 20),
	}

	for i := range 100 {
		key := "token-" + strconv.Itoa(i)
		profile.AuthByTokenHash[key] = platform.AuthContext{
			TenantID:          profile.TenantID,
			KeyID:             "key-" + strconv.Itoa(i),
			Permissions:       []string{"models:read", "models:invoke"},
			AllowedModels:     []string{"openai:account:gpt-4o", "anthropic:account:claude"},
			AllowedModelTypes: map[string][]string{"openai:account:gpt-4o": {"chat"}},
		}
	}

	for i := range 20 {
		id := strconv.Itoa(i)
		profile.Deployments[i] = platform.Deployment{ID: "deployment-" + id, ModelName: "model-" + id, ModelInfo: map[string]interface{}{"context_window": 128000}}
		profile.ProviderAPIKeys["deployment-"+id] = "secret"
	}

	for i := range 10 {
		profile.RateLimitRules[i] = platform.RateLimitRule{ID: "rate-" + strconv.Itoa(i), Conditions: map[string]interface{}{"model": "model"}, Limit: map[string]interface{}{"rpm": 1000}}
		profile.BudgetRules[i] = platform.BudgetLimitRule{ID: "budget-" + strconv.Itoa(i), Conditions: map[string]interface{}{"model": "model"}, Alerts: map[string]interface{}{"threshold": 80}}
	}

	return profile
}

func TestManagerLoadsProfileFromL2(t *testing.T) {
	cache := newFakeDistributedCache()
	profile := &Profile{SchemaVersion: ProfileSchemaVersion, TenantID: "tenant-1", Version: 9}
	raw, _ := json.Marshal(profile)
	cache.values[ActiveVersionKey("tenant-1")] = "9"
	cache.values[ProfileKey("tenant-1", 9)] = string(raw)

	manager := NewManager(nil, nil, cache, zap.NewNop(), Config{StaleGrace: time.Minute, VersionCheckInterval: time.Minute})

	loaded, err := manager.Get(context.Background(), "tenant-1")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}

	if loaded.Version != 9 {
		t.Fatalf("expected L2 version 9, got %d", loaded.Version)
	}

	if _, ok := manager.getL1(context.Background(), "tenant-1"); !ok {
		t.Fatal("expected L2 load to populate L1")
	}
}

func TestManagerReturnsUnavailableWithoutL1L2OrCompiler(t *testing.T) {
	manager := NewManager(nil, nil, nil, zap.NewNop(), Config{StaleGrace: time.Minute})

	_, err := manager.Get(context.Background(), "tenant-1")
	if !errors.Is(err, ErrProfileUnavailable) {
		t.Fatalf("expected ErrProfileUnavailable, got %v", err)
	}
}

func TestManagerCompilesAndPublishesOnColdMiss(t *testing.T) {
	repo := &fakeSourceRepo{tenant: testTenant("tenant-1")}
	compiler := NewCompiler(repo)
	store := &fakeProfileStore{}
	cache := newFakeDistributedCache()
	manager := NewManager(compiler, store, cache, zap.NewNop(), Config{StaleGrace: time.Minute, VersionCheckInterval: time.Minute})

	profile, err := manager.Get(context.Background(), "tenant-1")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}

	if profile.Version != 1 || store.saved == nil {
		t.Fatalf("expected cold miss to compile version 1, got profile=%+v saved=%+v", profile, store.saved)
	}

	if cache.values[ActiveVersionKey("tenant-1")] != "1" {
		t.Fatalf("expected active version in cache, got %q", cache.values[ActiveVersionKey("tenant-1")])
	}
}

func TestManagerCompileAndPublishPersistsAndPublishes(t *testing.T) {
	repo := &fakeSourceRepo{tenant: testTenant("tenant-1")}
	compiler := NewCompiler(repo)
	store := &fakeProfileStore{}
	cache := newFakeDistributedCache()
	manager := NewManager(compiler, store, cache, zap.NewNop(), Config{StaleGrace: time.Minute, VersionCheckInterval: time.Minute})

	profile, err := manager.CompileAndPublish(context.Background(), "tenant-1", "test")
	if err != nil {
		t.Fatalf("CompileAndPublish returned error: %v", err)
	}

	if profile.Version != 1 || store.saved == nil {
		t.Fatalf("expected version 1 to be saved, got profile=%+v saved=%+v", profile, store.saved)
	}

	if cache.values[ActiveVersionKey("tenant-1")] != "1" {
		t.Fatalf("expected active version in cache, got %q", cache.values[ActiveVersionKey("tenant-1")])
	}

	if cache.published == "" {
		t.Fatal("expected profile event to be published")
	}
}

func testTenant(id string) platform.Tenant {
	return platform.Tenant{ID: id, Slug: "tenant", Domain: "tenant.localhost", Status: "active", CreatedAt: time.Now()}
}

type fakeProfileStore struct {
	saved *StoredProfile
}

func (s *fakeProfileStore) NextTenantRuntimeProfileVersion(context.Context, string) (int64, error) {
	if s.saved == nil {
		return 1, nil
	}

	return s.saved.Version + 1, nil
}

func (s *fakeProfileStore) SaveTenantRuntimeProfile(_ context.Context, stored StoredProfile) error {
	storedCopy := stored
	s.saved = &storedCopy

	return nil
}

func (s *fakeProfileStore) LoadLatestTenantRuntimeProfile(context.Context, string) (*StoredProfile, error) {
	if s.saved == nil {
		return nil, ErrProfileUnavailable
	}

	return s.saved, nil
}

type fakeDistributedCache struct {
	values    map[string]string
	published string
}

type blockingVersionCache struct {
	started chan struct{}
	release chan struct{}
	once    sync.Once
}

func (c *blockingVersionCache) DistributedGet(context.Context, string) (string, error) {
	c.once.Do(func() { close(c.started) })
	<-c.release

	return "2", nil
}

func (*blockingVersionCache) DistributedSet(context.Context, string, interface{}, time.Duration) error {
	return nil
}

func newFakeDistributedCache() *fakeDistributedCache {
	return &fakeDistributedCache{values: map[string]string{}}
}

func (c *fakeDistributedCache) DistributedGet(_ context.Context, key string) (string, error) {
	value, ok := c.values[key]
	if !ok {
		return "", ErrProfileUnavailable
	}

	return value, nil
}

func (c *fakeDistributedCache) DistributedSet(_ context.Context, key string, value interface{}, _ time.Duration) error {
	c.values[key] = value.(string)
	return nil
}

func (c *fakeDistributedCache) Publish(_ context.Context, _ string, message string) error {
	c.published = message
	return nil
}

func TestManagerDiscardsProfilesFromAnOlderSchema(t *testing.T) {
	cache := newFakeDistributedCache()
	stale := &Profile{TenantID: "tenant-1", Version: 9, Tenant: platform.Tenant{ID: "tenant-1"}}
	raw, _ := json.Marshal(stale)
	cache.values[ActiveVersionKey("tenant-1")] = "9"
	cache.values[ProfileKey("tenant-1", 9)] = string(raw)

	manager := NewManager(nil, nil, cache, zap.NewNop(), Config{StaleGrace: time.Minute, VersionCheckInterval: time.Minute})

	if _, err := manager.Get(context.Background(), "tenant-1"); err == nil {
		t.Fatal("a profile compiled before the billing unit existed must not be served")
	}
}
