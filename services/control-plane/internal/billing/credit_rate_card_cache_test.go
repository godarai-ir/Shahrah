package billing

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/domain/credits"
)

type fakeCreditRateCardStore struct {
	mu    sync.Mutex
	card  credits.RateCard
	err   error
	calls int
}

func (s *fakeCreditRateCardStore) GetPublishedCreditRateCard(context.Context) (credits.RateCard, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.calls++

	return s.card, s.err
}

func (s *fakeCreditRateCardStore) callCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.calls
}

func TestCreditRateCardCacheServesIdentityBeforeLoad(t *testing.T) {
	t.Parallel()

	cache := NewCreditRateCardCache(nil, time.Minute, nil)
	if got := cache.CurrentRateCard(); got.Version != 1 {
		t.Fatalf("CurrentRateCard().Version = %d, want identity version 1", got.Version)
	}
}

func TestCreditRateCardCachePrimeAndKeepLastGood(t *testing.T) {
	t.Parallel()

	store := &fakeCreditRateCardStore{card: credits.RateCard{Version: 4}}
	cache := NewCreditRateCardCache(store, time.Minute, nil)

	if err := cache.Prime(context.Background()); err != nil {
		t.Fatalf("Prime: %v", err)
	}

	if got := cache.CurrentRateCard(); got.Version != 4 {
		t.Fatalf("CurrentRateCard().Version = %d, want 4", got.Version)
	}

	store.mu.Lock()
	store.err = errors.New("db down")
	store.mu.Unlock()

	if err := cache.Prime(context.Background()); err == nil {
		t.Fatal("Prime with store error returned nil")
	}

	if got := cache.CurrentRateCard(); got.Version != 4 {
		t.Fatalf("CurrentRateCard().Version = %d after failed refresh, want last good 4", got.Version)
	}
}

func TestCreditRateCardCacheRefreshesWhenStale(t *testing.T) {
	t.Parallel()

	store := &fakeCreditRateCardStore{card: credits.RateCard{Version: 2}}
	cache := NewCreditRateCardCache(store, time.Minute, nil)

	now := time.Now()
	cache.now = func() time.Time { return now }

	if err := cache.Prime(context.Background()); err != nil {
		t.Fatalf("Prime: %v", err)
	}

	store.mu.Lock()
	store.card = credits.RateCard{Version: 3}
	store.mu.Unlock()

	if got := cache.CurrentRateCard(); got.Version != 2 {
		t.Fatalf("fresh read version = %d, want 2", got.Version)
	}

	cache.mu.Lock()
	cache.now = func() time.Time { return now.Add(2 * time.Minute) }
	cache.mu.Unlock()

	// The stale read returns the cached card and triggers a background refresh.
	if got := cache.CurrentRateCard(); got.Version != 2 {
		t.Fatalf("stale read version = %d, want cached 2", got.Version)
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if cache.CurrentRateCard().Version == 3 {
			return
		}

		time.Sleep(10 * time.Millisecond)
	}

	t.Fatalf("rate card was not refreshed; store calls = %d", store.callCount())
}
