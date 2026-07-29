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
