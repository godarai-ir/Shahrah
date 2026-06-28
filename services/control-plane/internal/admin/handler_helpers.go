package admin

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

type tenantListParams struct {
	Query  string
	Limit  int
	Offset int
}

func currentTenantOrErr(c *fiber.Ctx) (*platform.Tenant, error) {
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}

	return currentTenant, nil
}

func parseTenantListParams(c *fiber.Ctx, maxLimit int) (tenantListParams, error) {
	params := tenantListParams{
		Query: strings.TrimSpace(c.Query("query")),
		Limit: 50,
	}

	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		parsed, parseErr := strconv.Atoi(raw)
		if parseErr != nil || parsed <= 0 {
			return tenantListParams{}, fiber.NewError(fiber.StatusBadRequest, "limit must be a positive number")
		}

		if maxLimit > 0 && parsed > maxLimit {
			parsed = maxLimit
		}

		params.Limit = parsed
	}

	if raw := strings.TrimSpace(c.Query("offset")); raw != "" {
		parsed, parseErr := strconv.Atoi(raw)
		if parseErr != nil || parsed < 0 {
			return tenantListParams{}, fiber.NewError(fiber.StatusBadRequest, "offset must be zero or greater")
		}

		params.Offset = parsed
	}

	return params, nil
}

func buildNextOffset(offset, itemCount, total int) *int {
	if offset+itemCount >= total {
		return nil
	}

	next := offset + itemCount

	return &next
}

func parseTemporalBody(c *fiber.Ctx, dest interface{}, setTemporal func(effectiveFrom, expiresAt *time.Time)) error {
	var raw map[string]interface{}
	if err := c.BodyParser(&raw); err != nil {
		return err
	}

	effectiveFrom, err := parseOptionalTimeValue(raw["effective_from"])
	if err != nil {
		return err
	}

	expiresAt, err := parseOptionalTimeValue(raw["expires_at"])
	if err != nil {
		return err
	}

	raw["effective_from"] = nil
	raw["expires_at"] = nil

	encoded, err := json.Marshal(raw)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(encoded, dest); err != nil {
		return err
	}

	setTemporal(effectiveFrom, expiresAt)

	return nil
}

func handleScopedEnabledUpdate(
	c *fiber.Ctx,
	notFoundMessage string,
	update func(ctx context.Context, tenantID, itemID string, enabled bool) (interface{}, error),
) error {
	currentTenant, err := currentTenantOrErr(c)
	if err != nil {
		return err
	}

	itemID := strings.TrimSpace(c.Params("id"))
	if itemID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "id is required")
	}

	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	item, err := update(c.Context(), currentTenant.ID, itemID, req.Enabled)
	if err != nil {
		if strings.TrimSpace(notFoundMessage) != "" && errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, notFoundMessage)
		}

		return err
	}

	return c.JSON(item)
}

func parseBodyInto[T any](c *fiber.Ctx) (T, error) {
	var req T
	if err := c.BodyParser(&req); err != nil {
		return req, fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	return req, nil
}

func handleTenantScopedUpdate[T any](
	c *fiber.Ctx,
	paramName string,
	notFoundMessage string,
	parseBody func(*fiber.Ctx) (T, error),
	update func(ctx context.Context, tenantID, itemID string, req T) (interface{}, error),
) error {
	currentTenant, err := currentTenantOrErr(c)
	if err != nil {
		return err
	}

	itemID := strings.TrimSpace(c.Params(paramName))
	if itemID == "" {
		return fiber.NewError(fiber.StatusBadRequest, paramName+" is required")
	}

	req, err := parseBody(c)
	if err != nil {
		return err
	}

	item, err := update(c.Context(), currentTenant.ID, itemID, req)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, notFoundMessage)
		}

		return err
	}

	return c.JSON(item)
}

func handleTenantMembershipDelete(
	c *fiber.Ctx,
	parentParam string,
	notFoundMessage string,
	deleteFn func(ctx context.Context, tenantID, parentID, userID string) error,
) error {
	currentTenant, err := currentTenantOrErr(c)
	if err != nil {
		return err
	}

	parentID := strings.TrimSpace(c.Params(parentParam))
	userID := strings.TrimSpace(c.Params("userID"))
	if parentID == "" || userID == "" {
		return fiber.NewError(fiber.StatusBadRequest, parentParam+" and userID are required")
	}

	if err := deleteFn(c.Context(), currentTenant.ID, parentID, userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, notFoundMessage)
		}

		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}
