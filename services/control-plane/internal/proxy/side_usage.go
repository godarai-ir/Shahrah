package proxy

import (
	"context"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/credits"
	"github.com/Delpak/shahrah/services/control-plane/internal/guardrails"
	"go.uber.org/zap"
)

// sideUsageRecorder charges the tenant's budget windows for side usage.
type sideUsageRecorder interface {
	RecordSideUsage(ctx context.Context, auth *platform.AuthContext, breakdown platform.UsageCostBreakdown, amountMC int64) error
}

// SideUsageCharge is one batch of calls GodarAI made for a request.
type SideUsageCharge struct {
	Auth      *platform.AuthContext
	RequestID string
	TraceID   string
	// Stage names the side call, for example guardrail_input; it keeps the
	// side call's spend log apart from the request's own.
	Stage   string
	Product string
	Model   string
	// RequestModel and RequestType describe the originating request, so the
	// budget rules that match it also count its side usage.
	RequestModel string
	RequestType  string
	Calls        int
	PriceMC      int64
}

// ChargeSideUsage charges the tenant for side calls: the budget windows right
// away, and the balance through a spend log the credit ledger settles.
// Failures are logged and never fail the request.
func (p *RequestPipeline) ChargeSideUsage(ctx context.Context, logs usageLogger, charge SideUsageCharge) {
	amountMC := int64(charge.Calls) * charge.PriceMC
	if charge.Auth == nil || amountMC <= 0 || logs == nil {
		return
	}

	recorder, ok := p.billingCapabilities.BudgetReservations.(sideUsageRecorder)
	if !ok {
		return
	}

	breakdown := platform.UsageCostBreakdown{Model: charge.RequestModel, RequestType: charge.RequestType}

	if err := recorder.RecordSideUsage(ctx, charge.Auth, breakdown, amountMC); err != nil && p.logger != nil {
		p.logger.Warn("failed to record side usage in budget windows", zap.Error(err), zap.String("request_id", charge.RequestID))
	}

	now := time.Now()
	logSpendBestEffort(p.logger, logs, ctx, platform.SpendLog{
		RequestID:      strings.TrimSpace(charge.RequestID) + ":" + charge.Stage,
		TokenHash:      charge.Auth.TokenHash,
		TenantID:       charge.Auth.TenantID,
		OrganizationID: charge.Auth.OrganizationID,
		TeamID:         charge.Auth.TeamID,
		UserID:         charge.Auth.UserID,
		Model:          charge.Model,
		Status:         "success",
		// The ledger reads the trace id from the metadata, which links the
		// side call to the request it served.
		RequestMetadata: map[string]interface{}{
			"trace_id": charge.TraceID, "side_usage_stage": charge.Stage, "side_usage_calls": charge.Calls,
			"request_model": charge.RequestModel,
		},
		Credits: &platform.SpendLogCredits{
			Millicredits: amountMC,
			Product:      charge.Product,
			Source:       string(credits.SourceFixedPrice),
		},
		CompletionStart: now,
		CompletionEnd:   now,
	})
}

// chargeJudgeUsage charges the semantic-guardrail judge calls counted on usage
// and clears the count.
func (p *RequestPipeline) chargeJudgeUsage(ctx context.Context, logs usageLogger, auth *platform.AuthContext, requestID, traceID, requestModel, stage string, usage *guardrails.JudgeUsage) {
	calls := usage.Calls()
	if calls == 0 {
		return
	}

	usage.Reset()

	model := usage.Model()
	if model == "" {
		model = "guardrail-judge"
	}

	p.ChargeSideUsage(ctx, logs, SideUsageCharge{
		Auth: auth, RequestID: requestID, TraceID: traceID, Stage: stage,
		Product: platform.CreditProductGuardrail, Model: model, RequestModel: requestModel, RequestType: "chat_completions",
		Calls: calls, PriceMC: credits.GuardrailJudgeCallMC,
	})
}
