package storage

import (
	"context"
	"strings"
)

func deploymentCacheKey(tenantID string) string {
	return "deployments:tenant:" + strings.TrimSpace(tenantID)
}

func walletCacheKey(tenantID string) string {
	return "wallet:tenant:" + strings.TrimSpace(tenantID)
}

func (s *Store) invalidateDeploymentCache(ctx context.Context, tenantID string) {
	if s == nil || s.entitlementCache == nil {
		return
	}
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return
	}
	s.entitlementCache.Delete(ctx, deploymentCacheKey(tenantID))
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
