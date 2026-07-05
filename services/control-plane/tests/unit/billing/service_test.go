package billing_test

import (
	"context"
	"errors"
	"sync"
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
	mu           sync.Mutex
	wallet       *platform.Wallet
	rate         *platform.ExchangeRate
	rates        []*platform.ExchangeRate
	rateErr      error
	rateCalls    int
	upsertCalls  int
	upsertedRate *platform.ExchangeRate
}

func (s *walletStoreStub) GetWallet(_ context.Context, tenantID string) (*platform.Wallet, error) {
	return s.wallet, nil
}
func (s *walletStoreStub) GetLatestExchangeRate(_ context.Context) (*platform.ExchangeRate, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.rateCalls++
	if s.rateErr != nil {
		return nil, s.rateErr
	}
	if len(s.rates) > 0 {
		rate := s.rates[0]
		if len(s.rates) > 1 {
			s.rates = s.rates[1:]
		}
		return rate, nil
	}

	return s.rate, nil
}
func (s *walletStoreStub) UpsertExchangeRate(_ context.Context, rate platform.ExchangeRate) (*platform.ExchangeRate, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.upsertCalls++
	s.upsertedRate = &rate
	return &rate, nil
}

func (s *walletStoreStub) rateCallCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.rateCalls
}

type entitlementWalletStoreStub struct {
	walletStoreStub
	effective *platform.EffectiveEntitlementsResponse
	used      float64
}

func (s *entitlementWalletStoreStub) ResolveEffectiveEntitlements(_ context.Context, tenantID string) (*platform.EffectiveEntitlementsResponse, error) {
	return s.effective, nil
}

