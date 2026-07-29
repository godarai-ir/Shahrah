package platform

import (
	"encoding/json"
	"testing"
	"time"
)

func TestModelRuntimeProfileBindingValidate(t *testing.T) {
	binding := validModelRuntimeProfileBinding()

	if err := binding.Validate(); err != nil {
		t.Fatalf("expected valid binding, got %v", err)
	}

	binding.Profile.Thinking.AllowedModes = []string{"manual"}
	if err := binding.Validate(); err == nil {
		t.Fatal("expected default thinking mode mismatch to fail")
	}
}

func TestTunnelRequestPreservesModelRuntimeProfileBinding(t *testing.T) {
	original := TunnelRequest{
		ModelRuntimeProfiles: []ModelRuntimeProfileBinding{
			validModelRuntimeProfileBinding(),
		},
	}

	payload, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal tunnel request: %v", err)
	}

	var decoded TunnelRequest
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("unmarshal tunnel request: %v", err)
	}
	if len(decoded.ModelRuntimeProfiles) != 1 {
		t.Fatalf(
			"expected one runtime profile, got %d",
			len(decoded.ModelRuntimeProfiles),
		)
	}

	got := decoded.ModelRuntimeProfiles[0]
	if got.CatalogRevision != "catalog-revision-1" {
		t.Fatalf("unexpected catalog revision %q", got.CatalogRevision)
	}
	if !got.MatchesModel("anthropic/claude-test") {
		t.Fatal("expected decoded profile to match model slug")
	}
}

func TestModelRuntimeProfileResolvesEffectiveSpeedAndBetaOutputLimit(
	t *testing.T,
) {
	profile := validModelRuntimeProfileBinding().Profile
	profile.Endpoints = append(profile.Endpoints, "message_batches")
	profile.RequestControls.SpeedValues = []string{"standard", "fast"}
	profile.RequestControls.EffectiveSpeed = map[string]string{
		"standard": "standard",
		"fast":     "standard",
	}
	profile.BetaHeaders = append(
		profile.BetaHeaders,
		"output-300k-2026-03-24",
	)
	profile.BetaCapabilities = []ModelRuntimeBetaCapability{{
		Header:          "output-300k-2026-03-24",
		Endpoints:       []string{"message_batches"},
		MaxOutputTokens: 300000,
	}}

	if got := profile.EffectiveSpeed("fast"); got != "standard" {
		t.Fatalf("expected standard effective speed, got %q", got)
	}
	if got := profile.MaxOutputTokensFor(
		"message_batches",
		[]string{"output-300k-2026-03-24"},
	); got != 300000 {
		t.Fatalf("expected 300k beta output limit, got %d", got)
	}
	if got := profile.MaxOutputTokensFor(
		"messages",
		[]string{"output-300k-2026-03-24"},
	); got != profile.Limits.MaxOutputTokens {
		t.Fatalf("messages endpoint must retain base limit, got %d", got)
	}
}

func TestGeminiRuntimeProfileAllowsNativeAndCompatibleEndpoints(t *testing.T) {
	binding := validModelRuntimeProfileBinding()
	binding.Provider = "google-gemini"
	binding.CanonicalID = "google-gemini/gemini-3.1-flash-lite"
	binding.Slug = "gemini-3.1-flash-lite"
	binding.Profile.Endpoints = []string{
		"models",
		"chat_completions",
		"embeddings",
		"image_generations",
		"audio_speech",
		"videos",
		"interactions",
		"live",
	}
	binding.Profile.Modalities = ModelRuntimeModalities{
		Input:  []string{"text", "image", "pdf", "audio", "video"},
		Output: []string{"text", "image", "audio", "video", "embedding"},
	}
	binding.Profile.Thinking = ModelRuntimeThinkingProfile{
		DefaultMode:  "manual",
		AllowedModes: []string{"manual"},
		EffortValues: []string{"none", "minimal", "low", "medium", "high"},
	}
	binding.Profile.RequestControls.ServiceTierValues = []string{
		"standard",
		"batch",
		"flex",
		"priority",
	}
	binding.Profile.Limits.ContextWindow = 0
	binding.Profile.Limits.MaxOutputTokens = 0

	if err := binding.Validate(); err != nil {
		t.Fatalf("expected valid Gemini runtime binding, got %v", err)
	}
	if !binding.MatchesModel("google-gemini/gemini-3.1-flash-lite") {
		t.Fatal("expected provider-prefixed Gemini model to match")
	}
}

func validModelRuntimeProfileBinding() ModelRuntimeProfileBinding {
	return ModelRuntimeProfileBinding{
		Provider:        "anthropic",
		CatalogRevision: "catalog-revision-1",
		CanonicalID:     "claude-test-20260729",
		Slug:            "claude-test",
		ProfileVersion:  1,
		EffectiveFrom:   time.Date(2026, 7, 29, 0, 0, 0, 0, time.UTC),
		Profile: ModelRuntimeProfile{
			Version:          1,
			Endpoints:        []string{"messages", "count_tokens"},
			AssistantPrefill: false,
			Thinking: ModelRuntimeThinkingProfile{
				DefaultMode:  "adaptive",
				AllowedModes: []string{"adaptive", "disabled"},
				EffortValues: []string{"low", "medium", "high"},
			},
			Sampling: ModelRuntimeSamplingProfile{
				Parameters:             []string{"temperature", "top_p", "top_k"},
				DisallowedWithThinking: []string{"temperature"},
			},
			Tools: ModelRuntimeToolsProfile{
				Custom:           true,
				Server:           true,
				Parallel:         true,
				StructuredOutput: true,
			},
			Modalities: ModelRuntimeModalities{
				Input:  []string{"text", "image", "pdf"},
				Output: []string{"text"},
			},
			RequestControls: ModelRuntimeRequestControls{
				ServiceTierValues:  []string{"auto", "standard_only"},
				InferenceGeoValues: []string{"global", "us"},
				SpeedValues:        []string{"standard"},
			},
			BetaHeaders: []string{"example-beta-2026-07-29"},
			Limits: ModelRuntimeLimits{
				ContextWindow:   200000,
				MaxOutputTokens: 64000,
				MessagesBytes:   32 << 20,
				BatchBytes:      256 << 20,
				FileBytes:       500 << 20,
			},
		},
	}
}
