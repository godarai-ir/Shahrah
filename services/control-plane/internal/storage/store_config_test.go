package storage

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyPostgresPoolConfigUsesExplicitDefaults(t *testing.T) {
	cfg, err := pgxpool.ParseConfig("postgres://user:pass@localhost:5432/shahrah?sslmode=disable")
	require.NoError(t, err)

	applyPostgresPoolConfig(cfg, DefaultConfig().Postgres)

	assert.Equal(t, int32(32), cfg.MaxConns)
	assert.Equal(t, int32(4), cfg.MinConns)
	assert.Equal(t, 1800*time.Second, cfg.MaxConnLifetime)
	assert.Equal(t, 300*time.Second, cfg.MaxConnIdleTime)
	assert.Equal(t, 30*time.Second, cfg.HealthCheckPeriod)
}

func TestApplyPostgresPoolConfigMapsValues(t *testing.T) {
	cfg, err := pgxpool.ParseConfig("postgres://user:pass@localhost:5432/shahrah?sslmode=disable")
	require.NoError(t, err)

	applyPostgresPoolConfig(cfg, PostgresConfig{
		MaxConnections:    64,
		MinConnections:    8,
		MaxLifetime:       900 * time.Second,
		MaxIdleTime:       120 * time.Second,
		HealthCheckPeriod: 15 * time.Second,
	})

	assert.Equal(t, int32(64), cfg.MaxConns)
	assert.Equal(t, int32(8), cfg.MinConns)
	assert.Equal(t, 900*time.Second, cfg.MaxConnLifetime)
	assert.Equal(t, 120*time.Second, cfg.MaxConnIdleTime)
	assert.Equal(t, 15*time.Second, cfg.HealthCheckPeriod)
}

func TestApplyPostgresPoolConfigClampsMinConnsToMaxConns(t *testing.T) {
	cfg, err := pgxpool.ParseConfig("postgres://user:pass@localhost:5432/shahrah?sslmode=disable")
	require.NoError(t, err)

	settings := DefaultConfig().Postgres
	settings.MaxConnections = 6
	settings.MinConnections = 12
	applyPostgresPoolConfig(cfg, settings)

	assert.Equal(t, int32(6), cfg.MaxConns)
	assert.Equal(t, int32(6), cfg.MinConns)
}
