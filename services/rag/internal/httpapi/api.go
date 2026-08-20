// Package httpapi exposes the managed RAG management and runtime contracts.
package httpapi

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/rag/internal/app"
	"github.com/Delpak/shahrah/services/rag/internal/domain"
	"github.com/Delpak/shahrah/services/rag/internal/files"
	"github.com/Delpak/shahrah/services/rag/internal/store"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	headerInternalToken = "x-godarai-rag-token"
	headerTenantID      = "x-godarai-tenant-id"
	headerTenantSlug    = "x-godarai-tenant"
	headerActorID       = "x-godarai-user-id"
	headerClientIP      = "x-godarai-client-ip"
)

// Dependencies contains the complete RAG API collaborators.
type Dependencies struct {
	Store         *store.Store
	Application   *app.Service
	Files         *files.Service
	Logger        *zap.Logger
	InternalToken string
	SessionSecret []byte
	Profile       domain.RAGProfile
}

// New creates the API-mode HTTP application.
func New(dependencies Dependencies) (*fiber.App, error) {
	if dependencies.Store == nil || dependencies.Application == nil || dependencies.Files == nil || dependencies.Logger == nil {
		return nil, errors.New("RAG HTTP dependencies are required")
	}

	if len(strings.TrimSpace(dependencies.InternalToken)) < 32 || len(dependencies.SessionSecret) < 32 {
		return nil, errors.New("RAG internal token and session secret must be at least 32 bytes")
	}

	if err := dependencies.Profile.Validate(); err != nil {
		return nil, fmt.Errorf("invalid active RAG profile: %w", err)
	}

	handler := &handler{dependencies: dependencies}
	server := fiber.New(fiber.Config{
		BodyLimit:    int(files.MaxFileBytes + 1024*1024),
		ReadTimeout:  2 * time.Minute,
		WriteTimeout: 2 * time.Minute,
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			var fiberError *fiber.Error
			if errors.As(err, &fiberError) {
				return writeError(ctx, fiberError.Code, "request_error", fiberError.Message)
			}

			return writeError(ctx, fiber.StatusInternalServerError, "server_error", "Internal Server Error")
		},
	})

	server.Get("/health", func(ctx *fiber.Ctx) error {
		return ctx.JSON(fiber.Map{"status": "healthy", "profile": dependencies.Profile.Revision})
	})
	server.Get("/ready", handler.ready)

	management := server.Group("/api/v1", handler.requireInternal)
	handler.registerManagement(management)

	runtime := server.Group("/v1")
	runtime.Post("/retrieval", handler.requireInternal, handler.retrieve)
	runtime.Post("/chatbots/:publicID/responses", handler.requireProxy, handler.answer)
	runtime.Delete("/chatbots/:publicID/sessions/:sessionID", handler.requireProxy, handler.deleteSession)
	runtime.Get("/chatbots/:publicID/openai/models", handler.requireProxy, handler.openAIModels)
	runtime.Post("/chatbots/:publicID/openai/chat/completions", handler.requireProxy, handler.openAIChatCompletions)

	return server, nil
}

type handler struct {
	dependencies Dependencies
}

func (h *handler) registerManagement(router fiber.Router) {
	router.Post("/knowledge-bases", h.createKnowledgeBase)
	router.Get("/knowledge-bases", h.listKnowledgeBases)
	router.Get("/knowledge-bases/:id", h.getKnowledgeBase)
	router.Patch("/knowledge-bases/:id", h.updateKnowledgeBase)
	router.Delete("/knowledge-bases/:id", h.deleteKnowledgeBase)
	router.Post("/knowledge-bases/:id/documents", h.createDocument)
	router.Get("/knowledge-bases/:id/documents", h.listDocuments)
	router.Get("/knowledge-bases/:id/documents/:documentID", h.getDocument)
	router.Delete("/knowledge-bases/:id/documents/:documentID", h.deleteDocument)
	router.Post("/knowledge-bases/:id/documents/:documentID/retry", h.retryDocument)

	router.Post("/chatbots", h.createChatbot)
	router.Get("/chatbots", h.listChatbots)
	router.Get("/chatbots/:id", h.getChatbot)
	router.Put("/chatbots/:id", h.updateChatbot)
	router.Patch("/chatbots/:id", h.updateChatbot)
	router.Delete("/chatbots/:id", h.deleteChatbot)
	router.Post("/chatbots/:id/publish", h.publishChatbot)
	router.Post("/chatbots/:id/unpublish", h.unpublishChatbot)
	router.Post("/chatbots/:id/session-tokens", h.createSessionToken)
	router.Post("/chatbots/:id/preview-responses", h.previewAnswer)
	router.Get("/chatbots/:id/analytics", h.chatbotAnalytics)
}

