package virtualmodels_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

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

func (r repo) GetVirtualModelByAlias(ctx context.Context, tenantID, alias string) (*platform.VirtualModel, error) {
	if r.err != nil {
		return nil, r.err
	}

	return r.vm, nil
}

type countingRepo struct {
	mu      sync.Mutex
	calls   int
	aliases []string
	vms     []*platform.VirtualModel
	errs    []error
}

func (r *countingRepo) GetVirtualModelByAlias(ctx context.Context, tenantID, alias string) (*platform.VirtualModel, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.calls++
	r.aliases = append(r.aliases, alias)
	index := r.calls - 1
	if index < len(r.errs) && r.errs[index] != nil {
		return nil, r.errs[index]
	}
	if index < len(r.vms) {
		return r.vms[index], nil
	}

	return nil, nil
}

func (r *countingRepo) callCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.calls
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

func TestResolverCachesVirtualModelAliasHits(t *testing.T) {
	repo := &countingRepo{vms: []*platform.VirtualModel{{
		Alias:        "gpt-4o-smart",
		Enabled:      true,
		ExposedModes: []string{"chat"},
		Targets: []platform.VirtualModelTarget{{
			Provider: "openai",
			Model:    "gpt-4o",
			Policy: platform.VirtualModelPolicy{
				Retry:    platform.VirtualModelRetryPolicy{StatusCodes: []int{429}},
				Fallback: platform.VirtualModelFallbackPolicy{StatusCodes: []int{500}},
			},
		}},
	}}}
	resolver := virtualmodels.NewResolver(repo, virtualmodels.WithAliasCacheTTLs(time.Minute, time.Minute))

	first, err := resolver.Resolve(context.Background(), "tenant-1", "gpt-4o-smart")
	require.NoError(t, err)
	first.Targets[0].Policy.Retry.StatusCodes[0] = 400

	second, err := resolver.Resolve(context.Background(), "tenant-1", "gpt-4o-smart")
	require.NoError(t, err)

	assert.Equal(t, 1, repo.callCount())
	assert.Equal(t, []int{429}, second.Targets[0].Policy.Retry.StatusCodes)
}

func TestResolverCachesVirtualModelAliasMisses(t *testing.T) {
	notFoundErr := errors.New("virtual model not found")
	repo := &countingRepo{errs: []error{notFoundErr}}
	resolver := virtualmodels.NewResolver(repo,
		virtualmodels.WithAliasCacheTTLs(time.Minute, time.Minute),
		virtualmodels.WithNotFoundError(func(err error) bool { return errors.Is(err, notFoundErr) }),
	)

	_, err := resolver.Resolve(context.Background(), "tenant-1", "static-model")
	require.ErrorIs(t, err, virtualmodels.ErrNotVirtualModel)
	_, err = resolver.Resolve(context.Background(), "tenant-1", "static-model")
	require.ErrorIs(t, err, virtualmodels.ErrNotVirtualModel)

	assert.Equal(t, 1, repo.callCount())
}

func TestResolverDoesNotCacheTransientErrorsAsMisses(t *testing.T) {
	transientErr := errors.New("database unavailable")
	repo := &countingRepo{
		errs: []error{transientErr},
		vms:  []*platform.VirtualModel{nil, {Alias: "vm", Enabled: true}},
	}
	resolver := virtualmodels.NewResolver(repo,
		virtualmodels.WithAliasCacheTTLs(time.Minute, time.Minute),
		virtualmodels.WithNotFoundError(func(err error) bool { return false }),
	)

	_, err := resolver.Resolve(context.Background(), "tenant-1", "vm")
	require.ErrorIs(t, err, transientErr)
	vm, err := resolver.Resolve(context.Background(), "tenant-1", "vm")
	require.NoError(t, err)

	assert.Equal(t, "vm", vm.Alias)
	assert.Equal(t, 2, repo.callCount())
}

func TestResolverNormalizesAliasBeforeCaching(t *testing.T) {
	repo := &countingRepo{vms: []*platform.VirtualModel{{Alias: "vm", Enabled: true}}}
	resolver := virtualmodels.NewResolver(repo, virtualmodels.WithAliasCacheTTLs(time.Minute, time.Minute))

	_, err := resolver.Resolve(context.Background(), "tenant-1", "  vm  ")
	require.NoError(t, err)
	_, err = resolver.Resolve(context.Background(), "tenant-1", "vm")
	require.NoError(t, err)

	assert.Equal(t, 1, repo.callCount())
	assert.Equal(t, []string{"vm"}, repo.aliases)
}

