package admin

import (
	"errors"
	"regexp"
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

	currentTenant, err := currentTenantOrErr(c)
	if err != nil {
		return err
	}

	params, err := parseTenantListParams(c, 0)
	if err != nil {
		return err
	}

	items, total, err := h.models.ListTenantGuardrails(c.Context(), currentTenant.ID, params.Query, params.Limit, params.Offset)
	if err != nil {
		return err
	}

	return c.JSON(platform.TenantGuardrailListResponse{
		Data:       items,
		Total:      total,
		Limit:      params.Limit,
		Offset:     params.Offset,
		NextOffset: buildNextOffset(params.Offset, len(items), total),
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
	case "keyword_blocklist", "regex", "secrets_detection", "basic_pii",
		"jwt_validation", "regex_replace", "sentence_count", "word_count", "character_count",
		"json_schema", "json_keys", "contains_code", "not_null", "inline_image_urls":
		return nil
	default:
		return errors.New("unsupported guardrail type")
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

	switch guardrailType {
	case "sentence_count", "word_count", "character_count", "json_schema", "json_keys", "contains_code", "not_null", "jwt_validation":
		if value != platform.GuardrailOperationValidate {
			return errors.New("this guardrail type only supports validate operation")
		}
	case "regex_replace", "inline_image_urls":
		if value != platform.GuardrailOperationMutate {
			return errors.New("this guardrail type only supports mutate operation")
		}
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
	case "regex_replace":
		pattern := strings.TrimSpace(stringFromAny(config["pattern"]))
		if pattern == "" {
			return errors.New("config.pattern is required")
		}

		if _, err := regexp.Compile(pattern); err != nil {
			return errors.New("config.pattern contains an invalid regex")
		}
	case "sentence_count", "word_count", "character_count":
		minValue, hasMin := intFromAny(config["min"])

		maxValue, hasMax := intFromAny(config["max"])
		if !hasMin && !hasMax {
			return errors.New("config.min or config.max is required")
		}

		if hasMin && minValue < 0 {
			return errors.New("config.min must be zero or greater")
		}

		if hasMax && maxValue < 0 {
			return errors.New("config.max must be zero or greater")
		}

		if hasMin && hasMax && minValue > maxValue {
			return errors.New("config.min must be less than or equal to config.max")
		}
	case "json_schema":
		schema, ok := config["schema"].(map[string]interface{})
		if !ok || len(schema) == 0 {
			return errors.New("config.schema is required")
		}
	case "json_keys":
		if len(stringListFromConfig(config, "keys")) == 0 {
			return errors.New("config.keys must contain at least one key")
		}

		mode := strings.TrimSpace(stringFromAny(config["mode"]))
		if mode == "" {
			mode = "all"
		}

		if mode != "any" && mode != "all" && mode != "none" {
			return errors.New("config.mode must be any, all, or none")
		}
	case "contains_code":
		mode := strings.TrimSpace(stringFromAny(config["mode"]))
		if mode != "" && mode != "must_contain" && mode != "must_not_contain" {
			return errors.New("config.mode must be must_contain or must_not_contain")
		}
	case "jwt_validation":
		inlineJWKS, hasInlineJWKS := config["inline_jwks"].(map[string]interface{})

		hasInlineJWKS = hasInlineJWKS && jwksHasKeys(inlineJWKS)
		if strings.TrimSpace(stringFromAny(config["shared_secret"])) == "" &&
			strings.TrimSpace(stringFromAny(config["jwks_uri"])) == "" &&
			!hasInlineJWKS &&
			strings.TrimSpace(stringFromAny(config["introspection_url"])) == "" {
			return errors.New("JWT validation requires shared_secret, inline_jwks, jwks_uri, or introspection_url")
		}
	case "inline_image_urls":
		if maxBytes, ok := intFromAny(config["max_bytes"]); ok && maxBytes <= 0 {
			return errors.New("config.max_bytes must be positive")
		}
	}

	return nil
}

func jwksHasKeys(jwks map[string]interface{}) bool {
	switch keys := jwks["keys"].(type) {
	case []interface{}:
		return len(keys) > 0
	case []map[string]interface{}:
		return len(keys) > 0
	default:
		return false
	}
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

func stringFromAny(value interface{}) string {
	if text, ok := value.(string); ok {
		return text
	}

	return ""
}

func intFromAny(value interface{}) (int, bool) {
	switch typed := value.(type) {
	case int:
		return typed, true
	case int64:
		return int(typed), true
	case float64:
		return int(typed), true
	default:
		return 0, false
	}
}
