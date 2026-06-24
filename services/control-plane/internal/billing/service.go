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
	cache        cache.Cache
	store        walletStore
	entitlements entitlementStore
	defaultRate  platform.ExchangeRate
}

// NewService creates a billing enforcement service.
func NewService(c cache.Cache, store walletStore, defaultUSDtoIRR float64, source string) *Service {
	entitlements, _ := store.(entitlementStore)
	return &Service{
		cache:        c,
		store:        store,
		entitlements: entitlements,
		defaultRate: platform.ExchangeRate{
			USDtoIRR: defaultUSDtoIRR,
			Source:   source,
		},
	}
}

// CheckRateLimit validates per-minute request and token limits.
// Returns the primary rate-limit result to expose in response headers.
func (s *Service) CheckRateLimit(ctx context.Context, auth *platform.AuthContext, model string, estimatedTokens int) (*cache.RateLimitResult, error) {
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
	rate, err := s.store.GetLatestExchangeRate(ctx)
	if err != nil {
		rate = &s.defaultRate
		_, _ = s.store.UpsertExchangeRate(ctx, s.defaultRate)
	}

	promptTokens := estimatePromptTokens(req)
	outputTokens := 256
	if req.MaxTokens != nil && *req.MaxTokens > 0 {
		outputTokens = *req.MaxTokens
	}

	price := modelPrice(req.Model)
	inputCost := (float64(promptTokens) / 1000.0) * price.Input
	outputCost := (float64(outputTokens) / 1000.0) * price.Output
	totalUSD := inputCost + outputCost

	return platform.UsageCostBreakdown{
		Model:         req.Model,
		InputTokens:   promptTokens,
		OutputTokens:  outputTokens,
		TotalTokens:   promptTokens + outputTokens,
		InputCostUSD:  inputCost,
		OutputCostUSD: outputCost,
		TotalCostUSD:  totalUSD,
		ExchangeRate:  rate.USDtoIRR,
		TotalCostIRR:  totalUSD * rate.USDtoIRR,
	}, nil
}

// EstimateEmbeddingCost estimates cost for an embeddings request.
func (s *Service) EstimateEmbeddingCost(ctx context.Context, req openai.EmbeddingRequest) (platform.UsageCostBreakdown, error) {
	rate, err := s.store.GetLatestExchangeRate(ctx)
	if err != nil {
		rate = &s.defaultRate
		_, _ = s.store.UpsertExchangeRate(ctx, s.defaultRate)
	}

	promptTokens := estimateEmbeddingPromptTokens(req.Input)
	price := modelPrice(req.Model)
	inputCost := (float64(promptTokens) / 1000.0) * price.Input

	return platform.UsageCostBreakdown{
		Model:         req.Model,
		InputTokens:   promptTokens,
		OutputTokens:  0,
		TotalTokens:   promptTokens,
		InputCostUSD:  inputCost,
		OutputCostUSD: 0,
		TotalCostUSD:  inputCost,
		ExchangeRate:  rate.USDtoIRR,
		TotalCostIRR:  inputCost * rate.USDtoIRR,
	}, nil
}

// FinalizeFromUsage recalculates final spend from provider usage.
func (s *Service) FinalizeFromUsage(ctx context.Context, model string, usage *openai.Usage, estimated platform.UsageCostBreakdown) (platform.UsageCostBreakdown, error) {
	if usage == nil {
		return estimated, nil
	}

	rate, err := s.store.GetLatestExchangeRate(ctx)
	if err != nil {
		rate = &s.defaultRate
	}

	price := modelPrice(model)
	inputCost := (float64(usage.PromptTokens) / 1000.0) * price.Input
	outputCost := (float64(usage.CompletionTokens) / 1000.0) * price.Output
	totalUSD := inputCost + outputCost

	return platform.UsageCostBreakdown{
		Model:         model,
		InputTokens:   usage.PromptTokens,
		OutputTokens:  usage.CompletionTokens,
		TotalTokens:   usage.TotalTokens,
		InputCostUSD:  inputCost,
		OutputCostUSD: outputCost,
		TotalCostUSD:  totalUSD,
		ExchangeRate:  rate.USDtoIRR,
		TotalCostIRR:  totalUSD * rate.USDtoIRR,
	}, nil
}
