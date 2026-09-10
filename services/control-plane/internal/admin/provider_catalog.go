package admin

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/billing"
)

var errDiscoveryNotFound = errors.New("discovery resource not found")
var errDiscoveryUnavailable = errors.New("model-discovery service unavailable")

const discoveryTimeout = 8 * time.Second

// DiscoveryConfig contains immutable model-discovery client and cache settings.
type DiscoveryConfig struct {
	BaseURL  string
	Token    string
	CacheTTL time.Duration
}

func (c DiscoveryConfig) normalized() DiscoveryConfig {
	c.BaseURL = strings.TrimSuffix(strings.TrimSpace(c.BaseURL), "/")
	c.Token = strings.TrimSpace(c.Token)

	return c
}

const (
	internalMockProvider = "mock"
	internalMockModel    = "mock/chat"
)

type providerCatalogCache struct {
	mu          sync.RWMutex
	providers   []platform.ProviderCatalogProvider
	providersAt time.Time
	models      map[string]cachedModels
}

type cachedModels struct {
	items []platform.ProviderCatalogModel
	at    time.Time
}

var discoveryCache = &providerCatalogCache{
	models: make(map[string]cachedModels),
}

type discoveryTelemetryCounters struct {
	providersCacheHit  uint64
	providersCacheMiss uint64
	modelsCacheHit     uint64
	modelsCacheMiss    uint64
	discoveryError     uint64
	discoveryFallback  uint64
}

var discoveryTelemetry discoveryTelemetryCounters

type discoveryProvider struct {
	Slug                    string            `json:"slug"`
	Name                    string            `json:"name"`
	Description             string            `json:"description"`
	DocsURL                 string            `json:"docsUrl"`
	LogoURL                 string            `json:"logoUrl"`
	ModelCount              *int              `json:"modelCount"`
	Models                  []json.RawMessage `json:"models"`
	Enabled                 bool              `json:"enabled"`
	SupportsByok            bool              `json:"supportsByok"`
	SupportsRegionalRouting bool              `json:"supportsRegionalRouting"`
}

type modelDiscoveryPricing struct {
	InputPerMillionTokens  *float64 `json:"inputPerMillionTokens"`
	OutputPerMillionTokens *float64 `json:"outputPerMillionTokens"`
}

type modelDiscoveryModel struct {
	CanonicalID      string                 `json:"canonicalId"`
	Name             string                 `json:"name"`
	Mode             []string               `json:"mode"`
	ContextWindow    *int                   `json:"contextWindow"`
	Capabilities     []string               `json:"capabilities"`
	InputModalities  []string               `json:"inputModalities"`
	OutputModalities []string               `json:"outputModalities"`
	Pricing          *modelDiscoveryPricing `json:"pricing"`
	Status           string                 `json:"status"`
	PricingRates     []struct {
		Dimension     string                 `json:"dimension"`
		Unit          string                 `json:"unit"`
		UnitPriceUSD  string                 `json:"unitPriceUsd"`
		Condition     map[string]interface{} `json:"condition"`
		Region        string                 `json:"region"`
		EffectiveFrom *time.Time             `json:"effectiveFrom"`
		EffectiveTo   *time.Time             `json:"effectiveTo"`
		SourceURL     string                 `json:"sourceUrl"`
	} `json:"pricingRates"`
}

func normalizeProviderSlug(value string) string {
	normalized := strings.TrimSpace(strings.ToLower(value))
	switch normalized {
	case "perplexity":
		return "perplexity-ai"
	default:
		return normalized
	}
}

func discoveryProviderExists(ctx context.Context, provider string, cfg DiscoveryConfig) (bool, error) {
	if normalizeProviderSlug(provider) == internalMockProvider {
		return true, nil
	}

	items, err := discoveryProviders(ctx, cfg)
	if err != nil {
		return false, err
	}

	normalized := normalizeProviderSlug(provider)
	for _, item := range items {
		if normalizeProviderSlug(item.Provider) == normalized {
			return true, nil
		}
	}

	return false, nil
}

