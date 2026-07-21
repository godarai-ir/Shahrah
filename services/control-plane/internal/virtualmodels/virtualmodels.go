// Package virtualmodels resolves virtual model aliases and route targets.
package virtualmodels

import (
	"context"
	"errors"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/config"
	"github.com/Delpak/shahrah/services/control-plane/internal/modelid"
	"github.com/Delpak/shahrah/services/control-plane/internal/router"
	"golang.org/x/sync/singleflight"
)

var (
	// ErrNotVirtualModel indicates the requested alias is not virtual.
	ErrNotVirtualModel      = errors.New("not a virtual model")
	ErrDisabledVirtualModel = errors.New("virtual model disabled")
	ErrNoEligibleTargets    = errors.New("no eligible virtual model targets")
	ErrFallbackExhausted    = errors.New("virtual model fallback exhausted")
)

type VirtualModelRepository interface {
	GetVirtualModelByAlias(ctx context.Context, tenantID string, alias string) (*platform.VirtualModel, error)
}

type VirtualModelResolver interface {
	Resolve(ctx context.Context, tenantID, alias string) (*platform.VirtualModel, error)
}

type RoutingEngine interface {
	SelectTargets(vm *platform.VirtualModel) ([]platform.VirtualModelTarget, error)
}

type ErrorNormalizer interface {
	ShouldRetryOnStatus(status int, target *platform.VirtualModelTarget) bool
	ShouldFallbackOnStatus(status int, target *platform.VirtualModelTarget) bool
}

type FallbackExecutor interface {
	ExecuteSync(ctx context.Context, targets []platform.VirtualModelTarget, doCall func(ctx context.Context, dep *router.Deployment) (*platform.TunnelResponse, error)) (*platform.TunnelResponse, *router.Deployment, int, string, error)
}

type Resolver struct {
	repo             VirtualModelRepository
	cacheTTL         time.Duration
	negativeCacheTTL time.Duration
	isNotFoundError  func(error) bool
	mu               sync.Mutex
	cache            map[string]aliasCacheEntry
	loadGroup        singleflight.Group
}

type aliasCacheEntry struct {
	vm        *platform.VirtualModel
	notFound  bool
	expiresAt time.Time
}

type ResolverOption func(*Resolver)

type StrategyRoutingEngine struct {
	priority *PriorityRoutingEngine
	weighted *WeightedRoutingEngine
}

func NewResolver(repo VirtualModelRepository, opts ...ResolverOption) *Resolver {
	r := &Resolver{
		repo:             repo,
		cacheTTL:         secondsEnvOrDefault("VIRTUAL_MODEL_ALIAS_CACHE_TTL_SECONDS", 60*time.Second),
		negativeCacheTTL: secondsEnvOrDefault("VIRTUAL_MODEL_ALIAS_NEGATIVE_CACHE_TTL_SECONDS", 10*time.Second),
		cache:            make(map[string]aliasCacheEntry),
	}

	for _, opt := range opts {
		if opt != nil {
			opt(r)
		}
	}

	return r
}

func WithNotFoundError(fn func(error) bool) ResolverOption {
	return func(r *Resolver) {
		r.isNotFoundError = fn
	}
}

func WithAliasCacheTTLs(hitTTL, negativeTTL time.Duration) ResolverOption {
	return func(r *Resolver) {
		r.cacheTTL = hitTTL
		r.negativeCacheTTL = negativeTTL
	}
}

func NewRoutingEngine() *StrategyRoutingEngine {
	return &StrategyRoutingEngine{
		priority: NewPriorityRoutingEngine(),
		weighted: NewWeightedRoutingEngine(),
	}
}

func (r *Resolver) Resolve(ctx context.Context, tenantID, alias string) (*platform.VirtualModel, error) {
	tenantID = strings.TrimSpace(tenantID)
	alias = strings.TrimSpace(alias)

	key := virtualModelAliasCacheKey(tenantID, alias)
	if entry, ok := r.getCached(key); ok {
		if entry.notFound {
			return nil, ErrNotVirtualModel
		}

		return cloneVirtualModel(entry.vm), nil
	}

	value, err, _ := r.loadGroup.Do(key, func() (interface{}, error) {
		if entry, ok := r.getCached(key); ok {
			return entry, nil
		}

		vm, err := r.repo.GetVirtualModelByAlias(ctx, tenantID, alias)
		if err != nil {
			if r.shouldCacheNotFound(err) {
				entry := aliasCacheEntry{notFound: true}
				r.setCached(key, entry, r.negativeCacheTTL)

				return entry, nil
			}

			return nil, err
		}

		if vm != nil && vm.Enabled {
			entry := aliasCacheEntry{vm: cloneVirtualModel(vm)}
			r.setCached(key, entry, r.cacheTTL)

			return entry, nil
		}

		return cloneVirtualModel(vm), nil
	})
	if err != nil {
		return nil, err
	}

	if entry, ok := value.(aliasCacheEntry); ok {
		if entry.notFound {
			return nil, ErrNotVirtualModel
		}

		vm := cloneVirtualModel(entry.vm)
		if vm == nil {
			return nil, ErrNotVirtualModel
		}

		if !vm.Enabled {
			return nil, ErrDisabledVirtualModel
		}

		return vm, nil
	}

	vm, ok := value.(*platform.VirtualModel)
	if !ok || vm == nil {
		return nil, ErrNotVirtualModel
	}

	if !vm.Enabled {
		return nil, ErrDisabledVirtualModel
	}

	return vm, nil
}

