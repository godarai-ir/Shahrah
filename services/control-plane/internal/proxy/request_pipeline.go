package proxy

import (
	"context"
	"errors"
	"io"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"go.uber.org/zap"
)

var errTunnelTransportUnavailable = errors.New("tunnel transport is unavailable")

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

// RequestPipeline owns the shared reserve, dispatch, and settlement lifecycle.
// Operation-specific payload decoding and response encoding remain in handlers.
type RequestPipeline struct {
	tunnel  platform.TunnelTransport
	billing budgeter
	wallets walletStore
	logger  *zap.Logger
}

// NewRequestPipeline constructs the shared gateway request pipeline.
func NewRequestPipeline(
	tunnel platform.TunnelTransport,
	billing budgeter,
	wallets walletStore,
	logger *zap.Logger,
) *RequestPipeline {
	return &RequestPipeline{
		tunnel:  tunnel,
		billing: billing,
		wallets: wallets,
		logger:  logger,
	}
}

// Reserve executes the budget admission stage without depending on HTTP.
func (p *RequestPipeline) Reserve(ctx context.Context, input BudgetRequest) (*platform.BudgetReservation, error) {
	if firstBillingMode(input.BillingMode) == platform.BillingModeExternalBilled {
		return nil, nil
	}

	if manager, ok := p.billing.(budgetReservationManager); ok {
		return manager.ReserveBudget(ctx, input.Auth, input.Estimated, input.RequestID)
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

	if manager, ok := p.billing.(budgetReservationManager); ok && input.Reservation != nil {
		return manager.CommitBudget(ctx, input.Reservation, input.Actual)
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

	manager, ok := p.billing.(budgetReservationManager)
	if !ok {
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

	rawTunnel, ok := p.tunnel.(platform.TunnelRawTransport)
	if !ok {
		return nil, errTunnelTransportUnavailable
	}

	return rawTunnel.ProxyRaw(ctx, request, body, contentLength)
}

// HasRawTransport reports whether raw body dispatch is supported.
func (p *RequestPipeline) HasRawTransport() bool {
	if p == nil || p.tunnel == nil {
		return false
	}

	_, ok := p.tunnel.(platform.TunnelRawTransport)

	return ok
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
