package httpapi

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/rag/internal/sessiontoken"
	"github.com/Delpak/shahrah/services/rag/internal/store"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func (h *handler) createChatbot(ctx *fiber.Ctx) error {
	request, parseErr := parseChatbotRequest(ctx)
	if parseErr != nil {
		return writeError(ctx, fiber.StatusBadRequest, "invalid_request", parseErr.Error())
	}

	capacityErr := h.dependencies.Store.RequireChatbotCapacity(ctx.Context(), tenantID(ctx))
	if capacityErr != nil {
		return h.entitlementError(ctx, capacityErr)
	}

	knowledgeBaseErr := h.dependencies.Store.RequireKnowledgeBasesPerBot(ctx.Context(), tenantID(ctx), len(request.KnowledgeBaseIDs))
	if knowledgeBaseErr != nil {
		return h.entitlementError(ctx, knowledgeBaseErr)
	}

	if request.AccessMode == "public" {
		publicBotErr := h.dependencies.Store.RequirePublicBots(ctx.Context(), tenantID(ctx))
		if publicBotErr != nil {
			return h.entitlementError(ctx, publicBotErr)
		}
	}

	result, err := h.dependencies.Store.CreateChatbot(ctx.Context(), tenantID(ctx), actorID(ctx), request)
	if err != nil {
		return h.storageError(ctx, err)
	}

	if err := h.audit(ctx, "chatbot.create", "chatbot", result.ID, map[string]any{"access_mode": result.AccessMode}); err != nil {
		return h.storageError(ctx, err)
	}

	return ctx.Status(fiber.StatusCreated).JSON(result)
}

func (h *handler) listChatbots(ctx *fiber.Ctx) error {
	limit, offset := pagination(ctx)

	result, err := h.dependencies.Store.ListChatbots(ctx.Context(), tenantID(ctx), limit, offset)
	if err != nil {
		return h.storageError(ctx, err)
	}

	return ctx.JSON(result)
}

func (h *handler) getChatbot(ctx *fiber.Ctx) error {
	result, err := h.dependencies.Store.GetChatbot(ctx.Context(), tenantID(ctx), ctx.Params("id"))
	if err != nil {
		return h.storageError(ctx, err)
	}

	return ctx.JSON(result)
}

func (h *handler) updateChatbot(ctx *fiber.Ctx) error {
	request, parseErr := parseChatbotRequest(ctx)
	if parseErr != nil {
		return writeError(ctx, fiber.StatusBadRequest, "invalid_request", parseErr.Error())
	}

	knowledgeBaseErr := h.dependencies.Store.RequireKnowledgeBasesPerBot(ctx.Context(), tenantID(ctx), len(request.KnowledgeBaseIDs))
	if knowledgeBaseErr != nil {
		return h.entitlementError(ctx, knowledgeBaseErr)
	}

	if request.AccessMode == "public" {
		publicBotErr := h.dependencies.Store.RequirePublicBots(ctx.Context(), tenantID(ctx))
		if publicBotErr != nil {
			return h.entitlementError(ctx, publicBotErr)
		}
	}

	result, err := h.dependencies.Store.UpdateChatbot(ctx.Context(), tenantID(ctx), ctx.Params("id"), request)
	if err != nil {
		return h.storageError(ctx, err)
	}

	if err := h.audit(ctx, "chatbot.access_change", "chatbot", result.ID, map[string]any{"access_mode": result.AccessMode}); err != nil {
		return h.storageError(ctx, err)
	}

	return ctx.JSON(result)
}

