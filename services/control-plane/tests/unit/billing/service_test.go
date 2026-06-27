package billing_test

import (
	"context"
	"testing"
	"time"

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

type entitlementWalletStoreStub struct {
	walletStoreStub
	effective *platform.EffectiveEntitlementsResponse
	used      float64
}

func (s *entitlementWalletStoreStub) ResolveEffectiveEntitlements(ctx context.Context, tenantID string) (*platform.EffectiveEntitlementsResponse, error) {
	return s.effective, nil
}

func (s *entitlementWalletStoreStub) CheckAndIncrementTenantUsageCounter(ctx context.Context, tenantID string, featureKey string, limit float64, increment float64, now time.Time) (*platform.UsageCounter, bool, error) {
	allowed := s.used+increment <= limit
	if allowed {
		s.used += increment
	}

	return &platform.UsageCounter{
		FeatureKey:  featureKey,
		WindowStart: now,
		WindowEnd:   now.Add(30 * 24 * time.Hour),
		Used:        s.used,
		Limit:       &limit,
	}, allowed, nil
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

func TestBillingService_ShouldEstimateEmbeddingCost(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)
	svc := billing.NewService(memCache, walletStoreStub{
		wallet: &platform.Wallet{TenantID: "tenant-1", BalanceIRR: 1_000_000},
		rate:   &platform.ExchangeRate{USDtoIRR: 850000, Source: "test"},
	}, 850000, "test")

	out, err := svc.EstimateEmbeddingCost(context.Background(), openai.EmbeddingRequest{
		Model: "text-embedding-3-small",
		Input: []string{"hello", "world"},
	})
	require.NoError(t, err)
	assert.Greater(t, out.TotalTokens, 0)
	assert.Equal(t, 0, out.OutputTokens)
	assert.Greater(t, out.TotalCostIRR, 0.0)
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

func TestBillingService_ShouldRejectWhenWalletBalanceIsInsufficient(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)
	svc := billing.NewService(memCache, walletStoreStub{
		wallet: &platform.Wallet{TenantID: "tenant-1", BalanceIRR: 10},
		rate:   &platform.ExchangeRate{USDtoIRR: 850000, Source: "test"},
	}, 850000, "test")

	err := svc.CheckBudget(context.Background(), &platform.AuthContext{
		TenantID: "tenant-1",
	}, platform.UsageCostBreakdown{TotalCostIRR: 100})
	assert.ErrorIs(t, err, billing.ErrBudgetExceeded)
}

func TestBillingService_ShouldAllowWhenWalletBalanceIsSufficient(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)
	svc := billing.NewService(memCache, walletStoreStub{
		wallet: &platform.Wallet{TenantID: "tenant-1", BalanceIRR: 1000000},
		rate:   &platform.ExchangeRate{USDtoIRR: 850000, Source: "test"},
	}, 850000, "test")

	err := svc.CheckBudget(context.Background(), &platform.AuthContext{
		TenantID: "tenant-1",
	}, platform.UsageCostBreakdown{TotalCostIRR: 100})
	assert.NoError(t, err)
}

func TestBillingService_ShouldSkipLegacyRateLimits(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)
	svc := billing.NewService(memCache, walletStoreStub{
		wallet: &platform.Wallet{TenantID: "tenant-1", BalanceIRR: 1000000},
		rate:   &platform.ExchangeRate{USDtoIRR: 850000, Source: "test"},
	}, 850000, "test")

	for i := 0; i < 10; i++ {
		res, err := svc.CheckRateLimit(context.Background(), &platform.AuthContext{
			TenantID:  "tenant-1",
			TokenHash: "key-a",
		}, "gpt-4", 500)
		assert.NoError(t, err)
		assert.Nil(t, res)
	}
}

func TestBillingService_ShouldAllowTrialWhenSubscriptionQuotaRemains(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)
	store := &entitlementWalletStoreStub{
		walletStoreStub: walletStoreStub{
			wallet: &platform.Wallet{TenantID: "tenant-1", BalanceIRR: 1000000},
			rate:   &platform.ExchangeRate{USDtoIRR: 850000, Source: "test"},
		},
		used: 4,
		effective: &platform.EffectiveEntitlementsResponse{
			Subscription: &platform.TenantSubscription{Status: "trialing"},
			Features: map[string]interface{}{
				platform.FeatureGatewayRequestsMonthly: float64(5),
			},
		},
	}
	svc := billing.NewService(memCache, store, 850000, "test")

	res, err := svc.CheckRateLimit(context.Background(), &platform.AuthContext{TenantID: "tenant-1"}, "gpt-4", 500)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.True(t, res.Allowed)
	assert.Equal(t, int64(5), res.Current)
	assert.Equal(t, int64(5), res.Limit)
}

func TestBillingService_ShouldRejectWhenSubscriptionQuotaIsConsumed(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)
	store := &entitlementWalletStoreStub{
		walletStoreStub: walletStoreStub{
			wallet: &platform.Wallet{TenantID: "tenant-1", BalanceIRR: 1000000},
			rate:   &platform.ExchangeRate{USDtoIRR: 850000, Source: "test"},
		},
		used: 5,
		effective: &platform.EffectiveEntitlementsResponse{
			Subscription: &platform.TenantSubscription{Status: "active"},
			Features: map[string]interface{}{
				platform.FeatureGatewayRequestsMonthly: float64(5),
			},
		},
	}
	svc := billing.NewService(memCache, store, 850000, "test")

	res, err := svc.CheckRateLimit(context.Background(), &platform.AuthContext{TenantID: "tenant-1"}, "gpt-4", 500)
	assert.ErrorIs(t, err, billing.ErrRateLimitExceeded)
	require.NotNil(t, res)
	assert.False(t, res.Allowed)
	assert.Equal(t, int64(5), res.Current)
}

func TestBillingService_ShouldRejectInactiveSubscription(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)
	store := &entitlementWalletStoreStub{
		walletStoreStub: walletStoreStub{
			wallet: &platform.Wallet{TenantID: "tenant-1", BalanceIRR: 1000000},
			rate:   &platform.ExchangeRate{USDtoIRR: 850000, Source: "test"},
		},
		effective: &platform.EffectiveEntitlementsResponse{
			Subscription: &platform.TenantSubscription{Status: "canceled"},
			Features:     map[string]interface{}{},
		},
	}
	svc := billing.NewService(memCache, store, 850000, "test")

	res, err := svc.CheckRateLimit(context.Background(), &platform.AuthContext{TenantID: "tenant-1"}, "gpt-4", 500)
	assert.ErrorIs(t, err, billing.ErrSubscriptionInactive)
	assert.Nil(t, res)
}
