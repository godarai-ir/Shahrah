package proxy

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	"github.com/Delpak/shahrah/services/control-plane/internal/billing"
	"github.com/Delpak/shahrah/services/control-plane/internal/router"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type budgeter interface {
	EstimateCost(ctx context.Context, req contractsopenai.ChatCompletionRequest) (platform.UsageCostBreakdown, error)
	CheckRateLimit(ctx context.Context, auth *platform.AuthContext, model string, estimatedTokens int) error
	CheckBudget(ctx context.Context, auth *platform.AuthContext, breakdown platform.UsageCostBreakdown) error
	FinalizeFromUsage(ctx context.Context, model string, usage *contractsopenai.Usage, estimated platform.UsageCostBreakdown) (platform.UsageCostBreakdown, error)
}

type walletStore interface {
	DebitWallet(ctx context.Context, tenantID string, breakdown platform.UsageCostBreakdown, requestID string) (*platform.Transaction, error)
}

type usageLogger interface {
	LogRequest(ctx context.Context, entry platform.RequestLogEntry) error
	LogSpend(ctx context.Context, log platform.SpendLog) error
}

// ChatCompletionHandler handles chat completion requests.
type ChatCompletionHandler struct {
	llmRouter *router.Router
	tunnel    platform.TunnelTransport
	billing   budgeter
	wallets   walletStore
	logs      usageLogger
	logger    *zap.Logger
}

// NewChatCompletionHandler creates a new handler.
func NewChatCompletionHandler(llmRouter *router.Router, tunnel platform.TunnelTransport, billing budgeter, wallets walletStore, logs usageLogger, logger *zap.Logger) *ChatCompletionHandler {
	return &ChatCompletionHandler{
		llmRouter: llmRouter,
		tunnel:    tunnel,
		billing:   billing,
		wallets:   wallets,
		logs:      logs,
		logger:    logger,
	}
}

// Handle processes a chat completion request.
func (h *ChatCompletionHandler) Handle(c *fiber.Ctx) error {
	startTime := time.Now()
	authInfo := auth.GetAuthInfo(c)
	if authInfo == nil {
		return writeAPIError(c, fiber.StatusUnauthorized, "Authentication required", "authentication_error")
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return writeAPIError(c, fiber.StatusBadRequest, "Tenant resolution required", "invalid_request_error")
	}
	authInfo.TenantID = currentTenant.ID

	var req contractsopenai.ChatCompletionRequest
	if err := c.BodyParser(&req); err != nil {
		return writeAPIError(c, fiber.StatusBadRequest, "Invalid request body", "invalid_request_error")
	}
	if req.Model == "" {
		return writeAPIError(c, fiber.StatusBadRequest, "Model is required", "invalid_request_error")
	}
	if len(req.Messages) == 0 {
		return writeAPIError(c, fiber.StatusBadRequest, "messages array must not be empty", "invalid_request_error")
	}
	for i, msg := range req.Messages {
		if msg.Role == "" {
			return writeAPIError(c, fiber.StatusBadRequest, fmt.Sprintf("message[%d].role is required", i), "invalid_request_error")
		}
	}
	if req.Temperature != nil && (*req.Temperature < 0 || *req.Temperature > 2) {
		return writeAPIError(c, fiber.StatusBadRequest, "temperature must be between 0 and 2", "invalid_request_error")
	}
	if req.TopP != nil && (*req.TopP < 0 || *req.TopP > 1) {
		return writeAPIError(c, fiber.StatusBadRequest, "top_p must be between 0 and 1", "invalid_request_error")
	}
	if req.MaxTokens != nil && *req.MaxTokens < 1 {
		return writeAPIError(c, fiber.StatusBadRequest, "max_tokens must be at least 1", "invalid_request_error")
	}
	if len(authInfo.AllowedModels) > 0 && !containsModel(authInfo.AllowedModels, req.Model) && !containsModel(authInfo.AllowedModels, "*") {
		return writeAPIError(c, fiber.StatusForbidden, "Model access denied", "authorization_error")
	}

	estimated, err := h.billing.EstimateCost(c.Context(), req)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadGateway, "Failed to estimate request cost", "billing_error")
	}

	if err := h.billing.CheckRateLimit(c.Context(), authInfo, req.Model, estimated.TotalTokens); err != nil {
		status := fiber.StatusTooManyRequests
		if err == billing.ErrBudgetExceeded {
			status = fiber.StatusPaymentRequired
		}
		return writeAPIError(c, status, err.Error(), "rate_limit_error")
	}
	if err := h.billing.CheckBudget(c.Context(), authInfo, estimated); err != nil {
		return writeAPIError(c, fiber.StatusPaymentRequired, err.Error(), "budget_exceeded")
	}

	deployment, err := h.llmRouter.Route(c.Context(), currentTenant.ID, req.Model)
	if err != nil {
		return writeAPIError(c, fiber.StatusServiceUnavailable, fmt.Sprintf("No available deployment for model: %s", req.Model), "routing_error")
	}

	if err := h.tunnel.Healthy(c.Context()); err != nil {
		return writeAPIError(c, fiber.StatusServiceUnavailable, "Tunnel path unavailable", "tunnel_unavailable")
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadRequest, "Failed to encode request", "invalid_request_error")
	}

	requestID := c.GetRespHeader("x-request-id")
	if requestID == "" {
		requestID = fmt.Sprintf("req_%d", time.Now().UnixNano())
	}

	tunnelReq := platform.TunnelRequest{
		RequestID:  requestID,
		TenantID:   currentTenant.ID,
		TraceID:    requestID,
		Stream:     req.Stream,
		Deployment: toContractDeployment(deployment),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: payload,
		RequestMetadata: map[string]interface{}{
			"tenant_slug": currentTenant.Slug,
			"model":       req.Model,
		},
	}

	if req.Stream {
		return h.handleStream(c, req, authInfo, estimated, tunnelReq, startTime)
	}
	return h.handleSync(c, req, authInfo, estimated, tunnelReq, startTime)
}

