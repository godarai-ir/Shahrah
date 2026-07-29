package server

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
)

func TestConvertToAnthropicRequestMapsToolsStructuredOutputAndParameters(t *testing.T) {
	parallel := false
	topK := 20
	maxTokens := 8192
	request := contractsopenai.ChatCompletionRequest{
		Model:             "anthropic/claude-opus-4-7",
		MaxTokens:         &maxTokens,
		TopK:              &topK,
		Stop:              []interface{}{"END", "STOP"},
		ReasoningEffort:   "high",
		ParallelToolCalls: &parallel,
		ToolChoice: map[string]interface{}{
			"type":     "function",
			"function": map[string]interface{}{"name": "weather"},
		},
		Tools: []contractsopenai.ToolDefinition{{
			Type: "function",
			Function: &contractsopenai.FunctionDef{
				Name: "weather",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"city": map[string]interface{}{"type": "string"},
					},
				},
			},
		}},
		ResponseFormat: map[string]interface{}{
			"type": "json_schema",
			"json_schema": map[string]interface{}{
				"name": "weather",
				"schema": map[string]interface{}{
					"type":     "object",
					"required": []interface{}{"temperature"},
				},
			},
		},
		Metadata:     map[string]interface{}{"user_id": "external-user"},
		ServiceTier:  "auto",
		InferenceGeo: "us",
		Messages: []contractsopenai.ChatMessage{{
			Role:    "user",
			Content: "weather in Tehran",
		}},
	}

	// Adaptive thinking and sampling cannot be combined.
	if _, err := convertToAnthropicRequest(request); err == nil {
		t.Fatal("expected adaptive thinking sampling conflict")
	}
	request.TopK = nil
	converted, err := convertToAnthropicRequest(request)
	if err != nil {
		t.Fatalf("convertToAnthropicRequest: %v", err)
	}
	if converted.Thinking == nil || converted.Thinking.Type != "adaptive" {
		t.Fatalf("unexpected thinking: %#v", converted.Thinking)
	}
	if converted.ToolChoice == nil || converted.ToolChoice.Type != "tool" ||
		converted.ToolChoice.Name != "weather" || !converted.ToolChoice.DisableParallelToolUse {
		t.Fatalf("unexpected tool choice: %#v", converted.ToolChoice)
	}
	if len(converted.Tools) != 1 || converted.Tools[0].Name != "weather" {
		t.Fatalf("unexpected tools: %#v", converted.Tools)
	}
	if converted.Output == nil || converted.Output.Effort != "high" {
		t.Fatalf("unexpected output config: %#v", converted.Output)
	}
	format, ok := converted.Output.Format.(map[string]interface{})
	if !ok || format["type"] != "json_schema" {
		t.Fatalf("unexpected output format: %#v", converted.Output.Format)
	}
	schema := format["schema"].(map[string]interface{})
	if schema["type"] != "object" {
		t.Fatalf("schema semantics were not preserved: %#v", schema)
	}
	if converted.Metadata == nil || converted.Metadata.UserID != "external-user" {
		t.Fatalf("unexpected metadata: %#v", converted.Metadata)
	}
	if converted.InferenceGeo != "us" || len(converted.Stop) != 2 {
		t.Fatalf("parameters were not preserved: %#v", converted)
	}
}

func TestConvertToAnthropicRequestMapsToolMessages(t *testing.T) {
	request := contractsopenai.ChatCompletionRequest{
		Model: "claude-sonnet-4-5",
		Messages: []contractsopenai.ChatMessage{
			{
				Role: "assistant",
				ToolCalls: []contractsopenai.ToolCall{{
					ID:   "toolu_1",
					Type: "function",
					Function: &contractsopenai.FunctionCall{
						Name:      "weather",
						Arguments: `{"city":"Tehran"}`,
					},
				}},
			},
			{Role: "tool", ToolCallID: "toolu_1", Content: `{"temperature":31}`},
		},
	}
	converted, err := convertToAnthropicRequest(request)
	if err != nil {
		t.Fatalf("convertToAnthropicRequest: %v", err)
	}
	if got := converted.Messages[0].Content[0]; got.Type != "tool_use" || got.ID != "toolu_1" {
		t.Fatalf("unexpected tool_use: %#v", got)
	}
	if got := converted.Messages[1]; got.Role != "user" ||
		got.Content[0].Type != "tool_result" || got.Content[0].ToolUseID != "toolu_1" {
		t.Fatalf("unexpected tool_result: %#v", got)
	}
}

func TestConvertAnthropicResponseMapsToolUseAndFullUsage(t *testing.T) {
	var input interface{}
	if err := json.Unmarshal([]byte(`{"city":"Tehran"}`), &input); err != nil {
		t.Fatal(err)
	}
	response := anthropicResponse{
		ID:         "msg_1",
		StopReason: "tool_use",
		Content: []anthropicContentBlock{{
			Type: "tool_use", ID: "toolu_1", Name: "weather", Input: input,
		}},
		Usage: anthropicUsage{
			InputTokens:              100,
			OutputTokens:             20,
			CacheCreationInputTokens: 30,
			CacheReadInputTokens:     40,
			ServerToolUse:            map[string]int{"web_search_requests": 2},
		},
	}
	response.Usage.CacheCreation.Ephemeral5mInputTokens = 10
	response.Usage.CacheCreation.Ephemeral1hInputTokens = 20

	converted := convertToOpenAIResponse(response, "claude-sonnet-4-6")
	if converted.Choices[0].FinishReason != "tool_calls" {
		t.Fatalf("unexpected finish reason: %s", converted.Choices[0].FinishReason)
	}
	calls := converted.Choices[0].Message.ToolCalls
	if len(calls) != 1 || calls[0].Function.Name != "weather" ||
		calls[0].Function.Arguments != `{"city":"Tehran"}` {
		t.Fatalf("unexpected tool calls: %#v", calls)
	}
	if converted.Usage.CacheCreation5mInputTokens != 10 ||
		converted.Usage.CacheCreation1hInputTokens != 20 ||
		converted.Usage.CacheReadInputTokens != 40 ||
		converted.Usage.ServerToolUse["web_search_requests"] != 2 {
		t.Fatalf("usage was not preserved: %#v", converted.Usage)
	}
}

