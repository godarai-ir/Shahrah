package proxy

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
)

func TestValidateXAIChatRuntimeProfileEnforcesModelContract(t *testing.T) {
	binding := xaiProxyTestRuntimeProfile()

	for name, test := range map[string]struct {
		raw       string
		wantError string
	}{
		"supported": {
			raw: `{"model":"grok-4.5-latest","messages":[{"role":"user","content":"hi"}],"reasoning_effort":"low","tools":[{"type":"x_search"}]}`,
		},
		"reasoning cannot be disabled": {
			raw:       `{"model":"grok-4.5","messages":[{"role":"user","content":"hi"}],"reasoning_effort":"none"}`,
			wantError: "reasoning_effort=none",
		},
		"logprobs pass through": {
			raw: `{"model":"grok-4.5","messages":[{"role":"user","content":"hi"}],"logprobs":true}`,
		},
		"reasoning penalties pass through": {
			raw: `{"model":"grok-4.5","messages":[{"role":"user","content":"hi"}],"presence_penalty":0.5}`,
		},
		"unknown server tool": {
			raw:       `{"model":"grok-4.5","messages":[{"role":"user","content":"hi"}],"tools":[{"type":"unknown_search"}]}`,
			wantError: "does not support xAI tool",
		},
	} {
		t.Run(name, func(t *testing.T) {
			var request contractsopenai.ChatCompletionRequest
			if err := json.Unmarshal([]byte(test.raw), &request); err != nil {
				t.Fatal(err)
			}

			err := validateXAIChatRuntimeProfile(request, &binding)
			if test.wantError == "" && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if test.wantError != "" && (err == nil || !strings.Contains(err.Error(), test.wantError)) {
				t.Fatalf("expected %q, got %v", test.wantError, err)
			}
		})
	}
}

func TestRuntimeProfilesFromEstimateIncludesXAI(t *testing.T) {
	binding := xaiProxyTestRuntimeProfile()

	profiles, err := runtimeProfilesFromEstimate("xai", "grok-4.5-latest", platform.UsageCostBreakdown{
		PricingRevision:     binding.CatalogRevision,
		ModelRuntimeProfile: &binding,
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(profiles) != 1 || !profiles[0].MatchesModel("grok-4.5-latest") {
		t.Fatalf("unexpected xAI runtime profile: %+v", profiles)
	}
}

func xaiProxyTestRuntimeProfile() platform.ModelRuntimeProfileBinding {
	profile := platform.ModelRuntimeProfile{
		Version:    3,
		Aliases:    []string{"grok-4.5-latest", "grok-build-latest"},
		Endpoints:  []string{"chat_completions", "deferred_chat_completions", "responses", "responses_compact", "files", "file_public_urls"},
		Transports: []string{"http", "sse"},
		Lifecycle:  "active",
		Thinking: platform.ModelRuntimeThinkingProfile{
			DefaultMode:   "adaptive",
			AllowedModes:  []string{"adaptive"},
			EffortValues:  []string{"low", "medium", "high"},
			AlwaysEnabled: true,
		},
		Sampling: platform.ModelRuntimeSamplingProfile{Parameters: []string{"temperature", "top_p"}},
		Tools: platform.ModelRuntimeToolsProfile{
			Custom:           true,
			Server:           true,
			Parallel:         true,
			StructuredOutput: true,
			Supported:        []string{"function_calling", "web_search", "x_search", "code_execution", "attachment_search", "collections_search", "file_search", "mcp"},
		},
		Modalities: platform.ModelRuntimeModalities{Input: []string{"text", "image"}, Output: []string{"text"}},
		RequestControls: platform.ModelRuntimeRequestControls{
			ServiceTierValues:  []string{"default", "priority"},
			InferenceGeoValues: []string{"global"},
		},
		RequestHeaders: platform.ModelRuntimeRequestHeaders{
			Allowed:    []string{"x-grok-conv-id"},
			ByEndpoint: map[string][]string{"chat_completions": {"x-grok-conv-id"}},
		},
		DeferredCompletions:   platform.ModelRuntimeDeferredProfile{Eligible: true, RetentionSeconds: 86400, MaximumRetrievals: 1},
		UnsupportedParameters: []string{"logprobs", "top_logprobs", "presence_penalty", "frequency_penalty", "stop"},
		Limits: platform.ModelRuntimeLimits{
			ContextWindow:  500000,
			MaxInputTokens: 500000,
			MessagesBytes:  32 * 1024 * 1024,
			BatchBytes:     25 * 1024 * 1024,
			FileBytes:      48 * 1024 * 1024,
		},
	}

	return platform.ModelRuntimeProfileBinding{
		Provider:        "xai",
		CatalogRevision: "xai-test-revision",
		CanonicalID:     "xai/grok-4.5",
		Slug:            "grok-4.5",
		ProfileVersion:  3,
		Profile:         profile,
		SourceURL:       "https://docs.x.ai/developers/models/grok-4.5",
		EffectiveFrom:   time.Now().UTC(),
	}
}
