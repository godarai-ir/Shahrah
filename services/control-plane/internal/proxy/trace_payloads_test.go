package proxy

import (
	"encoding/json"
	"strings"
	"testing"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
)

func TestBuildTraceChatRequestPayloadRedactsImageDataURL(t *testing.T) {
	req := contractsopenai.ChatCompletionRequest{
		Model: "gpt-4o-mini",
		Messages: []contractsopenai.ChatMessage{{
			Role: "user",
			Content: []interface{}{
				map[string]interface{}{"type": "text", "text": "Describe this image."},
				map[string]interface{}{"type": "image_url", "image_url": map[string]interface{}{"url": "data:image/png;base64,AAAA"}},
			},
		}},
	}

	payload := buildTraceChatRequestPayload(req)
	messages := payload["messages"].([]map[string]interface{})
	content := messages[0]["content"].([]map[string]interface{})

	imageURL := content[1]["image_url"].(map[string]interface{})
	if got := imageURL["url"]; got != "data:image/png;base64,[redacted]" {
		t.Fatalf("image url = %#v", got)
	}
}

func TestDeepSeekTraceAndRequestLogsExcludeSensitiveContent(t *testing.T) {
	const (
		promptSecret    = "prompt-secret-4a98"
		reasoningSecret = "reasoning-secret-7b21"
		toolSecret      = "tool-arguments-secret-8c32"
		userIDSecret    = "person@example.com"
	)

	req := contractsopenai.ChatCompletionRequest{
		Model: "deepseek-v4-flash",
		Messages: []contractsopenai.ChatMessage{
			{Role: "user", Content: promptSecret},
			{
				Role:             "assistant",
				ReasoningContent: reasoningSecret,
				ToolCalls: []contractsopenai.ToolCall{{
					Type: "function",
					Function: &contractsopenai.FunctionCall{
						Name:      "lookup",
						Arguments: toolSecret,
					},
				}},
			},
		},
		User: userIDSecret,
	}

	payloads := []map[string]interface{}{
		buildProviderTraceChatRequestPayload("deepseek", req),
		buildProviderChatRequestLogPayload("deepseek", req, []byte(`{"user_id":"`+userIDSecret+`","messages":[{"content":"`+promptSecret+`"}]}`)),
		buildProviderStreamLogPayload("deepseek", "completed", map[string]interface{}{"raw_sse": promptSecret}),
		buildProviderTraceErrorPayload("deepseek", 400, []byte(promptSecret)),
		buildProviderHTTPErrorLogPayloadForProvider("deepseek", 400, []byte(promptSecret), false),
	}

	assertSensitiveValuesAbsent(t, payloads, promptSecret, reasoningSecret, toolSecret, userIDSecret)

	if got := payloads[0]["model"]; got != req.Model {
		t.Fatalf("model = %#v", got)
	}

	if got := payloads[0]["tool_call_count"]; got != 1 {
		t.Fatalf("tool_call_count = %#v", got)
	}
}

func TestDeepSeekResponseLogsExcludeContentAndRetainUsage(t *testing.T) {
	const (
		responseSecret  = "response-secret-1d43"
		reasoningSecret = "reasoning-secret-2e54"
		toolSecret      = "tool-arguments-secret-3f65"
	)

	completion := contractsopenai.ChatCompletionResponse{
		Model: "deepseek-v4-pro",
		Choices: []contractsopenai.Choice{{
			Index: 0,
			Message: &contractsopenai.ChatMessage{
				Role:             "assistant",
				Content:          responseSecret,
				ReasoningContent: reasoningSecret,
				ToolCalls: []contractsopenai.ToolCall{{
					Type:     "function",
					Function: &contractsopenai.FunctionCall{Name: "lookup", Arguments: toolSecret},
				}},
			},
			FinishReason: "stop",
		}},
		Usage: &contractsopenai.Usage{
			PromptTokens:         12,
			CompletionTokens:     7,
			TotalTokens:          19,
			PromptCacheHitTokens: 5,
		},
	}

	payloads := []map[string]interface{}{
		buildProviderTraceChatResponsePayload("deepseek", completion),
		buildProviderChatResponseLogPayload("deepseek", nil, completion),
	}

	assertSensitiveValuesAbsent(t, payloads, responseSecret, reasoningSecret, toolSecret)

	usage, ok := payloads[0]["usage"].(map[string]interface{})
	if !ok {
		t.Fatalf("usage = %#v", payloads[0]["usage"])
	}

	if got := usage["total_tokens"]; got != 19 {
		t.Fatalf("total_tokens = %#v", got)
	}
}

func assertSensitiveValuesAbsent(t *testing.T, payloads []map[string]interface{}, secrets ...string) {
	t.Helper()

	encoded, err := json.Marshal(payloads)
	if err != nil {
		t.Fatalf("marshal payloads: %v", err)
	}

	serialized := string(encoded)
	for _, secret := range secrets {
		if strings.Contains(serialized, secret) {
			t.Fatalf("sensitive value %q leaked in %s", secret, serialized)
		}
	}
}
