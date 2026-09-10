package server

import (
	"encoding/json"
	"strings"
	"testing"
)

func rawJSONBody(text string) map[string]json.RawMessage {
	var body map[string]json.RawMessage
	if err := json.Unmarshal([]byte(text), &body); err != nil {
		panic(err)
	}

	return body
}

// TestParseTokenBazaarImageParameterSnapshotsOnlyKeepsInScopeImageModels
// confirms the /v1/models parser only extracts the 4 in-scope image models,
// requires endpoint_types to include "images", and ignores everything else
// (chat models, the embeddings model, out-of-catalog ids) — this delivery
// never invents a snapshot for a model TokenBazaar's own response didn't
// describe as an image model.
func TestParseTokenBazaarImageParameterSnapshotsOnlyKeepsInScopeImageModels(t *testing.T) {
	body := []byte(`{
		"data": [
			{"id":"gpt-image-2","supported_endpoint_types":["images"],"supported_parameters":{"size":["1024x1024","1536x1024"],"quality":["standard","high"]}},
			{"id":"gemini-3-pro-image","supported_endpoint_types":["images"],"supported_parameters":{"size":["1024x1024"],"quality":[]}},
			{"id":"claude-sonnet-5","supported_endpoint_types":["chat/completions"]},
			{"id":"text-embedding-3-large","supported_endpoint_types":["embeddings"]},
			{"id":"gemini-3.1-flash-lite-image","supported_endpoint_types":["chat/completions"]}
		]
	}`)

	snapshots, err := parseTokenBazaarImageParameterSnapshots("https://zx1.deepwl.net/v1/models", body)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if len(snapshots) != 2 {
		t.Fatalf("expected exactly 2 snapshots (gpt-image-2, gemini-3-pro-image), got %d: %v", len(snapshots), snapshots)
	}

	gptSnapshot, ok := snapshots["gpt-image-2"]
	if !ok {
		t.Fatal("missing gpt-image-2 snapshot")
	}

	if _, ok := gptSnapshot.SupportedSizes["1024x1024"]; !ok {
		t.Fatal("gpt-image-2 snapshot missing size 1024x1024")
	}

	if _, ok := gptSnapshot.SupportedQuality["high"]; !ok {
		t.Fatal("gpt-image-2 snapshot missing quality high")
	}

	if gptSnapshot.SourceURL != "https://zx1.deepwl.net/v1/models" {
		t.Fatalf("provenance SourceURL = %q", gptSnapshot.SourceURL)
	}

	if _, ok := snapshots["claude-sonnet-5"]; ok {
		t.Fatal("a chat-only model must never get an image parameter snapshot")
	}

	if _, ok := snapshots["gemini-3.1-flash-lite-image"]; ok {
		t.Fatal("an in-scope image model reported without endpoint_types containing \"images\" must not get a snapshot")
	}

	if _, ok := snapshots["text-embedding-3-large"]; ok {
		t.Fatal("the embeddings model must never get an image parameter snapshot")
	}
}

func TestParseTokenBazaarImageParameterSnapshotsRejectsInvalidJSON(t *testing.T) {
	_, err := parseTokenBazaarImageParameterSnapshots("https://zx1.deepwl.net/v1/models", []byte("not json"))
	if err == nil || !strings.Contains(err.Error(), "tokenbazaar_models_response_invalid") {
		t.Fatalf("expected tokenbazaar_models_response_invalid, got %v", err)
	}
}

// TestRefreshTokenBazaarImageParameterCacheInstallsIntoValidation confirms
// the discovery-job entry point (refreshTokenBazaarImageParameterCache)
// actually installs snapshots that validateTokenBazaarImageParameters then
// honors — this is the intended live-discovery integration point described
// in plans/tokenbazaar-production-readiness.md conflict #12, exercised here
// end-to-end with a fixture response instead of a live account.
func TestRefreshTokenBazaarImageParameterCacheInstallsIntoValidation(t *testing.T) {
	clearTokenBazaarImageParameterCache()
	t.Cleanup(clearTokenBazaarImageParameterCache)

	body := []byte(`{"data":[{"id":"gpt-image-2","supported_endpoint_types":["images"],"supported_parameters":{"size":["1024x1024"],"quality":["standard"]}}]}`)

	if err := refreshTokenBazaarImageParameterCache("https://zx1.deepwl.net/v1/models", body); err != nil {
		t.Fatalf("refresh: %v", err)
	}

	if err := validateTokenBazaarImageParameters("gpt-image-2", rawJSONBody(`{"size":"1024x1024"}`)); err != nil {
		t.Fatalf("expected confirmed size to be accepted, got %v", err)
	}

	if err := validateTokenBazaarImageParameters("gpt-image-2", rawJSONBody(`{"quality":"hd"}`)); err == nil {
		t.Fatal("expected unconfirmed quality to be rejected")
	}
}
