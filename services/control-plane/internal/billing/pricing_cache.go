package billing

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
	"golang.org/x/sync/singleflight"
)

type PricingCacheConfig struct {
	Enabled bool
	TTL     time.Duration
}

type pricingLoader func(model string) priceInfo

type pricingCache struct {
	cache  cache.Cache
	cfg    PricingCacheConfig
	loader pricingLoader
	group  singleflight.Group
}

func loadPricingCacheConfigFromEnv() PricingCacheConfig {
	return PricingCacheConfig{
		Enabled: boolEnvOrDefault("BILLING_PRICING_CACHE_ENABLED", true),
		TTL:     secondsEnvOrDefault("BILLING_PRICING_CACHE_TTL_SECONDS", 300*time.Second),
	}
}

func normalizePricingCacheConfig(cfg PricingCacheConfig) PricingCacheConfig {
	if cfg.TTL <= 0 {
		cfg.TTL = 300 * time.Second
	}

	return cfg
}

func newPricingCache(c cache.Cache, cfg PricingCacheConfig, loader pricingLoader) *pricingCache {
	if loader == nil {
		loader = modelPriceFromRegistry
	}

	return &pricingCache{
		cache:  c,
		cfg:    normalizePricingCacheConfig(cfg),
		loader: loader,
	}
}

func (c *pricingCache) Get(ctx context.Context, model string) priceInfo {
	model = strings.TrimSpace(model)
	if c == nil || !c.cfg.Enabled || c.cache == nil {
		return c.load(model)
	}

	key := pricingCacheKey(model)
	if price, ok := c.cachedPrice(ctx, key); ok {
		return price
	}

	value, err, _ := c.group.Do(key, func() (interface{}, error) {
		if price, ok := c.cachedPrice(ctx, key); ok {
			return price, nil
		}

		price := c.load(model)
		c.store(ctx, key, price)

		return price, nil
	})
	if err != nil {
		return c.load(model)
	}

	price, ok := value.(priceInfo)
	if !ok {
		return c.load(model)
	}

	return price
}

func (c *pricingCache) Invalidate(ctx context.Context, model string) {
	if c == nil || c.cache == nil {
		return
	}

	c.cache.Delete(ctx, pricingCacheKey(model))
}

func (c *pricingCache) load(model string) priceInfo {
	if c == nil || c.loader == nil {
		return modelPriceFromRegistry(model)
	}

	return c.loader(model)
}

func (c *pricingCache) cachedPrice(ctx context.Context, key string) (priceInfo, bool) {
	value, ok := c.cache.Get(ctx, key)
	if !ok {
		return priceInfo{}, false
	}

	var raw []byte

	switch typed := value.(type) {
	case string:
		raw = []byte(typed)
	case []byte:
		raw = typed
	case priceInfo:
		return typed, true
	default:
		return priceInfo{}, false
	}

	var price priceInfo
	if err := json.Unmarshal(raw, &price); err != nil {
		c.cache.Delete(ctx, key)
		return priceInfo{}, false
	}

	return price, true
}

func (c *pricingCache) store(ctx context.Context, key string, price priceInfo) {
	raw, err := json.Marshal(price)
	if err != nil {
		return
	}

	c.cache.Set(ctx, key, string(raw), c.cfg.TTL)
}

func pricingCacheKey(model string) string {
	normalized := strings.ToLower(strings.TrimSpace(model))
	if normalized == "" {
		normalized = "default"
	}

	return fmt.Sprintf("pricing:%s", normalized)
}

func boolEnvOrDefault(key string, fallback bool) bool {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}

	value, err := strconv.ParseBool(raw)
	if err != nil {
		return fallback
	}

	return value
}