func discoveryProviders(ctx context.Context, cfg DiscoveryConfig) ([]platform.ProviderCatalogProvider, error) {
	cfg = cfg.normalized()

	if cached, ok := getCachedProviders(cfg.CacheTTL); ok {
		atomic.AddUint64(&discoveryTelemetry.providersCacheHit, 1)
		return cached, nil
	}

	atomic.AddUint64(&discoveryTelemetry.providersCacheMiss, 1)

	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("%w: SHAHRAH_MODEL_DISCOVERY_URL is not configured; retry after restoring model-discovery connectivity", errDiscoveryUnavailable)
	}

	body, status, err := discoveryGET(ctx, cfg.BaseURL+"/providers", cfg)
	if err != nil {
		atomic.AddUint64(&discoveryTelemetry.discoveryError, 1)
		return nil, fmt.Errorf("%w: failed to fetch providers; retry after restoring model-discovery connectivity: %v", errDiscoveryUnavailable, err)
	}

	if status == http.StatusNotFound {
		atomic.AddUint64(&discoveryTelemetry.discoveryError, 1)
		return nil, fmt.Errorf("%w: /providers endpoint not found on model-discovery; retry after fixing service routing", errDiscoveryUnavailable)
	}

	if status >= http.StatusBadRequest {
		atomic.AddUint64(&discoveryTelemetry.discoveryError, 1)
		return nil, fmt.Errorf("%w: providers request failed with status %d; retry after restoring model-discovery connectivity", errDiscoveryUnavailable, status)
	}

	items, err := decodeProvidersPayload(body)
	if err != nil {
		atomic.AddUint64(&discoveryTelemetry.discoveryError, 1)
		return nil, fmt.Errorf("%w: invalid providers payload; retry after fixing model-discovery response: %v", errDiscoveryUnavailable, err)
	}

	normalized := normalizeProviders(items)
	if len(normalized) == 0 {
		atomic.AddUint64(&discoveryTelemetry.discoveryError, 1)
		return nil, fmt.Errorf("%w: model-discovery returned an empty provider catalog; retry after restoring discovery data", errDiscoveryUnavailable)
	}

	setCachedProviders(normalized)

	return normalized, nil
}

func discoveryProviderModels(ctx context.Context, provider string, cfg DiscoveryConfig) ([]platform.ProviderCatalogModel, error) {
	cfg = cfg.normalized()

	normalizedProvider := normalizeProviderSlug(provider)
	if normalizedProvider == "" {
		return nil, fmt.Errorf("provider is required")
	}

	if normalizedProvider == internalMockProvider {
		zeroPrice := 0.0

		return []platform.ProviderCatalogModel{{
			Provider:         internalMockProvider,
			Model:            internalMockModel,
			DisplayName:      "Internal Mock Chat",
			Modes:            []string{"chat", "completion"},
			Types:            []platform.ModelType{platform.ModelTypeChat, platform.ModelTypeCompletion},
			InputModalities:  []string{"text"},
			OutputModalities: []string{"text"},
			Capabilities:     []string{"chat", "completion"},
			Features:         []string{"chat", "completion"},
			PublicPricing: &platform.ModelPrice{
				InputUSDPer1M:  &zeroPrice,
				OutputUSDPer1M: &zeroPrice,
			},
		}}, nil
	}

	if normalizedProvider == internalMockProvider {
		zeroPrice := 0.0

		return []platform.ProviderCatalogModel{{
			Provider:         internalMockProvider,
			Model:            internalMockModel,
			DisplayName:      "Internal Mock Chat",
			Modes:            []string{"chat", "completion"},
			Types:            []platform.ModelType{platform.ModelTypeChat, platform.ModelTypeCompletion},
			InputModalities:  []string{"text"},
			OutputModalities: []string{"text"},
			Capabilities:     []string{"chat", "completion"},
			Features:         []string{"chat", "completion"},
			PublicPricing: &platform.ModelPrice{
				InputUSDPer1M:  &zeroPrice,
				OutputUSDPer1M: &zeroPrice,
			},
		}}, nil
	}

	if cached, ok := getCachedProviderModels(normalizedProvider, cfg.CacheTTL); ok {
		atomic.AddUint64(&discoveryTelemetry.modelsCacheHit, 1)
		return cached, nil
	}

	atomic.AddUint64(&discoveryTelemetry.modelsCacheMiss, 1)

	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("%w: SHAHRAH_MODEL_DISCOVERY_URL is not configured; retry after restoring model-discovery connectivity", errDiscoveryUnavailable)
	}

	body, status, err := discoveryGET(ctx, cfg.BaseURL+"/providers/"+url.PathEscape(normalizedProvider)+"/models", cfg)
	if err == nil && status == http.StatusNotFound {
		items, queryErr := discoveryProviderModelsByQuery(ctx, cfg.BaseURL, normalizedProvider, cfg)
		if queryErr != nil {
			return nil, queryErr
		}

		setCachedProviderModels(normalizedProvider, items)

		return items, nil
	}

	if err != nil {
		atomic.AddUint64(&discoveryTelemetry.discoveryError, 1)
		return nil, fmt.Errorf("%w: failed to fetch models for provider %q; retry after restoring model-discovery connectivity: %v", errDiscoveryUnavailable, normalizedProvider, err)
	}

	if status == http.StatusNotFound {
		return nil, errDiscoveryNotFound
	}

	if status >= http.StatusBadRequest {
		atomic.AddUint64(&discoveryTelemetry.discoveryError, 1)
		return nil, fmt.Errorf("%w: model request for provider %q failed with status %d; retry after restoring model-discovery connectivity", errDiscoveryUnavailable, normalizedProvider, status)
	}

	items, err := decodeProviderModelsPayload(body)
	if err != nil {
		atomic.AddUint64(&discoveryTelemetry.discoveryError, 1)
		return nil, fmt.Errorf("%w: invalid model payload for provider %q; retry after fixing model-discovery response: %v", errDiscoveryUnavailable, normalizedProvider, err)
	}

	for i := range items {
		items[i].Provider = normalizedProvider
		if items[i].Model == "" {
			atomic.AddUint64(&discoveryTelemetry.discoveryError, 1)
			return nil, fmt.Errorf("%w: model-discovery returned an invalid model entry for provider %q; retry after fixing discovery data", errDiscoveryUnavailable, normalizedProvider)
		}

		items[i].Types = normalizeModelTypes(items[i].Types)
		items[i].PublicPricing = fillModelPricingFromRegistry(normalizedProvider, items[i].Model, items[i].PublicPricing)
	}

	setCachedProviderModels(normalizedProvider, items)

	return items, nil
}

