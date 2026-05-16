package storage

import (
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

type authzCacheEntry struct {
	auth      platform.AuthContext
	expiresAt time.Time
	epoch     uint64
}

func (s *Store) authzCacheKey(tenantID string, userID string) string {
	return strings.TrimSpace(strings.ToLower(tenantID)) + ":" + strings.TrimSpace(strings.ToLower(userID))
}

func (s *Store) getAuthzCacheEntry(tenantID string, userID string) (*platform.AuthContext, bool) {
	key := s.authzCacheKey(tenantID, userID)
	now := time.Now()

	s.authzCacheMu.RLock()
	entry, ok := s.authzCache[key]
	currentEpoch := s.authzTenantEpoch[strings.TrimSpace(strings.ToLower(tenantID))]
	s.authzCacheMu.RUnlock()
	if !ok {
		return nil, false
	}
	if now.After(entry.expiresAt) || entry.epoch != currentEpoch {
		s.authzCacheMu.Lock()
		delete(s.authzCache, key)
		s.authzCacheMu.Unlock()
		return nil, false
	}

	copyAuth := entry.auth
	return &copyAuth, true
}

func (s *Store) setAuthzCacheEntry(tenantID string, userID string, auth platform.AuthContext) {
	key := s.authzCacheKey(tenantID, userID)
	tenantKey := strings.TrimSpace(strings.ToLower(tenantID))
	now := time.Now()

	s.authzCacheMu.Lock()
	defer s.authzCacheMu.Unlock()

	if len(s.authzCache) >= s.authzCacheMaxEntries {
		for cacheKey, entry := range s.authzCache {
			if now.After(entry.expiresAt) {
				delete(s.authzCache, cacheKey)
			}
		}
		if len(s.authzCache) >= s.authzCacheMaxEntries {
			for cacheKey := range s.authzCache {
				delete(s.authzCache, cacheKey)
				break
			}
		}
	}

	s.authzCache[key] = authzCacheEntry{
		auth:      auth,
		expiresAt: now.Add(s.authzCacheTTL),
		epoch:     s.authzTenantEpoch[tenantKey],
	}
}

func (s *Store) invalidateAuthzCacheForUser(tenantID string, userID string) {
	key := s.authzCacheKey(tenantID, userID)
	s.authzCacheMu.Lock()
	delete(s.authzCache, key)
	s.authzCacheMu.Unlock()
}

func (s *Store) bumpAuthzTenantEpoch(tenantID string) {
	tenantKey := strings.TrimSpace(strings.ToLower(tenantID))
	s.authzCacheMu.Lock()
	s.authzTenantEpoch[tenantKey] = s.authzTenantEpoch[tenantKey] + 1
	s.authzCacheMu.Unlock()
}
