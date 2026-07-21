package storage

import (
	"context"
	"strings"
)

func (s *Store) invalidateTenantRuntimeProfile(ctx context.Context, tenantID string, reason string) {
	if s == nil || s.runtimeProfiles == nil {
		return
	}

	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return
	}

	s.runtimeProfiles.Invalidate(ctx, tenantID, reason)
}