func (h *ChatCompletionHandler) handleSync(c *fiber.Ctx, req contractsopenai.ChatCompletionRequest, authInfo *platform.AuthContext, estimated platform.UsageCostBreakdown, tunnelReq platform.TunnelRequest, start time.Time) error {
	resp, err := h.tunnel.Proxy(c.Context(), tunnelReq)
	if err != nil {
		h.llmRouter.MarkFailed(c.Context(), &router.Deployment{ID: tunnelReq.Deployment.ID}, err.Error())
		return writeAPIError(c, fiber.StatusBadGateway, fmt.Sprintf("Tunnel error: %v", err), "api_error")
	}

	if resp.StatusCode >= httpErrorFloor {
		return transformProviderError(c, resp.StatusCode, resp.Body, resp.Headers)
	}

	var completion contractsopenai.ChatCompletionResponse
	if err := json.Unmarshal(resp.Body, &completion); err != nil {
		return writeAPIError(c, fiber.StatusBadGateway, "Failed to decode provider response", "api_error")
	}

	finalCost, err := h.billing.FinalizeFromUsage(c.Context(), req.Model, completion.Usage, estimated)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadGateway, "Failed to finalize usage", "billing_error")
	}
	if _, err := h.wallets.DebitWallet(c.Context(), authInfo.TenantID, finalCost, tunnelReq.RequestID); err != nil {
		return writeAPIError(c, fiber.StatusPaymentRequired, err.Error(), "budget_exceeded")
	}

	responseTimeMS := int(time.Since(start).Milliseconds())
	_ = h.logs.LogRequest(c.Context(), platform.RequestLogEntry{
		RequestID:      tunnelReq.RequestID,
		TokenHash:      authInfo.TokenHash,
		TenantID:       authInfo.TenantID,
		Model:          req.Model,
		Status:         "success",
		CostUSD:        finalCost.TotalCostUSD,
		ResponseTimeMS: responseTimeMS,
		RequestBody:    map[string]interface{}{"model": req.Model},
		ResponseBody:   map[string]interface{}{"id": completion.ID},
	})
	_ = h.logs.LogSpend(c.Context(), platform.SpendLog{
		RequestID:       tunnelReq.RequestID,
		TokenHash:       authInfo.TokenHash,
		TenantID:        authInfo.TenantID,
		OrganizationID:  authInfo.OrganizationID,
		TeamID:          authInfo.TeamID,
		Model:           req.Model,
		InputTokens:     finalCost.InputTokens,
		OutputTokens:    finalCost.OutputTokens,
		TotalTokens:     finalCost.TotalTokens,
		CostUSD:         finalCost.TotalCostUSD,
		CostIRR:         finalCost.TotalCostIRR,
		ResponseTimeMS:  responseTimeMS,
		Status:          "success",
		RequestMetadata: tunnelReq.RequestMetadata,
		CompletionStart: start,
		CompletionEnd:   time.Now(),
	})

	// Forward allowed provider response headers
	forwardProviderHeaders(c, resp.Headers)

	c.Set("x-shahrah-trace-id", tunnelReq.TraceID)
	c.Set("x-shahrah-cost-usd", fmt.Sprintf("%.8f", finalCost.TotalCostUSD))
	c.Set("x-shahrah-cost-irr", fmt.Sprintf("%.2f", finalCost.TotalCostIRR))
	return c.Status(resp.StatusCode).JSON(completion)
}

