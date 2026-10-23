package proxy

import (
	"context"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

const pricingChangelogCacheControl = "public, max-age=300"

// PricingChangelogStore reads the public pricing changelog.
type PricingChangelogStore interface {
	ListPricingChangelog(ctx context.Context) (*platform.PricingChangelog, error)
}

// NewPricingChangelogHandler serves GET /v1/pricing/changelog: every rate-card
// version and credit list price with its effective dates. It is public and
// cacheable, and it never exposes the USD conversion factor.
func NewPricingChangelogHandler(store PricingChangelogStore, logger *zap.Logger) fiber.Handler {
	if logger == nil {
		logger = zap.NewNop()
	}

	return func(c *fiber.Ctx) error {
		changelog, err := store.ListPricingChangelog(c.Context())
		if err != nil {
			logger.Warn("pricing changelog read failed", zap.Error(err))

			return fiber.NewError(fiber.StatusServiceUnavailable, "pricing changelog is unavailable")
		}

		c.Set(fiber.HeaderCacheControl, pricingChangelogCacheControl)

		return c.JSON(changelog)
	}
}
