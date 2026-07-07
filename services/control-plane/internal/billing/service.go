package billing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
	"github.com/Delpak/shahrah/services/control-plane/internal/runtimeprofile"
)

var (
	// ErrRateLimitExceeded is returned when per-key RPM/TPM is exhausted.
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
	// ErrBudgetExceeded is returned when max budget or wallet checks fail.
	ErrBudgetExceeded = errors.New("budget exceeded")
	// ErrBudgetUnavailable is returned when the spend allowance window cannot be read or updated.
	ErrBudgetUnavailable = errors.New("budget unavailable")
	// ErrSubscriptionInactive is returned when a tenant has no usable commercial subscription.
	ErrSubscriptionInactive = errors.New("subscription inactive")
)

type walletStore interface {
	GetWallet(ctx context.Context, tenantID string) (*platform.Wallet, error)
	GetLatestExchangeRate(ctx context.Context) (*platform.ExchangeRate, error)
	UpsertExchangeRate(ctx context.Context, rate platform.ExchangeRate) (*platform.ExchangeRate, error)
}

type entitlementStore interface {
	ResolveEffectiveEntitlementLimits(ctx context.Context, tenantID string) (*platform.EffectiveEntitlementLimits, error)
	CheckAndIncrementTenantUsageCounter(ctx context.Context, tenantID string, featureKey string, limit float64, increment float64, now time.Time) (*platform.UsageCounter, bool, error)
}

type budgetWindowCache interface {
	SeedBudgetWindow(ctx context.Context, seed cache.BudgetWindowSeed) (*cache.BudgetWindowResult, error)
	ReserveBudget(ctx context.Context, req cache.BudgetReservationRequest) (*cache.BudgetWindowResult, error)
	CommitBudget(ctx context.Context, req cache.BudgetCommitRequest) (*cache.BudgetWindowResult, error)
	ReleaseBudget(ctx context.Context, req cache.BudgetReleaseRequest) (*cache.BudgetWindowResult, error)
}

// Service enforces rate limits and budget checks.
type Service struct {
	cache           cache.Cache
	store           walletStore
	entitlements    entitlementStore
	runtimeProfiles interface {
		Get(ctx context.Context, tenantID string) (*runtimeprofile.Profile, error)
	}
	defaultRate          platform.ExchangeRate
	exchangeRateCache    *exchangeRateCache
	pricingCache         *pricingCache
	walletCacheTTL       time.Duration
	budgetWindowTTLGrace time.Duration
	rateLimitFailOpen    bool
}

// NewService creates a billing enforcement service.
func NewService(c cache.Cache, store walletStore, defaultUSDtoIRR float64, source string) *Service {
	return NewServiceWithCacheConfig(c, store, defaultUSDtoIRR, source, loadExchangeRateCacheConfigFromEnv(), loadPricingCacheConfigFromEnv())
}

// NewServiceWithExchangeRateCacheConfig creates a billing service with explicit cache settings.
func NewServiceWithExchangeRateCacheConfig(c cache.Cache, store walletStore, defaultUSDtoIRR float64, source string, exchangeRateCacheConfig ExchangeRateCacheConfig) *Service {
	return NewServiceWithCacheConfig(c, store, defaultUSDtoIRR, source, exchangeRateCacheConfig, loadPricingCacheConfigFromEnv())
}

// NewServiceWithCacheConfig creates a billing service with explicit hot-path cache settings.
func NewServiceWithCacheConfig(c cache.Cache, store walletStore, defaultUSDtoIRR float64, source string, exchangeRateCacheConfig ExchangeRateCacheConfig, pricingCacheConfig PricingCacheConfig) *Service {
	var entitlements entitlementStore
	if resolvedEntitlements, ok := store.(entitlementStore); ok {
		entitlements = resolvedEntitlements
	}

	defaultRate := platform.ExchangeRate{
		USDtoIRR: defaultUSDtoIRR,
		Source:   source,
	}

	svc := &Service{
		cache:                c,
		store:                store,
		entitlements:         entitlements,
		defaultRate:          defaultRate,
		walletCacheTTL:       secondsEnvOrDefault("BILLING_WALLET_CACHE_TTL_SECONDS", 5*time.Second),
		budgetWindowTTLGrace: secondsEnvOrDefault("BILLING_BUDGET_WINDOW_GRACE_SECONDS", 86400*time.Second),
		rateLimitFailOpen:    boolEnvOrDefault("RATE_LIMIT_POLICY_REDIS_FAIL_OPEN", false),
	}
	svc.exchangeRateCache = newExchangeRateCache(store, c, defaultRate, exchangeRateCacheConfig)
	svc.pricingCache = newPricingCache(c, pricingCacheConfig, modelPriceFromRegistry)

	return svc
}

