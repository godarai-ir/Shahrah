package billing

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
	prommetrics "github.com/Delpak/shahrah/services/control-plane/internal/observability/prometheus"
	"github.com/Delpak/shahrah/services/control-plane/internal/runtimeprofile"
	"github.com/shopspring/decimal"
)

// ApplyPricingOverride replaces catalog component rates with the effective
// tenant price card after routing. The catalog revision remains pinned and a
// deterministic private suffix identifies the exact override set.
func (s *Service) ApplyPricingOverride(
	_ context.Context,
	estimated platform.UsageCostBreakdown,
	effective *platform.ModelPrice,
) (platform.UsageCostBreakdown, error) {
	if effective == nil || estimated.PricingSnapshot == nil || len(estimated.Components) == 0 {
		return platform.UsageCostBreakdown{}, ErrPricingUnavailable
	}

	bySignature := make(map[string]platform.ModelPricingRate, len(effective.Rates))
	for _, rate := range effective.Rates {
		signature := pricingComponentSignature(rate.Dimension, rate.Unit, rate.Condition)
		if _, duplicate := bySignature[signature]; duplicate {
			return platform.UsageCostBreakdown{}, fmt.Errorf(
				"%w: effective private pricing contains an ambiguous rate",
				ErrPricingUnavailable,
			)
		}

		bySignature[signature] = rate
	}

	snapshot := &platform.UsagePricingSnapshot{
		Provider: estimated.PricingSnapshot.Provider, Revision: estimated.PricingSnapshot.Revision,
		GeneratedAt: estimated.PricingSnapshot.GeneratedAt,
		Rates:       make(map[string]platform.UsageCostComponent, len(estimated.PricingSnapshot.Rates)),
		RateList:    append([]platform.UsageCostComponent(nil), estimated.PricingSnapshot.RateList...),
	}
	for key, component := range estimated.PricingSnapshot.Rates {
		snapshot.Rates[key] = component
	}

	inputExact, outputExact := decimal.Zero, decimal.Zero

	for index := range estimated.Components {
		component := estimated.Components[index]

		rate, ok := bySignature[pricingComponentSignature(component.Dimension, component.Unit, component.Condition)]
		if !ok {
			return platform.UsageCostBreakdown{}, fmt.Errorf(
				"%w: effective private pricing is missing %s",
				ErrPricingUnavailable,
				component.Dimension,
			)
		}

		price, err := decimal.NewFromString(rate.UnitPriceUSD)

		quantity, quantityErr := decimal.NewFromString(component.Quantity)
		if err != nil || quantityErr != nil || price.IsNegative() || quantity.IsNegative() {
			return platform.UsageCostBreakdown{}, ErrPricingUnavailable
		}

		cost := price.Mul(quantity)
		component.UnitPriceUSD = price.String()
		component.CostUSD = cost.String()

		estimated.Components[index] = component
		if strings.Contains(component.Dimension, "output") {
			outputExact = outputExact.Add(cost)
		} else {
			inputExact = inputExact.Add(cost)
		}

		for rateIndex := range snapshot.RateList {
			item := snapshot.RateList[rateIndex]
			if pricingComponentSignature(item.Dimension, item.Unit, item.Condition) ==
				pricingComponentSignature(component.Dimension, component.Unit, component.Condition) {
				item.UnitPriceUSD = price.String()
				snapshot.RateList[rateIndex] = item
			}
		}

		if item, exists := snapshot.Rates[component.Dimension]; exists &&
			pricingComponentSignature(item.Dimension, item.Unit, item.Condition) ==
				pricingComponentSignature(component.Dimension, component.Unit, component.Condition) {
			item.UnitPriceUSD = price.String()
			snapshot.Rates[component.Dimension] = item
		}
	}

	totalExact := inputExact.Add(outputExact)
	estimated.InputCostUSD, _ = inputExact.Float64()
	estimated.OutputCostUSD, _ = outputExact.Float64()
	estimated.TotalCostUSD, _ = totalExact.Float64()
	estimated.TotalCostIRR = estimated.TotalCostUSD * estimated.ExchangeRate
	fingerprint, _ := json.Marshal(effective.Rates) //nolint:errcheck // The value is composed exclusively of JSON-safe fields.
	revision := fmt.Sprintf("%s:private:%x", snapshot.Revision, sha256.Sum256(fingerprint))
	snapshot.Revision = revision
	estimated.PricingRevision = revision
	estimated.PricingSnapshot = snapshot

	return estimated, nil
}

func pricingComponentSignature(
	dimension string,
	unit string,
	condition map[string]interface{},
) string {
	encoded, _ := json.Marshal(struct { //nolint:errcheck // The value is composed exclusively of JSON-safe fields.
		Dimension string                 `json:"dimension"`
		Unit      string                 `json:"unit"`
		Condition map[string]interface{} `json:"condition,omitempty"`
	}{
		Dimension: strings.TrimSpace(dimension),
		Unit:      strings.TrimSpace(unit),
		Condition: condition,
	})

	return string(encoded)
}

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
//
//nolint:gocritic // Pricing override methods precede this core type to keep related pricing logic together.
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
	anthropicPricing     *AnthropicPricingSnapshotManager
	catalogPricing       map[string]*CatalogPricingSnapshotManager
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
		catalogPricing:       make(map[string]*CatalogPricingSnapshotManager),
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

func (s *Service) SetAnthropicPricingSnapshotManager(
	manager *AnthropicPricingSnapshotManager,
) {
	s.anthropicPricing = manager
}

// SetCatalogPricingSnapshotManager registers a provider-neutral catalog
// snapshot for providers whose pricing is maintained by Model Discovery.
func (s *Service) SetCatalogPricingSnapshotManager(
	provider string,
	manager *CatalogPricingSnapshotManager,
) {
	if s == nil {
		return
	}

	if s.catalogPricing == nil {
		s.catalogPricing = make(map[string]*CatalogPricingSnapshotManager)
	}

	s.catalogPricing[strings.ToLower(strings.TrimSpace(provider))] = manager
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
		if err := s.ReleaseBudget(ctx, reservation, "compat_check_only"); err != nil {
			return err
		}
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
			releaseErr := s.releaseBudgetTargetsBestEffort(ctx, budgetCache, reserved, now)
			return nil, errors.Join(reserveErr, releaseErr)
		}

		if result == nil {
			releaseErr := s.releaseBudgetTargetsBestEffort(ctx, budgetCache, reserved, now)
			return nil, errors.Join(ErrBudgetUnavailable, releaseErr)
		}

		if !result.Allowed {
			releaseErr := s.releaseBudgetTargetsBestEffort(ctx, budgetCache, reserved, now)

			if result.State == "exhausted" {
				return nil, errors.Join(ErrBudgetExceeded, releaseErr)
			}

			return nil, errors.Join(ErrBudgetUnavailable, releaseErr)
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
		AllowanceIRR: float64(wallet.BalanceIRR),
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

	price, err := s.modelPrice(ctx, req.Model)
	if err != nil {
		return platform.UsageCostBreakdown{}, err
	}

	if err := validateOpenAIServiceTier(req.Model, req.ServiceTier, price); err != nil {
		return platform.UsageCostBreakdown{}, err
	}

	breakdown := costBreakdown(req.Model, promptTokens, outputTokens, price, rate.USDtoIRR)
	if price.Provider == "anthropic" &&
		strings.EqualFold(strings.TrimSpace(req.Speed), "fast") {
		prefix := ""
		if strings.EqualFold(strings.TrimSpace(req.InferenceGeo), "us") {
			prefix = "data_residency_"
		}

		conditional, conditionErr := anthropicTokenCostBreakdown(
			req.Model, promptTokens, outputTokens, price, rate.USDtoIRR,
			prefix+"speed_input_tokens", prefix+"speed_output_tokens",
		)
		if conditionErr != nil {
			return platform.UsageCostBreakdown{}, conditionErr
		}

		conditional.PricingConditions = map[string]string{"speed": "fast"}
		breakdown = conditional
	}

	if price.Provider == "anthropic" &&
		!strings.EqualFold(strings.TrimSpace(req.Speed), "fast") &&
		strings.EqualFold(strings.TrimSpace(req.InferenceGeo), "us") {
		conditional, conditionErr := anthropicTokenCostBreakdown(
			req.Model, promptTokens, outputTokens, price, rate.USDtoIRR,
			"data_residency_input_tokens", "data_residency_output_tokens",
		)
		if conditionErr != nil {
			return platform.UsageCostBreakdown{}, conditionErr
		}

		conditional.PricingConditions = map[string]string{"inference_geo": "us"}
		breakdown = conditional
	}

	if price.Provider == "google-gemini" {
		tier, tierErr := googleGeminiServiceTier(req.ServiceTier)
		if tierErr != nil {
			return platform.UsageCostBreakdown{}, tierErr
		}

		if price.RuntimeProfile != nil &&
			!price.RuntimeProfile.Profile.AllowsServiceTier(tier) {
			return platform.UsageCostBreakdown{}, fmt.Errorf(
				"%w: model does not support Gemini service tier %s",
				ErrPricingUnavailable,
				tier,
			)
		}

		conditional, conditionErr := googleGeminiUsageCostBreakdown(
			req.Model,
			&openai.Usage{
				PromptTokens:     promptTokens,
				CompletionTokens: outputTokens,
				TotalTokens:      promptTokens + outputTokens,
				ServiceTier:      tier,
			},
			price,
			rate.USDtoIRR,
			map[string]string{"service_tier": tier},
		)
		if conditionErr != nil {
			return platform.UsageCostBreakdown{}, conditionErr
		}

		breakdown = conditional
	}

	if price.Provider == "openai" {
		conditional, conditionErr := openAIUsageCostBreakdown(
			req.Model,
			&openai.Usage{
				PromptTokens:     promptTokens,
				CompletionTokens: outputTokens,
				TotalTokens:      promptTokens + outputTokens,
				ServiceTier:      req.ServiceTier,
			},
			price,
			rate.USDtoIRR,
			map[string]string{"service_tier": req.ServiceTier},
		)
		if conditionErr != nil {
			return platform.UsageCostBreakdown{}, conditionErr
		}

		breakdown = conditional
	}

	if price.Provider == "mistral-ai" {
		conditional, conditionErr := mistralUsageCostBreakdown(
			req.Model,
			&openai.Usage{PromptTokens: promptTokens, CompletionTokens: outputTokens, TotalTokens: promptTokens + outputTokens, ServiceTier: req.ServiceTier},
			price,
			rate.USDtoIRR,
			map[string]string{"service_tier": req.ServiceTier},
		)
		if conditionErr != nil {
			return platform.UsageCostBreakdown{}, conditionErr
		}

		breakdown = conditional
	}

	return breakdown, nil
}

