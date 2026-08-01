package billing

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/shopspring/decimal"
)

// CatalogPricingSnapshotManager loads an atomic, provider-scoped runtime and
// pricing snapshot from Model Discovery. It is provider-neutral; specialized
// billing logic can select dimensions from the same normalized rate set.
type CatalogPricingSnapshotManager struct {
	provider string
	endpoint string
	client   *http.Client
	now      func() time.Time
	state    atomic.Pointer[anthropicPricingSnapshot]
}

func NewCatalogPricingSnapshotManagerFromEnv(
	provider string,
	client *http.Client,
) (*CatalogPricingSnapshotManager, error) {
	provider = strings.ToLower(strings.TrimSpace(provider))
	if provider == "" {
		return nil, fmt.Errorf("%w: snapshot provider is missing", ErrPricingUnavailable)
	}

	base := strings.TrimRight(
		strings.TrimSpace(firstNonEmptyEnvironment(
			"SHAHRAH_MODEL_DISCOVERY_URL",
			"SHAHRAH_DISCOVERY_URL",
		)),
		"/",
	)
	if base == "" {
		return nil, fmt.Errorf("%w: model discovery URL is missing", ErrPricingUnavailable)
	}

	parsed, err := url.Parse(base)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("%w: model discovery URL is invalid", ErrPricingUnavailable)
	}

	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}

	return &CatalogPricingSnapshotManager{
		provider: provider,
		endpoint: base + "/models/runtime-snapshot/" + url.PathEscape(provider),
		client:   client,
		now:      time.Now,
	}, nil
}

//nolint:govet // Short-lived branch errors are checked immediately and intentionally scoped with short declarations.
func (m *CatalogPricingSnapshotManager) Reload(ctx context.Context) error {
	if m == nil {
		return ErrPricingUnavailable
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, m.endpoint, http.NoBody)
	if err != nil {
		return ErrPricingUnavailable
	}

	resp, err := m.client.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %s snapshot fetch failed", ErrPricingUnavailable, m.provider)
	}
	defer resp.Body.Close() //nolint:errcheck // Cleanup is best-effort and cannot change the completed operation.

	if resp.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64*1024)) //nolint:errcheck // A downstream disconnect is handled by terminating the stream.
		return fmt.Errorf("%w: %s snapshot HTTP %d", ErrPricingUnavailable, m.provider, resp.StatusCode)
	}

	var payload discoveryRuntimeSnapshot
	if err := json.NewDecoder(io.LimitReader(resp.Body, 8*1024*1024)).Decode(&payload); err != nil {
		return fmt.Errorf("%w: %s snapshot payload is invalid", ErrPricingUnavailable, m.provider)
	}

	snapshot, err := normalizeCatalogPricingSnapshot(payload, m.provider)
	if err != nil {
		return err
	}

	m.state.Store(snapshot)

	return nil
}

func (m *CatalogPricingSnapshotManager) Run(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = 5 * time.Minute
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = m.Reload(ctx) //nolint:errcheck // Failure is intentionally handled by the surrounding best-effort path.
		}
	}
}

func (m *CatalogPricingSnapshotManager) price(
	model string,
) (anthropicPricingModel, *anthropicPricingSnapshot, error) {
	if m == nil {
		return anthropicPricingModel{}, nil, ErrPricingUnavailable
	}

	snapshot := m.state.Load()
	if snapshot == nil || m.now().Sub(snapshot.GeneratedAt) > anthropicPricingMaxAge {
		return anthropicPricingModel{}, nil, ErrPricingUnavailable
	}

	item, ok := snapshot.Models[normalizeCatalogPricingModelID(model, m.provider)]
	if !ok {
		return anthropicPricingModel{}, nil, ErrPricingUnavailable
	}

	return item, snapshot, nil
}

func (m *CatalogPricingSnapshotManager) Ready() bool {
	if m == nil {
		return false
	}

	snapshot := m.state.Load()

	return snapshot != nil && m.now().Sub(snapshot.GeneratedAt) <= anthropicPricingMaxAge
}

