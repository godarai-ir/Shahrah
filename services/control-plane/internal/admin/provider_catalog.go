package admin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
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
	Slug                    string `json:"slug"`
	Name                    string `json:"name"`
	LogoURL                 string `json:"logoUrl"`
	Enabled                 bool   `json:"enabled"`
	SupportsByok            bool   `json:"supportsByok"`
	SupportsRegionalRouting bool   `json:"supportsRegionalRouting"`
}

type modelDiscoveryPricing struct {
	InputPerMillionTokens  *float64 `json:"inputPerMillionTokens"`
	OutputPerMillionTokens *float64 `json:"outputPerMillionTokens"`
}

type modelDiscoveryModel struct {
	CanonicalID      string                 `json:"canonicalId"`
	Name             string                 `json:"name"`
	ContextWindow    *int                   `json:"contextWindow"`
	Capabilities     []string               `json:"capabilities"`
	InputModalities  []string               `json:"inputModalities"`
	OutputModalities []string               `json:"outputModalities"`
	Pricing          *modelDiscoveryPricing `json:"pricing"`
}

func normalizeProviderSlug(value string) string {
	normalized := strings.TrimSpace(strings.ToLower(value))
	switch normalized {
	case "perplexity":
		return "perplexity-ai"
	case "zai":
		return "xai"
	default:
		return normalized
	}
}