func (s *entitlementWalletStoreStub) CheckAndIncrementTenantUsageCounter(_ context.Context, tenantID, featureKey string, limit, increment float64, now time.Time) (*platform.UsageCounter, bool, error) {
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
	svc := billing.NewService(memCache, &walletStoreStub{
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
	svc := billing.NewService(memCache, &walletStoreStub{
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
	svc := billing.NewService(memCache, &walletStoreStub{
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

func TestBillingService_ShouldCacheExchangeRateAcrossCostEstimations(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)
	store := &walletStoreStub{
		wallet: &platform.Wallet{TenantID: "tenant-1", BalanceIRR: 1_000_000},
		rate:   &platform.ExchangeRate{USDtoIRR: 850000, Source: "db"},
	}
	svc := billing.NewServiceWithExchangeRateCacheConfig(memCache, store, 800000, "default", billing.ExchangeRateCacheConfig{
		FreshTTL: time.Minute,
		MaxStale: time.Hour,
	})

	for i := 0; i < 3; i++ {
		out, err := svc.EstimateCost(context.Background(), openai.ChatCompletionRequest{
			Model: "gpt-3.5-turbo",
			Messages: []openai.ChatMessage{
				{Role: "user", Content: "hello world"},
			},
		})
		require.NoError(t, err)
		assert.Equal(t, 850000.0, out.ExchangeRate)
	}

	assert.Equal(t, 1, store.rateCallCount())
}

func TestBillingService_ShouldUseStaleExchangeRateAndRefreshAsync(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)
	store := &walletStoreStub{
		wallet: &platform.Wallet{TenantID: "tenant-1", BalanceIRR: 1_000_000},
		rates: []*platform.ExchangeRate{
			{USDtoIRR: 850000, Source: "initial"},
			{USDtoIRR: 900000, Source: "refresh"},
		},
	}
	svc := billing.NewServiceWithExchangeRateCacheConfig(memCache, store, 800000, "default", billing.ExchangeRateCacheConfig{
		FreshTTL: time.Nanosecond,
		MaxStale: time.Hour,
	})

	first, err := svc.EstimateCost(context.Background(), openai.ChatCompletionRequest{Model: "gpt-3.5-turbo"})
	require.NoError(t, err)
	assert.Equal(t, 850000.0, first.ExchangeRate)

	time.Sleep(time.Millisecond)
	stale, err := svc.EstimateCost(context.Background(), openai.ChatCompletionRequest{Model: "gpt-3.5-turbo"})
	require.NoError(t, err)
	assert.Equal(t, 850000.0, stale.ExchangeRate)

	require.Eventually(t, func() bool {
		return store.rateCallCount() >= 2
	}, time.Second, 10*time.Millisecond)

	refreshed, err := svc.EstimateCost(context.Background(), openai.ChatCompletionRequest{Model: "gpt-3.5-turbo"})
	require.NoError(t, err)
	assert.Equal(t, 900000.0, refreshed.ExchangeRate)
}

func TestBillingService_ShouldFallbackToDefaultRateWhenExchangeRateUnavailable(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)
	store := &walletStoreStub{
		wallet:  &platform.Wallet{TenantID: "tenant-1", BalanceIRR: 1_000_000},
		rateErr: errors.New("db unavailable"),
	}
	svc := billing.NewServiceWithExchangeRateCacheConfig(memCache, store, 810000, "default", billing.ExchangeRateCacheConfig{
		FreshTTL: time.Minute,
		MaxStale: time.Hour,
	})

	out, err := svc.EstimateCost(context.Background(), openai.ChatCompletionRequest{Model: "gpt-3.5-turbo"})
	require.NoError(t, err)
	assert.Equal(t, 810000.0, out.ExchangeRate)
	assert.Equal(t, 1, store.rateCallCount())
	assert.Equal(t, 1, store.upsertCalls)
	require.NotNil(t, store.upsertedRate)
	assert.Equal(t, 810000.0, store.upsertedRate.USDtoIRR)
}

func TestBillingService_ShouldKeepStaleExchangeRateWhenAsyncRefreshFails(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)
	store := &walletStoreStub{
		wallet: &platform.Wallet{TenantID: "tenant-1", BalanceIRR: 1_000_000},
		rates: []*platform.ExchangeRate{
			{USDtoIRR: 850000, Source: "initial"},
		},
	}
	svc := billing.NewServiceWithExchangeRateCacheConfig(memCache, store, 810000, "default", billing.ExchangeRateCacheConfig{
		FreshTTL: time.Nanosecond,
		MaxStale: time.Hour,
	})

	first, err := svc.EstimateCost(context.Background(), openai.ChatCompletionRequest{Model: "gpt-3.5-turbo"})
	require.NoError(t, err)
	assert.Equal(t, 850000.0, first.ExchangeRate)

	store.mu.Lock()
	store.rateErr = errors.New("db unavailable")
	store.mu.Unlock()

	time.Sleep(time.Millisecond)
	stale, err := svc.EstimateCost(context.Background(), openai.ChatCompletionRequest{Model: "gpt-3.5-turbo"})
	require.NoError(t, err)
	assert.Equal(t, 850000.0, stale.ExchangeRate)

	require.Eventually(t, func() bool {
		return store.rateCallCount() >= 2
	}, time.Second, 10*time.Millisecond)

	afterFailedRefresh, err := svc.EstimateCost(context.Background(), openai.ChatCompletionRequest{Model: "gpt-3.5-turbo"})
	require.NoError(t, err)
	assert.Equal(t, 850000.0, afterFailedRefresh.ExchangeRate)
	assert.Equal(t, 0, store.upsertCalls)
}

func TestBillingService_ShouldCoalesceConcurrentExchangeRateMisses(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)
	store := &walletStoreStub{
		wallet: &platform.Wallet{TenantID: "tenant-1", BalanceIRR: 1_000_000},
		rate:   &platform.ExchangeRate{USDtoIRR: 850000, Source: "db"},
	}
	svc := billing.NewServiceWithExchangeRateCacheConfig(memCache, store, 800000, "default", billing.ExchangeRateCacheConfig{
		FreshTTL: time.Minute,
		MaxStale: time.Hour,
	})

	const workers = 20
	var wg sync.WaitGroup
	wg.Add(workers)
	errs := make(chan error, workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			_, err := svc.EstimateCost(context.Background(), openai.ChatCompletionRequest{Model: "gpt-3.5-turbo"})
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		require.NoError(t, err)
	}
	assert.Equal(t, 1, store.rateCallCount())
}

func TestBillingService_ShouldFinalizeCostWithPer1KScaling(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := cache.NewMemoryCache(logger)
	svc := billing.NewService(memCache, &walletStoreStub{
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
	svc := billing.NewService(memCache, &walletStoreStub{
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
	svc := billing.NewService(memCache, &walletStoreStub{
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
	svc := billing.NewService(memCache, &walletStoreStub{
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
