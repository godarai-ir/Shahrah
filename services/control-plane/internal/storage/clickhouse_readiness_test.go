package storage

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClickHouseReadinessReportSeparatesSubsystems(t *testing.T) {
	db, err := sql.Open("clickhouse", "clickhouse://127.0.0.1:9000")
	require.NoError(t, err)

	defer closeBestEffort(db)

	store := &Store{
		clickhouse:             db,
		clickhouseReady:        true,
		clickhouseTraceReady:   false,
		clickhouseMetricsReady: true,
		clickhouseLogsReady:    true,
	}

	report := store.ClickHouseReadinessReport()

	assert.Equal(t, true, report["connected"])
	assert.Equal(t, false, report["traces"])
	assert.Equal(t, true, report["metrics"])
	assert.Equal(t, true, report["logs"])
	assert.False(t, store.clickHouseTraceReady())
	assert.True(t, store.clickHouseMetricsReady())
	assert.True(t, store.clickHouseLogsReady())
}

func TestClickHouseReadinessReportNilStore(t *testing.T) {
	var store *Store

	report := store.ClickHouseReadinessReport()

	assert.Equal(t, false, report["connected"])
	assert.Equal(t, false, report["traces"])
	assert.Equal(t, false, report["metrics"])
	assert.Equal(t, false, report["logs"])
}
