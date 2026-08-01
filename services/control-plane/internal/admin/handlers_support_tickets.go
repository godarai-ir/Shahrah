package admin

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/apiheaders"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	prommetrics "github.com/Delpak/shahrah/services/control-plane/internal/observability/prometheus"
	"github.com/Delpak/shahrah/services/control-plane/internal/storage"
	"github.com/Delpak/shahrah/services/control-plane/internal/supportfiles"
	"github.com/gofiber/fiber/v2"
)

const (
	defaultSupportTicketPageSize = 25
	maxSupportTicketPageSize     = 100
)

func (h *Handlers) ListTenantSupportTickets(c *fiber.Ctx) error {
	currentTenant, err := currentTenantOrErr(c)
	if err != nil {
		return err
	}

	actorID := delegatedUserID(c)

	readAll, allowed := tenantTicketReadAccess(c)
	if !allowed || actorID == "" {
		return writeSupportTicketError(c, fiber.StatusForbidden, "support_ticket_forbidden", "support ticket access denied")
	}

	params, err := parseSupportTicketListParams(c)
	if err != nil {
		return err
	}

	params.TenantID = currentTenant.ID
	params.RequesterUserID = actorID
	params.ReadAll = readAll

	items, total, err := h.requireSupportTicketStoreAndList(c.Context(), params)
	if err != nil {
		return mapSupportTicketError(c, err)
	}

	return c.JSON(platform.SupportTicketListResponse{
		Data:       items,
		NextCursor: nextOffsetCursor(params.Offset, len(items), total),
	})
}

//nolint:govet // Short-lived branch errors are checked immediately and intentionally scoped with short declarations.
func (h *Handlers) CreateTenantSupportTicket(c *fiber.Ctx) error {
	if !tenantTicketHasPermission(c, "support_ticket:create") {
		return writeSupportTicketError(c, fiber.StatusForbidden, "support_ticket_forbidden", "support ticket create permission required")
	}

	currentTenant, err := currentTenantOrErr(c)
	if err != nil {
		return err
	}

	actorID := delegatedUserID(c)
	if actorID == "" {
		return writeSupportTicketError(c, fiber.StatusForbidden, "support_ticket_forbidden", "authenticated user context missing")
	}

	req, files, err := parseCreateSupportTicketRequest(c)
	if err != nil {
		return mapSupportTicketError(c, err)
	}

	if err := validateCreateSupportTicketRequest(req, len(files)); err != nil {
		return mapSupportTicketError(c, err)
	}

	req.IdempotencyKey = supportIdempotencyKey(c, req.IdempotencyKey)
	if req.IdempotencyKey == "" {
		return writeSupportTicketError(c, fiber.StatusBadRequest, "invalid_request", "idempotency key is required")
	}

	uploads, err := h.uploadSupportFiles(c.Context(), currentTenant.ID, "new-"+shortHash(req.IdempotencyKey), files)
	if err != nil {
		return mapSupportTicketError(c, err)
	}

	req.RequestHash = hashSupportRequest(req, uploads)

	store, err := h.requireSupportTicketStore()
	if err != nil {
		h.deleteSupportUploads(uploads)
		return err
	}

	detail, replay, err := store.CreateSupportTicket(
		c.Context(),
		currentTenant.ID,
		actorID,
		req,
		uploads,
		tenantSupportTicketActor(c),
		intEnv("SUPPORT_TICKETS_PER_USER_PER_HOUR", 10),
	)
	if err != nil {
		h.deleteSupportUploads(uploads)
		return mapSupportTicketError(c, err)
	}

	if replay {
		h.deleteSupportUploads(uploads)
		c.Set("Idempotency-Replayed", "true")
	} else {
		prommetrics.RecordSupportTicketCreated()
		prommetrics.RecordSupportMessage(platform.SupportTicketAuthorTenantUser)
	}

	return c.Status(fiber.StatusCreated).JSON(detail)
}

func (h *Handlers) GetTenantSupportTicket(c *fiber.Ctx) error {
	currentTenant, err := currentTenantOrErr(c)
	if err != nil {
		return err
	}

	actorID := delegatedUserID(c)

	readAll, allowed := tenantTicketReadAccess(c)
	if !allowed || actorID == "" {
		return writeSupportTicketError(c, fiber.StatusForbidden, "support_ticket_forbidden", "support ticket access denied")
	}

	store, err := h.requireSupportTicketStore()
	if err != nil {
		return err
	}

	detail, err := store.GetSupportTicketDetail(c.Context(), currentTenant.ID, c.Params("ticketID"), actorID, readAll)
	if err != nil {
		return mapSupportTicketError(c, err)
	}

	return c.JSON(detail)
}

