package billing

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
	"github.com/Delpak/shahrah/services/control-plane/internal/runtimeprofile"
	"go.uber.org/zap"
)

var (
	// ErrRateLimitExceeded is returned when per-key RPM is exhausted.
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
	// ErrBudgetExceeded is returned when max budget or wallet checks fail.
	ErrBudgetExceeded = errors.New("budget exceeded")
	// ErrBudgetUnavailable is returned when the spend allowance window cannot be read or updated.
	ErrBudgetUnavailable = errors.New("budget unavailable")
	// ErrSubscriptionInactive is returned when a tenant has no usable commercial subscription.
	ErrSubscriptionInactive = errors.New("subscription inactive")
)

// exchangeRateStore persists the USD/IRR exchange rate the pricing engine uses.
type exchangeRateStore interface {
	GetLatestExchangeRate(ctx context.Context) (*platform.ExchangeRate, error)
	UpsertExchangeRate(ctx context.Context, rate platform.ExchangeRate) (*platform.ExchangeRate, error)
}

type entitlementStore interface {
	ResolveEffectiveEntitlementLimits(ctx context.Context, tenantID string) (*platform.EffectiveEntitlementLimits, error)
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
	GetSnapshot(ctx context.Context, tenantID string) (*runtimeprofile.Profile, error)
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
	entitlements         entitlementStore
	runtimeProfiles      RuntimeProfileStore
	creditPools          CreditPoolStore
	creditRateCards      CreditRateCardProvider
	logger               *zap.Logger
	budgetWindowTTLGrace time.Duration
}

// RateLimiter owns policy and subscription-based request admission.
type RateLimiter struct {
	cache             cache.Cache
	entitlements      entitlementStore
	runtimeProfiles   RuntimeProfileStore
	rateLimitFailOpen bool
	policyCacheMu     sync.RWMutex
	policyCache       map[string]compiledRateLimitProfile
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
	ExchangeRates   exchangeRateStore
	Entitlements    entitlementStore
	RuntimeProfiles RuntimeProfileStore
	// CreditPools and CreditRateCards enforce credit tenants in credits.
	CreditPools     CreditPoolStore
	CreditRateCards CreditRateCardProvider
	Logger          *zap.Logger
	Config          *ServiceConfig
}

// ServiceConfig controls billing hot paths without reading process state.
type ServiceConfig struct {
	ExchangeRateCache ExchangeRateCacheConfig
	PricingCache      PricingCacheConfig
	BudgetWindowGrace time.Duration
	RateLimitFailOpen bool
}

// NewService creates a billing enforcement service.
func NewService(c cache.Cache, store exchangeRateStore, defaultUSDtoIRR float64, source string) *Service {
	return NewServiceWithCacheConfig(c, store, defaultUSDtoIRR, source, defaultExchangeRateCacheConfig(), defaultPricingCacheConfig())
}

// NewServiceWithRuntimeProfiles creates a billing service with compiled policy state.
func NewServiceWithRuntimeProfiles(
	c cache.Cache,
	store exchangeRateStore,
	runtimeProfiles RuntimeProfileStore,
	defaultUSDtoIRR float64,
	source string,
) *Service {
	return NewServiceWithDependencies(Dependencies{
		Cache:           c,
		ExchangeRates:   store,
		RuntimeProfiles: runtimeProfiles,
	}, defaultUSDtoIRR, source)
}

// NewServiceWithRepositories creates a billing service with explicit feature repositories.
func NewServiceWithRepositories(
	c cache.Cache,
	wallets exchangeRateStore,
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
	wallets exchangeRateStore,
	entitlements entitlementStore,
	runtimeProfiles RuntimeProfileStore,
	defaultUSDtoIRR float64,
	source string,
) *Service {
	return NewServiceWithDependencies(Dependencies{
		Cache:           c,
		ExchangeRates:   wallets,
		Entitlements:    entitlements,
		RuntimeProfiles: runtimeProfiles,
	}, defaultUSDtoIRR, source)
}

