package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/dgraph-io/ristretto/v2"
	"go.uber.org/zap"
)

// MemoryCache implements Cache using Ristretto v2 with generics.
type MemoryCache struct {
	cache              *ristretto.Cache[string, any]
	logger             *zap.Logger
	mu                 sync.Mutex
	fixedWindows       map[string]fixedWindowCounter
	slidingWindows     map[string][]slidingWindowEntry
	budgetWindows      map[string]memoryBudgetWindow
	budgetReservations map[string]memoryBudgetReservation
}

type fixedWindowCounter struct {
	count     int64
	expiresAt time.Time
}

type slidingWindowEntry struct {
	tsMS   int64
	weight int64
}

type memoryBudgetWindow struct {
	allowanceIRR float64
	reservedIRR  float64
	committedIRR float64
	overageIRR   float64
	expiresAt    time.Time
	expiresCache time.Time
	status       string
	version      int64
}

type memoryBudgetReservation struct {
	windowKey    string
	status       string
	estimatedIRR float64
	actualIRR    float64
	expiresAt    time.Time
}

// NewMemoryCache creates a new in-memory cache using Ristretto v2.
func NewMemoryCache(logger *zap.Logger) (*MemoryCache, error) {
	cache, err := ristretto.NewCache(&ristretto.Config[string, any]{
		NumCounters: 1e7,     // number of keys to track frequency of (10M)
		MaxCost:     1 << 30, // maximum cost of cache (1GB)
		BufferItems: 64,      // number of keys per Get buffer
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create ristretto v2 cache: %w", err)
	}

	return &MemoryCache{
		cache:              cache,
		logger:             logger,
		fixedWindows:       make(map[string]fixedWindowCounter),
		slidingWindows:     make(map[string][]slidingWindowEntry),
		budgetWindows:      make(map[string]memoryBudgetWindow),
		budgetReservations: make(map[string]memoryBudgetReservation),
	}, nil
}

// Get retrieves a value from the cache with proper type handling.
func (m *MemoryCache) Get(_ context.Context, key string) (interface{}, bool) {
	val, found := m.cache.Get(key)
	if !found {
		return nil, false
	}

	return val, true
}

// Set stores a value in the cache with TTL support.
func (m *MemoryCache) Set(_ context.Context, key string, value interface{}, ttl time.Duration) {
	// Ristretto v2 uses SetWithTTL with cost parameter
	m.cache.SetWithTTL(key, value, 1, ttl)
	m.cache.Wait()
}

// Delete removes a value from the cache.
func (m *MemoryCache) Delete(_ context.Context, key string) {
	m.cache.Del(key)
}

// Increment safely increments an int64 value in the cache.
func (m *MemoryCache) Increment(_ context.Context, key string, value int64, ttl time.Duration) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	val, found := m.cache.Get(key)
	if !found {
		m.cache.SetWithTTL(key, value, 1, ttl)
		m.cache.Wait()

		return value, nil
	}

	// Type-safe access with proper error handling
	intVal, ok := val.(int64)
	if !ok {
		return 0, fmt.Errorf("expected int64 value, got %T", val)
	}

	newVal := intVal + value
	m.cache.SetWithTTL(key, newVal, 1, ttl)
	m.cache.Wait()

	return newVal, nil
}

