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

const discoveryTimeout = 8 * time.Second

type providerCatalogCache struct {
	mu         sync.RWMutex
	providers  []platform.ProviderCatalogProvider
	providersAt time.Time
	models     map[string]cachedModels
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

func discoveryProviderExists(ctx context.Context, provider string) (bool, error) {
	items, err := discoveryProviders(ctx)
	if err != nil {
		return false, err
	}
	normalized := strings.TrimSpace(strings.ToLower(provider))
	for _, item := range items {
		if strings.TrimSpace(strings.ToLower(item.Provider)) == normalized {
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
		items := localProviderCatalog()
		setCachedProviders(items)
		return items, nil
	}
	body, status, err := discoveryGET(ctx, baseURL+"/providers")
	if err != nil {
		atomic.AddUint64(&discoveryTelemetry.discoveryError, 1)
		atomic.AddUint64(&discoveryTelemetry.discoveryFallback, 1)
		items := localProviderCatalog()
		setCachedProviders(items)
		return items, nil
	}
	if status == http.StatusNotFound {
		atomic.AddUint64(&discoveryTelemetry.discoveryFallback, 1)
		items := localProviderCatalog()
		setCachedProviders(items)
		return items, nil
	}
	if status >= http.StatusBadRequest {
		atomic.AddUint64(&discoveryTelemetry.discoveryError, 1)
		atomic.AddUint64(&discoveryTelemetry.discoveryFallback, 1)
		items := localProviderCatalog()
		setCachedProviders(items)
		return items, nil
	}

	items, err := decodeProvidersPayload(body)
	if err != nil {
		atomic.AddUint64(&discoveryTelemetry.discoveryError, 1)
		atomic.AddUint64(&discoveryTelemetry.discoveryFallback, 1)
		items := localProviderCatalog()
		setCachedProviders(items)
		return items, nil
	}
	normalized := normalizeProviders(items)
	if len(normalized) == 0 {
		atomic.AddUint64(&discoveryTelemetry.discoveryFallback, 1)
		items := localProviderCatalog()
		setCachedProviders(items)
		return items, nil
	}
	setCachedProviders(normalized)
	return normalized, nil
}

func discoveryProviderModels(ctx context.Context, provider string) ([]platform.ProviderCatalogModel, error) {
	normalizedProvider := strings.TrimSpace(strings.ToLower(provider))
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
		items, err := localProviderModels(normalizedProvider)
		if err == nil {
			setCachedProviderModels(normalizedProvider, items)
		}
		return items, err
	}
	body, status, err := discoveryGET(ctx, baseURL+"/providers/"+url.PathEscape(normalizedProvider)+"/models")
	if err != nil {
		atomic.AddUint64(&discoveryTelemetry.discoveryError, 1)
		atomic.AddUint64(&discoveryTelemetry.discoveryFallback, 1)
		items, errLocal := localProviderModels(normalizedProvider)
		if errLocal == nil {
			setCachedProviderModels(normalizedProvider, items)
		}
		return items, errLocal
	}
	if status == http.StatusNotFound {
		atomic.AddUint64(&discoveryTelemetry.discoveryFallback, 1)
		items, errLocal := localProviderModels(normalizedProvider)
		if errLocal == nil {
			setCachedProviderModels(normalizedProvider, items)
		}
		return items, errLocal
	}
	if status >= http.StatusBadRequest {
		atomic.AddUint64(&discoveryTelemetry.discoveryError, 1)
		atomic.AddUint64(&discoveryTelemetry.discoveryFallback, 1)
		items, errLocal := localProviderModels(normalizedProvider)
		if errLocal == nil {
			setCachedProviderModels(normalizedProvider, items)
		}
		return items, errLocal
	}

	items, err := decodeProviderModelsPayload(body)
	if err != nil {
		atomic.AddUint64(&discoveryTelemetry.discoveryError, 1)
		atomic.AddUint64(&discoveryTelemetry.discoveryFallback, 1)
		items, errLocal := localProviderModels(normalizedProvider)
		if errLocal == nil {
			setCachedProviderModels(normalizedProvider, items)
		}
		return items, errLocal
	}
	for i := range items {
		items[i].Provider = normalizedProvider
		if items[i].Model == "" {
			atomic.AddUint64(&discoveryTelemetry.discoveryFallback, 1)
			items, errLocal := localProviderModels(normalizedProvider)
			if errLocal == nil {
				setCachedProviderModels(normalizedProvider, items)
			}
			return items, errLocal
		}
		items[i].Types = normalizeModelTypes(items[i].Types)
		items[i].PublicPricing = fillModelPricingFromRegistry(normalizedProvider, items[i].Model, items[i].PublicPricing)
	}
	setCachedProviderModels(normalizedProvider, items)
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

func validateProviderModelsFromCatalog(ctx context.Context, provider string, models []platform.TenantProviderAccountModel) ([]string, []string, error) {
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
	invalidType := make([]string, 0)
	seen := make(map[string]struct{})
	seenType := make(map[string]struct{})
	for _, item := range models {
		model := strings.TrimSpace(item.Model)
		if model == "" {
			continue
		}
		cm, ok := allowed[model]
		if !ok {
			if _, ok := seen[model]; !ok {
				seen[model] = struct{}{}
				invalid = append(invalid, model)
			}
			continue
		}
		if len(item.Types) > 0 && len(cm.Types) > 0 {
			allowedTypes := map[platform.ModelType]struct{}{}
			for _, t := range cm.Types {
				allowedTypes[platform.ModelType(strings.TrimSpace(strings.ToLower(string(t))))] = struct{}{}
			}
			for _, mt := range item.Types {
				normalizedMT := platform.ModelType(strings.TrimSpace(strings.ToLower(string(mt))))
				if normalizedMT == "" {
					continue
				}
				if _, ok := allowedTypes[normalizedMT]; !ok {
					key := model + ":" + string(normalizedMT)
					if _, dup := seenType[key]; !dup {
						seenType[key] = struct{}{}
						invalidType = append(invalidType, key)
					}
				}
			}
		}
	}
	sort.Strings(invalid)
	sort.Strings(invalidType)
	return invalid, invalidType, nil
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
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, 0, err
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
		return nil, 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return body, resp.StatusCode, nil
}

func decodeProvidersPayload(body []byte) ([]platform.ProviderCatalogProvider, error) {
	if len(body) == 0 {
		return nil, fmt.Errorf("discovery providers payload is empty")
	}

	var direct []platform.ProviderCatalogProvider
	if err := json.Unmarshal(body, &direct); err == nil && len(direct) > 0 {
		return direct, nil
	}

	var wrapped struct {
		Data      []platform.ProviderCatalogProvider `json:"data"`
		Providers []platform.ProviderCatalogProvider `json:"providers"`
	}
	if err := json.Unmarshal(body, &wrapped); err != nil {
		return nil, fmt.Errorf("decode discovery providers payload: %w", err)
	}
	if len(wrapped.Data) > 0 {
		return wrapped.Data, nil
	}
	if len(wrapped.Providers) > 0 {
		return wrapped.Providers, nil
	}
	return nil, fmt.Errorf("discovery providers payload does not contain providers")
}

func decodeProviderModelsPayload(body []byte) ([]platform.ProviderCatalogModel, error) {
	if len(body) == 0 {
		return nil, fmt.Errorf("discovery provider models payload is empty")
	}

	var direct []platform.ProviderCatalogModel
	if err := json.Unmarshal(body, &direct); err == nil && len(direct) > 0 {
		return direct, nil
	}

	var wrapped struct {
		Data   []platform.ProviderCatalogModel `json:"data"`
		Models []platform.ProviderCatalogModel `json:"models"`
	}
	if err := json.Unmarshal(body, &wrapped); err != nil {
		return nil, fmt.Errorf("decode discovery provider models payload: %w", err)
	}
	if len(wrapped.Data) > 0 {
		return wrapped.Data, nil
	}
	if len(wrapped.Models) > 0 {
		return wrapped.Models, nil
	}
	return nil, fmt.Errorf("discovery provider models payload does not contain models")
}

func normalizeProviders(items []platform.ProviderCatalogProvider) []platform.ProviderCatalogProvider {
	out := make([]platform.ProviderCatalogProvider, 0, len(items))
	for _, item := range items {
		item.Provider = strings.TrimSpace(strings.ToLower(item.Provider))
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

func localProviderCatalog() []platform.ProviderCatalogProvider {
	return []platform.ProviderCatalogProvider{
		{
			Provider:                "openai",
			DisplayName:             "OpenAI",
			Enabled:                 true,
			AuthModes:               []platform.ProviderAuthMode{platform.ProviderAuthModeAPIKey},
			SupportsBringYourOwnKey: true,
			SupportsVirtualKeys:     true,
		},
		{
			Provider:                "anthropic",
			DisplayName:             "Anthropic",
			Enabled:                 true,
			AuthModes:               []platform.ProviderAuthMode{platform.ProviderAuthModeAPIKey},
			SupportsBringYourOwnKey: true,
			SupportsVirtualKeys:     true,
		},
		{
			Provider:                "google",
			DisplayName:             "Google",
			Enabled:                 true,
			AuthModes:               []platform.ProviderAuthMode{platform.ProviderAuthModeAPIKey, platform.ProviderAuthModeGoogleServiceAccount},
			SupportsBringYourOwnKey: true,
			SupportsVirtualKeys:     true,
		},
		{
			Provider:                "bedrock",
			DisplayName:             "AWS Bedrock",
			Enabled:                 true,
			AuthModes:               []platform.ProviderAuthMode{platform.ProviderAuthModeAWSSigV4},
			SupportsBringYourOwnKey: true,
			SupportsVirtualKeys:     true,
		},
		{
			Provider:                "gemini",
			DisplayName:             "Google Gemini",
			Enabled:                 true,
			AuthModes:               []platform.ProviderAuthMode{platform.ProviderAuthModeAPIKey, platform.ProviderAuthModeGoogleServiceAccount},
			SupportsBringYourOwnKey: true,
			SupportsVirtualKeys:     true,
		},
		{
			Provider:                "vertexai",
			DisplayName:             "Google Vertex AI",
			Enabled:                 true,
			AuthModes:               []platform.ProviderAuthMode{platform.ProviderAuthModeGoogleServiceAccount},
			SupportsBringYourOwnKey: true,
			SupportsVirtualKeys:     true,
		},
		{
			Provider:                "groq",
			DisplayName:             "Groq",
			Enabled:                 true,
			AuthModes:               []platform.ProviderAuthMode{platform.ProviderAuthModeAPIKey},
			SupportsBringYourOwnKey: true,
			SupportsVirtualKeys:     true,
		},
		{
			Provider:                "meta",
			DisplayName:             "Meta",
			Enabled:                 true,
			AuthModes:               []platform.ProviderAuthMode{platform.ProviderAuthModeAPIKey},
			SupportsBringYourOwnKey: true,
			SupportsVirtualKeys:     true,
		},
		{
			Provider:                "mistral",
			DisplayName:             "Mistral",
			Enabled:                 true,
			AuthModes:               []platform.ProviderAuthMode{platform.ProviderAuthModeAPIKey},
			SupportsBringYourOwnKey: true,
			SupportsVirtualKeys:     true,
		},
		{
			Provider:                "xai",
			DisplayName:             "xAI",
			Enabled:                 true,
			AuthModes:               []platform.ProviderAuthMode{platform.ProviderAuthModeAPIKey},
			SupportsBringYourOwnKey: true,
			SupportsVirtualKeys:     true,
		},
		{
			Provider:                "togetherai",
			DisplayName:             "Together AI",
			Enabled:                 true,
			AuthModes:               []platform.ProviderAuthMode{platform.ProviderAuthModeAPIKey},
			SupportsBringYourOwnKey: true,
			SupportsVirtualKeys:     true,
		},
		{
			Provider:                "cohere",
			DisplayName:             "Cohere",
			Enabled:                 true,
			AuthModes:               []platform.ProviderAuthMode{platform.ProviderAuthModeAPIKey},
			SupportsBringYourOwnKey: true,
			SupportsVirtualKeys:     true,
		},
		{
			Provider:                "stability",
			DisplayName:             "Stability AI",
			Enabled:                 true,
			AuthModes:               []platform.ProviderAuthMode{platform.ProviderAuthModeAPIKey},
			SupportsBringYourOwnKey: true,
			SupportsVirtualKeys:     true,
		},
		{
			Provider:                "deepseek",
			DisplayName:             "DeepSeek",
			Enabled:                 true,
			AuthModes:               []platform.ProviderAuthMode{platform.ProviderAuthModeAPIKey},
			SupportsBringYourOwnKey: true,
			SupportsVirtualKeys:     true,
		},
		{
			Provider:                "alibaba",
			DisplayName:             "Alibaba",
			Enabled:                 true,
			AuthModes:               []platform.ProviderAuthMode{platform.ProviderAuthModeAPIKey},
			SupportsBringYourOwnKey: true,
			SupportsVirtualKeys:     true,
		},
		{
			Provider:                "bfl",
			DisplayName:             "BFL (FLUX)",
			Enabled:                 true,
			AuthModes:               []platform.ProviderAuthMode{platform.ProviderAuthModeAPIKey},
			SupportsBringYourOwnKey: true,
			SupportsVirtualKeys:     true,
		},
		{
			Provider:                "cloudflare",
			DisplayName:             "Cloudflare",
			Enabled:                 true,
			AuthModes:               []platform.ProviderAuthMode{platform.ProviderAuthModeAPIKey},
			SupportsBringYourOwnKey: true,
			SupportsVirtualKeys:     true,
		},
		{
			Provider:                "byteplus",
			DisplayName:             "BytePlus",
			Enabled:                 true,
			AuthModes:               []platform.ProviderAuthMode{platform.ProviderAuthModeAPIKey},
			SupportsBringYourOwnKey: true,
			SupportsVirtualKeys:     true,
		},
		{
			Provider:                "zai",
			DisplayName:             "Z.AI",
			Enabled:                 true,
			AuthModes:               []platform.ProviderAuthMode{platform.ProviderAuthModeAPIKey},
			SupportsBringYourOwnKey: true,
			SupportsVirtualKeys:     true,
		},
		{
			Provider:                "tavily",
			DisplayName:             "Tavily",
			Enabled:                 true,
			AuthModes:               []platform.ProviderAuthMode{platform.ProviderAuthModeAPIKey},
			SupportsBringYourOwnKey: true,
			SupportsVirtualKeys:     true,
		},
		{
			Provider:                "parallelai",
			DisplayName:             "Parallel AI",
			Enabled:                 true,
			AuthModes:               []platform.ProviderAuthMode{platform.ProviderAuthModeAPIKey},
			SupportsBringYourOwnKey: true,
			SupportsVirtualKeys:     true,
		},
		{
			Provider:                "exa",
			DisplayName:             "Exa AI",
			Enabled:                 true,
			AuthModes:               []platform.ProviderAuthMode{platform.ProviderAuthModeAPIKey},
			SupportsBringYourOwnKey: true,
			SupportsVirtualKeys:     true,
		},
		{
			Provider:                "dataforseo",
			DisplayName:             "DataForSEO",
			Enabled:                 true,
			AuthModes:               []platform.ProviderAuthMode{platform.ProviderAuthModeAPIKey},
			SupportsBringYourOwnKey: true,
			SupportsVirtualKeys:     true,
		},
		{
			Provider:                "perplexity",
			DisplayName:             "Perplexity",
			Enabled:                 true,
			AuthModes:               []platform.ProviderAuthMode{platform.ProviderAuthModeAPIKey},
			SupportsBringYourOwnKey: true,
			SupportsVirtualKeys:     true,
		},
		{
			Provider:                "google_pse",
			DisplayName:             "Google PSE",
			Enabled:                 true,
			AuthModes:               []platform.ProviderAuthMode{platform.ProviderAuthModeAPIKey},
			SupportsBringYourOwnKey: true,
			SupportsVirtualKeys:     true,
		},
		{
			Provider:                "firecrawl",
			DisplayName:             "Firecrawl",
			Enabled:                 true,
			AuthModes:               []platform.ProviderAuthMode{platform.ProviderAuthModeAPIKey},
			SupportsBringYourOwnKey: true,
			SupportsVirtualKeys:     true,
		},
		{
			Provider:                "moonshot",
			DisplayName:             "Moonshot.ai",
			Enabled:                 true,
			AuthModes:               []platform.ProviderAuthMode{platform.ProviderAuthModeAPIKey},
			SupportsBringYourOwnKey: true,
			SupportsVirtualKeys:     true,
		},
		{
			Provider:                "nvidia_nim",
			DisplayName:             "Nvidia NIM",
			Enabled:                 true,
			AuthModes:               []platform.ProviderAuthMode{platform.ProviderAuthModeAPIKey},
			SupportsBringYourOwnKey: true,
			SupportsVirtualKeys:     true,
		},
		{
			Provider:                "minimax",
			DisplayName:             "MiniMax",
			Enabled:                 true,
			AuthModes:               []platform.ProviderAuthMode{platform.ProviderAuthModeAPIKey},
			SupportsBringYourOwnKey: true,
			SupportsVirtualKeys:     true,
		},
	}
}

func localProviderModels(provider string) ([]platform.ProviderCatalogModel, error) {
	if items, ok := builtinProviderModelCatalog(provider); ok {
		return items, nil
	}
	items := make([]platform.ProviderCatalogModel, 0)
	for _, pricing := range billing.PricingRegistry {
		if strings.TrimSpace(strings.ToLower(pricing.Provider)) != provider {
			continue
		}
		input := pricing.InputPrice
		output := pricing.OutputPrice
		items = append(items, platform.ProviderCatalogModel{
			Provider:    provider,
			Model:       pricing.Model,
			DisplayName: pricing.Model,
			Types:       []platform.ModelType{platform.ModelTypeChat},
			PublicPricing: &platform.ModelPrice{
				InputUSDPer1M:  &input,
				OutputUSDPer1M: &output,
			},
		})
	}
	if len(items) == 0 {
		return nil, errDiscoveryNotFound
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Model < items[j].Model
	})
	return items, nil
}

func builtinProviderModelCatalog(provider string) ([]platform.ProviderCatalogModel, bool) {
	entries, ok := builtinProviderModels()[provider]
	if !ok || len(entries) == 0 {
		return nil, false
	}
	items := make([]platform.ProviderCatalogModel, 0, len(entries))
	for _, item := range entries {
		model := strings.TrimSpace(item.Model)
		if model == "" {
			continue
		}
		entry := platform.ProviderCatalogModel{
			Provider:    provider,
			Model:       model,
			DisplayName: model,
			Types:       normalizeModelTypes(item.Types),
		}
		if item.InputUSDPer1M != nil || item.OutputUSDPer1M != nil {
			entry.PublicPricing = &platform.ModelPrice{
				InputUSDPer1M:  item.InputUSDPer1M,
				OutputUSDPer1M: item.OutputUSDPer1M,
			}
		}
		entry.PublicPricing = fillModelPricingFromRegistry(provider, model, entry.PublicPricing)
		items = append(items, entry)
	}
	if len(items) == 0 {
		return nil, false
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Model < items[j].Model
	})
	return items, true
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

type builtInModelEntry struct {
	Model         string
	Types         []platform.ModelType
	InputUSDPer1M *float64
	OutputUSDPer1M *float64
}

func builtinProviderModels() map[string][]builtInModelEntry {
	return map[string][]builtInModelEntry{
		"openai": {
			{Model: "gpt-4.1", Types: []platform.ModelType{platform.ModelTypeChat}},
			{Model: "gpt-4.1-mini", Types: []platform.ModelType{platform.ModelTypeChat}},
			{Model: "gpt-4o", Types: []platform.ModelType{platform.ModelTypeChat}},
			{Model: "text-embedding-3-large", Types: []platform.ModelType{platform.ModelTypeEmbedding}},
		},
		"anthropic": {
			{Model: "claude-3-5-sonnet-20240620", Types: []platform.ModelType{platform.ModelTypeChat}},
			{Model: "claude-3-7-sonnet-latest", Types: []platform.ModelType{platform.ModelTypeChat}},
			{Model: "claude-3-haiku-20240307", Types: []platform.ModelType{platform.ModelTypeChat}},
		},
		"google": {
			{Model: "gemini-2.5-pro", Types: []platform.ModelType{platform.ModelTypeChat}},
			{Model: "gemini-2.5-flash", Types: []platform.ModelType{platform.ModelTypeChat}},
			{Model: "text-embedding-004", Types: []platform.ModelType{platform.ModelTypeEmbedding}},
		},
		"gemini": {
			{Model: "gemini-2.5-pro", Types: []platform.ModelType{platform.ModelTypeChat}},
			{Model: "gemini-2.5-flash", Types: []platform.ModelType{platform.ModelTypeChat}},
		},
		"groq": {
			{Model: "llama-3.3-70b-versatile", Types: []platform.ModelType{platform.ModelTypeChat}},
			{Model: "mixtral-8x7b-32768", Types: []platform.ModelType{platform.ModelTypeChat}},
		},
		"meta": {
			{Model: "llama-3.3-70b-instruct", Types: []platform.ModelType{platform.ModelTypeChat}},
			{Model: "llama-3.1-8b-instruct", Types: []platform.ModelType{platform.ModelTypeChat}},
		},
		"mistral": {
			{Model: "mistral-large-latest", Types: []platform.ModelType{platform.ModelTypeChat}},
			{Model: "ministral-8b-latest", Types: []platform.ModelType{platform.ModelTypeChat}},
			{Model: "mistral-embed", Types: []platform.ModelType{platform.ModelTypeEmbedding}},
		},
		"xai": {
			{Model: "grok-2-latest", Types: []platform.ModelType{platform.ModelTypeChat}},
			{Model: "grok-beta", Types: []platform.ModelType{platform.ModelTypeChat}},
		},
		"cohere": {
			{Model: "command-r-plus", Types: []platform.ModelType{platform.ModelTypeChat}},
			{Model: "embed-english-v3.0", Types: []platform.ModelType{platform.ModelTypeEmbedding}},
			{Model: "rerank-v3.5", Types: []platform.ModelType{platform.ModelTypeRerank}},
		},
		"stability": {
			{Model: "stable-image-ultra", Types: []platform.ModelType{platform.ModelTypeImageGeneration}},
			{Model: "stable-diffusion-3.5-large", Types: []platform.ModelType{platform.ModelTypeImageGeneration}},
		},
		"deepseek": {
			{Model: "deepseek-chat", Types: []platform.ModelType{platform.ModelTypeChat}},
			{Model: "deepseek-reasoner", Types: []platform.ModelType{platform.ModelTypeChat}},
		},
		"alibaba": {
			{Model: "qwen-max", Types: []platform.ModelType{platform.ModelTypeChat}},
			{Model: "qwen-plus", Types: []platform.ModelType{platform.ModelTypeChat}},
			{Model: "text-embedding-v3", Types: []platform.ModelType{platform.ModelTypeEmbedding}},
		},
		"bfl": {
			{Model: "flux-pro-1.1", Types: []platform.ModelType{platform.ModelTypeImageGeneration}},
			{Model: "flux-dev", Types: []platform.ModelType{platform.ModelTypeImageGeneration}},
		},
		"cloudflare": {
			{Model: "@cf/meta/llama-3.1-8b-instruct", Types: []platform.ModelType{platform.ModelTypeChat}},
			{Model: "@cf/baai/bge-base-en-v1.5", Types: []platform.ModelType{platform.ModelTypeEmbedding}},
		},
		"byteplus": {
			{Model: "doubao-pro-32k", Types: []platform.ModelType{platform.ModelTypeChat}},
			{Model: "doubao-lite-32k", Types: []platform.ModelType{platform.ModelTypeChat}},
		},
		"zai": {
			{Model: "glm-4-plus", Types: []platform.ModelType{platform.ModelTypeChat}},
			{Model: "embedding-3", Types: []platform.ModelType{platform.ModelTypeEmbedding}},
		},
		"tavily": {
			{Model: "tavily-search", Types: []platform.ModelType{platform.ModelTypeSearch}},
		},
		"parallelai": {
			{Model: "parallel-chat", Types: []platform.ModelType{platform.ModelTypeChat}},
		},
		"exa": {
			{Model: "exa-search", Types: []platform.ModelType{platform.ModelTypeSearch}},
		},
		"dataforseo": {
			{Model: "dataforseo-serp-live", Types: []platform.ModelType{platform.ModelTypeSearch}},
		},
		"perplexity": {
			{Model: "sonar-pro", Types: []platform.ModelType{platform.ModelTypeChat, platform.ModelTypeSearch}},
			{Model: "sonar-small-online", Types: []platform.ModelType{platform.ModelTypeChat, platform.ModelTypeSearch}},
		},
		"google_pse": {
			{Model: "google-custom-search", Types: []platform.ModelType{platform.ModelTypeSearch}},
		},
		"firecrawl": {
			{Model: "firecrawl-search", Types: []platform.ModelType{platform.ModelTypeSearch}},
		},
		"moonshot": {
			{Model: "moonshot-v1-8k", Types: []platform.ModelType{platform.ModelTypeChat}},
			{Model: "moonshot-v1-32k", Types: []platform.ModelType{platform.ModelTypeChat}},
		},
		"nvidia_nim": {
			{Model: "meta/llama-3.1-70b-instruct", Types: []platform.ModelType{platform.ModelTypeChat}},
			{Model: "nvidia/nv-embedqa-e5-v5", Types: []platform.ModelType{platform.ModelTypeEmbedding}},
		},
		"minimax": {
			{Model: "abab6.5-chat", Types: []platform.ModelType{platform.ModelTypeChat}},
			{Model: "minimax-embedding", Types: []platform.ModelType{platform.ModelTypeEmbedding}},
		},
	}
}
