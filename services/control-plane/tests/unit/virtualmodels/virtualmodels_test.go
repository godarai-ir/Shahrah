package virtualmodels_test

import (
	"context"
	"errors"
	"net/http"
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

func TestResolverCachesNilVirtualModelAliasMisses(t *testing.T) {
	repo := &countingRepo{}
	resolver := virtualmodels.NewResolver(
		repo,
		virtualmodels.WithAliasCacheTTLs(time.Minute, time.Minute),
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

func TestExecutorSyncRetryDelayHonorsContextCancellation(t *testing.T) {
	exec := virtualmodels.NewExecutor(virtualmodels.NewDefaultErrorNormalizer())
	targets := []platform.VirtualModelTarget{{
		Provider: "openai",
		Model:    "gpt-4o",
		Policy: platform.VirtualModelPolicy{
			Retry: platform.VirtualModelRetryPolicy{
				Attempts: 2,
				DelayMS:  10_000,
			},
		},
	}}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	attempts := 0

	_, _, _, _, err := exec.ExecuteSync(
		ctx,
		targets,
		func(context.Context, *router.Deployment) (*platform.TunnelResponse, error) {
			attempts++

			return nil, errors.New("provider unavailable")
		},
	)

	require.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, 1, attempts)
}

func TestExecutorSyncSkipsRetriesForMarkedError(t *testing.T) {
	exec := virtualmodels.NewExecutor(virtualmodels.NewDefaultErrorNormalizer())
	targets := []platform.VirtualModelTarget{
		{
			Provider: "openai",
			Model:    "gpt-4o",
			Policy: platform.VirtualModelPolicy{
				Retry:    platform.VirtualModelRetryPolicy{Attempts: 3},
				Fallback: platform.VirtualModelFallbackPolicy{Candidate: true},
			},
		},
		{
			Provider: "anthropic",
			Model:    "claude-sonnet",
			Policy: platform.VirtualModelPolicy{
				Retry:    platform.VirtualModelRetryPolicy{Attempts: 1},
				Fallback: platform.VirtualModelFallbackPolicy{Candidate: true},
			},
		},
	}
	firstTargetAttempts := 0

	response, deployment, fallbackCount, _, err := exec.ExecuteSync(
		context.Background(),
		targets,
		func(_ context.Context, deployment *router.Deployment) (*platform.TunnelResponse, error) {
			if deployment.Provider == "openai" {
				firstTargetAttempts++

				return nil, virtualmodels.WithoutRetry(errors.New("target preparation failed"))
			}

			return &platform.TunnelResponse{StatusCode: http.StatusOK}, nil
		},
	)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.Equal(t, "anthropic", deployment.Provider)
	assert.Equal(t, 1, fallbackCount)
	assert.Equal(t, 1, firstTargetAttempts)
}

func TestExecutorSyncFallsBackAfterEmptyResponse(t *testing.T) {
	exec := virtualmodels.NewExecutor(virtualmodels.NewDefaultErrorNormalizer())
	targets := []platform.VirtualModelTarget{
		{
			Provider: "openai",
			Model:    "gpt-4o",
			Policy: platform.VirtualModelPolicy{
				Retry:    platform.VirtualModelRetryPolicy{Attempts: 1},
				Fallback: platform.VirtualModelFallbackPolicy{Candidate: true},
			},
		},
		{
			Provider: "anthropic",
			Model:    "claude-sonnet",
			Policy: platform.VirtualModelPolicy{
				Retry:    platform.VirtualModelRetryPolicy{Attempts: 1},
				Fallback: platform.VirtualModelFallbackPolicy{Candidate: true},
			},
		},
	}
	attempts := 0

	response, deployment, fallbackCount, reason, err := exec.ExecuteSync(
		context.Background(),
		targets,
		func(context.Context, *router.Deployment) (*platform.TunnelResponse, error) {
			attempts++
			if attempts == 1 {
				return nil, nil
			}

			return &platform.TunnelResponse{StatusCode: 200}, nil
		},
	)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, 200, response.StatusCode)
	assert.Equal(t, "anthropic", deployment.Provider)
	assert.Equal(t, 1, fallbackCount)
	assert.Equal(t, "provider returned an empty response", reason)
}

func TestExecutorSyncDoesNotCountSkippedFallbackCandidates(t *testing.T) {
	exec := virtualmodels.NewExecutor(virtualmodels.NewDefaultErrorNormalizer())
	targets := []platform.VirtualModelTarget{
		{
			Provider: "openai",
			Model:    "gpt-4o",
			Policy: platform.VirtualModelPolicy{
				Retry:    platform.VirtualModelRetryPolicy{Attempts: 1},
				Fallback: platform.VirtualModelFallbackPolicy{Candidate: true},
			},
		},
		{
			Provider: "anthropic",
			Model:    "claude-sonnet",
			Policy: platform.VirtualModelPolicy{
				Retry:    platform.VirtualModelRetryPolicy{Attempts: 1},
				Fallback: platform.VirtualModelFallbackPolicy{Candidate: false},
			},
		},
	}

	_, _, fallbackCount, _, err := exec.ExecuteSync(
		context.Background(),
		targets,
		func(context.Context, *router.Deployment) (*platform.TunnelResponse, error) {
			return nil, errors.New("provider unavailable")
		},
	)

	require.ErrorIs(t, err, virtualmodels.ErrFallbackExhausted)
	assert.Zero(t, fallbackCount)
}

func TestExecutorStreamFallsBackBeforeCommitAndClosesFailedStream(t *testing.T) {
	exec := virtualmodels.NewExecutor(virtualmodels.NewDefaultErrorNormalizer())
	targets := []platform.VirtualModelTarget{
		{
			Order:    1,
			Provider: "openai",
			Model:    "gpt-4o",
			Policy: platform.VirtualModelPolicy{
				Retry:    platform.VirtualModelRetryPolicy{Attempts: 1, StatusCodes: []int{429}},
				Fallback: platform.VirtualModelFallbackPolicy{StatusCodes: []int{429}, Candidate: true},
			},
		},
		{
			Order:    2,
			Provider: "anthropic",
			Model:    "claude-sonnet",
			Policy: platform.VirtualModelPolicy{
				Retry:    platform.VirtualModelRetryPolicy{Attempts: 1, StatusCodes: []int{429}},
				Fallback: platform.VirtualModelFallbackPolicy{StatusCodes: []int{429}, Candidate: true},
			},
		},
	}

	failed := newExecutorTestStream(429)
	succeeded := newExecutorTestStream(200)
	attempt := 0

	stream, deployment, fallbackCount, _, err := exec.ExecuteStream(
		context.Background(),
		targets,
		func(context.Context, *router.Deployment) (platform.TunnelStream, error) {
			attempt++
			if attempt == 1 {
				return failed, nil
			}

			return succeeded, nil
		},
	)
	require.NoError(t, err)
	assert.Same(t, succeeded, stream)
	assert.Equal(t, "anthropic", deployment.Provider)
	assert.Equal(t, 1, fallbackCount)
	assert.True(t, failed.closed)
	assert.False(t, succeeded.closed)
}

func TestExecutorStreamSkipsRetriesForMarkedError(t *testing.T) {
	exec := virtualmodels.NewExecutor(virtualmodels.NewDefaultErrorNormalizer())
	targets := []platform.VirtualModelTarget{
		{
			Provider: "openai",
			Model:    "gpt-4o",
			Policy: platform.VirtualModelPolicy{
				Retry:    platform.VirtualModelRetryPolicy{Attempts: 3},
				Fallback: platform.VirtualModelFallbackPolicy{Candidate: true},
			},
		},
		{
			Provider: "anthropic",
			Model:    "claude-sonnet",
			Policy: platform.VirtualModelPolicy{
				Retry:    platform.VirtualModelRetryPolicy{Attempts: 1},
				Fallback: platform.VirtualModelFallbackPolicy{Candidate: true},
			},
		},
	}
	firstTargetAttempts := 0
	want := newExecutorTestStream(http.StatusOK)

	stream, deployment, fallbackCount, _, err := exec.ExecuteStream(
		context.Background(),
		targets,
		func(_ context.Context, deployment *router.Deployment) (platform.TunnelStream, error) {
			if deployment.Provider == "openai" {
				firstTargetAttempts++

				return nil, virtualmodels.WithoutRetry(errors.New("target preparation failed"))
			}

			return want, nil
		},
	)

	require.NoError(t, err)
	assert.Same(t, want, stream)
	assert.Equal(t, "anthropic", deployment.Provider)
	assert.Equal(t, 1, fallbackCount)
	assert.Equal(t, 1, firstTargetAttempts)
}

func TestExecutorStreamReturnsNonFallbackHTTPErrorToCaller(t *testing.T) {
	exec := virtualmodels.NewExecutor(virtualmodels.NewDefaultErrorNormalizer())
	targets := []platform.VirtualModelTarget{{
		Provider: "openai",
		Model:    "gpt-4o",
		Policy: platform.VirtualModelPolicy{
			Fallback: platform.VirtualModelFallbackPolicy{StatusCodes: []int{429}, Candidate: true},
		},
	}}
	want := newExecutorTestStream(400)

	stream, deployment, fallbackCount, reason, err := exec.ExecuteStream(
		context.Background(),
		targets,
		func(context.Context, *router.Deployment) (platform.TunnelStream, error) {
			return want, nil
		},
	)
	require.NoError(t, err)
	assert.Same(t, want, stream)
	assert.Equal(t, "openai", deployment.Provider)
	assert.Zero(t, fallbackCount)
	assert.Equal(t, "status_error", reason)
	assert.False(t, want.closed)
}

type executorTestStream struct {
	statusCode int
	chunks     chan platform.TunnelChunk
	errors     chan error
	closed     bool
}

func newExecutorTestStream(statusCode int) *executorTestStream {
	return &executorTestStream{
		statusCode: statusCode,
		chunks:     make(chan platform.TunnelChunk),
		errors:     make(chan error),
	}
}

func (s *executorTestStream) Chunks() <-chan platform.TunnelChunk { return s.chunks }
func (s *executorTestStream) Errors() <-chan error                { return s.errors }
func (s *executorTestStream) Close() error {
	s.closed = true

	return nil
}
func (s *executorTestStream) Headers() map[string]string { return nil }
func (s *executorTestStream) StatusCode() int            { return s.statusCode }

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
	require.NotEmpty(t, targets)
	assert.Contains(t, []string{"anthropic", "openai", "google"}, targets[0].Provider)

	expectedFallbacks := []string{"anthropic", "openai"}

	switch targets[0].Provider {
	case "anthropic":
		expectedFallbacks = []string{"openai"}
	case "openai":
		expectedFallbacks = []string{"anthropic"}
	}

	actualFallbacks := make([]string, 0, len(targets)-1)
	for _, target := range targets[1:] {
		assert.True(t, target.Policy.Fallback.Candidate)
		actualFallbacks = append(actualFallbacks, target.Provider)
	}

	assert.Equal(t, expectedFallbacks, actualFallbacks)
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