func TestOpus47RejectsManualThinking(t *testing.T) {
	request := contractsopenai.ChatCompletionRequest{
		Model:    "claude-opus-4-7",
		Thinking: &contractsopenai.ThinkingConfig{Type: "enabled", BudgetTokens: 2048},
		Messages: []contractsopenai.ChatMessage{{Role: "user", Content: "test"}},
	}
	if _, err := convertToAnthropicRequest(request); err == nil {
		t.Fatal("expected manual thinking to be rejected for Opus 4.7")
	}
}

func TestAnthropicTypedOperationBuildsFixedPathAndFiltersBeta(t *testing.T) {
	var gotMethod, gotPath, gotQuery, gotAPIKey, gotAuthorization, gotBeta string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		gotMethod, gotPath, gotQuery = request.Method, request.URL.Path, request.URL.RawQuery
		gotAPIKey = request.Header.Get("x-api-key")
		gotAuthorization = request.Header.Get("Authorization")
		gotBeta = request.Header.Get("anthropic-beta")
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"data":[],"has_more":false}`)
	}))
	defer upstream.Close()

	adapter := &anthropicAdapter{}
	response, err := adapter.Dispatch(t.Context(), upstream.Client(), platform.TunnelRequest{
		Operation: platform.ProviderOperationAnthropicFilesList,
		OperationInput: platform.ProviderOperationInput{
			AfterID: "file_1",
			Limit:   12,
		},
		Headers: map[string]string{
			"anthropic-beta": "files-api-2025-04-14,unapproved-beta",
			"authorization":  "Bearer virtual-key-must-not-leak",
		},
		Deployment: platform.Deployment{
			Provider: "anthropic",
			APIBase:  upstream.URL,
			APIKey:   "sk-ant-provider",
			ProviderHeaders: map[string]string{
				"Anthropic-Beta-Allowlist": "files-api-2025-04-14",
			},
		},
	})
	if err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	defer response.Body.Close()
	if gotMethod != http.MethodGet || gotPath != "/v1/files" ||
		!strings.Contains(gotQuery, "after_id=file_1") || !strings.Contains(gotQuery, "limit=12") {
		t.Fatalf("unexpected upstream target: %s %s?%s", gotMethod, gotPath, gotQuery)
	}
	if gotAPIKey != "sk-ant-provider" || gotAuthorization != "" {
		t.Fatalf("provider auth isolation failed: x-api-key=%q authorization=%q", gotAPIKey, gotAuthorization)
	}
	if gotBeta != "files-api-2025-04-14" {
		t.Fatalf("unexpected beta header: %q", gotBeta)
	}
}

func TestAnthropicTypedOperationRequiresResourceID(t *testing.T) {
	_, _, _, err := anthropicOperationTarget(platform.ProviderOperationAnthropicFilesContent, platform.ProviderOperationInput{})
	if err == nil {
		t.Fatal("expected resource_id validation error")
	}
}

func TestRawTunnelStreamsAnthropicFileUpload(t *testing.T) {
	payloadBytes := bytes.Repeat([]byte("x"), 2*1024*1024)
	var received int
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/v1/files" || request.Method != http.MethodPost {
			t.Fatalf("unexpected target: %s %s", request.Method, request.URL.Path)
		}
		data, err := io.ReadAll(request.Body)
		if err != nil {
			t.Fatal(err)
		}
		received = len(data)
		_, _ = io.WriteString(w, `{"id":"file_upstream","type":"file"}`)
	}))
	defer upstream.Close()

	metadata, err := json.Marshal(platform.TunnelRequest{
		Operation: platform.ProviderOperationAnthropicFilesCreate,
		Headers:   map[string]string{"content-type": "multipart/form-data; boundary=test"},
		Deployment: platform.Deployment{
			Provider: "anthropic", APIBase: upstream.URL, APIKey: "sk-provider",
			ProviderHeaders: map[string]string{"Anthropic-Beta-Allowlist": "files-api-2025-04-14"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	var prefix [4]byte
	binary.BigEndian.PutUint32(prefix[:], uint32(len(metadata)))
	body := io.MultiReader(bytes.NewReader(prefix[:]), bytes.NewReader(metadata), bytes.NewReader(payloadBytes))
	request := httptest.NewRequest(http.MethodPost, "/dispatch/raw", body)
	request.ContentLength = int64(4 + len(metadata) + len(payloadBytes))
	recorder := httptest.NewRecorder()
	NewHandler(nil).ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK || received != len(payloadBytes) {
		t.Fatalf("raw upload failed: status=%d received=%d body=%s", recorder.Code, received, recorder.Body.String())
	}
}
