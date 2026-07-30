package billing

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
	"sync/atomic"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	prommetrics "github.com/Delpak/shahrah/services/control-plane/internal/observability/prometheus"
	"github.com/shopspring/decimal"
)

var ErrPricingUnavailable = errors.New("pricing_unavailable")

const anthropicPricingMaxAge = 24 * time.Hour

type discoveryRuntimeSnapshot struct {
	Provider    string                  `json:"provider"`
	Revision    string                  `json:"revision"`
	GeneratedAt time.Time               `json:"generatedAt"`
	Models      []discoveryRuntimeModel `json:"models"`
}

type discoveryRuntimeModel struct {
	CanonicalID    string                           `json:"canonicalId"`
	Slug           string                           `json:"slug"`
	Status         string                           `json:"status"`
	RuntimeProfile *discoveryRuntimeProfileRevision `json:"runtimeProfile"`
	PricingRates   []discoveryPricingRate           `json:"pricingRates"`
}

type discoveryRuntimeProfileRevision struct {
	Version       int                          `json:"version"`
	Profile       platform.ModelRuntimeProfile `json:"profile"`
	SourceURL     string                       `json:"sourceUrl"`
	EffectiveFrom time.Time                    `json:"effectiveFrom"`
}

type discoveryPricingRate struct {
	Dimension     string                 `json:"dimension"`
	Unit          string                 `json:"unit"`
	UnitPriceUSD  string                 `json:"unitPriceUsd"`
	Currency      string                 `json:"currency"`
	Condition     map[string]interface{} `json:"condition"`
	Region        string                 `json:"region"`
	SourceURL     string                 `json:"sourceUrl"`
	EffectiveFrom time.Time              `json:"effectiveFrom"`
	EffectiveTo   *time.Time             `json:"effectiveTo"`
}

type normalizedPricingRate struct {
	Dimension     string
	Unit          string
	UnitPriceUSD  decimal.Decimal
	Condition     map[string]interface{}
	Region        string
	SourceURL     string
	EffectiveFrom time.Time
	EffectiveTo   *time.Time
}

type anthropicPricingModel struct {
	CanonicalID    string
	Slug           string
	Status         string
	RuntimeProfile *platform.ModelRuntimeProfileBinding
	Rates          map[string]normalizedPricingRate
	RateList       []normalizedPricingRate
}

type anthropicPricingSnapshot struct {
	Revision    string
	GeneratedAt time.Time
	Models      map[string]anthropicPricingModel
}

type AnthropicPricingSnapshotManager struct {
	endpoint string
	client   *http.Client
	now      func() time.Time
	state    atomic.Pointer[anthropicPricingSnapshot]
}

func NewAnthropicPricingSnapshotManagerFromEnv(
	client *http.Client,
) (*AnthropicPricingSnapshotManager, error) {
	base := strings.TrimRight(
		strings.TrimSpace(os.Getenv("SHAHRAH_MODEL_DISCOVERY_URL")),
		"/",
	)
	if base == "" {
		base = strings.TrimRight(
			strings.TrimSpace(os.Getenv("SHAHRAH_DISCOVERY_URL")),
			"/",
		)
	}
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
	return &AnthropicPricingSnapshotManager{
		endpoint: base + "/models/runtime-snapshot/anthropic",
		client:   client,
		now:      time.Now,
	}, nil
}

func (m *AnthropicPricingSnapshotManager) Reload(ctx context.Context) error {
	if m == nil {
		return ErrPricingUnavailable
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, m.endpoint, http.NoBody)
	if err != nil {
		return ErrPricingUnavailable
	}
	resp, err := m.client.Do(req)
	if err != nil {
		prommetrics.RecordAnthropicPricingFailure("fetch")
		return fmt.Errorf("%w: snapshot fetch failed", ErrPricingUnavailable)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		prommetrics.RecordAnthropicPricingFailure("fetch")
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64*1024))
		return fmt.Errorf("%w: snapshot HTTP %d", ErrPricingUnavailable, resp.StatusCode)
	}

	var payload discoveryRuntimeSnapshot
	decoder := json.NewDecoder(io.LimitReader(resp.Body, 8*1024*1024))
	if err := decoder.Decode(&payload); err != nil {
		prommetrics.RecordAnthropicPricingFailure("decode")
		return fmt.Errorf("%w: snapshot payload is invalid", ErrPricingUnavailable)
	}
	snapshot, err := normalizeAnthropicPricingSnapshot(payload)
	if err != nil {
		prommetrics.RecordAnthropicPricingFailure("validate")
		return err
	}
	m.state.Store(snapshot)
	return nil
}

