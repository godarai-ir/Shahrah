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
	defer response.Body.Close() //nolint:errcheck // Test cleanup is best-effort.

	if response.StatusCode != fiber.StatusNoContent {
		t.Fatalf("status = %d", response.StatusCode)
	}
}

func TestRewriteMistralListFiltersUnknownUpstreamResources(t *testing.T) {
	body := []byte(`{"data":[{"id":"upstream-owned"},{"id":"upstream-other"}]}`)
	rewritten, err := rewriteMistralResponseJSON(body, map[string]string{"upstream-owned": "magent_gdr_owned"}, true)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(rewritten), "upstream-") || strings.Contains(string(rewritten), "other") {
		t.Fatalf("upstream or unowned resource leaked: %s", rewritten)
	}
	if !strings.Contains(string(rewritten), "magent_gdr_owned") {
		t.Fatalf("public id missing: %s", rewritten)
	}
}

func TestMistralBatchUsageAggregatesJSONLResponses(t *testing.T) {
	input := strings.Join([]string{
		`{"response":{"body":{"usage":{"prompt_tokens":10,"completion_tokens":3,"total_tokens":13}}}}`,
		`{"result":{"usage":{"prompt_tokens":5,"completion_tokens":2,"total_tokens":7}}}`,
	}, "\n")

	usage, err := parseMistralBatchUsage(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if usage.PromptTokens != 15 || usage.CompletionTokens != 5 || usage.TotalTokens != 20 {
		t.Fatalf("unexpected usage: %+v", usage)
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
