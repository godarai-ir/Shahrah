package billing

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
	"github.com/Delpak/shahrah/services/control-plane/internal/runtimeprofile"
)

var (
	// ErrRateLimitExceeded is returned when per-key RPM/TPM is exhausted.
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
	// ErrBudgetExceeded is returned when max budget or wallet checks fail.
	ErrBudgetExceeded = errors.New("budget exceeded")
	// ErrBudgetUnavailable is returned when the spend allowance window cannot be read or updated.
	ErrBudgetUnavailable = errors.New("budget unavailable")
	// ErrSubscriptionInactive is returned when a tenant has no usable commercial subscription.
	ErrSubscriptionInactive = errors.New("subscription inactive")
)

type walletStore interface {
	GetWallet(ctx context.Context, tenantID string) (*platform.Wallet, error)
	GetLatestExchangeRate(ctx context.Context) (*platform.ExchangeRate, error)
	UpsertExchangeRate(ctx context.Context, rate platform.ExchangeRate) (*platform.ExchangeRate, error)
}

type entitlementStore interface {
	ResolveEffectiveEntitlementLimits(ctx context.Context, tenantID string) (*platform.EffectiveEntitlementLimits, error)
	CheckAndIncrementTenantUsageCounter(ctx context.Context, tenantID string, featureKey string, limit float64, increment float64, now time.Time) (*platform.UsageCounter, bool, error)
}

// BudgetWindowCache provides atomic spend-allowance operations.
type BudgetWindowCache interface {
	SeedBudgetWindow(ctx context.Context, seed cache.BudgetWindowSeed) (*cache.BudgetWindowResult, error)
	ReserveBudget(ctx context.Context, req cache.BudgetReservationRequest) (*cache.BudgetWindowResult, error)
	CommitBudget(ctx context.Context, req cache.BudgetCommitRequest) (*cache.BudgetWindowResult, error)
	ReleaseBudget(ctx context.Context, req cache.BudgetReleaseRequest) (*cache.BudgetWindowResult, error)
}

// RuntimeProfileStore resolves compiled billing state.
type RuntimeProfileStore interface {
	Get(ctx context.Context, tenantID string) (*runtimeprofile.Profile, error)
}

// PricingEngine estimates and finalizes provider usage costs.
type PricingEngine struct {
	defaultRate       platform.ExchangeRate
	exchangeRateCache *exchangeRateCache
	pricingCache      *pricingCache
	anthropicPricing  *AnthropicPricingSnapshotManager
	catalogPricing    map[string]*CatalogPricingSnapshotManager
}

// BudgetService owns budget admission, reservation, and settlement.
type BudgetService struct {
	cache                cache.Cache
	budgetWindows        BudgetWindowCache
	store                walletStore
	entitlements         entitlementStore
	runtimeProfiles      RuntimeProfileStore
	walletCacheTTL       time.Duration
	budgetWindowTTLGrace time.Duration
}

// RateLimiter owns policy and subscription-based request admission.
type RateLimiter struct {
	cache             cache.Cache
	entitlements      entitlementStore
	runtimeProfiles   RuntimeProfileStore
	rateLimitFailOpen bool
}

// Service is the compatibility facade for focused billing components.
type Service struct {
	*PricingEngine
	*BudgetService
	*RateLimiter
}

// Dependencies declares the storage and runtime capabilities used by billing.
type Dependencies struct {
	Cache           cache.Cache
	BudgetWindows   BudgetWindowCache
	Wallets         walletStore
	Entitlements    entitlementStore
	RuntimeProfiles RuntimeProfileStore
}

// NewService creates a billing enforcement service.
func NewService(c cache.Cache, store walletStore, defaultUSDtoIRR float64, source string) *Service {
	return NewServiceWithCacheConfig(c, store, defaultUSDtoIRR, source, loadExchangeRateCacheConfigFromEnv(), loadPricingCacheConfigFromEnv())
}

// NewServiceWithRuntimeProfiles creates a billing service with compiled policy state.
func NewServiceWithRuntimeProfiles(
	c cache.Cache,
	store walletStore,
	runtimeProfiles RuntimeProfileStore,
	defaultUSDtoIRR float64,
	source string,
) *Service {
	return NewServiceWithDependencies(Dependencies{
		Cache:           c,
		Wallets:         store,
		RuntimeProfiles: runtimeProfiles,
	}, defaultUSDtoIRR, source)
}

// NewServiceWithRepositories creates a billing service with explicit feature repositories.
func NewServiceWithRepositories(
	c cache.Cache,
	wallets walletStore,
	entitlements entitlementStore,
	defaultUSDtoIRR float64,
	source string,
) *Service {
	return NewServiceWithRepositoriesAndRuntimeProfiles(
		c,
		wallets,
		entitlements,
		nil,
		defaultUSDtoIRR,
		source,
	)
}

