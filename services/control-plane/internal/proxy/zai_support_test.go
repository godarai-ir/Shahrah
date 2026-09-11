package proxy

import (
	"encoding/json"
	"testing"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
)

func TestZAILayoutParsingValidatesOfficialContract(t *testing.T) {
	valid := map[string]interface{}{
		"model": "glm-ocr", "file": "https://example.com/document.pdf",
		"start_page_id": float64(1), "end_page_id": float64(30),
		"return_crop_images": true,
	}
	if err := validateZAILayoutParsingRequest(valid); err != nil {
		t.Fatalf("valid GLM-OCR request rejected: %v", err)
	}

	invalid := map[string]interface{}{
		"model": "glm-ocr", "file": "https://example.com/document.pdf",
		"start_page_id": float64(1), "end_page_id": float64(31),
	}
	if err := validateZAILayoutParsingRequest(invalid); err == nil {
		t.Fatal("31-page GLM-OCR request was accepted")
	}

	invalid["end_page_id"] = float64(30)

	invalid["start_page_id"] = float64(0)
	if err := validateZAILayoutParsingRequest(invalid); err == nil {
		t.Fatal("zero-based GLM-OCR page was accepted")
	}
}

func TestZAIOCRUsageRequiresProviderTokenCounts(t *testing.T) {
	usage, err := zaiOCRUsage([]byte(`{"usage":{"prompt_tokens":11,"completion_tokens":7,"total_tokens":18}}`))
	if err != nil {
		t.Fatalf("valid GLM-OCR usage rejected: %v", err)
	}

	if usage.PromptTokens != 11 || usage.CompletionTokens != 7 || usage.TotalTokens != 18 {
		t.Fatalf("unexpected GLM-OCR usage: %+v", usage)
	}

	if _, err := zaiOCRUsage([]byte(`{"id":"missing_usage"}`)); err == nil {
		t.Fatal("missing GLM-OCR usage was accepted")
	}
}

func TestZAIThinkingRequestRespectsAlwaysEnabledModels(t *testing.T) {
	toggleable := platform.ModelRuntimeProfile{
		Thinking: platform.ModelRuntimeThinkingProfile{
			DefaultMode: "adaptive", AllowedModes: []string{"adaptive", "disabled"},
		},
	}
	alwaysEnabled := platform.ModelRuntimeProfile{
		Thinking: platform.ModelRuntimeThinkingProfile{
			DefaultMode: "adaptive", AllowedModes: []string{"adaptive"}, AlwaysEnabled: true,
		},
	}

	disabled := &contractsopenai.ThinkingConfig{Type: "disabled"}
	enabled := &contractsopenai.ThinkingConfig{Type: "enabled"}

	if err := validateZAIThinkingRequest(contractsopenai.ChatCompletionRequest{Thinking: disabled}, toggleable); err != nil {
		t.Fatalf("thinking.type=disabled rejected for a toggleable Z.ai model: %v", err)
	}

	if err := validateZAIThinkingRequest(contractsopenai.ChatCompletionRequest{Thinking: enabled}, alwaysEnabled); err != nil {
		t.Fatalf("thinking.type=enabled rejected for glm-5.3-flash-style always-enabled model: %v", err)
	}

	if err := validateZAIThinkingRequest(contractsopenai.ChatCompletionRequest{Thinking: disabled}, alwaysEnabled); err == nil {
		t.Fatal("thinking.type=disabled was accepted for an always-enabled-thinking Z.ai model")
	}

	if err := validateZAIThinkingRequest(contractsopenai.ChatCompletionRequest{}, alwaysEnabled); err != nil {
		t.Fatalf("omitted thinking was rejected for an always-enabled-thinking Z.ai model: %v", err)
	}
}

// TestZAIEffortValuesRespectPerModelEnum proves glm-5.3's narrower
// low/high/max reasoning_effort enum rejects values the wider 7-value set
// used by glm-5.2 and the other zai chat models still accepts.
func TestZAIEffortValuesRespectPerModelEnum(t *testing.T) {
	glm53 := platform.ModelRuntimeProfile{
		Thinking: platform.ModelRuntimeThinkingProfile{
			DefaultMode: "adaptive", AllowedModes: []string{"adaptive"}, AlwaysEnabled: true,
			EffortValues: []string{"low", "high", "max"},
		},
	}
	glm52 := platform.ModelRuntimeProfile{
		Thinking: platform.ModelRuntimeThinkingProfile{
			DefaultMode: "adaptive", AllowedModes: []string{"adaptive", "disabled"},
			EffortValues: []string{"none", "minimal", "low", "medium", "high", "xhigh", "max"},
		},
	}

	if glm53.AllowsEffort("medium") {
		t.Fatal("glm-5.3 accepted an unsupported reasoning_effort value")
	}

	if !glm53.AllowsEffort("low") || !glm53.AllowsEffort("high") || !glm53.AllowsEffort("max") {
		t.Fatal("glm-5.3 rejected one of its own documented reasoning_effort values")
	}

	if !glm52.AllowsEffort("medium") {
		t.Fatal("glm-5.2's wider reasoning_effort enum regressed")
	}
}

func TestZAIStopGateFailsClosed(t *testing.T) {
	if count := zaiStopSequenceCount([]string{"one", "two"}); count != 2 {
		t.Fatalf("stop count = %d", count)
	}
}

func TestZAIRawChatControlsFollowUpstreamContract(t *testing.T) {
	req := contractsopenai.ChatCompletionRequest{
		Stream: true,
		Tools: []contractsopenai.ToolDefinition{{
			Type: "function",
			Function: &contractsopenai.FunctionDef{
				Name: "lookup", Description: "Lookup a record", Parameters: map[string]interface{}{"type": "object"},
			},
		}},
		AdditionalFields: map[string]json.RawMessage{
			"tool_stream": json.RawMessage(`true`),
			"request_id":  json.RawMessage(`"request-123"`),
			"user_id":     json.RawMessage(`"user-123"`),
		},
	}
	if err := validateZAIRawChatControls(req); err != nil {
		t.Fatalf("valid controls rejected: %v", err)
	}

	req.Stream = false
	if err := validateZAIRawChatControls(req); err == nil {
		t.Fatal("tool_stream without stream=true was accepted")
	}

	req.AdditionalFields["tool_stream"] = json.RawMessage(`false`)

	req.AdditionalFields["request_id"] = json.RawMessage(`"short"`)
	if err := validateZAIRawChatControls(req); err == nil {
		t.Fatal("short request_id was accepted")
	}
}

func TestZAIContentPartRejectsInlineVideoAndOversizedImage(t *testing.T) {
	inlineVideo := contractsopenai.ChatContentPart{
		Type: "video_url", VideoURL: &contractsopenai.ChatVideoURL{URL: "data:video/mp4;base64,AAAA"},
	}
	if err := validateZAIContentPart(inlineVideo); err == nil {
		t.Fatal("inline Z.ai video was accepted")
	}

	validImage := contractsopenai.ChatContentPart{
		Type: "image_url", ImageURL: &contractsopenai.ChatImageURL{URL: "data:image/png;base64,AAAA"},
	}
	if err := validateZAIContentPart(validImage); err != nil {
		t.Fatalf("valid inline Z.ai image rejected: %v", err)
	}
}