func (s *Service) SetRuntimeProfiles(runtimeProfiles interface {
	Get(ctx context.Context, tenantID string) (*runtimeprofile.Profile, error)
}) {
	s.runtimeProfiles = runtimeProfiles
}

// CheckRateLimit validates per-minute request and token limits.
// Returns the primary rate-limit result to expose in response headers.
func (s *Service) CheckRateLimit(ctx context.Context, auth *platform.AuthContext, model string, estimatedTokens int) (*cache.RateLimitResult, error) {
	policyResult, err := s.checkRateLimitPolicies(ctx, auth, platform.UsageCostBreakdown{
		Model:       model,
		TotalTokens: estimatedTokens,
	})
	if err != nil {
		return policyResult, err
	}
	if policyResult != nil && !policyResult.Allowed {
		return policyResult, ErrRateLimitExceeded
	}

	if s.entitlements == nil || auth == nil || auth.TenantID == "" {
		return policyResult, nil
	}

	effective, err := s.resolveEffectiveEntitlementLimits(ctx, auth.TenantID)
	if err != nil {
		return nil, ErrSubscriptionInactive
	}

	if effective.Subscription == nil || (effective.Subscription.Status != "active" && effective.Subscription.Status != "trialing") {
		return nil, ErrSubscriptionInactive
	}

	limit, limited := entitlementNumber(effective.Features[platform.FeatureGatewayRequestsMonthly])
	if !limited {
		return policyResult, nil
	}

	counter, allowed, err := s.entitlements.CheckAndIncrementTenantUsageCounter(ctx, auth.TenantID, platform.FeatureGatewayRequestsMonthly, limit, 1, time.Now().UTC())
	if err != nil {
		return nil, err
	}

	result := &cache.RateLimitResult{
		Allowed: allowed,
		Current: int64(counter.Used),
		Limit:   int64(limit),
		ResetAt: counter.WindowEnd,
	}
	if !allowed {
		return result, ErrRateLimitExceeded
	}

	return mostRestrictiveRateLimitResult(policyResult, result), nil
}

func (s *Service) resolveEffectiveEntitlementLimits(ctx context.Context, tenantID string) (*platform.EffectiveEntitlementLimits, error) {
	if s.runtimeProfiles != nil {
		profile, err := s.runtimeProfiles.Get(ctx, tenantID)
		if err == nil && profile != nil && profile.Entitlements != nil {
			out := *profile.Entitlements
			out.Features = cloneInterfaceMap(profile.Entitlements.Features)
			if profile.Entitlements.Subscription != nil {
				sub := *profile.Entitlements.Subscription
				sub.PriceSnapshot = cloneInterfaceMap(sub.PriceSnapshot)
				sub.Metadata = cloneInterfaceMap(sub.Metadata)
				out.Subscription = &sub
			}
			return &out, nil
		}
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("tenant runtime profile entitlements unavailable")
	}
	return s.entitlements.ResolveEffectiveEntitlementLimits(ctx, tenantID)
}

// CheckBudget validates wallet balance constraints for compatibility callers.
func (s *Service) CheckBudget(ctx context.Context, auth *platform.AuthContext, breakdown platform.UsageCostBreakdown) error {
	if auth == nil {
		return ErrBudgetExceeded
	}

	reservationID := fmt.Sprintf("compat:%s:%d", strings.TrimSpace(auth.TokenHash), time.Now().UnixNano())
	reservation, err := s.ReserveBudget(ctx, auth, breakdown, reservationID)
	if err != nil {
		return err
	}
	if reservation != nil {
		_ = s.ReleaseBudget(ctx, reservation, "compat_check_only")
	}

	return nil
}

