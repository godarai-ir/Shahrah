// Package proxy hosts OpenAI-compatible request handlers for the control plane.
package proxy

import (
	"context"
	"errors"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	metricsobs "github.com/Delpak/shahrah/services/control-plane/internal/observability/metrics"
	"github.com/Delpak/shahrah/services/control-plane/internal/router"
	"github.com/Delpak/shahrah/services/control-plane/internal/storage"
	"github.com/Delpak/shahrah/services/control-plane/internal/virtualmodels"
	"go.uber.org/zap"
)

// RequestLogger persists request lifecycle records.
type RequestLogger interface {
	LogRequest(ctx context.Context, entry platform.RequestLogEntry) error
}

// SpendLogger persists spend records.
type SpendLogger interface {
	LogSpend(ctx context.Context, log platform.SpendLog) error
}

type anthropicModelRuntimeProfileResolver interface {
	ResolveModelRuntimeProfiles(
		models []string,
		expectedRevision string,
	) ([]platform.ModelRuntimeProfileBinding, error)
}

// BaseHandlerDependencies groups shared proxy handler dependencies.
type BaseHandlerDependencies struct {
	Router                   *router.Router
	Tunnel                   platform.TunnelTransport
	Billing                  budgeter
	Wallets                  walletStore
	AnthropicRuntimeProfiles anthropicModelRuntimeProfileResolver
	RequestLogger            RequestLogger
	SpendLogger              SpendLogger
	Logger                   *zap.Logger
	EmitMetric               func(input metricsobs.BuildInput)
}

// ChatCompletionDependencies extends the base wiring for chat completion routing.
type ChatCompletionDependencies struct {
	Base              BaseHandlerDependencies
	VirtualModelStore interface {
		GetVirtualModelByAlias(ctx context.Context, tenantID string, alias string) (*platform.VirtualModel, error)
	}
	GuardrailResolver guardrailResolver
	PromptResolver    promptResolver
}

type usageLoggers struct {
	requests RequestLogger
	spend    SpendLogger
}

func (l usageLoggers) LogRequest(ctx context.Context, entry platform.RequestLogEntry) error {
	if l.requests == nil {
		return nil
	}

	return l.requests.LogRequest(ctx, entry)
}

func (l usageLoggers) LogSpend(ctx context.Context, log platform.SpendLog) error {
	if l.spend == nil {
		return nil
	}

	return l.spend.LogSpend(ctx, log)
}

func (d BaseHandlerDependencies) logger() usageLogger {
	return usageLoggers{
		requests: d.RequestLogger,
		spend:    d.SpendLogger,
	}
}

func (d BaseHandlerDependencies) metricEmitter() func(input metricsobs.BuildInput) {
	return d.EmitMetric
}

func newChatVirtualModelDeps(store interface {
	GetVirtualModelByAlias(ctx context.Context, tenantID string, alias string) (*platform.VirtualModel, error)
}) (virtualmodels.VirtualModelResolver, virtualmodels.RoutingEngine, virtualmodels.FallbackExecutor) {
	if store == nil {
		return nil, nil, nil
	}

	return virtualmodels.NewResolver(store, virtualmodels.WithNotFoundError(func(err error) bool {
		return errors.Is(err, storage.ErrVirtualModelNotFound)
	})), virtualmodels.NewRoutingEngine(), virtualmodels.NewExecutor(virtualmodels.NewDefaultErrorNormalizer())
}
