package storage

import (
	"encoding/json"
	"testing"

	"github.com/Delpak/shahrah/services/control-plane/internal/observability"
)

func TestPrepareRequestLogPayloadKeepsMetadataModeLazy(t *testing.T) {
	body := []byte(`{"prompt":"secret","model":"gpt-test"}`)
	payload := map[string]interface{}{
		observability.LazyJSONPayloadField: json.RawMessage(body),
		"model":                            "override",
	}

	content, metadata, truncated := prepareRequestLogPayload(RequestLoggingModeMetadataOnly, payload)
	if content != nil || truncated {
		t.Fatalf("metadata-only mode should not materialize content: %#v", content)
	}

	if metadata["bytes"] != len(body) {
		t.Fatalf("expected original payload byte count, got %#v", metadata)
	}
}

func TestPrepareRequestLogPayloadMaterializesLazyFullContent(t *testing.T) {
	payload := map[string]interface{}{
		observability.LazyJSONPayloadField: json.RawMessage(`{"prompt":"secret","model":"gpt-test"}`),
		"model":                            "override",
	}

	content, _, _ := prepareRequestLogPayload(RequestLoggingModeFullContent, payload)
	if content["prompt"] != "secret" || content["model"] != "override" {
		t.Fatalf("unexpected materialized content: %#v", content)
	}

	if _, exists := content[observability.LazyJSONPayloadField]; exists {
		t.Fatal("internal lazy payload marker leaked into persisted content")
	}
}

func TestPrepareRequestLogPayloadHandlesJSONNull(t *testing.T) {
	payload := map[string]interface{}{
		observability.LazyJSONPayloadField: json.RawMessage("null"),
		"model":                            "override",
	}

	content, _, _ := prepareRequestLogPayload(RequestLoggingModeFullContent, payload)
	if content["model"] != "override" {
		t.Fatalf("expected override after null payload, got %#v", content)
	}
}
