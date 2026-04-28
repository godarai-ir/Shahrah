package proxy

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	"github.com/Delpak/shahrah/services/control-plane/internal/router"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// CompletionsHandler handles legacy text completion requests.
type CompletionsHandler struct {
	llmRouter *router.Router
	tunnel    platform.TunnelTransport
	billing   budgeter
	wallets   walletStore
	logs      usageLogger
	logger    *zap.Logger
}

// NewCompletionsHandler creates a new legacy completions handler.
func NewCompletionsHandler(llmRouter *router.Router, tunnel platform.TunnelTransport, billing budgeter, wallets walletStore, logs usageLogger, logger *zap.Logger) *CompletionsHandler {
	return &CompletionsHandler{
		llmRouter: llmRouter,
		tunnel:    tunnel,
		billing:   billing,
		wallets:   wallets,
		logs:      logs,
		logger:    logger,
	}
}

// Handle processes a legacy text completion request.
func (h *CompletionsHandler) Handle(c *fiber.Ctx) error {
	startTime := time.Now()

	var req contractsopenai.CompletionRequest
	if err := c.BodyParser(&req); err != nil {
		return writeAPIError(c, fiber.StatusBadRequest, "invalid request body", "invalid_request_error")
	}

	// Validate required fields
	if req.Model == "" {
		return writeAPIError(c, fiber.StatusBadRequest, "model is required", "invalid_request_error")
	}
	if req.Prompt == "" {
		return writeAPIError(c, fiber.StatusBadRequest, "prompt is required", "invalid_request_error")
	}

	// Get auth context from middleware
	authInfo := auth.GetAuthInfo(c)
	if authInfo == nil {
		return writeAPIError(c, fiber.StatusUnauthorized, "authentication required", "authentication_error")
	}

	// Get tenant context
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return writeAPIError(c, fiber.StatusBadRequest, "tenant resolution required", "invalid_request_error")
	}
	authInfo.TenantID = currentTenant.ID

	if len(authInfo.AllowedModels) > 0 && !containsModel(authInfo.AllowedModels, req.Model) && !containsModel(authInfo.AllowedModels, "*") {
		return writeAPIError(c, fiber.StatusForbidden, "model access denied", "authorization_error")
	}

	// Transform to chat completion format internally
	chatReq := contractsopenai.ChatCompletionRequest{
		Model:       req.Model,
		Messages:    []contractsopenai.ChatMessage{{Role: "user", Content: req.Prompt}},
		Stream:      req.Stream,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		TopP:        req.TopP,
	}

	// Estimate cost using the same billing logic
	estimated, err := h.billing.EstimateCost(c.Context(), chatReq)
	if err != nil {
		h.logger.Error("failed to estimate cost", zap.Error(err))
		estimated = platform.UsageCostBreakdown{
			Model:        req.Model,
			TotalTokens:  512,
			TotalCostUSD: 0.001,
			TotalCostIRR: 850,
		}
	}

	// Check rate limit
	rateLimit, err := h.billing.CheckRateLimit(c.Context(), authInfo, req.Model, estimated.TotalTokens)
	setRateLimitHeaders(c, rateLimit)
	if err != nil {
		return writeAPIError(c, fiber.StatusTooManyRequests, "rate limit exceeded", "rate_limit_error")
	}

	// Check budget
	if err := h.billing.CheckBudget(c.Context(), authInfo, estimated); err != nil {
		return writeAPIError(c, fiber.StatusPaymentRequired, err.Error(), "budget_exceeded")
	}

	// Route to deployment
	dep, err := h.llmRouter.Route(c.Context(), currentTenant.ID, req.Model)
	if err != nil {
		return writeAPIError(c, fiber.StatusInternalServerError, "no available deployments", "internal_error")
	}

	// Build tunnel request
	requestID := c.GetRespHeader("x-request-id")
	if requestID == "" {
		requestID = fmt.Sprintf("req_%d", time.Now().UnixNano())
	}

	tunnelReq := platform.TunnelRequest{
		RequestID: requestID,
		TenantID:  currentTenant.ID,
		TraceID:   requestID,
		Stream:    req.Stream,
		Deployment: platform.Deployment{
			ID:         dep.ID,
			TenantID:   dep.TenantID,
			ModelName:  dep.ModelName,
			Provider:   dep.Provider,
			Model:      dep.Model,
			APIBase:    dep.Params.APIBase,
			APIKey:     dep.Params.APIKey,
			APIVersion: dep.Params.APIVersion,
		},
		Headers: map[string]string{"Authorization": "Bearer " + dep.Params.APIKey},
		Body:    mustMarshalJSON(chatReq),
	}

	// Execute request
	if req.Stream {
		return h.handleStream(c, chatReq, authInfo, estimated, tunnelReq, startTime)
	}
	return h.handleSync(c, chatReq, authInfo, estimated, tunnelReq, startTime)
}