func (h *ChatCompletionHandler) handleStream(c *fiber.Ctx, req contractsopenai.ChatCompletionRequest, authInfo *platform.AuthContext, estimated platform.UsageCostBreakdown, tunnelReq platform.TunnelRequest, start time.Time) error {
	stream, err := h.tunnel.Stream(c.Context(), tunnelReq)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadGateway, fmt.Sprintf("Tunnel error: %v", err), "api_error")
	}

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("x-shahrah-trace-id", tunnelReq.TraceID)

	// Forward provider headers from stream
	forwardProviderHeaders(c, stream.Headers())

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		defer stream.Close()

		for {
			select {
			case chunk, ok := <-stream.Chunks():
				if !ok {
					finalCost, _ := h.billing.FinalizeFromUsage(c.Context(), req.Model, nil, estimated)
					_, _ = h.wallets.DebitWallet(c.Context(), authInfo.TenantID, finalCost, tunnelReq.RequestID)
					_ = h.logs.LogSpend(c.Context(), platform.SpendLog{
						RequestID:       tunnelReq.RequestID,
						TokenHash:       authInfo.TokenHash,
						TenantID:        authInfo.TenantID,
						OrganizationID:  authInfo.OrganizationID,
						TeamID:          authInfo.TeamID,
						Model:           req.Model,
						InputTokens:     finalCost.InputTokens,
						OutputTokens:    finalCost.OutputTokens,
						TotalTokens:     finalCost.TotalTokens,
						CostUSD:         finalCost.TotalCostUSD,
						CostIRR:         finalCost.TotalCostIRR,
						ResponseTimeMS:  int(time.Since(start).Milliseconds()),
						Status:          "success",
						RequestMetadata: tunnelReq.RequestMetadata,
						CompletionStart: start,
						CompletionEnd:   time.Now(),
					})
					return
				}
				if len(chunk.Data) > 0 {
					_, _ = w.Write(chunk.Data)
					_ = w.Flush()
				}
				if chunk.IsDone {
					return
				}
			case err, ok := <-stream.Errors():
				if ok && err != nil {
					errorPayload, _ := json.Marshal(contractsopenai.APIError{
						Error: contractsopenai.ErrorInfo{
							Message: err.Error(),
							Type:    "streaming_error",
						},
					})
					_, _ = w.Write(bytes.Join([][]byte{[]byte("data: "), errorPayload, []byte("\n\n")}, nil))
					_ = w.Flush()
				}
				return
			case <-c.Context().Done():
				return
			}
		}
	})

	return nil
}

const httpErrorFloor = 400

func toContractDeployment(dep *router.Deployment) platform.Deployment {
	return platform.Deployment{
		ID:         dep.ID,
		TenantID:   dep.TenantID,
		ModelName:  dep.ModelName,
		Provider:   dep.Provider,
		Model:      dep.Model,
		APIBase:    dep.Params.APIBase,
		APIKey:     dep.Params.APIKey,
		APIVersion: dep.Params.APIVersion,
		RegionName: dep.Params.RegionName,
		RPM:        dep.Params.RPM,
		TPM:        dep.Params.TPM,
		Weight:     dep.Weight,
		TimeoutSec: dep.Params.Timeout,
	}
}

func containsModel(models []string, model string) bool {
	for _, candidate := range models {
		if candidate == model {
			return true
		}
	}
	return false
}

func writeAPIError(c *fiber.Ctx, status int, message string, errorType string) error {
	return c.Status(status).JSON(contractsopenai.APIError{
		Error: contractsopenai.ErrorInfo{
			Message: message,
			Type:    errorType,
			Code:    fmt.Sprintf("%d", status),
		},
	})
}

// transformProviderError converts non-OpenAI provider errors to OpenAI-compatible format.
func transformProviderError(c *fiber.Ctx, statusCode int, body []byte, headers map[string]string) error {
	// Forward provider headers even on error
	forwardProviderHeaders(c, headers)

	// Try to parse as OpenAI error first
	var apiErr contractsopenai.APIError
	if err := json.Unmarshal(body, &apiErr); err == nil && apiErr.Error.Message != "" {
		return c.Status(statusCode).JSON(apiErr)
	}

	// Fall back to generic OpenAI-compatible error
	message := string(body)
	if len(message) > 500 {
		message = message[:500] + "..."
	}
	if message == "" {
		message = http.StatusText(statusCode)
	}

	errorType := "api_error"
	switch {
	case statusCode == http.StatusTooManyRequests:
		errorType = "rate_limit_error"
	case statusCode == http.StatusUnauthorized || statusCode == http.StatusForbidden:
		errorType = "authentication_error"
	case statusCode >= http.StatusInternalServerError:
		errorType = "server_error"
	}

	return c.Status(statusCode).JSON(contractsopenai.APIError{
		Error: contractsopenai.ErrorInfo{
			Message: message,
			Type:    errorType,
			Code:    fmt.Sprintf("%d", statusCode),
		},
	})
}

// forwardProviderHeaders forwards allowed provider response headers to the client.
func forwardProviderHeaders(c *fiber.Ctx, headers map[string]string) {
	allowedPrefixes := []string{
		"x-ratelimit-",
		"x-request-id",
		"x-trace-id",
		"x-correlation-id",
		"cf-ray",
		"traceparent",
	}

	for key, val := range headers {
		if isAllowedHeader(key, allowedPrefixes) {
			c.Set(key, val)
		}
	}
}

// isAllowedHeader checks if a header key matches allowed prefixes (case-insensitive).
func isAllowedHeader(key string, prefixes []string) bool {
	lower := strings.ToLower(key)
	for _, prefix := range prefixes {
		if strings.HasPrefix(lower, prefix) {
			return true
		}
	}
	return false
}
