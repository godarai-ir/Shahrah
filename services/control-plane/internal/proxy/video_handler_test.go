package proxy

import (
	"context"
	"testing"

	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
	"github.com/Delpak/shahrah/services/control-plane/internal/config"
	"github.com/Delpak/shahrah/services/control-plane/internal/router"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestVideoDeploymentResolverUsesPersistedDeployment(t *testing.T) {
	t.Parallel()

	mem, err := cache.NewMemoryCache(zap.NewNop())
	require.NoError(t, err)

	rt, err := router.NewRouter([]config.ModelDeployment{
		{
			ModelName: "video-model",
			ProviderEndpoint: config.ProviderEndpointConfig{
				Provider: "openai",
				Model:    "video-model",
				APIKey:   "key-a",
			},
			ModelInfo: map[string]interface{}{"id": "deployment-a"},
		},
		{
			ModelName: "video-model",
			ProviderEndpoint: config.ProviderEndpointConfig{
				Provider: "openai",
				Model:    "video-model",
				APIKey:   "key-b",
			},
			ModelInfo: map[string]interface{}{"id": "deployment-b"},
		},
	}, config.RouterSettings{CooldownTime: 60, NumRetries: 1}, mem, nil, zap.NewNop())
	require.NoError(t, err)

	rt.MarkFailed(context.Background(), &router.Deployment{ID: "deployment-a"}, "test cooldown")

	deployment, err := NewVideoDeploymentResolver(rt)(
		context.Background(),
		"tenant-a",
		"video-model",
		"deployment-a",
	)
	require.NoError(t, err)
	require.Equal(t, "deployment-a", deployment.ID)
	require.Equal(t, "key-a", deployment.APIKey)
}