func (h *CompletionsHandler) handleSync(c *fiber.Ctx, req contractsopenai.ChatCompletionRequest, authInfo *platform.AuthContext, estimated platform.UsageCostBreakdown, tunnelReq platform.TunnelRequest, start time.Time) error {
	resp, err := h.tunnel.Proxy(c.Context(), tunnelReq)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadGateway, "provider request failed", "bad_gateway")
	}

	if resp.StatusCode >= httpErrorFloor {
		return transformProviderError(c, resp.StatusCode, resp.Body, resp.Headers)
	}
	forwardProviderHeaders(c, resp.Headers)

	// Transform chat completion response to legacy completion format
	var chatResp contractsopenai.ChatCompletionResponse
	if err := json.Unmarshal(resp.Body, &chatResp); err != nil {
		return c.Status(fiber.StatusOK).Send(resp.Body)
	}

	// Convert to legacy format
	completionResp := contractsopenai.CompletionResponse{
		ID:      chatResp.ID,
		Object:  "text_completion",
		Created: chatResp.Created,
		Model:   chatResp.Model,
	}

	choices := make([]contractsopenai.CompletionChoice, 0, len(chatResp.Choices))
	for _, choice := range chatResp.Choices {
		if choice.Message != nil {
			choices = append(choices, contractsopenai.CompletionChoice{
				Text:         choice.Message.Content,
				Index:        choice.Index,
				FinishReason: choice.FinishReason,
			})
		}
	}
	completionResp.Choices = choices
	completionResp.Usage = chatResp.Usage

	// Finalize billing
	finalCost, _ := h.billing.FinalizeFromUsage(c.Context(), req.Model, chatResp.Usage, estimated)
	if _, err := h.wallets.DebitWallet(c.Context(), authInfo.TenantID, finalCost, tunnelReq.RequestID); err != nil {
		h.logger.Error("failed to debit wallet", zap.Error(err))
	}
	setCostHeaders(c, finalCost)
	c.Set("x-shahrah-trace-id", tunnelReq.RequestID)

	// Log asynchronously
	responseTimeMS := int(time.Since(start).Milliseconds())
	go func() {
		_ = h.logs.LogRequest(context.Background(), platform.RequestLogEntry{
			RequestID:      tunnelReq.RequestID,
			TokenHash:      authInfo.TokenHash,
			TenantID:       authInfo.TenantID,
			Model:          req.Model,
			Status:         fmt.Sprintf("%d", resp.StatusCode),
			CostUSD:        finalCost.TotalCostUSD,
			ResponseTimeMS: responseTimeMS,
		})
	}()

	return c.Status(resp.StatusCode).JSON(completionResp)
}