func (h *Handlers) ListTenantSupportTicketMessages(c *fiber.Ctx) error {
	currentTenant, err := currentTenantOrErr(c)
	if err != nil {
		return err
	}

	actorID := delegatedUserID(c)

	readAll, allowed := tenantTicketReadAccess(c)
	if !allowed || actorID == "" {
		return writeSupportTicketError(c, fiber.StatusForbidden, "support_ticket_forbidden", "support ticket access denied")
	}

	store, err := h.requireSupportTicketStore()
	if err != nil {
		return err
	}

	ticket, err := store.GetSupportTicket(c.Context(), currentTenant.ID, c.Params("ticketID"), actorID, readAll)
	if err != nil {
		return mapSupportTicketError(c, err)
	}

	limit, offset, err := parseCursorPagination(c)
	if err != nil {
		return err
	}

	items, err := store.ListSupportTicketMessages(c.Context(), ticket.ID, limit+1, offset)
	if err != nil {
		return mapSupportTicketError(c, err)
	}

	next := ""

	if len(items) > limit {
		items = items[:limit]
		next = encodeOffsetCursor(offset + limit)
	}

	return c.JSON(fiber.Map{"data": items, "next_cursor": next})
}

func (h *Handlers) AddTenantSupportTicketMessage(c *fiber.Ctx) error {
	if !tenantTicketHasPermission(c, "support_ticket:reply") {
		return writeSupportTicketError(c, fiber.StatusForbidden, "support_ticket_forbidden", "support ticket reply permission required")
	}

	currentTenant, err := currentTenantOrErr(c)
	if err != nil {
		return err
	}

	actorID := delegatedUserID(c)

	readAll, allowed := tenantTicketReadAccess(c)
	if !allowed || actorID == "" {
		return writeSupportTicketError(c, fiber.StatusForbidden, "support_ticket_forbidden", "support ticket access denied")
	}

	req, files, err := parseAddSupportTicketMessageRequest(c)
	if err != nil {
		return mapSupportTicketError(c, err)
	}

	if strings.TrimSpace(req.Body) == "" && len(files) == 0 {
		return writeSupportTicketError(c, fiber.StatusBadRequest, "invalid_request", "message body or attachment is required")
	}

	if len([]rune(req.Body)) > 10000 {
		return writeSupportTicketError(c, fiber.StatusBadRequest, "invalid_request", "message body must not exceed 10000 characters")
	}

	req.IdempotencyKey = supportIdempotencyKey(c, req.IdempotencyKey)
	if req.IdempotencyKey == "" {
		return writeSupportTicketError(c, fiber.StatusBadRequest, "invalid_request", "idempotency key is required")
	}

	store, err := h.requireSupportTicketStore()
	if err != nil {
		return err
	}

	ticket, err := store.GetSupportTicket(c.Context(), currentTenant.ID, c.Params("ticketID"), actorID, readAll)
	if err != nil {
		return mapSupportTicketError(c, err)
	}

	uploads, err := h.uploadSupportFiles(c.Context(), currentTenant.ID, ticket.ID, files)
	if err != nil {
		return mapSupportTicketError(c, err)
	}

	req.NextStatus = ""
	req.RequestHash = hashSupportRequest(req, uploads)

	detail, replay, err := store.AddSupportTicketMessage(
		c.Context(),
		currentTenant.ID,
		ticket.ID,
		req,
		uploads,
		tenantSupportTicketActor(c),
		readAll,
		intEnv("SUPPORT_MESSAGES_PER_USER_PER_TICKET_PER_HOUR", 30),
	)
	if err != nil {
		h.deleteSupportUploads(uploads)
		return mapSupportTicketError(c, err)
	}

	if replay {
		h.deleteSupportUploads(uploads)
		c.Set("Idempotency-Replayed", "true")
	} else {
		prommetrics.RecordSupportMessage(platform.SupportTicketAuthorTenantUser)
	}

	return c.Status(fiber.StatusCreated).JSON(detail)
}

