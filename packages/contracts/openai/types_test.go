package openai

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestChatMessageContentUnmarshalString(t *testing.T) {
	var req ChatCompletionRequest
	err := json.Unmarshal([]byte(`{"model":"gpt-4o-mini","messages":[{"role":"user","content":"hello"}]}`), &req)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got := ChatContentText(req.Messages[0].Content); got != "hello" {
		t.Fatalf("ChatContentText = %q, want hello", got)
	}
	if err := ValidateChatMessageContent(req.Messages[0].Content); err != nil {
		t.Fatalf("ValidateChatMessageContent: %v", err)
	}
}

func TestChatMessageContentUnmarshalMultimodalParts(t *testing.T) {
	var req ChatCompletionRequest
	err := json.Unmarshal([]byte(`{
		"model":"gpt-4o-mini",
		"messages":[{
			"role":"user",
			"content":[
				{"type":"text","text":"Describe this image."},
				{"type":"image_url","image_url":{"url":"data:image/png;base64,AAAA"}}
			]
		}]
	}`), &req)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got := ChatContentText(req.Messages[0].Content); got != "Describe this image." {
		t.Fatalf("ChatContentText = %q", got)
	}
	if got := ChatContentImageCount(req.Messages[0].Content); got != 1 {
		t.Fatalf("ChatContentImageCount = %d, want 1", got)
	}
	if err := ValidateChatMessageContent(req.Messages[0].Content); err != nil {
		t.Fatalf("ValidateChatMessageContent: %v", err)
	}
}

func TestChatMessageContentUnmarshalFilePart(t *testing.T) {
	var req ChatCompletionRequest
	err := json.Unmarshal([]byte(`{
		"model":"gpt-4o-mini",
		"messages":[{
			"role":"user",
			"content":[
				{"type":"text","text":"Summarize this PDF."},
				{"type":"file","file":{"filename":"guide.pdf","file_data":"data:application/pdf;base64,JVBERi0x"}}
			]
		}]
	}`), &req)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got := ChatContentText(req.Messages[0].Content); got != "Summarize this PDF." {
		t.Fatalf("ChatContentText = %q", got)
	}
	if got := ChatContentFileCount(req.Messages[0].Content); got != 1 {
		t.Fatalf("ChatContentFileCount = %d, want 1", got)
	}
	if err := ValidateChatMessageContent(req.Messages[0].Content); err != nil {
		t.Fatalf("ValidateChatMessageContent: %v", err)
	}
}

