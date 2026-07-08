package storage

import (
	"context"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"golang.org/x/sync/singleflight"
)

type hotPathCacheState int

const (
	hotPathCacheMiss hotPathCacheState = iota
	hotPathCacheFresh
	hotPathCacheStale
)

type hotPathCacheEntry struct {
	value      interface{}
	freshUntil time.Time
	staleUntil time.Time
}

func deploymentCacheKey(tenantID string) string {
	return "deployments:tenant:" + strings.TrimSpace(tenantID)
}

func virtualModelsCacheKey(tenantID string) string {
	return "virtual-models:tenant:" + strings.TrimSpace(tenantID)
}

func walletCacheKey(tenantID string) string {
	return "wallet:tenant:" + strings.TrimSpace(tenantID)
}

func (s *Store) invalidateDeploymentCache(ctx context.Context, tenantID string) {
	s.deleteLocalHotPathCache(deploymentCacheKey(tenantID))
	if s == nil || s.entitlementCache == nil {
		return
	}
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return
	}
	s.entitlementCache.Delete(ctx, deploymentCacheKey(tenantID))
}

func (s *Store) invalidateVirtualModelsCache(_ context.Context, tenantID string) {
	s.deleteLocalHotPathCache(virtualModelsCacheKey(tenantID))
}

func (s *Store) invalidateWalletCache(ctx context.Context, tenantID string) {
	if s == nil || s.entitlementCache == nil {
		return
	}
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return
	}
	s.entitlementCache.Delete(ctx, walletCacheKey(tenantID))
	s.syncCurrentBudgetWindowAllowance(ctx, tenantID)
}

func (s *Store) getCachedDeployments(key string) ([]platform.Deployment, hotPathCacheState) {
	value, state := s.getLocalHotPathCache(key)
	if state == hotPathCacheMiss {
		return nil, hotPathCacheMiss
	}

	deps, ok := value.([]platform.Deployment)
	if !ok {
		s.deleteLocalHotPathCache(key)
		return nil, hotPathCacheMiss
	}

	return cloneDeployments(deps), state
}

func (s *Store) cacheDeployments(key string, deps []platform.Deployment) {
	s.setLocalHotPathCache(key, cloneDeployments(deps))
}

func (s *Store) getCachedVirtualModels(key string) ([]platform.VirtualModel, hotPathCacheState) {
	value, state := s.getLocalHotPathCache(key)
	if state == hotPathCacheMiss {
		return nil, hotPathCacheMiss
	}

	models, ok := value.([]platform.VirtualModel)
	if !ok {
		s.deleteLocalHotPathCache(key)
		return nil, hotPathCacheMiss
	}

	return cloneVirtualModels(models), state
}

func (s *Store) cacheVirtualModels(key string, models []platform.VirtualModel) {
	s.setLocalHotPathCache(key, cloneVirtualModels(models))
}

func (s *Store) getLocalHotPathCache(key string) (interface{}, hotPathCacheState) {
	if s == nil || strings.TrimSpace(key) == "" {
		return nil, hotPathCacheMiss
	}

	s.hotPathCacheMu.Lock()
	defer s.hotPathCacheMu.Unlock()

	if s.hotPathCache == nil {
		return nil, hotPathCacheMiss
	}
	entry, ok := s.hotPathCache[key]
	if !ok {
		return nil, hotPathCacheMiss
	}

	now := time.Now()
	staleUntil := entry.staleUntil
	if staleUntil.IsZero() {
		staleUntil = entry.freshUntil
	}
	if !staleUntil.IsZero() && now.After(staleUntil) {
		delete(s.hotPathCache, key)
		return nil, hotPathCacheMiss
	}
	if !entry.freshUntil.IsZero() && now.After(entry.freshUntil) {
		return entry.value, hotPathCacheStale
	}

	return entry.value, hotPathCacheFresh
}

func (s *Store) setLocalHotPathCache(key string, value interface{}) {
	if s == nil || strings.TrimSpace(key) == "" || s.hotPathCacheTTL <= 0 {
		return
	}

	s.hotPathCacheMu.Lock()
	defer s.hotPathCacheMu.Unlock()

	if s.hotPathCache == nil {
		s.hotPathCache = make(map[string]hotPathCacheEntry)
	}
	now := time.Now()
	freshUntil := now.Add(s.hotPathCacheTTL)
	staleUntil := freshUntil
	if s.hotPathCacheStaleTTL > 0 {
		staleUntil = freshUntil.Add(s.hotPathCacheStaleTTL)
	}
	s.hotPathCache[key] = hotPathCacheEntry{
		value:      value,
		freshUntil: freshUntil,
		staleUntil: staleUntil,
	}
}

