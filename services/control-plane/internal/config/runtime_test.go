package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLoadStorageSettingsUsesDefaultsForInvalidValues(t *testing.T) {
	t.Setenv("POSTGRES_POOL_MAX_CONNS", "invalid")
	t.Setenv("POSTGRES_POOL_MIN_CONNS", "-1")
	t.Setenv("GATEWAY_DB_TIMEOUT_SECONDS", "0")

	settings := LoadStorageSettings()

	require.Equal(t, int32(32), settings.PostgresMaxConnections)
	require.Equal(t, int32(4), settings.PostgresMinConnections)
	require.Equal(t, 3*time.Second, settings.GatewayDatabaseTimeout)
}

func TestLoadStorageSettingsClampsPostgresMinimum(t *testing.T) {
	t.Setenv("POSTGRES_POOL_MAX_CONNS", "6")
	t.Setenv("POSTGRES_POOL_MIN_CONNS", "12")

	settings := LoadStorageSettings()

	require.Equal(t, int32(6), settings.PostgresMaxConnections)
	require.Equal(t, int32(6), settings.PostgresMinConnections)
}

func TestLoadRuntimeSettingsParsesCompositionValues(t *testing.T) {
	t.Setenv("REDIS_SENTINEL_ADDRS", " redis-a:26379,redis-b:26379 ")
	t.Setenv("ROUTER_DEPLOYMENT_CACHE_TTL_SECONDS", "45")
	t.Setenv("PLATFORM_PROVIDER_CREDENTIAL_ENVIRONMENT", "staging")
	t.Setenv("VIRTUAL_MODEL_ALIAS_CACHE_TTL_SECONDS", "75")
	t.Setenv("VIRTUAL_MODEL_ALIAS_NEGATIVE_CACHE_TTL_SECONDS", "12")
	t.Setenv("NOBITEX_EXCHANGE_RATE_ENABLED", "false")

	settings := LoadRuntimeSettings()

	require.Equal(t, []string{"redis-a:26379", "redis-b:26379"}, settings.RedisSentinelAddresses)
	require.Equal(t, 45*time.Second, settings.RouterDeploymentCacheTTL)
	require.Equal(t, "staging", settings.PlatformCredentialEnv)
	require.Equal(t, 75*time.Second, settings.VirtualModelAliasCacheTTL)
	require.Equal(t, 12*time.Second, settings.VirtualModelNegativeTTL)
	require.True(t, settings.NobitexExchangeRateConfigured)
	require.False(t, settings.NobitexExchangeRateEnabled)
}
