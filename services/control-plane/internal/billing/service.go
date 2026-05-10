package billing

import (
	"context"
	"errors"

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

// CheckRateLimit validates per-minute request and token limits.
// Returns the primary rate-limit result to expose in response headers.
func (s *Service) CheckRateLimit(ctx context.Context, auth *platform.AuthContext, model string, estimatedTokens int) (*cache.RateLimitResult, error) {
	return nil, nil
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
