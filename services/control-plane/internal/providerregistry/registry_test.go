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
		"tokenbazaar":   true,
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
		"tokenbazaar": {
			pricingSnapshot: PricingSnapshotCatalog,
			pricingStrategy: PricingStrategyTokenBazaar,
			capabilities: []Capability{
				CapabilityChatCompletions, CapabilityImages, CapabilityEmbeddings,
				CapabilityResponses, CapabilityMessages, CapabilityGenerateContent,
			},
			operations: []platform.ProviderOperation{
				platform.ProviderOperationCredentialProbe,
				platform.ProviderOperationChatCompletions,
				platform.ProviderOperationOpenAIImagesGenerations,
				platform.ProviderOperationOpenAIImagesEdits,
				platform.ProviderOperationOpenAIEmbeddings,
				platform.ProviderOperationOpenAIResponsesCreate,
				platform.ProviderOperationAnthropicMessages,
				platform.ProviderOperationGeminiGenerateContent,
				platform.ProviderOperationGeminiStreamGenerateContent,
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
		"openai:primary:gpt-5":                  "openai",
		"anthropic:primary:claude-opus-5":       "anthropic",
		"xai:primary:grok-4.5":                  "xai",
		"deepseek:primary:deepseek-v4-flash":    "deepseek",
		"moonshot:primary:kimi-k3":              "moonshot",
		"gemini/gemini-3.5-flash":               "google-gemini",
		"mistral:primary:mistral-large-2607":    "mistral-ai",
		"claude-sonnet-5":                       "anthropic",
		"voxtral-mini-2602":                     "mistral-ai",
		"tokenbazaar:primary:claude-sonnet-5":   "tokenbazaar",
		"tokenbazaar:primary:deepseek-v4-flash": "tokenbazaar",
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

// TestTokenBazaarNeverCollidesWithDirectProviders asserts the reseller
// "tokenbazaar" slug is distinct from every direct provider it fronts, even
// though many TokenBazaar model IDs (e.g. "claude-sonnet-5") intentionally
// reuse a direct provider's model-id string. A bare, unqualified model id
// must keep resolving to the direct provider (legacy behavior); only the
// explicit "tokenbazaar:<account>:<model-id>" FQN reaches the reseller.
func TestTokenBazaarNeverCollidesWithDirectProviders(t *testing.T) {
	definition, ok := Lookup("tokenbazaar")
	if !ok || definition.Slug != "tokenbazaar" {
		t.Fatalf("tokenbazaar is not registered as a first-class provider")
	}

	for _, other := range []string{"anthropic", "deepseek", "xai", "moonshot", "google-gemini", "openai"} {
		if definition.Slug == other {
			t.Fatalf("tokenbazaar slug collides with direct provider %q", other)
		}
	}

	bare, ok := FromModelIdentifier("claude-sonnet-5")
	if !ok || bare.Slug != "anthropic" {
		t.Fatalf("bare model id resolved to %q, %t; want anthropic (unqualified ids must not route to tokenbazaar)", bare.Slug, ok)
	}

	qualified, ok := FromModelIdentifier("tokenbazaar:primary:claude-sonnet-5")
	if !ok || qualified.Slug != "tokenbazaar" {
		t.Fatalf("FQN model id resolved to %q, %t; want tokenbazaar", qualified.Slug, ok)
	}
}

// TestTokenBazaarSupportsImagesCapability asserts the 4 TokenBazaar image
// models' operations (POST /v1/images/generations, POST /v1/images/edits)
// are registered as first-class TokenBazaar operations/capabilities, per
// plans/tokenbazaar-production-readiness.md's image-endpoint delivery.
func TestTokenBazaarSupportsImagesCapability(t *testing.T) {
	definition, ok := Lookup("tokenbazaar")
	if !ok {
		t.Fatalf("tokenbazaar is not registered as a first-class provider")
	}

	if !definition.Supports(CapabilityImages) {
		t.Fatal("tokenbazaar does not declare CapabilityImages")
	}

	if !definition.Supports(CapabilityChatCompletions) {
		t.Fatal("tokenbazaar lost CapabilityChatCompletions while gaining CapabilityImages")
	}

	for _, operation := range []platform.ProviderOperation{
		platform.ProviderOperationOpenAIImagesGenerations,
		platform.ProviderOperationOpenAIImagesEdits,
	} {
		if !definition.SupportsOperation(operation) {
			t.Fatalf("tokenbazaar does not support operation %q", operation)
		}
	}
}

// TestTokenBazaarImageModelsNeverCollideWithDirectProviders asserts that the
// 4 TokenBazaar image model IDs, even though they are string-identical to
// the corresponding direct-provider model IDs (OpenAI's "gpt-image-2" and
// Google Gemini's 3 image models), only ever route to "tokenbazaar" via the
// explicit "tokenbazaar:<account>:<model-id>" FQN — a bare model id must
// keep resolving to the direct provider that actually owns it.
func TestTokenBazaarImageModelsNeverCollideWithDirectProviders(t *testing.T) {
	tests := []struct {
		bareModel      string
		wantBareOwner  string
		qualifiedModel string
	}{
		{"gpt-image-2", "", "tokenbazaar:primary:gpt-image-2"},
		{"gemini-3-pro-image", "", "tokenbazaar:primary:gemini-3-pro-image"},
		{"gemini-3.1-flash-image", "", "tokenbazaar:primary:gemini-3.1-flash-image"},
		{"gemini-3.1-flash-lite-image", "", "tokenbazaar:primary:gemini-3.1-flash-lite-image"},
	}

	for _, test := range tests {
		t.Run(test.bareModel, func(t *testing.T) {
			// A bare image-model id must never resolve to tokenbazaar: this
			// registry carries no OperationNamespaces/LegacyModelPrefixes for
			// tokenbazaar, so FromModelIdentifier on an unqualified id either
			// resolves to a direct provider (via its own legacy prefix, if
			// any) or does not resolve at all — either way, never
			// "tokenbazaar".
			if bare, ok := FromModelIdentifier(test.bareModel); ok && bare.Slug == "tokenbazaar" {
				t.Fatalf("bare model id %q resolved to tokenbazaar; must only be reachable via an explicit FQN", test.bareModel)
			}

			qualified, ok := FromModelIdentifier(test.qualifiedModel)
			if !ok || qualified.Slug != "tokenbazaar" {
				t.Fatalf("FromModelIdentifier(%q) = %q, %t; want tokenbazaar", test.qualifiedModel, qualified.Slug, ok)
			}
		})
	}
}

// TestTokenBazaarSupportsEmbeddingsCapability asserts text-embedding-3-large's
// operation (POST /v1/embeddings) is registered as a first-class TokenBazaar
// operation/capability, per plans/tokenbazaar-production-readiness.md's
// embeddings-endpoint delivery, and that gaining it did not disturb the
// already-registered chat and images capabilities.
func TestTokenBazaarSupportsEmbeddingsCapability(t *testing.T) {
	definition, ok := Lookup("tokenbazaar")
	if !ok {
		t.Fatalf("tokenbazaar is not registered as a first-class provider")
	}

	if !definition.Supports(CapabilityEmbeddings) {
		t.Fatal("tokenbazaar does not declare CapabilityEmbeddings")
	}

	if !definition.Supports(CapabilityChatCompletions) {
		t.Fatal("tokenbazaar lost CapabilityChatCompletions while gaining CapabilityEmbeddings")
	}

	if !definition.Supports(CapabilityImages) {
		t.Fatal("tokenbazaar lost CapabilityImages while gaining CapabilityEmbeddings")
	}

	if !definition.SupportsOperation(platform.ProviderOperationOpenAIEmbeddings) {
		t.Fatalf("tokenbazaar does not support operation %q", platform.ProviderOperationOpenAIEmbeddings)
	}
}

// TestTokenBazaarEmbeddingsModelNeverCollidesWithDirectProviders asserts that
// "text-embedding-3-large", even though it is string-identical to the direct
// OpenAI provider's own embeddings model id, only ever routes to
// "tokenbazaar" via the explicit "tokenbazaar:<account>:<model-id>" FQN — a
// bare model id must keep resolving to the direct provider that actually
// owns it (or fail to resolve), and must never collide with, or shadow, the
// direct "openai" provider's own embeddings dispatch path.
func TestTokenBazaarEmbeddingsModelNeverCollidesWithDirectProviders(t *testing.T) {
	const model = "text-embedding-3-large"

	if bare, ok := FromModelIdentifier(model); ok && bare.Slug == "tokenbazaar" {
		t.Fatalf("bare model id %q resolved to tokenbazaar; must only be reachable via an explicit FQN", model)
	}

	qualified, ok := FromModelIdentifier("tokenbazaar:primary:" + model)
	if !ok || qualified.Slug != "tokenbazaar" {
		t.Fatalf("FromModelIdentifier(%q) = %q, %t; want tokenbazaar", "tokenbazaar:primary:"+model, qualified.Slug, ok)
	}

	// The direct "openai" provider must still resolve independently and must
	// never be confused with the "tokenbazaar" registration.
	openaiDefinition, ok := Lookup("openai")
	if !ok {
		t.Fatalf("openai is not registered as a first-class provider")
	}

	if openaiDefinition.Slug == qualified.Slug {
		t.Fatalf("openai and tokenbazaar registrations unexpectedly collide for model %q", model)
	}
}

// TestTokenBazaarSupportsResponsesCapability asserts the 6 in-scope
// TokenBazaar Responses models' operation (POST /v1/responses, create-only)
// is registered as a first-class TokenBazaar operation/capability, per
// plans/tokenbazaar-production-readiness.md's Responses-endpoint delivery,
// and that gaining it did not disturb the already-registered chat, images,
// and embeddings capabilities.
func TestTokenBazaarSupportsResponsesCapability(t *testing.T) {
	definition, ok := Lookup("tokenbazaar")
	if !ok {
		t.Fatalf("tokenbazaar is not registered as a first-class provider")
	}

	if !definition.Supports(CapabilityResponses) {
		t.Fatal("tokenbazaar does not declare CapabilityResponses")
	}

	if !definition.Supports(CapabilityChatCompletions) {
		t.Fatal("tokenbazaar lost CapabilityChatCompletions while gaining CapabilityResponses")
	}

	if !definition.Supports(CapabilityImages) {
		t.Fatal("tokenbazaar lost CapabilityImages while gaining CapabilityResponses")
	}

	if !definition.Supports(CapabilityEmbeddings) {
		t.Fatal("tokenbazaar lost CapabilityEmbeddings while gaining CapabilityResponses")
	}

	if !definition.SupportsOperation(platform.ProviderOperationOpenAIResponsesCreate) {
		t.Fatalf("tokenbazaar does not support operation %q", platform.ProviderOperationOpenAIResponsesCreate)
	}

	// This delivery is deliberately "create only": TokenBazaar's stateful
	// Responses resource surface (GET/DELETE/CANCEL/InputItems/Compact/
	// InputTokens) is not a capability TokenBazaar's own docs describe, and
	// must never be registered here as a side effect of adding create.
	for _, resourceOperation := range []platform.ProviderOperation{
		platform.ProviderOperationOpenAIResponsesGet,
		platform.ProviderOperationOpenAIResponsesDelete,
		platform.ProviderOperationOpenAIResponsesCancel,
		platform.ProviderOperationOpenAIResponsesInputItems,
		platform.ProviderOperationOpenAIResponsesCompact,
		platform.ProviderOperationOpenAIResponsesInputTokens,
	} {
		if definition.SupportsOperation(resourceOperation) {
			t.Fatalf("tokenbazaar must not support stateful Responses resource operation %q (create-only delivery)", resourceOperation)
		}
	}
}

// TestTokenBazaarResponsesModelsNeverCollideWithDirectProviders asserts that
// the 6 in-scope Responses model IDs (4 OpenAI-family, 2 DeepSeek), even
// though they are string-identical to the corresponding direct provider's
// model IDs, only ever route to "tokenbazaar" via the explicit
// "tokenbazaar:<account>:<model-id>" FQN -- a bare model id must keep
// resolving to the direct provider that actually owns it (or fail to
// resolve), never to "tokenbazaar".
func TestTokenBazaarResponsesModelsNeverCollideWithDirectProviders(t *testing.T) {
	for _, model := range []string{
		"gpt-5.5", "gpt-5.6-luna", "gpt-5.6-sol", "gpt-5.6-terra",
		"deepseek-v4-flash", "deepseek-v4-pro",
	} {
		t.Run(model, func(t *testing.T) {
			if bare, ok := FromModelIdentifier(model); ok && bare.Slug == "tokenbazaar" {
				t.Fatalf("bare model id %q resolved to tokenbazaar; must only be reachable via an explicit FQN", model)
			}

			qualified, ok := FromModelIdentifier("tokenbazaar:primary:" + model)
			if !ok || qualified.Slug != "tokenbazaar" {
				t.Fatalf("FromModelIdentifier(%q) = %q, %t; want tokenbazaar", "tokenbazaar:primary:"+model, qualified.Slug, ok)
			}
		})
	}
}

func TestTokenBazaarSupportsMessagesCapability(t *testing.T) {
	definition, ok := Lookup("tokenbazaar")
	if !ok {
		t.Fatalf("tokenbazaar is not registered as a first-class provider")
	}

	if !definition.Supports(CapabilityMessages) {
		t.Fatal("tokenbazaar does not declare CapabilityMessages")
	}

	if !definition.Supports(CapabilityChatCompletions) {
		t.Fatal("tokenbazaar lost CapabilityChatCompletions while gaining CapabilityMessages")
	}

	if !definition.Supports(CapabilityResponses) {
		t.Fatal("tokenbazaar lost CapabilityResponses while gaining CapabilityMessages")
	}

	if !definition.SupportsOperation(platform.ProviderOperationAnthropicMessages) {
		t.Fatalf("tokenbazaar does not support operation %q", platform.ProviderOperationAnthropicMessages)
	}

	// This delivery is deliberately "create only": TokenBazaar's own docs
	// show a single stateless request/response shape for /v1/messages, with
	// no count_tokens, message_batches, or files surface. None of those must
	// ever be registered here as a side effect of adding Messages create.
	for _, unsupported := range []platform.ProviderOperation{
		platform.ProviderOperationAnthropicCountTokens,
		platform.ProviderOperationAnthropicBatchesCreate,
		platform.ProviderOperationAnthropicBatchesList,
		platform.ProviderOperationAnthropicBatchesGet,
		platform.ProviderOperationAnthropicBatchesCancel,
		platform.ProviderOperationAnthropicBatchesResults,
		platform.ProviderOperationAnthropicFilesCreate,
		platform.ProviderOperationAnthropicFilesList,
		platform.ProviderOperationAnthropicFilesGet,
		platform.ProviderOperationAnthropicFilesDelete,
		platform.ProviderOperationAnthropicFilesContent,
		platform.ProviderOperationAnthropicModelsList,
		platform.ProviderOperationAnthropicModelsGet,
	} {
		if definition.SupportsOperation(unsupported) {
			t.Fatalf("tokenbazaar must not support Anthropic operation %q (Messages-create-only delivery)", unsupported)
		}
	}
}

// TestTokenBazaarMessagesModelsNeverCollideWithDirectProviders asserts that
// the 4 in-scope Messages model IDs, even though they are string-identical to
// the direct "anthropic" provider's own model IDs, only ever route to
// "tokenbazaar" via the explicit "tokenbazaar:<account>:<model-id>" FQN -- a
// bare model id must keep resolving to the direct "anthropic" provider (or
// fail to resolve), never to "tokenbazaar".
func TestTokenBazaarMessagesModelsNeverCollideWithDirectProviders(t *testing.T) {
	for _, model := range []string{
		"claude-fable-5", "claude-haiku-4-5", "claude-opus-5", "claude-sonnet-5",
	} {
		t.Run(model, func(t *testing.T) {
			if bare, ok := FromModelIdentifier(model); ok && bare.Slug == "tokenbazaar" {
				t.Fatalf("bare model id %q resolved to tokenbazaar; must only be reachable via an explicit FQN", model)
			}

			qualified, ok := FromModelIdentifier("tokenbazaar:primary:" + model)
			if !ok || qualified.Slug != "tokenbazaar" {
				t.Fatalf("FromModelIdentifier(%q) = %q, %t; want tokenbazaar", "tokenbazaar:primary:"+model, qualified.Slug, ok)
			}
		})
	}
}

// TestTokenBazaarSupportsGenerateContentCapability asserts the 6 in-scope
// Gemini text models' Gemini-native operations
// (POST /v1beta/models/{model}:generateContent and
// .../{model}:streamGenerateContent) are registered as first-class
// TokenBazaar operations/capabilities, per
// plans/tokenbazaar-production-readiness.md's native Gemini generateContent
// delivery.
func TestTokenBazaarSupportsGenerateContentCapability(t *testing.T) {
	definition, ok := Lookup("tokenbazaar")
	if !ok {
		t.Fatalf("tokenbazaar is not registered as a first-class provider")
	}

	if !definition.Supports(CapabilityGenerateContent) {
		t.Fatal("tokenbazaar does not declare CapabilityGenerateContent")
	}

	if !definition.Supports(CapabilityChatCompletions) {
		t.Fatal("tokenbazaar lost CapabilityChatCompletions while gaining CapabilityGenerateContent")
	}

	if !definition.Supports(CapabilityMessages) {
		t.Fatal("tokenbazaar lost CapabilityMessages while gaining CapabilityGenerateContent")
	}

	for _, operation := range []platform.ProviderOperation{
		platform.ProviderOperationGeminiGenerateContent,
		platform.ProviderOperationGeminiStreamGenerateContent,
	} {
		if !definition.SupportsOperation(operation) {
			t.Fatalf("tokenbazaar does not support operation %q", operation)
		}
	}

	// This delivery is scoped to exactly generateContent/streamGenerateContent
	// (create-only, text models only). None of Gemini's other native
	// operations (caching, interactions, files, batches, countTokens,
	// embedContent) must be registered here as a side effect.
	for _, unsupported := range []platform.ProviderOperation{
		platform.ProviderOperationGeminiCountTokens,
		platform.ProviderOperationGeminiEmbedContent,
		platform.ProviderOperationGeminiBatchEmbedContents,
		platform.ProviderOperationGeminiInteractionsCreate,
		platform.ProviderOperationGeminiCachesCreate,
		platform.ProviderOperationGeminiCachesList,
		platform.ProviderOperationGeminiCachesGet,
		platform.ProviderOperationGeminiCachesUpdate,
		platform.ProviderOperationGeminiCachesDelete,
		platform.ProviderOperationGeminiBatchesCreate,
		platform.ProviderOperationGeminiFilesCreate,
		platform.ProviderOperationGeminiModelsList,
	} {
		if definition.SupportsOperation(unsupported) {
			t.Fatalf("tokenbazaar must not support Gemini operation %q (generateContent-only delivery)", unsupported)
		}
	}
}

// TestTokenBazaarGenerateContentModelsNeverCollideWithDirectProviders asserts
// that the 6 in-scope generateContent model IDs, even though they are
// string-identical to the direct "google-gemini" provider's own model IDs,
// only ever route to "tokenbazaar" via the explicit
// "tokenbazaar:<account>:<model-id>" FQN.
func TestTokenBazaarGenerateContentModelsNeverCollideWithDirectProviders(t *testing.T) {
	for _, model := range []string{
		"gemini-2.5-flash", "gemini-2.5-pro", "gemini-3.1-pro-preview",
		"gemini-3.5-flash", "gemini-3.6-flash", "gemini-3.7-flash",
	} {
		t.Run(model, func(t *testing.T) {
			if bare, ok := FromModelIdentifier(model); ok && bare.Slug == "tokenbazaar" {
				t.Fatalf("bare model id %q resolved to tokenbazaar; must only be reachable via an explicit FQN", model)
			}

			qualified, ok := FromModelIdentifier("tokenbazaar:primary:" + model)
			if !ok || qualified.Slug != "tokenbazaar" {
				t.Fatalf("FromModelIdentifier(%q) = %q, %t; want tokenbazaar", "tokenbazaar:primary:"+model, qualified.Slug, ok)
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
