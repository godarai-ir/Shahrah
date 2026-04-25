package billing

import (
	"context"
	"errors"
	"strings"
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

// CheckRateLimit validates per-minute request and token limits.
func (s *Service) CheckRateLimit(ctx context.Context, auth *platform.AuthContext, model string, estimatedTokens int) error {
	window := time.Minute
	nowBucket := time.Now().UTC().Format("200601021504")
	if auth.RPMLimit != nil {
		key := "rpm:" + auth.TokenHash + ":" + model + ":" + nowBucket
		value, err := s.cache.Increment(ctx, key, 1, window)
		if err != nil {
			return err
		}
		if value > *auth.RPMLimit {
			return ErrRateLimitExceeded
		}
	}
	if auth.TPMLimit != nil && estimatedTokens > 0 {
		key := "tpm:" + auth.TokenHash + ":" + model + ":" + nowBucket
		value, err := s.cache.Increment(ctx, key, int64(estimatedTokens), window)
		if err != nil {
			return err
		}
		if value > *auth.TPMLimit {
			return ErrRateLimitExceeded
		}
	}
	return nil
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

type pricing struct {
	Input  float64
	Output float64
}

func modelPrice(model string) pricing {
	switch {
	case strings.Contains(model, "gpt-4"):
		return pricing{Input: 0.00003, Output: 0.00006}
	default:
		return pricing{Input: 0.0000005, Output: 0.0000015}
	}
}

func estimatePromptTokens(req openai.ChatCompletionRequest) int {
	totalChars := 0
	for _, msg := range req.Messages {
		totalChars += len(msg.Content) + len(msg.Role) + len(msg.Name)
	}
	if totalChars == 0 {
		return 1
	}
	return (totalChars / 4) + len(req.Messages)*4
}