func (h *CompletionsHandler) handleStream(c *fiber.Ctx, req contractsopenai.ChatCompletionRequest, authInfo *platform.AuthContext, estimated platform.UsageCostBreakdown, tunnelReq platform.TunnelRequest, start time.Time) error {
	stream, err := h.tunnel.Stream(c.Context(), tunnelReq)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadGateway, "stream failed", "bad_gateway")
	}

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("X-Accel-Buffering", "no")
	c.Set("x-shahrah-trace-id", tunnelReq.TraceID)

	// Forward provider headers from stream
	forwardProviderHeaders(c, stream.Headers())

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		defer stream.Close()

		for {
			select {
			case chunk, ok := <-stream.Chunks():
				if !ok {
					// Stream done - finalize billing
					finalCost, _ := h.billing.FinalizeFromUsage(c.Context(), req.Model, nil, estimated)
					_, _ = h.wallets.DebitWallet(c.Context(), authInfo.TenantID, finalCost, tunnelReq.RequestID)
					_ = h.logs.LogSpend(c.Context(), platform.SpendLog{
						RequestID:      tunnelReq.RequestID,
						TokenHash:      authInfo.TokenHash,
						TenantID:       authInfo.TenantID,
						Model:          req.Model,
						InputTokens:    finalCost.InputTokens,
						OutputTokens:   finalCost.OutputTokens,
						TotalTokens:    finalCost.TotalTokens,
						CostUSD:        finalCost.TotalCostUSD,
						CostIRR:        finalCost.TotalCostIRR,
						ResponseTimeMS: int(time.Since(start).Milliseconds()),
						Status:         "success",
					})
					return
				}

				// Transform chat chunk to legacy completion format
				if len(chunk.Data) > 0 {
					raw := bytes.TrimSpace(chunk.Data)
					if bytes.Equal(raw, []byte("data: [DONE]")) || bytes.Equal(raw, []byte("[DONE]")) {
						_, _ = w.Write([]byte("data: [DONE]\n\n"))
						_ = w.Flush()
						return
					}

					normalized := raw
					if bytes.HasPrefix(normalized, []byte("data: ")) {
						normalized = bytes.TrimSpace(bytes.TrimPrefix(normalized, []byte("data: ")))
					}

					var chatChunk contractsopenai.StreamChunk
					if err := json.Unmarshal(normalized, &chatChunk); err != nil {
						_, _ = w.Write(chunk.Data)
						if !bytes.HasSuffix(chunk.Data, []byte("\n\n")) {
							_, _ = w.Write([]byte("\n\n"))
						}
						_ = w.Flush()
						continue
					}

					// Extract text from chat delta
					var text string
					for _, choice := range chatChunk.Choices {
						if choice.Delta != nil {
							text = choice.Delta.Content
							break
						}
					}

					if text != "" {
						legacyChunk := map[string]interface{}{
							"id":      chatChunk.ID,
							"object":  "text_completion",
							"created": chatChunk.Created,
							"model":   chatChunk.Model,
							"choices": []map[string]interface{}{
								{"text": text, "index": 0, "finish_reason": nil},
							},
						}
						data, _ := json.Marshal(legacyChunk)
						_, _ = w.Write([]byte("data: "))
						_, _ = w.Write(data)
						_, _ = w.Write([]byte("\n\n"))
						_ = w.Flush()
					}

					if chunk.IsDone {
						_, _ = w.Write([]byte("data: [DONE]\n\n"))
						_ = w.Flush()
						return
					}
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
			case <-c.Context().Done():
				return
			}
		}
	})

	return nil
}

// setCostHeaders sets cost-related response headers.
func setCostHeaders(c *fiber.Ctx, cost platform.UsageCostBreakdown) {
	c.Set("x-shahrah-cost-usd", fmt.Sprintf("%.8f", cost.TotalCostUSD))
	c.Set("x-shahrah-cost-irr", fmt.Sprintf("%.2f", cost.TotalCostIRR))
}

// mustMarshalJSON marshals to JSON or panics.
func mustMarshalJSON(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}

func normalizePrompt(prompt interface{}) string {
	switch p := prompt.(type) {
	case string:
		return p
	case []string:
		return strings.TrimSpace(strings.Join(p, "\n"))
	case []interface{}:
		parts := make([]string, 0, len(p))
		for _, item := range p {
			if s, ok := item.(string); ok && strings.TrimSpace(s) != "" {
				parts = append(parts, s)
			}
		}
		return strings.TrimSpace(strings.Join(parts, "\n"))
	default:
		return ""
	}
}