func discoveryProviderModelsByQuery(ctx context.Context, baseURL string, provider string, cfg DiscoveryConfig) ([]platform.ProviderCatalogModel, error) {
	const (
		pageSize = 100
		maxPages = 20
	)

	items := make([]platform.ProviderCatalogModel, 0, pageSize)
	for page := 1; page <= maxPages; page++ {
		queryURL := fmt.Sprintf("%s/models?provider=%s&limit=%d&page=%d", baseURL, url.QueryEscape(provider), pageSize, page)

		body, status, err := discoveryGET(ctx, queryURL, cfg)
		if err != nil {
			atomic.AddUint64(&discoveryTelemetry.discoveryError, 1)
			return nil, fmt.Errorf("%w: failed to fetch models for provider %q; retry after restoring model-discovery connectivity: %v", errDiscoveryUnavailable, provider, err)
		}

		if status == http.StatusNotFound {
			return nil, errDiscoveryNotFound
		}

		if status >= http.StatusBadRequest {
			atomic.AddUint64(&discoveryTelemetry.discoveryError, 1)
			return nil, fmt.Errorf("%w: model request for provider %q failed with status %d; retry after restoring model-discovery connectivity", errDiscoveryUnavailable, provider, status)
		}

		pageItems, err := decodeProviderModelsPayload(body)
		if err != nil {
			atomic.AddUint64(&discoveryTelemetry.discoveryError, 1)
			return nil, fmt.Errorf("%w: invalid model payload for provider %q; retry after fixing model-discovery response: %v", errDiscoveryUnavailable, provider, err)
		}

		if len(pageItems) == 0 {
			break
		}

		items = append(items, pageItems...)
		if len(pageItems) < pageSize {
			break
		}
	}

	for i := range items {
		items[i].Provider = provider
		if items[i].Model == "" {
			atomic.AddUint64(&discoveryTelemetry.discoveryError, 1)
			return nil, fmt.Errorf("%w: model-discovery returned an invalid model entry for provider %q; retry after fixing discovery data", errDiscoveryUnavailable, provider)
		}

		items[i].Types = normalizeModelTypes(items[i].Types)
		items[i].PublicPricing = fillModelPricingFromRegistry(provider, items[i].Model, items[i].PublicPricing)
	}

	return items, nil
}

func invalidateDiscoveryCache() {
	discoveryCache.mu.Lock()
	defer discoveryCache.mu.Unlock()

	discoveryCache.providers = nil
	discoveryCache.providersAt = time.Time{}
	discoveryCache.models = make(map[string]cachedModels)
}