//nolint:govet // Short-lived branch errors are checked immediately and intentionally scoped with short declarations.
func (h *Handlers) RateTenantSupportTicket(c *fiber.Ctx) error {
	if !tenantTicketHasPermission(c, "support_ticket:rate") {
		return writeSupportTicketError(c, fiber.StatusForbidden, "support_ticket_forbidden", "support ticket rating permission required")
	}

	currentTenant, err := currentTenantOrErr(c)
	if err != nil {
		return err
	}

	actorID := delegatedUserID(c)
	if actorID == "" {
		return writeSupportTicketError(c, fiber.StatusForbidden, "support_ticket_forbidden", "authenticated user context missing")
	}

	var req platform.RateSupportTicketRequest
	if err := c.BodyParser(&req); err != nil {
		return writeSupportTicketError(c, fiber.StatusBadRequest, "invalid_request", "invalid rating request")
	}

	if req.Score < 1 || req.Score > 5 || len([]rune(req.Comment)) > 500 {
		return writeSupportTicketError(c, fiber.StatusBadRequest, "invalid_request", "score must be between 1 and 5 and comment must not exceed 500 characters")
	}

	req.IdempotencyKey = supportIdempotencyKey(c, req.IdempotencyKey)
	if req.IdempotencyKey == "" {
		return writeSupportTicketError(c, fiber.StatusBadRequest, "invalid_request", "idempotency key is required")
	}

	store, err := h.requireSupportTicketStore()
	if err != nil {
		return err
	}

	ticket, err := store.RateSupportTicket(c.Context(), currentTenant.ID, c.Params("ticketID"), actorID, req, tenantSupportTicketActor(c))
	if err != nil {
		return mapSupportTicketError(c, err)
	}

	prommetrics.ObserveSupportRating(req.Score)

	return c.JSON(ticket)
}

func (h *Handlers) DownloadTenantSupportTicketAttachment(c *fiber.Ctx) error {
	currentTenant, err := currentTenantOrErr(c)
	if err != nil {
		return err
	}

	actorID := delegatedUserID(c)

	readAll, allowed := tenantTicketReadAccess(c)
	if !allowed || actorID == "" {
		return writeSupportTicketError(c, fiber.StatusForbidden, "support_ticket_forbidden", "support ticket access denied")
	}

	return h.downloadSupportTicketAttachment(c, currentTenant.ID, actorID, readAll, tenantSupportTicketActor(c))
}

func (h *Handlers) ListPlatformSupportTickets(c *fiber.Ctx) error {
	params, err := parseSupportTicketListParams(c)
	if err != nil {
		return err
	}

	params.ReadAll = true

	items, total, err := h.requireSupportTicketStoreAndList(c.Context(), params)
	if err != nil {
		return mapSupportTicketError(c, err)
	}

	return c.JSON(platform.SupportTicketListResponse{
		Data:       items,
		NextCursor: nextOffsetCursor(params.Offset, len(items), total),
	})
}

func (h *Handlers) CountPlatformSupportTickets(c *fiber.Ctx) error {
	params, err := parseSupportTicketListParams(c)
	if err != nil {
		return err
	}

	params.ReadAll = true

	store, err := h.requireSupportTicketStore()
	if err != nil {
		return err
	}

	counts, err := store.CountSupportTickets(c.Context(), params)
	if err != nil {
		return mapSupportTicketError(c, err)
	}

	prommetrics.SetSupportStatusCounts(map[string]int{
		"open": counts.Open, "in_progress": counts.InProgress,
		"waiting_for_customer": counts.WaitingForCustomer,
		"resolved":             counts.Resolved, "closed": counts.Closed,
	})

	return c.JSON(counts)
}

func (h *Handlers) GetPlatformSupportTicket(c *fiber.Ctx) error {
	store, err := h.requireSupportTicketStore()
	if err != nil {
		return err
	}

	detail, err := store.GetSupportTicketDetail(c.Context(), "", c.Params("ticketID"), "", true)
	if err != nil {
		return mapSupportTicketError(c, err)
	}

	return c.JSON(detail)
}

func (h *Handlers) ListPlatformSupportTicketMessages(c *fiber.Ctx) error {
	store, err := h.requireSupportTicketStore()
	if err != nil {
		return err
	}

	ticket, err := store.GetSupportTicket(c.Context(), "", c.Params("ticketID"), "", true)
	if err != nil {
		return mapSupportTicketError(c, err)
	}

	limit, offset, err := parseCursorPagination(c)
	if err != nil {
		return err
	}

	items, err := store.ListSupportTicketMessages(c.Context(), ticket.ID, limit+1, offset)
	if err != nil {
		return mapSupportTicketError(c, err)
	}

	next := ""

	if len(items) > limit {
		items = items[:limit]
		next = encodeOffsetCursor(offset + limit)
	}

	return c.JSON(fiber.Map{"data": items, "next_cursor": next})
}

