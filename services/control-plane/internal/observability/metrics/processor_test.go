package metrics

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type batchExporterStub struct {
	mu      sync.Mutex
	batches [][]platform.ModelMetricEvent
}

func (s *batchExporterStub) InsertModelMetricEvents(_ context.Context, events []platform.ModelMetricEvent) error {
	copied := append([]platform.ModelMetricEvent(nil), events...)

	s.mu.Lock()
	defer s.mu.Unlock()

	s.batches = append(s.batches, copied)

	return nil
}

func TestBuildEventPreservesFirstClassCacheSavings(t *testing.T) {
	processor := &Processor{}
	similarity := 0.94
	event := processor.BuildEvent(BuildInput{
		TS: time.Now(), TenantID: "tenant-1", RequestID: "request-1", ModelName: "gpt-4",
		RequestType: "chat_completions", StatusCode: 200,
		CacheMode: "SMART", CacheStatus: "hit", CacheMatch: "semantic", CacheSimilarity: &similarity,
		CacheOriginRequestID: "origin-1", AvoidedInputTokens: 10, AvoidedOutputTokens: 5,
		AvoidedCostUSD: 0.25, AvoidedCostIRR: 212500, PricingVersion: "gateway-pricing-v1",
	})

	assert.Equal(t, "smart", event.CacheMode)
	assert.Equal(t, "hit", event.CacheStatus)
	assert.Equal(t, "semantic", event.CacheMatch)
	assert.Equal(t, 0.25, event.AvoidedCostUSD)
	assert.Equal(t, "origin-1", event.CacheOriginRequestID)
}

func TestBuildEventDefaultsCacheDimensionsToNone(t *testing.T) {
	event := (&Processor{}).BuildEvent(BuildInput{TS: time.Now(), StatusCode: 200})
	assert.Equal(t, "none", event.CacheMode)
	assert.Equal(t, "none", event.CacheStatus)
	assert.Equal(t, "none", event.CacheMatch)
}

func TestProcessorFlushExportsDequeuedEventsAsSingleBatch(t *testing.T) {
	queue := NewQueue(8)
	require.True(t, queue.Enqueue(platform.ModelMetricEvent{RequestID: "request-1"}))
	require.True(t, queue.Enqueue(platform.ModelMetricEvent{RequestID: "request-2"}))
	require.True(t, queue.Enqueue(platform.ModelMetricEvent{RequestID: "request-3"}))

	exporter := &batchExporterStub{}
	processor := NewProcessor(Config{
		BatchSize:        3,
		ExportTimeout:    time.Second,
		RetryMaxAttempts: 1,
		RetryBaseDelay:   time.Millisecond,
		RetryMaxDelay:    time.Millisecond,
	}, queue, exporter, zap.NewNop())

	processor.flush()

	require.Len(t, exporter.batches, 1)
	require.Len(t, exporter.batches[0], 3)
	assert.Equal(t, "request-1", exporter.batches[0][0].RequestID)
	assert.Equal(t, "request-3", exporter.batches[0][2].RequestID)
}
