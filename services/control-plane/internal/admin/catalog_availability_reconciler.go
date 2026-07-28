package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/modelid"
	"github.com/Delpak/shahrah/services/control-plane/internal/storage"
	"go.uber.org/zap"
)

type CatalogAvailabilityStore interface {
	ListCatalogAvailabilityProviders(ctx context.Context) ([]string, error)
	ReconcileProviderCatalogAvailability(ctx context.Context, provider string, publishedModels []string) (int64, error)
	ReconcileLegacyDeploymentCatalogPricing(ctx context.Context, provider string, models []platform.ProviderCatalogModel) (int64, error)
	ReconcileLegacyDeploymentCatalogAvailability(ctx context.Context, catalog []storage.CatalogModelPublication) (int64, error)
}

func ReconcileCatalogAvailability(ctx context.Context, store CatalogAvailabilityStore) (int, int64, error) {
	providers, err := store.ListCatalogAvailabilityProviders(ctx)
	if err != nil {
		return 0, 0, err
	}
	invalidateDiscoveryCache()
	var changed int64
	for _, provider := range providers {
		models, fetchErr := discoveryProviderModels(ctx, provider)
		if fetchErr != nil {
			return len(providers), changed, fetchErr
		}
		published := make([]string, 0, len(models))
		for _, model := range models {
			published = append(published, model.Model)
		}
		count, reconcileErr := store.ReconcileProviderCatalogAvailability(
			ctx,
			provider,
			published,
		)
		if reconcileErr != nil {
			return len(providers), changed, reconcileErr
		}
		changed += count
		pricingCount, reconcileErr := store.ReconcileLegacyDeploymentCatalogPricing(ctx, provider, models)
		if reconcileErr != nil {
			return len(providers), changed, reconcileErr
		}
		changed += pricingCount
	}
	adminCatalog, err := discoveryAdminModelPublications(ctx)
	if err != nil {
		return len(providers), changed, err
	}
	legacyChanged, err := store.ReconcileLegacyDeploymentCatalogAvailability(ctx, adminCatalog)
	if err != nil {
		return len(providers), changed, err
	}
	changed += legacyChanged
	return len(providers), changed, nil
}

func discoveryAdminModelPublications(ctx context.Context) ([]storage.CatalogModelPublication, error) {
	baseURL := getDiscoveryBaseURL()
	if baseURL == "" {
		return nil, fmt.Errorf("%w: model-discovery URL is not configured", errDiscoveryUnavailable)
	}
	const pageSize = 200
	out := make([]storage.CatalogModelPublication, 0, pageSize)
	for page := 1; ; page++ {
		rawURL := fmt.Sprintf("%s/admin/models?limit=%d&page=%d", baseURL, pageSize, page)
		body, status, err := discoveryGET(ctx, rawURL)
		if err != nil {
			return nil, err
		}
		if status == http.StatusNotFound {
			// Supports rolling upgrades; provider-account availability remains enforced.
			return out, nil
		}
		if status >= http.StatusBadRequest {
			return nil, fmt.Errorf("admin model catalog returned HTTP %d", status)
		}
		var payload struct {
			Data []struct {
				CanonicalID     string `json:"canonicalId"`
				LifecycleStatus string `json:"lifecycleStatus"`
				Provider        struct {
					Slug string `json:"slug"`
				} `json:"provider"`
			} `json:"data"`
			Pagination struct {
				TotalPages int `json:"totalPages"`
			} `json:"pagination"`
		}
		if err := json.Unmarshal(body, &payload); err != nil {
			return nil, err
		}
		for _, item := range payload.Data {
			model := item.CanonicalID
			if item.Provider.Slug != "" {
				model = modelid.CanonicalOfficialModel(item.Provider.Slug, model)
			}
			out = append(out, storage.CatalogModelPublication{
				Provider:  item.Provider.Slug,
				Model:     model,
				Published: item.LifecycleStatus == "published",
			})
		}
		if page >= payload.Pagination.TotalPages || len(payload.Data) < pageSize {
			break
		}
	}
	return out, nil
}

func discoveryAdminModelPublication(ctx context.Context, provider, model string) (known bool, published bool, err error) {
	baseURL := getDiscoveryBaseURL()
	if baseURL == "" {
		return false, false, nil
	}
	canonicalID := provider + "/" + modelid.CanonicalOfficialModel(provider, model)
	body, status, err := discoveryGET(ctx, baseURL+"/admin/models/detail/"+url.PathEscape(canonicalID))
	if err != nil {
		return false, false, err
	}
	if status == http.StatusNotFound {
		return false, false, nil
	}
	if status >= http.StatusBadRequest {
		return false, false, fmt.Errorf("admin model lookup returned HTTP %d", status)
	}
	var payload struct {
		LifecycleStatus string `json:"lifecycleStatus"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return false, false, err
	}
	return true, payload.LifecycleStatus == "published", nil
}

func RunCatalogAvailabilityReconciler(ctx context.Context, store CatalogAvailabilityStore, interval time.Duration, logger *zap.Logger) {
	if interval <= 0 {
		interval = time.Minute
	}
	run := func() {
		providers, changed, err := ReconcileCatalogAvailability(ctx, store)
		if err != nil {
			logger.Warn("model catalog availability reconcile failed", zap.Error(err))
			return
		}
		if changed > 0 {
			logger.Info("model catalog availability reconciled", zap.Int("providers", providers), zap.Int64("changed", changed))
		}
	}
	run()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			run()
		}
	}
}
