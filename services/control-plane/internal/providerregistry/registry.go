// Package providerregistry defines the first-class provider capabilities used
// across control-plane routing, pricing, and runtime-profile validation.
package providerregistry

import "strings"

// PricingSnapshotKind identifies the authoritative pricing snapshot loader.
type PricingSnapshotKind string

const (
	// PricingSnapshotAnthropic uses the Anthropic-specific snapshot manager.
	PricingSnapshotAnthropic PricingSnapshotKind = "anthropic"
	// PricingSnapshotCatalog uses the provider-neutral catalog snapshot manager.
	PricingSnapshotCatalog PricingSnapshotKind = "catalog"
)

// Definition describes one first-class provider and its control-plane behavior.
type Definition struct {
	Slug                           string
	DisplayName                    string
	Aliases                        []string
	PricingSnapshot                PricingSnapshotKind
	DispatchRuntimeProfileRequired bool
	LegacyModelPrefixes            []string
}

var definitions = []Definition{
	{
		Slug:            "openai",
		DisplayName:     "OpenAI",
		PricingSnapshot: PricingSnapshotCatalog,
	},
	{
		Slug:                           "anthropic",
		DisplayName:                    "Anthropic",
		PricingSnapshot:                PricingSnapshotAnthropic,
		DispatchRuntimeProfileRequired: true,
		LegacyModelPrefixes:            []string{"claude-"},
	},
	{
		Slug:                           "xai",
		DisplayName:                    "xAI",
		PricingSnapshot:                PricingSnapshotCatalog,
		DispatchRuntimeProfileRequired: true,
	},
	{
		Slug:                           "deepseek",
		DisplayName:                    "DeepSeek",
		PricingSnapshot:                PricingSnapshotCatalog,
		DispatchRuntimeProfileRequired: true,
	},
	{
		Slug:                           "moonshot",
		DisplayName:                    "Moonshot",
		PricingSnapshot:                PricingSnapshotCatalog,
		DispatchRuntimeProfileRequired: true,
	},
	{
		Slug:                           "google-gemini",
		DisplayName:                    "Google Gemini",
		Aliases:                        []string{"gemini"},
		PricingSnapshot:                PricingSnapshotCatalog,
		DispatchRuntimeProfileRequired: true,
	},
	{
		Slug:                           "mistral-ai",
		DisplayName:                    "Mistral AI",
		Aliases:                        []string{"mistral"},
		PricingSnapshot:                PricingSnapshotCatalog,
		DispatchRuntimeProfileRequired: true,
		LegacyModelPrefixes:            []string{"mistral-", "codestral-", "voxtral-"},
	},
}

// Definitions returns all first-class provider definitions.
func Definitions() []Definition {
	out := make([]Definition, len(definitions))
	copy(out, definitions)

	return out
}

// PricingProviders returns providers backed by an authoritative pricing snapshot.
func PricingProviders() []Definition {
	out := make([]Definition, 0, len(definitions))
	for _, definition := range definitions {
		if definition.PricingSnapshot != "" {
			out = append(out, definition)
		}
	}

	return out
}

// Lookup resolves a canonical provider slug or a supported alias.
func Lookup(value string) (Definition, bool) {
	value = strings.ToLower(strings.TrimSpace(value))
	for _, definition := range definitions {
		if value == definition.Slug || containsFold(definition.Aliases, value) {
			return definition, true
		}
	}

	return Definition{}, false
}

// Normalize returns the canonical slug for a first-class provider.
func Normalize(value string) string {
	if definition, ok := Lookup(value); ok {
		return definition.Slug
	}

	return strings.ToLower(strings.TrimSpace(value))
}

// Matches reports whether two provider values resolve to the same definition.
func Matches(left, right string) bool {
	leftDefinition, leftOK := Lookup(left)
	rightDefinition, rightOK := Lookup(right)

	return leftOK && rightOK && leftDefinition.Slug == rightDefinition.Slug
}

// FromModelIdentifier resolves the provider encoded in an FQN, a provider/model
// identifier, or an explicitly supported legacy model prefix.
func FromModelIdentifier(model string) (Definition, bool) {
	value := strings.ToLower(strings.TrimSpace(model))
	if value == "" {
		return Definition{}, false
	}

	if separator := strings.IndexByte(value, ':'); separator > 0 {
		if definition, ok := Lookup(value[:separator]); ok {
			return definition, true
		}
	}

	if separator := strings.IndexByte(value, '/'); separator > 0 {
		if definition, ok := Lookup(value[:separator]); ok {
			return definition, true
		}
	}

	for _, definition := range definitions {
		for _, prefix := range definition.LegacyModelPrefixes {
			if strings.HasPrefix(value, prefix) {
				return definition, true
			}
		}
	}

	return Definition{}, false
}

// SplitQualifiedModel splits a first-class provider/model identifier and
// canonicalizes the provider slug.
func SplitQualifiedModel(model string) (Definition, string, bool) {
	model = strings.TrimSpace(model)
	separator := strings.IndexByte(model, '/')
	if separator <= 0 || separator == len(model)-1 {
		return Definition{}, "", false
	}

	definition, ok := Lookup(model[:separator])
	if !ok {
		return Definition{}, "", false
	}

	officialModel := strings.TrimSpace(model[separator+1:])
	if officialModel == "" {
		return Definition{}, "", false
	}

	return definition, officialModel, true
}

func containsFold(values []string, target string) bool {
	for _, value := range values {
		if strings.EqualFold(value, target) {
			return true
		}
	}

	return false
}