func (h *Handlers) AddPlatformSupportTicketMessage(c *fiber.Ctx) error {
	req, files, err := parseAddSupportTicketMessageRequest(c)
	if err != nil {
		return mapSupportTicketError(c, err)
	}

	if strings.TrimSpace(req.Body) == "" && len(files) == 0 {
		return writeSupportTicketError(c, fiber.StatusBadRequest, "invalid_request", "message body or attachment is required")
	}

	if len([]rune(req.Body)) > 10000 {
		return writeSupportTicketError(c, fiber.StatusBadRequest, "invalid_request", "message body must not exceed 10000 characters")
	}

	req.IdempotencyKey = supportIdempotencyKey(c, req.IdempotencyKey)
	if req.IdempotencyKey == "" {
		return writeSupportTicketError(c, fiber.StatusBadRequest, "invalid_request", "idempotency key is required")
	}

	if req.NextStatus == "" {
		req.NextStatus = platform.SupportTicketStatusInProgress
	}

	if !validTicketStatus(req.NextStatus) || req.NextStatus == platform.SupportTicketStatusClosed {
		return writeSupportTicketError(c, fiber.StatusBadRequest, "support_ticket_invalid_transition", "invalid next status")
	}

	store, err := h.requireSupportTicketStore()
	if err != nil {
		return err
	}

	ticket, err := store.GetSupportTicket(c.Context(), "", c.Params("ticketID"), "", true)
	if err != nil {
		return mapSupportTicketError(c, err)
	}

	uploads, err := h.uploadSupportFiles(c.Context(), ticket.TenantID, ticket.ID, files)
	if err != nil {
		return mapSupportTicketError(c, err)
	}

	req.RequestHash = hashSupportRequest(req, uploads)

	detail, replay, err := store.AddSupportTicketMessage(
		c.Context(),
		"",
		ticket.ID,
		req,
		uploads,
		platformSupportTicketActor(c),
		true,
		intEnv("SUPPORT_MESSAGES_PER_USER_PER_TICKET_PER_HOUR", 30),
	)
	if err != nil {
		h.deleteSupportUploads(uploads)
		return mapSupportTicketError(c, err)
	}

	if replay {
		h.deleteSupportUploads(uploads)
		c.Set("Idempotency-Replayed", "true")
	} else {
		prommetrics.RecordSupportMessage(platform.SupportTicketAuthorSupportAgent)

		if ticket.FirstSupportResponseAt == nil && detail.Ticket.FirstSupportResponseAt != nil {
			prommetrics.ObserveSupportFirstResponse(detail.Ticket.FirstSupportResponseAt.Sub(ticket.CreatedAt))
		}
	}

	return c.Status(fiber.StatusCreated).JSON(detail)
}

func (h *Handlers) TriagePlatformSupportTicket(c *fiber.Ctx) error {
	var req platform.TriageSupportTicketRequest
	if err := c.BodyParser(&req); err != nil {
		return writeSupportTicketError(c, fiber.StatusBadRequest, "invalid_request", "invalid triage request")
	}

	if req.Status != nil && !validTicketStatus(*req.Status) {
		return writeSupportTicketError(c, fiber.StatusBadRequest, "invalid_request", "invalid ticket status")
	}

	if req.Priority != nil && !validTicketPriority(*req.Priority) {
		return writeSupportTicketError(c, fiber.StatusBadRequest, "invalid_request", "invalid ticket priority")
	}

	actor := platformSupportTicketActor(c)

	if req.AssigneeID != nil {
		value := strings.TrimSpace(*req.AssigneeID)
		switch value {
		case "":
			// Explicitly unassign.
		case "self":
			if actor.ID == "" {
				return writeSupportTicketError(c, fiber.StatusForbidden, "support_ticket_forbidden", "support agent identity is required")
			}

			value = actor.ID
		default:
			if value != actor.ID {
				return writeSupportTicketError(c, fiber.StatusForbidden, "support_ticket_forbidden", "assignee must be the authenticated support agent")
			}
		}

		req.AssigneeID = &value
	}

	req.IdempotencyKey = supportIdempotencyKey(c, req.IdempotencyKey)
	if req.IdempotencyKey == "" {
		return writeSupportTicketError(c, fiber.StatusBadRequest, "invalid_request", "idempotency key is required")
	}

	store, err := h.requireSupportTicketStore()
	if err != nil {
		return err
	}

	ticket, err := store.TriageSupportTicket(c.Context(), c.Params("ticketID"), req, actor)
	if err != nil {
		if errors.Is(err, storage.ErrSupportTicketStaleVersion) {
			prommetrics.RecordSupportTransition("stale")
		} else if errors.Is(err, storage.ErrSupportTicketInvalidTransition) {
			prommetrics.RecordSupportTransition("invalid")
		} else {
			prommetrics.RecordSupportTransition("error")
		}

		return mapSupportTicketError(c, err)
	}

	prommetrics.RecordSupportTransition("success")

	return c.JSON(ticket)
}