// CheckRateLimit checks a fixed-window counter in memory.
func (m *MemoryCache) CheckRateLimit(_ context.Context, key string, limit int64, window time.Duration) (*RateLimitResult, error) {
	if limit <= 0 {
		return &RateLimitResult{
			Allowed: true,
			Current: 0,
			Limit:   limit,
			ResetAt: time.Now().Add(window),
		}, nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()

	counter, exists := m.fixedWindows[key]
	if !exists || now.After(counter.expiresAt) {
		counter = fixedWindowCounter{
			count:     0,
			expiresAt: now.Add(window),
		}
	}

	if counter.count >= limit {
		return &RateLimitResult{
			Allowed: false,
			Current: counter.count,
			Limit:   limit,
			ResetAt: counter.expiresAt,
		}, nil
	}

	counter.count++
	m.fixedWindows[key] = counter

	return &RateLimitResult{
		Allowed: true,
		Current: counter.count,
		Limit:   limit,
		ResetAt: counter.expiresAt,
	}, nil
}

// CheckSlidingWindowRateLimit checks a true sliding-window log in memory.
func (m *MemoryCache) CheckSlidingWindowRateLimit(ctx context.Context, key string, limit int64, window time.Duration) (*RateLimitResult, error) {
	return m.CheckSlidingWindowRateLimitWithWeight(ctx, key, limit, 1, window)
}

// CheckSlidingWindowRateLimitWithWeight checks a true sliding-window log in memory with weighted entries.
func (m *MemoryCache) CheckSlidingWindowRateLimitWithWeight(_ context.Context, key string, limit, weight int64, window time.Duration) (*RateLimitResult, error) {
	if limit <= 0 {
		return &RateLimitResult{
			Allowed: true,
			Current: 0,
			Limit:   limit,
			ResetAt: time.Now().Add(window),
		}, nil
	}

	if weight < 1 {
		weight = 1
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	nowMS := now.UnixMilli()
	cutoffMS := now.Add(-window).UnixMilli()
	entries := m.slidingWindows[key]

	pruned := make([]slidingWindowEntry, 0, len(entries))

	var total int64

	for _, ts := range entries {
		if ts.tsMS > cutoffMS {
			pruned = append(pruned, ts)
			total += ts.weight
		}
	}

	if total+weight > limit {
		resetAt := now.Add(window)
		if len(pruned) > 0 {
			resetAt = time.UnixMilli(pruned[0].tsMS).Add(window)
		}

		m.slidingWindows[key] = pruned

		return &RateLimitResult{
			Allowed: false,
			Current: total,
			Limit:   limit,
			ResetAt: resetAt,
		}, nil
	}

	pruned = append(pruned, slidingWindowEntry{tsMS: nowMS, weight: weight})
	m.slidingWindows[key] = pruned

	return &RateLimitResult{
		Allowed: true,
		Current: total + weight,
		Limit:   limit,
		ResetAt: now.Add(window),
	}, nil
}

// SeedBudgetWindow creates or refreshes a tenant spend allowance window.
func (m *MemoryCache) SeedBudgetWindow(_ context.Context, seed BudgetWindowSeed) (*BudgetWindowResult, error) {
	if seed.WindowKey == "" {
		return nil, fmt.Errorf("budget window key is required")
	}
	if seed.TTL <= 0 {
		return nil, fmt.Errorf("budget window ttl must be positive")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now().UTC()
	m.pruneBudgetStateLocked(now)
	window := m.budgetWindows[seed.WindowKey]
	if window.status == "" {
		window.status = "active"
	}
	window.allowanceIRR = seed.AllowanceIRR
	window.expiresAt = seed.ExpiresAt
	window.expiresCache = now.Add(seed.TTL)
	window.version = seed.Version
	m.budgetWindows[seed.WindowKey] = window
	return memoryBudgetResult(window, true, window.status), nil
}

// ReserveBudget atomically reserves estimated spend from a tenant allowance window.
func (m *MemoryCache) ReserveBudget(_ context.Context, req BudgetReservationRequest) (*BudgetWindowResult, error) {
	if req.WindowKey == "" || req.ReservationKey == "" {
		return nil, fmt.Errorf("budget window and reservation keys are required")
	}
	if req.TTL <= 0 {
		return nil, fmt.Errorf("budget reservation ttl must be positive")
	}
	now := req.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.pruneBudgetStateLocked(now)
	window, ok := m.budgetWindows[req.WindowKey]
	if !ok || window.status != "active" || (!window.expiresAt.IsZero() && now.After(window.expiresAt)) {
		return &BudgetWindowResult{Allowed: false, State: "window_unavailable"}, nil
	}
	if reservation, ok := m.budgetReservations[req.ReservationKey]; ok {
		return memoryBudgetResult(window, true, reservation.status), nil
	}

	estimated := req.EstimatedIRR
	if estimated < 0 {
		estimated = 0
	}
	if window.reservedIRR+window.committedIRR+estimated > window.allowanceIRR {
		return memoryBudgetResult(window, false, "exhausted"), nil
	}

	window.reservedIRR += estimated
	m.budgetWindows[req.WindowKey] = window
	m.budgetReservations[req.ReservationKey] = memoryBudgetReservation{
		windowKey:    req.WindowKey,
		status:       "reserved",
		estimatedIRR: estimated,
		expiresAt:    now.Add(req.TTL),
	}
	return memoryBudgetResult(window, true, "reserved"), nil
}

// CommitBudget atomically converts a reservation into committed spend.
func (m *MemoryCache) CommitBudget(_ context.Context, req BudgetCommitRequest) (*BudgetWindowResult, error) {
	if req.WindowKey == "" || req.ReservationKey == "" {
		return nil, fmt.Errorf("budget window and reservation keys are required")
	}
	now := req.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.pruneBudgetStateLocked(now)
	window, ok := m.budgetWindows[req.WindowKey]
	if !ok {
		return &BudgetWindowResult{Allowed: false, State: "window_unavailable"}, nil
	}
	reservation, ok := m.budgetReservations[req.ReservationKey]
	if !ok {
		return &BudgetWindowResult{Allowed: false, State: "reservation_unavailable"}, nil
	}
	if reservation.status == "committed" || reservation.status == "released" {
		return memoryBudgetResult(window, true, reservation.status), nil
	}

	actual := req.ActualIRR
	if actual < 0 {
		actual = 0
	}
	if window.reservedIRR-reservation.estimatedIRR+window.committedIRR+actual > window.allowanceIRR {
		return memoryBudgetResult(window, false, "exhausted"), nil
	}
	window.reservedIRR -= reservation.estimatedIRR
	if window.reservedIRR < 0 {
		window.reservedIRR = 0
	}
	window.committedIRR += actual
	reservation.status = "committed"
	reservation.actualIRR = actual
	m.budgetWindows[req.WindowKey] = window
	m.budgetReservations[req.ReservationKey] = reservation
	return memoryBudgetResult(window, true, "committed"), nil
}

// ReleaseBudget atomically releases an uncommitted reservation.
func (m *MemoryCache) ReleaseBudget(_ context.Context, req BudgetReleaseRequest) (*BudgetWindowResult, error) {
	if req.WindowKey == "" || req.ReservationKey == "" {
		return nil, fmt.Errorf("budget window and reservation keys are required")
	}
	now := req.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.pruneBudgetStateLocked(now)
	window, ok := m.budgetWindows[req.WindowKey]
	if !ok {
		return &BudgetWindowResult{Allowed: false, State: "window_unavailable"}, nil
	}
	reservation, ok := m.budgetReservations[req.ReservationKey]
	if !ok {
		return memoryBudgetResult(window, true, "missing"), nil
	}
	if reservation.status == "reserved" {
		window.reservedIRR -= reservation.estimatedIRR
		if window.reservedIRR < 0 {
			window.reservedIRR = 0
		}
		reservation.status = "released"
		m.budgetWindows[req.WindowKey] = window
		m.budgetReservations[req.ReservationKey] = reservation
	}
	return memoryBudgetResult(window, true, "released"), nil
}

func (m *MemoryCache) pruneBudgetStateLocked(now time.Time) {
	for key, window := range m.budgetWindows {
		if !window.expiresCache.IsZero() && now.After(window.expiresCache) {
			delete(m.budgetWindows, key)
		}
	}
	for key, reservation := range m.budgetReservations {
		if !reservation.expiresAt.IsZero() && now.After(reservation.expiresAt) {
			delete(m.budgetReservations, key)
		}
	}
}

func memoryBudgetResult(window memoryBudgetWindow, allowed bool, state string) *BudgetWindowResult {
	return &BudgetWindowResult{
		Allowed:      allowed,
		State:        state,
		AllowanceIRR: window.allowanceIRR,
		ReservedIRR:  window.reservedIRR,
		CommittedIRR: window.committedIRR,
		RemainingIRR: window.allowanceIRR - window.reservedIRR - window.committedIRR,
		OverageIRR:   window.overageIRR,
	}
}

// GetBudgetWindow returns the current state of one in-memory budget window.
func (m *MemoryCache) GetBudgetWindow(_ context.Context, windowKey string) (*BudgetWindowSnapshot, error) {
	if windowKey == "" {
		return nil, fmt.Errorf("budget window key is required")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now().UTC()
	m.pruneBudgetStateLocked(now)
	window, ok := m.budgetWindows[windowKey]
	if !ok {
		return nil, nil
	}

	return &BudgetWindowSnapshot{
		WindowKey:    windowKey,
		Status:       window.status,
		AllowanceIRR: window.allowanceIRR,
		ReservedIRR:  window.reservedIRR,
		CommittedIRR: window.committedIRR,
		RemainingIRR: window.allowanceIRR - window.reservedIRR - window.committedIRR,
		OverageIRR:   window.overageIRR,
		ExpiresAt:    window.expiresAt,
		Version:      window.version,
	}, nil
}

// ListActiveBudgetWindows returns a bounded list of active in-memory budget windows.
func (m *MemoryCache) ListActiveBudgetWindows(_ context.Context, batchSize int) ([]BudgetWindowSnapshot, error) {
	if batchSize <= 0 {
		batchSize = 100
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now().UTC()
	m.pruneBudgetStateLocked(now)

	out := make([]BudgetWindowSnapshot, 0, min(batchSize, len(m.budgetWindows)))
	for key, window := range m.budgetWindows {
		out = append(out, BudgetWindowSnapshot{
			WindowKey:    key,
			Status:       window.status,
			AllowanceIRR: window.allowanceIRR,
			ReservedIRR:  window.reservedIRR,
			CommittedIRR: window.committedIRR,
			RemainingIRR: window.allowanceIRR - window.reservedIRR - window.committedIRR,
			OverageIRR:   window.overageIRR,
			ExpiresAt:    window.expiresAt,
			Version:      window.version,
		})
		if len(out) >= batchSize {
			break
		}
	}

	return out, nil
}

func min(left int, right int) int {
	if left < right {
		return left
	}
	return right
}

// Close closes the cache and releases resources.
func (m *MemoryCache) Close() error {
	m.cache.Close()
	return nil
}