func (r *Resolver) shouldCacheNotFound(err error) bool {
	if err == nil || r.isNotFoundError == nil {
		return false
	}

	return r.isNotFoundError(err)
}

func (r *Resolver) getCached(key string) (aliasCacheEntry, bool) {
	if r == nil || key == "" {
		return aliasCacheEntry{}, false
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	entry, ok := r.cache[key]
	if !ok {
		return aliasCacheEntry{}, false
	}

	if !entry.expiresAt.IsZero() && time.Now().After(entry.expiresAt) {
		delete(r.cache, key)
		return aliasCacheEntry{}, false
	}

	return entry, true
}

func (r *Resolver) setCached(key string, entry aliasCacheEntry, ttl time.Duration) {
	if r == nil || key == "" || ttl <= 0 {
		return
	}

	entry.expiresAt = time.Now().Add(ttl)
	if entry.vm != nil {
		entry.vm = cloneVirtualModel(entry.vm)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.cache[key] = entry
}

func virtualModelAliasCacheKey(tenantID, alias string) string {
	tenantID = strings.TrimSpace(tenantID)
	alias = strings.TrimSpace(alias)

	return tenantID + "\x00" + alias
}

func secondsEnvOrDefault(key string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}

	seconds, err := strconv.Atoi(raw)
	if err != nil || seconds <= 0 {
		return fallback
	}

	return time.Duration(seconds) * time.Second
}

func cloneVirtualModel(in *platform.VirtualModel) *platform.VirtualModel {
	if in == nil {
		return nil
	}

	out := *in
	out.ExposedModes = append([]string(nil), in.ExposedModes...)
	out.Targets = cloneVirtualModelTargets(in.Targets)
	out.Access = platform.TenantProviderAccountAccessControl{
		Managers: append([]platform.AccessPrincipal(nil), in.Access.Managers...),
		Users:    append([]platform.AccessPrincipal(nil), in.Access.Users...),
		Rules:    append([]platform.TenantProviderAccountAccessRule(nil), in.Access.Rules...),
	}

	return &out
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

func (e *StrategyRoutingEngine) SelectTargets(vm *platform.VirtualModel) ([]platform.VirtualModelTarget, error) {
	if vm == nil {
		return nil, ErrNoEligibleTargets
	}

	switch strings.ToLower(strings.TrimSpace(vm.RoutingStrategy)) {
	case "weighted":
		return e.weighted.SelectTargets(vm)
	case "", "priority":
		return e.priority.SelectTargets(vm)
	default:
		return e.priority.SelectTargets(vm)
	}
}

type PriorityRoutingEngine struct{}

type WeightedRoutingEngine struct{}

func NewPriorityRoutingEngine() *PriorityRoutingEngine {
	return &PriorityRoutingEngine{}
}

func NewWeightedRoutingEngine() *WeightedRoutingEngine {
	return &WeightedRoutingEngine{}
}

func (e *PriorityRoutingEngine) SelectTargets(vm *platform.VirtualModel) ([]platform.VirtualModelTarget, error) {
	out := normalizedEligibleTargets(vm)
	if len(out) == 0 {
		return nil, ErrNoEligibleTargets
	}

	return out, nil
}

func (e *WeightedRoutingEngine) SelectTargets(vm *platform.VirtualModel) ([]platform.VirtualModelTarget, error) {
	eligible := normalizedEligibleTargets(vm)
	if len(eligible) == 0 {
		return nil, ErrNoEligibleTargets
	}

	primaryIndex := weightedPickIndex(eligible)
	if primaryIndex < 0 {
		return nil, ErrNoEligibleTargets
	}

	primary := eligible[primaryIndex]
	out := make([]platform.VirtualModelTarget, 0, len(eligible))
	out = append(out, primary)

	for i, target := range eligible {
		if i == primaryIndex {
			continue
		}

		if !target.Policy.Fallback.Candidate {
			continue
		}

		out = append(out, target)
	}

	return out, nil
}

type DefaultErrorNormalizer struct{}

func NewDefaultErrorNormalizer() *DefaultErrorNormalizer {
	return &DefaultErrorNormalizer{}
}

func (n *DefaultErrorNormalizer) ShouldRetryOnStatus(status int, target *platform.VirtualModelTarget) bool {
	return containsInt(target.Policy.Retry.StatusCodes, status)
}

func (n *DefaultErrorNormalizer) ShouldFallbackOnStatus(status int, target *platform.VirtualModelTarget) bool {
	return containsInt(target.Policy.Fallback.StatusCodes, status)
}

type Executor struct {
	normalizer ErrorNormalizer
}

func NewExecutor(normalizer ErrorNormalizer) *Executor {
	return &Executor{normalizer: normalizer}
}

func (e *Executor) ExecuteSync(ctx context.Context, targets []platform.VirtualModelTarget, doCall func(ctx context.Context, dep *router.Deployment) (*platform.TunnelResponse, error)) (*platform.TunnelResponse, *router.Deployment, int, string, error) {
	fallbackCount := 0
	lastReason := ""

	for idx := range targets {
		target := &targets[idx]
		if idx > 0 {
			fallbackCount++
		}

		if !target.Policy.Fallback.Candidate && idx > 0 {
			continue
		}

		retries := target.Policy.Retry.Attempts
		if retries <= 0 {
			retries = 1
		}

		delay := time.Duration(target.Policy.Retry.DelayMS) * time.Millisecond

		for attempt := 1; attempt <= retries; attempt++ {
			dep := buildDeployment(target)

			resp, err := doCall(ctx, dep)
			if err != nil {
				lastReason = err.Error()

				if attempt < retries {
					sleepRetryDelay(delay)
					continue
				}

				break
			}

			if resp.StatusCode < 400 {
				return resp, dep, fallbackCount, lastReason, nil
			}

			if e.normalizer.ShouldRetryOnStatus(resp.StatusCode, target) && attempt < retries {
				sleepRetryDelay(delay)
				continue
			}

			lastReason = "status_error"
			if !e.normalizer.ShouldFallbackOnStatus(resp.StatusCode, target) {
				return resp, dep, fallbackCount, lastReason, nil
			}

			break
		}
	}

	return nil, nil, fallbackCount, lastReason, ErrFallbackExhausted
}

func buildDeployment(target *platform.VirtualModelTarget) *router.Deployment {
	provider := strings.TrimSpace(target.Provider)
	modelForProvider := strings.TrimSpace(target.Model)

	if p, _, official, ok := splitModelFQN(target.Model); ok {
		if provider == "" {
			provider = p
		}

		modelForProvider = official
	}

	return &router.Deployment{
		ID:        target.ID,
		ModelName: target.Model,
		Provider:  provider,
		Model:     modelForProvider,
		Params: config.ProviderEndpointConfig{
			Timeout: target.Policy.TimeoutSec,
		},
		Healthy: true,
		Weight:  target.Weight,
	}
}

func sleepRetryDelay(delay time.Duration) {
	if delay > 0 {
		time.Sleep(delay)
	}
}

func splitModelFQN(input string) (provider, accountAlias, officialModel string, ok bool) {
	return modelid.SplitModelFQN(input)
}

func containsInt(values []int, target int) bool {
	for _, v := range values {
		if v == target {
			return true
		}
	}

	return false
}

func normalizedEligibleTargets(vm *platform.VirtualModel) []platform.VirtualModelTarget {
	if vm == nil || len(vm.Targets) == 0 {
		return nil
	}

	out := make([]platform.VirtualModelTarget, 0, len(vm.Targets))
	for i := range vm.Targets {
		t := &vm.Targets[i]
		if strings.TrimSpace(t.Provider) == "" || strings.TrimSpace(t.Model) == "" {
			continue
		}

		out = append(out, *t)
	}

	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Order == out[j].Order {
			return out[i].Model < out[j].Model
		}

		return out[i].Order < out[j].Order
	})

	return out
}

func weightedPickIndex(targets []platform.VirtualModelTarget) int {
	totalWeight := 0.0

	for _, target := range targets {
		if target.Weight > 0 {
			totalWeight += target.Weight
		}
	}

	if totalWeight <= 0 {
		return -1
	}

	pick := rand.Float64() * totalWeight
	cumulative := 0.0

	for i, target := range targets {
		if target.Weight <= 0 {
			continue
		}

		cumulative += target.Weight
		if pick <= cumulative {
			return i
		}
	}

	return len(targets) - 1
}
