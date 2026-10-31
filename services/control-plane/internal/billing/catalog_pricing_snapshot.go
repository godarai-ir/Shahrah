package billing

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/shopspring/decimal"
)

const catalogPricingMaxAge = 10 * time.Minute

// CatalogPricingSnapshotManager loads an atomic, provider-scoped runtime and
// pricing snapshot from Model Discovery. It is provider-neutral; specialized
// billing logic can select dimensions from the same normalized rate set.
type CatalogPricingSnapshotManager struct {
	provider string
	endpoint string
	client   *http.Client
	now      func() time.Time
	reloadMu sync.Mutex
	state    atomic.Pointer[anthropicPricingSnapshot]
}

// NewCatalogPricingSnapshotManager creates a provider-scoped discovery snapshot manager.
func NewCatalogPricingSnapshotManager(
	provider string,
	baseURL string,
	client *http.Client,
) (*CatalogPricingSnapshotManager, error) {
	provider = strings.ToLower(strings.TrimSpace(provider))
	if provider == "" {
		return nil, fmt.Errorf("%w: snapshot provider is missing", ErrPricingUnavailable)
	}

	base := strings.TrimRight(strings.TrimSpace(baseURL), "/")
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

func (m *CatalogPricingSnapshotManager) Reload(ctx context.Context) error {
	if m == nil {
		return ErrPricingUnavailable
	}

	m.reloadMu.Lock()
	defer m.reloadMu.Unlock()

	return m.reloadLocked(ctx)
}

func (m *CatalogPricingSnapshotManager) reloadLocked(ctx context.Context) error {
	if m.client == nil || strings.TrimSpace(m.endpoint) == "" {
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
	if decodeErr := json.NewDecoder(io.LimitReader(resp.Body, 8*1024*1024)).Decode(&payload); decodeErr != nil {
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
	ctx context.Context,
	model string,
) (anthropicPricingModel, *anthropicPricingSnapshot, error) {
	if m == nil {
		return anthropicPricingModel{}, nil, ErrPricingUnavailable
	}

	if err := m.ensureFresh(ctx); err != nil {
		return anthropicPricingModel{}, nil, err
	}

	snapshot := m.state.Load()
	if snapshot == nil || !catalogPricingSnapshotCurrent(snapshot, m.now()) {
		return anthropicPricingModel{}, nil, ErrPricingUnavailable
	}

	item, ok := snapshot.Models[normalizeCatalogPricingModelID(model, m.provider)]
	if !ok {
		return anthropicPricingModel{}, nil, fmt.Errorf(
			"%w: %s model %s is missing from the pricing snapshot",
			ErrPricingModelUnavailable,
			m.provider,
			model,
		)
	}

	return item, snapshot, nil
}

func (m *CatalogPricingSnapshotManager) Ready() bool {
	if m == nil {
		return false
	}

	snapshot := m.state.Load()

	return snapshot != nil && catalogPricingSnapshotCurrent(snapshot, m.now())
}

func (m *CatalogPricingSnapshotManager) ensureFresh(ctx context.Context) error {
	snapshot := m.state.Load()
	if snapshot != nil && catalogPricingSnapshotCurrent(snapshot, m.now()) {
		return nil
	}

	m.reloadMu.Lock()
	defer m.reloadMu.Unlock()

	snapshot = m.state.Load()
	if snapshot != nil && catalogPricingSnapshotCurrent(snapshot, m.now()) {
		return nil
	}

	return m.reloadLocked(ctx)
}

func catalogPricingSnapshotCurrent(
	snapshot *anthropicPricingSnapshot,
	now time.Time,
) bool {
	if snapshot == nil || now.Sub(snapshot.GeneratedAt) > catalogPricingMaxAge {
		return false
	}

	transition := snapshot.NextPricingTransitionAt

	return transition == nil || now.Before(*transition)
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

	if payload.NextPricingTransitionAt != nil &&
		!payload.NextPricingTransitionAt.After(payload.GeneratedAt) {
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
		signatures := make(map[string]struct{}, len(item.PricingRates))

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

			if _, duplicate := rates[input.Dimension]; duplicate && !catalogProviderAllowsConditionalRates(provider) {
				return nil, fmt.Errorf(
					"%w: duplicate %s rate dimension %s",
					ErrPricingUnavailable,
					provider,
					input.Dimension,
				)
			}

			signature := pricingComponentSignature(
				input.Dimension,
				input.Unit,
				input.Condition,
			)
			if _, duplicate := signatures[signature]; duplicate {
				return nil, fmt.Errorf(
					"%w: duplicate %s pricing signature %s",
					ErrPricingUnavailable,
					provider,
					signature,
				)
			}

			signatures[signature] = struct{}{}

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

		if provider == "moonshot" && !moonshotPricingDimensionsComplete(rates) {
			return nil, fmt.Errorf("%w: incomplete Moonshot pricing for %s", ErrPricingUnavailable, item.CanonicalID)
		}

		if provider == "deepseek" && !deepSeekPricingDimensionsComplete(rates) {
			return nil, fmt.Errorf("%w: incomplete DeepSeek pricing for %s", ErrPricingUnavailable, item.CanonicalID)
		}

		if provider == "zai" && !zaiPricingDimensionsComplete(runtimeBinding.Profile, rates) {
			return nil, fmt.Errorf("%w: incomplete Z.ai pricing for %s", ErrPricingUnavailable, item.CanonicalID)
		}

		if len(runtimeBinding.Profile.VideoControls.Operations) > 0 && !videoPricingDimensionsComplete(rates) {
			return nil, fmt.Errorf("%w: incomplete video pricing for %s", ErrPricingUnavailable, item.CanonicalID)
		}

		if provider == "google-gemini" && !googleGeminiPricingDimensionsComplete(runtimeBinding.Profile, rates) {
			return nil, fmt.Errorf("%w: incomplete Google Gemini pricing for %s", ErrPricingUnavailable, item.CanonicalID)
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
		Revision:                payload.Revision,
		GeneratedAt:             payload.GeneratedAt,
		NextPricingTransitionAt: payload.NextPricingTransitionAt,
		Models:                  models,
	}, nil
}

// videoPricingDimensionsComplete requires an unconditional per-second rate, so
// every video model can be reserved and settled before it is advertised.
func videoPricingDimensionsComplete(rates map[string]normalizedPricingRate) bool {
	rate, ok := rates[VideoOutputSecondsDimension]

	return ok && rate.UnitPriceUSD.IsPositive() && len(rate.Condition) == 0
}

func moonshotPricingDimensionsComplete(rates map[string]normalizedPricingRate) bool {
	for _, dimension := range []string{"input_tokens", "cache_read_tokens", "output_tokens"} {
		rate, ok := rates[dimension]
		if !ok || rate.UnitPriceUSD.IsNegative() || len(rate.Condition) != 0 {
			return false
		}
	}

	return true
}

func zaiPricingDimensionsComplete(
	profile platform.ModelRuntimeProfile,
	rates map[string]normalizedPricingRate,
) bool {
	required := []string{"input_tokens", "output_tokens"}
	if profile.AllowsEndpoint("chat_completions") {
		required = append(required, "cache_read_tokens")
	}

	for _, dimension := range required {
		rate, ok := rates[dimension]
		if !ok || !rate.UnitPriceUSD.IsPositive() || len(rate.Condition) != 0 ||
			strings.TrimSpace(rate.SourceURL) == "" {
			return false
		}
	}

	return true
}

func deepSeekPricingDimensionsComplete(rates map[string]normalizedPricingRate) bool {
	for _, dimension := range []string{"input_tokens", "cache_read_tokens", "output_tokens"} {
		rate, ok := rates[dimension]
		if !ok || !rate.UnitPriceUSD.IsPositive() || len(rate.Condition) != 0 || strings.TrimSpace(rate.SourceURL) == "" {
			return false
		}
	}

	return true
}

func googleGeminiPricingDimensionsComplete(
	profile platform.ModelRuntimeProfile,
	rates map[string]normalizedPricingRate,
) bool {
	input, ok := rates["input_tokens"]
	if !ok || input.UnitPriceUSD.IsNegative() || len(input.Condition) != 0 {
		return false
	}

	for _, modality := range profile.Modalities.Output {
		var dimension string

		switch modality {
		case "text":
			dimension = "output_tokens"
		case "audio":
			dimension = "audio_output_tokens"
		default:
			continue
		}

		rate, exists := rates[dimension]
		if !exists || rate.UnitPriceUSD.IsNegative() || len(rate.Condition) != 0 {
			return false
		}
	}

	return true
}

func catalogProviderAllowsConditionalRates(provider string) bool {
	switch provider {
	case "openai", "xai", "mistral-ai", "deepseek":
		return true
	default:
		return false
	}
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