func (h *Handlers) ReopenPlatformSupportTicket(c *fiber.Ctx) error {
	var req platform.ReopenSupportTicketRequest
	if err := c.BodyParser(&req); err != nil {
		return writeSupportTicketError(c, fiber.StatusBadRequest, "invalid_request", "invalid reopen request")
	}

	req.IdempotencyKey = supportIdempotencyKey(c, req.IdempotencyKey)
	if req.IdempotencyKey == "" {
		return writeSupportTicketError(c, fiber.StatusBadRequest, "invalid_request", "idempotency key is required")
	}

	store, err := h.requireSupportTicketStore()
	if err != nil {
		return err
	}

	ticket, err := store.ReopenSupportTicket(c.Context(), c.Params("ticketID"), req, platformSupportTicketActor(c))
	if err != nil {
		return mapSupportTicketError(c, err)
	}

	return c.JSON(ticket)
}

func (h *Handlers) ListPlatformSupportTicketEvents(c *fiber.Ctx) error {
	store, err := h.requireSupportTicketStore()
	if err != nil {
		return err
	}

	ticket, err := store.GetSupportTicket(c.Context(), "", c.Params("ticketID"), "", true)
	if err != nil {
		return mapSupportTicketError(c, err)
	}

	limit, offset, err := parseCursorPagination(c)
	if err != nil {
		return err
	}

	items, err := store.ListSupportTicketEvents(c.Context(), ticket.ID, limit+1, offset)
	if err != nil {
		return mapSupportTicketError(c, err)
	}

	next := ""

	if len(items) > limit {
		items = items[:limit]
		next = encodeOffsetCursor(offset + limit)
	}

	return c.JSON(fiber.Map{"data": items, "next_cursor": next})
}

func (h *Handlers) DownloadPlatformSupportTicketAttachment(c *fiber.Ctx) error {
	return h.downloadSupportTicketAttachment(c, "", "", true, platformSupportTicketActor(c))
}

func (h *Handlers) downloadSupportTicketAttachment(
	c *fiber.Ctx,
	tenantID string,
	requesterUserID string,
	readAll bool,
	actor platform.SupportTicketActor,
) error {
	store, err := h.requireSupportTicketStore()
	if err != nil {
		return err
	}

	attachment, err := store.GetSupportTicketAttachment(c.Context(), tenantID, c.Params("attachmentID"), requesterUserID, readAll)
	if err != nil {
		return mapSupportTicketError(c, err)
	}

	if attachment.ScanStatus != "clean" {
		prommetrics.RecordSupportAttachment("not_ready")

		_ = store.RecordSupportTicketAttachmentDownload(c.Context(), attachment, actor, "not_ready") //nolint:errcheck // Failure is intentionally handled by the surrounding best-effort path.

		return writeSupportTicketError(c, fiber.StatusLocked, "support_attachment_not_ready", "attachment is not available")
	}

	if h.supportFiles == nil {
		prommetrics.RecordSupportAttachment("scan_failed")
		return writeSupportTicketError(c, fiber.StatusServiceUnavailable, "support_attachment_scan_failed", "attachment storage is not configured")
	}

	object, err := h.supportFiles.Get(c.Context(), attachment.ObjectKey)
	if err != nil {
		prommetrics.RecordSupportAttachment("storage_error")

		_ = store.RecordSupportTicketAttachmentDownload(c.Context(), attachment, actor, "storage_error") //nolint:errcheck // Failure is intentionally handled by the surrounding best-effort path.

		return writeSupportTicketError(c, fiber.StatusBadGateway, "support_attachment_not_ready", "attachment could not be loaded")
	}

	_ = store.RecordSupportTicketAttachmentDownload(c.Context(), attachment, actor, "allowed") //nolint:errcheck // Failure is intentionally handled by the surrounding best-effort path.

	prommetrics.RecordSupportAttachment("downloaded")

	c.Set(fiber.HeaderContentType, attachment.DetectedContentType)
	c.Set(fiber.HeaderContentDisposition, contentDispositionAttachment(attachment.OriginalFilename))
	c.Set(fiber.HeaderCacheControl, "private, no-store")
	c.Set("X-Content-Type-Options", "nosniff")
	c.Context().SetBodyStream(object.Body, int(attachment.SizeBytes))

	return nil
}

func (h *Handlers) requireSupportTicketStore() (supportTicketStore, error) {
	if h.supportTickets == nil {
		return nil, fiber.NewError(fiber.StatusNotImplemented, "support ticket store not configured")
	}

	return h.supportTickets, nil
}

func (h *Handlers) requireSupportTicketStoreAndList(ctx context.Context, params platform.SupportTicketListParams) ([]platform.SupportTicket, int, error) {
	store, err := h.requireSupportTicketStore()
	if err != nil {
		return nil, 0, err
	}

	return store.ListSupportTickets(ctx, params)
}

