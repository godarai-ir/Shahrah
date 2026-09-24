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

			attributes := map[string]interface{}{
				"selector":    check.Selector,
				"status":      check.Status,
				"operation":   check.Operation,
				"enforcement": check.Enforcement,
				"latency_ms":  check.LatencyMS,
				"mutated":     check.Mutated,
				"reason":      check.Reason,
			}
			addSemanticTraceAttributes(attributes, check.Semantic)

			tracing.RecordStepSpan(c, tracing.QueueFromContext(c), "", "guardrail."+hook, "internal", time.Now(), status, attributes)
		}
	}

	setGuardrailReviewHeader(c, checks)
}

// addSemanticTraceAttributes records rule probabilities on the trace only; API responses never carry them.
func addSemanticTraceAttributes(attributes map[string]interface{}, detail *guardrails.SemanticCheckDetail) {
	if detail == nil {
		return
	}

	attributes["semantic_decision"] = detail.Decision
	attributes["semantic_language"] = detail.Language
	attributes["semantic_rules"] = detail.Rules
	attributes["semantic_truncated"] = detail.Truncated
	attributes["semantic_cached"] = detail.Cached
	attributes["semantic_model"] = detail.Model

	if detail.Severity != nil {
		attributes["semantic_severity"] = *detail.Severity
	}
}

// setGuardrailReviewHeader lists the guardrails that put this exchange in the review band. Review never blocks in v1.
func setGuardrailReviewHeader(c *fiber.Ctx, checks map[string][]guardrails.Check) {
	var selectors []string

	for _, hookChecks := range checks {
		for _, check := range hookChecks {
			if check.Status == "review" {
				selectors = append(selectors, guardrailSlugFromSelector(check.Selector))
			}
		}
	}

	if len(selectors) == 0 {
		return
	}

	if existing := strings.TrimSpace(string(c.Response().Header.Peek(apiheaders.HeaderGuardrailReview))); existing != "" {
		selectors = append(strings.Split(existing, ","), selectors...)
	}

	c.Set(apiheaders.HeaderGuardrailReview, strings.Join(uniqueNonEmpty(selectors), ","))
}

func uniqueNonEmpty(values []string) []string {
	out := make([]string, 0, len(values))
	seen := make(map[string]bool, len(values))

	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}

		seen[value] = true
		out = append(out, value)
	}

	return out
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
