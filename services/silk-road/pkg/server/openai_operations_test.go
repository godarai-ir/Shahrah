package server

import (
	"net/http"
	"strings"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

func TestResolveOpenAIOperationUsesFixedEscapedResourcePath(t *testing.T) {
	method, path, err := resolveOpenAIOperation(platform.TunnelRequest{
		Operation: platform.ProviderOperationOpenAIResponsesCancel,
		OperationInput: platform.ProviderOperationInput{
			ResourceID: "resp/tenant-bound",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if method != http.MethodPost || path != "/responses/resp%2Ftenant-bound/cancel" {
		t.Fatalf("unexpected operation target: %s %s", method, path)
	}
}

func TestResolveOpenAIOperationRejectsArbitraryLegacyPath(t *testing.T) {
	_, _, err := resolveOpenAIOperation(platform.TunnelRequest{
		RequestMetadata: map[string]interface{}{
			"provider_endpoint_path": "https://attacker.invalid/secrets",
		},
	})
	if err == nil {
		t.Fatal("expected arbitrary provider path to be rejected")
	}
}

func TestResolveOpenAIOperationBuildsBoundedListQuery(t *testing.T) {
	method, path, err := resolveOpenAIOperation(platform.TunnelRequest{
		Operation: platform.ProviderOperationOpenAIBatchesList,
		OperationInput: platform.ProviderOperationInput{
			AfterID: "batch_1",
			Limit:   25,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if method != http.MethodGet || path != "/batches?after=batch_1&limit=25" {
		t.Fatalf("unexpected list operation target: %s %s", method, path)
	}
}

func TestResolveOpenAIResponsesRetrievePreservesOfficialQuery(t *testing.T) {
	includeObfuscation := false
	startingAfter := int64(9)
	stream := true
	method, path, err := resolveOpenAIOperation(platform.TunnelRequest{
		Operation: platform.ProviderOperationOpenAIResponsesGet,
		OperationInput: platform.ProviderOperationInput{
			ResourceID:         "resp_1",
			Include:            []string{"reasoning.encrypted_content", "message.output_text.logprobs"},
			IncludeObfuscation: &includeObfuscation,
			StartingAfter:      &startingAfter,
			Stream:             &stream,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if method != http.MethodGet || !strings.HasPrefix(path, "/responses/resp_1?") {
		t.Fatalf("unexpected retrieve operation target: %s %s", method, path)
	}
	for _, expected := range []string{
		"include=reasoning.encrypted_content",
		"include=message.output_text.logprobs",
		"include_obfuscation=false",
		"starting_after=9",
		"stream=true",
	} {
		if !strings.Contains(path, expected) {
			t.Fatalf("expected %q in %q", expected, path)
		}
	}
}

func TestResolveMistralNativeOperationsUseFixedPaths(t *testing.T) {
	createdByMe := true
	page := 2
	pageSize := 25

	method, path, err := resolveOpenAIOperation(platform.TunnelRequest{
		Operation: platform.ProviderOperationMistralAgentsList,
		OperationInput: platform.ProviderOperationInput{
			CreatedByMe: &createdByMe,
			Page:        &page,
			PageSize:    &pageSize,
			OrderBy:     "created_at",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if method != http.MethodGet || !strings.HasPrefix(path, "/agents?") {
		t.Fatalf("unexpected Mistral list target: %s %s", method, path)
	}

	for _, expected := range []string{"created_by_me=true", "page=2", "page_size=25", "order_by=created_at"} {
		if !strings.Contains(path, expected) {
			t.Fatalf("expected %q in %q", expected, path)
		}
	}

	method, path, err = resolveOpenAIOperation(platform.TunnelRequest{
		Operation:      platform.ProviderOperationMistralConversationsRestart,
		OperationInput: platform.ProviderOperationInput{ResourceID: "conv/tenant-bound"},
	})
	if err != nil {
		t.Fatal(err)
	}

	if method != http.MethodPost || path != "/conversations/conv%2Ftenant-bound/restart" {
		t.Fatalf("unexpected Mistral resource target: %s %s", method, path)
	}
}
