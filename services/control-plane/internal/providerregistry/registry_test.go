package providerregistry

import (
	"reflect"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

func TestDefinitionsCoverFirstClassProviders(t *testing.T) {
	want := map[string]bool{
		"openai":        false,
		"anthropic":     true,
		"xai":           true,
		"deepseek":      true,
		"zai":           true,
		"moonshot":      true,
		"google-gemini": true,
		"mistral-ai":    true,
	}

	got := make(map[string]bool, len(definitions))

	for _, definition := range Definitions() {
		if definition.PricingSnapshot == "" {
			t.Fatalf("provider %q has no pricing snapshot", definition.Slug)
		}

		got[definition.Slug] = definition.DispatchRuntimeProfileRequired
	}

	if len(got) != len(want) {
		t.Fatalf("expected %d providers, got %d", len(want), len(got))
	}

	for provider, runtimeProfileRequired := range want {
		if got[provider] != runtimeProfileRequired {
			t.Fatalf("provider %q runtime-profile requirement = %t, want %t", provider, got[provider], runtimeProfileRequired)
		}
	}
}

func TestDefinitionsSatisfyProviderContracts(t *testing.T) {
	if err := Validate(Definitions()); err != nil {
		t.Fatalf("provider definitions are invalid: %v", err)
	}

	type providerContract struct {
		pricingSnapshot     PricingSnapshotKind
		pricingStrategy     PricingStrategyKind
		aliases             []string
		capabilities        []Capability
		operationNamespaces []string
		operations          []platform.ProviderOperation
	}

	contracts := map[string]providerContract{
		"openai": {
			pricingSnapshot:     PricingSnapshotCatalog,
			pricingStrategy:     PricingStrategyOpenAI,
			operationNamespaces: []string{"openai_"},
			capabilities: []Capability{
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
			operations: []platform.ProviderOperation{
				platform.ProviderOperationCredentialProbe,
				platform.ProviderOperationChatCompletions,
			},
		},
		"anthropic": {
			pricingSnapshot:     PricingSnapshotAnthropic,
			pricingStrategy:     PricingStrategyAnthropic,
			operationNamespaces: []string{"anthropic_"},
			capabilities: []Capability{
				CapabilityChatCompletions,
				CapabilityTokenCounting,
				CapabilityBatches,
				CapabilityFiles,
				CapabilityModels,
			},
			operations: []platform.ProviderOperation{
				platform.ProviderOperationCredentialProbe,
				platform.ProviderOperationChatCompletions,
			},
		},
		"xai": {
			pricingSnapshot:     PricingSnapshotCatalog,
			pricingStrategy:     PricingStrategyXAI,
			operationNamespaces: []string{"xai_"},
			capabilities: []Capability{
				CapabilityChatCompletions,
				CapabilityResponses,
				CapabilityImages,
				CapabilityBatches,
				CapabilityFiles,
				CapabilityModels,
			},
			operations: []platform.ProviderOperation{
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
		},
		"deepseek": {
			pricingSnapshot: PricingSnapshotCatalog,
			pricingStrategy: PricingStrategyDeepSeek,
			capabilities:    []Capability{CapabilityChatCompletions},
			operations: []platform.ProviderOperation{
				platform.ProviderOperationCredentialProbe,
				platform.ProviderOperationChatCompletions,
			},
		},
		"zai": {
			pricingSnapshot:     PricingSnapshotCatalog,
			pricingStrategy:     PricingStrategyZAI,
			operationNamespaces: []string{"zai_"},
			capabilities:        []Capability{CapabilityChatCompletions, CapabilityOCR},
			operations: []platform.ProviderOperation{
				platform.ProviderOperationCredentialProbe,
				platform.ProviderOperationChatCompletions,
				platform.ProviderOperationZAILayoutParsing,
			},
		},
		"moonshot": {
			pricingSnapshot: PricingSnapshotCatalog,
			pricingStrategy: PricingStrategyMoonshot,
			capabilities:    []Capability{CapabilityChatCompletions},
			operations: []platform.ProviderOperation{
				platform.ProviderOperationCredentialProbe,
				platform.ProviderOperationChatCompletions,
			},
		},
		"google-gemini": {
			pricingSnapshot:     PricingSnapshotCatalog,
			pricingStrategy:     PricingStrategyGoogleGemini,
			aliases:             []string{"gemini"},
			operationNamespaces: []string{"gemini_"},
			capabilities: []Capability{
				CapabilityChatCompletions,
				CapabilityEmbeddings,
				CapabilityBatches,
				CapabilityFiles,
				CapabilityModels,
				CapabilityTokenCounting,
				CapabilityInteractions,
				CapabilityContextCaching,
			},
			operations: []platform.ProviderOperation{
				platform.ProviderOperationCredentialProbe,
				platform.ProviderOperationChatCompletions,
			},
		},
		"mistral-ai": {
			pricingSnapshot:     PricingSnapshotCatalog,
			pricingStrategy:     PricingStrategyMistral,
			aliases:             []string{"mistral"},
			operationNamespaces: []string{"mistral_"},
			capabilities: []Capability{
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
			operations: []platform.ProviderOperation{
				platform.ProviderOperationCredentialProbe,
				platform.ProviderOperationChatCompletions,
				platform.ProviderOperationOpenAICompletions,
				platform.ProviderOperationOpenAIAudioSpeech,
				platform.ProviderOperationOpenAIAudioTranscriptions,
				platform.ProviderOperationOpenAIAudioTranslations,
			},
		},
	}

	definitions := Definitions()
	if len(definitions) != len(contracts) {
		t.Fatalf("expected %d provider contracts, got %d definitions", len(contracts), len(definitions))
	}

	for _, definition := range definitions {
		contract, ok := contracts[definition.Slug]
		if !ok {
			t.Fatalf("provider %q has no contract", definition.Slug)
		}

		if definition.PricingStrategy != contract.pricingStrategy {
			t.Errorf(
				"provider %q pricing strategy = %q, want %q",
				definition.Slug,
				definition.PricingStrategy,
				contract.pricingStrategy,
			)
		}

		if definition.PricingSnapshot != contract.pricingSnapshot {
			t.Errorf(
				"provider %q pricing snapshot = %q, want %q",
				definition.Slug,
				definition.PricingSnapshot,
				contract.pricingSnapshot,
			)
		}

		if !reflect.DeepEqual(definition.Aliases, contract.aliases) {
			t.Errorf("provider %q aliases = %v, want %v", definition.Slug, definition.Aliases, contract.aliases)
		}

		if !reflect.DeepEqual(definition.Capabilities, contract.capabilities) {
			t.Errorf(
				"provider %q capabilities = %v, want %v",
				definition.Slug,
				definition.Capabilities,
				contract.capabilities,
			)
		}

		if !reflect.DeepEqual(definition.OperationNamespaces, contract.operationNamespaces) {
			t.Errorf(
				"provider %q operation namespaces = %v, want %v",
				definition.Slug,
				definition.OperationNamespaces,
				contract.operationNamespaces,
			)
		}

		if !reflect.DeepEqual(definition.Operations, contract.operations) {
			t.Errorf(
				"provider %q operations = %v, want %v",
				definition.Slug,
				definition.Operations,
				contract.operations,
			)
		}

		for _, capability := range contract.capabilities {
			if !definition.Supports(capability) {
				t.Errorf("provider %q does not support capability %q", definition.Slug, capability)
			}
		}

		for _, operation := range contract.operations {
			if !definition.SupportsOperation(operation) {
				t.Errorf("provider %q does not support operation %q", definition.Slug, operation)
			}
		}

		for _, namespace := range contract.operationNamespaces {
			operation := platform.ProviderOperation(namespace + "contract_probe")
			if !definition.SupportsOperation(operation) {
				t.Errorf("provider %q does not support namespace %q", definition.Slug, namespace)
			}
		}
	}
}

func TestDefinitionsAreImmutableSnapshots(t *testing.T) {
	items := Definitions()
	items[0].Aliases = append(items[0].Aliases, "mutated")
	items[0].Capabilities[0] = Capability("mutated")
	items[0].Operations[0] = platform.ProviderOperation("mutated")

	definition, ok := Lookup("openai")
	if !ok {
		t.Fatal("OpenAI definition is missing")
	}

	if containsFold(definition.Aliases, "mutated") ||
		definition.Supports(Capability("mutated")) ||
		definition.SupportsOperation(platform.ProviderOperation("mutated")) {
		t.Fatal("provider registry leaked mutable definition state")
	}
}

func TestFromModelIdentifierResolvesEveryFirstClassProvider(t *testing.T) {
	tests := map[string]string{
		"openai:primary:gpt-5":               "openai",
		"anthropic:primary:claude-opus-5":    "anthropic",
		"xai:primary:grok-4.5":               "xai",
		"deepseek:primary:deepseek-v4-flash": "deepseek",
		"moonshot:primary:kimi-k3":           "moonshot",
		"gemini/gemini-3.5-flash":            "google-gemini",
		"mistral:primary:mistral-large-2607": "mistral-ai",
		"claude-sonnet-5":                    "anthropic",
		"voxtral-mini-2602":                  "mistral-ai",
	}

	for model, want := range tests {
		t.Run(model, func(t *testing.T) {
			definition, ok := FromModelIdentifier(model)
			if !ok || definition.Slug != want {
				t.Fatalf("FromModelIdentifier(%q) = %q, %t; want %q", model, definition.Slug, ok, want)
			}
		})
	}
}

func TestAliasesResolveToCanonicalProviders(t *testing.T) {
	for alias, want := range map[string]string{
		"gemini":  "google-gemini",
		"mistral": "mistral-ai",
	} {
		definition, ok := Lookup(alias)
		if !ok || definition.Slug != want {
			t.Fatalf("Lookup(%q) = %q, %t; want %q", alias, definition.Slug, ok, want)
		}
	}
}
