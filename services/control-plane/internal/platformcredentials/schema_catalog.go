package platformcredentials

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

// SchemaCatalog is used only by operator workflows. It is intentionally absent
// from Manager.Resolve so model-discovery can never enter the gateway hot path.
type SchemaCatalog interface {
	Schemas(context.Context, string) ([]platform.ProviderConnectionSchema, error)
}

type schemaCacheEntry struct {
	items     []platform.ProviderConnectionSchema
	expiresAt time.Time
}

type ModelDiscoverySchemaCatalog struct {
	baseURL string
	token   string
	client  *http.Client
	ttl     time.Duration
	mu      sync.RWMutex
	cache   map[string]schemaCacheEntry
}

func NewModelDiscoverySchemaCatalogFromEnv(client *http.Client) (*ModelDiscoverySchemaCatalog, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("SHAHRAH_MODEL_DISCOVERY_URL")), "/")
	if baseURL == "" {
		return nil, fmt.Errorf("model-discovery schema catalog is not configured")
	}
	if !strings.Contains(baseURL, "/api/") {
		baseURL += "/api/v1"
	}
	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("model-discovery schema catalog URL is invalid")
	}
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	return &ModelDiscoverySchemaCatalog{
		baseURL: baseURL,
		token:   strings.TrimSpace(os.Getenv("SHAHRAH_DISCOVERY_TOKEN")),
		client:  client,
		ttl:     5 * time.Minute,
		cache:   make(map[string]schemaCacheEntry),
	}, nil
}

func (c *ModelDiscoverySchemaCatalog) Schemas(ctx context.Context, providerSlug string) ([]platform.ProviderConnectionSchema, error) {
	providerSlug = normalizeProvider(providerSlug)
	if providerSlug == "" {
		return nil, ErrNotFound
	}
	now := time.Now()
	c.mu.RLock()
	entry, ok := c.cache[providerSlug]
	c.mu.RUnlock()
	if ok && now.Before(entry.expiresAt) {
		return cloneSchemas(entry.items), nil
	}

	endpoint := c.baseURL + "/providers/" + url.PathEscape(providerSlug) + "/connection-schemas"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, http.NoBody)
	if err != nil {
		return nil, err
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("model-discovery schema request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("model-discovery schema request returned HTTP %d", resp.StatusCode)
	}
	var payload struct {
		Slug              string `json:"slug"`
		ConnectionSchemas []struct {
			AuthMode     string                             `json:"authMode"`
			Version      int                                `json:"version"`
			RuntimeReady bool                               `json:"runtimeReady"`
			Fields       []platform.ProviderConnectionField `json:"fieldsJson"`
		} `json:"connectionSchemas"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("model-discovery schema response is invalid: %w", err)
	}
	if normalizeProvider(payload.Slug) != providerSlug || len(payload.ConnectionSchemas) == 0 {
		return nil, ErrNotFound
	}
	items := make([]platform.ProviderConnectionSchema, 0, len(payload.ConnectionSchemas))
	for _, item := range payload.ConnectionSchemas {
		items = append(items, platform.ProviderConnectionSchema{
			ProviderSlug: providerSlug,
			AuthMode:     strings.ToLower(strings.TrimSpace(item.AuthMode)),
			Version:      item.Version,
			RuntimeReady: item.RuntimeReady,
			Fields:       item.Fields,
		})
	}
	c.mu.Lock()
	c.cache[providerSlug] = schemaCacheEntry{items: cloneSchemas(items), expiresAt: now.Add(c.ttl)}
	c.mu.Unlock()
	return items, nil
}

func cloneSchemas(items []platform.ProviderConnectionSchema) []platform.ProviderConnectionSchema {
	result := make([]platform.ProviderConnectionSchema, len(items))
	copy(result, items)
	for index := range result {
		result[index].Fields = append([]platform.ProviderConnectionField(nil), items[index].Fields...)
	}
	return result
}