// ReserveBudget reserves the estimated request cost from the tenant credit window and matching budget policy windows.
func (s *Service) ReserveBudget(ctx context.Context, auth *platform.AuthContext, breakdown platform.UsageCostBreakdown, requestID string) (*platform.BudgetReservation, error) {
	if auth == nil || strings.TrimSpace(auth.TenantID) == "" {
		return nil, ErrBudgetExceeded
	}
	budgetCache, ok := s.cache.(budgetWindowCache)
	if !ok || budgetCache == nil {
		return nil, ErrBudgetUnavailable
	}

	tenantID := strings.TrimSpace(auth.TenantID)
	now := time.Now().UTC()
	reservationID := budgetReservationID(requestID)
	targets, err := s.budgetReservationTargets(ctx, auth, breakdown, reservationID, now)
	if err != nil {
		return nil, err
	}

	reserved := make([]budgetReservationTarget, 0, len(targets))
	for _, target := range targets {
		result, reserveErr := s.reserveBudgetTarget(ctx, budgetCache, target, reservationID, now)
		if reserveErr != nil {
			s.releaseBudgetTargetsBestEffort(ctx, budgetCache, reserved, now)
			return nil, reserveErr
		}
		if result == nil {
			s.releaseBudgetTargetsBestEffort(ctx, budgetCache, reserved, now)
			return nil, ErrBudgetUnavailable
		}
		if !result.Allowed {
			s.releaseBudgetTargetsBestEffort(ctx, budgetCache, reserved, now)
			if result.State == "exhausted" {
				return nil, ErrBudgetExceeded
			}
			return nil, ErrBudgetUnavailable
		}
		reserved = append(reserved, target)
	}
	if len(reserved) == 0 {
		return nil, ErrBudgetUnavailable
	}
	publicTargets := make([]platform.BudgetReservationTarget, 0, len(reserved))
	for _, target := range reserved {
		publicTargets = append(publicTargets, target.BudgetReservationTarget)
	}
	primary := publicTargets[0]

	return &platform.BudgetReservation{
		ID:             reservationID,
		TenantID:       tenantID,
		WindowID:       primary.WindowID,
		WindowKey:      primary.WindowKey,
		ReservationKey: primary.ReservationKey,
		EstimatedIRR:   breakdown.TotalCostIRR,
		CreatedAt:      now,
		Targets:        publicTargets,
	}, nil
}

// CommitBudget finalizes a reservation using the actual provider usage cost.
func (s *Service) CommitBudget(ctx context.Context, reservation *platform.BudgetReservation, breakdown platform.UsageCostBreakdown) error {
	if reservation == nil {
		return nil
	}
	budgetCache, ok := s.cache.(budgetWindowCache)
	if !ok || budgetCache == nil {
		return ErrBudgetUnavailable
	}
	targets := reservation.Targets
	if len(targets) == 0 {
		targets = []platform.BudgetReservationTarget{{
			WindowKey:      reservation.WindowKey,
			ReservationKey: reservation.ReservationKey,
		}}
	}
	now := time.Now().UTC()
	for _, target := range targets {
		result, err := budgetCache.CommitBudget(ctx, cache.BudgetCommitRequest{
			WindowKey:      target.WindowKey,
			ReservationKey: target.ReservationKey,
			ActualIRR:      breakdown.TotalCostIRR,
			Now:            now,
		})
		if err != nil {
			return fmt.Errorf("%w: commit failed: %v", ErrBudgetUnavailable, err)
		}
		if result == nil || !result.Allowed {
			if result != nil && result.State == "exhausted" {
				return ErrBudgetExceeded
			}
			return ErrBudgetUnavailable
		}
	}
	return nil
}

