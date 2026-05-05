package virtualmodels_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/router"
	"github.com/Delpak/shahrah/services/control-plane/internal/virtualmodels"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type repo struct {
	vm  *platform.VirtualModel
	err error
}

func (r repo) GetVirtualModelByAlias(ctx context.Context, tenantID string, alias string) (*platform.VirtualModel, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.vm, nil
}

func TestResolverResolveSuccess(t *testing.T) {
	resolver := virtualmodels.NewResolver(repo{
		vm: &platform.VirtualModel{Alias: "gpt-4o-smart", Enabled: true},
	})
	vm, err := resolver.Resolve(context.Background(), "t1", "gpt-4o-smart")
	require.NoError(t, err)
	assert.Equal(t, "gpt-4o-smart", vm.Alias)
}

func TestResolverUnknownVirtualModel(t *testing.T) {
	resolver := virtualmodels.NewResolver(repo{err: errors.New("not found")})
	_, err := resolver.Resolve(context.Background(), "t1", "missing")
	require.Error(t, err)
}

func TestResolverDisabledVirtualModel(t *testing.T) {
	resolver := virtualmodels.NewResolver(repo{vm: &platform.VirtualModel{Alias: "x", Enabled: false}})
	_, err := resolver.Resolve(context.Background(), "t1", "x")
	assert.ErrorIs(t, err, virtualmodels.ErrDisabledVirtualModel)
}

func TestExecutorFallbackOn429(t *testing.T) {
	exec := virtualmodels.NewExecutor(virtualmodels.NewDefaultErrorNormalizer())
	targets := []platform.VirtualModelTarget{
		{
			Order:    1,
			Provider: "openai",
			Model:    "gpt-4o",
			Policy: platform.VirtualModelPolicy{
				Retry:    platform.VirtualModelRetryPolicy{Attempts: 1, StatusCodes: []int{429, 500}},
				Fallback: platform.VirtualModelFallbackPolicy{StatusCodes: []int{429, 500}, Candidate: true},
			},
		},
		{
			Order:    2,
			Provider: "anthropic",
			Model:    "claude-sonnet",
			Policy: platform.VirtualModelPolicy{
				Retry:    platform.VirtualModelRetryPolicy{Attempts: 1, StatusCodes: []int{429, 500}},
				Fallback: platform.VirtualModelFallbackPolicy{StatusCodes: []int{429, 500}, Candidate: true},
			},
		},
	}
	attempt := 0
	resp, dep, fallbackCount, _, err := exec.ExecuteSync(context.Background(), targets, func(ctx context.Context, dep *router.Deployment) (*platform.TunnelResponse, error) {
		attempt++
		if attempt == 1 {
			return &platform.TunnelResponse{StatusCode: 429}, nil
		}
		return &platform.TunnelResponse{StatusCode: 200, Body: []byte(`{}`)}, nil
	})
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "anthropic", dep.Provider)
	assert.Equal(t, 1, fallbackCount)
}

func TestExecutorNonRetryableStops(t *testing.T) {
	exec := virtualmodels.NewExecutor(virtualmodels.NewDefaultErrorNormalizer())
	targets := []platform.VirtualModelTarget{
		{
			Order:    1,
			Provider: "openai",
			Model:    "gpt-4o",
			Policy: platform.VirtualModelPolicy{
				Retry:    platform.VirtualModelRetryPolicy{Attempts: 1, StatusCodes: []int{429, 500}},
				Fallback: platform.VirtualModelFallbackPolicy{StatusCodes: []int{429, 500}, Candidate: true},
			},
		},
	}
	resp, _, _, _, err := exec.ExecuteSync(context.Background(), targets, func(ctx context.Context, dep *router.Deployment) (*platform.TunnelResponse, error) {
		return &platform.TunnelResponse{StatusCode: 400}, nil
	})
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}