func parseSupportTicketListParams(c *fiber.Ctx) (platform.SupportTicketListParams, error) {
	limit, offset, err := parseCursorPagination(c)
	if err != nil {
		return platform.SupportTicketListParams{}, err
	}

	params := platform.SupportTicketListParams{
		Queue:      strings.TrimSpace(c.Query("queue")),
		Status:     strings.TrimSpace(c.Query("status")),
		Category:   strings.TrimSpace(c.Query("category")),
		Priority:   strings.TrimSpace(c.Query("priority")),
		TenantID:   strings.TrimSpace(c.Query("tenant_id")),
		AssigneeID: strings.TrimSpace(c.Query("assignee_id")),
		Query:      strings.TrimSpace(c.Query("query")),
		Limit:      limit,
		Offset:     offset,
	}
	if params.Status != "" && !validTicketStatus(params.Status) {
		return platform.SupportTicketListParams{}, fiber.NewError(fiber.StatusBadRequest, "invalid support ticket status")
	}

	if params.Category != "" && !validTicketCategory(params.Category) {
		return platform.SupportTicketListParams{}, fiber.NewError(fiber.StatusBadRequest, "invalid support ticket category")
	}

	if params.Priority != "" && !validTicketPriority(params.Priority) {
		return platform.SupportTicketListParams{}, fiber.NewError(fiber.StatusBadRequest, "invalid support ticket priority")
	}

	if raw := strings.TrimSpace(c.Query("updated_from")); raw != "" {
		value, parseErr := time.Parse(time.RFC3339, raw)
		if parseErr != nil {
			return platform.SupportTicketListParams{}, fiber.NewError(fiber.StatusBadRequest, "updated_from must be RFC3339")
		}

		params.UpdatedFrom = &value
	}

	if raw := strings.TrimSpace(c.Query("updated_to")); raw != "" {
		value, parseErr := time.Parse(time.RFC3339, raw)
		if parseErr != nil {
			return platform.SupportTicketListParams{}, fiber.NewError(fiber.StatusBadRequest, "updated_to must be RFC3339")
		}

		params.UpdatedTo = &value
	}

	return params, nil
}

func parseCursorPagination(c *fiber.Ctx) (int, int, error) {
	limit := defaultSupportTicketPageSize

	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value <= 0 {
			return 0, 0, fiber.NewError(fiber.StatusBadRequest, "limit must be a positive integer")
		}

		if value > maxSupportTicketPageSize {
			value = maxSupportTicketPageSize
		}

		limit = value
	}

	offset := 0

	if raw := strings.TrimSpace(c.Query("cursor")); raw != "" {
		value, err := decodeOffsetCursor(raw)
		if err != nil {
			return 0, 0, fiber.NewError(fiber.StatusBadRequest, "invalid cursor")
		}

		offset = value
	}

	return limit, offset, nil
}

func encodeOffsetCursor(offset int) string {
	return base64.RawURLEncoding.EncodeToString([]byte(strconv.Itoa(offset)))
}

func decodeOffsetCursor(cursor string) (int, error) {
	raw, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(cursor))
	if err != nil {
		return 0, err
	}

	value, err := strconv.Atoi(string(raw))
	if err != nil || value < 0 {
		return 0, errors.New("invalid cursor")
	}

	return value, nil
}

func nextOffsetCursor(offset, count, total int) string {
	if offset+count >= total {
		return ""
	}

	return encodeOffsetCursor(offset + count)
}

func parseCreateSupportTicketRequest(c *fiber.Ctx) (platform.CreateSupportTicketRequest, []*multipart.FileHeader, error) {
	var req platform.CreateSupportTicketRequest

	files, err := parseSupportMultipart(c, &req)

	return req, files, err
}

func parseAddSupportTicketMessageRequest(c *fiber.Ctx) (platform.AddSupportTicketMessageRequest, []*multipart.FileHeader, error) {
	var req platform.AddSupportTicketMessageRequest

	files, err := parseSupportMultipart(c, &req)

	return req, files, err
}

func parseSupportMultipart(c *fiber.Ctx, destination interface{}) ([]*multipart.FileHeader, error) {
	contentType := strings.ToLower(strings.TrimSpace(c.Get(fiber.HeaderContentType)))
	if !strings.HasPrefix(contentType, "multipart/form-data") {
		if err := c.BodyParser(destination); err != nil {
			return nil, fiber.NewError(fiber.StatusBadRequest, "invalid request body")
		}

		return nil, nil
	}

	form, err := c.MultipartForm()
	if err != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, "invalid multipart request body")
	}

	payload := strings.TrimSpace(c.FormValue("payload"))
	if payload == "" {
		return nil, fiber.NewError(fiber.StatusBadRequest, "multipart payload field is required")
	}

	if err := json.Unmarshal([]byte(payload), destination); err != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, "invalid multipart payload")
	}

	files := form.File["files"]
	if len(files) > supportfiles.MaxFilesPerMessage {
		return nil, supportfiles.ErrTooManyFiles
	}

	return files, nil
}

