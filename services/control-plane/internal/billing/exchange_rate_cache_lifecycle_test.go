package billing

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/stretchr/testify/require"
)

type lifecycleExchangeRateStore struct {
	mu              sync.Mutex
	calls           int
	refreshStarted  chan struct{}
	refreshFinished chan struct{}
}

func (s *lifecycleExchangeRateStore) GetWallet(context.Context, string) (*platform.Wallet, error) {
	return nil, nil
}

func (s *lifecycleExchangeRateStore) GetLatestExchangeRate(ctx context.Context) (*platform.ExchangeRate, error) {
	s.mu.Lock()
	s.calls++
	call := s.calls
	s.mu.Unlock()

	if call == 1 {
		return &platform.ExchangeRate{USDtoIRR: 850_000, Source: "initial"}, nil
	}

	close(s.refreshStarted)
	<-ctx.Done()
	close(s.refreshFinished)

	return nil, ctx.Err()
}

func (s *lifecycleExchangeRateStore) UpsertExchangeRate(context.Context, platform.ExchangeRate) (*platform.ExchangeRate, error) {
	return nil, nil
}

func (s *lifecycleExchangeRateStore) callCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.calls
}

func TestExchangeRateCacheShutdownCancelsAndDrainsRefresh(t *testing.T) {
	store := &lifecycleExchangeRateStore{
		refreshStarted:  make(chan struct{}),
		refreshFinished: make(chan struct{}),
	}
	baseTime := time.Date(2026, time.August, 7, 8, 0, 0, 0, time.UTC)
	rateCache := newExchangeRateCache(store, nil, platform.ExchangeRate{}, ExchangeRateCacheConfig{
		FreshTTL: time.Minute,
		MaxStale: 10 * time.Minute,
	})
	rateCache.now = func() time.Time { return baseTime }

	initial := rateCache.Get(context.Background())
	require.NotNil(t, initial)
	require.Equal(t, 850_000.0, initial.USDtoIRR)

	rateCache.now = func() time.Time { return baseTime.Add(2 * time.Minute) }
	stale := rateCache.Get(context.Background())
	require.NotNil(t, stale)

	select {
	case <-store.refreshStarted:
	case <-time.After(time.Second):
		t.Fatal("background refresh did not start")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	require.NoError(t, rateCache.shutdown(shutdownCtx))

	select {
	case <-store.refreshFinished:
	default:
		t.Fatal("shutdown returned before the background refresh finished")
	}

	require.NotNil(t, rateCache.Get(context.Background()))
	require.Equal(t, 2, store.callCount())
}
