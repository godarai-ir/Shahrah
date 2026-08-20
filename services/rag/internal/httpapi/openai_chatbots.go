package httpapi

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/rag/internal/store"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type openAIChatCompletionRequest struct {
	Model     string              `json:"model"`
	Messages  []openAIChatMessage `json:"messages"`
	Stream    bool                `json:"stream"`
	ChatID    string              `json:"chat_id"`
	SessionID string              `json:"session_id"`
	User      string              `json:"user"`
	Metadata  map[string]any      `json:"metadata"`
}

type openAIChatMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

func (h *handler) previewAnswer(ctx *fiber.Ctx) error {
	chatbot, err := h.dependencies.Store.GetChatbot(ctx.Context(), tenantID(ctx), ctx.Params("id"))
	if err != nil {
		return h.storageError(ctx, err)
	}

	if chatbot.PublishedVersionID == nil {
		return writeError(ctx, fiber.StatusConflict, "chatbot_not_published", "A published chatbot is required")
	}

	published, err := h.dependencies.Store.GetPublishedChatbot(ctx.Context(), chatbot.PublicID)
	if err != nil {
		return h.storageError(ctx, err)
	}

	var request platform.ChatbotResponseRequest

	bodyErr := ctx.BodyParser(&request)
	if bodyErr != nil || strings.TrimSpace(request.Message) == "" {
		return writeError(ctx, fiber.StatusBadRequest, "invalid_request", "message is required")
	}

	if validationErr := validateChatMessage(request.Message); validationErr != nil {
		return writeError(ctx, fiber.StatusBadRequest, "invalid_request", validationErr.Error())
	}

	result, err := h.dependencies.Application.Answer(ctx.Context(), *published, request)
	if err != nil {
		return h.answerError(ctx, err)
	}

	return ctx.JSON(result)
}

func (h *handler) openAIModels(ctx *fiber.Ctx) error {
	chatbot, err := h.authorizedOpenAIChatbot(ctx)
	if err != nil {
		return err
	}

	return ctx.JSON(fiber.Map{
		"object": "list",
		"data": []fiber.Map{{
			"id":       chatbot.Chatbot.PublicID,
			"object":   "model",
			"created":  chatbot.Version.PublishedAt.Unix(),
			"owned_by": "godarai",
			"name":     chatbot.Chatbot.Name,
		}},
	})
}

func (h *handler) openAIChatCompletions(ctx *fiber.Ctx) error {
	chatbot, err := h.authorizedOpenAIChatbot(ctx)
	if err != nil {
		return err
	}

	var request openAIChatCompletionRequest

	bodyErr := ctx.BodyParser(&request)
	if bodyErr != nil {
		return writeOpenAIError(ctx, fiber.StatusBadRequest, "invalid_request_error", "Invalid request body")
	}

	if model := strings.TrimSpace(request.Model); model != "" && model != chatbot.Chatbot.PublicID {
		return writeOpenAIError(ctx, fiber.StatusNotFound, "invalid_request_error", "The requested chatbot model was not found")
	}

	message, err := lastOpenAIUserMessage(request.Messages)
	if err != nil {
		return writeOpenAIError(ctx, fiber.StatusBadRequest, "invalid_request_error", err.Error())
	}

	if validationErr := validateChatMessage(message); validationErr != nil {
		return writeOpenAIError(ctx, fiber.StatusBadRequest, "invalid_request_error", validationErr.Error())
	}

	sessionID := openAISessionID(request)
	if chatbot.Chatbot.AccessMode == "public" {
		allowed, retryAfter, limitErr := h.dependencies.Store.AllowPublicChatbotRequest(
			ctx.Context(), chatbot.Chatbot.PublicID, ctx.Get(headerClientIP), sessionID,
		)
		if limitErr != nil {
			return writeOpenAIError(ctx, fiber.StatusServiceUnavailable, "api_error", "Public chatbot admission unavailable")
		}

		if !allowed {
			ctx.Set(fiber.HeaderRetryAfter, strconv.Itoa(max(1, int(retryAfter.Seconds()))))

			return writeOpenAIError(ctx, fiber.StatusTooManyRequests, "rate_limit_error", "Public chatbot rate limit exceeded")
		}
	}

	result, err := h.dependencies.Application.Answer(ctx.Context(), *chatbot, platform.ChatbotResponseRequest{
		Message:   message,
		SessionID: sessionID,
	})
	if err != nil {
		return h.openAIAnswerError(ctx, err)
	}

	completionID := "chatcmpl_" + strings.ReplaceAll(uuid.NewString(), "-", "")

	content := answerWithSources(result.Answer, result.Citations)
	if request.Stream {
		return streamOpenAICompletion(ctx, completionID, chatbot.Chatbot.PublicID, content, result)
	}

	return ctx.JSON(fiber.Map{
		"id": completionID, "object": "chat.completion", "created": time.Now().Unix(),
		"model": chatbot.Chatbot.PublicID,
		"choices": []fiber.Map{{
			"index": 0, "message": fiber.Map{"role": "assistant", "content": content}, "finish_reason": "stop",
		}},
		"usage":     fiber.Map{"prompt_tokens": 0, "completion_tokens": 0, "total_tokens": 0},
		"citations": result.Citations,
		"godarai":   openAIGodarMetadata(result),
	})
}