func (m *AnthropicPricingSnapshotManager) Run(
	ctx context.Context,
	interval time.Duration,
) {
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
			_ = m.Reload(ctx)
		}
	}
}

func (m *AnthropicPricingSnapshotManager) Price(
	model string,
) (anthropicPricingModel, *anthropicPricingSnapshot, error) {
	if m == nil {
		return anthropicPricingModel{}, nil, ErrPricingUnavailable
	}
	snapshot := m.state.Load()
	if snapshot == nil ||
		m.now().Sub(snapshot.GeneratedAt) > anthropicPricingMaxAge {
		prommetrics.RecordAnthropicPricingFailure("stale")
		return anthropicPricingModel{}, nil, ErrPricingUnavailable
	}
	prommetrics.SetAnthropicPricingSnapshotAge(m.now().Sub(snapshot.GeneratedAt))
	normalized := normalizeAnthropicPricingModelID(model)
	price, ok := snapshot.Models[normalized]
	if !ok {
		prommetrics.RecordAnthropicPricingFailure("missing_rate")
		return anthropicPricingModel{}, nil, ErrPricingUnavailable
	}
	return price, snapshot, nil
}

func (m *AnthropicPricingSnapshotManager) Ready() bool {
	if m == nil {
		return false
	}
	snapshot := m.state.Load()
	return snapshot != nil &&
		m.now().Sub(snapshot.GeneratedAt) <= anthropicPricingMaxAge
}

func (m *AnthropicPricingSnapshotManager) ResolveModelRuntimeProfiles(
	models []string,
	expectedRevision string,
) ([]platform.ModelRuntimeProfileBinding, error) {
	if m == nil {
		return nil, ErrPricingUnavailable
	}
	snapshot := m.state.Load()
	if snapshot == nil ||
		m.now().Sub(snapshot.GeneratedAt) > anthropicPricingMaxAge ||
		(expectedRevision != "" && snapshot.Revision != expectedRevision) {
		return nil, ErrPricingUnavailable
	}
	profiles := make([]platform.ModelRuntimeProfileBinding, 0, len(models))
	seen := make(map[string]struct{}, len(models))
	for _, model := range models {
		normalized := normalizeAnthropicPricingModelID(model)
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		item, ok := snapshot.Models[normalized]
		if !ok || item.RuntimeProfile == nil {
			return nil, ErrPricingUnavailable
		}
		binding := *item.RuntimeProfile
		if err := binding.Validate(); err != nil {
			return nil, ErrPricingUnavailable
		}
		profiles = append(profiles, binding)
	}
	if len(profiles) == 0 {
		return nil, ErrPricingUnavailable
	}
	return profiles, nil
}

func normalizeAnthropicPricingSnapshot(
	payload discoveryRuntimeSnapshot,
) (*anthropicPricingSnapshot, error) {
	if payload.Provider != "anthropic" ||
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
			rates[input.Dimension] = normalizedPricingRate{
				Dimension:    input.Dimension,
				Unit:         input.Unit,
				UnitPriceUSD: value,
				Condition:    input.Condition,
				Region:       input.Region,
				SourceURL:    input.SourceURL,
			}
		}
		if _, ok := rates["input_tokens"]; !ok {
			return nil, ErrPricingUnavailable
		}
		if _, ok := rates["output_tokens"]; !ok {
			return nil, ErrPricingUnavailable
		}
		model := anthropicPricingModel{
			CanonicalID:    item.CanonicalID,
			Slug:           item.Slug,
			Status:         item.Status,
			RuntimeProfile: runtimeBinding,
			Rates:          rates,
		}
		models[strings.ToLower(strings.TrimSpace(item.Slug))] = model
		models[strings.ToLower(strings.TrimSpace(item.CanonicalID))] = model
	}
	return &anthropicPricingSnapshot{
		Revision:    payload.Revision,
		GeneratedAt: payload.GeneratedAt,
		Models:      models,
	}, nil
}

func normalizeAnthropicPricingModelID(model string) string {
	value := strings.ToLower(strings.TrimSpace(model))
	if strings.Count(value, ":") == 2 {
		parts := strings.SplitN(value, ":", 3)
		value = parts[2]
	}
	value = strings.TrimPrefix(value, "anthropic/")
	return value
}

func isAnthropicModel(model string) bool {
	value := strings.ToLower(strings.TrimSpace(model))
	return strings.HasPrefix(value, "anthropic/") ||
		strings.HasPrefix(value, "anthropic:") ||
		strings.HasPrefix(value, "claude-")
}
