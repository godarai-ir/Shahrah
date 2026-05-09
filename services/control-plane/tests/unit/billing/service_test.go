package billing_test

import (
	"context"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/billing"
	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type walletStoreStub struct {
	wallet *platform.Wallet
	rate   *platform.ExchangeRate
}

func (s walletStoreStub) GetWallet(ctx context.Context, tenantID string) (*platform.Wallet, error) {
	return s.wallet, nil
}
func (s walletStoreStub) GetLatestExchangeRate(ctx context.Context) (*platform.ExchangeRate, error) {
	return s.rate, nil
}
func (s walletStoreStub) UpsertExchangeRate(ctx context.Context, rate platform.ExchangeRate) (*platform.ExchangeRate, error) {
	return &rate, nil
}

func TestBillingService_ShouldEstimateCost(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)
	svc := billing.NewService(memCache, walletStoreStub{
		wallet: &platform.Wallet{TenantID: "tenant-1", BalanceIRR: 1000000},
		rate:   &platform.ExchangeRate{USDtoIRR: 850000, Source: "test"},
	}, 850000, "test")

	out, err := svc.EstimateCost(context.Background(), openai.ChatCompletionRequest{
		Model: "gpt-3.5-turbo",
		Messages: []openai.ChatMessage{
			{Role: "user", Content: "hello world"},
		},
	})
	assert.NoError(t, err)
	assert.Greater(t, out.TotalTokens, 0)
	assert.Greater(t, out.TotalCostIRR, 0.0)
}

func TestBillingService_ShouldEstimateCostWithPer1KScaling(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)
	svc := billing.NewService(memCache, walletStoreStub{
		wallet: &platform.Wallet{TenantID: "tenant-1", BalanceIRR: 1_000_000},
		rate:   &platform.ExchangeRate{USDtoIRR: 850000, Source: "test"},
	}, 850000, "test")

	maxTokens := 250
	out, err := svc.EstimateCost(context.Background(), openai.ChatCompletionRequest{
		Model:     "gpt-3.5-turbo",
		MaxTokens: &maxTokens,
		Messages: []openai.ChatMessage{
			{Role: "user", Content: "hello world"},
		},
	})
	require.NoError(t, err)

	// gpt-3.5-turbo: input=$0.50/1M, output=$1.50/1M
	// modelPrice(): input=0.0005 and output=0.0015 per 1K
	// promptTokens estimator returns 100
	// inputCost=(100/1000)*0.0005=0.00005
	// outputCost=(250/1000)*0.0015=0.000375
	expectedUSD := 0.000425
	expectedIRR := expectedUSD * 850000.0

	assert.InDelta(t, expectedUSD, out.TotalCostUSD, 1e-12)
	assert.InDelta(t, expectedIRR, out.TotalCostIRR, 1e-6)
}

func TestBillingService_ShouldFinalizeCostWithPer1KScaling(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)
	svc := billing.NewService(memCache, walletStoreStub{
		wallet: &platform.Wallet{TenantID: "tenant-1", BalanceIRR: 1_000_000},
		rate:   &platform.ExchangeRate{USDtoIRR: 850000, Source: "test"},
	}, 850000, "test")

	usage := &openai.Usage{
		PromptTokens:     120,
		CompletionTokens: 80,
		TotalTokens:      200,
	}
	out, err := svc.FinalizeFromUsage(context.Background(), "gpt-3.5-turbo", usage, platform.UsageCostBreakdown{})
	require.NoError(t, err)

	// inputCost=(120/1000)*0.0005=0.00006
	// outputCost=(80/1000)*0.0015=0.00012
	expectedUSD := 0.00018
	expectedIRR := expectedUSD * 850000.0

	assert.InDelta(t, expectedUSD, out.TotalCostUSD, 1e-12)
	assert.InDelta(t, expectedIRR, out.TotalCostIRR, 1e-6)
}

