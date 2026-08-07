package admin

import (
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/gofiber/fiber/v2"
)

type topUpIntentContextKey struct{}

func setTopUpIntent(c *fiber.Ctx, intent *platform.WalletTopUpIntent) {
	c.Locals(topUpIntentContextKey{}, intent)
}

func topUpIntentFromContext(c *fiber.Ctx) *platform.WalletTopUpIntent {
	if c == nil {
		return nil
	}

	intent, ok := c.Locals(topUpIntentContextKey{}).(*platform.WalletTopUpIntent)
	if !ok {
		return nil
	}

	return intent
}