// EstimateEmbeddingCost estimates cost for an embeddings request.
func (s *Service) EstimateEmbeddingCost(ctx context.Context, req openai.EmbeddingRequest) (platform.UsageCostBreakdown, error) {
	rate := s.exchangeRate(ctx)

	promptTokens := estimateEmbeddingPromptTokens(req.Input)

	price, err := s.modelPrice(ctx, req.Model)
	if err != nil {
		return platform.UsageCostBreakdown{}, err
	}

	return costBreakdown(req.Model, promptTokens, 0, price, rate.USDtoIRR), nil
}

// FinalizeFromUsage recalculates final spend from provider usage.
func (s *Service) FinalizeFromUsage(ctx context.Context, model string, usage *openai.Usage, estimated platform.UsageCostBreakdown) (platform.UsageCostBreakdown, error) {
	if usage == nil {
		return estimated, nil
	}

	rate := s.exchangeRate(ctx)

	price, err := s.pinnedOrCurrentModelPrice(ctx, model, estimated)
	if err != nil {
		return platform.UsageCostBreakdown{}, err
	}

	breakdown, err := usageCostBreakdown(
		model,
		usage,
		price,
		rate.USDtoIRR,
		estimated.PricingConditions,
	)
	if err == nil {
		breakdown.ModelRuntimeProfile = estimated.ModelRuntimeProfile

		breakdown.Metadata = cloneUsageMetadata(estimated.Metadata)
		if strings.EqualFold(price.Provider, "xai") {
			breakdown, err = reconcileXAIProviderCost(breakdown, usage.CostInUSDTicks, rate.USDtoIRR)
			if err == nil {
				breakdown.Metadata["num_server_side_tools_used"] = usage.NumServerSideToolsUsed
			}
		}
	}

	return breakdown, err
}

// FinalizeBatchFromUsage uses the price revision pinned when the batch was
// reserved and applies batch token dimensions instead of synchronous rates.
func (s *Service) FinalizeBatchFromUsage(ctx context.Context, model string, usage *openai.Usage, estimated platform.UsageCostBreakdown) (platform.UsageCostBreakdown, error) {
	if usage == nil {
		return estimated, nil
	}

	rate := s.exchangeRate(ctx)

	price, err := s.pinnedOrCurrentModelPrice(ctx, model, estimated)
	if err != nil {
		return platform.UsageCostBreakdown{}, err
	}

	var breakdown platform.UsageCostBreakdown

	if strings.EqualFold(price.Provider, "openai") || strings.EqualFold(price.Provider, "xai") {
		if strings.Contains(strings.ToLower(model), "gpt-image") {
			breakdown, err = openAIImageCostBreakdown(
				model,
				usage,
				usage.PromptTokens,
				usage.CompletionTokens,
				usage.TotalTokens,
				price,
				rate.USDtoIRR,
				estimated.Metadata,
				"batch",
			)
		} else if strings.EqualFold(price.Provider, "xai") && strings.Contains(strings.ToLower(model), "grok-imagine-image") {
			breakdown, err = xAIImageCostBreakdown(
				model,
				usage.PromptTokens,
				usage.CompletionTokens,
				usage.TotalTokens,
				price,
				rate.USDtoIRR,
				estimated.Metadata,
			)
			breakdown.Batch = true
		} else {
			batchUsage := *usage
			batchUsage.ServiceTier = "batch"
			breakdown, err = openAIUsageCostBreakdown(
				model,
				&batchUsage,
				price,
				rate.USDtoIRR,
				estimated.PricingConditions,
			)
		}
	} else if strings.EqualFold(price.Provider, "mistral-ai") || strings.EqualFold(price.Provider, "mistral") {
		batchUsage := *usage
		batchUsage.ServiceTier = "batch"
		breakdown, err = mistralUsageCostBreakdown(
			model,
			&batchUsage,
			price,
			rate.USDtoIRR,
			estimated.PricingConditions,
		)
	} else {
		breakdown, err = anthropicUsageCostBreakdown(
			model,
			usage,
			price,
			rate.USDtoIRR,
			true,
			estimated.PricingConditions,
		)
	}

	if err == nil {
		breakdown.ModelRuntimeProfile = estimated.ModelRuntimeProfile

		breakdown.Metadata = cloneUsageMetadata(estimated.Metadata)
		if strings.EqualFold(price.Provider, "xai") {
			breakdown, err = reconcileXAIProviderCost(breakdown, usage.CostInUSDTicks, rate.USDtoIRR)
		}
	}

	return breakdown, err
}

// FinalizeImageUsage recalculates final spend for image endpoints.
func (s *Service) FinalizeImageUsage(ctx context.Context, model string, usage *openai.Usage, providerCostUSDTicks *int64, estimated platform.UsageCostBreakdown) (platform.UsageCostBreakdown, error) {
	if usage == nil && providerCostUSDTicks == nil {
		return estimated, nil
	}

	rate := s.exchangeRate(ctx)
	inputTokens, outputTokens, totalTokens := imageUsageTokenBreakdown(usage, estimated)

	price, err := s.pinnedOrCurrentModelPrice(ctx, model, estimated)
	if err != nil {
		return platform.UsageCostBreakdown{}, err
	}

	breakdown, err := imageCostBreakdown(
		model,
		usage,
		inputTokens,
		outputTokens,
		totalTokens,
		price,
		rate.USDtoIRR,
		estimated.Metadata,
	)
	if err != nil {
		return platform.UsageCostBreakdown{}, err
	}

	if strings.EqualFold(price.Provider, "xai") {
		var ticks int64
		if providerCostUSDTicks != nil {
			ticks = *providerCostUSDTicks
		} else if usage != nil {
			ticks = usage.CostInUSDTicks
		}

		return reconcileXAIProviderCost(breakdown, ticks, rate.USDtoIRR)
	}

	if providerCostUSDTicks != nil && *providerCostUSDTicks >= 0 {
		totalUSD := float64(*providerCostUSDTicks) / usdTicksPerDollar
		breakdown = applyProviderImageCost(breakdown, totalUSD)
		breakdown.ExchangeRate = rate.USDtoIRR

		breakdown.TotalCostIRR = totalUSD * rate.USDtoIRR
		if breakdown.Metadata == nil {
			breakdown.Metadata = cloneUsageMetadata(estimated.Metadata)
		}

		breakdown.Metadata["provider_cost_in_usd_ticks"] = *providerCostUSDTicks
		breakdown.Metadata["provider_cost_exact"] = true
	}

	return breakdown, nil
}

// PriceMeteredUsage prices non-token provider usage (for example speech
// characters or transcription minutes) against the same pinned public catalog
// snapshot used by token billing. Unknown legacy-provider dimensions leave the
// estimate unchanged; catalog-governed Mistral dimensions fail closed.
func (s *Service) PriceMeteredUsage(
	ctx context.Context,
	model string,
	dimension string,
	quantity string,
	estimated platform.UsageCostBreakdown,
) (platform.UsageCostBreakdown, error) {
	normalizedDimension := strings.TrimSpace(dimension)

	normalizedQuantity := strings.TrimSpace(quantity)
	if normalizedDimension == "" || normalizedQuantity == "" {
		return estimated, nil
	}

	parsedQuantity, err := decimal.NewFromString(normalizedQuantity)
	if err != nil || parsedQuantity.IsNegative() {
		return platform.UsageCostBreakdown{}, fmt.Errorf(
			"%w: invalid metered quantity",
			ErrPricingUnavailable,
		)
	}

	rate := s.exchangeRate(ctx)

	price, err := s.pinnedOrCurrentModelPrice(ctx, model, estimated)
	if err != nil {
		return platform.UsageCostBreakdown{}, err
	}

	pricingRate, ok := price.Rates[normalizedDimension]
	if !ok {
		if price.Provider == "mistral-ai" {
			return platform.UsageCostBreakdown{}, fmt.Errorf("%w: missing Mistral rate for %s", ErrPricingUnavailable, normalizedDimension)
		}

		return estimated, nil
	}

	component := usageCostComponentDecimal(
		normalizedDimension,
		parsedQuantity,
		pricingRate,
	)

	totalExact, err := decimal.NewFromString(component.CostUSD)
	if err != nil {
		return platform.UsageCostBreakdown{}, fmt.Errorf(
			"%w: invalid metered cost",
			ErrPricingUnavailable,
		)
	}

	totalUSD, _ := totalExact.Float64()

	out := estimated
	out.Model = model
	out.InputCostUSD = totalUSD
	out.OutputCostUSD = 0
	out.TotalCostUSD = totalUSD
	out.ExchangeRate = rate.USDtoIRR
	out.TotalCostIRR = totalUSD * rate.USDtoIRR
	out.PricingRevision = price.Revision
	out.ModelRuntimeProfile = price.RuntimeProfile
	out.Components = []platform.UsageCostComponent{component}
	out.Quantities = map[string]string{
		normalizedDimension: parsedQuantity.String(),
	}
	out.PricingSnapshot = usagePricingSnapshot(price)

	return out, nil
}

