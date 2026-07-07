package storage

import (
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/stretchr/testify/assert"
)

func TestApproximateWindowedTotalSupportsLimitPlusOnePagination(t *testing.T) {
	assert.Equal(t, 50, approximateWindowedTotal(0, 50, false))
	assert.Equal(t, 51, approximateWindowedTotal(0, 50, true))
	assert.Equal(t, 151, approximateWindowedTotal(100, 50, true))
}

func TestBuildTraceProviderSpanWhereIsTimeBounded(t *testing.T) {
	from := time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC)
	to := time.Date(2026, 7, 8, 11, 0, 0, 0, time.UTC)

	where, args := buildTraceProviderSpanWhere(platform.ListRequestTracesV2Params{
		TenantID: "tenant-1",
		From:     &from,
		To:       &to,
	})

	assert.Contains(t, where, "tenant_id = ?")
	assert.Contains(t, where, "name = 'provider.call'")
	assert.Contains(t, where, "started_at >= ?")
	assert.Contains(t, where, "started_at <= ?")
	assert.Equal(t, []interface{}{"tenant-1", from, to}, args)
}