func validateCreateSupportTicketRequest(req platform.CreateSupportTicketRequest, fileCount int) error {
	subjectLength := len([]rune(strings.TrimSpace(req.Subject)))
	if subjectLength < 5 || subjectLength > 160 {
		return fiber.NewError(fiber.StatusBadRequest, "subject must contain 5 to 160 characters")
	}

	if !validTicketCategory(req.Category) {
		return fiber.NewError(fiber.StatusBadRequest, "invalid support ticket category")
	}

	if len([]rune(req.Body)) > 10000 {
		return fiber.NewError(fiber.StatusBadRequest, "message body must not exceed 10000 characters")
	}

	if strings.TrimSpace(req.Body) == "" && fileCount == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "message body or attachment is required")
	}

	return nil
}

func (h *Handlers) uploadSupportFiles(ctx context.Context, tenantID, scope string, files []*multipart.FileHeader) ([]platform.SupportTicketAttachmentUpload, error) {
	if len(files) == 0 {
		return nil, nil
	}

	if h.supportFiles == nil {
		return nil, supportfiles.ErrNotConfigured
	}

	uploads, err := h.supportFiles.UploadFiles(ctx, tenantID, scope, files)
	if err != nil {
		switch {
		case errors.Is(err, supportfiles.ErrInfected), errors.Is(err, supportfiles.ErrTypeNotAllowed), errors.Is(err, supportfiles.ErrFileTooLarge), errors.Is(err, supportfiles.ErrTooManyFiles):
			prommetrics.RecordSupportAttachment("rejected")
		default:
			prommetrics.RecordSupportAttachment("scan_failed")
		}

		return nil, err
	}

	for range uploads {
		prommetrics.RecordSupportAttachment("uploaded")
	}

	return uploads, nil
}

func (h *Handlers) deleteSupportUploads(uploads []platform.SupportTicketAttachmentUpload) {
	if h.supportFiles == nil || len(uploads) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	h.supportFiles.DeleteUploads(ctx, uploads)
}

func supportIdempotencyKey(c *fiber.Ctx, bodyValue string) string {
	value := strings.TrimSpace(c.Get("Idempotency-Key"))
	if value == "" {
		value = strings.TrimSpace(bodyValue)
	}

	if len(value) > 128 {
		return ""
	}

	return value
}

func hashSupportRequest(value interface{}, uploads []platform.SupportTicketAttachmentUpload) string {
	payload := struct {
		Value       interface{} `json:"value"`
		Attachments []string    `json:"attachments"`
	}{Value: value, Attachments: make([]string, 0, len(uploads))}
	for _, item := range uploads {
		payload.Attachments = append(payload.Attachments, item.SHA256)
	}

	raw, _ := json.Marshal(payload) //nolint:errcheck // The value is composed exclusively of JSON-safe fields.
	hash := sha256.Sum256(raw)

	return hex.EncodeToString(hash[:])
}

func shortHash(value string) string {
	hash := sha256.Sum256([]byte(value))
	return hex.EncodeToString(hash[:8])
}

func tenantSupportTicketActor(c *fiber.Ctx) platform.SupportTicketActor {
	info := auth.GetAuthInfo(c)

	role := ""
	if info != nil {
		role = info.Role
	}

	return platform.SupportTicketActor{
		Type:      platform.SupportTicketAuthorTenantUser,
		ID:        delegatedUserID(c),
		Role:      role,
		RequestID: requestID(c),
	}
}

func platformSupportTicketActor(c *fiber.Ctx) platform.SupportTicketActor {
	return platform.SupportTicketActor{
		Type:      platform.SupportTicketAuthorSupportAgent,
		ID:        strings.TrimSpace(c.Get("x-shahrah-actor-id")),
		Role:      strings.TrimSpace(c.Get("x-shahrah-actor-role")),
		RequestID: requestID(c),
	}
}

func requestID(c *fiber.Ctx) string {
	if value := strings.TrimSpace(c.Get("x-request-id")); value != "" {
		return value
	}

	return apiheaders.GetTraceID(c)
}

func tenantTicketReadAccess(c *fiber.Ctx) (readAll bool, allowed bool) {
	info := auth.GetAuthInfo(c)
	if auth.IsTenantAdminOrHigher(info) {
		return true, true
	}

	if hasPermissionValue(info, "support_ticket:read_all") {
		return true, true
	}

	if hasPermissionValue(info, "support_ticket:read_own") {
		return false, true
	}

	return false, false
}

func tenantTicketHasPermission(c *fiber.Ctx, permission string) bool {
	info := auth.GetAuthInfo(c)
	return auth.IsTenantAdminOrHigher(info) || hasPermissionValue(info, permission)
}