func costBreakdown(model string, inputTokens int, outputTokens int, price priceInfo, exchangeRate float64) platform.UsageCostBreakdown {
	inputRate := price.InputPerToken
	if inputRate.IsZero() && price.Input != 0 {
		inputRate = decimal.NewFromFloat(price.Input).Div(decimal.NewFromInt(1000))
	}

	outputRate := price.OutputPerToken
	if outputRate.IsZero() && price.Output != 0 {
		outputRate = decimal.NewFromFloat(price.Output).Div(decimal.NewFromInt(1000))
	}

	inputCostExact := inputRate.Mul(decimal.NewFromInt(int64(inputTokens)))
	outputCostExact := outputRate.Mul(decimal.NewFromInt(int64(outputTokens)))
	inputCost, _ := inputCostExact.Float64()
	outputCost, _ := outputCostExact.Float64()
	totalExact := inputCostExact.Add(outputCostExact)
	totalUSD, _ := totalExact.Float64()

	breakdown := platform.UsageCostBreakdown{
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
	if price.Provider != "" && len(price.Rates) > 0 {
		breakdown.PricingRevision = price.Revision

		breakdown.ModelRuntimeProfile = price.RuntimeProfile
		for _, dimension := range []string{"input_tokens", "output_tokens"} {
			rate, ok := price.Rates[dimension]
			if !ok {
				continue
			}

			quantity := inputTokens
			if dimension == "output_tokens" {
				quantity = outputTokens
			}

			breakdown.Components = append(
				breakdown.Components,
				usageCostComponent(dimension, int64(quantity), rate),
			)
		}

		breakdown.PricingSnapshot = usagePricingSnapshot(price)
	}

	return breakdown
}

func anthropicTokenCostBreakdown(model string, inputTokens, outputTokens int, price priceInfo, exchangeRate float64, inputDimension, outputDimension string) (platform.UsageCostBreakdown, error) {
	inputRate, inputOK := price.Rates[inputDimension]

	outputRate, outputOK := price.Rates[outputDimension]
	if !inputOK || !outputOK {
		return platform.UsageCostBreakdown{}, fmt.Errorf("%w: conditional Anthropic token rate is missing", ErrPricingUnavailable)
	}

	inputExact := inputRate.UnitPriceUSD.Mul(decimal.NewFromInt(int64(inputTokens)))
	outputExact := outputRate.UnitPriceUSD.Mul(decimal.NewFromInt(int64(outputTokens)))
	totalExact := inputExact.Add(outputExact)
	inputUSD, _ := inputExact.Float64()
	outputUSD, _ := outputExact.Float64()
	totalUSD, _ := totalExact.Float64()

	return platform.UsageCostBreakdown{
		Model: model, InputTokens: inputTokens, OutputTokens: outputTokens,
		TotalTokens: inputTokens + outputTokens, InputCostUSD: inputUSD,
		OutputCostUSD: outputUSD, TotalCostUSD: totalUSD, ExchangeRate: exchangeRate,
		TotalCostIRR: totalUSD * exchangeRate, PricingRevision: price.Revision,
		ModelRuntimeProfile: price.RuntimeProfile,
		Components: []platform.UsageCostComponent{
			usageCostComponent(inputDimension, int64(inputTokens), inputRate),
			usageCostComponent(outputDimension, int64(outputTokens), outputRate),
		},
		PricingSnapshot: usagePricingSnapshot(price),
	}, nil
}

func usageCostBreakdown(
	model string,
	usage *openai.Usage,
	price priceInfo,
	exchangeRate float64,
	conditions map[string]string,
) (platform.UsageCostBreakdown, error) {
	if price.Provider != "anthropic" {
		if price.Provider == "google-gemini" {
			return googleGeminiUsageCostBreakdown(
				model,
				usage,
				price,
				exchangeRate,
				conditions,
			)
		}

		if price.Provider == "openai" {
			return openAIUsageCostBreakdown(
				model,
				usage,
				price,
				exchangeRate,
				conditions,
			)
		}

		if price.Provider == "mistral-ai" {
			return mistralUsageCostBreakdown(model, usage, price, exchangeRate, conditions)
		}

		if price.Provider == "moonshot" {
			return moonshotUsageCostBreakdown(model, usage, price, exchangeRate)
		}

		if price.Provider == "deepseek" {
			return deepSeekUsageCostBreakdown(model, usage, price, exchangeRate)
		}

		return costBreakdown(
			model,
			usage.PromptTokens,
			usage.CompletionTokens,
			price,
			exchangeRate,
		), nil
	}

	return anthropicUsageCostBreakdown(model, usage, price, exchangeRate, false, conditions)
}

func moonshotUsageCostBreakdown(
	model string,
	usage *openai.Usage,
	price priceInfo,
	exchangeRate float64,
) (platform.UsageCostBreakdown, error) {
	if usage == nil {
		return platform.UsageCostBreakdown{}, fmt.Errorf("%w: Moonshot usage is missing", ErrPricingUnavailable)
	}

	cached := 0
	if usage.PromptTokensDetails != nil {
		cached = usage.PromptTokensDetails.CachedTokens
	}

	if cached < 0 || cached > usage.PromptTokens {
		return platform.UsageCostBreakdown{}, fmt.Errorf("%w: invalid Moonshot cached token usage", ErrPricingUnavailable)
	}

	quantities := map[string]int64{
		"input_tokens":      int64(usage.PromptTokens - cached),
		"cache_read_tokens": int64(cached),
		"output_tokens":     int64(usage.CompletionTokens),
	}
	dimensions := []string{"input_tokens", "cache_read_tokens", "output_tokens"}
	components := make([]platform.UsageCostComponent, 0, len(dimensions))
	inputExact, outputExact := decimal.Zero, decimal.Zero

	for _, dimension := range dimensions {
		quantity := quantities[dimension]
		if quantity == 0 {
			continue
		}

		rate, ok := price.Rates[dimension]
		if !ok || len(rate.Condition) != 0 {
			return platform.UsageCostBreakdown{}, fmt.Errorf("%w: missing Moonshot rate for %s", ErrPricingUnavailable, dimension)
		}

		cost := rate.UnitPriceUSD.Mul(decimal.NewFromInt(quantity))
		if dimension == "output_tokens" {
			outputExact = outputExact.Add(cost)
		} else {
			inputExact = inputExact.Add(cost)
		}

		components = append(components, usageCostComponent(dimension, quantity, rate))
	}

	inputUSD, _ := inputExact.Float64()
	outputUSD, _ := outputExact.Float64()
	totalUSD := inputUSD + outputUSD

	totalTokens := usage.TotalTokens
	if totalTokens <= 0 {
		totalTokens = usage.PromptTokens + usage.CompletionTokens
	}

	return platform.UsageCostBreakdown{
		Model: model, InputTokens: usage.PromptTokens, OutputTokens: usage.CompletionTokens,
		TotalTokens: totalTokens, CacheReadInputTokens: cached,
		InputCostUSD: inputUSD, OutputCostUSD: outputUSD, TotalCostUSD: totalUSD,
		ExchangeRate: exchangeRate, TotalCostIRR: totalUSD * exchangeRate,
		PricingRevision: price.Revision, ModelRuntimeProfile: price.RuntimeProfile,
		PricingSnapshot: usagePricingSnapshot(price), Components: components,
	}, nil
}

func deepSeekUsageCostBreakdown(
	model string,
	usage *openai.Usage,
	price priceInfo,
	exchangeRate float64,
) (platform.UsageCostBreakdown, error) {
	if usage == nil {
		return platform.UsageCostBreakdown{}, fmt.Errorf("%w: DeepSeek usage is missing", ErrPricingUnavailable)
	}

	hit := usage.PromptCacheHitTokens

	miss := usage.PromptCacheMissTokens
	if hit < 0 || miss < 0 || hit+miss != usage.PromptTokens {
		return platform.UsageCostBreakdown{}, fmt.Errorf("%w: invalid DeepSeek cache token usage", ErrPricingUnavailable)
	}

	quantities := map[string]int64{
		"input_tokens":      int64(miss),
		"cache_read_tokens": int64(hit),
		"output_tokens":     int64(usage.CompletionTokens),
	}
	dimensions := []string{"input_tokens", "cache_read_tokens", "output_tokens"}
	components := make([]platform.UsageCostComponent, 0, len(dimensions))
	inputExact, outputExact := decimal.Zero, decimal.Zero

	for _, dimension := range dimensions {
		quantity := quantities[dimension]
		if quantity == 0 {
			continue
		}

		rate, ok := price.Rates[dimension]
		if !ok || len(rate.Condition) != 0 {
			return platform.UsageCostBreakdown{}, fmt.Errorf("%w: missing DeepSeek rate for %s", ErrPricingUnavailable, dimension)
		}

		cost := rate.UnitPriceUSD.Mul(decimal.NewFromInt(quantity))
		if dimension == "output_tokens" {
			outputExact = outputExact.Add(cost)
		} else {
			inputExact = inputExact.Add(cost)
		}

		components = append(components, usageCostComponent(dimension, quantity, rate))
	}

	inputUSD, _ := inputExact.Float64()
	outputUSD, _ := outputExact.Float64()
	totalUSD := inputUSD + outputUSD

	totalTokens := usage.TotalTokens
	if totalTokens <= 0 {
		totalTokens = usage.PromptTokens + usage.CompletionTokens
	}

	return platform.UsageCostBreakdown{
		Model: model, InputTokens: usage.PromptTokens, OutputTokens: usage.CompletionTokens,
		TotalTokens: totalTokens, CacheReadInputTokens: hit,
		InputCostUSD: inputUSD, OutputCostUSD: outputUSD, TotalCostUSD: totalUSD,
		ExchangeRate: exchangeRate, TotalCostIRR: totalUSD * exchangeRate,
		PricingRevision: price.Revision, ModelRuntimeProfile: price.RuntimeProfile,
		PricingSnapshot: usagePricingSnapshot(price), Components: components,
	}, nil
}

func mistralUsageCostBreakdown(model string, usage *openai.Usage, price priceInfo, exchangeRate float64, conditions map[string]string) (platform.UsageCostBreakdown, error) {
	if usage == nil {
		return platform.UsageCostBreakdown{}, fmt.Errorf("%w: Mistral usage is missing", ErrPricingUnavailable)
	}

	cached := 0
	if usage.PromptTokensDetails != nil {
		cached = usage.PromptTokensDetails.CachedTokens
	}

	if cached < 0 || cached > usage.PromptTokens {
		return platform.UsageCostBreakdown{}, fmt.Errorf("%w: invalid Mistral cached token usage", ErrPricingUnavailable)
	}

	inputDimension, outputDimension, cacheDimension := "input_tokens", "output_tokens", "cache_read_tokens"

	if strings.EqualFold(strings.TrimSpace(usage.ServiceTier), "batch") || strings.EqualFold(strings.TrimSpace(conditions["service_tier"]), "batch") {
		if conditions == nil {
			conditions = map[string]string{}
		}

		conditions["api"] = "batches"
	}

	quantities := map[string]int64{
		inputDimension:  int64(usage.PromptTokens - cached),
		outputDimension: int64(usage.CompletionTokens),
	}
	if cached > 0 {
		if _, ok := price.Rates[cacheDimension]; !ok {
			return platform.UsageCostBreakdown{}, fmt.Errorf("%w: missing Mistral rate for %s", ErrPricingUnavailable, cacheDimension)
		}

		quantities[cacheDimension] = int64(cached)
	}

	for dimension, quantity := range usage.ServerToolUse {
		if quantity > 0 {
			quantities[dimension] += int64(quantity)
		}
	}

	dimensions := make([]string, 0, len(quantities))
	for dimension := range quantities {
		dimensions = append(dimensions, dimension)
	}

	sort.Strings(dimensions)

	totalExact, inputExact, outputExact := decimal.Zero, decimal.Zero, decimal.Zero

	components := make([]platform.UsageCostComponent, 0, len(dimensions))
	for _, dimension := range dimensions {
		quantity := quantities[dimension]
		if quantity == 0 {
			continue
		}

		rate, rateErr := selectOpenAIRate(price.RateList, dimension, usage.PromptTokens, "", conditions)
		if rateErr != nil {
			return platform.UsageCostBreakdown{}, fmt.Errorf("%w: missing or ambiguous Mistral rate for %s", ErrPricingUnavailable, dimension)
		}

		cost := rate.UnitPriceUSD.Mul(decimal.NewFromInt(quantity))

		totalExact = totalExact.Add(cost)
		if dimension == outputDimension || strings.Contains(dimension, "output") {
			outputExact = outputExact.Add(cost)
		} else {
			inputExact = inputExact.Add(cost)
		}

		components = append(components, usageCostComponent(dimension, quantity, rate))
	}

	inputUSD, _ := inputExact.Float64()
	outputUSD, _ := outputExact.Float64()
	totalUSD, _ := totalExact.Float64()

	totalTokens := usage.TotalTokens
	if totalTokens <= 0 {
		totalTokens = usage.PromptTokens + usage.CompletionTokens
	}

	return platform.UsageCostBreakdown{
		Model: model, InputTokens: usage.PromptTokens, OutputTokens: usage.CompletionTokens,
		TotalTokens: totalTokens, CacheReadInputTokens: cached,
		InputCostUSD: inputUSD, OutputCostUSD: outputUSD, TotalCostUSD: totalUSD,
		ExchangeRate: exchangeRate, TotalCostIRR: totalUSD * exchangeRate,
		PricingRevision: price.Revision, ModelRuntimeProfile: price.RuntimeProfile,
		PricingSnapshot: usagePricingSnapshot(price), Components: components,
		ServerToolUse: usage.ServerToolUse,
	}, nil
}

func openAIUsageCostBreakdown(
	model string,
	usage *openai.Usage,
	price priceInfo,
	exchangeRate float64,
	conditions map[string]string,
) (platform.UsageCostBreakdown, error) {
	requestedTier := strings.TrimSpace(usage.ServiceTier)
	if requestedTier == "" {
		requestedTier = strings.TrimSpace(conditions["service_tier"])
	}

	if err := validateOpenAIServiceTier(model, requestedTier, price); err != nil {
		return platform.UsageCostBreakdown{}, err
	}

	tier, err := normalizeOpenAIServiceTier(requestedTier)
	if err != nil {
		return platform.UsageCostBreakdown{}, err
	}

	promptAudio, promptCached, promptCacheWrite, completionAudio := 0, 0, 0, 0
	if usage.PromptTokensDetails != nil {
		promptAudio = usage.PromptTokensDetails.AudioTokens
		promptCached = usage.PromptTokensDetails.CachedTokens
		promptCacheWrite = usage.PromptTokensDetails.CacheWriteTokens
	}

	if usage.CompletionTokensDetails != nil {
		completionAudio = usage.CompletionTokensDetails.AudioTokens
	}

	if promptAudio < 0 || promptCached < 0 || promptCacheWrite < 0 || completionAudio < 0 ||
		promptAudio+promptCached+promptCacheWrite > usage.PromptTokens ||
		completionAudio > usage.CompletionTokens {
		return platform.UsageCostBreakdown{}, fmt.Errorf(
			"%w: invalid OpenAI usage token detail totals",
			ErrPricingUnavailable,
		)
	}

	inputDimension, outputDimension := "input_tokens", "output_tokens"
	cacheDimension := "cache_read_tokens"
	cacheWriteDimension := "cache_write_tokens"

	switch tier {
	case "batch":
		inputDimension, outputDimension = "batch_input_tokens", "batch_output_tokens"
		cacheDimension = "batch_cache_read_tokens"
		cacheWriteDimension = "batch_cache_write_tokens"
	case "flex":
		inputDimension, outputDimension = "flex_input_tokens", "flex_output_tokens"
		cacheDimension = "flex_cache_read_tokens"
		cacheWriteDimension = "flex_cache_write_tokens"
	case "priority":
		inputDimension, outputDimension = "speed_input_tokens", "speed_output_tokens"
		cacheDimension = "speed_cache_read_tokens"
		cacheWriteDimension = "speed_cache_write_tokens"
	}

	longContextThreshold := 272000
	if strings.EqualFold(price.Provider, "xai") {
		longContextThreshold = 200000
	}

	if usage.PromptTokens >= longContextThreshold {
		switch tier {
		case "priority":
			if strings.EqualFold(price.Provider, "xai") {
				inputDimension, outputDimension = "speed_long_context_input_tokens", "speed_long_context_output_tokens"
				cacheDimension = "speed_long_context_cache_read_tokens"
				cacheWriteDimension = "speed_long_context_cache_write_tokens"
			} else {
				return platform.UsageCostBreakdown{}, newUnsupportedServiceTierError(
					price,
					model,
					requestedTier,
				)
			}
		case "batch":
			inputDimension, outputDimension = "batch_long_context_input_tokens", "batch_long_context_output_tokens"
			cacheDimension = "batch_long_context_cache_read_tokens"
			cacheWriteDimension = "batch_long_context_cache_write_tokens"
		case "flex":
			inputDimension, outputDimension = "flex_long_context_input_tokens", "flex_long_context_output_tokens"
			cacheDimension = "flex_long_context_cache_read_tokens"
			cacheWriteDimension = "flex_long_context_cache_write_tokens"
		default:
			inputDimension, outputDimension = "long_context_input_tokens", "long_context_output_tokens"
			cacheDimension = "long_context_cache_read_tokens"
			cacheWriteDimension = "long_context_cache_write_tokens"
		}
	}

	if promptCached > 0 && !hasOpenAIRate(price.RateList, cacheDimension, usage.PromptTokens, tier, conditions) {
		// Pro models publish no cached-input discount. Cached tokens are still
		// billable, at the applicable ordinary input rate.
		cacheDimension = inputDimension
	}

	if promptCacheWrite > 0 && !hasOpenAIRate(price.RateList, cacheWriteDimension, usage.PromptTokens, tier, conditions) {
		cacheWriteDimension = inputDimension
	}

	quantities := make(map[string]int64, 7)
	quantities[inputDimension] += int64(usage.PromptTokens - promptAudio - promptCached - promptCacheWrite)
	quantities[outputDimension] += int64(usage.CompletionTokens - completionAudio)
	quantities[cacheDimension] += int64(promptCached)
	quantities[cacheWriteDimension] += int64(promptCacheWrite)
	quantities["audio_input_tokens"] += int64(promptAudio)
	quantities["audio_output_tokens"] += int64(completionAudio)

	for tool, quantity := range usage.ServerToolUse {
		if quantity > 0 {
			quantities[tool] += int64(quantity)
		}
	}

	if webSearchCalls := usage.ServerToolUse["web_search_requests"]; webSearchCalls > 0 &&
		openAIWebSearchUsesFixedInputBlock(model) {
		quantities[inputDimension] += int64(webSearchCalls * 8000)
	}

	dimensions := make([]string, 0, len(quantities))
	for dimension := range quantities {
		dimensions = append(dimensions, dimension)
	}

	sort.Strings(dimensions)

	totalExact, inputExact, outputExact := decimal.Zero, decimal.Zero, decimal.Zero

	components := make([]platform.UsageCostComponent, 0, len(dimensions))
	for _, dimension := range dimensions {
		quantity := quantities[dimension]
		if quantity == 0 {
			continue
		}

		selected, err := selectOpenAIRate(price.RateList, dimension, usage.PromptTokens, tier, conditions)
		if err != nil {
			return platform.UsageCostBreakdown{}, err
		}

		cost := selected.UnitPriceUSD.Mul(decimal.NewFromInt(quantity))

		totalExact = totalExact.Add(cost)
		if strings.Contains(dimension, "output") {
			outputExact = outputExact.Add(cost)
		} else {
			inputExact = inputExact.Add(cost)
		}

		components = append(components, usageCostComponent(dimension, quantity, selected))
	}

	inputUSD, _ := inputExact.Float64()
	outputUSD, _ := outputExact.Float64()
	totalUSD, _ := totalExact.Float64()

	totalTokens := usage.TotalTokens
	if totalTokens <= 0 {
		totalTokens = usage.PromptTokens + usage.CompletionTokens
	}

	applied := map[string]string{"service_tier": tier}

	for key, value := range conditions {
		if strings.TrimSpace(value) != "" {
			applied[key] = value
		}
	}

	return platform.UsageCostBreakdown{
		Model: model, InputTokens: usage.PromptTokens,
		OutputTokens: usage.CompletionTokens, TotalTokens: totalTokens,
		InputCostUSD: inputUSD, OutputCostUSD: outputUSD,
		TotalCostUSD: totalUSD, ExchangeRate: exchangeRate,
		TotalCostIRR:               totalUSD * exchangeRate,
		CacheCreation5mInputTokens: promptCacheWrite,
		CacheReadInputTokens:       promptCached,
		ServerToolUse:              usage.ServerToolUse,
		ServiceTier:                tier,
		PricingConditions:          applied, PricingRevision: price.Revision,
		ModelRuntimeProfile: price.RuntimeProfile, Components: components,
		PricingSnapshot: usagePricingSnapshot(price),
	}, nil
}

func openAIWebSearchUsesFixedInputBlock(model string) bool {
	value := strings.ToLower(strings.TrimSpace(model))
	value = strings.TrimPrefix(value, "openai/")
	value = strings.TrimPrefix(value, "openai:")

	return value == "gpt-4o-mini" ||
		strings.HasPrefix(value, "gpt-4o-mini-") ||
		value == "gpt-4.1-mini" ||
		strings.HasPrefix(value, "gpt-4.1-mini-")
}

func hasOpenAIRate(
	rates []normalizedPricingRate,
	dimension string,
	promptTokens int,
	tier string,
	conditions map[string]string,
) bool {
	for _, rate := range rates {
		if rate.Dimension == dimension &&
			openAIRateConditionsMatch(rate.Condition, promptTokens, tier, conditions) {
			return true
		}
	}

	return false
}

func selectOpenAIRate(
	rates []normalizedPricingRate,
	dimension string,
	promptTokens int,
	tier string,
	conditions map[string]string,
) (normalizedPricingRate, error) {
	matches := make([]normalizedPricingRate, 0, 1)
	bestSpecificity := -1

	for _, rate := range rates {
		if rate.Dimension != dimension ||
			!openAIRateConditionsMatch(rate.Condition, promptTokens, tier, conditions) {
			continue
		}

		specificity := len(rate.Condition)
		if specificity > bestSpecificity {
			matches = matches[:0]
			bestSpecificity = specificity
		}

		if specificity == bestSpecificity {
			matches = append(matches, rate)
		}
	}

	if len(matches) != 1 {
		return normalizedPricingRate{}, fmt.Errorf(
			"%w: OpenAI rate selection for %s matched %d rates",
			ErrPricingRateUnavailable,
			dimension,
			len(matches),
		)
	}

	return matches[0], nil
}

func normalizeOpenAIServiceTier(value string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "default", "standard":
		return "standard", nil
	case "batch", "flex", "priority":
		return strings.ToLower(strings.TrimSpace(value)), nil
	case "fast":
		// OpenAI reports "priority" when the request uses the newer "fast" alias.
		return "priority", nil
	default:
		return "", &UnsupportedServiceTierError{
			Provider: "openai",
			Tier:     value,
		}
	}
}

