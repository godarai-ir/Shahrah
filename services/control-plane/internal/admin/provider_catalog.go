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
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

var errDiscoveryNotFound = errors.New("discovery resource not found")

const discoveryTimeout = 8 * time.Second

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
	baseURL, err := getDiscoveryBaseURL()
	if err != nil {
		return nil, err
	}
	body, status, err := discoveryGET(ctx, baseURL+"/providers")
	if err != nil {
		return nil, err
	}
	if status == http.StatusNotFound {
		return nil, errDiscoveryNotFound
	}
	if status >= http.StatusBadRequest {
		return nil, fmt.Errorf("discovery providers failed with status %d", status)
	}

	items, err := decodeProvidersPayload(body)
	if err != nil {
		return nil, err
	}
	return normalizeProviders(items), nil
}

func discoveryProviderModels(ctx context.Context, provider string) ([]platform.ProviderCatalogModel, error) {
	baseURL, err := getDiscoveryBaseURL()
	if err != nil {
		return nil, err
	}
	normalizedProvider := strings.TrimSpace(strings.ToLower(provider))
	if normalizedProvider == "" {
		return nil, fmt.Errorf("provider is required")
	}
	body, status, err := discoveryGET(ctx, baseURL+"/providers/"+url.PathEscape(normalizedProvider)+"/models")
	if err != nil {
		return nil, err
	}
	if status == http.StatusNotFound {
		return nil, errDiscoveryNotFound
	}
	if status >= http.StatusBadRequest {
		return nil, fmt.Errorf("discovery provider models failed with status %d", status)
	}

	items, err := decodeProviderModelsPayload(body)
	if err != nil {
		return nil, err
	}
	for i := range items {
		items[i].Provider = normalizedProvider
		if items[i].Model == "" {
			return nil, fmt.Errorf("discovery payload contains empty model value")
		}
	}
	return items, nil
}

func getDiscoveryBaseURL() (string, error) {
	baseURL := strings.TrimSpace(os.Getenv("SHAHRAH_MODEL_DISCOVERY_URL"))
	if baseURL == "" {
		baseURL = strings.TrimSpace(os.Getenv("SHAHRAH_DISCOVERY_URL"))
	}
	if baseURL == "" {
		return "", fmt.Errorf("model discovery service URL is not configured (set SHAHRAH_MODEL_DISCOVERY_URL)")
	}
	baseURL = strings.TrimSuffix(baseURL, "/")
	return baseURL, nil
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
