package storage

import (
	"context"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

type hotPathCacheEntry struct {
	value     interface{}
	expiresAt time.Time
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
}

func (s *Store) getCachedDeployments(key string) ([]platform.Deployment, bool) {
	value, ok := s.getLocalHotPathCache(key)
	if !ok {
		return nil, false
	}

	deps, ok := value.([]platform.Deployment)
	if !ok {
		s.deleteLocalHotPathCache(key)
		return nil, false
	}

	return cloneDeployments(deps), true
}

func (s *Store) cacheDeployments(key string, deps []platform.Deployment) {
	s.setLocalHotPathCache(key, cloneDeployments(deps))
}

func (s *Store) getCachedVirtualModels(key string) ([]platform.VirtualModel, bool) {
	value, ok := s.getLocalHotPathCache(key)
	if !ok {
		return nil, false
	}

	models, ok := value.([]platform.VirtualModel)
	if !ok {
		s.deleteLocalHotPathCache(key)
		return nil, false
	}

	return cloneVirtualModels(models), true
}

func (s *Store) cacheVirtualModels(key string, models []platform.VirtualModel) {
	s.setLocalHotPathCache(key, cloneVirtualModels(models))
}

func (s *Store) getLocalHotPathCache(key string) (interface{}, bool) {
	if s == nil || strings.TrimSpace(key) == "" {
		return nil, false
	}

	s.hotPathCacheMu.Lock()
	defer s.hotPathCacheMu.Unlock()

	if s.hotPathCache == nil {
		return nil, false
	}
	entry, ok := s.hotPathCache[key]
	if !ok {
		return nil, false
	}
	if !entry.expiresAt.IsZero() && time.Now().After(entry.expiresAt) {
		delete(s.hotPathCache, key)
		return nil, false
	}

	return entry.value, true
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
	s.hotPathCache[key] = hotPathCacheEntry{
		value:     value,
		expiresAt: time.Now().Add(s.hotPathCacheTTL),
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