func validateOpenAIServiceTier(model, requestedTier string, price priceInfo) error {
	if !strings.EqualFold(strings.TrimSpace(price.Provider), "openai") ||
		strings.TrimSpace(requestedTier) == "" {
		return nil
	}

	normalizedTier, err := normalizeOpenAIServiceTier(requestedTier)
	if err != nil {
		var tierErr *UnsupportedServiceTierError
		if errors.As(err, &tierErr) {
			tierErr.Model = openAIModelName(model, price)
		}

		return err
	}

	if price.RuntimeProfile == nil {
		return nil
	}

	if price.RuntimeProfile.Profile.AllowsServiceTier(normalizedTier) {
		return nil
	}

	return newUnsupportedServiceTierError(price, model, requestedTier)
}

func newUnsupportedServiceTierError(
	price priceInfo,
	model string,
	tier string,
) error {
	return &UnsupportedServiceTierError{
		Provider: price.Provider,
		Model:    openAIModelName(model, price),
		Tier:     tier,
	}
}

func openAIModelName(model string, price priceInfo) string {
	if price.RuntimeProfile != nil &&
		strings.TrimSpace(price.RuntimeProfile.Slug) != "" {
		return strings.TrimSpace(price.RuntimeProfile.Slug)
	}

	return normalizeCatalogPricingModelID(model, "openai")
}

