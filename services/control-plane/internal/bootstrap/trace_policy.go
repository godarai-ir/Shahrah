package bootstrap

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/observability/tracing"
	"github.com/Delpak/shahrah/services/control-plane/internal/storage"
)

type tenantTracePolicyResolver struct {
	store *storage.ObservabilityRepository
	mu    sync.RWMutex
	cache map[string]tenantPolicyCache
	ttl   time.Duration
}

type tenantPolicyCache struct {
	policy    tracing.CapturePolicy
	expiresAt time.Time
}

func newTenantTracePolicyResolver(store *storage.ObservabilityRepository) *tenantTracePolicyResolver {
	return &tenantTracePolicyResolver{
		store: store,
		cache: map[string]tenantPolicyCache{},
		ttl:   60 * time.Second,
	}
}

func (r *tenantTracePolicyResolver) ResolveTenantPolicy(tenantID string) (tracing.CapturePolicy, bool) {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return "", false
	}

	now := time.Now()

	r.mu.RLock()
	cached, ok := r.cache[tenantID]
	r.mu.RUnlock()

	if ok && now.Before(cached.expiresAt) {
		return cached.policy, true
	}

	ctx, cancel := context.WithTimeout(context.Background(), 400*time.Millisecond)
	defer cancel()

	policy, exists, err := r.store.ResolveTenantTraceCapturePolicy(ctx, tenantID)
	if err != nil || !exists {
		return "", false
	}

	out := tracing.ParseCapturePolicy(policy)

	r.mu.Lock()
	r.cache[tenantID] = tenantPolicyCache{policy: out, expiresAt: now.Add(r.ttl)}
	r.mu.Unlock()

	return out, true
}