func getCachedProviders(ttl time.Duration) ([]platform.ProviderCatalogProvider, bool) {
	if ttl <= 0 {
		return nil, false
	}

	discoveryCache.mu.RLock()
	defer discoveryCache.mu.RUnlock()

	if len(discoveryCache.providers) == 0 || discoveryCache.providersAt.IsZero() {
		return nil, false
	}

	if time.Since(discoveryCache.providersAt) > ttl {
		return nil, false
	}

	out := make([]platform.ProviderCatalogProvider, len(discoveryCache.providers))
	copy(out, discoveryCache.providers)

	return out, true
}

func setCachedProviders(items []platform.ProviderCatalogProvider) {
	discoveryCache.mu.Lock()
	defer discoveryCache.mu.Unlock()

	out := make([]platform.ProviderCatalogProvider, len(items))
	copy(out, items)
	discoveryCache.providers = out
	discoveryCache.providersAt = time.Now().UTC()
}

func getCachedProviderModels(provider string, ttl time.Duration) ([]platform.ProviderCatalogModel, bool) {
	if ttl <= 0 {
		return nil, false
	}

	discoveryCache.mu.RLock()
	defer discoveryCache.mu.RUnlock()

	entry, ok := discoveryCache.models[provider]
	if !ok || len(entry.items) == 0 || entry.at.IsZero() {
		return nil, false
	}

	if time.Since(entry.at) > ttl {
		return nil, false
	}

	out := make([]platform.ProviderCatalogModel, len(entry.items))
	copy(out, entry.items)

	return out, true
}

func setCachedProviderModels(provider string, items []platform.ProviderCatalogModel) {
	discoveryCache.mu.Lock()
	defer discoveryCache.mu.Unlock()

	out := make([]platform.ProviderCatalogModel, len(items))
	copy(out, items)
	discoveryCache.models[provider] = cachedModels{
		items: out,
		at:    time.Now().UTC(),
	}
}

func discoveryTelemetrySnapshot() map[string]uint64 {
	return map[string]uint64{
		"providers_cache_hit":  atomic.LoadUint64(&discoveryTelemetry.providersCacheHit),
		"providers_cache_miss": atomic.LoadUint64(&discoveryTelemetry.providersCacheMiss),
		"models_cache_hit":     atomic.LoadUint64(&discoveryTelemetry.modelsCacheHit),
		"models_cache_miss":    atomic.LoadUint64(&discoveryTelemetry.modelsCacheMiss),
		"discovery_error":      atomic.LoadUint64(&discoveryTelemetry.discoveryError),
		"discovery_fallback":   atomic.LoadUint64(&discoveryTelemetry.discoveryFallback),
	}
}

func resolveProviderAccountModelSelections(ctx context.Context, provider string, selections []platform.UpsertTenantProviderAccountModelSelection, cfg DiscoveryConfig) ([]platform.TenantProviderAccountModel, []string, map[string][]string, error) {
	catalog, err := discoveryProviderModels(ctx, provider, cfg)
	if err != nil {
		return nil, nil, nil, err
	}

	allowed := make(map[string]platform.ProviderCatalogModel, len(catalog))
	for _, item := range catalog {
		model := strings.TrimSpace(item.Model)
		if model == "" {
			continue
		}

		allowed[model] = item
	}

	invalid := make([]string, 0)
	invalidCapabilities := make(map[string][]string)
	seen := make(map[string]struct{}, len(selections))

	resolved := make([]platform.TenantProviderAccountModel, 0, len(selections))
	for _, item := range selections {
		model := strings.TrimSpace(item.Model)
		if model == "" {
			continue
		}

		if _, ok := seen[model]; ok {
			continue
		}

		seen[model] = struct{}{}

		cm, ok := allowed[model]
		if !ok {
			invalid = append(invalid, model)
			continue
		}

		if cm.IsDeprecated {
			invalid = append(invalid, model)
			continue
		}

		requestedCapabilities := normalizeModelTypes(item.EffectiveCapabilities)
		allowedCapabilities := normalizeModelTypes(cm.Types)

		allowedSet := make(map[platform.ModelType]struct{}, len(allowedCapabilities))
		for _, capability := range allowedCapabilities {
			allowedSet[capability] = struct{}{}
		}

		if len(requestedCapabilities) == 0 {
			invalidCapabilities[model] = []string{}
			continue
		}

		disallowed := make([]string, 0)

		for _, capability := range requestedCapabilities {
			if _, ok := allowedSet[capability]; ok {
				continue
			}

			disallowed = append(disallowed, string(capability))
		}

		if len(disallowed) > 0 {
			invalidCapabilities[model] = disallowed
			continue
		}

		pricingMode := item.PricingMode
		if pricingMode == "" {
			pricingMode = platform.PricingModePublic
		}

		resolved = append(resolved, platform.TenantProviderAccountModel{
			Model:            model,
			Enabled:          true,
			Types:            requestedCapabilities,
			InputModalities:  normalizeStringList(cm.InputModalities),
			OutputModalities: normalizeStringList(cm.OutputModalities),
			PricingMode:      pricingMode,
			PublicPricing:    cloneModelPrice(cm.PublicPricing),
			PrivatePricing:   cloneModelPrice(item.PrivatePricing),
			Settings:         map[string]interface{}{},
			Readiness:        append([]platform.CapabilityReadiness(nil), cm.Readiness...),
		})
	}

	sort.Strings(invalid)

	return resolved, invalid, invalidCapabilities, nil
}

