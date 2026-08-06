package storage

import (
	"context"
	"strings"
)

func (s *Store) invalidateTenantRuntimeProfile(ctx context.Context, tenantID string, reason string) {
	if s == nil || s.runtimeProfileInvalidator == nil {
		return
	}

	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return
	}

	s.runtimeProfileInvalidator.Invalidate(ctx, tenantID, reason)
}