func (h *handler) requireInternal(ctx *fiber.Ctx) error {
	if !h.proxyAuthenticated(ctx) {
		return writeError(ctx, fiber.StatusUnauthorized, "authentication_error", "Internal authentication required")
	}

	if strings.TrimSpace(ctx.Get(headerTenantID)) == "" || strings.TrimSpace(ctx.Get(headerTenantSlug)) == "" {
		return writeError(ctx, fiber.StatusBadRequest, "tenant_required", "Tenant context required")
	}

	if err := h.dependencies.Store.RequireRAGEnabled(ctx.Context(), tenantID(ctx)); err != nil {
		return writeError(ctx, fiber.StatusForbidden, "rag_addon_required", "Managed RAG add-on required")
	}

	return ctx.Next()
}

func (h *handler) requireProxy(ctx *fiber.Ctx) error {
	if !h.proxyAuthenticated(ctx) {
		return writeError(ctx, fiber.StatusUnauthorized, "authentication_error", "Internal authentication required")
	}

	return ctx.Next()
}

func (h *handler) proxyAuthenticated(ctx *fiber.Ctx) bool {
	expected := []byte(strings.TrimSpace(h.dependencies.InternalToken))

	actual := []byte(strings.TrimSpace(ctx.Get(headerInternalToken)))

	return len(actual) == len(expected) && subtle.ConstantTimeCompare(actual, expected) == 1
}

func (h *handler) ready(ctx *fiber.Ctx) error {
	if err := h.dependencies.Store.Ready(ctx.Context()); err != nil {
		return writeError(ctx, fiber.StatusServiceUnavailable, "dependency_unavailable", "RAG data stores unavailable")
	}

	profile, err := h.dependencies.Store.LoadActiveRAGProfile(ctx.Context())
	if err != nil {
		return writeError(ctx, fiber.StatusServiceUnavailable, "profile_unavailable", "Active RAG profile unavailable")
	}

	return ctx.JSON(fiber.Map{"status": "ready", "profile": profile.Revision})
}

func (h *handler) createKnowledgeBase(ctx *fiber.Ctx) error {
	var request platform.CreateKnowledgeBaseRequest
	if err := ctx.BodyParser(&request); err != nil {
		return writeError(ctx, fiber.StatusBadRequest, "invalid_request", "Invalid request body")
	}

	request.Name = strings.TrimSpace(request.Name)

	request.Description = strings.TrimSpace(request.Description)
	if err := validateTextLength("name", request.Name, 1, 160); err != nil {
		return writeError(ctx, fiber.StatusBadRequest, "invalid_request", err.Error())
	}

	if err := validateTextLength("description", request.Description, 0, 2000); err != nil {
		return writeError(ctx, fiber.StatusBadRequest, "invalid_request", err.Error())
	}

	if err := h.dependencies.Store.RequireKnowledgeBaseCapacity(ctx.Context(), tenantID(ctx)); err != nil {
		return h.entitlementError(ctx, err)
	}

	result, err := h.dependencies.Store.CreateKnowledgeBase(ctx.Context(), tenantID(ctx), actorID(ctx), request)
	if err != nil {
		return h.storageError(ctx, err)
	}

	if err := h.audit(ctx, "knowledge_base.create", "knowledge_base", result.ID, nil); err != nil {
		return h.storageError(ctx, err)
	}

	return ctx.Status(fiber.StatusCreated).JSON(result)
}

func (h *handler) listKnowledgeBases(ctx *fiber.Ctx) error {
	limit, offset := pagination(ctx)

	result, err := h.dependencies.Store.ListKnowledgeBases(ctx.Context(), tenantID(ctx), limit, offset)
	if err != nil {
		return h.storageError(ctx, err)
	}

	return ctx.JSON(result)
}

func (h *handler) getKnowledgeBase(ctx *fiber.Ctx) error {
	result, err := h.dependencies.Store.GetKnowledgeBase(ctx.Context(), tenantID(ctx), ctx.Params("id"))
	if err != nil {
		return h.storageError(ctx, err)
	}

	return ctx.JSON(result)
}