func discoveryGET(ctx context.Context, rawURL string, cfg DiscoveryConfig) ([]byte, int, error) {
	body, status, _, err := discoveryFetch(ctx, rawURL, cfg)
	return body, status, err
}

func discoveryFetch(ctx context.Context, rawURL string, cfg DiscoveryConfig) ([]byte, int, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, http.NoBody)
	if err != nil {
		return nil, 0, "", err
	}

	req.Header.Set("Accept", "application/json")

	if strings.TrimSpace(cfg.Token) != "" {
		req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(cfg.Token))
	}

	client := &http.Client{Timeout: discoveryTimeout}

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, "", err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			return
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, "", err
	}

	return body, resp.StatusCode, resp.Header.Get("Content-Type"), nil
}

func decodeProvidersPayload(body []byte) ([]platform.ProviderCatalogProvider, error) {
	return decodeDiscoveryPayload(body, "discovery providers payload is empty", "discovery providers payload does not contain providers", "providers", hasCanonicalProviderEntries, mapDiscoveryProviders)
}

func decodeProviderModelsPayload(body []byte) ([]platform.ProviderCatalogModel, error) {
	return decodeDiscoveryPayload(body, "discovery provider models payload is empty", "discovery provider models payload does not contain models", "models", hasCanonicalModelEntries, mapDiscoveryModels)
}

func decodeDiscoveryPayload[T any, D any](
	body []byte,
	emptyMessage string,
	missingMessage string,
	altKey string,
	hasCanonical func([]T) bool,
	mapDiscovery func([]D) []T,
) ([]T, error) {
	if len(body) == 0 {
		return nil, fmt.Errorf("%s", emptyMessage)
	}

	trimmedBody := strings.TrimSpace(string(body))

	var direct []T
	if err := json.Unmarshal(body, &direct); err == nil {
		if len(direct) == 0 && strings.HasPrefix(trimmedBody, "[") {
			return direct, nil
		}

		if hasCanonical(direct) {
			return direct, nil
		}
	}

	var mdDirect []D
	if err := json.Unmarshal(body, &mdDirect); err == nil {
		if len(mdDirect) == 0 && strings.HasPrefix(trimmedBody, "[") {
			return mapDiscovery(mdDirect), nil
		}

		if len(mdDirect) > 0 {
			return mapDiscovery(mdDirect), nil
		}
	}

	var wrapped map[string]json.RawMessage
	if err := json.Unmarshal(body, &wrapped); err == nil {
		for _, key := range []string{"data", altKey} {
			raw := wrapped[key]
			if len(raw) == 0 {
				continue
			}

			var wrappedDirect []T
			if err := json.Unmarshal(raw, &wrappedDirect); err == nil {
				if len(wrappedDirect) == 0 {
					return wrappedDirect, nil
				}

				if hasCanonical(wrappedDirect) {
					return wrappedDirect, nil
				}
			}

			var wrappedDiscovery []D
			if err := json.Unmarshal(raw, &wrappedDiscovery); err == nil {
				if len(wrappedDiscovery) == 0 {
					return mapDiscovery(wrappedDiscovery), nil
				}

				if len(wrappedDiscovery) > 0 {
					return mapDiscovery(wrappedDiscovery), nil
				}
			}
		}
	}

	return nil, fmt.Errorf("%s", missingMessage)
}

func hasCanonicalProviderEntries(items []platform.ProviderCatalogProvider) bool {
	if len(items) == 0 {
		return false
	}

	for _, item := range items {
		if strings.TrimSpace(item.Provider) != "" {
			return true
		}
	}

	return false
}

func hasCanonicalModelEntries(items []platform.ProviderCatalogModel) bool {
	if len(items) == 0 {
		return false
	}

	for _, item := range items {
		if strings.TrimSpace(item.Model) != "" {
			return true
		}
	}

	return false
}

