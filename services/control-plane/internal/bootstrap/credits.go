package bootstrap

import (
	"context"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/billing"
	"github.com/Delpak/shahrah/services/control-plane/internal/proxy"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

const creditRateCardPrimeTimeout = 5 * time.Second

// buildCreditRateCards creates the credit rate-card cache and loads the published card.
func (b *builder) buildCreditRateCards() {
	var store billing.CreditRateCardStore
	if b.store != nil && b.store.WalletRepository != nil && b.store.CreditRateCardRepository != nil {
		store = b.store.CreditRateCardRepository
	}

	b.creditRateCards = billing.NewCreditRateCardCache(store, b.cfg.Runtime.BillingCreditRateCardTTL, b.logger)

	ctx, cancel := context.WithTimeout(b.ctx, creditRateCardPrimeTimeout)
	defer cancel()

	if err := b.creditRateCards.Prime(ctx); err != nil {
		b.logger.Warn("Initial credit rate card is unavailable; serving the identity card until refresh succeeds", zap.Error(err))
	}
}

// withCreditShadow wraps the spend logger with credit pricing. Every tenant is
// billed in credits, and the ledger settles only spend logs that carry credits.
func (b *builder) withCreditShadow(next proxy.SpendLogger) proxy.SpendLogger {
	return proxy.NewCreditShadowSpendLogger(next, b.creditRateCards, b.logger)
}

// creditSurfaceMiddleware returns the response credit-reporting middleware.
func (b *builder) creditSurfaceMiddleware() fiber.Handler {
	return proxy.CreditSurfaceMiddleware(b.creditRateCards)
}

// pricingChangelogHandler serves the public pricing changelog when storage is available.
func (b *builder) pricingChangelogHandler() fiber.Handler {
	if b.store == nil || b.store.WalletRepository == nil {
		return nil
	}

	return proxy.NewPricingChangelogHandler(b.store.CreditPriceRepository, b.logger)
}
