package admin

import (
	"errors"
	"strconv"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/prompts"
	"github.com/Delpak/shahrah/services/control-plane/internal/storage"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgconn"
)

func (h *Handlers) ListPrompts(c *fiber.Ctx) error {
	if err := requirePolicyListOrManage(c); err != nil {
		return err
	}

	store, err := h.requirePromptStore()
	if err != nil {
		return err
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}

	query := strings.TrimSpace(c.Query("query"))

	limit, offset, err := parsePromptLimitOffset(c, 50)
	if err != nil {
		return err
	}

	items, total, err := store.ListPrompts(c.Context(), currentTenant.ID, query, limit, offset)
	if err != nil {
		return err
	}

	return c.JSON(platform.PromptListResponse{Data: items, Total: total, Limit: limit, Offset: offset, NextOffset: nextOffset(offset, len(items), total)})
}

func (h *Handlers) CreatePrompt(c *fiber.Ctx) error {
	if err := requirePolicyManage(c); err != nil {
		return err
	}

	store, err := h.requirePromptStore()
	if err != nil {
		return err
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}

	var req platform.CreatePromptRequest

	parseErr := c.BodyParser(&req)
	if parseErr != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	actor := strings.TrimSpace(c.Get("x-shahrah-user-id"))

	item, err := store.CreatePrompt(c.Context(), currentTenant.ID, currentTenant.Slug, req, actor)
	if err != nil {
		return mapPromptStoreError(err)
	}

	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *Handlers) GetPrompt(c *fiber.Ctx) error {
	if err := requirePolicyListOrManage(c); err != nil {
		return err
	}

	store, err := h.requirePromptStore()
	if err != nil {
		return err
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}

	promptID := strings.TrimSpace(c.Params("promptID"))
	if promptID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "prompt_id is required")
	}

	item, err := store.GetPrompt(c.Context(), currentTenant.ID, promptID)
	if err != nil {
		return mapPromptStoreError(err)
	}

	return c.JSON(item)
}

func (h *Handlers) DeletePrompt(c *fiber.Ctx) error {
	if err := requirePolicyManage(c); err != nil {
		return err
	}

	store, err := h.requirePromptStore()
	if err != nil {
		return err
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}

	promptID := strings.TrimSpace(c.Params("promptID"))
	if promptID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "prompt_id is required")
	}

	if err := store.DeletePrompt(c.Context(), currentTenant.ID, promptID); err != nil {
		return mapPromptStoreError(err)
	}

	return c.JSON(fiber.Map{"success": true})
}

func (h *Handlers) ListPromptVersions(c *fiber.Ctx) error {
	if err := requirePolicyListOrManage(c); err != nil {
		return err
	}

	store, err := h.requirePromptStore()
	if err != nil {
		return err
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}

	promptID := strings.TrimSpace(c.Params("promptID"))
	if promptID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "prompt_id is required")
	}

	limit, offset, err := parsePromptLimitOffset(c, 50)
	if err != nil {
		return err
	}

	items, total, err := store.ListPromptVersions(c.Context(), currentTenant.ID, promptID, limit, offset)
	if err != nil {
		return err
	}

	return c.JSON(platform.PromptVersionListResponse{Data: items, Total: total, Limit: limit, Offset: offset, NextOffset: nextOffset(offset, len(items), total)})
}

func (h *Handlers) CreatePromptVersion(c *fiber.Ctx) error {
	if err := requirePolicyManage(c); err != nil {
		return err
	}

	store, err := h.requirePromptStore()
	if err != nil {
		return err
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}

	promptID := strings.TrimSpace(c.Params("promptID"))
	if promptID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "prompt_id is required")
	}

	var req platform.CreatePromptVersionRequest

	parseErr := c.BodyParser(&req)
	if parseErr != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	actor := strings.TrimSpace(c.Get("x-shahrah-user-id"))

	item, err := store.CreatePromptVersion(c.Context(), currentTenant.ID, currentTenant.Slug, promptID, req, actor)
	if err != nil {
		return mapPromptStoreError(err)
	}

	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *Handlers) GetPromptVersion(c *fiber.Ctx) error {
	if err := requirePolicyListOrManage(c); err != nil {
		return err
	}

	store, err := h.requirePromptStore()
	if err != nil {
		return err
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}

	versionID := strings.TrimSpace(c.Params("versionID"))
	if versionID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "version_id is required")
	}

	item, err := store.GetPromptVersion(c.Context(), currentTenant.ID, versionID)
	if err != nil {
		return mapPromptStoreError(err)
	}

	return c.JSON(item)
}

func (h *Handlers) RenderPromptVersion(c *fiber.Ctx) error {
	if err := requirePolicyListOrManage(c); err != nil {
		return err
	}

	store, err := h.requirePromptStore()
	if err != nil {
		return err
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}

	versionID := strings.TrimSpace(c.Params("versionID"))
	if versionID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "version_id is required")
	}

	var req platform.RenderPromptVersionRequest
	if len(c.Body()) > 0 {
		parseErr := c.BodyParser(&req)
		if parseErr != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
		}
	}

	version, err := store.GetPromptVersion(c.Context(), currentTenant.ID, versionID)
	if err != nil {
		return mapPromptStoreError(err)
	}

	rendered, err := prompts.Render(*version, req.Variables, req.Messages, req.Model)
	if err != nil {
		return mapPromptStoreError(err)
	}

	return c.JSON(rendered)
}

func (h *Handlers) requirePromptStore() (promptStore, error) {
	if h.prompts == nil {
		return nil, fiber.NewError(fiber.StatusNotImplemented, "prompt store not configured")
	}

	return h.prompts, nil
}

func parsePromptLimitOffset(c *fiber.Ctx, defaultLimit int) (int, int, error) {
	limit := defaultLimit
	offset := 0

	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		parsed, parseErr := strconv.Atoi(raw)
		if parseErr != nil || parsed <= 0 {
			return 0, 0, fiber.NewError(fiber.StatusBadRequest, "limit must be a positive number")
		}

		limit = parsed
	}

	if raw := strings.TrimSpace(c.Query("offset")); raw != "" {
		parsed, parseErr := strconv.Atoi(raw)
		if parseErr != nil || parsed < 0 {
			return 0, 0, fiber.NewError(fiber.StatusBadRequest, "offset must be zero or greater")
		}

		offset = parsed
	}

	return limit, offset, nil
}

func nextOffset(offset int, count int, total int) *int {
	if offset+count >= total {
		return nil
	}

	next := offset + count

	return &next
}

func mapPromptStoreError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, storage.ErrPromptNotFound) {
		return fiber.NewError(fiber.StatusNotFound, "prompt not found")
	}

	if errors.Is(err, prompts.ErrInvalidManifest) || errors.Is(err, prompts.ErrMissingVariable) {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return fiber.NewError(fiber.StatusConflict, "prompt already exists")
		case "23514", "23502", "22P02":
			return fiber.NewError(fiber.StatusBadRequest, "invalid prompt payload")
		case "42P01", "42703":
			return fiber.NewError(fiber.StatusServiceUnavailable, "prompt storage migration has not been applied")
		}
	}

	return err
}