func openAIRateConditionsMatch(
	rateConditions map[string]interface{},
	promptTokens int,
	tier string,
	conditions map[string]string,
) bool {
	for key, expected := range rateConditions {
		switch key {
		case "min_prompt_tokens":
			minimum, ok := numericCondition(expected)
			if !ok || int64(promptTokens) < minimum {
				return false
			}
		case "service_tier":
			if !strings.EqualFold(strings.TrimSpace(fmt.Sprint(expected)), tier) {
				return false
			}
		default:
			actual, ok := conditions[key]
			if !ok || !strings.EqualFold(strings.TrimSpace(actual), strings.TrimSpace(fmt.Sprint(expected))) {
				return false
			}
		}
	}

	return true
}

func numericCondition(value interface{}) (int64, bool) {
	switch typed := value.(type) {
	case float64:
		return int64(typed), typed >= 0 && typed == float64(int64(typed))
	case int:
		return int64(typed), typed >= 0
	case int64:
		return typed, typed >= 0
	case json.Number:
		out, err := typed.Int64()
		return out, err == nil && out >= 0
	default:
		out, err := strconv.ParseInt(strings.TrimSpace(fmt.Sprint(value)), 10, 64)
		return out, err == nil && out >= 0
	}
}

func googleGeminiServiceTier(value string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "default", "standard":
		return "standard", nil
	case "flex":
		return "flex", nil
	case "priority":
		return "priority", nil
	default:
		return "", fmt.Errorf(
			"%w: unsupported Gemini service tier %q",
			ErrPricingUnavailable,
			value,
		)
	}
}

func googleGeminiUsageCostBreakdown(
	model string,
	usage *openai.Usage,
	price priceInfo,
	exchangeRate float64,
	conditions map[string]string,
) (platform.UsageCostBreakdown, error) {
	tierValue := usage.ServiceTier
	if strings.TrimSpace(tierValue) == "" {
		tierValue = conditions["service_tier"]
	}

	tier, err := googleGeminiServiceTier(tierValue)
	if err != nil {
		return platform.UsageCostBreakdown{}, err
	}

	promptAudio := 0
	promptCached := 0

	if usage.PromptTokensDetails != nil {
		promptAudio = usage.PromptTokensDetails.AudioTokens
		promptCached = usage.PromptTokensDetails.CachedTokens
	}

	completionAudio := 0
	completionImage := 0

	if usage.CompletionTokensDetails != nil {
		completionAudio = usage.CompletionTokensDetails.AudioTokens
		completionImage = usage.CompletionTokensDetails.ImageTokens
	}

	if promptAudio < 0 || promptCached < 0 || completionAudio < 0 ||
		promptAudio+promptCached > usage.PromptTokens ||
		completionAudio+completionImage > usage.CompletionTokens {
		return platform.UsageCostBreakdown{}, fmt.Errorf(
			"%w: invalid Gemini token detail totals",
			ErrPricingUnavailable,
		)
	}
	// The OpenAI-compatible usage schema does not identify the modality of
	// cached tokens. Refuse an ambiguous mixed audio/cache bill instead of
	// silently applying the cheaper text cache rate.
	if promptAudio > 0 && promptCached > 0 {
		return platform.UsageCostBreakdown{}, fmt.Errorf(
			"%w: Gemini audio cache modality is ambiguous",
			ErrPricingUnavailable,
		)
	}

	prefix := ""
	if tier != "standard" {
		prefix = tier + "_"
	}

	if usage.PromptTokens > 200000 {
		prefix += "long_context_"
	}

	inputDimension := prefix + "input_tokens"
	audioInputDimension := prefix + "audio_input_tokens"

	textInputTokens := usage.PromptTokens - promptAudio - promptCached
	if promptAudio > 0 {
		if _, hasAudioRate := price.Rates[audioInputDimension]; !hasAudioRate {
			// A missing audio-specific dimension means the official table uses
			// one published input rate for every supported input modality.
			textInputTokens += promptAudio
			promptAudio = 0
		}
	}

	quantities := map[string]int64{
		inputDimension:                 int64(textInputTokens),
		prefix + "output_tokens":       int64(usage.CompletionTokens - completionAudio - completionImage),
		audioInputDimension:            int64(promptAudio),
		prefix + "audio_output_tokens": int64(completionAudio),
		prefix + "image_output_tokens": int64(completionImage),
		prefix + "cache_read_tokens":   int64(promptCached),
	}
	for dimension, quantity := range usage.ServerToolUse {
		if quantity > 0 {
			quantities[dimension] += int64(quantity)
		}
	}

	dimensions := make([]string, 0, len(quantities))
	for dimension := range quantities {
		dimensions = append(dimensions, dimension)
	}

	sort.Strings(dimensions)

	totalExact := decimal.Zero
	inputExact := decimal.Zero
	outputExact := decimal.Zero

	components := make([]platform.UsageCostComponent, 0, len(dimensions))
	for _, dimension := range dimensions {
		quantity := quantities[dimension]
		if quantity == 0 {
			continue
		}

		rate, ok := price.Rates[dimension]
		if !ok {
			return platform.UsageCostBreakdown{}, fmt.Errorf(
				"%w: missing Gemini rate for %s",
				ErrPricingUnavailable,
				dimension,
			)
		}

		cost := rate.UnitPriceUSD.Mul(decimal.NewFromInt(quantity))

		totalExact = totalExact.Add(cost)
		if strings.Contains(dimension, "output_") {
			outputExact = outputExact.Add(cost)
		} else {
			inputExact = inputExact.Add(cost)
		}

		components = append(
			components,
			usageCostComponent(dimension, quantity, rate),
		)
	}

	inputUSD, _ := inputExact.Float64()
	outputUSD, _ := outputExact.Float64()
	totalUSD, _ := totalExact.Float64()

	appliedConditions := make(map[string]string, len(conditions)+1)
	for key, value := range conditions {
		appliedConditions[key] = value
	}

	appliedConditions["service_tier"] = tier

	totalTokens := usage.TotalTokens
	if totalTokens <= 0 {
		totalTokens = usage.PromptTokens + usage.CompletionTokens
	}

	return platform.UsageCostBreakdown{
		Model:                model,
		InputTokens:          usage.PromptTokens,
		OutputTokens:         usage.CompletionTokens,
		TotalTokens:          totalTokens,
		InputCostUSD:         inputUSD,
		OutputCostUSD:        outputUSD,
		TotalCostUSD:         totalUSD,
		ExchangeRate:         exchangeRate,
		TotalCostIRR:         totalUSD * exchangeRate,
		CacheReadInputTokens: promptCached,
		ServiceTier:          tier,
		PricingConditions:    appliedConditions,
		PricingRevision:      price.Revision,
		ModelRuntimeProfile:  price.RuntimeProfile,
		Components:           components,
		PricingSnapshot:      usagePricingSnapshot(price),
	}, nil
}