func discoveryProviderExists(ctx context.Context, provider string) (bool, error) {
	items, err := discoveryProviders(ctx)
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

func discoveryProviders(ctx context.Context) ([]platform.ProviderCatalogProvider, error) {
	if cached, ok := getCachedProviders(); ok {
		atomic.AddUint64(&discoveryTelemetry.providersCacheHit, 1)
		return cached, nil
	}
	atomic.AddUint64(&discoveryTelemetry.providersCacheMiss, 1)
	baseURL := getDiscoveryBaseURL()
	if baseURL == "" {
		return nil, fmt.Errorf("%w: SHAHRAH_MODEL_DISCOVERY_URL is not configured; retry after restoring model-discovery connectivity", errDiscoveryUnavailable)
	}
	body, status, err := discoveryGET(ctx, baseURL+"/providers")
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

func discoveryProviderModels(ctx context.Context, provider string) ([]platform.ProviderCatalogModel, error) {
	normalizedProvider := normalizeProviderSlug(provider)
	if normalizedProvider == "" {
		return nil, fmt.Errorf("provider is required")
	}
	if cached, ok := getCachedProviderModels(normalizedProvider); ok {
		atomic.AddUint64(&discoveryTelemetry.modelsCacheHit, 1)
		return cached, nil
	}
	atomic.AddUint64(&discoveryTelemetry.modelsCacheMiss, 1)
	baseURL := getDiscoveryBaseURL()
	if baseURL == "" {
		return nil, fmt.Errorf("%w: SHAHRAH_MODEL_DISCOVERY_URL is not configured; retry after restoring model-discovery connectivity", errDiscoveryUnavailable)
	}
	body, status, err := discoveryGET(ctx, baseURL+"/providers/"+url.PathEscape(normalizedProvider)+"/models")
	if err == nil && status == http.StatusNotFound {
		items, queryErr := discoveryProviderModelsByQuery(ctx, baseURL, normalizedProvider)
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

func discoveryProviderModelsByQuery(ctx context.Context, baseURL string, provider string) ([]platform.ProviderCatalogModel, error) {
	const pageSize = 100
	const maxPages = 20

	items := make([]platform.ProviderCatalogModel, 0, pageSize)
	for page := 1; page <= maxPages; page++ {
		queryURL := fmt.Sprintf("%s/models?provider=%s&limit=%d&page=%d", baseURL, url.QueryEscape(provider), pageSize, page)
		body, status, err := discoveryGET(ctx, queryURL)
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

func getCachedProviders() ([]platform.ProviderCatalogProvider, bool) {
	ttl := discoveryCacheTTL()
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

func getCachedProviderModels(provider string) ([]platform.ProviderCatalogModel, bool) {
	ttl := discoveryCacheTTL()
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

func discoveryCacheTTL() time.Duration {
	raw := strings.TrimSpace(os.Getenv("SHAHRAH_DISCOVERY_CACHE_TTL_SEC"))
	if raw == "" {
		return 5 * time.Minute
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return 5 * time.Minute
	}
	if n == 0 {
		return 0
	}
	return time.Duration(n) * time.Second
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

func resolveProviderAccountModelSelections(ctx context.Context, provider string, selections []platform.UpsertTenantProviderAccountModelSelection) ([]platform.TenantProviderAccountModel, []string, error) {
	catalog, err := discoveryProviderModels(ctx, provider)
	if err != nil {
		return nil, nil, err
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
		pricingMode := item.PricingMode
		if pricingMode == "" {
			pricingMode = platform.PricingModePublic
		}
		resolved = append(resolved, platform.TenantProviderAccountModel{
			Model:            model,
			Enabled:          true,
			Types:            normalizeModelTypes(cm.Types),
			InputModalities:  normalizeStringList(cm.InputModalities),
			OutputModalities: normalizeStringList(cm.OutputModalities),
			PricingMode:      pricingMode,
			PublicPricing:    cloneModelPrice(cm.PublicPricing),
			PrivatePricing:   cloneModelPrice(item.PrivatePricing),
			Settings:         map[string]interface{}{},
		})
	}
	sort.Strings(invalid)
	return resolved, invalid, nil
}

func getDiscoveryBaseURL() string {
	baseURL := strings.TrimSpace(os.Getenv("SHAHRAH_MODEL_DISCOVERY_URL"))
	if baseURL == "" {
		baseURL = strings.TrimSpace(os.Getenv("SHAHRAH_DISCOVERY_URL"))
	}
	if baseURL == "" {
		return ""
	}
	baseURL = strings.TrimSuffix(baseURL, "/")
	return baseURL
}

func discoveryGET(ctx context.Context, rawURL string) ([]byte, int, error) {
	body, status, _, err := discoveryFetch(ctx, rawURL)
	return body, status, err
}

func discoveryFetch(ctx context.Context, rawURL string) ([]byte, int, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, 0, "", err
	}
	req.Header.Set("Accept", "application/json")

	token := strings.TrimSpace(os.Getenv("SHAHRAH_DISCOVERY_TOKEN"))
	if token == "" {
		token = strings.TrimSpace(os.Getenv("MASTER_KEY"))
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: discoveryTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, "", err
	}
	return body, resp.StatusCode, resp.Header.Get("Content-Type"), nil
}

func decodeProvidersPayload(body []byte) ([]platform.ProviderCatalogProvider, error) {
	if len(body) == 0 {
		return nil, fmt.Errorf("discovery providers payload is empty")
	}

	var direct []platform.ProviderCatalogProvider
	if err := json.Unmarshal(body, &direct); err == nil && hasCanonicalProviderEntries(direct) {
		return direct, nil
	}

	var mdDirect []discoveryProvider
	if err := json.Unmarshal(body, &mdDirect); err == nil && len(mdDirect) > 0 {
		return mapDiscoveryProviders(mdDirect), nil
	}

	var wrapped struct {
		Data      []platform.ProviderCatalogProvider `json:"data"`
		Providers []platform.ProviderCatalogProvider `json:"providers"`
	}
	if err := json.Unmarshal(body, &wrapped); err == nil {
		if hasCanonicalProviderEntries(wrapped.Data) {
			return wrapped.Data, nil
		}
		if hasCanonicalProviderEntries(wrapped.Providers) {
			return wrapped.Providers, nil
		}
	}

	var mdWrapped struct {
		Data []discoveryProvider `json:"data"`
	}
	if err := json.Unmarshal(body, &mdWrapped); err == nil && len(mdWrapped.Data) > 0 {
		return mapDiscoveryProviders(mdWrapped.Data), nil
	}
	return nil, fmt.Errorf("discovery providers payload does not contain providers")
}

func decodeProviderModelsPayload(body []byte) ([]platform.ProviderCatalogModel, error) {
	if len(body) == 0 {
		return nil, fmt.Errorf("discovery provider models payload is empty")
	}

	var direct []platform.ProviderCatalogModel
	if err := json.Unmarshal(body, &direct); err == nil && hasCanonicalModelEntries(direct) {
		return direct, nil
	}

	var mdDirect []modelDiscoveryModel
	if err := json.Unmarshal(body, &mdDirect); err == nil && len(mdDirect) > 0 {
		return mapDiscoveryModels(mdDirect), nil
	}

	var wrapped struct {
		Data   []platform.ProviderCatalogModel `json:"data"`
		Models []platform.ProviderCatalogModel `json:"models"`
	}
	if err := json.Unmarshal(body, &wrapped); err == nil {
		if hasCanonicalModelEntries(wrapped.Data) {
			return wrapped.Data, nil
		}
		if hasCanonicalModelEntries(wrapped.Models) {
			return wrapped.Models, nil
		}
	}

	var modelsWrapped struct {
		Data []modelDiscoveryModel `json:"data"`
	}
	if err := json.Unmarshal(body, &modelsWrapped); err == nil && len(modelsWrapped.Data) > 0 {
		return mapDiscoveryModels(modelsWrapped.Data), nil
	}
	return nil, fmt.Errorf("discovery provider models payload does not contain models")
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
		out = append(out, platform.ProviderCatalogProvider{
			Provider:                provider,
			DisplayName:             displayName,
			LogoURL:                 strings.TrimSpace(item.LogoURL),
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

		types := make([]platform.ModelType, 0, len(item.Capabilities))
		for _, capability := range item.Capabilities {
			mt := mapCapabilityToModelType(capability)
			if mt == "" {
				continue
			}
			types = append(types, mt)
		}

		if len(types) == 0 {
			for _, modality := range append(item.InputModalities, item.OutputModalities...) {
				mt := mapModalityToModelType(modality)
				if mt == "" {
					continue
				}
				types = append(types, mt)
			}
		}

		entry := platform.ProviderCatalogModel{
			Model:            model,
			DisplayName:      displayName,
			Types:            normalizeModelTypes(types),
			ContextWindow:    item.ContextWindow,
			InputModalities:  normalizeStringList(item.InputModalities),
			OutputModalities: normalizeStringList(item.OutputModalities),
			Features:         normalizeStringList(item.Capabilities),
		}
		if item.Pricing != nil {
			entry.PublicPricing = &platform.ModelPrice{
				InputUSDPer1M:  item.Pricing.InputPerMillionTokens,
				OutputUSDPer1M: item.Pricing.OutputPerMillionTokens,
			}
		}
		out = append(out, entry)
	}
	return out
}

func discoveryDisplaySettings(ctx context.Context) (*platform.ProviderCatalogDisplaySettings, error) {
	baseURL := getDiscoveryBaseURL()
	if baseURL == "" {
		return nil, errDiscoveryNotFound
	}
	body, status, err := discoveryGET(ctx, baseURL+"/settings/display")
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
	case "image", "image_generation", "vision":
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
	default:
		return ""
	}
}

func mapModalityToModelType(modality string) platform.ModelType {
	switch strings.TrimSpace(strings.ToLower(modality)) {
	case "text":
		return platform.ModelTypeChat
	case "image":
		return platform.ModelTypeImageGeneration
	case "audio":
		return platform.ModelTypeAudioTranscription
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
	return &out
}

func fillModelPricingFromRegistry(provider string, model string, current *platform.ModelPrice) *platform.ModelPrice {
	hasInput := current != nil && current.InputUSDPer1M != nil && *current.InputUSDPer1M > 0
	hasOutput := current != nil && current.OutputUSDPer1M != nil && *current.OutputUSDPer1M > 0
	if hasInput && hasOutput {
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
	if (out.InputUSDPer1M == nil || *out.InputUSDPer1M <= 0) && pricing.InputPrice > 0 {
		v := pricing.InputPrice
		out.InputUSDPer1M = &v
	}
	if (out.OutputUSDPer1M == nil || *out.OutputUSDPer1M <= 0) && pricing.OutputPrice > 0 {
		v := pricing.OutputPrice
		out.OutputUSDPer1M = &v
	}
	if out.InputUSDPer1M == nil && out.OutputUSDPer1M == nil {
		return nil
	}
	return out
}
