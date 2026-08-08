package config

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestResolveEnvVarsUsesSilkRoadProviderEgressURL(t *testing.T) {
	t.Setenv("SILK_ROAD_URL", "http://silk-road:8083")
	t.Setenv("TUNNEL_CLIENT_URL", "http://legacy-tunnel-client:8081")

	cfg := &Config{
		ProviderEgress: ProviderEgressSettings{URL: "${SILK_ROAD_URL}"},
	}

	err := resolveEnvVars(cfg, zap.NewNop())
	require.NoError(t, err)
	require.Equal(t, "http://silk-road:8083", cfg.ProviderEgress.URL)
}
