package runtimeprofile

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"
)

type Store interface {
	NextTenantRuntimeProfileVersion(ctx context.Context, tenantID string) (int64, error)
	SaveTenantRuntimeProfile(ctx context.Context, stored StoredProfile) error
	LoadLatestTenantRuntimeProfile(ctx context.Context, tenantID string) (*StoredProfile, error)
}

type DistributedCache interface {
	DistributedGet(ctx context.Context, key string) (string, error)
	DistributedSet(ctx context.Context, key string, value interface{}, ttl time.Duration) error
}

type Publisher interface {
	Publish(ctx context.Context, channel string, message string) error
}

type Config struct {
	VersionCheckInterval time.Duration
	StaleGrace           time.Duration
	L2TTL                time.Duration
}

type Manager struct {
	compiler *Compiler
	store    Store
	cache    DistributedCache
	pub      Publisher
	logger   *zap.Logger
	cfg      Config

	mu    sync.RWMutex
	l1    map[string]l1Entry
	group singleflight.Group
}

type l1Entry struct {
	profile   *Profile
	loadedAt  time.Time
	expiresAt time.Time
}

func NewManager(compiler *Compiler, store Store, cache DistributedCache, logger *zap.Logger, cfg Config) *Manager {
	if cfg.VersionCheckInterval <= 0 {
		cfg.VersionCheckInterval = time.Second
	}

	if cfg.StaleGrace <= 0 {
		cfg.StaleGrace = 300 * time.Second
	}

	if cfg.L2TTL <= 0 {
		cfg.L2TTL = time.Hour
	}

	var pub Publisher
	if p, ok := cache.(Publisher); ok {
		pub = p
	}

	return &Manager{
		compiler: compiler,
		store:    store,
		cache:    cache,
		pub:      pub,
		logger:   logger,
		cfg:      cfg,
		l1:       make(map[string]l1Entry),
	}
}

func (m *Manager) Get(ctx context.Context, tenantID string) (*Profile, error) {
	profile, err := m.GetSnapshot(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	return cloneProfile(profile), nil
}

// GetSnapshot returns the immutable in-memory profile snapshot for hot-path readers.
// Callers must never mutate the returned profile or any nested map or slice.
func (m *Manager) GetSnapshot(ctx context.Context, tenantID string) (*Profile, error) {
	if m == nil {
		return nil, ErrProfileUnavailable
	}

	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return nil, ErrProfileUnavailable
	}

	if profile, ok := m.getL1(ctx, tenantID); ok {
		return profile, nil
	}

	value, err, _ := m.group.Do(tenantID, func() (interface{}, error) {
		if profile, ok := m.getL1(ctx, tenantID); ok {
			return profile, nil
		}

		if profile, err := m.loadFromL2(ctx, tenantID); err == nil && profile != nil {
			profile = m.hydrateL1Secrets(ctx, tenantID, profile)
			m.setL1(tenantID, profile)

			return profile, nil
		}

		if profile, err := m.loadFromStore(ctx, tenantID); err == nil && profile != nil {
			profile = m.hydrateL1Secrets(ctx, tenantID, profile)
			m.setL1(tenantID, profile)

			if writeErr := m.writeL2(ctx, profile); writeErr != nil {
				m.logger.Warn("failed to refresh runtime profile L2 cache", zap.Error(writeErr), zap.String("tenant_id", tenantID))
			}

			return profile, nil
		}

		if m.compiler != nil && m.store != nil {
			profile, err := m.CompileAndPublish(ctx, tenantID, "cold_miss")
			if err == nil {
				return profile, nil
			}
		}

		return nil, ErrProfileUnavailable
	})
	if err != nil {
		return nil, err
	}

	profile, ok := value.(*Profile)
	if !ok || profile == nil {
		return nil, ErrProfileUnavailable
	}

	return profile, nil
}

func (m *Manager) GetAuthByTokenHash(ctx context.Context, tenantID, tokenHash string) (*platform.AuthContext, error) {
	profile, err := m.GetSnapshot(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	authContext, ok := profile.AuthByTokenHash[strings.TrimSpace(tokenHash)]
	if !ok {
		return nil, ErrAuthNotFound
	}

	cloned := cloneAuthContext(authContext)

	return &cloned, nil
}

func (m *Manager) Invalidate(ctx context.Context, tenantID string, reason string) {
	if m == nil {
		return
	}

	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return
	}

	m.mu.Lock()
	delete(m.l1, tenantID)

	m.mu.Unlock()
	go func() {
		_, err := m.CompileAndPublish(context.Background(), tenantID, reason)
		if err != nil && m.logger != nil {
			m.logger.Warn("tenant runtime profile warm compile failed", zap.String("tenant_id", tenantID), zap.Error(err))
		}
	}()

	_ = ctx
}

