package billing

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
)

var (
	// ErrRateLimitExceeded is returned when per-key RPM/TPM is exhausted.
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
	// ErrBudgetExceeded is returned when max budget or wallet checks fail.
	ErrBudgetExceeded = errors.New("budget exceeded")
	// ErrSubscriptionInactive is returned when a tenant has no usable commercial subscription.
	ErrSubscriptionInactive = errors.New("subscription inactive")
)

type walletStore interface {
	GetWallet(ctx context.Context, tenantID string) (*platform.Wallet, error)
	GetLatestExchangeRate(ctx context.Context) (*platform.ExchangeRate, error)
	UpsertExchangeRate(ctx context.Context, rate platform.ExchangeRate) (*platform.ExchangeRate, error)
}

type entitlementStore interface {
	ResolveEffectiveEntitlements(ctx context.Context, tenantID string) (*platform.EffectiveEntitlementsResponse, error)
	CheckAndIncrementTenantUsageCounter(ctx context.Context, tenantID string, featureKey string, limit float64, increment float64, now time.Time) (*platform.UsageCounter, bool, error)
}

// Service enforces rate limits and budget checks.
type Service struct {
	cache             cache.Cache
	store             walletStore
	entitlements      entitlementStore
	defaultRate       platform.ExchangeRate
	exchangeRateCache *exchangeRateCache
	pricingCache      *pricingCache
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
		cache:        c,
		store:        store,
		entitlements: entitlements,
		defaultRate:  defaultRate,
	}
	svc.exchangeRateCache = newExchangeRateCache(store, c, defaultRate, exchangeRateCacheConfig)
	svc.pricingCache = newPricingCache(c, pricingCacheConfig, modelPriceFromRegistry)

	return svc
}

// CheckRateLimit validates per-minute request and token limits.
// Returns the primary rate-limit result to expose in response headers.
func (s *Service) CheckRateLimit(ctx context.Context, auth *platform.AuthContext, _ string, _ int) (*cache.RateLimitResult, error) {
	if s.entitlements == nil || auth == nil || auth.TenantID == "" {
		return nil, nil
	}

	effective, err := s.entitlements.ResolveEffectiveEntitlements(ctx, auth.TenantID)
	if err != nil {
		return nil, ErrSubscriptionInactive
	}

	if effective.Subscription == nil || (effective.Subscription.Status != "active" && effective.Subscription.Status != "trialing") {
		return nil, ErrSubscriptionInactive
	}

	limit, limited := entitlementNumber(effective.Features[platform.FeatureGatewayRequestsMonthly])
	if !limited {
		return nil, nil
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

	return result, nil
}

// CheckBudget validates wallet balance constraints.
func (s *Service) CheckBudget(ctx context.Context, auth *platform.AuthContext, breakdown platform.UsageCostBreakdown) error {
	wallet, err := s.store.GetWallet(ctx, auth.TenantID)
	if err != nil {
		return err
	}

	if wallet.BalanceIRR < breakdown.TotalCostIRR {
		return ErrBudgetExceeded
	}

	return nil
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