func anthropicUsageCostBreakdown(
	model string,
	usage *openai.Usage,
	price priceInfo,
	exchangeRate float64,
	batch bool,
	conditions map[string]string,
) (platform.UsageCostBreakdown, error) {
	inputDimension := "input_tokens"
	outputDimension := "output_tokens"
	cache5mDimension := "cache_write_5m_tokens"
	cache1hDimension := "cache_write_1h_tokens"
	cacheReadDimension := "cache_read_tokens"

	if batch {
		inputDimension = "batch_input_tokens"
		outputDimension = "batch_output_tokens"
		cache5mDimension = "batch_cache_write_5m_tokens"
		cache1hDimension = "batch_cache_write_1h_tokens"
		cacheReadDimension = "batch_cache_read_tokens"
	}

	if strings.EqualFold(strings.TrimSpace(usage.Speed), "fast") {
		if batch {
			return platform.UsageCostBreakdown{}, fmt.Errorf(
				"%w: Anthropic fast-mode is unavailable with Batch",
				ErrPricingUnavailable,
			)
		}

		inputDimension = "speed_input_tokens"
		outputDimension = "speed_output_tokens"
		cache5mDimension = "speed_cache_write_5m_tokens"
		cache1hDimension = "speed_cache_write_1h_tokens"
		cacheReadDimension = "speed_cache_read_tokens"
	}

	if strings.EqualFold(conditions["inference_geo"], "us") {
		inputDimension = "data_residency_" + inputDimension
		outputDimension = "data_residency_" + outputDimension
		cache5mDimension = "data_residency_" + cache5mDimension
		cache1hDimension = "data_residency_" + cache1hDimension
		cacheReadDimension = "data_residency_" + cacheReadDimension
	}

	quantities := map[string]int64{
		inputDimension:     int64(usage.PromptTokens),
		outputDimension:    int64(usage.CompletionTokens),
		cache5mDimension:   int64(usage.CacheCreation5mInputTokens),
		cache1hDimension:   int64(usage.CacheCreation1hInputTokens),
		cacheReadDimension: int64(usage.CacheReadInputTokens),
	}
	if usage.CacheCreationInputTokens > 0 &&
		quantities[cache5mDimension] == 0 &&
		quantities[cache1hDimension] == 0 {
		quantities[cache5mDimension] = int64(
			usage.CacheCreationInputTokens,
		)
	}

	for tool, quantity := range usage.ServerToolUse {
		quantities[tool] = int64(quantity)
	}

	totalExact := decimal.Zero
	inputExact := decimal.Zero
	outputExact := decimal.Zero
	components := make([]platform.UsageCostComponent, 0, len(quantities))

	dimensions := make([]string, 0, len(quantities))
	for dimension := range quantities {
		dimensions = append(dimensions, dimension)
	}

	sort.Strings(dimensions)

	for _, dimension := range dimensions {
		quantity := quantities[dimension]
		if quantity == 0 {
			continue
		}

		rate, ok := price.Rates[dimension]
		if dimension == "code_execution_requests" {
			rate, ok = anthropicCodeExecutionUsageRate(price, conditions)
			if !ok {
				return platform.UsageCostBreakdown{}, fmt.Errorf(
					"%w: billable code execution requires "+
						"provider cost reconciliation",
					ErrPricingUnavailable,
				)
			}
		}

		if !ok {
			return platform.UsageCostBreakdown{}, fmt.Errorf(
				"%w: missing rate for %s",
				ErrPricingUnavailable,
				dimension,
			)
		}

		cost := rate.UnitPriceUSD.Mul(decimal.NewFromInt(quantity))

		totalExact = totalExact.Add(cost)
		if dimension == outputDimension {
			outputExact = outputExact.Add(cost)
		} else {
			inputExact = inputExact.Add(cost)
		}

		components = append(
			components,
			usageCostComponent(dimension, quantity, rate),
		)
		prommetrics.RecordAnthropicUsageDimension(dimension, quantity)
	}

	inputUSD, _ := inputExact.Float64()
	outputUSD, _ := outputExact.Float64()
	totalUSD, _ := totalExact.Float64()

	return platform.UsageCostBreakdown{
		Model:                      model,
		InputTokens:                usage.PromptTokens,
		OutputTokens:               usage.CompletionTokens,
		TotalTokens:                usage.TotalTokens,
		InputCostUSD:               inputUSD,
		OutputCostUSD:              outputUSD,
		TotalCostUSD:               totalUSD,
		ExchangeRate:               exchangeRate,
		TotalCostIRR:               totalUSD * exchangeRate,
		CacheCreation5mInputTokens: usage.CacheCreation5mInputTokens,
		CacheCreation1hInputTokens: usage.CacheCreation1hInputTokens,
		CacheReadInputTokens:       usage.CacheReadInputTokens,
		ServerToolUse:              usage.ServerToolUse,
		ServiceTier:                usage.ServiceTier,
		Speed:                      usage.Speed,
		Batch:                      batch,
		PricingConditions:          conditions,
		PricingRevision:            price.Revision,
		ModelRuntimeProfile:        price.RuntimeProfile,
		Components:                 components,
		PricingSnapshot:            usagePricingSnapshot(price),
	}, nil
}

func anthropicCodeExecutionUsageRate(
	price priceInfo,
	conditions map[string]string,
) (normalizedPricingRate, bool) {
	if conditions[platform.AnthropicCodeExecutionPricingCondition] !=
		platform.AnthropicCodeExecutionFreeWithCurrentWebTools {
		return normalizedPricingRate{}, false
	}

	hourlyRate, ok := price.Rates["code_execution_container_hours"]
	if !ok || hourlyRate.Unit != "container_hour" {
		return normalizedPricingRate{}, false
	}

	return normalizedPricingRate{
		Dimension: "code_execution_requests",
		Unit:      "request",
		Condition: map[string]interface{}{
			platform.AnthropicCodeExecutionPricingCondition: platform.AnthropicCodeExecutionFreeWithCurrentWebTools,
			"source_dimension": hourlyRate.Dimension,
		},
		UnitPriceUSD: decimal.Zero,
		Region:       hourlyRate.Region,
		SourceURL:    hourlyRate.SourceURL,
	}, true
}

const usdTicksPerDollar = 10_000_000_000.0

func imageUsageTokenBreakdown(usage *openai.Usage, estimated platform.UsageCostBreakdown) (int, int, int) {
	if usage == nil {
		return estimated.InputTokens, estimated.OutputTokens, estimated.TotalTokens
	}

	inputTokens := usage.PromptTokens
	outputTokens := usage.CompletionTokens

	totalTokens := usage.TotalTokens
	if totalTokens <= 0 {
		totalTokens = inputTokens + outputTokens
	}

	if inputTokens == 0 && outputTokens == 0 && totalTokens > 0 {
		if estimated.InputTokens > 0 && totalTokens > estimated.InputTokens {
			inputTokens = estimated.InputTokens
			outputTokens = totalTokens - inputTokens
		} else {
			outputTokens = totalTokens
		}
	}

	if totalTokens < inputTokens+outputTokens {
		totalTokens = inputTokens + outputTokens
	}

	return inputTokens, outputTokens, totalTokens
}

func imageCostBreakdown(
	model string,
	usage *openai.Usage,
	inputTokens int,
	outputTokens int,
	totalTokens int,
	price priceInfo,
	exchangeRate float64,
	metadata map[string]interface{},
) (platform.UsageCostBreakdown, error) {
	if price.Provider == "google-gemini" {
		return googleGeminiImageCostBreakdown(
			model,
			inputTokens,
			outputTokens,
			totalTokens,
			price,
			exchangeRate,
			metadata,
		)
	}

	if price.Provider == "openai" {
		return openAIImageCostBreakdown(
			model,
			usage,
			inputTokens,
			outputTokens,
			totalTokens,
			price,
			exchangeRate,
			metadata,
			"standard",
		)
	}

	if price.Provider == "xai" {
		return xAIImageCostBreakdown(model, inputTokens, outputTokens, totalTokens, price, exchangeRate, metadata)
	}

	out := costBreakdown(model, inputTokens, outputTokens, price, exchangeRate)
	if totalTokens > 0 {
		out.TotalTokens = totalTokens
	}

	return out, nil
}

func xAIImageCostBreakdown(
	model string,
	inputTokens int,
	outputTokens int,
	totalTokens int,
	price priceInfo,
	exchangeRate float64,
	metadata map[string]interface{},
) (platform.UsageCostBreakdown, error) {
	inputImages := intFromBillingMetadata(metadata["image_count"])

	outputImages := intFromBillingMetadata(metadata["n"])
	if outputImages <= 0 {
		outputImages = 1
	}

	resolution := strings.ToLower(strings.TrimSpace(stringFromBillingMetadata(metadata["resolution"])))
	if resolution == "" {
		resolution = "1k"
	}

	components := make([]platform.UsageCostComponent, 0, 2)
	inputExact := decimal.Zero

	if inputImages > 0 {
		rate, ok := selectXAIImageRate(price, "input_images", "")
		if !ok {
			return platform.UsageCostBreakdown{}, fmt.Errorf("%w: missing input image rate", ErrPricingUnavailable)
		}

		inputExact = rate.UnitPriceUSD.Mul(decimal.NewFromInt(int64(inputImages)))
		components = append(components, usageCostComponent("input_images", int64(inputImages), rate))
	}

	outputRate, ok := selectXAIImageRate(price, "output_images", resolution)
	if !ok {
		return platform.UsageCostBreakdown{}, fmt.Errorf("%w: missing %s output image rate", ErrPricingUnavailable, resolution)
	}

	outputExact := outputRate.UnitPriceUSD.Mul(decimal.NewFromInt(int64(outputImages)))
	components = append(components, usageCostComponent("output_images", int64(outputImages), outputRate))
	inputUSD, _ := inputExact.Float64()
	outputUSD, _ := outputExact.Float64()
	totalUSD := inputUSD + outputUSD

	return platform.UsageCostBreakdown{
		Model: model, Metadata: cloneUsageMetadata(metadata),
		InputTokens: inputTokens, OutputTokens: outputTokens, TotalTokens: totalTokens,
		InputCostUSD: inputUSD, OutputCostUSD: outputUSD, TotalCostUSD: totalUSD,
		ExchangeRate: exchangeRate, TotalCostIRR: totalUSD * exchangeRate,
		PricingRevision: price.Revision, ModelRuntimeProfile: price.RuntimeProfile,
		Components: components, PricingSnapshot: usagePricingSnapshot(price),
		Quantities: map[string]string{"input_images": strconv.Itoa(inputImages), "output_images": strconv.Itoa(outputImages)},
	}, nil
}

func selectXAIImageRate(price priceInfo, dimension, resolution string) (normalizedPricingRate, bool) {
	for _, rate := range price.RateList {
		if rate.Dimension != dimension {
			continue
		}

		conditionResolution := strings.ToLower(strings.TrimSpace(fmt.Sprint(rate.Condition["resolution"])))
		if resolution == "" && conditionResolution == "" {
			return rate, true
		}

		if resolution != "" && conditionResolution == resolution {
			return rate, true
		}
	}

	rate, ok := price.Rates[dimension]

	return rate, ok && (resolution == "" || len(rate.Condition) == 0)
}