func (h *handler) updateKnowledgeBase(ctx *fiber.Ctx) error {
	var request platform.UpdateKnowledgeBaseRequest
	if err := ctx.BodyParser(&request); err != nil {
		return writeError(ctx, fiber.StatusBadRequest, "invalid_request", "Invalid request body")
	}

	if request.Name == nil && request.Description == nil {
		return writeError(ctx, fiber.StatusBadRequest, "invalid_request", "name or description is required")
	}

	if request.Name != nil {
		name := strings.TrimSpace(*request.Name)
		if err := validateTextLength("name", name, 1, 160); err != nil {
			return writeError(ctx, fiber.StatusBadRequest, "invalid_request", err.Error())
		}

		request.Name = &name
	}

	if request.Description != nil {
		description := strings.TrimSpace(*request.Description)
		if err := validateTextLength("description", description, 0, 2000); err != nil {
			return writeError(ctx, fiber.StatusBadRequest, "invalid_request", err.Error())
		}

		request.Description = &description
	}

	result, err := h.dependencies.Store.UpdateKnowledgeBase(ctx.Context(), tenantID(ctx), ctx.Params("id"), request)
	if err != nil {
		return h.storageError(ctx, err)
	}

	if err := h.audit(ctx, "knowledge_base.update", "knowledge_base", result.ID, nil); err != nil {
		return h.storageError(ctx, err)
	}

	return ctx.JSON(result)
}

