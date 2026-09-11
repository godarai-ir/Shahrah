package server

import (
	"encoding/json"
	"fmt"
	"strings"
)

// tokenBazaarModelsResponse is the minimal shape this delivery reads out of
// TokenBazaar's documented, authenticated GET /v1/models response (see
// plans/tokenbazaar-production-readiness.md, "Confirmed wire contract" and
// conflict-register item #12). Only the fields needed to populate the
// per-model image-parameter confirmation cache are modeled; every other
// documented field (pricing_kind, pricing_defaults, price_rows,
// compare_price_rows, etc.) is intentionally left unparsed by this function.
type tokenBazaarModelsResponse struct {
	Data []struct {
		ID                    string   `json:"id"`
		SupportedEndpointType []string `json:"supported_endpoint_types"`
		SupportedParameters   struct {
			Size    []string `json:"size"`
			Quality []string `json:"quality"`
		} `json:"supported_parameters"`
	} `json:"data"`
}

// parseTokenBazaarImageParameterSnapshots decodes a GET /v1/models response
// body and returns one tokenBazaarImageParameterSnapshot per in-scope image
// model that the response actually describes. It does not itself perform
// the HTTP call or install anything into the cache — callers (a future
// scheduled discovery job, or a credential-test flow, per conflict #12's
// still-open wiring gap) are expected to call setTokenBazaarImageParameterSnapshot
// for each returned entry. This separation keeps the parsing logic unit
// testable without a live account.
func parseTokenBazaarImageParameterSnapshots(sourceURL string, body []byte) (map[string]tokenBazaarImageParameterSnapshot, error) {
	var response tokenBazaarModelsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("tokenbazaar_models_response_invalid: %w", err)
	}

	snapshots := make(map[string]tokenBazaarImageParameterSnapshot)

	for _, entry := range response.Data {
		model := strings.ToLower(strings.TrimSpace(entry.ID))
		if _, inScope := tokenBazaarImageModels[model]; !inScope {
			continue
		}

		if !containsTokenBazaarEndpointType(entry.SupportedEndpointType, "images") {
			continue
		}

		snapshots[model] = tokenBazaarImageParameterSnapshot{
			SupportedSizes:   toTokenBazaarStringSet(entry.SupportedParameters.Size),
			SupportedQuality: toTokenBazaarStringSet(entry.SupportedParameters.Quality),
			SourceURL:        sourceURL,
		}
	}

	return snapshots, nil
}

// refreshTokenBazaarImageParameterCache parses a GET /v1/models response and
// installs every in-scope image model's snapshot into the shared cache used
// by validateTokenBazaarImageParameters. This is the single integration
// point a future live discovery job (or credential-test flow) should call;
// it is not currently invoked by any automatic trigger in this delivery
// (there is no platform-managed TokenBazaar credential to call GET
// /v1/models with yet — see
// plans/tokenbazaar-production-readiness.md conflict #12).
func refreshTokenBazaarImageParameterCache(sourceURL string, body []byte) error {
	snapshots, err := parseTokenBazaarImageParameterSnapshots(sourceURL, body)
	if err != nil {
		return err
	}

	for model, snapshot := range snapshots {
		setTokenBazaarImageParameterSnapshot(model, snapshot)
	}

	return nil
}

func containsTokenBazaarEndpointType(values []string, target string) bool {
	for _, value := range values {
		if strings.EqualFold(strings.TrimSpace(value), target) {
			return true
		}
	}

	return false
}

func toTokenBazaarStringSet(values []string) map[string]struct{} {
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}

		set[trimmed] = struct{}{}
	}

	return set
}
