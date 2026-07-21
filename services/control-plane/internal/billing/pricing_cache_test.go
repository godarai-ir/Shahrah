package billing

import (
	"context"
	"math"
	"sync"
	"testing"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
	"go.uber.org/zap"
)

func TestPricingCacheKeyNormalizesModel(t *testing.T) {
	got := pricingCacheKey(" OpenAI/GPT-4O-MINI ")
	if got != "pricing:openai/gpt-4o-mini" {
		t.Fatalf("unexpected pricing cache key: %s", got)
	}
}

func TestPricingCacheHitAvoidsLoader(t *testing.T) {
	ctx := context.Background()

	mem, err := cache.NewMemoryCache(zap.NewNop())
	if err != nil {
		t.Fatal(err)
	}
	defer mem.Close()

	key := pricingCacheKey("openai/gpt-4o-mini")
	mem.Set(ctx, key, `{"Input":0.123,"Output":0.456}`, time.Minute)

	loads := 0
	prices := newPricingCache(mem, PricingCacheConfig{Enabled: true, TTL: time.Minute}, func(model string) priceInfo {
		loads++
		return priceInfo{Input: 1, Output: 2}
	})

	got := prices.Get(ctx, "openai/gpt-4o-mini")
	if got.Input != 0.123 || got.Output != 0.456 {
		t.Fatalf("unexpected cached price: %+v", got)
	}

	if loads != 0 {
		t.Fatalf("expected loader not to run on cache hit, got %d calls", loads)
	}
}

func TestPricingCacheConcurrentMissIsCoalesced(t *testing.T) {
	ctx := context.Background()

	mem, err := cache.NewMemoryCache(zap.NewNop())
	if err != nil {
		t.Fatal(err)
	}
	defer mem.Close()

	var (
		mu    sync.Mutex
		loads int
	)

	prices := newPricingCache(mem, PricingCacheConfig{Enabled: true, TTL: time.Minute}, func(model string) priceInfo {
		time.Sleep(20 * time.Millisecond)
		mu.Lock()
		loads++
		mu.Unlock()

		return priceInfo{Input: 0.5, Output: 1.5}
	})

	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			got := prices.Get(ctx, "openai/gpt-3.5-turbo")
			if got.Input != 0.5 || got.Output != 1.5 {
				t.Errorf("unexpected price: %+v", got)
			}
		}()
	}

	wg.Wait()

	mu.Lock()
	defer mu.Unlock()

	if loads != 1 {
		t.Fatalf("expected one loader call, got %d", loads)
	}
}

func TestCostBreakdownUsesPureArithmetic(t *testing.T) {
	got := costBreakdown("model", 100, 250, priceInfo{Input: 0.0005, Output: 0.0015}, 850000)

	if got.TotalTokens != 350 {
		t.Fatalf("unexpected total tokens: %d", got.TotalTokens)
	}

	if math.Abs(got.TotalCostUSD-0.000425) > 1e-12 {
		t.Fatalf("unexpected total usd: %.12f", got.TotalCostUSD)
	}

	if math.Abs(got.TotalCostIRR-361.25) > 1e-9 {
		t.Fatalf("unexpected total irr: %.12f", got.TotalCostIRR)
	}
}

func TestModelPriceFromRegistryIncludesDallE2LegacyImagePricing(t *testing.T) {
	got := modelPriceFromRegistry("openai/dall-e-2")

	if got.Input != 0 {
		t.Fatalf("unexpected DALL-E 2 input price: %.12f", got.Input)
	}

	if math.Abs(got.Output-0.02) > 1e-12 {
		t.Fatalf("unexpected DALL-E 2 output price: %.12f", got.Output)
	}
}
