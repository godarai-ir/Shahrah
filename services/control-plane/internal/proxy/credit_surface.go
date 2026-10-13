package proxy

import (
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/billing"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/credits"
	"github.com/gofiber/fiber/v2"
)

type creditSurfaceKey struct{}

// CreditSurfaceMiddleware exposes the credit rate card to gateway handlers so
// responses can report credits next to the existing cost headers. Register it
// only when shadow credits are enabled.
func CreditSurfaceMiddleware(rateCards billing.CreditRateCardProvider) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if rateCards != nil && strings.HasPrefix(c.Path(), "/v1/") {
			c.Locals(creditSurfaceKey{}, rateCards)
		}

		return c.Next()
	}
}

// creditSurface returns the rate-card provider for this request, or nil when
// credit reporting is disabled. Capture it before starting a body stream writer.
func creditSurface(c *fiber.Ctx) billing.CreditRateCardProvider {
	if c == nil {
		return nil
	}

	provider, ok := c.Locals(creditSurfaceKey{}).(billing.CreditRateCardProvider)
	if !ok {
		return nil
	}

	return provider
}

// creditsUsed formats the credit charge for a final cost, or "" when credit
// reporting is disabled or the cost cannot be converted.
func creditsUsed(rateCards billing.CreditRateCardProvider, cost platform.UsageCostBreakdown) string {
	if rateCards == nil {
		return ""
	}

	charge, err := credits.FromBreakdown(cost, rateCards.CurrentRateCard())
	if err != nil {
		return ""
	}

	return charge.Credits()
}
