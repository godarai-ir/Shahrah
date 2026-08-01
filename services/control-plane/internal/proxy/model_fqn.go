package proxy

import (
	"strings"

	"github.com/Delpak/shahrah/services/control-plane/internal/modelid"
)

// splitModelFQN parses a model in provider:account_alias:model_id format.
func splitModelFQN(input string) (provider string, accountAlias string, officialModel string, ok bool) {
	return modelid.SplitModelFQN(input)
}

func normalizeModelIdentifier(model string) string {
	model = strings.TrimSpace(model)
	if model == "" {
		return ""
	}

	provider, accountAlias, officialModel, ok := splitModelFQN(model)
	if !ok {
		return model
	}

	return modelid.BuildModelFQN(provider, accountAlias, officialModel)
}

// modelRoutingCandidates returns route candidates for an inbound model id.
// For strict tenant/account isolation, FQN values are routed only as-is.
func modelRoutingCandidates(model string) []string {
	model = normalizeModelIdentifier(model)
	if model == "" {
		return nil
	}

	return []string{model}
}

func providerModelID(model string) string {
	model = normalizeModelIdentifier(model)

	_, _, officialModel, ok := splitModelFQN(model)
	if ok {
		return officialModel
	}

	return model
}

// billingEstimateModel preserves provider-qualified identifiers for catalog
// pricing providers. Their pricing snapshots are selected by provider prefix,
// while the provider-facing request still uses providerModelID.
func billingEstimateModel(model string) string {
	model = normalizeModelIdentifier(model)

	provider, _, _, ok := splitModelFQN(model)
	if ok && (isGoogleGeminiProvider(provider) || isMistralProvider(provider) || strings.EqualFold(provider, "openai") || strings.EqualFold(provider, "xai")) {
		return model
	}

	value := strings.ToLower(model)
	if strings.HasPrefix(value, "google-gemini/") ||
		strings.HasPrefix(value, "gemini/") ||
		strings.HasPrefix(value, "mistral-ai/") ||
		strings.HasPrefix(value, "mistral/") ||
		strings.HasPrefix(value, "openai/") ||
		strings.HasPrefix(value, "xai/") {
		return model
	}

	return providerModelID(model)
}

func isMistralProvider(provider string) bool {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "mistral-ai", "mistral":
		return true
	default:
		return false
	}
}

func isMistralModelIdentifier(model string) bool {
	value := strings.ToLower(strings.TrimSpace(model))
	if value == "" {
		return false
	}

	if separator := strings.IndexByte(value, ':'); separator > 0 {
		return isMistralProvider(value[:separator])
	}

	if separator := strings.IndexByte(value, '/'); separator > 0 && isMistralProvider(value[:separator]) {
		return true
	}

	return strings.HasPrefix(value, "mistral-") ||
		strings.HasPrefix(value, "codestral-") ||
		strings.HasPrefix(value, "voxtral-")
}

func isGoogleGeminiProvider(provider string) bool {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "google-gemini", "gemini":
		return true
	default:
		return false
	}
}
