package billing

import (
	"context"
	"sync"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/domain/credits"
	"go.uber.org/zap"
)

const (
	defaultCreditRateCardTTL     = 5 * time.Minute
	creditRateCardRefreshTimeout = 5 * time.Second
)

// CreditRateCardStore loads the published credit rate card.
type CreditRateCardStore interface {
	GetPublishedCreditRateCard(ctx context.Context) (credits.RateCard, error)
}

// CreditRateCardProvider returns the credit rate card for hot-path use.
type CreditRateCardProvider interface {
	CurrentRateCard() credits.RateCard
}

// CreditRateCardCache serves the published rate card from memory. Reads never
// block on the database: a stale card triggers one background refresh and the
// last good card (or the identity card before the first load) is returned.
type CreditRateCardCache struct {
	store  CreditRateCardStore
	ttl    time.Duration
	logger *zap.Logger
	now    func() time.Time

	mu         sync.RWMutex
	card       credits.RateCard
	loadedAt   time.Time
	refreshing bool
}

// NewCreditRateCardCache creates a rate-card cache. A non-positive ttl uses the default.
func NewCreditRateCardCache(store CreditRateCardStore, ttl time.Duration, logger *zap.Logger) *CreditRateCardCache {
	if ttl <= 0 {
		ttl = defaultCreditRateCardTTL
	}

	if logger == nil {
		logger = zap.NewNop()
	}

	return &CreditRateCardCache{
		store:  store,
		ttl:    ttl,
		logger: logger,
		now:    time.Now,
		card:   credits.IdentityRateCard(),
	}
}

// Prime loads the rate card synchronously, typically at startup.
func (c *CreditRateCardCache) Prime(ctx context.Context) error {
	if c == nil || c.store == nil {
		return nil
	}

	card, err := c.store.GetPublishedCreditRateCard(ctx)
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.card = card
	c.loadedAt = c.now()
	c.mu.Unlock()

	return nil
}

// CurrentRateCard returns the cached rate card and refreshes it in the background when stale.
func (c *CreditRateCardCache) CurrentRateCard() credits.RateCard {
	if c == nil {
		return credits.IdentityRateCard()
	}

	c.mu.RLock()
	card := c.card
	stale := c.now().Sub(c.loadedAt) >= c.ttl
	c.mu.RUnlock()

	if stale {
		c.refreshAsync()
	}

	return card
}

func (c *CreditRateCardCache) refreshAsync() {
	if c.store == nil {
		return
	}

	c.mu.Lock()
	if c.refreshing {
		c.mu.Unlock()

		return
	}

	c.refreshing = true
	c.mu.Unlock()

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), creditRateCardRefreshTimeout)
		defer cancel()

		err := c.Prime(ctx)

		c.mu.Lock()
		c.refreshing = false

		if err != nil {
			// Back off for a full TTL instead of retrying on every request.
			c.loadedAt = c.now()
		}
		c.mu.Unlock()

		if err != nil {
			c.logger.Warn("credit rate card refresh failed; serving last good card", zap.Error(err))
		}
	}()
}
