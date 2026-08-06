package storage

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
)

type fakeUsageCounterCache struct {
	result *cache.UsageCounterResult
	err    error
	req    cache.UsageCounterRequest
}

func (f *fakeUsageCounterCache) CheckUsageCounter(_ context.Context, req cache.UsageCounterRequest) (*cache.UsageCounterResult, error) {
	f.req = req
	if f.err != nil {
		return nil, f.err
	}

	return f.result, nil
}

func (f *fakeUsageCounterCache) ListDirtyUsageCounters(context.Context, string, int) ([]cache.UsageCounterSnapshot, error) {
	return nil, nil
}

func (f *fakeUsageCounterCache) AckDirtyUsageCounters(context.Context, string, []string) {
}

func TestUsageCounterMonthWindowUsesUTC(t *testing.T) {
	now := time.Date(2026, 7, 15, 12, 30, 0, 0, time.FixedZone("IRST", 3*60*60+30*60))

	start, end := usageCounterMonthWindow(now)

	if want := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC); !start.Equal(want) {
		t.Fatalf("unexpected window start: got %s want %s", start, want)
	}

	if want := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC); !end.Equal(want) {
		t.Fatalf("unexpected window end: got %s want %s", end, want)
	}
}

func TestUsageCounterRedisPathBuildsKeyAndCounter(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 7, 5, 10, 0, 0, 0, time.UTC)
	fake := &fakeUsageCounterCache{result: &cache.UsageCounterResult{Allowed: true, Used: 7}}
	store := &PlanRepository{
		usageCounterCache:     fake,
		usageCountersEnabled:  true,
		usageCounterKeyPrefix: "test-usage",
		usageCounterTTLBuffer: time.Hour,
	}

	counter, allowed, err := store.CheckAndIncrementTenantUsageCounter(ctx, "tenant-1", "gateway.requests.monthly", 10, 1, now)
	if err != nil {
		t.Fatal(err)
	}

	if !allowed {
		t.Fatal("expected request to be allowed")
	}

	if counter.Used != 7 {
		t.Fatalf("unexpected counter used: %v", counter.Used)
	}

	if fake.req.Key != "test-usage:tenant:tenant-1:feature:gateway.requests.monthly:window:1782864000" {
		t.Fatalf("unexpected redis key: %s", fake.req.Key)
	}

	if fake.req.DirtySetKey != "test-usage:dirty" {
		t.Fatalf("unexpected dirty key: %s", fake.req.DirtySetKey)
	}

	if fake.req.TTL <= time.Hour {
		t.Fatalf("expected ttl to include window remainder and buffer, got %s", fake.req.TTL)
	}
}

func TestUsageCounterRedisFailOpen(t *testing.T) {
	ctx := context.Background()
	store := &PlanRepository{
		usageCounterCache:     &fakeUsageCounterCache{err: errors.New("redis down")},
		usageCountersEnabled:  true,
		usageCountersFailOpen: true,
		usageCounterKeyPrefix: "test-usage",
		usageCounterTTLBuffer: time.Hour,
	}

	counter, allowed, err := store.CheckAndIncrementTenantUsageCounter(ctx, "tenant-1", "gateway.requests.monthly", 10, 1, time.Date(2026, 7, 5, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}

	if !allowed {
		t.Fatal("expected fail-open request to be allowed")
	}

	if counter.Used != 0 {
		t.Fatalf("unexpected fail-open counter used: %v", counter.Used)
	}
}

func TestUsageCounterKeyRoundTrip(t *testing.T) {
	store := &PlanRepository{usageCounterKeyPrefix: "usage"}
	windowStart := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	key := store.usageCounterKey("tenant-1", "gateway.requests.monthly", windowStart)

	parts, err := store.parseUsageCounterKey(key)
	if err != nil {
		t.Fatal(err)
	}

	if parts.TenantID != "tenant-1" || parts.FeatureKey != "gateway.requests.monthly" || !parts.WindowStart.Equal(windowStart) {
		t.Fatalf("unexpected parsed key parts: %+v", parts)
	}
}
