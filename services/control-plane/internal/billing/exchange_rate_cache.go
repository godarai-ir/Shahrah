package billing

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
)

const exchangeRateCacheKey = "billing:exchange-rate:latest"

// ExchangeRateCacheConfig controls hot-path exchange-rate caching.
type ExchangeRateCacheConfig struct {
	FreshTTL time.Duration
	MaxStale time.Duration
}

type exchangeRateCacheEntry struct {
	Rate     platform.ExchangeRate `json:"rate"`
	CachedAt time.Time             `json:"cached_at"`
}

type exchangeRateCache struct {
	store       walletStore
	cache       cache.Cache
	defaultRate platform.ExchangeRate
	cfg         ExchangeRateCacheConfig
	now         func() time.Time

	mu         sync.RWMutex
	entry      *exchangeRateCacheEntry
	refreshing bool
	loadMu     sync.Mutex
}

func loadExchangeRateCacheConfigFromEnv() ExchangeRateCacheConfig {
	return ExchangeRateCacheConfig{
		FreshTTL: secondsEnvOrDefault("BILLING_EXCHANGE_RATE_CACHE_TTL_SECONDS", 60*time.Second),
		MaxStale: secondsEnvOrDefault("BILLING_EXCHANGE_RATE_CACHE_MAX_STALE_SECONDS", 10*time.Minute),
	}
}

func normalizeExchangeRateCacheConfig(cfg ExchangeRateCacheConfig) ExchangeRateCacheConfig {
	if cfg.FreshTTL <= 0 {
		cfg.FreshTTL = 60 * time.Second
	}

	if cfg.MaxStale < cfg.FreshTTL {
		cfg.MaxStale = 10 * cfg.FreshTTL
	}

	return cfg
}

func newExchangeRateCache(store walletStore, c cache.Cache, defaultRate platform.ExchangeRate, cfg ExchangeRateCacheConfig) *exchangeRateCache {
	defaultRate = cloneExchangeRate(defaultRate)

	return &exchangeRateCache{
		store:       store,
		cache:       c,
		defaultRate: defaultRate,
		cfg:         normalizeExchangeRateCacheConfig(cfg),
		now:         time.Now,
	}
}

func (c *exchangeRateCache) Get(ctx context.Context) *platform.ExchangeRate {
	if c == nil || c.store == nil {
		return nil
	}

	if rate, state := c.cachedRate(ctx); rate != nil {
		if state == exchangeRateCacheStateStale {
			c.refreshAsync()
		}

		return rate
	}

	return c.refreshSync(ctx)
}

type exchangeRateCacheState int

const (
	exchangeRateCacheStateMissing exchangeRateCacheState = iota
	exchangeRateCacheStateFresh
	exchangeRateCacheStateStale
)

func (c *exchangeRateCache) cachedRate(ctx context.Context) (*platform.ExchangeRate, exchangeRateCacheState) {
	now := c.now()

	c.mu.RLock()
	entry := c.entry
	c.mu.RUnlock()

	if entry == nil {
		entry = c.distributedEntry(ctx)
		if entry != nil {
			c.storeEntry(entry)
		}
	}

	if entry == nil {
		return nil, exchangeRateCacheStateMissing
	}

	age := now.Sub(entry.CachedAt)
	if age <= c.cfg.FreshTTL {
		rate := cloneExchangeRate(entry.Rate)
		return &rate, exchangeRateCacheStateFresh
	}

	if age <= c.cfg.MaxStale {
		rate := cloneExchangeRate(entry.Rate)
		return &rate, exchangeRateCacheStateStale
	}

	return nil, exchangeRateCacheStateMissing
}

func (c *exchangeRateCache) refreshSync(ctx context.Context) *platform.ExchangeRate {
	c.loadMu.Lock()
	defer c.loadMu.Unlock()

	if rate, state := c.cachedRate(ctx); rate != nil && state == exchangeRateCacheStateFresh {
		return rate
	}

	return c.loadAndStoreLocked(ctx, true)
}

func (c *exchangeRateCache) refreshAsync() {
	c.mu.Lock()
	if c.refreshing {
		c.mu.Unlock()
		return
	}

	c.refreshing = true
	c.mu.Unlock()

	go func() {
		defer func() {
			c.mu.Lock()
			c.refreshing = false
			c.mu.Unlock()
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		c.loadMu.Lock()
		defer c.loadMu.Unlock()

		_ = c.loadAndStoreLocked(ctx, false)
	}()
}

func (c *exchangeRateCache) loadAndStoreLocked(ctx context.Context, fallbackDefault bool) *platform.ExchangeRate {
	rate, err := c.store.GetLatestExchangeRate(ctx)
	if err != nil || rate == nil {
		if !fallbackDefault {
			return nil
		}

		if err := c.persistDefaultExchangeRate(ctx); err != nil {
			return nil
		}

		rate = &c.defaultRate
	}

	cloned := cloneExchangeRate(*rate)
	entry := &exchangeRateCacheEntry{Rate: cloned, CachedAt: c.now()}
	c.storeEntry(entry)
	c.storeDistributed(ctx, entry)

	return &cloned
}

func (c *exchangeRateCache) storeEntry(entry *exchangeRateCacheEntry) {
	if entry == nil {
		return
	}

	c.mu.Lock()
	cloned := *entry
	cloned.Rate = cloneExchangeRate(entry.Rate)
	c.entry = &cloned
	c.mu.Unlock()
}

func (c *exchangeRateCache) distributedEntry(ctx context.Context) *exchangeRateCacheEntry {
	if c.cache == nil {
		return nil
	}

	value, ok := c.cache.Get(ctx, exchangeRateCacheKey)
	if !ok {
		return nil
	}

	var raw []byte

	switch v := value.(type) {
	case string:
		raw = []byte(v)
	case []byte:
		raw = append([]byte(nil), v...)
	case exchangeRateCacheEntry:
		entry := v
		return &entry
	case *exchangeRateCacheEntry:
		if v == nil {
			return nil
		}

		entry := *v

		return &entry
	default:
		return nil
	}

	var entry exchangeRateCacheEntry
	if err := json.Unmarshal(raw, &entry); err != nil {
		return nil
	}

	return &entry
}

func (c *exchangeRateCache) storeDistributed(ctx context.Context, entry *exchangeRateCacheEntry) {
	if c.cache == nil || entry == nil {
		return
	}

	raw, err := json.Marshal(entry)
	if err != nil {
		return
	}

	c.cache.Set(ctx, exchangeRateCacheKey, string(raw), c.cfg.MaxStale)
}

func (c *exchangeRateCache) persistDefaultExchangeRate(ctx context.Context) error {
	_, err := c.store.UpsertExchangeRate(ctx, c.defaultRate)

	return err
}

func cloneExchangeRate(rate platform.ExchangeRate) platform.ExchangeRate {
	return platform.ExchangeRate{
		ID:        rate.ID,
		USDtoIRR:  rate.USDtoIRR,
		Source:    rate.Source,
		FetchedAt: rate.FetchedAt,
	}
}

func secondsEnvOrDefault(key string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}

	seconds, err := strconv.Atoi(raw)
	if err != nil || seconds <= 0 {
		return fallback
	}

	return time.Duration(seconds) * time.Second
}
