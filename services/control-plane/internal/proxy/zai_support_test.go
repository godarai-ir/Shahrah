package proxy

import (
	"encoding/json"
	"testing"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
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