func mapDiscoveryProviders(items []discoveryProvider) []platform.ProviderCatalogProvider {
	out := make([]platform.ProviderCatalogProvider, 0, len(items))
	for _, item := range items {
		provider := normalizeProviderSlug(item.Slug)
		if provider == "" {
			continue
		}

		displayName := strings.TrimSpace(item.Name)
		if displayName == "" {
			displayName = provider
		}

		modelCount := len(item.Models)
		if item.ModelCount != nil {
			modelCount = *item.ModelCount
		}

		out = append(out, platform.ProviderCatalogProvider{
			Provider:                provider,
			DisplayName:             displayName,
			Description:             strings.TrimSpace(item.Description),
			DocsURL:                 strings.TrimSpace(item.DocsURL),
			LogoURL:                 strings.TrimSpace(item.LogoURL),
			ModelCount:              modelCount,
			Enabled:                 item.Enabled,
			AuthModes:               []platform.ProviderAuthMode{platform.ProviderAuthModeAPIKey},
			SupportsBringYourOwnKey: item.SupportsByok,
			SupportsVirtualKeys:     true,
		})
	}

	return out
}

func mapDiscoveryModels(items []modelDiscoveryModel) []platform.ProviderCatalogModel {
	out := make([]platform.ProviderCatalogModel, 0, len(items))
	for _, item := range items {
		model := strings.TrimSpace(item.CanonicalID)
		if model == "" {
			continue
		}

		displayName := strings.TrimSpace(item.Name)
		if displayName == "" {
			displayName = model
		}

		modes := normalizeStringList(item.Mode)

		types := make([]platform.ModelType, 0, len(modes))
		for _, mode := range modes {
			mt := mapModeToModelType(mode)
			if mt == "" {
				continue
			}

			types = append(types, mt)
		}

		if len(types) == 0 {
			types = append(types, inferModelTypesFromCapabilities(item.Capabilities)...)
		}

		entry := platform.ProviderCatalogModel{
			Model:            model,
			DisplayName:      displayName,
			Modes:            modes,
			Types:            normalizeModelTypes(types),
			ContextWindow:    item.ContextWindow,
			InputModalities:  normalizeStringList(item.InputModalities),
			OutputModalities: normalizeStringList(item.OutputModalities),
			Capabilities:     normalizeStringList(item.Capabilities),
			Features:         normalizeStringList(item.Capabilities),
			IsDeprecated:     strings.TrimSpace(item.Status) != "" && !strings.EqualFold(strings.TrimSpace(item.Status), "active"),
		}
		if item.Pricing != nil {
			entry.PublicPricing = &platform.ModelPrice{
				InputUSDPer1M:  item.Pricing.InputPerMillionTokens,
				OutputUSDPer1M: item.Pricing.OutputPerMillionTokens,
			}
		}

		if len(item.PricingRates) > 0 {
			if entry.PublicPricing == nil {
				entry.PublicPricing = &platform.ModelPrice{}
			}

			entry.PublicPricing.Rates = make([]platform.ModelPricingRate, 0, len(item.PricingRates))
			for _, rate := range item.PricingRates {
				entry.PublicPricing.Rates = append(entry.PublicPricing.Rates, platform.ModelPricingRate{
					Key:           pricingRateKey(rate.Dimension, rate.Unit, rate.Region, rate.Condition, rate.EffectiveFrom, rate.EffectiveTo),
					Dimension:     strings.TrimSpace(rate.Dimension),
					Unit:          strings.TrimSpace(rate.Unit),
					UnitPriceUSD:  strings.TrimSpace(rate.UnitPriceUSD),
					Condition:     cloneInterfaceMap(rate.Condition),
					Region:        strings.TrimSpace(rate.Region),
					EffectiveFrom: rate.EffectiveFrom,
					EffectiveTo:   rate.EffectiveTo,
					SourceURL:     strings.TrimSpace(rate.SourceURL),
				})
			}
		}

		entry.Readiness = catalogCapabilityReadiness(entry, item.Status)

		out = append(out, entry)
	}

	return out
}

