package runtimeprofile

import (
	"context"
	"encoding/json"
	"errors"
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

func TestManagerLoadsProfileFromL2(t *testing.T) {
	cache := newFakeDistributedCache()
	profile := &Profile{TenantID: "tenant-1", Version: 9}
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
