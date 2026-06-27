package proxy

import (
	"context"
	"strings"
	"time"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/services/control-plane/internal/apiheaders"
	"github.com/Delpak/shahrah/services/control-plane/internal/guardrails"
	"github.com/Delpak/shahrah/services/control-plane/internal/observability/tracing"
	"github.com/gofiber/fiber/v2"
)

func (h *ChatCompletionHandler) resolveTenantGuardrails(ctx context.Context, tenantID string, selection guardrails.RequestSelection) (map[string]guardrails.Guardrail, error) {
	out := map[string]guardrails.Guardrail{}
	if h == nil || h.guardrailRepo == nil {
		return out, nil
	}

	slugs := guardrailSlugsFromSelection(selection)
	if len(slugs) == 0 {
		return out, nil
	}

	items, err := h.guardrailRepo.ListEnabledTenantGuardrailsBySlugs(ctx, tenantID, slugs)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		g := guardrails.NewFromTenantGuardrail(item)
		if g == nil {
			continue
		}

		out[normalizeGuardrailSelectorKey(item.Slug)] = g
		out[normalizeGuardrailSelectorKey(item.Selector)] = g
	}

	return out, nil
}

func normalizeGuardrailSelectorKey(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	return strings.ReplaceAll(value, "-", "_")
}

func guardrailSlugsFromSelection(selection guardrails.RequestSelection) []string {
	all := append([]string{}, selection.LLMInput...)
	all = append(all, selection.LLMOutput...)
	out := []string{}
	seen := map[string]bool{}

	for _, selector := range all {
		slug := guardrailSlugFromSelector(selector)
		if slug == "" || seen[slug] {
			continue
		}

		seen[slug] = true
		out = append(out, slug)
	}

	return out
}

func guardrailSlugFromSelector(selector string) string {
	selector = strings.TrimSpace(selector)
	if selector == "" {
		return ""
	}

	parts := strings.Split(selector, "/")

	return strings.TrimSpace(parts[len(parts)-1])
}

func (h *ChatCompletionHandler) parseGuardrails(c *fiber.Ctx) (guardrails.RequestSelection, error) {
	if c == nil {
		return guardrails.RequestSelection{}, nil
	}

	raw := apiheaders.GetGuardrails(c)
	scope := apiheaders.GetGuardrailsScope(c)

	return guardrails.ParseSelection(raw, scope)
}

func (h *ChatCompletionHandler) recordGuardrailChecks(c *fiber.Ctx, checks map[string][]guardrails.Check) {
	if c == nil || len(checks) == 0 {
		return
	}

	for hook, hookChecks := range checks {
		for _, check := range hookChecks {
			status := fiber.StatusOK
			if check.Status == "blocked" || check.Status == "error" {
				status = fiber.StatusBadRequest
			}

			tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "guardrail."+hook, "internal", time.Now(), status, map[string]interface{}{
				"selector":    check.Selector,
				"status":      check.Status,
				"operation":   check.Operation,
				"enforcement": check.Enforcement,
				"latency_ms":  check.LatencyMS,
				"mutated":     check.Mutated,
				"reason":      check.Reason,
			})
		}
	}
}

func writeGuardrailError(c *fiber.Ctx, status int, message string, checks map[string][]guardrails.Check) error {
	return c.Status(status).JSON(struct {
		Error           contractsopenai.ErrorInfo     `json:"error"`
		GuardrailChecks map[string][]guardrails.Check `json:"guardrail_checks"`
	}{
		Error: contractsopenai.ErrorInfo{
			Message: message,
			Type:    "guardrail_checks_failed",
			Code:    "guardrail_blocked",
		},
		GuardrailChecks: checks,
	})
}