func discoveryDisplaySettings(ctx context.Context, cfg DiscoveryConfig) (*platform.ProviderCatalogDisplaySettings, error) {
	cfg = cfg.normalized()
	if cfg.BaseURL == "" {
		return nil, errDiscoveryNotFound
	}

	body, status, err := discoveryGET(ctx, cfg.BaseURL+"/settings/display", cfg)
	if err != nil {
		return nil, err
	}

	if status == http.StatusNotFound {
		return nil, errDiscoveryNotFound
	}

	if status >= http.StatusBadRequest {
		return nil, fmt.Errorf("discovery display settings request failed with status %d", status)
	}

	var direct platform.ProviderCatalogDisplaySettings
	if err := json.Unmarshal(body, &direct); err == nil && (direct.CurrencyUnit != "" || direct.PriceScale != "" || direct.ExchangeRate != nil) {
		return &direct, nil
	}

	var wrapped struct {
		Data platform.ProviderCatalogDisplaySettings `json:"data"`
	}
	if err := json.Unmarshal(body, &wrapped); err == nil && (wrapped.Data.CurrencyUnit != "" || wrapped.Data.PriceScale != "" || wrapped.Data.ExchangeRate != nil) {
		return &wrapped.Data, nil
	}

	return nil, fmt.Errorf("discovery display settings payload does not contain settings")
}

func mapModeToModelType(mode string) platform.ModelType {
	switch strings.TrimSpace(strings.ToLower(mode)) {
	case "chat":
		return platform.ModelTypeChat
	case "completion":
		return platform.ModelTypeCompletion
	case "responses":
		return platform.ModelTypeResponses
	case "embedding":
		return platform.ModelTypeEmbedding
	case "realtime":
		return platform.ModelTypeRealtime
	case "image":
		return platform.ModelTypeImageGeneration
	case "video":
		return platform.ModelTypeVideoGeneration
	case "text_to_speech":
		return platform.ModelTypeTextToSpeech
	case "audio_transcription":
		return platform.ModelTypeAudioTranscription
	case "audio_translation":
		return platform.ModelTypeAudioTranslation
	case "moderation":
		return platform.ModelTypeModeration
	case "rerank":
		return platform.ModelTypeRerank
	case "ocr":
		return platform.ModelTypeOCR
	default:
		return ""
	}
}

func inferModelTypesFromCapabilities(capabilities []string) []platform.ModelType {
	types := make([]platform.ModelType, 0, len(capabilities))
	for _, capability := range capabilities {
		mt := mapCapabilityToModelType(capability)
		if mt == "" {
			continue
		}

		types = append(types, mt)
	}

	return normalizeModelTypes(types)
}

func mapCapabilityToModelType(capability string) platform.ModelType {
	switch strings.TrimSpace(strings.ToLower(capability)) {
	case "chat", "reasoning", "agent":
		return platform.ModelTypeChat
	case "completion", "text":
		return platform.ModelTypeCompletion
	case "embedding":
		return platform.ModelTypeEmbedding
	case "realtime":
		return platform.ModelTypeRealtime
	case "vision", "image_generation":
		return platform.ModelTypeImageGeneration
	case "audio", "speech", "tts", "text_to_speech":
		return platform.ModelTypeTextToSpeech
	case "transcription", "audio_transcription":
		return platform.ModelTypeAudioTranscription
	case "translation", "audio_translation":
		return platform.ModelTypeAudioTranslation
	case "moderation":
		return platform.ModelTypeModeration
	case "rerank":
		return platform.ModelTypeRerank
	case "search":
		return platform.ModelTypeSearch
	case "ocr", "bbox_extraction", "annotations":
		return platform.ModelTypeOCR
	default:
		return ""
	}
}

func normalizeProviders(items []platform.ProviderCatalogProvider) []platform.ProviderCatalogProvider {
	out := make([]platform.ProviderCatalogProvider, 0, len(items))
	for _, item := range items {
		item.Provider = normalizeProviderSlug(item.Provider)

		item.DisplayName = strings.TrimSpace(item.DisplayName)
		if item.Provider == "" {
			continue
		}

		if item.Provider == "azure" {
			continue
		}

		if item.DisplayName == "" {
			item.DisplayName = item.Provider
		}

		item.Description = strings.TrimSpace(item.Description)

		item.DocsURL = strings.TrimSpace(item.DocsURL)
		if len(item.AuthModes) == 0 {
			item.AuthModes = []platform.ProviderAuthMode{platform.ProviderAuthModeAPIKey}
		}

		out = append(out, item)
	}

	return out
}

func normalizeModelTypes(types []platform.ModelType) []platform.ModelType {
	if len(types) == 0 {
		return nil
	}

	out := make([]platform.ModelType, 0, len(types))

	seen := make(map[platform.ModelType]struct{}, len(types))
	for _, item := range types {
		normalized := platform.ModelType(strings.TrimSpace(strings.ToLower(string(item))))
		if normalized == "" {
			continue
		}

		if _, ok := seen[normalized]; ok {
			continue
		}

		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}

	return out
}

