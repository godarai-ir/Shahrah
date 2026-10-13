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

// withCreditShadow wraps the spend logger with shadow credit computation when enabled.
func (b *builder) withCreditShadow(next proxy.SpendLogger) proxy.SpendLogger {
	if !b.cfg.Runtime.BillingCreditsShadowEnabled {
		return next
	}

	return proxy.NewCreditShadowSpendLogger(next, b.creditRateCards, b.logger)
}

// creditSurfaceMiddleware returns the response credit-reporting middleware, or
// nil when shadow credits are disabled so responses stay unchanged.
func (b *builder) creditSurfaceMiddleware() fiber.Handler {
	if !b.cfg.Runtime.BillingCreditsShadowEnabled {
		return nil
	}

	return proxy.CreditSurfaceMiddleware(b.creditRateCards)
}