// ReleaseBudget releases a reservation that did not become chargeable usage.
func (s *Service) ReleaseBudget(ctx context.Context, reservation *platform.BudgetReservation, _ string) error {
	if reservation == nil {
		return nil
	}
	budgetCache, ok := s.cache.(budgetWindowCache)
	if !ok || budgetCache == nil {
		return ErrBudgetUnavailable
	}
	targets := reservation.Targets
	if len(targets) == 0 {
		targets = []platform.BudgetReservationTarget{{
			WindowKey:      reservation.WindowKey,
			ReservationKey: reservation.ReservationKey,
		}}
	}
	now := time.Now().UTC()
	for _, target := range targets {
		result, err := budgetCache.ReleaseBudget(ctx, cache.BudgetReleaseRequest{
			WindowKey:      target.WindowKey,
			ReservationKey: target.ReservationKey,
			Now:            now,
		})
		if err != nil {
			return fmt.Errorf("%w: release failed: %v", ErrBudgetUnavailable, err)
		}
		if result == nil || !result.Allowed {
			return ErrBudgetUnavailable
		}
	}
	return nil
}

func (s *Service) seedBudgetWindow(ctx context.Context, budgetCache budgetWindowCache, tenantID string, windowID string, windowKey string, windowEnd time.Time, ttl time.Duration) error {
	wallet, err := s.getWallet(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("%w: seed failed: %v", ErrBudgetUnavailable, err)
	}
	result, err := budgetCache.SeedBudgetWindow(ctx, cache.BudgetWindowSeed{
		WindowKey:    windowKey,
		AllowanceIRR: wallet.BalanceIRR,
		ExpiresAt:    windowEnd,
		TTL:          ttl,
		Version:      windowEnd.Unix(),
	})
	if err != nil {
		return fmt.Errorf("%w: seed failed: %v", ErrBudgetUnavailable, err)
	}
	if result == nil || !result.Allowed {
		return ErrBudgetUnavailable
	}
	return nil
}

func (s *Service) getWallet(ctx context.Context, tenantID string) (*platform.Wallet, error) {
	if s == nil || s.store == nil {
		return nil, errors.New("wallet store unavailable")
	}

	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" || s.cache == nil || s.walletCacheTTL <= 0 {
		return s.store.GetWallet(ctx, tenantID)
	}

	key := walletCacheKey(tenantID)
	if wallet, ok := s.cachedWallet(ctx, key); ok {
		return wallet, nil
	}

	wallet, err := s.store.GetWallet(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	s.storeWallet(ctx, key, wallet)
	return cloneWallet(wallet), nil
}

func (s *Service) cachedWallet(ctx context.Context, key string) (*platform.Wallet, bool) {
	value, ok := s.cache.Get(ctx, key)
	if !ok {
		return nil, false
	}

	switch typed := value.(type) {
	case string:
		return walletFromJSON([]byte(typed))
	case []byte:
		return walletFromJSON(typed)
	case platform.Wallet:
		return cloneWallet(&typed), true
	case *platform.Wallet:
		return cloneWallet(typed), typed != nil
	default:
		s.cache.Delete(ctx, key)
		return nil, false
	}
}

func (s *Service) storeWallet(ctx context.Context, key string, wallet *platform.Wallet) {
	if wallet == nil {
		return
	}

	raw, err := json.Marshal(wallet)
	if err != nil {
		return
	}

	s.cache.Set(ctx, key, string(raw), s.walletCacheTTL)
}

func walletFromJSON(raw []byte) (*platform.Wallet, bool) {
	var wallet platform.Wallet
	if err := json.Unmarshal(raw, &wallet); err != nil {
		return nil, false
	}
	return &wallet, true
}

func cloneWallet(in *platform.Wallet) *platform.Wallet {
	if in == nil {
		return nil
	}
	out := *in
	return &out
}

func walletCacheKey(tenantID string) string {
	return fmt.Sprintf("wallet:tenant:%s", strings.TrimSpace(tenantID))
}

func budgetWindowForTime(now time.Time) (string, time.Time) {
	utc := now.UTC()
	start := time.Date(utc.Year(), utc.Month(), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)
	return start.Format("200601"), end
}

func budgetWindowKey(tenantID string, windowID string) string {
	return fmt.Sprintf("budget:tenant:%s:window:%s", strings.TrimSpace(tenantID), strings.TrimSpace(windowID))
}

