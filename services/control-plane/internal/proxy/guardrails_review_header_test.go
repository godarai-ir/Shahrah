package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Delpak/shahrah/services/control-plane/internal/apiheaders"
	"github.com/Delpak/shahrah/services/control-plane/internal/guardrails"
	"github.com/gofiber/fiber/v2"
)

func TestSetGuardrailReviewHeaderMergesHooks(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		setGuardrailReviewHeader(c, map[string][]guardrails.Check{
			"llm_input_guardrails": {{Selector: "guardrails/security-baseline", Status: "review"}, {Selector: "guardrails/pii", Status: "passed"}},
		})
		setGuardrailReviewHeader(c, map[string][]guardrails.Check{
			"llm_output_guardrails": {{Selector: "commitments", Status: "review"}, {Selector: "guardrails/security-baseline", Status: "review"}},
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

	if got := resp.Header.Get(apiheaders.HeaderGuardrailReview); got != "security-baseline,commitments" {
		t.Fatalf("review header = %q", got)
	}
}
