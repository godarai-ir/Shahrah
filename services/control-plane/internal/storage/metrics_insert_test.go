package storage

import (
	"strings"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/stretchr/testify/assert"
)

func TestModelMetricInsertArgsMatchPlaceholderCount(t *testing.T) {
	ttft := 25
	similarity := 0.98
	args := modelMetricInsertArgs(platform.ModelMetricEvent{
		EventID:          "019b5f1d-f9a7-7abc-8123-0123456789ab",
		TS:               time.Unix(100, 0),
		TenantID:         "tenant-1",
		RequestID:        "request-1",
		ModelName:        "gpt-4",
		RequestType:      "chat_completions",
		TeamIDs:          []string{"team-1"},
		TeamNames:        []string{"platform"},
		StatusCode:       200,
		ErrorType:        "none",
		LatencyMS:        120,
		TTFTMS:           &ttft,
		CacheSimilarity:  &similarity,
		InputTokens:      10,
		OutputTokens:     20,
		CostUSD:          0.03,
		PricingVersion:   "gateway-pricing-v1",
		EmbeddingCostUSD: 0.01,
		Metadata: map[string]interface{}{
			"environment": "test",
		},
	})

	assert.Equal(t, modelMetricInsertArgCount, strings.Count(modelMetricInsertPlaceholders, "?"))
	assert.Len(t, args, modelMetricInsertArgCount)
	assert.Equal(t, "019b5f1d-f9a7-7abc-8123-0123456789ab", args[0])
}