func (s *Store) deleteLocalHotPathCache(key string) {
	if s == nil || strings.TrimSpace(key) == "" {
		return
	}

	s.hotPathCacheMu.Lock()
	defer s.hotPathCacheMu.Unlock()

	if s.hotPathCache != nil {
		delete(s.hotPathCache, key)
	}
}

func (s *Store) refreshHotPathCacheAsync(key string, refresh func(context.Context) error) <-chan singleflight.Result {
	if s == nil || strings.TrimSpace(key) == "" || refresh == nil {
		return nil
	}

	return s.hotPathRefreshGroup.DoChan(key, func() (interface{}, error) {
		ctx, cancel := context.WithTimeout(context.Background(), s.hotPathRefreshTimeout())
		defer cancel()
		return nil, refresh(ctx)
	})
}

func (s *Store) hotPathRefreshTimeout() time.Duration {
	if s == nil {
		return 30 * time.Second
	}

	timeout := s.hotPathCacheTTL + s.hotPathCacheStaleTTL
	if timeout < 5*time.Second {
		return 5 * time.Second
	}
	if timeout > 30*time.Second {
		return 30 * time.Second
	}

	return timeout
}

func (s *Store) warmDeploymentCacheAsync(tenantID string) {
	cacheKey := deploymentCacheKey(tenantID)
	s.refreshDeploymentsCacheAsync(tenantID, cacheKey)
}

func (s *Store) refreshDeploymentsCacheAsync(tenantID string, cacheKey string) {
	if s == nil || s.pool == nil {
		return
	}

	s.refreshHotPathCacheAsync(cacheKey, func(ctx context.Context) error {
		deps, err := s.listDeploymentsFromDB(ctx, tenantID)
		if err != nil {
			return err
		}
		s.cacheDeployments(cacheKey, deps)
		return nil
	})
}

func (s *Store) warmVirtualModelsCacheAsync(tenantID string) {
	cacheKey := virtualModelsCacheKey(tenantID)
	s.refreshVirtualModelsCacheAsync(tenantID, cacheKey)
}

func (s *Store) refreshVirtualModelsCacheAsync(tenantID string, cacheKey string) {
	if s == nil || s.pool == nil {
		return
	}

	s.refreshHotPathCacheAsync(cacheKey, func(ctx context.Context) error {
		models, err := s.listVirtualModelsFromDB(ctx, tenantID)
		if err != nil {
			return err
		}
		s.cacheVirtualModels(cacheKey, models)
		return nil
	})
}

func cloneDeployments(in []platform.Deployment) []platform.Deployment {
	if len(in) == 0 {
		return nil
	}

	out := make([]platform.Deployment, len(in))
	for i, dep := range in {
		out[i] = dep
		out[i].ModelInfo = cloneStringInterfaceMap(dep.ModelInfo)
	}

	return out
}

func cloneVirtualModels(in []platform.VirtualModel) []platform.VirtualModel {
	if len(in) == 0 {
		return nil
	}

	out := make([]platform.VirtualModel, len(in))
	for i, vm := range in {
		out[i] = cloneVirtualModelValue(vm)
	}

	return out
}

func cloneVirtualModelValue(in platform.VirtualModel) platform.VirtualModel {
	out := in
	out.ExposedModes = append([]string(nil), in.ExposedModes...)
	out.Targets = cloneVirtualModelTargets(in.Targets)
	out.Access = platform.TenantProviderAccountAccessControl{
		Managers: append([]platform.AccessPrincipal(nil), in.Access.Managers...),
		Users:    append([]platform.AccessPrincipal(nil), in.Access.Users...),
		Rules:    append([]platform.TenantProviderAccountAccessRule(nil), in.Access.Rules...),
	}

	return out
}

func cloneVirtualModelTargets(in []platform.VirtualModelTarget) []platform.VirtualModelTarget {
	if len(in) == 0 {
		return nil
	}

	out := make([]platform.VirtualModelTarget, len(in))
	for i, target := range in {
		out[i] = target
		out[i].Policy.Retry.StatusCodes = append([]int(nil), target.Policy.Retry.StatusCodes...)
		out[i].Policy.Fallback.StatusCodes = append([]int(nil), target.Policy.Fallback.StatusCodes...)
	}

	return out
}
