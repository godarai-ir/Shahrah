package proxy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	"github.com/Delpak/shahrah/services/control-plane/internal/providernative"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
)

const xaiDeferredResourceKind = "deferred_chat_completion"

func (h *ChatCompletionHandler) handleXAIDeferredCreate(
	c *fiber.Ctx,
	req contractsopenai.ChatCompletionRequest,
	originalBody []byte,
	authInfo *platform.AuthContext,
	estimated platform.UsageCostBreakdown,
	tunnelReq platform.TunnelRequest,
	reservation *platform.BudgetReservation,
	resp *platform.TunnelResponse,
	start time.Time,
) error {
	var result struct {
		RequestID string `json:"request_id"`
	}
	if json.Unmarshal(resp.Body, &result) != nil || strings.TrimSpace(result.RequestID) == "" {
		return writeAPIError(c, fiber.StatusBadGateway, "xAI returned an invalid deferred completion response", "api_error")
	}

	retention := 24 * time.Hour
	if binding, ok := platform.FindModelRuntimeProfile(tunnelReq.ModelRuntimeProfiles, tunnelReq.Deployment.Model); ok && binding.Profile.DeferredCompletions.RetentionSeconds > 0 {
		retention = time.Duration(binding.Profile.DeferredCompletions.RetentionSeconds) * time.Second
	}

	expiresAt := time.Now().UTC().Add(retention)
	requestID := strings.TrimSpace(result.RequestID)

	_, err := h.resources.UpsertProviderNativeResourceMapping(c.Context(), platform.ProviderNativeResourceMapping{
		PublicID:          requestID,
		TenantID:          authInfo.TenantID,
		Provider:          "xai",
		ProviderAccountID: tunnelReq.Deployment.ProviderAccountID,
		ResourceKind:      xaiDeferredResourceKind,
		UpstreamID:        requestID,
		Status:            "pending",
		ReservationID:     reservationID(reservation),
		Reservation:       reservation,
		PricingRevision:   estimated.PricingRevision,
		EstimatedUsage:    &estimated,
		ExpiresAt:         &expiresAt,
		Metadata: map[string]interface{}{
			"deployment_id": tunnelReq.Deployment.ID,
			"model":         req.Model,
			"billing_mode":  tunnelReq.Deployment.BillingMode,
			"request_id":    tunnelReq.RequestID,
		},
	})
	if err != nil {
		return writeAPIError(c, fiber.StatusServiceUnavailable, "Deferred completion routing state is unavailable", "routing_error")
	}

	responseTimeMS := int(time.Since(start).Milliseconds())
	entry := buildRequestLogEntry(c, authInfo, tunnelReq.Deployment, req.Model, "/v1/chat/completions", "chat_completions_deferred", "pending", resp.StatusCode, start, responseTimeMS)
	entry.RequestID = tunnelReq.RequestID
	entry.TraceID = tunnelReq.TraceID
	entry.RequestBody = buildChatRequestLogPayload(req, originalBody)
	entry.ResponseBody = map[string]interface{}{"request_id": requestID, "status": "pending"}
	logRequestBestEffort(h.logger, h.logs, c.Context(), entry)

	forwardProviderHeaders(c, resp.Headers)
	setTraceHeaders(c, tunnelReq.TraceID)
	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	return c.Status(resp.StatusCode).Send(resp.Body)
}

