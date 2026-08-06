package proxy

import (
	"context"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/stretchr/testify/require"
)

type pipelineBudgetManager struct {
	testBudgeter
	reservation   *platform.BudgetReservation
	reserveInput  BudgetRequest
	committed     platform.UsageCostBreakdown
	releaseReason string
}

func (m *pipelineBudgetManager) ReserveBudget(
	_ context.Context,
	auth *platform.AuthContext,
	estimated platform.UsageCostBreakdown,
	requestID string,
) (*platform.BudgetReservation, error) {
	m.reserveInput = BudgetRequest{Auth: auth, Estimated: estimated, RequestID: requestID}

	return m.reservation, nil
}

func (m *pipelineBudgetManager) CommitBudget(
	_ context.Context,
	_ *platform.BudgetReservation,
	actual platform.UsageCostBreakdown,
) error {
	m.committed = actual

	return nil
}

func (m *pipelineBudgetManager) ReleaseBudget(
	_ context.Context,
	_ *platform.BudgetReservation,
	reason string,
) error {
	m.releaseReason = reason

	return nil
}

type pipelineTunnel struct {
	request  platform.TunnelRequest
	response *platform.TunnelResponse
}

func (t *pipelineTunnel) Proxy(
	_ context.Context,
	request platform.TunnelRequest,
) (*platform.TunnelResponse, error) {
	t.request = request

	return t.response, nil
}

func (t *pipelineTunnel) Stream(context.Context, platform.TunnelRequest) (platform.TunnelStream, error) {
	return nil, nil
}

func (t *pipelineTunnel) Healthy(context.Context) error {
	return nil
}

func TestRequestPipelineReserveAndSettleWithReservationManager(t *testing.T) {
	reservation := &platform.BudgetReservation{ID: "reservation-1"}
	manager := &pipelineBudgetManager{reservation: reservation}
	pipeline := NewRequestPipeline(nil, manager, &testWallets{}, nil)
	authInfo := &platform.AuthContext{TenantID: "tenant-1"}
	estimated := platform.UsageCostBreakdown{TotalCostIRR: 1000}

	actualReservation, err := pipeline.Reserve(t.Context(), BudgetRequest{
		Auth:      authInfo,
		Estimated: estimated,
		RequestID: "request-1",
	})
	require.NoError(t, err)
	require.Same(t, reservation, actualReservation)
	require.Same(t, authInfo, manager.reserveInput.Auth)
	require.Equal(t, estimated, manager.reserveInput.Estimated)
	require.Equal(t, "request-1", manager.reserveInput.RequestID)

	actual := platform.UsageCostBreakdown{TotalCostIRR: 750}
	err = pipeline.Settle(t.Context(), SettlementRequest{
		Reservation: reservation,
		TenantID:    "tenant-1",
		Actual:      actual,
		RequestID:   "request-1",
	})
	require.NoError(t, err)
	require.Equal(t, actual, manager.committed)
}

func TestRequestPipelineExternalBillingSkipsReserveAndReleasesExistingReservation(t *testing.T) {
	reservation := &platform.BudgetReservation{ID: "reservation-1"}
	manager := &pipelineBudgetManager{reservation: reservation}
	pipeline := NewRequestPipeline(nil, manager, nil, nil)

	actualReservation, err := pipeline.Reserve(t.Context(), BudgetRequest{
		BillingMode: platform.BillingModeExternalBilled,
	})
	require.NoError(t, err)
	require.Nil(t, actualReservation)
	require.Empty(t, manager.reserveInput.RequestID)

	err = pipeline.Settle(t.Context(), SettlementRequest{
		Reservation: reservation,
		BillingMode: platform.BillingModeExternalBilled,
	})
	require.NoError(t, err)
	require.Equal(t, "external_billed", manager.releaseReason)
	require.Zero(t, manager.committed.TotalCostIRR)
}

func TestRequestPipelineFallsBackToWalletDebit(t *testing.T) {
	wallets := &testWallets{}
	pipeline := NewRequestPipeline(nil, &testBudgeter{}, wallets, nil)
	actual := platform.UsageCostBreakdown{TotalCostIRR: 500}

	err := pipeline.Settle(t.Context(), SettlementRequest{
		TenantID:  "tenant-1",
		Actual:    actual,
		RequestID: "request-1",
	})
	require.NoError(t, err)
	require.Equal(t, []walletDebitCall{{
		TenantID:  "tenant-1",
		Breakdown: actual,
		RequestID: "request-1",
	}}, wallets.debits)
}

func TestRequestPipelineDispatchKeepsProviderResponseRaw(t *testing.T) {
	response := &platform.TunnelResponse{
		StatusCode: 201,
		Headers:    map[string]string{"content-type": "application/provider+json"},
		Body:       []byte(`{"provider_specific":true}`),
	}
	tunnel := &pipelineTunnel{response: response}
	pipeline := NewRequestPipeline(tunnel, &testBudgeter{}, nil, nil)
	request := platform.TunnelRequest{RequestID: "request-1", Body: []byte(`{"raw":true}`)}

	actual, err := pipeline.Dispatch(t.Context(), request)
	require.NoError(t, err)
	require.Same(t, response, actual)
	require.Equal(t, request, tunnel.request)
}
