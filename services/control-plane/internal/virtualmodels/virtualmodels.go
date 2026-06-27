package virtualmodels

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/config"
	"github.com/Delpak/shahrah/services/control-plane/internal/modelid"
	"github.com/Delpak/shahrah/services/control-plane/internal/router"
)

var (
	ErrNotVirtualModel      = errors.New("not a virtual model")
	ErrDisabledVirtualModel = errors.New("virtual model disabled")
	ErrNoEligibleTargets    = errors.New("no eligible virtual model targets")
	ErrFallbackExhausted    = errors.New("virtual model fallback exhausted")
)

type VirtualModelRepository interface {
	GetVirtualModelByAlias(ctx context.Context, tenantID string, alias string) (*platform.VirtualModel, error)
}

type VirtualModelResolver interface {
	Resolve(ctx context.Context, tenantID, alias string) (*platform.VirtualModel, error)
}

type RoutingEngine interface {
	SelectTargets(vm *platform.VirtualModel) ([]platform.VirtualModelTarget, error)
}

type ErrorNormalizer interface {
	ShouldRetryOnStatus(status int, target platform.VirtualModelTarget) bool
	ShouldFallbackOnStatus(status int, target platform.VirtualModelTarget) bool
}

type FallbackExecutor interface {
	ExecuteSync(ctx context.Context, targets []platform.VirtualModelTarget, doCall func(ctx context.Context, dep *router.Deployment) (*platform.TunnelResponse, error)) (*platform.TunnelResponse, *router.Deployment, int, string, error)
}

type Resolver struct {
	repo VirtualModelRepository
}

func NewResolver(repo VirtualModelRepository) *Resolver {
	return &Resolver{repo: repo}
}

func (r *Resolver) Resolve(ctx context.Context, tenantID, alias string) (*platform.VirtualModel, error) {
	vm, err := r.repo.GetVirtualModelByAlias(ctx, tenantID, strings.TrimSpace(alias))
	if err != nil {
		return nil, err
	}

	if !vm.Enabled {
		return nil, ErrDisabledVirtualModel
	}

	return vm, nil
}

type PriorityRoutingEngine struct{}

func NewPriorityRoutingEngine() *PriorityRoutingEngine {
	return &PriorityRoutingEngine{}
}

func (e *PriorityRoutingEngine) SelectTargets(vm *platform.VirtualModel) ([]platform.VirtualModelTarget, error) {
	if vm == nil || len(vm.Targets) == 0 {
		return nil, ErrNoEligibleTargets
	}
	// Stored order is already deterministic.
	out := make([]platform.VirtualModelTarget, 0, len(vm.Targets))
	for _, t := range vm.Targets {
		if strings.TrimSpace(t.Provider) == "" || strings.TrimSpace(t.Model) == "" {
			continue
		}

		out = append(out, t)
	}

	if len(out) == 0 {
		return nil, ErrNoEligibleTargets
	}

	return out, nil
}

type DefaultErrorNormalizer struct{}

func NewDefaultErrorNormalizer() *DefaultErrorNormalizer {
	return &DefaultErrorNormalizer{}
}

func (n *DefaultErrorNormalizer) ShouldRetryOnStatus(status int, target platform.VirtualModelTarget) bool {
	return containsInt(target.Policy.Retry.StatusCodes, status)
}

func (n *DefaultErrorNormalizer) ShouldFallbackOnStatus(status int, target platform.VirtualModelTarget) bool {
	return containsInt(target.Policy.Fallback.StatusCodes, status)
}

type Executor struct {
	normalizer ErrorNormalizer
}

func NewExecutor(normalizer ErrorNormalizer) *Executor {
	return &Executor{normalizer: normalizer}
}

func (e *Executor) ExecuteSync(ctx context.Context, targets []platform.VirtualModelTarget, doCall func(ctx context.Context, dep *router.Deployment) (*platform.TunnelResponse, error)) (*platform.TunnelResponse, *router.Deployment, int, string, error) {
	fallbackCount := 0
	lastReason := ""

	for idx, target := range targets {
		if idx > 0 {
			fallbackCount++
		}

		if !target.Policy.Fallback.Candidate && idx > 0 {
			continue
		}

		retries := target.Policy.Retry.Attempts
		if retries <= 0 {
			retries = 1
		}

		delay := time.Duration(target.Policy.Retry.DelayMS) * time.Millisecond

		for attempt := 1; attempt <= retries; attempt++ {
			provider := strings.TrimSpace(target.Provider)

			modelForProvider := strings.TrimSpace(target.Model)
			if p, _, official, ok := splitModelFQN(target.Model); ok {
				if provider == "" {
					provider = p
				}

				modelForProvider = official
			}

			dep := &router.Deployment{
				ID:        target.ID,
				ModelName: target.Model,
				Provider:  provider,
				Model:     modelForProvider,
				Params: config.ProviderEndpointConfig{
					Timeout: target.Policy.TimeoutSec,
				},
				Healthy: true,
			}

			resp, err := doCall(ctx, dep)
			if err != nil {
				lastReason = err.Error()

				if attempt < retries {
					if delay > 0 {
						time.Sleep(delay)
					}

					continue
				}

				break
			}

			if resp.StatusCode < 400 {
				return resp, dep, fallbackCount, lastReason, nil
			}

			if e.normalizer.ShouldRetryOnStatus(resp.StatusCode, target) && attempt < retries {
				if delay > 0 {
					time.Sleep(delay)
				}

				continue
			}

			lastReason = "status_error"
			if !e.normalizer.ShouldFallbackOnStatus(resp.StatusCode, target) {
				return resp, dep, fallbackCount, lastReason, nil
			}

			break
		}
	}

	return nil, nil, fallbackCount, lastReason, ErrFallbackExhausted
}

func splitModelFQN(input string) (provider string, accountAlias string, officialModel string, ok bool) {
	return modelid.SplitModelFQN(input)
}

func containsInt(values []int, target int) bool {
	for _, v := range values {
		if v == target {
			return true
		}
	}

	return false
}