//nolint:govet // Short-lived branch errors are checked immediately and intentionally scoped with short declarations.
func (h *ChatCompletionHandler) HandleXAIDeferredGet(c *fiber.Ctx) error {
	authInfo := auth.GetAuthInfo(c)
	if authInfo == nil {
		return writeAPIError(c, fiber.StatusUnauthorized, "Authentication required", "authentication_error")
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return writeAPIError(c, fiber.StatusBadRequest, "Tenant resolution required", "invalid_request_error")
	}

	authInfo.TenantID = currentTenant.ID

	if h.resources == nil || h.deployments == nil {
		return writeAPIError(c, fiber.StatusServiceUnavailable, "Deferred completion routing state is unavailable", "routing_error")
	}

	publicID := strings.TrimSpace(c.Params("request_id"))

	mapping, err := h.resources.GetProviderNativeResourceMapping(c.Context(), currentTenant.ID, "xai", xaiDeferredResourceKind, publicID)
	if errors.Is(err, providernative.ErrResourceNotFound) {
		return writeAPIError(c, fiber.StatusNotFound, "Deferred completion not found", "invalid_request_error")
	}

	if err != nil || mapping == nil {
		return writeAPIError(c, fiber.StatusServiceUnavailable, "Deferred completion routing state is unavailable", "routing_error")
	}

	if strings.EqualFold(mapping.Status, "retrieved") {
		return writeAPIError(c, fiber.StatusGone, "Deferred completion has already been retrieved", "invalid_request_error")
	}

	if mapping.ExpiresAt != nil && time.Now().UTC().After(*mapping.ExpiresAt) {
		h.pipeline.Release(c.Context(), mapping.Reservation, "deferred_completion_expired")
		_ = h.resources.UpdateProviderNativeResourceStatus(c.Context(), currentTenant.ID, "xai", xaiDeferredResourceKind, publicID, "expired") //nolint:errcheck // Failure is intentionally handled by the surrounding best-effort path.

		return writeAPIError(c, fiber.StatusGone, "Deferred completion has expired", "invalid_request_error")
	}

	deployment, err := h.resolveXAIDeferredDeployment(c.Context(), currentTenant.ID, mapping.ProviderAccountID)
	if err != nil {
		return writeAPIError(c, fiber.StatusServiceUnavailable, err.Error(), "routing_error")
	}

	requestID := responseRequestID(c)

	tunnelReq := platform.TunnelRequest{
		RequestID: requestID,
		TenantID:  currentTenant.ID,
		TraceID:   requestID,
		Operation: platform.ProviderOperationXAIDeferredChatGet,
		OperationInput: platform.ProviderOperationInput{
			ResourceID: mapping.UpstreamID,
		},
		Deployment: deployment,
		Headers:    responseProviderHeaders(c),
		RequestMetadata: enrichRequestMetadataForDeployment(map[string]interface{}{
			"provider_endpoint_path": "/chat/deferred-completion/{request_id}",
			"resource_kind":          xaiDeferredResourceKind,
		}, deployment),
	}
	if err := h.pipeline.EnsureReady(c.Context()); err != nil {
		return writeAPIError(c, fiber.StatusServiceUnavailable, "Tunnel path unavailable", "tunnel_unavailable")
	}

	resp, err := h.pipeline.Dispatch(c.Context(), tunnelReq)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadGateway, fmt.Sprintf("Tunnel error: %v", err), "api_error")
	}

	if resp.StatusCode == fiber.StatusAccepted {
		forwardProviderHeaders(c, resp.Headers)
		setTraceHeaders(c, requestID)

		return c.Status(fiber.StatusAccepted).Send(resp.Body)
	}

	if resp.StatusCode >= httpErrorFloor {
		return transformProviderError(c, resp.StatusCode, resp.Body, resp.Headers)
	}

	var completion contractsopenai.ChatCompletionResponse
	if json.Unmarshal(resp.Body, &completion) != nil {
		return writeAPIError(c, fiber.StatusBadGateway, "Failed to decode deferred completion response", "api_error")
	}

	estimated := platform.UsageCostBreakdown{}
	if mapping.EstimatedUsage != nil {
		estimated = *mapping.EstimatedUsage
	}

	model, _ := mapping.Metadata["model"].(string) //nolint:errcheck // Optional metadata uses the zero value when absent or mismatched.

	finalCost, err := h.billing.FinalizeFromUsage(c.Context(), billingModelForDeployment(model, deployment), completion.Usage, estimated)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadGateway, "Failed to finalize deferred completion usage", "billing_error")
	}

	tunnelReq.RequestMetadata = withXAIUsageMetadata(tunnelReq.RequestMetadata, deployment.Provider, completion.Usage)
	billingMode, _ := mapping.Metadata["billing_mode"].(string) //nolint:errcheck // Optional metadata uses the zero value when absent or mismatched.

	originalRequestID, _ := mapping.Metadata["request_id"].(string) //nolint:errcheck // Optional metadata uses the zero value when absent or mismatched.
	if strings.TrimSpace(originalRequestID) == "" {
		originalRequestID = requestID
	}

	if err := h.pipeline.Settle(c.Context(), SettlementRequest{Reservation: mapping.Reservation, TenantID: currentTenant.ID, Actual: finalCost, RequestID: originalRequestID, BillingMode: billingMode}); err != nil {
		return writeBudgetCheckError(c, err)
	}

	if err := h.resources.UpdateProviderNativeResourceStatus(c.Context(), currentTenant.ID, "xai", xaiDeferredResourceKind, publicID, "retrieved"); err != nil {
		return writeAPIError(c, fiber.StatusServiceUnavailable, "Deferred completion status could not be finalized", "routing_error")
	}

	responseTimeMS := 0
	logSpendBestEffort(h.logger, h.logs, c.Context(), platform.SpendLog{
		RequestID: originalRequestID, TokenHash: authInfo.TokenHash, TenantID: currentTenant.ID,
		OrganizationID: authInfo.OrganizationID, TeamID: authInfo.TeamID, UserID: authInfo.UserID, OrgUnitID: authInfo.OrgUnitID, Model: model,
		InputTokens: finalCost.InputTokens, OutputTokens: finalCost.OutputTokens, TotalTokens: finalCost.TotalTokens,
		CostUSD: finalCost.TotalCostUSD, CostIRR: finalCost.TotalCostIRR, ResponseTimeMS: responseTimeMS,
		Status: "success", RequestMetadata: enrichPricingMetadata(tunnelReq.RequestMetadata, authInfo, model),
		UsageBreakdown: &finalCost, PricingSnapshot: finalCost.PricingSnapshot,
		CompletionStart: time.Now().UTC(), CompletionEnd: time.Now().UTC(),
	})

	forwardProviderHeaders(c, resp.Headers)
	setTraceHeaders(c, requestID)
	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	return c.Status(resp.StatusCode).Send(resp.Body)
}

func (h *ChatCompletionHandler) resolveXAIDeferredDeployment(ctx context.Context, tenantID, accountID string) (platform.Deployment, error) {
	deployments, err := h.deployments.ListDeployments(ctx, tenantID)
	if err != nil {
		return platform.Deployment{}, fmt.Errorf("deferred completion deployment routing is unavailable")
	}

	var found *platform.Deployment

	for i := range deployments {
		deployment := deployments[i]
		if !strings.EqualFold(deployment.Provider, "xai") || !strings.EqualFold(deployment.ProviderAccountID, accountID) {
			continue
		}

		if found != nil {
			return platform.Deployment{}, fmt.Errorf("deferred completion provider account is ambiguous")
		}

		found = &deployment
	}

	if found == nil {
		return platform.Deployment{}, fmt.Errorf("deferred completion provider account is unavailable")
	}

	return *found, nil
}

func reservationID(reservation *platform.BudgetReservation) string {
	if reservation == nil {
		return ""
	}

	return strings.TrimSpace(reservation.ID)
}
