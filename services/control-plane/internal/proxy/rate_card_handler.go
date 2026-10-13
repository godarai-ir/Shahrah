package proxy

import (
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/billing"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/credits"
	"github.com/gofiber/fiber/v2"
)

const rateCardCacheControl = "public, max-age=300"

// CreditRateCardBuilder builds the public credit rate card for a rate-card version.
type CreditRateCardBuilder interface {
	CreditRateCard(card credits.RateCard) (platform.CreditRateCard, string)
}

// NewRateCardHandler serves GET /v1/rate-card. The response is public and
// cacheable; a matching If-None-Match returns 304.
func NewRateCardHandler(builder CreditRateCardBuilder, rateCards billing.CreditRateCardProvider) fiber.Handler {
	return func(c *fiber.Ctx) error {
		card := credits.IdentityRateCard()
		if rateCards != nil {
			card = rateCards.CurrentRateCard()
		}

		body, revision := builder.CreditRateCard(card)
		etag := `"` + revision + `"`

		c.Set(fiber.HeaderCacheControl, rateCardCacheControl)
		c.Set(fiber.HeaderETag, etag)

		if etagMatches(c.Get(fiber.HeaderIfNoneMatch), etag) {
			return c.SendStatus(fiber.StatusNotModified)
		}

		return c.JSON(body)
	}
}

func etagMatches(header, etag string) bool {
	for _, candidate := range strings.Split(header, ",") {
		candidate = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(candidate), "W/"))
		if candidate == etag || candidate == "*" {
			return true
		}
	}

	return false
}
