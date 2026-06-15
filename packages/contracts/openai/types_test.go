package openai

import (
	"encoding/json"
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
