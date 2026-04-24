package proxy

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/Delpak/shahrah/internal/auth"
	"github.com/Delpak/shahrah/internal/cache"
	"github.com/Delpak/shahrah/internal/providers"
	"github.com/Delpak/shahrah/internal/providers/openai"
	"github.com/Delpak/shahrah/internal/router"
	"github.com/Delpak/shahrah/pkg/types"
)

// ChatCompletionHandler handles chat completion requests.
type ChatCompletionHandler struct {
	llmRouter *router.Router
	cache     cache.Cache
	logger    *zap.Logger
}

// NewChatCompletionHandler creates a new handler.
func NewChatCompletionHandler(llmRouter *router.Router, c cache.Cache, logger *zap.Logger) *ChatCompletionHandler {
	return &ChatCompletionHandler{
		llmRouter: llmRouter,
		cache:     c,
		logger:    logger,
	}
}

// Handle processes a chat completion request.
func (h *ChatCompletionHandler) Handle(c *fiber.Ctx) error {
	startTime := time.Now()

	// Get auth info
	keyInfo := auth.GetAuthInfo(c)
	if keyInfo == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": map[string]interface{}{
				"message": "Authentication required",
				"type":    "authentication_error",
			},
		})
	}

	// Parse request body
	var req types.ChatCompletionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": map[string]interface{}{
				"message": "Invalid request body",
				"type":    "invalid_request_error",
			},
		})
	}

	h.logger.Info("Received chat completion request",
		zap.String("model", req.Model),
		zap.Bool("stream", req.Stream),
		zap.Int("messages", len(req.Messages)),
	)

	// Route to a deployment
	deployment, err := h.llmRouter.Route(c.Context(), req.Model)
	if err != nil {
		h.logger.Error("Failed to route request", zap.Error(err))
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": map[string]interface{}{
				"message": fmt.Sprintf("No available deployment for model: %s", req.Model),
				"type":    "routing_error",
			},
		})
	}

	h.logger.Info("Routed to deployment",
		zap.String("deployment_id", deployment.ID),
		zap.String("provider", deployment.Provider),
		zap.String("model", deployment.Model),
	)

	// Create provider client
	providerClient := h.createProvider(deployment)

	// Execute the request
	if req.Stream {
		return h.handleStream(c, c.Context(), providerClient, deployment, keyInfo, startTime)
	}

	return h.handleSync(c, c.Context(), providerClient, deployment, req, keyInfo, startTime)
}

func (h *ChatCompletionHandler) handleSync(c *fiber.Ctx, ctx context.Context, provider providers.Provider, deployment *router.Deployment, req types.ChatCompletionRequest, keyInfo *auth.KeyInfo, startTime time.Time) error {
	resp, err := provider.Complete(ctx, req)
	if err != nil {
		h.llmRouter.MarkFailed(ctx, deployment, err.Error())
		h.logger.Error("Provider request failed", zap.Error(err))
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error": map[string]interface{}{
				"message": fmt.Sprintf("Provider error: %v", err),
				"type":    "api_error",
			},
		})
	}

	// Set response headers
	c.Set("x-litellm-model", deployment.Model)
	c.Set("x-litellm-provider", deployment.Provider)
	c.Set("x-litellm-deployment-id", deployment.ID)

	// Log response time
	responseTime := time.Since(startTime)
	h.logger.Info("Request completed",
		zap.String("deployment_id", deployment.ID),
		zap.Duration("response_time", responseTime),
		zap.Int("prompt_tokens", resp.Usage.PromptTokens),
		zap.Int("completion_tokens", resp.Usage.CompletionTokens),
	)

	return c.JSON(resp)
}

func (h *ChatCompletionHandler) handleStream(c *fiber.Ctx, ctx context.Context, provider providers.Provider, deployment *router.Deployment, keyInfo *auth.KeyInfo, startTime time.Time) error {
	// Set streaming headers
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("x-litellm-model", deployment.Model)
	c.Set("x-litellm-provider", deployment.Provider)

	// Create a streaming request
	req := types.ChatCompletionRequest{
		Stream: true,
		// Would populate from original request in real implementation
	}

	chunkCh, errCh, err := provider.StreamComplete(ctx, req)
	if err != nil {
		h.llmRouter.MarkFailed(ctx, deployment, err.Error())
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error": map[string]interface{}{
				"message": fmt.Sprintf("Streaming error: %v", err),
				"type":    "api_error",
			},
		})
	}

	// Write SSE stream
	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		for {
			select {
			case chunk, ok := <-chunkCh:
				if !ok {
					// Stream ended
					w.WriteString("data: [DONE]\n\n")
					w.Flush()
					return
				}
				data, _ := json.Marshal(chunk)
				w.WriteString(fmt.Sprintf("data: %s\n\n", data))
				w.Flush()

			case err, ok := <-errCh:
				if ok && err != nil {
					h.llmRouter.MarkFailed(ctx, deployment, err.Error())
					errorData, _ := json.Marshal(types.APIError{
						Error: types.ErrorInfo{
							Message: err.Error(),
							Type:    "streaming_error",
						},
					})
					w.WriteString(fmt.Sprintf("data: %s\n\n", errorData))
					w.Flush()
				}
				return

			case <-ctx.Done():
				return
			}
		}
	})

	return nil
}

func (h *ChatCompletionHandler) createProvider(deployment *router.Deployment) providers.Provider {
	cfg := providers.ProviderConfig{
		APIKey:     deployment.Params.APIKey,
		APIBase:    deployment.Params.APIBase,
		APIVersion: deployment.Params.APIVersion,
		Region:     deployment.Params.RegionName,
		Timeout:    deployment.Params.Timeout,
	}

	switch deployment.Provider {
	case "openai":
		return openai.NewClient(cfg)
	default:
		// Default to OpenAI-compatible for unknown providers
		return openai.NewClient(cfg)
	}
}