func reconcileXAIProviderCost(breakdown platform.UsageCostBreakdown, ticks int64, exchangeRate float64) (platform.UsageCostBreakdown, error) {
	if ticks <= 0 {
		return platform.UsageCostBreakdown{}, fmt.Errorf("%w: xAI response omitted cost_in_usd_ticks", ErrPricingUnavailable)
	}

	catalogTotalUSD := breakdown.TotalCostUSD
	exactUSD := decimal.NewFromInt(ticks).Div(decimal.NewFromInt(int64(usdTicksPerDollar)))
	totalUSD, _ := exactUSD.Float64()
	delta := exactUSD.Sub(decimal.NewFromFloat(catalogTotalUSD))

	breakdown = applyProviderImageCost(breakdown, totalUSD)
	if !delta.IsZero() {
		breakdown.Components = append(breakdown.Components, platform.UsageCostComponent{
			Dimension:    "provider_cost_reconciliation",
			Quantity:     "1",
			Unit:         "request",
			UnitPriceUSD: delta.String(),
			CostUSD:      delta.String(),
			Condition:    map[string]interface{}{"source": "cost_in_usd_ticks"},
		})
	}

	breakdown.ExchangeRate = exchangeRate

	breakdown.TotalCostIRR = totalUSD * exchangeRate
	if breakdown.Metadata == nil {
		breakdown.Metadata = map[string]interface{}{}
	}

	breakdown.Metadata["provider_cost_in_usd_ticks"] = ticks
	breakdown.Metadata["provider_cost_usd"] = exactUSD.String()
	breakdown.Metadata["catalog_estimated_cost_usd"] = decimal.NewFromFloat(catalogTotalUSD).String()
	breakdown.Metadata["provider_cost_reconciliation_delta_usd"] = delta.String()
	breakdown.Metadata["provider_cost_exact"] = true

	return breakdown, nil
}

func cloneUsageMetadata(input map[string]interface{}) map[string]interface{} {
	output := make(map[string]interface{}, len(input)+2)
	for key, value := range input {
		output[key] = value
	}

	return output
}

func stringFromBillingMetadata(value interface{}) string {
	text, _ := value.(string) //nolint:errcheck // Optional metadata uses the zero value when absent or mismatched.
	return text
}

func openAIImageCostBreakdown(
	model string,
	usage *openai.Usage,
	inputTokens int,
	outputTokens int,
	totalTokens int,
	price priceInfo,
	exchangeRate float64,
	metadata map[string]interface{},
	serviceTier string,
) (platform.UsageCostBreakdown, error) {
	textInputTokens := inputTokens
	imageInputTokens := 0
	cachedTextTokens := 0
	cachedImageTokens := 0

	if usage != nil && usage.PromptTokensDetails != nil {
		details := usage.PromptTokensDetails
		textInputTokens = details.TextTokens
		imageInputTokens = details.ImageTokens
		cachedTextTokens = details.CachedTextTokens

		cachedImageTokens = details.CachedImageTokens
		if cachedTextTokens+cachedImageTokens == 0 && details.CachedTokens > 0 {
			if imageInputTokens > 0 && textInputTokens == 0 {
				cachedImageTokens = details.CachedTokens
			} else {
				cachedTextTokens = details.CachedTokens
			}
		}
	}

	if imageInputTokens == 0 && intFromBillingMetadata(metadata["image_count"]) > 0 {
		estimatedImageTokens := intFromBillingMetadata(metadata["estimated_image_input_tokens"])
		if estimatedImageTokens > inputTokens {
			estimatedImageTokens = inputTokens
		}

		imageInputTokens = estimatedImageTokens
		textInputTokens = inputTokens - imageInputTokens
	}

	if textInputTokens+imageInputTokens == 0 {
		textInputTokens = inputTokens
	}

	quantities := map[string]int64{
		"input_tokens":            int64(max(textInputTokens-cachedTextTokens, 0)),
		"image_input_tokens":      int64(max(imageInputTokens-cachedImageTokens, 0)),
		"cache_read_tokens":       int64(cachedTextTokens),
		"image_cache_read_tokens": int64(cachedImageTokens),
		"image_output_tokens":     int64(outputTokens),
	}
	if strings.EqualFold(serviceTier, "batch") {
		quantities = map[string]int64{
			"batch_input_tokens":            int64(max(textInputTokens-cachedTextTokens, 0)),
			"batch_image_input_tokens":      int64(max(imageInputTokens-cachedImageTokens, 0)),
			"batch_cache_read_tokens":       int64(cachedTextTokens),
			"batch_image_cache_read_tokens": int64(cachedImageTokens),
			"batch_image_output_tokens":     int64(outputTokens),
		}
	}

	dimensions := make([]string, 0, len(quantities))
	for dimension, quantity := range quantities {
		if quantity > 0 {
			dimensions = append(dimensions, dimension)
		}
	}

	sort.Strings(dimensions)

	totalExact, inputExact, outputExact := decimal.Zero, decimal.Zero, decimal.Zero

	components := make([]platform.UsageCostComponent, 0, len(dimensions))
	for _, dimension := range dimensions {
		rate, ok := price.Rates[dimension]
		if !ok {
			fallback := dimension
			switch dimension {
			case "image_input_tokens", "batch_image_input_tokens":
				fallback = "input_tokens"
			case "image_cache_read_tokens", "batch_image_cache_read_tokens":
				fallback = "cache_read_tokens"
			case "image_output_tokens", "batch_image_output_tokens":
				fallback = "output_tokens"
			}

			rate, ok = price.Rates[fallback]
		}

		if !ok {
			return platform.UsageCostBreakdown{}, fmt.Errorf(
				"%w: missing OpenAI image rate for %s",
				ErrPricingUnavailable,
				dimension,
			)
		}

		quantity := quantities[dimension]
		cost := rate.UnitPriceUSD.Mul(decimal.NewFromInt(quantity))

		totalExact = totalExact.Add(cost)
		if dimension == "image_output_tokens" || dimension == "batch_image_output_tokens" {
			outputExact = outputExact.Add(cost)
		} else {
			inputExact = inputExact.Add(cost)
		}

		components = append(
			components,
			usageCostComponent(dimension, quantity, rate),
		)
	}

	inputUSD, _ := inputExact.Float64()
	outputUSD, _ := outputExact.Float64()
	totalUSD, _ := totalExact.Float64()

	if totalTokens <= 0 {
		totalTokens = inputTokens + outputTokens
	}

	return platform.UsageCostBreakdown{
		Model: model, InputTokens: inputTokens, OutputTokens: outputTokens,
		TotalTokens: totalTokens, InputCostUSD: inputUSD,
		OutputCostUSD: outputUSD, TotalCostUSD: totalUSD,
		ExchangeRate: exchangeRate, TotalCostIRR: totalUSD * exchangeRate,
		ServiceTier:     serviceTier,
		PricingRevision: price.Revision, ModelRuntimeProfile: price.RuntimeProfile,
		Components: components, PricingSnapshot: usagePricingSnapshot(price),
	}, nil
}

func intFromBillingMetadata(value interface{}) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	case json.Number:
		parsed, _ := typed.Int64() //nolint:errcheck // Failure is intentionally handled by the surrounding best-effort path.
		return int(parsed)
	default:
		return 0
	}
}

func googleGeminiImageCostBreakdown(
	model string,
	inputTokens int,
	outputTokens int,
	totalTokens int,
	price priceInfo,
	exchangeRate float64,
	metadata map[string]interface{},
) (platform.UsageCostBreakdown, error) {
	quantities := make(map[string]int64)

	if inputTokens > 0 {
		if _, ok := price.Rates["input_tokens"]; ok {
			quantities["input_tokens"] = int64(inputTokens)
		}
	}

	if outputTokens > 0 {
		if _, ok := price.Rates["image_output_tokens"]; ok {
			quantities["image_output_tokens"] = int64(outputTokens)
		} else if _, ok := price.Rates["output_tokens"]; ok {
			quantities["output_tokens"] = int64(outputTokens)
		}
	}

	imageCount := int64(1)

	if raw, ok := metadata["n"]; ok {
		switch value := raw.(type) {
		case int:
			imageCount = int64(value)
		case float64:
			imageCount = int64(value)
		}
	}

	if imageCount < 1 {
		imageCount = 1
	}

	if _, hasTokenRate := quantities["image_output_tokens"]; !hasTokenRate {
		dimension := googleGeminiImageDimension(price, metadata)
		if dimension != "" {
			quantities[dimension] = imageCount
			delete(quantities, "output_tokens")
		}
	}

	if len(quantities) == 0 {
		return platform.UsageCostBreakdown{}, fmt.Errorf(
			"%w: Gemini image pricing is unavailable",
			ErrPricingUnavailable,
		)
	}

	dimensions := make([]string, 0, len(quantities))
	for dimension := range quantities {
		dimensions = append(dimensions, dimension)
	}

	sort.Strings(dimensions)

	totalExact := decimal.Zero
	inputExact := decimal.Zero
	outputExact := decimal.Zero

	components := make([]platform.UsageCostComponent, 0, len(dimensions))
	for _, dimension := range dimensions {
		quantity := quantities[dimension]

		rate, ok := price.Rates[dimension]
		if !ok {
			return platform.UsageCostBreakdown{}, fmt.Errorf(
				"%w: missing Gemini image rate for %s",
				ErrPricingUnavailable,
				dimension,
			)
		}

		cost := rate.UnitPriceUSD.Mul(decimal.NewFromInt(quantity))

		totalExact = totalExact.Add(cost)
		if dimension == "input_tokens" || strings.HasPrefix(dimension, "input_") {
			inputExact = inputExact.Add(cost)
		} else {
			outputExact = outputExact.Add(cost)
		}

		components = append(
			components,
			usageCostComponent(dimension, quantity, rate),
		)
	}

	inputUSD, _ := inputExact.Float64()
	outputUSD, _ := outputExact.Float64()
	totalUSD, _ := totalExact.Float64()

	if totalTokens <= 0 {
		totalTokens = inputTokens + outputTokens
	}

	return platform.UsageCostBreakdown{
		Model:               model,
		InputTokens:         inputTokens,
		OutputTokens:        outputTokens,
		TotalTokens:         totalTokens,
		InputCostUSD:        inputUSD,
		OutputCostUSD:       outputUSD,
		TotalCostUSD:        totalUSD,
		ExchangeRate:        exchangeRate,
		TotalCostIRR:        totalUSD * exchangeRate,
		PricingRevision:     price.Revision,
		ModelRuntimeProfile: price.RuntimeProfile,
		Components:          components,
		PricingSnapshot:     usagePricingSnapshot(price),
	}, nil
}