func (h *handler) deleteChatbot(ctx *fiber.Ctx) error {
	if err := h.dependencies.Store.DeleteChatbot(ctx.Context(), tenantID(ctx), ctx.Params("id")); err != nil {
		return h.storageError(ctx, err)
	}

	if err := h.audit(ctx, "chatbot.delete", "chatbot", ctx.Params("id"), nil); err != nil {
		return h.storageError(ctx, err)
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}

func (h *handler) publishChatbot(ctx *fiber.Ctx) error {
	result, err := h.dependencies.Store.PublishChatbot(ctx.Context(), tenantID(ctx), ctx.Params("id"))
	if err != nil {
		if strings.Contains(err.Error(), "ready documents") {
			return writeError(ctx, fiber.StatusConflict, "knowledge_base_not_ready", err.Error())
		}

		return h.storageError(ctx, err)
	}

	if err := h.audit(ctx, "chatbot.publish", "chatbot", ctx.Params("id"), map[string]any{"version_id": result.VersionID}); err != nil {
		return h.storageError(ctx, err)
	}

	return ctx.JSON(result)
}

func (h *handler) unpublishChatbot(ctx *fiber.Ctx) error {
	if err := h.dependencies.Store.UnpublishChatbot(ctx.Context(), tenantID(ctx), ctx.Params("id")); err != nil {
		return h.storageError(ctx, err)
	}

	if err := h.audit(ctx, "chatbot.unpublish", "chatbot", ctx.Params("id"), nil); err != nil {
		return h.storageError(ctx, err)
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}

func (h *handler) createSessionToken(ctx *fiber.Ctx) error {
	chatbot, getErr := h.dependencies.Store.GetChatbot(ctx.Context(), tenantID(ctx), ctx.Params("id"))
	if getErr != nil {
		return h.storageError(ctx, getErr)
	}

	if chatbot.AccessMode != "private" || chatbot.PublishedVersionID == nil {
		return writeError(ctx, fiber.StatusConflict, "chatbot_not_private_published", "A published private chatbot is required")
	}

	var request platform.ChatbotSessionTokenRequest
	if len(ctx.Body()) > 0 {
		if bodyErr := ctx.BodyParser(&request); bodyErr != nil {
			return writeError(ctx, fiber.StatusBadRequest, "invalid_request", "Invalid request body")
		}
	}

	ttl := request.TTLSeconds
	if ttl == 0 {
		ttl = 3600
	}

	if ttl < 1 || ttl > 3600 {
		return writeError(ctx, fiber.StatusBadRequest, "invalid_ttl", "ttl_seconds must be between 1 and 3600")
	}

	now := time.Now().UTC()

	token, expiresAt, err := sessiontoken.Sign(h.dependencies.SessionSecret, sessiontoken.Claims{
		ChatbotPublicID: chatbot.PublicID, ChatbotVersionID: *chatbot.PublishedVersionID, TenantID: chatbot.TenantID,
		ExpiresAt: now.Add(time.Duration(ttl) * time.Second).Unix(), Nonce: uuid.NewString(),
	}, now)
	if err != nil {
		return writeError(ctx, fiber.StatusInternalServerError, "token_error", "Unable to issue session token")
	}

	if err := h.audit(ctx, "chatbot.session_token.issue", "chatbot", chatbot.ID, map[string]any{"expires_at": expiresAt}); err != nil {
		return h.storageError(ctx, err)
	}

	return ctx.JSON(platform.ChatbotSessionToken{Token: token, ExpiresAt: expiresAt})
}

func (h *handler) chatbotAnalytics(ctx *fiber.Ctx) error {
	result, err := h.dependencies.Store.GetChatbotAnalytics(ctx.Context(), tenantID(ctx), ctx.Params("id"))
	if err != nil {
		return h.storageError(ctx, err)
	}

	return ctx.JSON(result)
}

func (h *handler) answer(ctx *fiber.Ctx) error {
	publicID := strings.TrimSpace(ctx.Params("publicID"))

	chatbot, getErr := h.dependencies.Store.GetPublishedChatbot(ctx.Context(), publicID)
	if getErr != nil {
		return h.storageError(ctx, getErr)
	}

	authorizationErr := h.authorizeRuntime(ctx, chatbot)
	if authorizationErr != nil {
		return writeError(ctx, fiber.StatusUnauthorized, "private_chatbot_token_required", authorizationErr.Error())
	}

	if !originAllowed(ctx.Get(fiber.HeaderOrigin), chatbot.Chatbot.AllowedOrigins) {
		return writeError(ctx, fiber.StatusForbidden, "origin_not_allowed", "Origin is not allowed for this chatbot")
	}

	var request platform.ChatbotResponseRequest

	bodyErr := ctx.BodyParser(&request)
	if bodyErr != nil || strings.TrimSpace(request.Message) == "" {
		return writeError(ctx, fiber.StatusBadRequest, "invalid_request", "message is required")
	}

	if err := validateChatMessage(request.Message); err != nil {
		return writeError(ctx, fiber.StatusBadRequest, "invalid_request", err.Error())
	}

	if chatbot.Chatbot.AccessMode == "public" {
		allowed, retryAfter, limitErr := h.dependencies.Store.AllowPublicChatbotRequest(
			ctx.Context(), publicID, ctx.Get(headerClientIP), request.SessionID,
		)
		if limitErr != nil {
			return writeError(ctx, fiber.StatusServiceUnavailable, "rate_limit_unavailable", "Public chatbot admission unavailable")
		}

		if !allowed {
			ctx.Set(fiber.HeaderRetryAfter, strconv.Itoa(max(1, int(retryAfter.Seconds()))))

			return writeError(ctx, fiber.StatusTooManyRequests, "rate_limit_exceeded", "Public chatbot rate limit exceeded")
		}
	}

	if request.Stream {
		return h.streamAnswer(ctx, chatbot, request)
	}

	result, err := h.dependencies.Application.Answer(ctx.Context(), *chatbot, request)
	if err != nil {
		return h.answerError(ctx, err)
	}

	return ctx.JSON(result)
}

func (h *handler) streamAnswer(
	ctx *fiber.Ctx,
	chatbot *store.PublishedChatbot,
	request platform.ChatbotResponseRequest,
) error {
	ctx.Set(fiber.HeaderContentType, "text/event-stream")
	ctx.Set(fiber.HeaderCacheControl, "no-cache, no-transform")
	ctx.Set(fiber.HeaderConnection, "keep-alive")
	ctx.Context().SetBodyStreamWriter(func(writer *bufio.Writer) {
		if err := writeSSE(writer, "status", map[string]string{"stage": "retrieving"}); err != nil {
			return
		}

		if err := writeSSE(writer, "status", map[string]string{"stage": "reranking"}); err != nil {
			return
		}

		answerContext, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		result, err := h.dependencies.Application.Answer(answerContext, *chatbot, request)
		if err != nil {
			if writeErr := writeSSE(writer, "error", map[string]string{"code": "rag_dependency_unavailable"}); writeErr != nil {
				return
			}

			return
		}

		if writeErr := writeSSE(writer, "final", result); writeErr != nil {
			return
		}
	})

	return nil
}

func writeSSE(writer *bufio.Writer, event string, value any) error {
	payload, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("encode SSE event: %w", err)
	}

	if _, err := fmt.Fprintf(writer, "event: %s\ndata: %s\n\n", event, payload); err != nil {
		return fmt.Errorf("write SSE event: %w", err)
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("flush SSE event: %w", err)
	}

	return nil
}

func (h *handler) deleteSession(ctx *fiber.Ctx) error {
	publicID := strings.TrimSpace(ctx.Params("publicID"))

	chatbot, err := h.dependencies.Store.GetPublishedChatbot(ctx.Context(), publicID)
	if err != nil {
		return h.storageError(ctx, err)
	}

	if err := h.authorizeRuntime(ctx, chatbot); err != nil {
		return writeError(ctx, fiber.StatusUnauthorized, "private_chatbot_token_required", err.Error())
	}

	if err := h.dependencies.Application.DeleteSession(ctx.Context(), *chatbot, ctx.Params("sessionID")); err != nil {
		return h.storageError(ctx, err)
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}

func (h *handler) authorizeRuntime(ctx *fiber.Ctx, chatbot *store.PublishedChatbot) error {
	if chatbot.Chatbot.AccessMode == "public" {
		return nil
	}

	scheme, token, found := strings.Cut(strings.TrimSpace(ctx.Get(fiber.HeaderAuthorization)), " ")
	if !found || !strings.EqualFold(scheme, "Bearer") || strings.TrimSpace(token) == "" {
		return errors.New("private chatbot session token required")
	}

	_, err := sessiontoken.Verify(
		h.dependencies.SessionSecret,
		strings.TrimSpace(token),
		chatbot.Chatbot.PublicID,
		chatbot.Version.VersionID,
		chatbot.Chatbot.TenantID,
		time.Now().UTC(),
	)

	return err
}

func originAllowed(origin string, allowlist []string) bool {
	origin = strings.TrimRight(strings.TrimSpace(origin), "/")
	if origin == "" {
		return true
	}

	for _, allowed := range allowlist {
		if strings.EqualFold(origin, strings.TrimRight(strings.TrimSpace(allowed), "/")) {
			return true
		}
	}

	return false
}

func parseChatbotRequest(ctx *fiber.Ctx) (platform.UpsertChatbotRequest, error) {
	var request platform.UpsertChatbotRequest
	if err := ctx.BodyParser(&request); err != nil {
		return request, errors.New("invalid request body")
	}

	if err := validateChatbotRequest(request); err != nil {
		return request, err
	}

	return request, nil
}
