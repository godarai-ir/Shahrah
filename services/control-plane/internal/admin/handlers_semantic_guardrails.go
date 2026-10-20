package admin

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/guardrails"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
)

const (
	maxSemanticTestTextRunes    = 8000
	semanticTestRateLimit       = 30
	semanticTestRateLimitWindow = time.Minute
)

// semanticPolicyTester evaluates unsaved semantic_policy configs for the wizard's live test.
type semanticPolicyTester interface {
	TestSemanticPolicy(ctx context.Context, config map[string]interface{}, text string) (platform.SemanticPolicyTestResponse, error)
}

// TestTenantSemanticGuardrail runs one sample text through an unsaved semantic_policy config. Every call is a paid
// judge request, so it is limited per tenant.
func (h *Handlers) TestTenantSemanticGuardrail(c *fiber.Ctx) error {
	if err := requirePolicyManage(c); err != nil {
		return err
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}

	var req platform.SemanticPolicyTestRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	text := strings.TrimSpace(req.Text)
	if text == "" || utf8.RuneCountInString(text) > maxSemanticTestTextRunes {
		return fiber.NewError(fiber.StatusBadRequest, "text must be 1-"+strconv.Itoa(maxSemanticTestTextRunes)+" characters")
	}

	if _, err := guardrails.ParseSemanticPolicyConfig(req.Config); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	allowed, wait := h.semanticTestLimiter.allow(c.Context(), authRateLimitPolicy{
		key:    authRateLimitKey("semantic-guardrail-test", currentTenant.ID),
		limit:  semanticTestRateLimit,
		window: semanticTestRateLimitWindow,
	})
	if !allowed {
		c.Set("Retry-After", strconv.Itoa(max(1, int(wait.Round(time.Second)/time.Second))))

		return fiber.NewError(fiber.StatusTooManyRequests, "too many semantic guardrail tests; try again shortly")
	}

	if h.semanticGuardrails == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "semantic guardrails are not configured")
	}

	result, err := h.semanticGuardrails.TestSemanticPolicy(c.Context(), req.Config, text)

	switch {
	case err == nil:
		return c.JSON(result)
	case errors.Is(err, guardrails.ErrInvalidSemanticPolicy):
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	case errors.Is(err, guardrails.ErrSemanticJudgeUnavailable):
		return fiber.NewError(fiber.StatusServiceUnavailable, "semantic guardrails are not configured")
	default:
		return fiber.NewError(fiber.StatusBadGateway, "semantic judge is unavailable")
	}
}
