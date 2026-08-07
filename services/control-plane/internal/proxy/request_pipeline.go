package proxy

import (
	"context"
	"errors"
	"io"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/router"
	"go.uber.org/zap"
)

var errTunnelTransportUnavailable = errors.New("tunnel transport is unavailable")

var (
	errMeteredUsagePricingUnavailable = errors.New("metered usage pricing capability is unavailable")
	errBudgetReservationUnavailable   = errors.New("budget reservation capability is unavailable")
)

// BudgetRequest is the normalized input for the pipeline budget stage.
type BudgetRequest struct {
	Auth        *platform.AuthContext
	Estimated   platform.UsageCostBreakdown
	RequestID   string
	BillingMode string
}

// SettlementRequest is the normalized input for the pipeline settlement stage.
type SettlementRequest struct {
	Reservation *platform.BudgetReservation
	TenantID    string
	Actual      platform.UsageCostBreakdown
	RequestID   string
	BillingMode string
}

// RequestPipelineConfig controls shared gateway lifecycle behavior.
type RequestPipelineConfig struct {
	DetachedWorkTimeout time.Duration
}

// RequestPipeline owns the shared reserve, dispatch, and settlement lifecycle.
// Operation-specific payload decoding and response encoding remain in handlers.
type RequestPipeline struct {
	tunnel              platform.TunnelTransport
	rawTunnel           platform.TunnelRawTransport
	billing             budgeter
	billingCapabilities BillingCapabilities
	wallets             walletStore
	logger              *zap.Logger
	detachedWorkTimeout time.Duration
}

// NewRequestPipeline constructs the shared gateway request pipeline.
func NewRequestPipeline(
	tunnel platform.TunnelTransport,
	rawTunnel platform.TunnelRawTransport,
	billing budgeter,
	billingCapabilities BillingCapabilities,
	wallets walletStore,
	logger *zap.Logger,
) *RequestPipeline {
	return NewRequestPipelineWithConfig(
		tunnel,
		rawTunnel,
		billing,
		billingCapabilities,
		wallets,
		logger,
		RequestPipelineConfig{},
	)
}

// NewRequestPipelineWithConfig constructs one configured gateway request pipeline.
func NewRequestPipelineWithConfig(
	tunnel platform.TunnelTransport,
	rawTunnel platform.TunnelRawTransport,
	billing budgeter,
	billingCapabilities BillingCapabilities,
	wallets walletStore,
	logger *zap.Logger,
	config RequestPipelineConfig,
) *RequestPipeline {
	if config.DetachedWorkTimeout <= 0 {
		config.DetachedWorkTimeout = 5 * time.Second
	}

	return &RequestPipeline{
		tunnel:              tunnel,
		rawTunnel:           rawTunnel,
		billing:             billing,
		billingCapabilities: billingCapabilities,
		wallets:             wallets,
		logger:              logger,
		detachedWorkTimeout: config.DetachedWorkTimeout,
	}
}

// DetachedContext creates a bounded context for settlement after client disconnect.
func (p *RequestPipeline) DetachedContext() (context.Context, context.CancelFunc) {
	timeout := 5 * time.Second
	if p != nil && p.detachedWorkTimeout > 0 {
		timeout = p.detachedWorkTimeout
	}

	return context.WithTimeout(context.Background(), timeout)
}

// ReleaseDetached bounds reservation cleanup that must survive client disconnect.
func (p *RequestPipeline) ReleaseDetached(reservation *platform.BudgetReservation, reason string) {
	ctx, cancel := p.DetachedContext()
	defer cancel()

	p.Release(ctx, reservation, reason)
}

// Reserve executes the budget admission stage without depending on HTTP.
func (p *RequestPipeline) Reserve(ctx context.Context, input BudgetRequest) (*platform.BudgetReservation, error) {
	if firstBillingMode(input.BillingMode) == platform.BillingModeExternalBilled {
		return nil, nil
	}

	reservationPolicy := normalizedCapabilityPolicy(p.billingCapabilities.BudgetReservationPolicy)
	if reservationPolicy != CapabilityDisabled && p.billingCapabilities.BudgetReservations != nil {
		return p.billingCapabilities.BudgetReservations.ReserveBudget(ctx, input.Auth, input.Estimated, input.RequestID)
	}

	if reservationPolicy == CapabilityFailClosed {
		return nil, errBudgetReservationUnavailable
	}

	if err := p.billing.CheckBudget(ctx, input.Auth, input.Estimated); err != nil {
		return nil, err
	}

	return nil, nil
}

// Settle executes the final budget commit or wallet debit stage.
func (p *RequestPipeline) Settle(ctx context.Context, input SettlementRequest) error {
	if firstBillingMode(input.BillingMode) == platform.BillingModeExternalBilled {
		p.Release(ctx, input.Reservation, "external_billed")

		return nil
	}

	reservationPolicy := normalizedCapabilityPolicy(p.billingCapabilities.BudgetReservationPolicy)
	if reservationPolicy != CapabilityDisabled && p.billingCapabilities.BudgetReservations != nil && input.Reservation != nil {
		return p.billingCapabilities.BudgetReservations.CommitBudget(ctx, input.Reservation, input.Actual)
	}

	if input.Reservation != nil && reservationPolicy == CapabilityFailClosed {
		return errBudgetReservationUnavailable
	}

	if p.wallets == nil {
		return nil
	}

	_, err := p.wallets.DebitWallet(ctx, input.TenantID, input.Actual, input.RequestID)
	if err != nil && p.logger != nil {
		p.logger.Warn("failed to debit wallet", zap.Error(err), zap.String("request_id", input.RequestID))
	}

	return err
}