func (h *handler) authorizedOpenAIChatbot(ctx *fiber.Ctx) (*store.PublishedChatbot, error) {
	chatbot, err := h.dependencies.Store.GetPublishedChatbot(ctx.Context(), strings.TrimSpace(ctx.Params("publicID")))
	if err != nil {
		return nil, h.storageError(ctx, err)
	}

	if err := h.authorizeRuntime(ctx, chatbot); err != nil {
		return nil, writeOpenAIError(ctx, fiber.StatusUnauthorized, "authentication_error", err.Error())
	}

	if !originAllowed(ctx.Get(fiber.HeaderOrigin), chatbot.Chatbot.AllowedOrigins) {
		return nil, writeOpenAIError(ctx, fiber.StatusForbidden, "permission_error", "Origin is not allowed for this chatbot")
	}

	return chatbot, nil
}

func (h *handler) answerError(ctx *fiber.Ctx, err error) error {
	if errors.Is(err, store.ErrEntitlementDenied) {
		return writeError(ctx, fiber.StatusTooManyRequests, "rag_query_quota_exceeded", "Managed RAG query quota exceeded")
	}

	if isRAGDependencyError(err) {
		return writeError(ctx, fiber.StatusServiceUnavailable, "rag_dependency_unavailable", "RAG dependency unavailable")
	}

	return h.storageError(ctx, err)
}

func (h *handler) openAIAnswerError(ctx *fiber.Ctx, err error) error {
	if errors.Is(err, store.ErrEntitlementDenied) {
		return writeOpenAIError(ctx, fiber.StatusTooManyRequests, "rate_limit_error", "Managed RAG query quota exceeded")
	}

	if isRAGDependencyError(err) {
		return writeOpenAIError(ctx, fiber.StatusServiceUnavailable, "api_error", "RAG dependency unavailable")
	}

	return writeOpenAIError(ctx, fiber.StatusInternalServerError, "api_error", "Chatbot response failed")
}

func isRAGDependencyError(err error) bool {
	message := err.Error()

	return strings.Contains(message, "retrieval") || strings.Contains(message, "rerank") ||
		strings.Contains(message, "generation")
}

func lastOpenAIUserMessage(messages []openAIChatMessage) (string, error) {
	for index := len(messages) - 1; index >= 0; index-- {
		if strings.EqualFold(strings.TrimSpace(messages[index].Role), "user") {
			content := openAITextContent(messages[index].Content)
			if content != "" {
				return content, nil
			}
		}
	}

	return "", errors.New("messages must contain a non-empty user message")
}

