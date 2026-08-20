// Package ragproxy forwards authenticated tenant context to the isolated RAG data plane.
package ragproxy

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
)

const maxRAGResponseBytes = 32 << 20

// Handler is a content-blind reverse proxy for managed RAG APIs.
type Handler struct {
	baseURL string
	token   string
	client  *http.Client
}

// New creates an optional RAG proxy. Empty configuration disables the feature surface.
func New(baseURL, token string, timeout time.Duration) (*Handler, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")

	token = strings.TrimSpace(token)
	if baseURL == "" && token == "" {
		return nil, nil
	}

	if baseURL == "" || len(token) < 32 {
		return nil, errors.New("RAG service URL and a 32-byte internal token must be configured together")
	}

	if timeout <= 0 {
		timeout = 2 * time.Minute
	}

	return &Handler{baseURL: baseURL, token: token, client: &http.Client{Timeout: timeout}}, nil
}

// ForwardManagement forwards tenant management requests after control-plane RBAC.
func (h *Handler) ForwardManagement(ctx *fiber.Ctx) error {
	return h.forward(ctx, ctx.OriginalURL(), true)
}

// ForwardRetrieval forwards the protected retrieval contract.
func (h *Handler) ForwardRetrieval(ctx *fiber.Ctx) error {
	authInfo := auth.GetAuthInfo(ctx)
	if !auth.IsTenantAdminOrHigher(authInfo) && !hasPermission(authInfo, "knowledge_base:query") {
		return fiber.NewError(fiber.StatusForbidden, "Knowledge base query permission required")
	}

	return h.forward(ctx, ctx.OriginalURL(), true)
}

// ForwardRuntime forwards public/private chatbot runtime admission to the RAG service.
func (h *Handler) ForwardRuntime(ctx *fiber.Ctx) error {
	return h.forward(ctx, ctx.OriginalURL(), false)
}

func (h *Handler) forward(ctx *fiber.Ctx, path string, requireTenant bool) error {
	request, err := http.NewRequestWithContext(ctx.Context(), ctx.Method(), h.baseURL+path, bytes.NewReader(ctx.Body()))
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, "Unable to create RAG request")
	}

	request.Header.Set("x-godarai-rag-token", h.token)
	request.Header.Set("x-godarai-client-ip", strings.TrimSpace(ctx.IP()))
	copyRequestHeader(ctx, request, fiber.HeaderContentType)
	copyRequestHeader(ctx, request, fiber.HeaderAccept)
	copyRequestHeader(ctx, request, fiber.HeaderAuthorization)
	copyRequestHeader(ctx, request, fiber.HeaderOrigin)
	copyRequestHeader(ctx, request, "x-client-request-id")

	if requireTenant {
		currentTenant := tenant.FromContext(ctx)
		if currentTenant == nil {
			return fiber.NewError(fiber.StatusBadRequest, "Tenant context required")
		}

		request.Header.Set("x-godarai-tenant-id", currentTenant.ID)
		request.Header.Set("x-godarai-tenant", currentTenant.Slug)

		if authInfo := auth.GetAuthInfo(ctx); authInfo != nil {
			request.Header.Set("x-godarai-user-id", strings.TrimSpace(authInfo.UserID))
		}
	}

	response, err := h.client.Do(request) //nolint:bodyclose // SSE response ownership transfers to the Fiber stream writer.
	if err != nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "Managed RAG service unavailable")
	}

	contentType := strings.TrimSpace(response.Header.Get(fiber.HeaderContentType))
	if strings.HasPrefix(strings.ToLower(contentType), "text/event-stream") {
		ctx.Set(fiber.HeaderContentType, contentType)
		ctx.Set(fiber.HeaderCacheControl, "no-cache, no-transform")
		ctx.Context().SetBodyStreamWriter(func(writer *bufio.Writer) {
			defer closeResponseBody(response.Body)

			if _, copyErr := io.Copy(writer, response.Body); copyErr != nil {
				slog.Warn("managed RAG SSE proxy copy failed", "error_type", "stream_copy")

				return
			}

			if flushErr := writer.Flush(); flushErr != nil {
				slog.Warn("managed RAG SSE proxy flush failed", "error_type", "stream_flush")
			}
		})

		return nil
	}
	defer closeResponseBody(response.Body)

	body, err := io.ReadAll(io.LimitReader(response.Body, maxRAGResponseBytes+1))
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, "Unable to read RAG response")
	}

	if len(body) > maxRAGResponseBytes {
		return fiber.NewError(fiber.StatusBadGateway, "RAG response exceeds size limit")
	}

	if contentType != "" {
		ctx.Set(fiber.HeaderContentType, contentType)
	}

	if retryAfter := strings.TrimSpace(response.Header.Get(fiber.HeaderRetryAfter)); retryAfter != "" {
		ctx.Set(fiber.HeaderRetryAfter, retryAfter)
	}

	return ctx.Status(response.StatusCode).Send(body)
}

func copyRequestHeader(ctx *fiber.Ctx, request *http.Request, key string) {
	if value := strings.TrimSpace(ctx.Get(key)); value != "" {
		request.Header.Set(key, value)
	}
}

func hasPermission(authInfo *platform.AuthContext, permission string) bool {
	if authInfo == nil {
		return false
	}

	for _, item := range authInfo.Permissions {
		if strings.EqualFold(strings.TrimSpace(item), permission) {
			return true
		}
	}

	return false
}

// Health verifies the RAG service without logging response content.
func (h *Handler) Health(ctx context.Context) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, h.baseURL+"/ready", http.NoBody)
	if err != nil {
		return err
	}

	response, err := h.client.Do(request)
	if err != nil {
		return err
	}
	defer closeResponseBody(response.Body)

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("RAG readiness returned HTTP %d", response.StatusCode)
	}

	return nil
}

func closeResponseBody(body io.Closer) {
	if err := body.Close(); err != nil {
		slog.Warn("managed RAG proxy response close failed", "error_type", "response_close")
	}
}
