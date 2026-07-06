package storage

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyPostgresPoolEnvUsesDefaultsForMissingOrInvalidValues(t *testing.T) {
	t.Setenv("POSTGRES_POOL_MAX_CONNS", "not-a-number")
	t.Setenv("POSTGRES_POOL_MIN_CONNS", "-1")
	t.Setenv("POSTGRES_POOL_MAX_CONN_LIFETIME_SECONDS", "0")
	t.Setenv("POSTGRES_POOL_MAX_CONN_IDLE_TIME_SECONDS", "invalid")
	t.Setenv("POSTGRES_POOL_HEALTH_CHECK_PERIOD_SECONDS", "-20")

	cfg, err := pgxpool.ParseConfig("postgres://user:pass@localhost:5432/shahrah?sslmode=disable")
	require.NoError(t, err)

	applyPostgresPoolEnv(cfg)

	assert.Equal(t, int32(32), cfg.MaxConns)
	assert.Equal(t, int32(4), cfg.MinConns)
	assert.Equal(t, 1800*time.Second, cfg.MaxConnLifetime)
	assert.Equal(t, 300*time.Second, cfg.MaxConnIdleTime)
	assert.Equal(t, 30*time.Second, cfg.HealthCheckPeriod)
}

func TestApplyPostgresPoolEnvMapsValidValues(t *testing.T) {
	t.Setenv("POSTGRES_POOL_MAX_CONNS", "64")
	t.Setenv("POSTGRES_POOL_MIN_CONNS", "8")
	t.Setenv("POSTGRES_POOL_MAX_CONN_LIFETIME_SECONDS", "900")
	t.Setenv("POSTGRES_POOL_MAX_CONN_IDLE_TIME_SECONDS", "120")
	t.Setenv("POSTGRES_POOL_HEALTH_CHECK_PERIOD_SECONDS", "15")

	cfg, err := pgxpool.ParseConfig("postgres://user:pass@localhost:5432/shahrah?sslmode=disable")
	require.NoError(t, err)

	applyPostgresPoolEnv(cfg)

	assert.Equal(t, int32(64), cfg.MaxConns)
	assert.Equal(t, int32(8), cfg.MinConns)
	assert.Equal(t, 900*time.Second, cfg.MaxConnLifetime)
	assert.Equal(t, 120*time.Second, cfg.MaxConnIdleTime)
	assert.Equal(t, 15*time.Second, cfg.HealthCheckPeriod)
}

func TestApplyPostgresPoolEnvClampsMinConnsToMaxConns(t *testing.T) {
	t.Setenv("POSTGRES_POOL_MAX_CONNS", "6")
	t.Setenv("POSTGRES_POOL_MIN_CONNS", "12")

	cfg, err := pgxpool.ParseConfig("postgres://user:pass@localhost:5432/shahrah?sslmode=disable")
	require.NoError(t, err)

	applyPostgresPoolEnv(cfg)

	assert.Equal(t, int32(6), cfg.MaxConns)
	assert.Equal(t, int32(6), cfg.MinConns)
}