func normalizeStringList(values []string) []string {
	if len(values) == 0 {
		return nil
	}

	out := make([]string, 0, len(values))

	seen := make(map[string]struct{}, len(values))
	for _, item := range values {
		normalized := strings.TrimSpace(strings.ToLower(item))
		if normalized == "" {
			continue
		}

		if _, ok := seen[normalized]; ok {
			continue
		}

		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}

	return out
}

func cloneModelPrice(price *platform.ModelPrice) *platform.ModelPrice {
	if price == nil {
		return nil
	}

	out := *price
	if len(price.Rates) > 0 {
		out.Rates = append([]platform.ModelPricingRate(nil), price.Rates...)
		for i := range out.Rates {
			out.Rates[i].Condition = cloneInterfaceMap(out.Rates[i].Condition)
		}
	}

	return &out
}

func cloneInterfaceMap(input map[string]interface{}) map[string]interface{} {
	if len(input) == 0 {
		return nil
	}

	out := make(map[string]interface{}, len(input))
	for key, value := range input {
		out[key] = value
	}

	return out
}

func pricingRateKey(
	dimension string,
	unit string,
	region string,
	condition map[string]interface{},
	effectiveFrom *time.Time,
	effectiveTo *time.Time,
) string {
	payload := struct {
		Dimension     string                 `json:"dimension"`
		Unit          string                 `json:"unit"`
		Region        string                 `json:"region"`
		Condition     map[string]interface{} `json:"condition,omitempty"`
		EffectiveFrom *time.Time             `json:"effective_from,omitempty"`
		EffectiveTo   *time.Time             `json:"effective_to,omitempty"`
	}{
		Dimension:     strings.TrimSpace(dimension),
		Unit:          strings.TrimSpace(unit),
		Region:        strings.TrimSpace(region),
		Condition:     condition,
		EffectiveFrom: effectiveFrom,
		EffectiveTo:   effectiveTo,
	}
	raw, _ := json.Marshal(payload) //nolint:errcheck // The value is composed exclusively of JSON-safe fields.
	sum := sha256.Sum256(raw)

	return hex.EncodeToString(sum[:])
}

func catalogCapabilityReadiness(
	model platform.ProviderCatalogModel,
	status string,
) []platform.CapabilityReadiness {
	state := platform.CapabilityReadinessContractReady
	reason := ""

	if strings.TrimSpace(status) != "" && !strings.EqualFold(strings.TrimSpace(status), "active") {
		state = platform.CapabilityReadinessBlockedLifecycle
		reason = "model_lifecycle_not_active"
	}

	result := make([]platform.CapabilityReadiness, 0, len(model.Modes))
	for _, mode := range model.Modes {
		transport := "http"
		if strings.EqualFold(mode, "realtime") {
			transport = "websocket"
		}

		result = append(result, platform.CapabilityReadiness{
			Endpoint:      mode,
			Transport:     transport,
			State:         state,
			Reason:        reason,
			LiveValidated: false,
		})
	}

	return result
}

func fillModelPricingFromRegistry(provider string, model string, current *platform.ModelPrice) *platform.ModelPrice {
	if hasPositiveModelPrice(current, true) && hasPositiveModelPrice(current, false) {
		return current
	}

	pricing, _ := billing.GetPricing(provider, model)
	if pricing.InputPrice <= 0 && pricing.OutputPrice <= 0 {
		return current
	}

	out := &platform.ModelPrice{}
	if current != nil {
		*out = *current
	}

	assignModelPriceIfMissing(out, true, pricing.InputPrice)
	assignModelPriceIfMissing(out, false, pricing.OutputPrice)

	if out.InputUSDPer1M == nil && out.OutputUSDPer1M == nil {
		return nil
	}

	return out
}

func hasPositiveModelPrice(price *platform.ModelPrice, input bool) bool {
	if price == nil {
		return false
	}

	value := price.OutputUSDPer1M
	if input {
		value = price.InputUSDPer1M
	}

	return value != nil && *value > 0
}

func assignModelPriceIfMissing(price *platform.ModelPrice, input bool, value float64) {
	if value <= 0 {
		return
	}

	target := &price.OutputUSDPer1M
	if input {
		target = &price.InputUSDPer1M
	}

	if *target != nil && **target > 0 {
		return
	}

	v := value
	*target = &v
}