func budgetReservationKey(tenantID string, reservationID string) string {
	return fmt.Sprintf("budget:tenant:%s:reservation:%s", strings.TrimSpace(tenantID), strings.TrimSpace(reservationID))
}

func budgetReservationID(requestID string) string {
	requestID = strings.TrimSpace(requestID)
	if requestID == "" {
		return fmt.Sprintf("reservation:%d", time.Now().UnixNano())
	}
	return requestID
}

func entitlementNumber(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case nil:
		return 0, false
	case int:
		return float64(v), v >= 0
	case int64:
		return float64(v), v >= 0
	case float64:
		return v, v >= 0
	case json.Number:
		out, err := v.Float64()
		return out, err == nil && out >= 0
	default:
		return 0, false
	}
}

func cloneInterfaceMap(in map[string]interface{}) map[string]interface{} {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]interface{}, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

// EstimateCost estimates tokens and cost for a request.
func (s *Service) EstimateCost(ctx context.Context, req openai.ChatCompletionRequest) (platform.UsageCostBreakdown, error) {
	rate := s.exchangeRate(ctx)

	promptTokens := estimatePromptTokens(req)

	outputTokens := 256
	if req.MaxTokens != nil && *req.MaxTokens > 0 {
		outputTokens = *req.MaxTokens
	}

	price := s.modelPrice(ctx, req.Model)
	return costBreakdown(req.Model, promptTokens, outputTokens, price, rate.USDtoIRR), nil
}

// EstimateEmbeddingCost estimates cost for an embeddings request.
func (s *Service) EstimateEmbeddingCost(ctx context.Context, req openai.EmbeddingRequest) (platform.UsageCostBreakdown, error) {
	rate := s.exchangeRate(ctx)

	promptTokens := estimateEmbeddingPromptTokens(req.Input)
	price := s.modelPrice(ctx, req.Model)
	return costBreakdown(req.Model, promptTokens, 0, price, rate.USDtoIRR), nil
}

// FinalizeFromUsage recalculates final spend from provider usage.
func (s *Service) FinalizeFromUsage(ctx context.Context, model string, usage *openai.Usage, estimated platform.UsageCostBreakdown) (platform.UsageCostBreakdown, error) {
	if usage == nil {
		return estimated, nil
	}

	rate := s.exchangeRate(ctx)

	price := s.modelPrice(ctx, model)
	return costBreakdown(model, usage.PromptTokens, usage.CompletionTokens, price, rate.USDtoIRR), nil
}

func costBreakdown(model string, inputTokens int, outputTokens int, price priceInfo, exchangeRate float64) platform.UsageCostBreakdown {
	inputCost := (float64(inputTokens) / 1000.0) * price.Input
	outputCost := (float64(outputTokens) / 1000.0) * price.Output
	totalUSD := inputCost + outputCost

	return platform.UsageCostBreakdown{
		Model:         model,
		InputTokens:   inputTokens,
		OutputTokens:  outputTokens,
		TotalTokens:   inputTokens + outputTokens,
		InputCostUSD:  inputCost,
		OutputCostUSD: outputCost,
		TotalCostUSD:  totalUSD,
		ExchangeRate:  exchangeRate,
		TotalCostIRR:  totalUSD * exchangeRate,
	}
}

func (s *Service) modelPrice(ctx context.Context, model string) priceInfo {
	if s != nil && s.pricingCache != nil {
		return s.pricingCache.Get(ctx, model)
	}
	return modelPriceFromRegistry(model)
}

func (s *Service) persistDefaultExchangeRate(ctx context.Context) {
	if s == nil || s.store == nil {
		return
	}

	_, err := s.store.UpsertExchangeRate(ctx, s.defaultRate)
	if err != nil {
		return
	}
}

func (s *Service) exchangeRate(ctx context.Context) *platform.ExchangeRate {
	if s == nil {
		return nil
	}
	if s.exchangeRateCache != nil {
		if rate := s.exchangeRateCache.Get(ctx); rate != nil {
			return rate
		}
	}

	return &s.defaultRate
}