// NewServiceWithDependencies creates a billing service with explicit capabilities.
func NewServiceWithDependencies(dependencies Dependencies, defaultUSDtoIRR float64, source string) *Service {
	serviceConfig := defaultServiceConfig()
	if dependencies.Config != nil {
		serviceConfig = *dependencies.Config
	}

	return newService(
		dependencies,
		defaultUSDtoIRR,
		source,
		serviceConfig,
	)
}

// NewServiceWithExchangeRateCacheConfig creates a billing service with explicit cache settings.
func NewServiceWithExchangeRateCacheConfig(c cache.Cache, store exchangeRateStore, defaultUSDtoIRR float64, source string, exchangeRateCacheConfig ExchangeRateCacheConfig) *Service {
	return NewServiceWithCacheConfig(c, store, defaultUSDtoIRR, source, exchangeRateCacheConfig, defaultPricingCacheConfig())
}

// NewServiceWithCacheConfig creates a billing service with explicit hot-path cache settings.
func NewServiceWithCacheConfig(c cache.Cache, store exchangeRateStore, defaultUSDtoIRR float64, source string, exchangeRateCacheConfig ExchangeRateCacheConfig, pricingCacheConfig PricingCacheConfig) *Service {
	serviceConfig := defaultServiceConfig()
	serviceConfig.ExchangeRateCache = exchangeRateCacheConfig
	serviceConfig.PricingCache = pricingCacheConfig

	return newService(Dependencies{Cache: c, ExchangeRates: store}, defaultUSDtoIRR, source, serviceConfig)
}

func newService(
	dependencies Dependencies,
	defaultUSDtoIRR float64,
	source string,
	serviceConfig ServiceConfig,
) *Service {
	serviceConfig = normalizeServiceConfig(serviceConfig)
	defaultRate := platform.ExchangeRate{
		USDtoIRR: defaultUSDtoIRR,
		Source:   source,
	}

	logger := dependencies.Logger
	if logger == nil {
		logger = zap.NewNop()
	}

	return &Service{
		PricingEngine: &PricingEngine{
			defaultRate:       defaultRate,
			exchangeRateCache: newExchangeRateCache(dependencies.ExchangeRates, dependencies.Cache, defaultRate, serviceConfig.ExchangeRateCache),
			pricingCache:      newPricingCache(dependencies.Cache, serviceConfig.PricingCache, modelPriceFromRegistry),
			catalogPricing:    make(map[string]*CatalogPricingSnapshotManager),
		},
		BudgetService: &BudgetService{
			cache:                dependencies.Cache,
			budgetWindows:        dependencies.BudgetWindows,
			entitlements:         dependencies.Entitlements,
			runtimeProfiles:      dependencies.RuntimeProfiles,
			creditPools:          dependencies.CreditPools,
			creditRateCards:      dependencies.CreditRateCards,
			logger:               logger,
			budgetWindowTTLGrace: serviceConfig.BudgetWindowGrace,
		},
		RateLimiter: &RateLimiter{
			cache:             dependencies.Cache,
			entitlements:      dependencies.Entitlements,
			runtimeProfiles:   dependencies.RuntimeProfiles,
			rateLimitFailOpen: serviceConfig.RateLimitFailOpen,
			policyCache:       make(map[string]compiledRateLimitProfile),
		},
	}
}

func defaultServiceConfig() ServiceConfig {
	return ServiceConfig{
		ExchangeRateCache: defaultExchangeRateCacheConfig(),
		PricingCache:      defaultPricingCacheConfig(),
		BudgetWindowGrace: 24 * time.Hour,
	}
}

func normalizeServiceConfig(cfg ServiceConfig) ServiceConfig {
	cfg.ExchangeRateCache = normalizeExchangeRateCacheConfig(cfg.ExchangeRateCache)

	cfg.PricingCache = normalizePricingCacheConfig(cfg.PricingCache)
	if cfg.BudgetWindowGrace <= 0 {
		cfg.BudgetWindowGrace = 24 * time.Hour
	}

	return cfg
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
