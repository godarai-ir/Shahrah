package proxy

import (
	"context"
	"encoding/json"
	"testing"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
)

func TestMistralOCRUsageHelpers(t *testing.T) {
	request := map[string]interface{}{"pages": []interface{}{0, 2}, "document_annotation_format": map[string]interface{}{"type": "json_schema"}}
	if pages := mistralOCRPageEstimate(request); pages != 2 {
		t.Fatalf("estimated pages = %d", pages)
	}

	if !mistralOCRUsesAnnotations(request) {
		t.Fatal("annotation pricing was not selected")
	}

	if pages := mistralOCRResponsePages([]byte(`{"usage_info":{"pages_processed":3}}`)); pages != 3 {
		t.Fatalf("processed pages = %d", pages)
	}
}

func TestMistralOCRValidatesOfficialRequestContract(t *testing.T) {
	valid := map[string]interface{}{
		"model":        "mistral-ocr-4-0",
		"document":     map[string]interface{}{"type": "document_url", "document_url": "https://example.com/document.pdf"},
		"pages":        "0,2-4",
		"table_format": "html",
	}
	if err := validateMistralOCRRequest(valid); err != nil {
		t.Fatalf("valid OCR request was rejected: %v", err)
	}

	if pages := mistralOCRPageEstimate(valid); pages != 4 {
		t.Fatalf("page selection estimate = %d, want 4", pages)
	}

	invalid := map[string]interface{}{
		"model":                      "mistral-ocr-4-0",
		"document":                   map[string]interface{}{"type": "document_url", "document_url": "https://example.com/document.pdf"},
		"document_annotation_prompt": "extract invoice fields",
	}
	if err := validateMistralOCRRequest(invalid); err == nil {
		t.Fatal("annotation prompt without document_annotation_format was accepted")
	}
}

func TestMistralRuntimeProfileRejectsUnsupportedParameter(t *testing.T) {
	budget := &testBudgeter{}

	estimated, err := budget.PriceMeteredUsage(context.Background(), "mistral-ai:acct:voxtral-mini-2602", "audio_input_seconds", "0", platform.UsageCostBreakdown{})
	if err != nil {
		t.Fatal(err)
	}

	parameters := map[string]json.RawMessage{"model": json.RawMessage(`"voxtral-mini-2602"`), "temperature": json.RawMessage(`0`), "unknown": json.RawMessage(`true`)}
	if err := validateMistralParameters(estimated.ModelRuntimeProfile, "audio_transcriptions", parameters); err == nil {
		t.Fatal("unsupported parameter was accepted")
	}
}

func TestAudioTranscriptionMistralBillingUsesSeconds(t *testing.T) {
	dimension, quantity := audioTranscriptionMeteredQuantity("mistral-ai", "1.5")
	if dimension != "audio_input_seconds" || quantity != "90" {
		t.Fatalf("got %s=%s", dimension, quantity)
	}
}

func TestMistralChatRejectsConversationOnlyBuiltInTools(t *testing.T) {
	budget := &testBudgeter{}

	estimated, err := budget.PriceMeteredUsage(context.Background(), "mistral-ai:acct:mistral-small-2603", "input_tokens", "0", platform.UsageCostBreakdown{})
	if err != nil {
		t.Fatal(err)
	}

	estimated.ModelRuntimeProfile.Profile.Tools.Supported = []string{"function_calling", "web_search", "image_generation"}

	request := contractsopenai.ChatCompletionRequest{
		Model: "mistral-small-2603",
		Tools: []contractsopenai.ToolDefinition{{Type: "web_search"}},
	}
	if err := validateMistralChatRuntimeProfile(request, estimated.ModelRuntimeProfile); err == nil {
		t.Fatal("web_search was accepted on the Mistral Chat Completions endpoint")
	}

	request.Tools = []contractsopenai.ToolDefinition{{Type: "image_generation"}}
	if err := validateMistralChatRuntimeProfile(request, estimated.ModelRuntimeProfile); err == nil {
		t.Fatal("image_generation was accepted without image billing and Files API support")
	}
}

func TestMistralEmbeddingNativeParameterValidation(t *testing.T) {
	for _, body := range []string{
		`{"output_dimension":0}`,
		`{"output_dtype":"float16"}`,
	} {
		if err := validateMistralEmbeddingRequest([]byte(body)); err == nil {
			t.Fatalf("invalid embedding request was accepted: %s", body)
		}
	}

	if err := validateMistralEmbeddingRequest([]byte(`{"output_dimension":256,"output_dtype":"int8"}`)); err != nil {
		t.Fatalf("valid Codestral Embed parameters were rejected: %v", err)
	}
}

func TestMistralEmbedUsesFixedDimensions(t *testing.T) {
	if modelSupportsCustomDimensions("mistral-embed") || modelSupportsCustomDimensions("mistral-embed-2312") {
		t.Fatal("mistral-embed must retain its documented fixed 1024-dimensional output")
	}

	if !modelSupportsCustomDimensions("codestral-embed-2505") {
		t.Fatal("Codestral Embed configurable dimensions were disabled")
	}
}

func TestMistralFIMValidatesTokenBounds(t *testing.T) {
	budget := &testBudgeter{}

	estimated, err := budget.PriceMeteredUsage(context.Background(), "mistral-ai:acct:codestral-2508", "input_tokens", "0", platform.UsageCostBreakdown{})
	if err != nil {
		t.Fatal(err)
	}

	minTokens, maxTokens := 20, 10

	request := contractsopenai.CompletionRequest{
		Model:     "codestral-2508",
		Prompt:    "func main()",
		MinTokens: &minTokens,
		MaxTokens: &maxTokens,
		AdditionalFields: map[string]json.RawMessage{
			"model":      json.RawMessage(`"codestral-2508"`),
			"prompt":     json.RawMessage(`"func main()"`),
			"min_tokens": json.RawMessage(`20`),
			"max_tokens": json.RawMessage(`10`),
		},
	}
	if err := validateMistralFIMRuntimeProfile(request, estimated.ModelRuntimeProfile); err == nil {
		t.Fatal("invalid FIM token bounds were accepted")
	}
}

func TestMistralFeatureFlagsFailClosed(t *testing.T) {
	features := ProviderFeatureConfig{
		Configured:       true,
		MistralEnabled:   true,
		MistralEndpoints: map[string]bool{"ocr": false},
	}

	if mistralEndpointEnabled(features, "ocr") {
		t.Fatal("invalid endpoint feature flag must fail closed")
	}

	features.MistralEndpoints["ocr"] = true

	if !mistralEndpointEnabled(features, "ocr") {
		t.Fatal("enabled endpoint feature flag was rejected")
	}
}
