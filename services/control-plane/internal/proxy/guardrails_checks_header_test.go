package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Delpak/shahrah/services/control-plane/internal/apiheaders"
	"github.com/Delpak/shahrah/services/control-plane/internal/guardrails"
	"github.com/gofiber/fiber/v2"
)

func TestSetGuardrailChecksHeaderSummarizesHooksInOrder(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		setGuardrailChecksHeader(c, map[string][]guardrails.Check{
			"llm_input_guardrails": {
				{Selector: "guardrails/pii-redaction", Status: "mutated"},
				{Selector: "guardrails/keyword-blocklist", Status: "passed"},
			},
		})
		setGuardrailChecksHeader(c, map[string][]guardrails.Check{
			"llm_output_guardrails": {{Selector: "guardrails/secrets-detection", Status: "passed"}},
			"llm_input_guardrails":  {{Selector: "", Status: "passed"}},
		})

		return c.SendStatus(fiber.StatusOK)
	})

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/", http.NoBody))
	if err != nil {
		t.Fatalf("request: %v", err)
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			t.Errorf("close body: %v", closeErr)
		}
	}()

	want := "input:pii-redaction=mutated,input:keyword-blocklist=passed,output:secrets-detection=passed"
	if got := resp.Header.Get(apiheaders.HeaderGuardrailChecks); got != want {
		t.Fatalf("checks header = %q, want %q", got, want)
	}
}