func (m *Manager) CompileAndPublish(ctx context.Context, tenantID string, reason string) (*Profile, error) {
	if m == nil {
		return nil, ErrProfileUnavailable
	}

	if m.compiler == nil || m.store == nil {
		return nil, ErrProfileUnavailable
	}

	version, err := m.store.NextTenantRuntimeProfileVersion(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	profile, err := m.compiler.Compile(ctx, tenantID, version)
	if err != nil {
		return nil, err
	}

	raw, err := json.Marshal(profile)
	if err != nil {
		return nil, err
	}

	if err := m.store.SaveTenantRuntimeProfile(ctx, StoredProfile{
		TenantID:        tenantID,
		Version:         profile.Version,
		ProfileJSON:     raw,
		Checksum:        profile.Checksum,
		CompiledAt:      profile.CompiledAt,
		SourceUpdatedAt: profile.SourceUpdatedAt,
		Status:          ProfileStatusActive,
	}); err != nil {
		return nil, err
	}

	m.setL1(tenantID, profile)

	if err := m.writeL2(ctx, profile); err != nil {
		m.logger.Warn("failed to update runtime profile L2 cache", zap.Error(err), zap.String("tenant_id", tenantID))
	}

	if err := m.publish(ctx, Event{TenantID: tenantID, Version: profile.Version, Reason: reason}); err != nil {
		m.logger.Warn("failed to publish runtime profile update", zap.Error(err), zap.String("tenant_id", tenantID))
	}

	return cloneProfile(profile), nil
}

func (m *Manager) getL1(ctx context.Context, tenantID string) (*Profile, bool) {
	for {
		m.mu.RLock()
		entry, ok := m.l1[tenantID]
		m.mu.RUnlock()

		if !ok || entry.profile == nil {
			return nil, false
		}

		now := time.Now()
		if !entry.expiresAt.IsZero() && now.After(entry.expiresAt) {
			if !m.deleteObservedL1Entry(tenantID, entry.profile) {
				continue
			}

			return nil, false
		}

		if m.cache != nil && now.Sub(entry.loadedAt) >= m.cfg.VersionCheckInterval {
			if active, err := m.activeVersion(ctx, tenantID); err == nil && active > entry.profile.Version {
				if !m.deleteObservedL1Entry(tenantID, entry.profile) {
					continue
				}

				return nil, false
			}

			currentUnchanged := false

			m.mu.Lock()
			if current, exists := m.l1[tenantID]; exists && current.profile == entry.profile {
				current.loadedAt = now
				m.l1[tenantID] = current
				currentUnchanged = true
			}
			m.mu.Unlock()

			if !currentUnchanged {
				continue
			}
		}

		return entry.profile, true
	}
}

func (m *Manager) deleteObservedL1Entry(tenantID string, observed *Profile) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	current, ok := m.l1[tenantID]
	if !ok || current.profile != observed {
		return false
	}

	delete(m.l1, tenantID)

	return true
}

func (m *Manager) setL1(tenantID string, profile *Profile) {
	if profile == nil {
		return
	}

	snapshot := cloneProfile(profile)
	sortProfilePolicies(snapshot)

	m.mu.Lock()
	m.l1[tenantID] = l1Entry{profile: snapshot, loadedAt: time.Now(), expiresAt: time.Now().Add(m.cfg.StaleGrace)}
	m.mu.Unlock()
}

func (m *Manager) activeVersion(ctx context.Context, tenantID string) (int64, error) {
	if m.cache == nil {
		return 0, ErrProfileUnavailable
	}

	raw, err := m.cache.DistributedGet(ctx, ActiveVersionKey(tenantID))
	if err != nil {
		return 0, err
	}

	return parseVersion(raw)
}

func (m *Manager) loadFromL2(ctx context.Context, tenantID string) (*Profile, error) {
	version, err := m.activeVersion(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	raw, err := m.cache.DistributedGet(ctx, ProfileKey(tenantID, version))
	if err != nil {
		return nil, err
	}

	return decodeProfile([]byte(raw))
}

func (m *Manager) loadFromStore(ctx context.Context, tenantID string) (*Profile, error) {
	if m.store == nil {
		return nil, ErrProfileUnavailable
	}

	stored, err := m.store.LoadLatestTenantRuntimeProfile(ctx, tenantID)
	if err != nil || stored == nil {
		return nil, err
	}

	return decodeProfile(stored.ProfileJSON)
}

func (m *Manager) writeL2(ctx context.Context, profile *Profile) error {
	if m.cache == nil || profile == nil {
		return nil
	}

	redacted := cloneProfile(profile)
	redacted.ProviderAPIKeys = nil

	raw, err := json.Marshal(redacted)
	if err != nil {
		return err
	}

	if err := m.cache.DistributedSet(ctx, ProfileKey(profile.TenantID, profile.Version), string(raw), m.cfg.L2TTL); err != nil {
		return err
	}

	return m.cache.DistributedSet(ctx, ActiveVersionKey(profile.TenantID), formatVersion(profile.Version), m.cfg.L2TTL)
}

func (m *Manager) hydrateL1Secrets(ctx context.Context, tenantID string, profile *Profile) *Profile {
	if profile == nil || m.compiler == nil || len(profile.ProviderAPIKeys) > 0 || len(profile.Deployments) == 0 {
		return profile
	}

	hydrated, err := m.compiler.Compile(ctx, tenantID, profile.Version)
	if err != nil {
		if m.logger != nil {
			m.logger.Warn("tenant runtime profile secret hydration failed", zap.String("tenant_id", tenantID), zap.Error(err))
		}

		return profile
	}

	hydrated.Checksum = profile.Checksum
	hydrated.CompiledAt = profile.CompiledAt
	hydrated.SourceUpdatedAt = profile.SourceUpdatedAt

	return hydrated
}

func (m *Manager) publish(ctx context.Context, event Event) error {
	if m.pub == nil {
		return nil
	}

	raw, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return m.pub.Publish(ctx, RedisChannel, string(raw))
}

func decodeProfile(raw []byte) (*Profile, error) {
	if len(raw) == 0 {
		return nil, ErrProfileUnavailable
	}

	var profile Profile
	if err := json.Unmarshal(raw, &profile); err != nil {
		return nil, err
	}

	if profile.SchemaVersion < ProfileSchemaVersion {
		return nil, ErrProfileUnavailable
	}

	if profile.ProviderAPIKeys == nil {
		profile.ProviderAPIKeys = map[string]string{}
	}

	return &profile, nil
}
