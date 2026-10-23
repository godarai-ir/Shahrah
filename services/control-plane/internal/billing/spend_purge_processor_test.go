package billing

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type fakeSpendPurgeStore struct {
	mu       sync.Mutex
	calls    int
	failures int
}

func (f *fakeSpendPurgeStore) RunPendingSpendPurges(context.Context) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.calls++
	if f.calls <= f.failures {
		return 1, errors.New("clickhouse down")
	}

	return 0, nil
}

func (f *fakeSpendPurgeStore) callCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.calls
}

func TestSpendPurgeProcessorStopsOnceNothingIsPending(t *testing.T) {
	t.Parallel()

	store := &fakeSpendPurgeStore{}
	processor := NewSpendPurgeProcessor(store, nil, time.Millisecond)

	processor.Start()
	<-processor.done
	processor.Stop()

	require.Equal(t, 1, store.callCount())
}

func TestSpendPurgeProcessorRetriesUntilThePurgeSucceeds(t *testing.T) {
	t.Parallel()

	store := &fakeSpendPurgeStore{failures: 2}
	processor := NewSpendPurgeProcessor(store, nil, time.Millisecond)

	processor.Start()
	<-processor.done
	processor.Stop()

	require.Equal(t, 3, store.callCount())
}