func TestResolverAliasCacheHonorsTTL(t *testing.T) {
	repo := &countingRepo{vms: []*platform.VirtualModel{
		{Alias: "vm", Enabled: true, Version: 1},
		{Alias: "vm", Enabled: true, Version: 2},
	}}
	resolver := virtualmodels.NewResolver(repo, virtualmodels.WithAliasCacheTTLs(10*time.Millisecond, time.Minute))

	first, err := resolver.Resolve(context.Background(), "tenant-1", "vm")
	require.NoError(t, err)
	time.Sleep(20 * time.Millisecond)
	second, err := resolver.Resolve(context.Background(), "tenant-1", "vm")
	require.NoError(t, err)

	assert.Equal(t, 1, first.Version)
	assert.Equal(t, 2, second.Version)
	assert.Equal(t, 2, repo.callCount())
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

func TestPriorityRoutingEngineSortsByOrder(t *testing.T) {
	engine := virtualmodels.NewPriorityRoutingEngine()
	targets, err := engine.SelectTargets(&platform.VirtualModel{
		RoutingStrategy: "priority",
		Targets: []platform.VirtualModelTarget{
			{Order: 3, Provider: "openai", Model: "gpt-4o"},
			{Order: 1, Provider: "anthropic", Model: "claude-sonnet"},
			{Order: 2, Provider: "google", Model: "gemini"},
		},
	})
	require.NoError(t, err)
	require.Len(t, targets, 3)
	assert.Equal(t, 1, targets[0].Order)
	assert.Equal(t, 2, targets[1].Order)
	assert.Equal(t, 3, targets[2].Order)
}

func TestWeightedRoutingEngineSelectsWeightedPrimaryThenOrderedFallbacks(t *testing.T) {
	engine := virtualmodels.NewWeightedRoutingEngine()
	targets, err := engine.SelectTargets(&platform.VirtualModel{
		RoutingStrategy: "weighted",
		Targets: []platform.VirtualModelTarget{
			{
				Order:    2,
				Weight:   100,
				Provider: "openai",
				Model:    "gpt-4o",
				Policy: platform.VirtualModelPolicy{
					Fallback: platform.VirtualModelFallbackPolicy{Candidate: true},
				},
			},
			{
				Order:    1,
				Weight:   50,
				Provider: "anthropic",
				Model:    "claude-sonnet",
				Policy: platform.VirtualModelPolicy{
					Fallback: platform.VirtualModelFallbackPolicy{Candidate: true},
				},
			},
			{
				Order:    3,
				Weight:   20,
				Provider: "google",
				Model:    "gemini",
				Policy: platform.VirtualModelPolicy{
					Fallback: platform.VirtualModelFallbackPolicy{Candidate: false},
				},
			},
		},
	})
	require.NoError(t, err)
	require.Len(t, targets, 2)
	assert.Equal(t, "openai", targets[0].Provider)
	assert.Equal(t, "anthropic", targets[1].Provider)
}

func TestRoutingEngineUsesWeightedStrategy(t *testing.T) {
	engine := virtualmodels.NewRoutingEngine()
	highCount := 0
	iterations := 1000

	for i := 0; i < iterations; i++ {
		targets, err := engine.SelectTargets(&platform.VirtualModel{
			RoutingStrategy: "weighted",
			Targets: []platform.VirtualModelTarget{
				{Order: 1, Weight: 90, Provider: "openai", Model: "gpt-4o", Policy: platform.VirtualModelPolicy{Fallback: platform.VirtualModelFallbackPolicy{Candidate: true}}},
				{Order: 2, Weight: 10, Provider: "anthropic", Model: "claude-sonnet", Policy: platform.VirtualModelPolicy{Fallback: platform.VirtualModelFallbackPolicy{Candidate: true}}},
			},
		})
		require.NoError(t, err)
		require.NotEmpty(t, targets)
		if targets[0].Provider == "openai" {
			highCount++
		}
	}

	assert.Greater(t, float64(highCount)/float64(iterations), 0.80)
}