// NewServiceWithRepositoriesAndRuntimeProfiles creates a fully wired billing service.
func NewServiceWithRepositoriesAndRuntimeProfiles(
	c cache.Cache,
	wallets walletStore,
	entitlements entitlementStore,
	runtimeProfiles RuntimeProfileStore,
	defaultUSDtoIRR float64,
	source string,
) *Service {
	return NewServiceWithDependencies(Dependencies{
		Cache:           c,
		Wallets:         wallets,
		Entitlements:    entitlements,
		RuntimeProfiles: runtimeProfiles,
	}, defaultUSDtoIRR, source)
}

// NewServiceWithDependencies creates a billing service with explicit capabilities.
func NewServiceWithDependencies(dependencies Dependencies, defaultUSDtoIRR float64, source string) *Service {
	return newService(
		dependencies,
		defaultUSDtoIRR,
		source,
		loadExchangeRateCacheConfigFromEnv(),
		loadPricingCacheConfigFromEnv(),
	)
}

// NewServiceWithExchangeRateCacheConfig creates a billing service with explicit cache settings.
func NewServiceWithExchangeRateCacheConfig(c cache.Cache, store walletStore, defaultUSDtoIRR float64, source string, exchangeRateCacheConfig ExchangeRateCacheConfig) *Service {
	return NewServiceWithCacheConfig(c, store, defaultUSDtoIRR, source, exchangeRateCacheConfig, loadPricingCacheConfigFromEnv())
}

// NewServiceWithCacheConfig creates a billing service with explicit hot-path cache settings.
func NewServiceWithCacheConfig(c cache.Cache, store walletStore, defaultUSDtoIRR float64, source string, exchangeRateCacheConfig ExchangeRateCacheConfig, pricingCacheConfig PricingCacheConfig) *Service {
	return newService(Dependencies{Cache: c, Wallets: store}, defaultUSDtoIRR, source, exchangeRateCacheConfig, pricingCacheConfig)
}

func newService(
	dependencies Dependencies,
	defaultUSDtoIRR float64,
	source string,
	exchangeRateCacheConfig ExchangeRateCacheConfig,
	pricingCacheConfig PricingCacheConfig,
) *Service {
	defaultRate := platform.ExchangeRate{
		USDtoIRR: defaultUSDtoIRR,
		Source:   source,
	}

	return &Service{
		PricingEngine: &PricingEngine{
			defaultRate:       defaultRate,
			exchangeRateCache: newExchangeRateCache(dependencies.Wallets, dependencies.Cache, defaultRate, exchangeRateCacheConfig),
			pricingCache:      newPricingCache(dependencies.Cache, pricingCacheConfig, modelPriceFromRegistry),
			catalogPricing:    make(map[string]*CatalogPricingSnapshotManager),
		},
		BudgetService: &BudgetService{
			cache:                dependencies.Cache,
			budgetWindows:        dependencies.BudgetWindows,
			store:                dependencies.Wallets,
			entitlements:         dependencies.Entitlements,
			runtimeProfiles:      dependencies.RuntimeProfiles,
			walletCacheTTL:       secondsEnvOrDefault("BILLING_WALLET_CACHE_TTL_SECONDS", 5*time.Second),
			budgetWindowTTLGrace: secondsEnvOrDefault("BILLING_BUDGET_WINDOW_GRACE_SECONDS", 86400*time.Second),
		},
		RateLimiter: &RateLimiter{
			cache:             dependencies.Cache,
			entitlements:      dependencies.Entitlements,
			runtimeProfiles:   dependencies.RuntimeProfiles,
			rateLimitFailOpen: boolEnvOrDefault("RATE_LIMIT_POLICY_REDIS_FAIL_OPEN", false),
		},
	}
}

// Shutdown cancels and drains background billing work.
func (s *Service) Shutdown(ctx context.Context) error {
	if s == nil || s.PricingEngine == nil || s.exchangeRateCache == nil {
		return nil
	}

	return s.exchangeRateCache.shutdown(ctx)
}

func (s *PricingEngine) SetAnthropicPricingSnapshotManager(
	manager *AnthropicPricingSnapshotManager,
) {
	s.anthropicPricing = manager
}

// SetCatalogPricingSnapshotManager registers a provider-neutral catalog
// snapshot for providers whose pricing is maintained by Model Discovery.
func (s *PricingEngine) SetCatalogPricingSnapshotManager(
	provider string,
	manager *CatalogPricingSnapshotManager,
) {
	if s == nil {
		return
	}

	if s.catalogPricing == nil {
		s.catalogPricing = make(map[string]*CatalogPricingSnapshotManager)
	}

	s.catalogPricing[strings.ToLower(strings.TrimSpace(provider))] = manager
}

func cloneInterfaceMap(in map[string]interface{}) map[string]interface{} {
	if len(in) == 0 {
		return nil
	}

	out := make(map[string]interface{}, len(in))
	for key, value := range in {
		out[key] = value
	}

	return out
}