// Release releases a reservation on every non-settled terminal path.
func (p *RequestPipeline) Release(
	ctx context.Context,
	reservation *platform.BudgetReservation,
	reason string,
) {
	if reservation == nil {
		return
	}

	if normalizedCapabilityPolicy(p.billingCapabilities.BudgetReservationPolicy) == CapabilityDisabled {
		return
	}

	manager := p.billingCapabilities.BudgetReservations
	if manager == nil {
		return
	}

	if err := manager.ReleaseBudget(ctx, reservation, reason); err != nil && p.logger != nil {
		p.logger.Warn(
			"failed to release budget reservation",
			zap.Error(err),
			zap.String("request_id", reservation.ID),
			zap.String("reason", reason),
		)
	}
}

func (p *RequestPipeline) priceMeteredUsage(
	ctx context.Context,
	model string,
	dimension string,
	quantity string,
	estimated platform.UsageCostBreakdown,
) (platform.UsageCostBreakdown, error) {
	if p != nil &&
		normalizedCapabilityPolicy(p.billingCapabilities.MeteredUsagePolicy) != CapabilityDisabled &&
		p.billingCapabilities.MeteredUsage != nil {
		return p.billingCapabilities.MeteredUsage.PriceMeteredUsage(ctx, model, dimension, quantity, estimated)
	}

	if p != nil && normalizedCapabilityPolicy(p.billingCapabilities.MeteredUsagePolicy) == CapabilityFailClosed {
		return platform.UsageCostBreakdown{}, errMeteredUsagePricingUnavailable
	}

	return estimated, nil
}

func (p *RequestPipeline) applyDeploymentPricingOverride(
	ctx context.Context,
	estimated platform.UsageCostBreakdown,
	deployment *router.Deployment,
) (platform.UsageCostBreakdown, error) {
	if deployment == nil || deployment.PricingMode != platform.PricingModePrivate {
		return estimated, nil
	}

	if p == nil || p.billingCapabilities.DeploymentPricing == nil {
		return platform.UsageCostBreakdown{}, errors.New("private pricing override capability is unavailable")
	}

	return p.billingCapabilities.DeploymentPricing.ApplyPricingOverride(ctx, estimated, deployment.EffectivePricing)
}

func (p *RequestPipeline) batchUsageFinalizer() batchUsageFinalizer {
	if p == nil || normalizedCapabilityPolicy(p.billingCapabilities.BatchUsagePolicy) == CapabilityDisabled {
		return nil
	}

	return p.billingCapabilities.BatchUsage
}

func normalizedCapabilityPolicy(policy CapabilityPolicy) CapabilityPolicy {
	if policy == "" {
		return CapabilityDisabled
	}

	return policy
}

// EnsureReady verifies that the dispatch transport is healthy.
func (p *RequestPipeline) EnsureReady(ctx context.Context) error {
	if p == nil || p.tunnel == nil {
		return errTunnelTransportUnavailable
	}

	return p.tunnel.Healthy(ctx)
}

// Dispatch sends a request and returns the raw provider response unchanged.
func (p *RequestPipeline) Dispatch(
	ctx context.Context,
	request platform.TunnelRequest,
) (*platform.TunnelResponse, error) {
	if p == nil || p.tunnel == nil {
		return nil, errTunnelTransportUnavailable
	}

	return p.tunnel.Proxy(ctx, request)
}

// DispatchStream opens a provider stream without interpreting provider chunks.
func (p *RequestPipeline) DispatchStream(
	ctx context.Context,
	request platform.TunnelRequest,
) (platform.TunnelStream, error) {
	if p == nil || p.tunnel == nil {
		return nil, errTunnelTransportUnavailable
	}

	return p.tunnel.Stream(ctx, request)
}

// DispatchRaw streams a large request body without interpreting provider data.
func (p *RequestPipeline) DispatchRaw(
	ctx context.Context,
	request platform.TunnelRequest,
	body io.Reader,
	contentLength int64,
) (*platform.TunnelRawResponse, error) {
	if p == nil || p.tunnel == nil {
		return nil, errTunnelTransportUnavailable
	}

	if p.rawTunnel == nil {
		return nil, errTunnelTransportUnavailable
	}

	return p.rawTunnel.ProxyRaw(ctx, request, body, contentLength)
}

// HasRawTransport reports whether raw body dispatch is supported.
func (p *RequestPipeline) HasRawTransport() bool {
	return p != nil && p.rawTunnel != nil
}

// HasTransport reports whether dispatch can be attempted.
func (p *RequestPipeline) HasTransport() bool {
	return p != nil && p.tunnel != nil
}

func firstBillingMode(mode string) string {
	mode = strings.TrimSpace(mode)
	if mode == "" {
		return platform.BillingModeWalletBilled
	}

	return mode
}
