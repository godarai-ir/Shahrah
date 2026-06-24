package metrics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

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
