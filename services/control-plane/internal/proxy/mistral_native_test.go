package proxy

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/gofiber/fiber/v2"
)

func TestMistralNativeFeatureFlagsDefaultOff(t *testing.T) {
	t.Setenv("SHAHRAH_MISTRAL_ENABLED", "true")
	t.Setenv("SHAHRAH_MISTRAL_FILES_ENABLED", "")

	if mistralNativeFeatureEnabled("FILES") {
		t.Fatal("native Files API must default off")
	}

	t.Setenv("SHAHRAH_MISTRAL_FILES_ENABLED", "true")

	if !mistralNativeFeatureEnabled("FILES") {
		t.Fatal("explicitly enabled native Files API was rejected")
	}
}

func TestMistralOperationParserUsesPublicResourceID(t *testing.T) {
	app := fiber.New()
	app.Post("/v1/conversations/:id/restart", func(c *fiber.Ctx) error {
		operation, input, err := mistralOperationFromRequest(c)
		if err != nil {
			return err
		}

		if operation != platform.ProviderOperationMistralConversationsRestart {
			t.Fatalf("operation = %s", operation)
		}

		if input.ResourceID != "mconv_gdr_owned" {
			t.Fatalf("resource id = %q", input.ResourceID)
		}

		return c.SendStatus(fiber.StatusNoContent)
	})

	request := httptest.NewRequest(http.MethodPost, "/v1/conversations/mconv_gdr_owned/restart", bytes.NewBufferString(`{"inputs":"continue"}`))

	response, err := app.Test(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()

	if response.StatusCode != fiber.StatusNoContent {
		t.Fatalf("status = %d", response.StatusCode)
	}
}

func TestRewriteMistralListFiltersUnknownUpstreamResources(t *testing.T) {
	body := []byte(`{"data":[{"id":"upstream-owned"},{"id":"upstream-other"}],"total":99}`)

	rewritten, err := rewriteMistralOwnedListResponseJSON(
		body,
		map[string]string{"upstream-owned": "magent_gdr_owned", "related-file": "mfile_gdr_related"},
		map[string]string{"upstream-owned": "magent_gdr_owned"},
		true,
	)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(string(rewritten), "upstream-") || strings.Contains(string(rewritten), "other") {
		t.Fatalf("upstream or unowned resource leaked: %s", rewritten)
	}

	if !strings.Contains(string(rewritten), "magent_gdr_owned") {
		t.Fatalf("public id missing: %s", rewritten)
	}

	if !strings.Contains(string(rewritten), `"total":1`) {
		t.Fatalf("owned count was not sanitized: %s", rewritten)
	}
}

func TestMistralResponseIdentityUsesConversationID(t *testing.T) {
	id, status, err := mistralResponseIdentity([]byte(`{"conversation_id":"conv-upstream"}`))
	if err != nil {
		t.Fatal(err)
	}

	if id != "conv-upstream" || status != "active" {
		t.Fatalf("identity = %q, %q", id, status)
	}
}

func TestMistralFileMetadataRetainsBatchLineCount(t *testing.T) {
	metadata := mistralResponseResourceMetadata([]byte(`{"id":"file-upstream","num_lines":12,"filename":"input.jsonl"}`))

	lines, ok := mistralMetadataPositiveInt(metadata, "num_lines")
	if !ok || lines != 12 {
		t.Fatalf("num_lines = %d, %v", lines, ok)
	}

	if _, leaked := metadata["id"]; leaked {
		t.Fatal("upstream id leaked into resource metadata")
	}
}

func TestMistralNestedBatchModel(t *testing.T) {
	body := []byte(`{"endpoint":"/v1/chat/completions","requests":[{"body":{"model":"mistral-large-2512"}}]}`)

	model, _, err := mistralRequestModel(platform.ProviderOperationMistralBatchCreate, body)
	if err != nil {
		t.Fatal(err)
	}

	if model != "mistral-large-2512" {
		t.Fatalf("model = %q", model)
	}
}

func TestMistralNestedBuiltInToolIsRejected(t *testing.T) {
	body := []byte(`{"requests":[{"body":{"tools":[{"type":"web_search"}]}}]}`)

	err := validateMistralNativePolicy(platform.ProviderOperationMistralBatchCreate, body)
	if err == nil || !strings.Contains(err.Error(), "web_search") {
		t.Fatalf("error = %v", err)
	}
}

func TestMistralBatchFileRequiresJobLevelModel(t *testing.T) {
	var value interface{}
	if err := json.Unmarshal([]byte(`{"custom_id":"1","body":{"model":"mistral-small-latest","messages":[]}}`), &value); err != nil {
		t.Fatal(err)
	}

	err := validateMistralBatchFileRequest(value)
	if err == nil || !strings.Contains(err.Error(), "creating the batch job") {
		t.Fatalf("error = %v", err)
	}
}

func TestMistralBatchTerminalStatuses(t *testing.T) {
	for _, status := range []string{"SUCCESS", "failed", "timeout_exceeded", "CANCELLED"} { //nolint:misspell // Verify Mistral's alternate provider spelling.
		if !mistralBatchStatusIsTerminal(status) {
			t.Fatalf("terminal status rejected: %s", status)
		}
	}

	if mistralBatchStatusIsTerminal("running") {
		t.Fatal("running batch was considered terminal")
	}
}

func TestMistralBatchUsageAggregatesJSONLResponses(t *testing.T) {
	input := `{"response":{"body":{"usage":{"prompt_tokens":10,"completion_tokens":3,"total_tokens":13}}}}` + "\n" +
		`{"result":{"usage":{"prompt_tokens":5,"completion_tokens":2,"total_tokens":7}}}`

	usage, err := parseMistralBatchUsage(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}

	if usage.PromptTokens != 15 || usage.CompletionTokens != 5 || usage.TotalTokens != 20 {
		t.Fatalf("unexpected usage: %+v", usage)
	}
}

func TestMistralUsageRejectsIncompleteTokenAccounting(t *testing.T) {
	var value interface{}
	if err := json.Unmarshal([]byte(`{"usage":{"total_tokens":12}}`), &value); err != nil {
		t.Fatal(err)
	}

	if usage := mistralUsageFromValue(value); usage != nil {
		t.Fatalf("incomplete usage was accepted: %+v", usage)
	}
}

func TestMistralBatchFileReferencesAreDiscovered(t *testing.T) {
	body, err := json.Marshal(map[string]interface{}{
		"output_file": "file-output",
		"errors":      map[string]interface{}{"error_file_id": "file-errors"},
	})
	if err != nil {
		t.Fatal(err)
	}

	references := mistralResponseFileReferences(body)
	if len(references) != 2 {
		t.Fatalf("references = %v", references)
	}
}
