package admin

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

var tenantGuardrailSlugPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{1,95}$`)

func (h *Handlers) ListTenantGuardrails(c *fiber.Ctx) error {
	if err := requirePolicyListOrManage(c); err != nil {
		return err
	}
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	query := strings.TrimSpace(c.Query("query"))
	limit := 50
	offset := 0
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		parsed, parseErr := strconv.Atoi(raw)
		if parseErr != nil || parsed <= 0 {
			return fiber.NewError(fiber.StatusBadRequest, "limit must be a positive number")
		}
		limit = parsed
	}
	if raw := strings.TrimSpace(c.Query("offset")); raw != "" {
		parsed, parseErr := strconv.Atoi(raw)
		if parseErr != nil || parsed < 0 {
			return fiber.NewError(fiber.StatusBadRequest, "offset must be zero or greater")
		}
		offset = parsed
	}
	items, total, err := h.models.ListTenantGuardrails(c.Context(), currentTenant.ID, query, limit, offset)
	if err != nil {
		return err
	}
	var nextOffset *int
	if offset+len(items) < total {
		n := offset + len(items)
		nextOffset = &n
	}
	return c.JSON(platform.TenantGuardrailListResponse{
		Data:       items,
		Total:      total,
		Limit:      limit,
		Offset:     offset,
		NextOffset: nextOffset,
	})
}

func (h *Handlers) CreateTenantGuardrail(c *fiber.Ctx) error {
	if err := requirePolicyManage(c); err != nil {
		return err
	}
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	var req platform.CreateTenantGuardrailRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if err := validateCreateTenantGuardrailRequest(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	actor := strings.TrimSpace(c.Get("x-shahrah-user-id"))
	item, err := h.models.CreateTenantGuardrail(c.Context(), currentTenant.ID, req, actor)
	if err != nil {
		return mapProviderAccountStoreError(err)
	}
	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *Handlers) GetTenantGuardrail(c *fiber.Ctx) error {
	if err := requirePolicyListOrManage(c); err != nil {
		return err
	}
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	slug := strings.TrimSpace(c.Params("slug"))
	if slug == "" {
		return fiber.NewError(fiber.StatusBadRequest, "slug is required")
	}
	item, err := h.models.GetTenantGuardrailBySlug(c.Context(), currentTenant.ID, slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "guardrail not found")
		}
		return err
	}
	return c.JSON(item)
}

func (h *Handlers) UpdateTenantGuardrail(c *fiber.Ctx) error {
	if err := requirePolicyManage(c); err != nil {
		return err
	}
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	slug := strings.TrimSpace(c.Params("slug"))
	if slug == "" {
		return fiber.NewError(fiber.StatusBadRequest, "slug is required")
	}
	var req platform.UpdateTenantGuardrailRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if err := validateUpdateTenantGuardrailRequest(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	actor := strings.TrimSpace(c.Get("x-shahrah-user-id"))
	item, err := h.models.UpdateTenantGuardrailBySlug(c.Context(), currentTenant.ID, slug, req, actor)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "guardrail not found")
		}
		return mapProviderAccountStoreError(err)
	}
	return c.JSON(item)
}

func (h *Handlers) DeleteTenantGuardrail(c *fiber.Ctx) error {
	if err := requirePolicyManage(c); err != nil {
		return err
	}
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	slug := strings.TrimSpace(c.Params("slug"))
	if slug == "" {
		return fiber.NewError(fiber.StatusBadRequest, "slug is required")
	}
	if err := h.models.DeleteTenantGuardrailBySlug(c.Context(), currentTenant.ID, slug); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "guardrail not found")
		}
		return err
	}
	return c.JSON(fiber.Map{"success": true})
}

func validateCreateTenantGuardrailRequest(req platform.CreateTenantGuardrailRequest) error {
	if !tenantGuardrailSlugPattern.MatchString(strings.TrimSpace(req.Slug)) {
		return errors.New("slug must start with a letter or number and contain 2-96 letters, numbers, dashes, or underscores")
	}
	if strings.TrimSpace(req.Name) == "" {
		return errors.New("name is required")
	}
	if err := validateTenantGuardrailType(strings.TrimSpace(req.Type)); err != nil {
		return err
	}
	if err := validateTenantGuardrailOperation(req.Operation, strings.TrimSpace(req.Type)); err != nil {
		return err
	}
	if err := validateTenantGuardrailEnforcement(req.Enforcement); err != nil {
		return err
	}
	return validateTenantGuardrailConfig(strings.TrimSpace(req.Type), req.Config)
}

func validateUpdateTenantGuardrailRequest(req platform.UpdateTenantGuardrailRequest) error {
	guardrailType := ""
	if req.Type != nil {
		guardrailType = strings.TrimSpace(*req.Type)
		if err := validateTenantGuardrailType(guardrailType); err != nil {
			return err
		}
	}
	if req.Operation != nil {
		if err := validateTenantGuardrailOperation(*req.Operation, guardrailType); err != nil {
			return err
		}
	}
	if req.Enforcement != nil {
		if err := validateTenantGuardrailEnforcement(*req.Enforcement); err != nil {
			return err
		}
	}
	if req.Config != nil && guardrailType != "" {
		return validateTenantGuardrailConfig(guardrailType, req.Config)
	}
	return nil
}

func validateTenantGuardrailType(value string) error {
	switch value {
	case "keyword_blocklist", "regex", "secrets_detection", "basic_pii":
		return nil
	default:
		return errors.New("type must be keyword_blocklist, regex, secrets_detection, or basic_pii")
	}
}

func validateTenantGuardrailOperation(value platform.GuardrailOperation, guardrailType string) error {
	switch value {
	case platform.GuardrailOperationValidate, platform.GuardrailOperationMutate:
	default:
		return errors.New("operation must be validate or mutate")
	}
	if guardrailType == "keyword_blocklist" && value != platform.GuardrailOperationValidate {
		return errors.New("keyword_blocklist only supports validate operation")
	}
	return nil
}

func validateTenantGuardrailEnforcement(value platform.GuardrailEnforcement) error {
	switch value {
	case platform.GuardrailEnforcementAudit, platform.GuardrailEnforcementEnforce, platform.GuardrailEnforcementEnforceIgnoreError:
		return nil
	default:
		return errors.New("enforcement must be audit, enforce, or enforce_ignore_error")
	}
}

func validateTenantGuardrailConfig(guardrailType string, config map[string]interface{}) error {
	switch guardrailType {
	case "keyword_blocklist":
		if len(stringListFromConfig(config, "terms")) == 0 {
			return errors.New("config.terms must contain at least one term")
		}
	case "regex":
		patterns := stringListFromConfig(config, "patterns")
		if len(patterns) == 0 {
			return errors.New("config.patterns must contain at least one regex")
		}
		for _, pattern := range patterns {
			if _, err := regexp.Compile(pattern); err != nil {
				return errors.New("config.patterns contains an invalid regex")
			}
		}
	}
	return nil
}

func stringListFromConfig(config map[string]interface{}, key string) []string {
	raw, ok := config[key]
	if !ok {
		return nil
	}
	switch values := raw.(type) {
	case []string:
		out := make([]string, 0, len(values))
		for _, value := range values {
			if trimmed := strings.TrimSpace(value); trimmed != "" {
				out = append(out, trimmed)
			}
		}
		return out
	case []interface{}:
		out := make([]string, 0, len(values))
		for _, value := range values {
			if text, ok := value.(string); ok {
				if trimmed := strings.TrimSpace(text); trimmed != "" {
					out = append(out, trimmed)
				}
			}
		}
		return out
	default:
		return nil
	}
}