func TestChatMessageContentRejectsFilePartWithoutDataOrID(t *testing.T) {
	content := []interface{}{
		map[string]interface{}{"type": "file", "file": map[string]interface{}{"filename": "guide.pdf"}},
	}
	err := ValidateChatMessageContent(content)
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestRedactChatContentRedactsDataURL(t *testing.T) {
	redacted := RedactChatContent([]interface{}{
		map[string]interface{}{"type": "image_url", "image_url": map[string]interface{}{"url": "data:image/png;base64,AAAA"}},
	})
	raw, err := json.Marshal(redacted)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if got := string(raw); got != `[{"image_url":{"url":"data:image/png;base64,[redacted]"},"type":"image_url"}]` {
		t.Fatalf("unexpected redacted content: %s", got)
	}
}

func TestRedactChatContentRedactsFileDataURL(t *testing.T) {
	redacted := RedactChatContent([]interface{}{
		map[string]interface{}{"type": "file", "file": map[string]interface{}{"filename": "guide.pdf", "file_data": "data:application/pdf;base64,JVBERi0x"}},
	})
	raw, err := json.Marshal(redacted)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if got := string(raw); got != `[{"file":{"file_data":"data:application/pdf;base64,[redacted]","filename":"guide.pdf"},"type":"file"}]` {
		t.Fatalf("unexpected redacted content: %s", got)
	}
}

func TestParseFileDataURL(t *testing.T) {
	mediaType, data, err := ParseFileDataURL("data:application/pdf;base64,JVBERi0x")
	if err != nil {
		t.Fatalf("ParseFileDataURL: %v", err)
	}
	if mediaType != "application/pdf" || data != "JVBERi0x" {
		t.Fatalf("unexpected parsed file data URL: mediaType=%q data=%q", mediaType, data)
	}
}

func TestChatCompletionToolsRoundTrip(t *testing.T) {
	payload := []byte(`{
		"model":"gpt-4o-mini",
		"parallel_tool_calls":false,
		"messages":[
			{"role":"user","content":"What is order A-1024?"},
			{"role":"assistant","tool_calls":[{"id":"call_123","type":"function","function":{"name":"get_order_status","arguments":"{\"order_id\":\"A-1024\"}"}}]},
			{"role":"tool","tool_call_id":"call_123","content":"{\"status\":\"shipped\"}"}
		],
		"tools":[{
			"type":"function",
			"function":{
				"name":"get_order_status",
				"description":"Get order status.",
				"strict":true,
				"parameters":{"type":"object","properties":{"order_id":{"type":"string"}},"required":["order_id"],"additionalProperties":false}
			}
		}],
		"tool_choice":{"type":"function","function":{"name":"get_order_status"}}
	}`)

	var req ChatCompletionRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if req.ParallelToolCalls == nil || *req.ParallelToolCalls {
		t.Fatalf("parallel_tool_calls was not preserved: %#v", req.ParallelToolCalls)
	}
	if len(req.Tools) != 1 || req.Tools[0].Function == nil || req.Tools[0].Function.Strict == nil || !*req.Tools[0].Function.Strict {
		t.Fatalf("tool definition was not preserved: %#v", req.Tools)
	}
	if len(req.Messages) != 3 || len(req.Messages[1].ToolCalls) != 1 {
		t.Fatalf("assistant tool calls were not preserved: %#v", req.Messages)
	}
	if got := req.Messages[2].ToolCallID; got != "call_123" {
		t.Fatalf("tool_call_id = %q, want call_123", got)
	}

	raw, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	for _, want := range []string{`"tools"`, `"tool_choice"`, `"tool_calls"`, `"tool_call_id":"call_123"`, `"parallel_tool_calls":false`, `"strict":true`} {
		if !strings.Contains(string(raw), want) {
			t.Fatalf("round-trip payload missing %s: %s", want, raw)
		}
	}
}

func TestChatCompletionResponseToolCallsRoundTrip(t *testing.T) {
	payload := []byte(`{
		"id":"chatcmpl-123",
		"object":"chat.completion",
		"created":1234567890,
		"model":"gpt-4o-mini",
		"choices":[{
			"index":0,
			"message":{"role":"assistant","tool_calls":[{"id":"call_123","type":"function","function":{"name":"get_order_status","arguments":"{\"order_id\":\"A-1024\"}"}}]},
			"finish_reason":"tool_calls"
		}],
		"usage":{"prompt_tokens":20,"completion_tokens":8,"total_tokens":28}
	}`)

	var resp ChatCompletionResponse
	if err := json.Unmarshal(payload, &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(resp.Choices) != 1 || resp.Choices[0].Message == nil || len(resp.Choices[0].Message.ToolCalls) != 1 {
		t.Fatalf("tool calls were not preserved: %#v", resp.Choices)
	}
	if got := resp.Choices[0].FinishReason; got != "tool_calls" {
		t.Fatalf("finish_reason = %q, want tool_calls", got)
	}
	raw, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(raw), `"tool_calls"`) || !strings.Contains(string(raw), `"finish_reason":"tool_calls"`) {
		t.Fatalf("round-trip response missing tool call fields: %s", raw)
	}
}

func TestChatCompletionResponseRefusalRoundTrip(t *testing.T) {
	payload := []byte(`{
		"id":"chatcmpl-123",
		"object":"chat.completion",
		"created":1234567890,
		"model":"gpt-4o-mini",
		"choices":[{
			"index":0,
			"message":{"role":"assistant","refusal":"I cannot help with that request."},
			"finish_reason":"stop"
		}],
		"usage":{"prompt_tokens":20,"completion_tokens":8,"total_tokens":28}
	}`)

	var resp ChatCompletionResponse
	if err := json.Unmarshal(payload, &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got := resp.Choices[0].Message.Refusal; got != "I cannot help with that request." {
		t.Fatalf("refusal = %q", got)
	}
	raw, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(raw), `"refusal":"I cannot help with that request."`) {
		t.Fatalf("round-trip response missing refusal: %s", raw)
	}
}

func TestChatCompletionExtendedThinkingRoundTrip(t *testing.T) {
	payload := []byte(`{
		"model":"gpt-5",
		"reasoning_effort":"high",
		"thinking":{"type":"enabled","budget_tokens":3200,"display":"omitted"},
		"messages":[
			{"role":"user","content":"Solve this."},
			{
				"role":"assistant",
				"content":"Final answer",
				"reasoning_content":"private reasoning",
				"thinking_blocks":[
					{"type":"thinking","thinking":"private reasoning","signature":"sig_123"},
					{"type":"redacted_thinking","data":"opaque_blob"}
				]
			}
		]
	}`)

	var req ChatCompletionRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if req.ReasoningEffort != "high" {
		t.Fatalf("reasoning_effort = %q", req.ReasoningEffort)
	}
	if req.Thinking == nil || req.Thinking.BudgetTokens != 3200 || req.Thinking.Display != "omitted" {
		t.Fatalf("thinking config was not preserved: %#v", req.Thinking)
	}
	assistant := req.Messages[1]
	if assistant.ReasoningContent != "private reasoning" || len(assistant.ThinkingBlocks) != 2 {
		t.Fatalf("thinking fields were not preserved: %#v", assistant)
	}
	if assistant.ThinkingBlocks[0]["signature"] != "sig_123" || assistant.ThinkingBlocks[1]["data"] != "opaque_blob" {
		t.Fatalf("thinking block details were not preserved: %#v", assistant.ThinkingBlocks)
	}

	raw, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	for _, want := range []string{`"reasoning_effort":"high"`, `"thinking"`, `"budget_tokens":3200`, `"reasoning_content":"private reasoning"`, `"thinking_blocks"`, `"signature":"sig_123"`, `"type":"redacted_thinking"`} {
		if !strings.Contains(string(raw), want) {
			t.Fatalf("round-trip request missing %s: %s", want, raw)
		}
	}
}

func TestChatCompletionResponseExtendedThinkingRoundTrip(t *testing.T) {
	payload := []byte(`{
		"id":"chatcmpl-thinking",
		"object":"chat.completion",
		"created":1234567890,
		"model":"gpt-5",
		"choices":[{
			"index":0,
			"message":{
				"role":"assistant",
				"content":"Final answer",
				"reasoning_content":"reasoning text",
				"thinking_blocks":[{"type":"thinking","thinking":"reasoning text","signature":"sig_456"}]
			},
			"finish_reason":"stop"
		}],
		"usage":{"prompt_tokens":20,"completion_tokens":8,"total_tokens":28}
	}`)

	var resp ChatCompletionResponse
	if err := json.Unmarshal(payload, &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	msg := resp.Choices[0].Message
	if msg == nil || msg.ReasoningContent != "reasoning text" || len(msg.ThinkingBlocks) != 1 {
		t.Fatalf("thinking response was not preserved: %#v", resp.Choices)
	}
	if msg.ThinkingBlocks[0]["signature"] != "sig_456" {
		t.Fatalf("signature was not preserved: %#v", msg.ThinkingBlocks)
	}
	raw, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(raw), `"thinking_blocks"`) || !strings.Contains(string(raw), `"reasoning_content":"reasoning text"`) {
		t.Fatalf("round-trip response missing thinking fields: %s", raw)
	}
}

func TestResponseRequestTextFormatRoundTrip(t *testing.T) {
	payload := []byte(`{
		"model":"gpt-4o-mini",
		"input":"Extract event details.",
		"text":{
			"format":{
				"type":"json_schema",
				"name":"event",
				"strict":true,
				"schema":{"type":"object","properties":{"name":{"type":"string"}},"required":["name"],"additionalProperties":false}
			}
		}
	}`)

	var req ResponseRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if req.Text == nil || req.Text.Format == nil {
		t.Fatalf("text.format was not preserved: %#v", req.Text)
	}
	raw, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	for _, want := range []string{`"text"`, `"format"`, `"type":"json_schema"`, `"strict":true`} {
		if !strings.Contains(string(raw), want) {
			t.Fatalf("round-trip response request missing %s: %s", want, raw)
		}
	}
}

func TestModerationResponseRoundTrip(t *testing.T) {
	payload := []byte(`{"id":"modr_1","model":"omni-moderation-latest","results":[{"flagged":true,"categories":{"self-harm/intent":true},"category_scores":{"self-harm/intent":0.98}}]}`)
	var resp ModerationResponse
	if err := json.Unmarshal(payload, &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.ID != "modr_1" || resp.Model != "omni-moderation-latest" {
		t.Fatalf("unexpected response metadata: %#v", resp)
	}
	if len(resp.Results) != 1 || !resp.Results[0].Flagged {
		t.Fatalf("unexpected results: %#v", resp.Results)
	}
	if !resp.Results[0].Categories["self-harm/intent"] {
		t.Fatalf("category not preserved: %#v", resp.Results[0].Categories)
	}
	if resp.Results[0].CategoryScores["self-harm/intent"] != 0.98 {
		t.Fatalf("score not preserved: %#v", resp.Results[0].CategoryScores)
	}
}
