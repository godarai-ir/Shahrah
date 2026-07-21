package storage

import (
	"context"
	"testing"
	"time"
)

func TestResolveTenantRequestLoggingModeUsesCache(t *testing.T) {
	ctx := context.Background()
	store := &Store{
		entitlementCache:      newEntitlementTestCache(),
		requestLoggingModeTTL: time.Minute,
	}
	store.cacheTenantRequestLoggingMode(ctx, "tenant-1", RequestLoggingModeFullContent, true)

	mode, ok, err := store.ResolveTenantRequestLoggingMode(ctx, "tenant-1")
	if err != nil {
		t.Fatal(err)
	}

	if !ok {
		t.Fatal("expected cached request logging mode")
	}

	if mode != RequestLoggingModeFullContent {
		t.Fatalf("unexpected mode: %s", mode)
	}
}

func TestInvalidateTenantRequestLoggingModeDeletesCache(t *testing.T) {
	ctx := context.Background()
	store := &Store{
		entitlementCache:      newEntitlementTestCache(),
		requestLoggingModeTTL: time.Minute,
	}
	store.cacheTenantRequestLoggingMode(ctx, "tenant-1", RequestLoggingModeFullContent, true)
	store.invalidateTenantRequestLoggingMode(ctx, "tenant-1")

	_, _, found := store.getCachedTenantRequestLoggingMode(ctx, "tenant-1")
	if found {
		t.Fatal("expected request logging mode cache to be invalidated")
	}
}
