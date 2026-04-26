package billing

import (
	"context"
	"errors"
	"fmt"
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
)

type walletStore interface {
	GetWallet(ctx context.Context, tenantID string) (*platform.Wallet, error)
	GetLatestExchangeRate(ctx context.Context) (*platform.ExchangeRate, error)
	UpsertExchangeRate(ctx context.Context, rate platform.ExchangeRate) (*platform.ExchangeRate, error)
}

// Service enforces rate limits and budget checks.
type Service struct {
	cache       cache.Cache
	store       walletStore
	defaultRate platform.ExchangeRate
}

// NewService creates a billing enforcement service.
func NewService(c cache.Cache, store walletStore, defaultUSDtoIRR float64, source string) *Service {
	return &Service{
		cache: c,
		store: store,
		defaultRate: platform.ExchangeRate{
			USDtoIRR: defaultUSDtoIRR,
			Source:   source,
		},
	}
}

const tenantLimitMultiplier int64 = 5

// CheckRateLimit validates per-minute request and token limits.
// Returns the primary rate-limit result to expose in response headers.
func (s *Service) CheckRateLimit(ctx context.Context, auth *platform.AuthContext, model string, estimatedTokens int) (*cache.RateLimitResult, error) {
	window := time.Minute
	var primary *cache.RateLimitResult

	if auth.RPMLimit != nil {
		key := fmt.Sprintf("rpm:key:%s:%s", auth.TokenHash, model)
		result, err := s.cache.CheckRateLimit(ctx, key, *auth.RPMLimit, window)
		if err != nil {
			return primary, err
		}
		if primary == nil {
			primary = result
		}
		if !result.Allowed {
			return primary, ErrRateLimitExceeded
		}

		if auth.TenantID != "" {
			tenantLimit := deriveTenantLimit(*auth.RPMLimit)
			tenantKey := fmt.Sprintf("rpm:tenant:%s:%s", auth.TenantID, model)
			tenantResult, err := s.cache.CheckRateLimit(ctx, tenantKey, tenantLimit, window)
			if err != nil {
				return primary, err
			}
			if !tenantResult.Allowed {
				return tenantResult, ErrRateLimitExceeded
			}
		}
	}

	if auth.TPMLimit != nil && estimatedTokens > 0 {
		key := fmt.Sprintf("tpm:key:%s:%s", auth.TokenHash, model)
		result, err := s.cache.CheckSlidingWindowRateLimitWithWeight(ctx, key, *auth.TPMLimit, int64(estimatedTokens), window)
		if err != nil {
			return primary, err
		}
		if primary == nil {
			primary = result
		}
		if !result.Allowed {
			return primary, ErrRateLimitExceeded
		}

		if auth.TenantID != "" {
			tenantLimit := deriveTenantLimit(*auth.TPMLimit)
			tenantKey := fmt.Sprintf("tpm:tenant:%s:%s", auth.TenantID, model)
			tenantResult, err := s.cache.CheckSlidingWindowRateLimitWithWeight(ctx, tenantKey, tenantLimit, int64(estimatedTokens), window)
			if err != nil {
				return primary, err
			}
			if !tenantResult.Allowed {
				return tenantResult, ErrRateLimitExceeded
			}
		}
	}
	return primary, nil
}

// CheckBudget validates wallet and hard budget constraints.
func (s *Service) CheckBudget(ctx context.Context, auth *platform.AuthContext, breakdown platform.UsageCostBreakdown) error {
	if auth.MaxBudget != nil && auth.Spend+breakdown.TotalCostUSD > *auth.MaxBudget {
		return ErrBudgetExceeded
	}
	wallet, err := s.store.GetWallet(ctx, auth.TenantID)
	if err != nil {
		return err
	}
	if wallet.BalanceIRR < breakdown.TotalCostIRR {
		return ErrBudgetExceeded
	}
	return nil
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
	inputCost := float64(promptTokens) * price.Input
	outputCost := float64(outputTokens) * price.Output
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
	inputCost := float64(usage.PromptTokens) * price.Input
	outputCost := float64(usage.CompletionTokens) * price.Output
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

func deriveTenantLimit(perKeyLimit int64) int64 {
	if perKeyLimit <= 0 {
		return perKeyLimit
	}
	return perKeyLimit * tenantLimitMultiplier
}