func googleGeminiImageDimension(
	price priceInfo,
	metadata map[string]interface{},
) string {
	size, _ := metadata["size"].(string) //nolint:errcheck // Optional metadata uses the zero value when absent or mismatched.

	size = strings.ToLower(strings.TrimSpace(size))
	switch {
	case strings.Contains(size, "4096") || strings.Contains(size, "4k"):
		if _, ok := price.Rates["output_images_4k"]; ok {
			return "output_images_4k"
		}
	case strings.Contains(size, "2048") || strings.Contains(size, "2k"):
		if _, ok := price.Rates["output_images_2k"]; ok {
			return "output_images_2k"
		}

		if _, ok := price.Rates["output_images_1k_2k"]; ok {
			return "output_images_1k_2k"
		}
	case strings.Contains(size, "512") || strings.Contains(size, "0.5k"):
		if _, ok := price.Rates["output_images_0_5k"]; ok {
			return "output_images_0_5k"
		}
	default:
		if _, ok := price.Rates["output_images_1k"]; ok {
			return "output_images_1k"
		}

		if _, ok := price.Rates["output_images_1k_2k"]; ok {
			return "output_images_1k_2k"
		}
	}

	if _, ok := price.Rates["output_images"]; ok {
		return "output_images"
	}

	return ""
}

func applyProviderImageCost(breakdown platform.UsageCostBreakdown, totalUSD float64) platform.UsageCostBreakdown {
	breakdown.TotalCostUSD = totalUSD
	if breakdown.InputCostUSD+breakdown.OutputCostUSD > 0 {
		pricedTotal := breakdown.InputCostUSD + breakdown.OutputCostUSD
		breakdown.InputCostUSD = totalUSD * (breakdown.InputCostUSD / pricedTotal)
		breakdown.OutputCostUSD = totalUSD - breakdown.InputCostUSD

		return breakdown
	}

	breakdown.InputCostUSD = 0
	breakdown.OutputCostUSD = totalUSD

	return breakdown
}

func (s *Service) modelPrice(
	ctx context.Context,
	model string,
) (priceInfo, error) {
	if isAnthropicModel(model) {
		if s == nil || s.anthropicPricing == nil {
			return priceInfo{}, ErrPricingUnavailable
		}

		item, snapshot, err := s.anthropicPricing.price(model)
		if err != nil {
			return priceInfo{}, err
		}

		return priceInfo{
			Input:          decimalToPerThousand(item.Rates["input_tokens"].UnitPriceUSD),
			Output:         decimalToPerThousand(item.Rates["output_tokens"].UnitPriceUSD),
			InputPerToken:  item.Rates["input_tokens"].UnitPriceUSD,
			OutputPerToken: item.Rates["output_tokens"].UnitPriceUSD,
			Rates:          item.Rates,
			RateList:       item.RateList,
			Revision:       snapshot.Revision,
			GeneratedAt:    snapshot.GeneratedAt,
			Provider:       "anthropic",
			RuntimeProfile: item.RuntimeProfile,
		}, nil
	}

	if provider := catalogPricingProvider(model); provider != "" {
		if s == nil || s.catalogPricing == nil || s.catalogPricing[provider] == nil {
			return priceInfo{}, ErrPricingUnavailable
		}

		item, snapshot, err := s.catalogPricing[provider].price(model)
		if err != nil {
			return priceInfo{}, err
		}

		if !catalogPricingModelStatusBillable(item.Status) {
			return priceInfo{}, fmt.Errorf(
				"%w: model lifecycle is %s",
				ErrPricingModelUnavailable,
				item.Status,
			)
		}

		inputRate := item.Rates["input_tokens"].UnitPriceUSD
		outputRate := item.Rates["output_tokens"].UnitPriceUSD

		return priceInfo{
			Input:          decimalToPerThousand(inputRate),
			Output:         decimalToPerThousand(outputRate),
			InputPerToken:  inputRate,
			OutputPerToken: outputRate,
			Rates:          item.Rates,
			RateList:       item.RateList,
			Revision:       snapshot.Revision,
			GeneratedAt:    snapshot.GeneratedAt,
			Provider:       provider,
			RuntimeProfile: item.RuntimeProfile,
		}, nil
	}

	if s != nil && s.pricingCache != nil {
		return s.pricingCache.Get(ctx, model), nil
	}

	return modelPriceFromRegistry(model), nil
}

func catalogPricingModelStatusBillable(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "active", "preview":
		return true
	default:
		return false
	}
}

func (s *Service) pinnedOrCurrentModelPrice(
	ctx context.Context,
	model string,
	estimated platform.UsageCostBreakdown,
) (priceInfo, error) {
	if (!isAnthropicModel(model) && catalogPricingProvider(model) == "") ||
		estimated.PricingSnapshot == nil {
		return s.modelPrice(ctx, model)
	}

	return priceInfoFromUsageSnapshot(estimated.PricingSnapshot)
}

func catalogPricingProvider(model string) string {
	value := strings.ToLower(strings.TrimSpace(model))
	switch {
	case strings.HasPrefix(value, "google-gemini/"),
		strings.HasPrefix(value, "google-gemini:"),
		strings.HasPrefix(value, "gemini/"),
		strings.HasPrefix(value, "gemini:"):
		return "google-gemini"
	case strings.HasPrefix(value, "openai/"),
		strings.HasPrefix(value, "openai:"):
		return "openai"
	case strings.HasPrefix(value, "xai/"),
		strings.HasPrefix(value, "xai:"):
		return "xai"
	case strings.HasPrefix(value, "moonshot/"),
		strings.HasPrefix(value, "moonshot:"):
		return "moonshot"
	case strings.HasPrefix(value, "deepseek/"),
		strings.HasPrefix(value, "deepseek:"):
		return "deepseek"
	case strings.HasPrefix(value, "mistral-ai/"),
		strings.HasPrefix(value, "mistral-ai:"),
		strings.HasPrefix(value, "mistral/"),
		strings.HasPrefix(value, "mistral:"),
		strings.HasPrefix(value, "mistral-"),
		strings.HasPrefix(value, "codestral-"),
		strings.HasPrefix(value, "voxtral-"):
		return "mistral-ai"
	default:
		return ""
	}
}

func decimalToPerThousand(value decimal.Decimal) float64 {
	out, _ := value.Mul(decimal.NewFromInt(1000)).Float64()
	return out
}

func usageCostComponent(
	dimension string,
	quantity int64,
	rate normalizedPricingRate,
) platform.UsageCostComponent {
	return usageCostComponentDecimal(
		dimension,
		decimal.NewFromInt(quantity),
		rate,
	)
}

func usageCostComponentDecimal(
	dimension string,
	quantity decimal.Decimal,
	rate normalizedPricingRate,
) platform.UsageCostComponent {
	cost := rate.UnitPriceUSD.Mul(quantity)

	return platform.UsageCostComponent{
		Dimension:    dimension,
		Quantity:     quantity.String(),
		Unit:         rate.Unit,
		UnitPriceUSD: rate.UnitPriceUSD.String(),
		CostUSD:      cost.String(),
		Condition:    cloneInterfaceMap(rate.Condition),
	}
}

func usagePricingSnapshot(price priceInfo) *platform.UsagePricingSnapshot {
	rates := make(map[string]platform.UsageCostComponent, len(price.Rates))
	for dimension, rate := range price.Rates {
		rates[dimension] = usageCostComponent(dimension, 0, rate)
	}

	rateList := make([]platform.UsageCostComponent, 0, len(price.RateList))
	for _, rate := range price.RateList {
		rateList = append(rateList, usageCostComponent(rate.Dimension, 0, rate))
	}

	return &platform.UsagePricingSnapshot{
		Provider:    price.Provider,
		Revision:    price.Revision,
		GeneratedAt: price.GeneratedAt,
		Rates:       rates,
		RateList:    rateList,
	}
}

func priceInfoFromUsageSnapshot(
	snapshot *platform.UsagePricingSnapshot,
) (priceInfo, error) {
	if snapshot == nil || strings.TrimSpace(snapshot.Provider) == "" ||
		snapshot.Revision == "" || snapshot.GeneratedAt.IsZero() {
		return priceInfo{}, ErrPricingUnavailable
	}

	rates := make(
		map[string]normalizedPricingRate,
		len(snapshot.Rates),
	)
	for dimension, component := range snapshot.Rates {
		value, err := decimal.NewFromString(component.UnitPriceUSD)
		if err != nil || component.Unit == "" {
			return priceInfo{}, ErrPricingUnavailable
		}

		rates[dimension] = normalizedPricingRate{
			Dimension:    dimension,
			Unit:         component.Unit,
			UnitPriceUSD: value,
			Condition:    cloneInterfaceMap(component.Condition),
		}
	}

	rateList := make([]normalizedPricingRate, 0, len(snapshot.RateList))
	for _, component := range snapshot.RateList {
		value, err := decimal.NewFromString(component.UnitPriceUSD)
		if err != nil || component.Dimension == "" || component.Unit == "" {
			return priceInfo{}, ErrPricingUnavailable
		}

		rateList = append(rateList, normalizedPricingRate{
			Dimension:    component.Dimension,
			Unit:         component.Unit,
			UnitPriceUSD: value,
			Condition:    cloneInterfaceMap(component.Condition),
		})
	}

	input := rates["input_tokens"]
	output := rates["output_tokens"]

	return priceInfo{
		Input:          decimalToPerThousand(input.UnitPriceUSD),
		Output:         decimalToPerThousand(output.UnitPriceUSD),
		InputPerToken:  input.UnitPriceUSD,
		OutputPerToken: output.UnitPriceUSD,
		Rates:          rates,
		RateList:       rateList,
		Revision:       snapshot.Revision,
		GeneratedAt:    snapshot.GeneratedAt,
		Provider:       snapshot.Provider,
	}, nil
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