func openAITextContent(content any) string {
	switch value := content.(type) {
	case string:
		return strings.TrimSpace(value)
	case []any:
		parts := make([]string, 0, len(value))
		for _, item := range value {
			part, ok := item.(map[string]any)
			if !ok || part["type"] != "text" {
				continue
			}

			text, ok := part["text"].(string)
			if ok && strings.TrimSpace(text) != "" {
				parts = append(parts, strings.TrimSpace(text))
			}
		}

		return strings.Join(parts, "\n")
	default:
		return ""
	}
}

func openAISessionID(request openAIChatCompletionRequest) string {
	candidates := []string{request.ChatID, request.SessionID}
	for _, key := range []string{"chat_id", "conversation_id", "session_id"} {
		if value, ok := request.Metadata[key].(string); ok {
			candidates = append(candidates, value)
		}
	}

	candidates = append(candidates, request.User)

	for _, candidate := range candidates {
		candidate = strings.TrimSpace(candidate)
		if candidate != "" && len(candidate) <= 200 {
			return "openai_" + candidate
		}
	}

	return ""
}

func answerWithSources(answer string, citations []platform.Citation) string {
	if len(citations) == 0 {
		return answer
	}

	var builder strings.Builder
	builder.WriteString(strings.TrimSpace(answer))
	builder.WriteString("\n\n---\n\n### Sources\n")

	for index, citation := range citations {
		builder.WriteString("\n- [")
		builder.WriteString(strconv.Itoa(index + 1))
		builder.WriteString("] ")
		builder.WriteString(citation.Title)

		if citation.Page != nil {
			builder.WriteString(" — page ")
			builder.WriteString(strconv.Itoa(*citation.Page))
		}
	}

	return builder.String()
}

func streamOpenAICompletion(
	ctx *fiber.Ctx,
	completionID string,
	model string,
	content string,
	result *platform.ChatbotResponse,
) error {
	ctx.Set(fiber.HeaderContentType, "text/event-stream")
	ctx.Set(fiber.HeaderCacheControl, "no-cache, no-transform")
	ctx.Set(fiber.HeaderConnection, "keep-alive")
	ctx.Context().SetBodyStreamWriter(func(writer *bufio.Writer) {
		chunks := []fiber.Map{
			openAIStreamChunk(completionID, model, fiber.Map{"role": "assistant"}, nil),
			openAIStreamChunk(completionID, model, fiber.Map{"content": content}, nil),
			openAIStreamChunk(completionID, model, fiber.Map{}, stringPointer("stop")),
		}

		for index, chunk := range chunks {
			if index == len(chunks)-1 {
				chunk["citations"] = result.Citations
				chunk["godarai"] = openAIGodarMetadata(result)
			}

			if err := writeOpenAIStreamEvent(writer, chunk); err != nil {
				return
			}
		}

		if _, err := writer.WriteString("data: [DONE]\n\n"); err != nil {
			return
		}

		if err := writer.Flush(); err != nil {
			return
		}
	})

	return nil
}

func openAIStreamChunk(completionID, model string, delta fiber.Map, finishReason *string) fiber.Map {
	return fiber.Map{
		"id": completionID, "object": "chat.completion.chunk", "created": time.Now().Unix(), "model": model,
		"choices": []fiber.Map{{"index": 0, "delta": delta, "finish_reason": finishReason}},
	}
}

func writeOpenAIStreamEvent(writer *bufio.Writer, value any) error {
	payload, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("encode OpenAI stream event: %w", err)
	}

	if _, err := fmt.Fprintf(writer, "data: %s\n\n", payload); err != nil {
		return fmt.Errorf("write OpenAI stream event: %w", err)
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("flush OpenAI stream event: %w", err)
	}

	return nil
}

func openAIGodarMetadata(result *platform.ChatbotResponse) fiber.Map {
	return fiber.Map{
		"request_id": result.RequestID, "trace_id": result.TraceID, "session_id": result.SessionID,
		"refusal": result.Refusal, "actual_model": result.ActualModel,
	}
}

func writeOpenAIError(ctx *fiber.Ctx, status int, errorType, message string) error {
	return ctx.Status(status).JSON(fiber.Map{"error": fiber.Map{
		"message": message, "type": errorType, "param": nil, "code": nil,
	}})
}

func stringPointer(value string) *string {
	return &value
}
