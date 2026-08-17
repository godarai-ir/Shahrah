// Package providerregistry defines the first-class provider capabilities used
// across control-plane routing, pricing, and runtime-profile validation.
package providerregistry

import (
	"fmt"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

// PricingSnapshotKind identifies the authoritative pricing snapshot loader.
type PricingSnapshotKind string

const (
	// PricingSnapshotAnthropic uses the Anthropic-specific snapshot manager.
	PricingSnapshotAnthropic PricingSnapshotKind = "anthropic"
	// PricingSnapshotCatalog uses the provider-neutral catalog snapshot manager.
	PricingSnapshotCatalog PricingSnapshotKind = "catalog"
)

// PricingStrategyKind identifies the provider-specific usage calculator.
type PricingStrategyKind string

const (
	PricingStrategyStandard     PricingStrategyKind = "standard"
	PricingStrategyOpenAI       PricingStrategyKind = "openai"
	PricingStrategyAnthropic    PricingStrategyKind = "anthropic"
	PricingStrategyXAI          PricingStrategyKind = "xai"
	PricingStrategyDeepSeek     PricingStrategyKind = "deepseek"
	PricingStrategyZAI          PricingStrategyKind = "zai"
	PricingStrategyMoonshot     PricingStrategyKind = "moonshot"
	PricingStrategyGoogleGemini PricingStrategyKind = "google-gemini"
	PricingStrategyMistral      PricingStrategyKind = "mistral"
)

// Capability identifies a provider surface exposed by the control plane.
type Capability string

const (
	CapabilityChatCompletions Capability = "chat_completions"
	CapabilityCompletions     Capability = "completions"
	CapabilityResponses       Capability = "responses"
	CapabilityEmbeddings      Capability = "embeddings"
	CapabilityModerations     Capability = "moderations"
	CapabilityImages          Capability = "images"
	CapabilityAudio           Capability = "audio"
	CapabilityBatches         Capability = "batches"
	CapabilityFiles           Capability = "files"
	CapabilityModels          Capability = "models"
	CapabilityRealtime        Capability = "realtime"
	CapabilityTokenCounting   Capability = "token_counting"
	CapabilityUploads         Capability = "uploads"
	CapabilityVectorStores    Capability = "vector_stores"
	CapabilityContainers      Capability = "containers"
	CapabilityVideos          Capability = "videos"
	CapabilityInteractions    Capability = "interactions"
	CapabilityContextCaching  Capability = "context_caching"
	CapabilityOCR             Capability = "ocr"
	CapabilityAgents          Capability = "agents"
	CapabilityConversations   Capability = "conversations"
)

// Definition describes one first-class provider and its control-plane behavior.
type Definition struct {
	Slug                           string
	DisplayName                    string
	Aliases                        []string
	PricingSnapshot                PricingSnapshotKind
	PricingStrategy                PricingStrategyKind
	Capabilities                   []Capability
	OperationNamespaces            []string
	Operations                     []platform.ProviderOperation
	DispatchRuntimeProfileRequired bool
	LegacyModelPrefixes            []string
}

var definitions = []Definition{
	{
		Slug:            "openai",
		DisplayName:     "OpenAI",
		PricingSnapshot: PricingSnapshotCatalog,
		PricingStrategy: PricingStrategyOpenAI,
		Capabilities: []Capability{
			CapabilityChatCompletions,
			CapabilityCompletions,
			CapabilityResponses,
			CapabilityEmbeddings,
			CapabilityModerations,
			CapabilityImages,
			CapabilityAudio,
			CapabilityBatches,
			CapabilityFiles,
			CapabilityModels,
			CapabilityRealtime,
			CapabilityUploads,
			CapabilityVectorStores,
			CapabilityContainers,
			CapabilityVideos,
			CapabilityConversations,
		},
		OperationNamespaces: []string{"openai_"},
		Operations:          []platform.ProviderOperation{platform.ProviderOperationCredentialProbe, platform.ProviderOperationChatCompletions},
	},
	{
		Slug:            "anthropic",
		DisplayName:     "Anthropic",
		PricingSnapshot: PricingSnapshotAnthropic,
		PricingStrategy: PricingStrategyAnthropic,
		Capabilities: []Capability{
			CapabilityChatCompletions,
			CapabilityTokenCounting,
			CapabilityBatches,
			CapabilityFiles,
			CapabilityModels,
		},
		OperationNamespaces: []string{"anthropic_"},
		Operations: []platform.ProviderOperation{
			platform.ProviderOperationCredentialProbe,
			platform.ProviderOperationChatCompletions,
		},
		DispatchRuntimeProfileRequired: true,
		LegacyModelPrefixes:            []string{"claude-"},
	},
	{
		Slug:            "xai",
		DisplayName:     "xAI",
		PricingSnapshot: PricingSnapshotCatalog,
		PricingStrategy: PricingStrategyXAI,
		Capabilities: []Capability{
			CapabilityChatCompletions,
			CapabilityResponses,
			CapabilityImages,
			CapabilityBatches,
			CapabilityFiles,
			CapabilityModels,
		},
		OperationNamespaces: []string{"xai_"},
		Operations: []platform.ProviderOperation{
			platform.ProviderOperationCredentialProbe,
			platform.ProviderOperationChatCompletions,
			platform.ProviderOperationOpenAIResponsesCreate,
			platform.ProviderOperationOpenAIResponsesGet,
			platform.ProviderOperationOpenAIResponsesDelete,
			platform.ProviderOperationOpenAIResponsesCancel,
			platform.ProviderOperationOpenAIResponsesInputItems,
			platform.ProviderOperationOpenAIResponsesInputTokens,
			platform.ProviderOperationOpenAIImagesGenerations,
			platform.ProviderOperationOpenAIFilesCreate,
			platform.ProviderOperationOpenAIFilesList,
			platform.ProviderOperationOpenAIFilesGet,
			platform.ProviderOperationOpenAIFilesDelete,
			platform.ProviderOperationOpenAIFilesContent,
			platform.ProviderOperationOpenAIModelsList,
			platform.ProviderOperationOpenAIModelsGet,
		},
		DispatchRuntimeProfileRequired: true,
	},
	{
		Slug:                           "deepseek",
		DisplayName:                    "DeepSeek",
		PricingSnapshot:                PricingSnapshotCatalog,
		PricingStrategy:                PricingStrategyDeepSeek,
		Capabilities:                   []Capability{CapabilityChatCompletions},
		Operations:                     []platform.ProviderOperation{platform.ProviderOperationCredentialProbe, platform.ProviderOperationChatCompletions},
		DispatchRuntimeProfileRequired: true,
	},
	{
		Slug:                "zai",
		DisplayName:         "Z.ai",
		PricingSnapshot:     PricingSnapshotCatalog,
		PricingStrategy:     PricingStrategyZAI,
		Capabilities:        []Capability{CapabilityChatCompletions, CapabilityOCR},
		OperationNamespaces: []string{"zai_"},
		Operations: []platform.ProviderOperation{
			platform.ProviderOperationCredentialProbe,
			platform.ProviderOperationChatCompletions,
			platform.ProviderOperationZAILayoutParsing,
		},
		DispatchRuntimeProfileRequired: true,
	},
	{
		Slug:                           "moonshot",
		DisplayName:                    "Moonshot",
		PricingSnapshot:                PricingSnapshotCatalog,
		PricingStrategy:                PricingStrategyMoonshot,
		Capabilities:                   []Capability{CapabilityChatCompletions},
		Operations:                     []platform.ProviderOperation{platform.ProviderOperationCredentialProbe, platform.ProviderOperationChatCompletions},
		DispatchRuntimeProfileRequired: true,
	},
	{
		Slug:            "google-gemini",
		DisplayName:     "Google Gemini",
		Aliases:         []string{"gemini"},
		PricingSnapshot: PricingSnapshotCatalog,
		PricingStrategy: PricingStrategyGoogleGemini,
		Capabilities: []Capability{
			CapabilityChatCompletions,
			CapabilityEmbeddings,
			CapabilityBatches,
			CapabilityFiles,
			CapabilityModels,
			CapabilityTokenCounting,
			CapabilityInteractions,
			CapabilityContextCaching,
		},
		OperationNamespaces:            []string{"gemini_"},
		Operations:                     []platform.ProviderOperation{platform.ProviderOperationCredentialProbe, platform.ProviderOperationChatCompletions},
		DispatchRuntimeProfileRequired: true,
	},
	{
		Slug:            "mistral-ai",
		DisplayName:     "Mistral AI",
		Aliases:         []string{"mistral"},
		PricingSnapshot: PricingSnapshotCatalog,
		PricingStrategy: PricingStrategyMistral,
		Capabilities: []Capability{
			CapabilityChatCompletions,
			CapabilityCompletions,
			CapabilityModerations,
			CapabilityAudio,
			CapabilityBatches,
			CapabilityFiles,
			CapabilityOCR,
			CapabilityAgents,
			CapabilityConversations,
		},
		OperationNamespaces: []string{"mistral_"},
		Operations: []platform.ProviderOperation{
			platform.ProviderOperationCredentialProbe,
			platform.ProviderOperationChatCompletions,
			platform.ProviderOperationOpenAICompletions,
			platform.ProviderOperationOpenAIAudioSpeech,
			platform.ProviderOperationOpenAIAudioTranscriptions,
			platform.ProviderOperationOpenAIAudioTranslations,
		},
		DispatchRuntimeProfileRequired: true,
		LegacyModelPrefixes:            []string{"mistral-", "codestral-", "voxtral-"},
	},
}

// Definitions returns all first-class provider definitions.
func Definitions() []Definition {
	out := make([]Definition, len(definitions))
	for index, definition := range definitions {
		out[index] = cloneDefinition(definition)
	}

	return out
}

// PricingProviders returns providers backed by an authoritative pricing snapshot.
func PricingProviders() []Definition {
	out := make([]Definition, 0, len(definitions))
	for _, definition := range definitions {
		if definition.PricingSnapshot != "" {
			out = append(out, cloneDefinition(definition))
		}
	}

	return out
}

// Lookup resolves a canonical provider slug or a supported alias.
func Lookup(value string) (Definition, bool) {
	value = strings.ToLower(strings.TrimSpace(value))
	for _, definition := range definitions {
		if value == definition.Slug || containsFold(definition.Aliases, value) {
			return cloneDefinition(definition), true
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

// Supports reports whether the provider declares the requested capability.
func (d Definition) Supports(capability Capability) bool {
	for _, candidate := range d.Capabilities {
		if candidate == capability {
			return true
		}
	}

	return false
}

// SupportsOperation reports whether the provider owns an exact operation or namespace.
func (d Definition) SupportsOperation(operation platform.ProviderOperation) bool {
	value := strings.ToLower(strings.TrimSpace(string(operation)))
	if value == "" {
		return false
	}

	for _, candidate := range d.Operations {
		if strings.EqualFold(value, string(candidate)) {
			return true
		}
	}

	for _, namespace := range d.OperationNamespaces {
		if strings.HasPrefix(value, namespace) {
			return true
		}
	}

	return false
}

// Validate verifies that provider definitions are complete and unambiguous.
func Validate(items []Definition) error {
	providers := make(map[string]struct{}, len(items))
	identifiers := make(map[string]string, len(items)*2)

	for _, definition := range items {
		if err := validateDefinition(definition); err != nil {
			return err
		}

		if _, duplicate := providers[definition.Slug]; duplicate {
			return fmt.Errorf("provider registry: duplicate slug %q", definition.Slug)
		}

		providers[definition.Slug] = struct{}{}
		for _, identifier := range append([]string{definition.Slug}, definition.Aliases...) {
			if owner, duplicate := identifiers[identifier]; duplicate {
				return fmt.Errorf("provider registry: identifier %q belongs to both %q and %q", identifier, owner, definition.Slug)
			}

			identifiers[identifier] = definition.Slug
		}
	}

	return nil
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
				return cloneDefinition(definition), true
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

func cloneDefinition(definition Definition) Definition {
	definition.Aliases = append([]string(nil), definition.Aliases...)
	definition.Capabilities = append([]Capability(nil), definition.Capabilities...)
	definition.OperationNamespaces = append([]string(nil), definition.OperationNamespaces...)
	definition.Operations = append([]platform.ProviderOperation(nil), definition.Operations...)
	definition.LegacyModelPrefixes = append([]string(nil), definition.LegacyModelPrefixes...)

	return definition
}

func validateDefinition(definition Definition) error {
	if definition.Slug == "" || definition.Slug != strings.ToLower(strings.TrimSpace(definition.Slug)) {
		return fmt.Errorf("provider registry: invalid slug %q", definition.Slug)
	}

	if strings.TrimSpace(definition.DisplayName) == "" {
		return fmt.Errorf("provider registry: provider %q has no display name", definition.Slug)
	}

	if !validPricingSnapshot(definition.PricingSnapshot) {
		return fmt.Errorf("provider registry: provider %q has invalid pricing snapshot %q", definition.Slug, definition.PricingSnapshot)
	}

	if !validPricingStrategy(definition.PricingStrategy) {
		return fmt.Errorf("provider registry: provider %q has invalid pricing strategy %q", definition.Slug, definition.PricingStrategy)
	}

	if len(definition.Capabilities) == 0 {
		return fmt.Errorf("provider registry: provider %q has no capabilities", definition.Slug)
	}

	if len(definition.OperationNamespaces) == 0 && len(definition.Operations) == 0 {
		return fmt.Errorf("provider registry: provider %q has no operations", definition.Slug)
	}

	if err := validateUniqueStrings(definition.Slug, "alias", definition.Aliases); err != nil {
		return err
	}

	if err := validateUniqueStrings(definition.Slug, "operation namespace", definition.OperationNamespaces); err != nil {
		return err
	}

	for _, namespace := range definition.OperationNamespaces {
		if !strings.HasSuffix(namespace, "_") {
			return fmt.Errorf(
				"provider registry: provider %q operation namespace %q must end with an underscore",
				definition.Slug,
				namespace,
			)
		}
	}

	if err := validateUniqueStrings(definition.Slug, "legacy model prefix", definition.LegacyModelPrefixes); err != nil {
		return err
	}

	capabilities := make(map[Capability]struct{}, len(definition.Capabilities))
	for _, capability := range definition.Capabilities {
		if !validCapability(capability) {
			return fmt.Errorf("provider registry: provider %q has invalid capability %q", definition.Slug, capability)
		}

		if _, duplicate := capabilities[capability]; duplicate {
			return fmt.Errorf("provider registry: provider %q repeats capability %q", definition.Slug, capability)
		}

		capabilities[capability] = struct{}{}
	}

	operations := make(map[platform.ProviderOperation]struct{}, len(definition.Operations))
	for _, operation := range definition.Operations {
		value := string(operation)
		if value == "" || value != strings.ToLower(strings.TrimSpace(value)) {
			return fmt.Errorf("provider registry: provider %q has invalid operation %q", definition.Slug, operation)
		}

		if _, duplicate := operations[operation]; duplicate {
			return fmt.Errorf("provider registry: provider %q repeats operation %q", definition.Slug, operation)
		}

		operations[operation] = struct{}{}
	}

	return nil
}

func validateUniqueStrings(provider string, field string, values []string) error {
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if value == "" || value != strings.ToLower(strings.TrimSpace(value)) {
			return fmt.Errorf("provider registry: provider %q has invalid %s %q", provider, field, value)
		}

		if _, duplicate := seen[value]; duplicate {
			return fmt.Errorf("provider registry: provider %q repeats %s %q", provider, field, value)
		}

		seen[value] = struct{}{}
	}

	return nil
}

func validPricingSnapshot(kind PricingSnapshotKind) bool {
	return kind == PricingSnapshotAnthropic || kind == PricingSnapshotCatalog
}

func validPricingStrategy(kind PricingStrategyKind) bool {
	switch kind {
	case PricingStrategyStandard,
		PricingStrategyOpenAI,
		PricingStrategyAnthropic,
		PricingStrategyXAI,
		PricingStrategyDeepSeek,
		PricingStrategyZAI,
		PricingStrategyMoonshot,
		PricingStrategyGoogleGemini,
		PricingStrategyMistral:
		return true
	default:
		return false
	}
}

func validCapability(capability Capability) bool {
	switch capability {
	case CapabilityChatCompletions,
		CapabilityCompletions,
		CapabilityResponses,
		CapabilityEmbeddings,
		CapabilityModerations,
		CapabilityImages,
		CapabilityAudio,
		CapabilityBatches,
		CapabilityFiles,
		CapabilityModels,
		CapabilityRealtime,
		CapabilityTokenCounting,
		CapabilityUploads,
		CapabilityVectorStores,
		CapabilityContainers,
		CapabilityVideos,
		CapabilityInteractions,
		CapabilityContextCaching,
		CapabilityOCR,
		CapabilityAgents,
		CapabilityConversations:
		return true
	default:
		return false
	}
}