func TestBillingService_ShouldRejectBudgetExceeded(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)
	maxBudget := 0.0000001
	svc := billing.NewService(memCache, walletStoreStub{
		wallet: &platform.Wallet{TenantID: "tenant-1", BalanceIRR: 1000000},
		rate:   &platform.ExchangeRate{USDtoIRR: 850000, Source: "test"},
	}, 850000, "test")

	err := svc.CheckBudget(context.Background(), &platform.AuthContext{
		TenantID:  "tenant-1",
		MaxBudget: &maxBudget,
	}, platform.UsageCostBreakdown{TotalCostUSD: 1})
	assert.ErrorIs(t, err, billing.ErrBudgetExceeded)
}

func TestBillingService_ShouldEnforcePerKeyRPM(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)
	svc := billing.NewService(memCache, walletStoreStub{
		wallet: &platform.Wallet{TenantID: "tenant-1", BalanceIRR: 1000000},
		rate:   &platform.ExchangeRate{USDtoIRR: 850000, Source: "test"},
	}, 850000, "test")

	limit := int64(2)
	auth := &platform.AuthContext{
		TenantID:  "tenant-1",
		TokenHash: "key-a",
		RPMLimit:  &limit,
	}

	first, err := svc.CheckRateLimit(context.Background(), auth, "gpt-4", 0)
	assert.NoError(t, err)
	assert.NotNil(t, first)
	assert.EqualValues(t, 1, first.Current)

	second, err := svc.CheckRateLimit(context.Background(), auth, "gpt-4", 0)
	assert.NoError(t, err)
	assert.NotNil(t, second)
	assert.EqualValues(t, 2, second.Current)

	third, err := svc.CheckRateLimit(context.Background(), auth, "gpt-4", 0)
	assert.ErrorIs(t, err, billing.ErrRateLimitExceeded)
	assert.NotNil(t, third)
	assert.False(t, third.Allowed)
}

func TestBillingService_ShouldEnforceTenantAggregateRPM(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)
	svc := billing.NewService(memCache, walletStoreStub{
		wallet: &platform.Wallet{TenantID: "tenant-1", BalanceIRR: 1000000},
		rate:   &platform.ExchangeRate{USDtoIRR: 850000, Source: "test"},
	}, 850000, "test")

	limit := int64(1)
	for _, token := range []string{"key-a", "key-b", "key-c", "key-d", "key-e"} {
		_, err := svc.CheckRateLimit(context.Background(), &platform.AuthContext{
			TenantID:  "tenant-1",
			TokenHash: token,
			RPMLimit:  &limit,
		}, "gpt-4", 0)
		assert.NoError(t, err)
	}

	// tenantLimitMultiplier=5 with RPMLimit=1 => tenant aggregate limit 5.
	tenantBlock, err := svc.CheckRateLimit(context.Background(), &platform.AuthContext{
		TenantID:  "tenant-1",
		TokenHash: "key-f",
		RPMLimit:  &limit,
	}, "gpt-4", 0)
	assert.ErrorIs(t, err, billing.ErrRateLimitExceeded)
	assert.NotNil(t, tenantBlock)
	assert.False(t, tenantBlock.Allowed)
}

func TestBillingService_ShouldEnforceSlidingTPM(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)
	svc := billing.NewService(memCache, walletStoreStub{
		wallet: &platform.Wallet{TenantID: "tenant-1", BalanceIRR: 1000000},
		rate:   &platform.ExchangeRate{USDtoIRR: 850000, Source: "test"},
	}, 850000, "test")

	limit := int64(100)
	auth := &platform.AuthContext{
		TenantID:  "tenant-1",
		TokenHash: "key-a",
		TPMLimit:  &limit,
	}

	first, err := svc.CheckRateLimit(context.Background(), auth, "gpt-4", 60)
	assert.NoError(t, err)
	assert.NotNil(t, first)
	assert.True(t, first.Allowed)

	second, err := svc.CheckRateLimit(context.Background(), auth, "gpt-4", 50)
	assert.ErrorIs(t, err, billing.ErrRateLimitExceeded)
	assert.NotNil(t, second)
	assert.False(t, second.Allowed)

}
