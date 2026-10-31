// Package proxy hosts OpenAI-compatible request handlers for the control plane.
package proxy

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/guardrails"
	metricsobs "github.com/Delpak/shahrah/services/control-plane/internal/observability/metrics"
	"github.com/Delpak/shahrah/services/control-plane/internal/responsecache"
	"github.com/Delpak/shahrah/services/control-plane/internal/router"
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

type virtualModelStore interface {
	GetVirtualModelByAlias(ctx context.Context, tenantID string, alias string) (*platform.VirtualModel, error)
}

type anthropicModelRuntimeProfileResolver interface {
	ResolveModelRuntimeProfiles(
		models []string,
		expectedRevision string,
	) ([]platform.ModelRuntimeProfileBinding, error)
}

// CapabilityPolicy defines how an unavailable optional gateway capability is handled.
type CapabilityPolicy string

const (
	// CapabilityDisabled makes the capability intentionally unavailable.
	CapabilityDisabled CapabilityPolicy = "disabled"
	// CapabilityFailOpen uses the compatibility path when the capability is unavailable.
	CapabilityFailOpen CapabilityPolicy = "fail_open"
	// CapabilityFailClosed rejects work when the capability is unavailable.
	CapabilityFailClosed CapabilityPolicy = "fail_closed"
)

// BillingCapabilities declares optional billing behavior used by the request pipeline.
type BillingCapabilities struct {
	MeteredUsage             meteredUsageBudgeter
	MeteredUsagePolicy       CapabilityPolicy
	DeploymentPricing        deploymentPricingBudgeter
	RequireDeploymentPricing bool
	BudgetReservations       budgetReservationManager
	BudgetReservationPolicy  CapabilityPolicy
	BatchUsage               batchUsageFinalizer
	BatchUsagePolicy         CapabilityPolicy
	VideoUsage               videoUsagePricer
}

// Validate verifies that required billing capabilities are available at startup.
func (c BillingCapabilities) Validate() error {
	if c.RequireDeploymentPricing && c.DeploymentPricing == nil {
		return errors.New("deployment pricing capability is required")
	}

	if err := validateCapability("metered usage pricing", c.MeteredUsagePolicy, c.MeteredUsage != nil); err != nil {
		return err
	}

	if err := validateCapability("budget reservations", c.BudgetReservationPolicy, c.BudgetReservations != nil); err != nil {
		return err
	}

	return validateCapability("batch usage finalization", c.BatchUsagePolicy, c.BatchUsage != nil)
}

func validateCapability(name string, policy CapabilityPolicy, available bool) error {
	if policy == "" {
		policy = CapabilityDisabled
	}

	switch policy {
	case CapabilityDisabled, CapabilityFailOpen:
		return nil
	case CapabilityFailClosed:
		if !available {
			return fmt.Errorf("%s capability is required by fail-closed policy", name)
		}

		return nil
	default:
		return fmt.Errorf("unsupported %s capability policy %q", name, policy)
	}
}

func compatibilityBillingCapabilities() BillingCapabilities {
	return BillingCapabilities{
		MeteredUsagePolicy:      CapabilityFailOpen,
		BudgetReservationPolicy: CapabilityFailOpen,
		BatchUsagePolicy:        CapabilityDisabled,
	}
}

// BaseHandlerDependencies groups shared proxy handler dependencies.
type BaseHandlerDependencies struct {
	Pipeline                  *RequestPipeline
	Router                    *router.Router
	Tunnel                    platform.TunnelTransport
	RawTunnel                 platform.TunnelRawTransport
	Billing                   budgeter
	BillingCapabilities       BillingCapabilities
	AnthropicRuntimeProfiles  anthropicModelRuntimeProfileResolver
	ProviderResources         providerNativeResourceStore
	AnthropicResources        anthropicNativeResourceStore
	MistralResources          mistralBatchResourceStore
	Deployments               responseDeploymentLister
	ResponseProviderAuth      responseProviderAuthResolver
	RequestLogger             RequestLogger
	SpendLogger               SpendLogger
	Logger                    *zap.Logger
	EmitMetric                func(input metricsobs.BuildInput)
	PipelineConfig            RequestPipelineConfig
	ProviderFeatures          ProviderFeatureConfig
	VirtualModelStore         virtualModelStore
	VirtualModelAliasCacheTTL time.Duration
	VirtualModelNegativeTTL   time.Duration
	ResponseCache             *responsecache.Service
}

// ProviderFeatureConfig is an immutable provider endpoint rollout snapshot.
type ProviderFeatureConfig struct {
	Configured                   bool
	MistralFilesEnabled          bool
	MistralBatchEnabled          bool
	MistralAgentsEnabled         bool
	MistralConversationsEnabled  bool
	MistralConversationStreaming bool
	AnthropicAccountIDs          []string
	VideosEnabled                bool
	VideoCanaryTenantIDs         []string
}

// ChatCompletionDependencies extends the base wiring for chat completion routing.
type ChatCompletionDependencies struct {
	Base                      BaseHandlerDependencies
	VirtualModelStore         virtualModelStore
	GuardrailResolver         guardrailResolver
	PromptResolver            promptResolver
	Guardrails                *guardrails.Engine
	ResponseCache             *responsecache.Service
	VirtualModelAliasCacheTTL time.Duration
	VirtualModelNegativeTTL   time.Duration
}

type usageLoggers struct {
	requests RequestLogger
	spend    SpendLogger
}

func (l usageLoggers) SupportsLazyRequestLogPayload() bool {
	sink, ok := l.requests.(interface {
		SupportsLazyRequestLogPayload() bool
	})

	return ok && sink.SupportsLazyRequestLogPayload()
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

func (d BaseHandlerDependencies) requestPipeline() *RequestPipeline {
	if d.Pipeline != nil {
		return d.Pipeline
	}

	return NewRequestPipelineWithConfig(
		d.Tunnel,
		d.RawTunnel,
		d.Billing,
		d.BillingCapabilities,
		d.Logger,
		d.PipelineConfig,
	)
}

func newVirtualModelDeps(store virtualModelStore, hitTTL time.Duration, negativeTTL time.Duration) (virtualmodels.VirtualModelResolver, virtualmodels.RoutingEngine, virtualmodels.FallbackExecutor) {
	if store == nil {
		return nil, nil, nil
	}

	options := []virtualmodels.ResolverOption{virtualmodels.WithNotFoundError(func(err error) bool {
		return errors.Is(err, virtualmodels.ErrNotFound)
	})}
	if hitTTL > 0 && negativeTTL > 0 {
		options = append(options, virtualmodels.WithAliasCacheTTLs(hitTTL, negativeTTL))
	}

	return virtualmodels.NewResolver(store, options...), virtualmodels.NewRoutingEngine(), virtualmodels.NewExecutor(virtualmodels.NewDefaultErrorNormalizer())
}
