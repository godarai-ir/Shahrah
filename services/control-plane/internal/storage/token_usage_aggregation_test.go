package storage

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
)

type fakeTokenUsageCache struct {
	queueSpendErr error
	queueSpend    struct {
		deltaKey string
		dirtyKey string
		deltaUSD float64
		ttl      time.Duration
		score    int64
		called   bool
	}
	queueLastErr error
	queueLast    struct {
		key      string
		dirtyKey string
		usedAt   int64
		ttl      time.Duration
		called   bool
	}
}

func (f *fakeTokenUsageCache) QueueTokenSpendDelta(_ context.Context, deltaKey string, dirtySetKey string, deltaUSD float64, ttl time.Duration, score int64) error {
	f.queueSpend.deltaKey = deltaKey
	f.queueSpend.dirtyKey = dirtySetKey
	f.queueSpend.deltaUSD = deltaUSD
	f.queueSpend.ttl = ttl
	f.queueSpend.score = score
	f.queueSpend.called = true

	return f.queueSpendErr
}

func (f *fakeTokenUsageCache) DrainTokenSpendDeltas(context.Context, string, int) ([]cache.TokenSpendDelta, error) {
	return nil, nil
}

func (f *fakeTokenUsageCache) QueueKeyLastUsed(_ context.Context, lastUsedKey string, dirtySetKey string, usedAtUnix int64, ttl time.Duration) (bool, error) {
	f.queueLast.key = lastUsedKey
	f.queueLast.dirtyKey = dirtySetKey
	f.queueLast.usedAt = usedAtUnix
	f.queueLast.ttl = ttl
	f.queueLast.called = true

	return true, f.queueLastErr
}

func (f *fakeTokenUsageCache) DrainKeyLastUsed(context.Context, string, int) ([]cache.KeyLastUsedSnapshot, error) {
	return nil, nil
}

func TestTokenUsageKeyRoundTrip(t *testing.T) {
	store := &Store{tokenUsageKeyPrefix: "usage:"}

	spendKey := store.tokenSpendDeltaKey("hash-1")

	tokenHash, err := store.parseTokenSpendDeltaKey(spendKey)
	if err != nil {
		t.Fatal(err)
	}

	if tokenHash != "hash-1" {
		t.Fatalf("unexpected token hash: %s", tokenHash)
	}

	if spendKey != "usage:token_spend_delta:hash-1" {
		t.Fatalf("unexpected spend key: %s", spendKey)
	}

	if store.tokenSpendDirtySetKey() != "usage:token_spend_delta:dirty" {
		t.Fatalf("unexpected spend dirty key: %s", store.tokenSpendDirtySetKey())
	}

	lastUsedKey := store.keyLastUsedKey("key-1")

	keyID, err := store.parseKeyLastUsedKey(lastUsedKey)
	if err != nil {
		t.Fatal(err)
	}

	if keyID != "key-1" {
		t.Fatalf("unexpected key id: %s", keyID)
	}

	if lastUsedKey != "usage:key_last_used:key-1" {
		t.Fatalf("unexpected last-used key: %s", lastUsedKey)
	}
}

func TestQueueTokenSpendDeltaUsesRedisAndFailOpen(t *testing.T) {
	ctx := context.Background()
	fake := &fakeTokenUsageCache{}
	store := &Store{
		tokenUsageCache:     fake,
		tokenUsageEnabled:   true,
		tokenUsageFailOpen:  true,
		tokenUsageKeyPrefix: "test-usage",
		tokenUsageDeltaTTL:  time.Hour,
	}

	if err := store.queueTokenSpendDelta(ctx, "hash-1", 1.25); err != nil {
		t.Fatal(err)
	}

	if !fake.queueSpend.called {
		t.Fatal("expected redis queue call")
	}

	if fake.queueSpend.deltaKey != "test-usage:token_spend_delta:hash-1" {
		t.Fatalf("unexpected delta key: %s", fake.queueSpend.deltaKey)
	}

	if fake.queueSpend.dirtyKey != "test-usage:token_spend_delta:dirty" {
		t.Fatalf("unexpected dirty key: %s", fake.queueSpend.dirtyKey)
	}

	if fake.queueSpend.deltaUSD != 1.25 {
		t.Fatalf("unexpected delta: %v", fake.queueSpend.deltaUSD)
	}

	fake.queueSpendErr = errors.New("redis down")

	if err := store.queueTokenSpendDelta(ctx, "hash-1", 1.25); err != nil {
		t.Fatalf("expected fail-open nil error, got %v", err)
	}

	store.tokenUsageFailOpen = false
	if err := store.queueTokenSpendDelta(ctx, "hash-1", 1.25); err == nil {
		t.Fatal("expected fail-closed error")
	}
}

func TestQueueKeyLastUsedUsesRedis(t *testing.T) {
	ctx := context.Background()
	fake := &fakeTokenUsageCache{}
	store := &Store{
		tokenUsageCache:     fake,
		tokenUsageEnabled:   true,
		tokenUsageFailOpen:  true,
		tokenUsageKeyPrefix: "test-usage",
		lastUsedTTL:         time.Hour,
	}
	usedAt := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)

	if err := store.queueKeyLastUsed(ctx, "key-1", usedAt); err != nil {
		t.Fatal(err)
	}

	if !fake.queueLast.called {
		t.Fatal("expected redis queue call")
	}

	if fake.queueLast.key != "test-usage:key_last_used:key-1" {
		t.Fatalf("unexpected last-used key: %s", fake.queueLast.key)
	}

	if fake.queueLast.dirtyKey != "test-usage:key_last_used:dirty" {
		t.Fatalf("unexpected dirty key: %s", fake.queueLast.dirtyKey)
	}

	if fake.queueLast.usedAt != usedAt.UnixMilli() {
		t.Fatalf("unexpected used_at unix: %d", fake.queueLast.usedAt)
	}
}
