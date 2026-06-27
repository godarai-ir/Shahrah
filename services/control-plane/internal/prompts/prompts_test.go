package prompts

import (
	"errors"
	"testing"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
)

func TestParseFQN(t *testing.T) {
	got, err := ParseFQN("chat_prompt:tenant-a/support_reply:3")
	if err != nil {
		t.Fatalf("ParseFQN returned error: %v", err)
	}

	if got.Type != TypeChatPrompt || got.TenantSlug != "tenant-a" || got.PromptSlug != "support_reply" || got.Version != 3 {
		t.Fatalf("unexpected FQN parse result: %+v", got)
	}
}

func TestRenderSubstitutesVariablesAndAppendsMessages(t *testing.T) {
	version := platform.PromptVersion{
		ID: "pv_1", PromptID: "p_1", FQN: "chat_prompt:acme/support_reply:1", Model: "gpt-4o-mini",
		Messages: []contractsopenai.ChatMessage{
			{Role: "system", Content: "Answer for {{customer_name}}."},
			{Role: "user", Content: "Issue: {{issue}}"},
		},
	}

	rendered, err := Render(version, map[string]string{"customer_name": "Ali", "issue": "late invoice"}, []contractsopenai.ChatMessage{
		{Role: "user", Content: "Use a short tone."},
	}, "")
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	if rendered.Model != "gpt-4o-mini" {
		t.Fatalf("expected prompt model, got %q", rendered.Model)
	}

	if len(rendered.Messages) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(rendered.Messages))
	}

	if rendered.Messages[0].Content != "Answer for Ali." || rendered.Messages[1].Content != "Issue: late invoice" {
		t.Fatalf("variables were not substituted: %+v", rendered.Messages)
	}

	if rendered.Messages[2].Content != "Use a short tone." {
		t.Fatalf("request messages were not appended: %+v", rendered.Messages)
	}
}

func TestRenderRequestModelOverridesPromptModel(t *testing.T) {
	rendered, err := Render(platform.PromptVersion{
		Model:    "gpt-4o-mini",
		Messages: []contractsopenai.ChatMessage{{Role: "user", Content: "hello"}},
	}, nil, nil, "gpt-4.1-mini")
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	if rendered.Model != "gpt-4.1-mini" {
		t.Fatalf("expected request model override, got %q", rendered.Model)
	}
}

func TestRenderMissingVariables(t *testing.T) {
	_, err := Render(platform.PromptVersion{
		Messages: []contractsopenai.ChatMessage{{Role: "user", Content: "Hello {{name}}"}},
	}, nil, nil, "")
	if !errors.Is(err, ErrMissingVariable) {
		t.Fatalf("expected ErrMissingVariable, got %v", err)
	}
}

func TestValidateManifestRejectsInvalidCacheAndStreamingOutputGuardrails(t *testing.T) {
	stream := true
	threshold := 0.7

	err := ValidateManifest(platform.PromptManifest{
		Type:     TypeChatPrompt,
		Name:     "support_reply",
		Messages: []contractsopenai.ChatMessage{{Role: "user", Content: "hello"}},
		ModelSettings: &platform.PromptModelSettings{
			Stream: &stream,
		},
		Cache: &platform.PromptCacheConfig{
			Enabled:             true,
			Type:                "smart",
			SimilarityThreshold: &threshold,
		},
	})
	if !errors.Is(err, ErrInvalidManifest) {
		t.Fatalf("expected invalid manifest error for threshold, got %v", err)
	}

	validThreshold := 0.9

	err = ValidateManifest(platform.PromptManifest{
		Type:     TypeChatPrompt,
		Name:     "support_reply",
		Messages: []contractsopenai.ChatMessage{{Role: "user", Content: "hello"}},
		ModelSettings: &platform.PromptModelSettings{
			Stream: &stream,
		},
		Cache: &platform.PromptCacheConfig{
			Enabled:             true,
			Type:                "smart",
			SimilarityThreshold: &validThreshold,
		},
		Guardrails: &platform.PromptGuardrailConfig{
			Output: []string{"guardrails/secrets-detection"},
		},
	})
	if !errors.Is(err, ErrInvalidManifest) {
		t.Fatalf("expected invalid manifest error for streaming output guardrails, got %v", err)
	}
}

func TestRenderFallsBackToManifestMessages(t *testing.T) {
	rendered, err := Render(platform.PromptVersion{
		Manifest: platform.PromptManifest{
			Model: "gpt-4.1-mini",
			Messages: []contractsopenai.ChatMessage{
				{Role: "system", Content: "Hello {{name}}"},
			},
		},
	}, map[string]string{"name": "Sara"}, nil, "")
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	if len(rendered.Messages) != 1 || rendered.Messages[0].Content != "Hello Sara" {
		t.Fatalf("expected manifest messages fallback, got %+v", rendered.Messages)
	}

	if rendered.Model != "gpt-4.1-mini" {
		t.Fatalf("expected manifest model fallback, got %q", rendered.Model)
	}
}