func normalizeCatalogPricingSnapshot(
	payload discoveryRuntimeSnapshot,
	provider string,
) (*anthropicPricingSnapshot, error) {
	if payload.Provider != provider ||
		strings.TrimSpace(payload.Revision) == "" ||
		payload.GeneratedAt.IsZero() ||
		len(payload.Models) == 0 {
		return nil, ErrPricingUnavailable
	}

	models := make(map[string]anthropicPricingModel, len(payload.Models)*2)
	for _, item := range payload.Models {
		if item.RuntimeProfile == nil {
			return nil, ErrPricingUnavailable
		}

		runtimeBinding := &platform.ModelRuntimeProfileBinding{
			Provider:        payload.Provider,
			CatalogRevision: payload.Revision,
			CanonicalID:     item.CanonicalID,
			Slug:            item.Slug,
			ProfileVersion:  item.RuntimeProfile.Version,
			Profile:         item.RuntimeProfile.Profile,
			SourceURL:       item.RuntimeProfile.SourceURL,
			EffectiveFrom:   item.RuntimeProfile.EffectiveFrom,
		}
		if err := runtimeBinding.Validate(); err != nil {
			return nil, ErrPricingUnavailable
		}

		rates := make(map[string]normalizedPricingRate, len(item.PricingRates))

		rateList := make([]normalizedPricingRate, 0, len(item.PricingRates))
		for _, input := range item.PricingRates {
			if input.Currency != "USD" || input.Region != "global" {
				continue
			}

			value, err := decimal.NewFromString(input.UnitPriceUSD)
			if err != nil || value.IsNegative() ||
				input.Dimension == "" || input.Unit == "" ||
				input.SourceURL == "" {
				return nil, ErrPricingUnavailable
			}

			if _, duplicate := rates[input.Dimension]; duplicate && provider != "openai" && provider != "xai" && provider != "mistral-ai" {
				return nil, fmt.Errorf(
					"%w: duplicate %s rate dimension %s",
					ErrPricingUnavailable,
					provider,
					input.Dimension,
				)
			}

			normalized := normalizedPricingRate{
				Dimension:     input.Dimension,
				Unit:          input.Unit,
				UnitPriceUSD:  value,
				Condition:     input.Condition,
				Region:        input.Region,
				SourceURL:     input.SourceURL,
				EffectiveFrom: input.EffectiveFrom,
				EffectiveTo:   input.EffectiveTo,
			}

			rateList = append(rateList, normalized)
			if current, exists := rates[input.Dimension]; !exists ||
				(len(current.Condition) > 0 && len(normalized.Condition) == 0) {
				rates[input.Dimension] = normalized
			}
		}

		if provider == "mistral-ai" && !mistralPricingDimensionsComplete(runtimeBinding.Profile, rates) {
			return nil, fmt.Errorf("%w: incomplete Mistral AI pricing for %s", ErrPricingUnavailable, item.CanonicalID)
		}

		model := anthropicPricingModel{
			CanonicalID:    item.CanonicalID,
			Slug:           item.Slug,
			Status:         item.Status,
			RuntimeProfile: runtimeBinding,
			Rates:          rates,
			RateList:       rateList,
		}
		models[normalizeCatalogPricingModelID(item.Slug, provider)] = model

		models[normalizeCatalogPricingModelID(item.CanonicalID, provider)] = model
		for _, alias := range item.RuntimeProfile.Profile.Aliases {
			models[normalizeCatalogPricingModelID(alias, provider)] = model
		}
	}

	return &anthropicPricingSnapshot{
		Revision:    payload.Revision,
		GeneratedAt: payload.GeneratedAt,
		Models:      models,
	}, nil
}

func mistralPricingDimensionsComplete(profile platform.ModelRuntimeProfile, rates map[string]normalizedPricingRate) bool {
	require := func(dimensions ...string) bool {
		for _, dimension := range dimensions {
			rate, ok := rates[dimension]
			if !ok || rate.UnitPriceUSD.IsNegative() || len(rate.Condition) != 0 {
				return false
			}
		}

		return true
	}
	requireAny := func(dimensions ...string) bool {
		for _, dimension := range dimensions {
			rate, ok := rates[dimension]
			if !ok || rate.UnitPriceUSD.IsNegative() {
				return false
			}
		}

		return true
	}

	for _, endpoint := range profile.Endpoints {
		switch endpoint {
		case "chat_completions", "fim_completions", "conversations", "agents":
			if !require("input_tokens", "output_tokens") {
				return false
			}
		case "embeddings", "moderations", "chat_moderations":
			if !require("input_tokens") {
				return false
			}
		case "audio_transcriptions":
			if !require("audio_input_seconds") {
				return false
			}
		case "audio_speech":
			if !require("speech_input_characters") {
				return false
			}
		case "ocr":
			if !require("ocr_input_pages") || !requireAny("ocr_annotated_pages") {
				return false
			}
		}
	}

	return true
}

func normalizeCatalogPricingModelID(model, provider string) string {
	value := strings.ToLower(strings.TrimSpace(model))
	if strings.Count(value, ":") == 2 {
		parts := strings.SplitN(value, ":", 3)
		value = parts[2]
	}

	value = strings.TrimPrefix(value, provider+"/")
	if provider == "google-gemini" {
		value = strings.TrimPrefix(value, "gemini/")
	}

	if provider == "mistral-ai" {
		value = strings.TrimPrefix(value, "mistral/")
	}

	return value
}

func firstNonEmptyEnvironment(keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}

	return ""
}