func hasPermissionValue(info *platform.AuthContext, permission string) bool {
	if info == nil {
		return false
	}

	for _, value := range info.Permissions {
		if strings.EqualFold(strings.TrimSpace(value), strings.TrimSpace(permission)) {
			return true
		}
	}

	return false
}

func validTicketStatus(value string) bool {
	switch strings.TrimSpace(value) {
	case platform.SupportTicketStatusOpen,
		platform.SupportTicketStatusInProgress,
		platform.SupportTicketStatusWaitingForCustomer,
		platform.SupportTicketStatusResolved,
		platform.SupportTicketStatusClosed:
		return true
	default:
		return false
	}
}

func validTicketPriority(value string) bool {
	switch strings.TrimSpace(value) {
	case platform.SupportTicketPriorityLow,
		platform.SupportTicketPriorityNormal,
		platform.SupportTicketPriorityHigh,
		platform.SupportTicketPriorityUrgent:
		return true
	default:
		return false
	}
}

func validTicketCategory(value string) bool {
	switch strings.TrimSpace(value) {
	case platform.SupportTicketCategoryTechnical,
		platform.SupportTicketCategoryBilling,
		platform.SupportTicketCategoryAccount,
		platform.SupportTicketCategoryFeatureRequest,
		platform.SupportTicketCategoryOther:
		return true
	default:
		return false
	}
}

func contentDispositionAttachment(filename string) string {
	filename = strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(filename), "\r", "_"), "\n", "_")
	escaped := url.QueryEscape(filename)

	return fmt.Sprintf("attachment; filename=\"attachment\"; filename*=UTF-8''%s", escaped)
}

func intEnv(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return fallback
	}

	return value
}

func writeSupportTicketError(c *fiber.Ctx, status int, code, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"error": fiber.Map{
			"code":    code,
			"message": message,
			"type":    "support_ticket_error",
		},
	})
}

func mapSupportTicketError(c *fiber.Ctx, err error) error {
	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		return writeSupportTicketError(c, fiberErr.Code, "invalid_request", fiberErr.Message)
	}

	switch {
	case errors.Is(err, storage.ErrSupportTicketNotFound):
		return writeSupportTicketError(c, fiber.StatusNotFound, "support_ticket_not_found", "support ticket not found")
	case errors.Is(err, storage.ErrSupportTicketForbidden):
		return writeSupportTicketError(c, fiber.StatusForbidden, "support_ticket_forbidden", "support ticket access denied")
	case errors.Is(err, storage.ErrSupportTicketClosed):
		return writeSupportTicketError(c, fiber.StatusConflict, "support_ticket_closed", "support ticket is closed")
	case errors.Is(err, storage.ErrSupportTicketInvalidTransition):
		return writeSupportTicketError(c, fiber.StatusUnprocessableEntity, "support_ticket_invalid_transition", "invalid support ticket status transition")
	case errors.Is(err, storage.ErrSupportTicketRatingNotAllowed):
		return writeSupportTicketError(c, fiber.StatusUnprocessableEntity, "support_ticket_rating_not_allowed", "support ticket rating is not allowed")
	case errors.Is(err, storage.ErrSupportTicketStaleVersion):
		return writeSupportTicketError(c, fiber.StatusConflict, "support_ticket_stale_version", "support ticket changed; refresh and retry")
	case errors.Is(err, storage.ErrSupportTicketIdempotency):
		return writeSupportTicketError(c, fiber.StatusConflict, "idempotency_conflict", "idempotency key was already used with another payload")
	case errors.Is(err, storage.ErrSupportTicketRateLimited):
		return writeSupportTicketError(c, fiber.StatusTooManyRequests, "rate_limit_exceeded", "support ticket rate limit exceeded")
	case errors.Is(err, supportfiles.ErrTooManyFiles):
		return writeSupportTicketError(c, fiber.StatusBadRequest, "support_attachment_too_many", "too many attachments")
	case errors.Is(err, supportfiles.ErrFileTooLarge):
		return writeSupportTicketError(c, fiber.StatusRequestEntityTooLarge, "support_attachment_too_large", "attachment size limit exceeded")
	case errors.Is(err, supportfiles.ErrTypeNotAllowed):
		return writeSupportTicketError(c, fiber.StatusUnsupportedMediaType, "support_attachment_type_not_allowed", "attachment type is not allowed")
	case errors.Is(err, supportfiles.ErrInfected):
		return writeSupportTicketError(c, fiber.StatusUnprocessableEntity, "support_attachment_infected", "attachment was rejected")
	case errors.Is(err, supportfiles.ErrScanFailed), errors.Is(err, supportfiles.ErrNotConfigured):
		return writeSupportTicketError(c, fiber.StatusServiceUnavailable, "support_attachment_scan_failed", "attachment scanning is unavailable")
	default:
		return err
	}
}
