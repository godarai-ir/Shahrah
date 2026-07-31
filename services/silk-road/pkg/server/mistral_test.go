package server

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

func TestMistralNativeOperationsUseFixedOfficialPaths(t *testing.T) {
	tests := []struct {
		operation platform.ProviderOperation
		method    string
		path      string
	}{
		{platform.ProviderOperationMistralFilesContent, http.MethodGet, "/files/id%2Fwith%2Fslashes/content"},
		{platform.ProviderOperationMistralBatchCancel, http.MethodPost, "/batch/jobs/id%2Fwith%2Fslashes/cancel"},
		{platform.ProviderOperationMistralAgentsUpdate, http.MethodPatch, "/agents/id%2Fwith%2Fslashes"},
		{platform.ProviderOperationMistralConversationsRestart, http.MethodPost, "/conversations/id%2Fwith%2Fslashes/restart"},
	}
	for _, test := range tests {
		method, path, err := resolveOpenAIOperation(platform.TunnelRequest{
			Operation:      test.operation,
			OperationInput: platform.ProviderOperationInput{ResourceID: "id/with/slashes"},
		})
		if err != nil || method != test.method || path != test.path {
			t.Fatalf("%s resolved to %s %s (%v), want %s %s", test.operation, method, path, err, test.method, test.path)
		}
	}
}

func TestNormalizeMistralJSONRequestMapsCompatibilityAliases(t *testing.T) {
	body, err := normalizeMistralJSONRequest(
		platform.ProviderOperationOpenAIEmbeddings,
		[]byte(`{"model":"codestral-embed-2505","input":"code","dimensions":256,"seed":7,"output_dtype":"int8"}`),
	)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]interface{}
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatal(err)
	}
	if got["output_dimension"] != float64(256) || got["random_seed"] != float64(7) || got["output_dtype"] != "int8" {
		t.Fatalf("unexpected normalized request: %#v", got)
	}
	if _, exists := got["dimensions"]; exists {
		t.Fatal("dimensions alias was not removed")
	}
	if _, exists := got["seed"]; exists {
		t.Fatal("seed alias was not removed")
	}
}

func TestNormalizeMistralJSONRequestRejectsAliasConflicts(t *testing.T) {
	_, err := normalizeMistralJSONRequest(
		platform.ProviderOperationOpenAIEmbeddings,
		[]byte(`{"dimensions":256,"output_dimension":512}`),
	)
	if err == nil || !strings.Contains(err.Error(), "both dimensions and output_dimension") {
		t.Fatalf("expected alias conflict, got %v", err)
	}
}