func (h *handler) deleteKnowledgeBase(ctx *fiber.Ctx) error {
	if err := h.dependencies.Store.DeleteKnowledgeBase(ctx.Context(), tenantID(ctx), ctx.Params("id")); err != nil {
		return h.storageError(ctx, err)
	}

	if err := h.audit(ctx, "knowledge_base.delete", "knowledge_base", ctx.Params("id"), nil); err != nil {
		return h.storageError(ctx, err)
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}

func (h *handler) createDocument(ctx *fiber.Ctx) error {
	contentType := strings.ToLower(ctx.Get(fiber.HeaderContentType))
	if strings.HasPrefix(contentType, fiber.MIMEMultipartForm) {
		return h.createFileDocument(ctx)
	}

	var request platform.CreateTextDocumentRequest
	if err := ctx.BodyParser(&request); err != nil || strings.TrimSpace(request.Title) == "" || strings.TrimSpace(request.Text) == "" {
		return writeError(ctx, fiber.StatusBadRequest, "invalid_request", "title and text are required")
	}

	request.Title = strings.TrimSpace(request.Title)
	if err := validateTextLength("title", request.Title, 1, 255); err != nil {
		return writeError(ctx, fiber.StatusBadRequest, "invalid_request", err.Error())
	}

	if err := h.dependencies.Store.RequireDocumentCapacity(ctx.Context(), tenantID(ctx), int64(len([]byte(request.Text)))); err != nil {
		return h.entitlementError(ctx, err)
	}

	result, err := h.dependencies.Store.CreateTextDocument(
		ctx.Context(), tenantID(ctx), ctx.Params("id"), actorID(ctx), request,
	)
	if err != nil {
		return h.storageError(ctx, err)
	}

	if err := h.audit(ctx, "document.upload", "document", result.ID, map[string]any{"source_type": "text"}); err != nil {
		return h.storageError(ctx, err)
	}

	return ctx.Status(fiber.StatusAccepted).JSON(result)
}

func (h *handler) createFileDocument(ctx *fiber.Ctx) error {
	header, formErr := ctx.FormFile("file")
	if formErr != nil || header == nil || header.Size <= 0 || header.Size > files.MaxFileBytes {
		return writeError(ctx, fiber.StatusBadRequest, "invalid_file", "A supported file is required")
	}

	capacityErr := h.dependencies.Store.RequireDocumentCapacity(ctx.Context(), tenantID(ctx), header.Size)
	if capacityErr != nil {
		return h.entitlementError(ctx, capacityErr)
	}

	file, err := header.Open()
	if err != nil {
		return writeError(ctx, fiber.StatusBadRequest, "invalid_file", "Unable to open file")
	}

	content, readErr := io.ReadAll(io.LimitReader(file, files.MaxFileBytes+1))

	closeErr := file.Close()
	if readErr != nil || closeErr != nil || int64(len(content)) > files.MaxFileBytes {
		return writeError(ctx, fiber.StatusBadRequest, "invalid_file", "Unable to read file")
	}

	upload, err := h.dependencies.Files.ValidateAndPut(ctx.Context(), tenantID(ctx), header.Filename, content)
	if err != nil {
		if errors.Is(err, files.ErrFileRejected) || errors.Is(err, files.ErrInfected) {
			return writeError(ctx, fiber.StatusUnprocessableEntity, "file_rejected", "File failed security or format validation")
		}

		return writeError(ctx, fiber.StatusServiceUnavailable, "scan_unavailable", "File scanning unavailable")
	}

	result, err := h.dependencies.Store.CreateFileDocument(
		ctx.Context(), tenantID(ctx), ctx.Params("id"), actorID(ctx), upload.Title,
		upload.MediaType, upload.ObjectKey, upload.SHA256, upload.SizeBytes,
	)
	if err != nil {
		if deleteErr := h.dependencies.Files.Delete(ctx.Context(), upload.ObjectKey); deleteErr != nil {
			h.dependencies.Logger.Warn("failed to clean up unreferenced RAG upload", zap.Error(deleteErr))
		}

		return h.storageError(ctx, err)
	}

	if err := h.audit(ctx, "document.upload", "document", result.ID, map[string]any{"source_type": "file", "media_type": upload.MediaType}); err != nil {
		return h.storageError(ctx, err)
	}

	return ctx.Status(fiber.StatusAccepted).JSON(result)
}

func (h *handler) listDocuments(ctx *fiber.Ctx) error {
	limit, offset := pagination(ctx)

	result, err := h.dependencies.Store.ListDocuments(ctx.Context(), tenantID(ctx), ctx.Params("id"), limit, offset)
	if err != nil {
		return h.storageError(ctx, err)
	}

	return ctx.JSON(result)
}

func (h *handler) getDocument(ctx *fiber.Ctx) error {
	result, err := h.dependencies.Store.GetDocument(ctx.Context(), tenantID(ctx), ctx.Params("id"), ctx.Params("documentID"))
	if err != nil {
		return h.storageError(ctx, err)
	}

	return ctx.JSON(result)
}

func (h *handler) deleteDocument(ctx *fiber.Ctx) error {
	err := h.dependencies.Store.DeleteDocument(ctx.Context(), tenantID(ctx), ctx.Params("id"), ctx.Params("documentID"))
	if err != nil {
		return h.storageError(ctx, err)
	}

	if err := h.audit(ctx, "document.delete", "document", ctx.Params("documentID"), nil); err != nil {
		return h.storageError(ctx, err)
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}

func (h *handler) retryDocument(ctx *fiber.Ctx) error {
	err := h.dependencies.Store.RetryDocument(ctx.Context(), tenantID(ctx), ctx.Params("id"), ctx.Params("documentID"))
	if err != nil {
		return h.storageError(ctx, err)
	}

	return ctx.SendStatus(fiber.StatusAccepted)
}

func (h *handler) retrieve(ctx *fiber.Ctx) error {
	var request platform.RetrievalRequest
	if err := ctx.BodyParser(&request); err != nil {
		return writeError(ctx, fiber.StatusBadRequest, "invalid_request", "Invalid request body")
	}

	result, err := h.dependencies.Application.Retrieve(ctx.Context(), tenantID(ctx), tenantSlug(ctx), request)
	if err != nil {
		if errors.Is(err, store.ErrEntitlementDenied) {
			return writeError(ctx, fiber.StatusTooManyRequests, "rag_query_quota_exceeded", "Managed RAG query quota exceeded")
		}

		if errors.Is(err, store.ErrNotFound) {
			return writeError(ctx, fiber.StatusNotFound, "not_found", "Knowledge base not found")
		}

		if strings.Contains(err.Error(), "required") || strings.Contains(err.Error(), "identifier") {
			return writeError(ctx, fiber.StatusBadRequest, "invalid_request", err.Error())
		}

		return writeError(ctx, fiber.StatusServiceUnavailable, "rag_dependency_unavailable", "Retrieval or rerank unavailable")
	}

	return ctx.JSON(result)
}

func (h *handler) entitlementError(ctx *fiber.Ctx, err error) error {
	if errors.Is(err, store.ErrEntitlementDenied) {
		return writeError(ctx, fiber.StatusForbidden, "rag_entitlement_required", "Managed RAG entitlement or capacity required")
	}

	return h.storageError(ctx, err)
}

func (h *handler) audit(ctx *fiber.Ctx, action, resourceType, resourceID string, metadata map[string]any) error {
	return h.dependencies.Store.RecordAudit(
		context.WithoutCancel(ctx.Context()), tenantID(ctx), actorID(ctx), action, resourceType, resourceID, metadata,
	)
}

func (h *handler) storageError(ctx *fiber.Ctx, err error) error {
	if errors.Is(err, store.ErrNotFound) {
		return writeError(ctx, fiber.StatusNotFound, "not_found", "Resource not found")
	}

	h.dependencies.Logger.Error("RAG storage operation failed", zap.Error(err))

	return writeError(ctx, fiber.StatusInternalServerError, "storage_error", "Storage operation failed")
}

func tenantID(ctx *fiber.Ctx) string   { return strings.TrimSpace(ctx.Get(headerTenantID)) }
func tenantSlug(ctx *fiber.Ctx) string { return strings.TrimSpace(ctx.Get(headerTenantSlug)) }
func actorID(ctx *fiber.Ctx) string    { return strings.TrimSpace(ctx.Get(headerActorID)) }

func pagination(ctx *fiber.Ctx) (int, int) {
	limit, err := strconv.Atoi(ctx.Query("limit", "50"))
	if err != nil || limit < 1 {
		limit = 50
	}

	limit = min(limit, 100)

	offset, err := strconv.Atoi(ctx.Query("offset", "0"))
	if err != nil || offset < 0 {
		offset = 0
	}

	return limit, offset
}

func writeError(ctx *fiber.Ctx, status int, code, message string) error {
	return ctx.Status(status).JSON(fiber.Map{"error": fiber.Map{
		"type": "rag_error", "code": code, "message": message,
	}})
}

func validateChatbotRequest(request platform.UpsertChatbotRequest) error {
	if err := validateTextLength("name", request.Name, 1, 160); err != nil {
		return err
	}

	if err := validateTextLength("description", request.Description, 0, 2000); err != nil {
		return err
	}

	if err := validateTextLength("system_instruction", request.SystemInstruction, 0, 8000); err != nil {
		return err
	}

	if err := validateTextLength("answer_model", request.AnswerModel, 1, 255); err != nil {
		return err
	}

	if len(request.KnowledgeBaseIDs) < 1 || len(request.KnowledgeBaseIDs) > 5 {
		return errors.New("knowledge_base_ids must contain between 1 and 5 items")
	}

	seenKnowledgeBases := make(map[string]struct{}, len(request.KnowledgeBaseIDs))
	for _, id := range request.KnowledgeBaseIDs {
		if _, err := uuid.Parse(id); err != nil {
			return fmt.Errorf("invalid knowledge base identifier")
		}

		if _, exists := seenKnowledgeBases[id]; exists {
			return errors.New("knowledge_base_ids must not contain duplicates")
		}

		seenKnowledgeBases[id] = struct{}{}
	}

	if request.RetentionDays != 0 && request.RetentionDays != 7 && request.RetentionDays != 30 {
		return errors.New("retention_days must be 0, 7, or 30")
	}

	accessMode := strings.ToLower(strings.TrimSpace(request.AccessMode))
	if accessMode != "" && accessMode != "private" && accessMode != "public" {
		return errors.New("access_mode must be private or public")
	}

	locale := strings.TrimSpace(request.DefaultLocale)
	if locale != "" && locale != "fa-IR" && locale != "en-US" {
		return errors.New("default_locale must be fa-IR or en-US")
	}

	answerLanguage := strings.ToLower(strings.TrimSpace(request.AnswerLanguage))
	if answerLanguage != "" && answerLanguage != "query" && answerLanguage != "fa" && answerLanguage != "en" {
		return errors.New("answer_language must be query, fa, or en")
	}

	if len(request.AllowedOrigins) > 50 {
		return errors.New("allowed_origins must not contain more than 50 items")
	}

	seenOrigins := make(map[string]struct{}, len(request.AllowedOrigins))
	for _, origin := range request.AllowedOrigins {
		if err := validateTextLength("allowed origin", origin, 1, 2048); err != nil {
			return err
		}

		if err := validateAllowedOrigin(origin); err != nil {
			return err
		}

		normalizedOrigin := strings.ToLower(strings.TrimRight(strings.TrimSpace(origin), "/"))
		if _, exists := seenOrigins[normalizedOrigin]; exists {
			return errors.New("allowed_origins must not contain duplicates")
		}

		seenOrigins[normalizedOrigin] = struct{}{}
	}

	branding, err := json.Marshal(request.Branding)
	if err != nil {
		return errors.New("branding must be valid JSON")
	}

	if len(branding) > 16*1024 {
		return errors.New("branding must not exceed 16 KiB")
	}

	return nil
}

func validateAllowedOrigin(value string) error {
	parsed, err := url.ParseRequestURI(strings.TrimSpace(value))
	if err != nil || parsed.Host == "" || parsed.User != nil ||
		parsed.RawQuery != "" || parsed.Fragment != "" ||
		(parsed.Scheme != "http" && parsed.Scheme != "https") ||
		(parsed.Path != "" && parsed.Path != "/") {
		return fmt.Errorf("allowed_origins must contain only HTTP or HTTPS origins")
	}

	return nil
}
