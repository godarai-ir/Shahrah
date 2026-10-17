package admin

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/apperrors"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/credits"
	"github.com/Delpak/shahrah/services/control-plane/internal/provideraccounts"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/Delpak/shahrah/services/control-plane/internal/virtualmodels"
	"github.com/gofiber/fiber/v2"
)

type createModelRequest struct {
	ModelName  string                 `json:"model_name"`
	Provider   string                 `json:"provider"`
	Model      string                 `json:"model"`
	APIBase    string                 `json:"api_base"`
	APIKey     string                 `json:"api_key"`
	APIVersion string                 `json:"api_version"`
	RegionName string                 `json:"region_name"`
	RPM        int64                  `json:"rpm"`
	TPM        int64                  `json:"tpm"`
	Weight     float64                `json:"weight"`
	TimeoutSec float64                `json:"timeout"`
	ModelInfo  map[string]interface{} `json:"model_info"`
}

// CreateModel creates a deployment for the current tenant.
func (h *Handlers) CreateModel(c *fiber.Ctx) error {
	if err := requireTenantAdminOrHigher(c); err != nil {
		return err
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}

	var req createModelRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	dep, err := h.models.CreateDeployment(c.Context(), platform.Deployment{
		TenantID:   currentTenant.ID,
		ModelName:  req.ModelName,
		Provider:   req.Provider,
		Model:      req.Model,
		APIBase:    req.APIBase,
		APIKey:     req.APIKey,
		APIVersion: req.APIVersion,
		RegionName: req.RegionName,
		RPM:        req.RPM,
		TPM:        req.TPM,
		Weight:     req.Weight,
		TimeoutSec: req.TimeoutSec,
		ModelInfo:  req.ModelInfo,
	})
	if err != nil {
		if errors.Is(err, provideraccounts.ErrInvalidAPIBase) {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}

		return err
	}

	return c.Status(fiber.StatusCreated).JSON(dep)
}

// ListModels lists tenant-visible model aliases.
func (h *Handlers) ListModels(c *fiber.Ctx) error {
	if err := requirePermissionAnyOrTenantAdmin(c, "model list permission required", []string{
		"playground:read",
		"playground:run",
		"policy:list",
		"policy:manage",
		"provider_account:read",
		"provider_account:manage",
		"provider_account:create",
		"virtual_account:read",
		"virtual_account:manage",
	}); err != nil {
		return err
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}

	deployments, err := h.models.ListDeployments(c.Context(), currentTenant.ID)
	if err != nil {
		return err
	}

	byAlias := make(map[string]openai.Model, len(deployments))
	for _, dep := range deployments {
		id := dep.ModelName
		if id == "" {
			id = dep.ID
		}

		model := byAlias[id]
		model.ID = id
		model.Object = "model"

		model.OwnedBy = "shahrah"
		if model.Provider == "" {
			model.Provider = dep.Provider
		}

		model.Aliases = appendUnique(model.Aliases, dep.ModelName)

		model.Aliases = appendUnique(model.Aliases, dep.Model)
		for _, mode := range supportedModesFromModelInfo(dep.ModelInfo) {
			model.SupportedModes = appendUnique(model.SupportedModes, mode)
		}

		byAlias[id] = model
	}

	virtualModels, err := h.models.ListVirtualModels(c.Context(), currentTenant.ID)
	if err != nil {
		return err
	}

	for _, vm := range virtualModels {
		alias := strings.TrimSpace(vm.Alias)
		if alias == "" || !vm.Enabled {
			continue
		}

		model := byAlias[alias]
		model.ID = alias
		model.Object = "model"
		model.OwnedBy = "shahrah"
		model.Provider = "virtual"

		model.Aliases = appendUnique(model.Aliases, alias)
		for _, mode := range vm.ExposedModes {
			model.SupportedModes = appendUnique(model.SupportedModes, mode)
		}

		byAlias[alias] = model
	}

	data := make([]openai.Model, 0, len(byAlias))
	for _, item := range byAlias {
		data = append(data, openai.Model{
			ID:             item.ID,
			Object:         "model",
			OwnedBy:        item.OwnedBy,
			Provider:       item.Provider,
			Aliases:        item.Aliases,
			SupportedModes: item.SupportedModes,
		})
	}

	return c.JSON(openai.ModelListResponse{Object: "list", Data: data})
}

func supportedModesFromModelInfo(info map[string]interface{}) []string {
	if len(info) == 0 {
		return nil
	}

	value, ok := info["supported_modes"]
	if !ok {
		return nil
	}

	switch modes := value.(type) {
	case []string:
		return append([]string(nil), modes...)
	case []interface{}:
		out := make([]string, 0, len(modes))
		for _, mode := range modes {
			if text, ok := mode.(string); ok && strings.TrimSpace(text) != "" {
				out = append(out, text)
			}
		}

		return out
	default:
		return nil
	}
}

func (h *Handlers) CreateVirtualModel(c *fiber.Ctx) error {
	if err := requirePermissionAnyOrTenantAdmin(c, "virtual account manage permission required", []string{
		"virtual_account:manage",
	}); err != nil {
		return err
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}

	var req platform.CreateVirtualModelRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	if err := validateCreateVirtualModelRequest(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	actor := delegatedUserID(c)

	item, err := h.models.CreateVirtualModel(c.Context(), currentTenant.ID, req, actor)
	if err != nil {
		if errors.Is(err, virtualmodels.ErrInvalidModes) {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}

		return err
	}

	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *Handlers) ListVirtualModels(c *fiber.Ctx) error {
	if err := requirePermissionAnyOrTenantAdmin(c, "virtual account read permission required", []string{
		"virtual_account:read",
		"virtual_account:manage",
		"policy:list",
		"policy:manage",
	}); err != nil {
		return err
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}

	items, err := h.models.ListVirtualModels(c.Context(), currentTenant.ID)
	if err != nil {
		return err
	}

	if items == nil {
		items = make([]platform.VirtualModel, 0)
	}

	return c.JSON(fiber.Map{"data": items})
}

func (h *Handlers) GetVirtualModel(c *fiber.Ctx) error {
	if err := requirePermissionAnyOrTenantAdmin(c, "virtual account read permission required", []string{
		"virtual_account:read",
		"virtual_account:manage",
		"policy:list",
		"policy:manage",
	}); err != nil {
		return err
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}

	id := strings.TrimSpace(c.Params("id"))
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "id is required")
	}

	item, err := h.models.GetVirtualModel(c.Context(), currentTenant.ID, id)
	if err != nil {
		if errors.Is(err, virtualmodels.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "virtual model not found")
		}

		return err
	}

	return c.JSON(item)
}

func (h *Handlers) UpdateVirtualModel(c *fiber.Ctx) error {
	if err := requirePermissionAnyOrTenantAdmin(c, "virtual account manage permission required", []string{
		"virtual_account:manage",
	}); err != nil {
		return err
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}

	id := strings.TrimSpace(c.Params("id"))
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "id is required")
	}

	var req platform.UpdateVirtualModelRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	currentStrategy := "priority"

	if req.RoutingStrategy == nil && len(req.Targets) > 0 {
		existing, err := h.models.GetVirtualModel(c.Context(), currentTenant.ID, id)
		if err != nil {
			if errors.Is(err, virtualmodels.ErrNotFound) {
				return fiber.NewError(fiber.StatusNotFound, "virtual model not found")
			}

			return err
		}

		currentStrategy = existing.RoutingStrategy
	}

	if err := validateUpdateVirtualModelRequest(req, currentStrategy); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	actor := delegatedUserID(c)

	item, err := h.models.UpdateVirtualModel(c.Context(), currentTenant.ID, id, req, actor)
	if err != nil {
		if errors.Is(err, virtualmodels.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "virtual model not found")
		}

		if errors.Is(err, virtualmodels.ErrInvalidModes) {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}

		return err
	}

	return c.JSON(item)
}

func (h *Handlers) DeleteVirtualModel(c *fiber.Ctx) error {
	if err := requirePermissionAnyOrTenantAdmin(c, "virtual account delete permission required", []string{
		"virtual_account:delete",
		"virtual_account:manage",
	}); err != nil {
		return err
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}

	id := strings.TrimSpace(c.Params("id"))
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "id is required")
	}

	if err := h.models.DeleteVirtualModel(c.Context(), currentTenant.ID, id); err != nil {
		if errors.Is(err, virtualmodels.ErrNotFound) || errors.Is(err, apperrors.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "virtual model not found")
		}

		return err
	}

	return c.JSON(fiber.Map{"success": true})
}

func validateCreateVirtualModelRequest(req platform.CreateVirtualModelRequest) error {
	if strings.TrimSpace(req.Alias) == "" {
		return errors.New("alias is required")
	}

	if strings.TrimSpace(req.RoutingStrategy) == "" {
		req.RoutingStrategy = "priority"
	}

	strategy := normalizeVirtualModelRoutingStrategy(req.RoutingStrategy)
	if strategy == "" {
		return errors.New("routing_strategy must be one of: priority, weighted")
	}

	if len(req.Targets) == 0 {
		return errors.New("at least one target is required")
	}

	return validateVirtualModelTargets(req.Targets, strategy)
}

func validateUpdateVirtualModelRequest(req platform.UpdateVirtualModelRequest, currentStrategy string) error {
	strategy := normalizeVirtualModelRoutingStrategy(currentStrategy)
	if strategy == "" {
		strategy = "priority"
	}

	if req.RoutingStrategy != nil {
		strategy = normalizeVirtualModelRoutingStrategy(*req.RoutingStrategy)
		if strategy == "" {
			return errors.New("routing_strategy must be one of: priority, weighted")
		}
	}

	if len(req.Targets) > 0 {
		return validateVirtualModelTargets(req.Targets, strategy)
	}

	return nil
}

func normalizeVirtualModelRoutingStrategy(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "priority":
		return "priority"
	case "weighted":
		return "weighted"
	default:
		return ""
	}
}

func validateVirtualModelTargets(targets []platform.VirtualModelTarget, strategy string) error {
	seenOrders := make(map[int]struct{}, len(targets))
	totalWeight := 0.0

	for i := range targets {
		t := targets[i]
		if strings.TrimSpace(t.Provider) == "" {
			return errors.New("target.provider is required")
		}

		if strings.TrimSpace(t.Model) == "" {
			return errors.New("target.model is required")
		}

		if t.Order <= 0 {
			return errors.New("target.order must be greater than 0")
		}

		if _, exists := seenOrders[t.Order]; exists {
			return errors.New("target.order must be unique")
		}

		seenOrders[t.Order] = struct{}{}
		if strategy == "weighted" {
			if t.Weight <= 0 {
				return errors.New("target.weight must be greater than 0 for weighted routing")
			}

			totalWeight += t.Weight
		}

		if t.Policy.Retry.Attempts < 0 || t.Policy.Retry.Attempts > 5 {
			return errors.New("target.retry.attempts must be between 0 and 5")
		}

		if t.Policy.Retry.DelayMS < 0 || t.Policy.Retry.DelayMS > 10000 {
			return errors.New("target.retry.delay_ms must be between 0 and 10000")
		}
	}

	if strategy == "weighted" && math.Abs(totalWeight-100) > 0.000001 {
		return errors.New("target.weight values must sum to 100 for weighted routing")
	}

	return nil
}

func (h *Handlers) ListRateLimitRules(c *fiber.Ctx) error {
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

	items, total, err := h.policies.ListRateLimitRules(c.Context(), currentTenant.ID, params.Query, params.Limit, params.Offset)
	if err != nil {
		return err
	}

	return c.JSON(platform.RateLimitRuleListResponse{
		Data:       items,
		Total:      total,
		Limit:      params.Limit,
		Offset:     params.Offset,
		NextOffset: buildNextOffset(params.Offset, len(items), total),
	})
}

func (h *Handlers) CreateRateLimitRule(c *fiber.Ctx) error {
	if err := requirePolicyManage(c); err != nil {
		return err
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}

	if err := h.requireEntitlementEnabled(c, currentTenant.ID, platform.FeatureRateLimitRulesEnabled); err != nil {
		return err
	}

	req, err := parseCreateRateLimitRuleRequest(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	validateErr := validateCreateRateLimitRuleRequest(req)
	if validateErr != nil {
		return fiber.NewError(fiber.StatusBadRequest, validateErr.Error())
	}

	item, err := h.policies.CreateRateLimitRule(c.Context(), currentTenant.ID, req)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return fiber.NewError(fiber.StatusUnprocessableEntity, "ثبت قانون انجام نشد. لطفا فیلدهای قانون و شرایط را دوباره بررسی کنید.")
		}

		return mapProviderAccountStoreError(err)
	}

	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *Handlers) GetRateLimitRule(c *fiber.Ctx) error {
	if err := requirePolicyListOrManage(c); err != nil {
		return err
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}

	ruleID := strings.TrimSpace(c.Params("id"))
	if ruleID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "id is required")
	}

	item, err := h.policies.GetRateLimitRuleByRuleID(c.Context(), currentTenant.ID, ruleID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "rate limit rule not found")
		}

		return err
	}

	return c.JSON(item)
}

func (h *Handlers) UpdateRateLimitRule(c *fiber.Ctx) error {
	if err := requirePolicyManage(c); err != nil {
		return err
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}

	ruleID := strings.TrimSpace(c.Params("id"))
	if ruleID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "id is required")
	}

	req, err := parseUpdateRateLimitRuleRequest(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	validateErr := validateUpdateRateLimitRuleRequest(req)
	if validateErr != nil {
		return fiber.NewError(fiber.StatusBadRequest, validateErr.Error())
	}

	item, err := h.policies.UpdateRateLimitRuleByRuleID(c.Context(), currentTenant.ID, ruleID, req)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "rate limit rule not found")
		}

		return mapProviderAccountStoreError(err)
	}

	return c.JSON(item)
}

func (h *Handlers) DeleteRateLimitRule(c *fiber.Ctx) error {
	if err := requirePolicyManage(c); err != nil {
		return err
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}

	ruleID := strings.TrimSpace(c.Params("id"))
	if ruleID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "id is required")
	}

	if err := h.policies.DeleteRateLimitRuleByRuleID(c.Context(), currentTenant.ID, ruleID); err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "rate limit rule not found")
		}

		return err
	}

	return c.JSON(fiber.Map{"success": true})
}

func (h *Handlers) UpdateRateLimitRuleStatus(c *fiber.Ctx) error {
	if err := requirePolicyManage(c); err != nil {
		return err
	}

	return handleScopedEnabledUpdate(c, "rate limit rule not found", func(ctx context.Context, tenantID, itemID string, enabled bool) (interface{}, error) {
		return h.policies.UpdateRateLimitRuleStatusByRuleID(ctx, tenantID, itemID, enabled)
	})
}

func validateCreateRateLimitRuleRequest(req platform.CreateRateLimitRuleRequest) error {
	if strings.TrimSpace(req.RuleID) == "" {
		return errors.New("rule_id is required")
	}

	if len(req.Limit) == 0 {
		return errors.New("limit is required")
	}

	return nil
}

func validateUpdateRateLimitRuleRequest(req platform.UpdateRateLimitRuleRequest) error {
	if req.Priority != nil && *req.Priority < 0 {
		return errors.New("priority cannot be negative")
	}

	return nil
}

func parseCreateRateLimitRuleRequest(c *fiber.Ctx) (platform.CreateRateLimitRuleRequest, error) {
	var req platform.CreateRateLimitRuleRequest
	if err := parseTemporalBody(c, &req, func(effectiveFrom, expiresAt *time.Time) {
		req.EffectiveFrom = effectiveFrom
		req.ExpiresAt = expiresAt
	}); err != nil {
		return platform.CreateRateLimitRuleRequest{}, err
	}

	return req, nil
}

func parseUpdateRateLimitRuleRequest(c *fiber.Ctx) (platform.UpdateRateLimitRuleRequest, error) {
	var req platform.UpdateRateLimitRuleRequest
	if err := parseTemporalBody(c, &req, func(effectiveFrom, expiresAt *time.Time) {
		req.EffectiveFrom = effectiveFrom
		req.ExpiresAt = expiresAt
	}); err != nil {
		return platform.UpdateRateLimitRuleRequest{}, err
	}

	return req, nil
}

func (h *Handlers) ListBudgetLimitRules(c *fiber.Ctx) error {
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

	items, total, err := h.policies.ListBudgetLimitRules(c.Context(), currentTenant.ID, params.Query, params.Limit, params.Offset)
	if err != nil {
		return err
	}

	return c.JSON(platform.BudgetLimitRuleListResponse{
		Data:       items,
		Total:      total,
		Limit:      params.Limit,
		Offset:     params.Offset,
		NextOffset: buildNextOffset(params.Offset, len(items), total),
	})
}

func (h *Handlers) CreateBudgetLimitRule(c *fiber.Ctx) error {
	if err := requirePolicyManage(c); err != nil {
		return err
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}

	if err := h.requireEntitlementEnabled(c, currentTenant.ID, platform.FeatureBudgetLimitRulesEnabled); err != nil {
		return err
	}

	var req platform.CreateBudgetLimitRuleRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	if err := validateCreateBudgetLimitRuleRequest(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	item, err := h.policies.CreateBudgetLimitRule(c.Context(), currentTenant.ID, req)
	if err != nil {
		if isBudgetRuleCreditError(err) {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}

		if errors.Is(err, apperrors.ErrNotFound) {
			return fiber.NewError(fiber.StatusUnprocessableEntity, "failed to create budget limit rule")
		}

		return mapProviderAccountStoreError(err)
	}

	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *Handlers) GetBudgetLimitRule(c *fiber.Ctx) error {
	if err := requirePolicyListOrManage(c); err != nil {
		return err
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}

	ruleID := strings.TrimSpace(c.Params("id"))
	if ruleID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "id is required")
	}

	item, err := h.policies.GetBudgetLimitRuleByRuleID(c.Context(), currentTenant.ID, ruleID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "budget limit rule not found")
		}

		return err
	}

	return c.JSON(item)
}

func (h *Handlers) UpdateBudgetLimitRule(c *fiber.Ctx) error {
	if err := requirePolicyManage(c); err != nil {
		return err
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}

	ruleID := strings.TrimSpace(c.Params("id"))
	if ruleID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "id is required")
	}

	var req platform.UpdateBudgetLimitRuleRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	if err := validateUpdateBudgetLimitRuleRequest(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	item, err := h.policies.UpdateBudgetLimitRuleByRuleID(c.Context(), currentTenant.ID, ruleID, req)
	if err != nil {
		if isBudgetRuleCreditError(err) {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}

		if errors.Is(err, apperrors.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "budget limit rule not found")
		}

		return mapProviderAccountStoreError(err)
	}

	return c.JSON(item)
}

func (h *Handlers) DeleteBudgetLimitRule(c *fiber.Ctx) error {
	if err := requirePolicyManage(c); err != nil {
		return err
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}

	ruleID := strings.TrimSpace(c.Params("id"))
	if ruleID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "id is required")
	}

	if err := h.policies.DeleteBudgetLimitRuleByRuleID(c.Context(), currentTenant.ID, ruleID); err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "budget limit rule not found")
		}

		return err
	}

	return c.JSON(fiber.Map{"success": true})
}

func (h *Handlers) UpdateBudgetLimitRuleStatus(c *fiber.Ctx) error {
	if err := requirePolicyManage(c); err != nil {
		return err
	}

	return handleScopedEnabledUpdate(c, "budget limit rule not found", func(ctx context.Context, tenantID, itemID string, enabled bool) (interface{}, error) {
		return h.policies.UpdateBudgetLimitRuleStatusByRuleID(ctx, tenantID, itemID, enabled)
	})
}

func validateCreateBudgetLimitRuleRequest(req platform.CreateBudgetLimitRuleRequest) error {
	if strings.TrimSpace(req.RuleID) == "" {
		return errors.New("rule_id is required")
	}

	if req.LimitTo < 0 || (req.LimitTo == 0 && req.Limit == nil) {
		return errors.New("limit or limit_to must be greater than zero")
	}

	if err := validateBudgetRuleCreditFields(req.Limit, req.OnExceed, req.BlockOnBudgetExceed, req.AlertThresholds); err != nil {
		return err
	}

	switch req.Unit {
	case platform.BudgetLimitCostPerDay, platform.BudgetLimitCostPerWeek, platform.BudgetLimitCostPerMonth:
	default:
		return errors.New("unit is invalid")
	}

	if req.Priority != nil && *req.Priority < 0 {
		return errors.New("priority cannot be negative")
	}

	if req.BudgetAppliesPer != nil && strings.TrimSpace(*req.BudgetAppliesPer) != "" {
		if err := validateBudgetAppliesPer(*req.BudgetAppliesPer); err != nil {
			return err
		}
	}

	return nil
}

func validateUpdateBudgetLimitRuleRequest(req platform.UpdateBudgetLimitRuleRequest) error {
	if req.Priority != nil && *req.Priority < 0 {
		return errors.New("priority cannot be negative")
	}

	if req.LimitTo != nil && *req.LimitTo <= 0 {
		return errors.New("limit_to must be greater than zero")
	}

	if err := validateBudgetRuleCreditFields(req.Limit, req.OnExceed, req.BlockOnBudgetExceed, req.AlertThresholds); err != nil {
		return err
	}

	if req.Unit != nil {
		switch *req.Unit {
		case platform.BudgetLimitCostPerDay, platform.BudgetLimitCostPerWeek, platform.BudgetLimitCostPerMonth:
		default:
			return errors.New("unit is invalid")
		}
	}

	if req.BudgetAppliesPer != nil && strings.TrimSpace(*req.BudgetAppliesPer) != "" {
		if err := validateBudgetAppliesPer(*req.BudgetAppliesPer); err != nil {
			return err
		}
	}

	return nil
}

// validateBudgetRuleCreditFields checks the credit fields a request sets. An
// empty limit string is allowed: it clears the limit on update.
func validateBudgetRuleCreditFields(limit *string, onExceed *platform.BudgetOnExceed, block *bool, thresholds []int) error {
	if limit != nil && strings.TrimSpace(*limit) != "" {
		if _, err := credits.ParseCreditAmountMC(*limit); err != nil {
			return err
		}
	}

	if onExceed != nil {
		if _, err := credits.ResolveOnExceed(onExceed, block, ""); err != nil {
			return err
		}
	}

	if len(thresholds) > 0 {
		if _, err := credits.NormalizeAlertThresholds(thresholds); err != nil {
			return err
		}
	}

	return nil
}

func isBudgetRuleCreditError(err error) bool {
	return errors.Is(err, credits.ErrInvalidCreditAmount) ||
		errors.Is(err, credits.ErrInvalidOnExceed) ||
		errors.Is(err, credits.ErrInconsistentOnExceed) ||
		errors.Is(err, credits.ErrInvalidAlertThreshold)
}

func validateBudgetAppliesPer(value string) error {
	normalized := strings.TrimSpace(value)
	switch normalized {
	case string(platform.BudgetAppliesPerTenant),
		string(platform.BudgetAppliesPerUser),
		string(platform.BudgetAppliesPerTeam),
		string(platform.BudgetAppliesPerOrgUnit),
		string(platform.BudgetAppliesPerModel),
		string(platform.BudgetAppliesPerVirtualAccount):
		return nil
	}

	if strings.HasPrefix(normalized, "metadata.") && strings.TrimPrefix(normalized, "metadata.") != "" {
		return nil
	}

	return errors.New("budget_applies_per is invalid")
}

func parseOptionalTimeValue(value interface{}) (*time.Time, error) {
	if value == nil {
		return nil, nil
	}

	switch v := value.(type) {
	case string:
		trimmed := strings.TrimSpace(v)
		if trimmed == "" {
			return nil, nil
		}

		parsed, err := time.Parse(time.RFC3339, trimmed)
		if err != nil {
			return nil, err
		}

		return &parsed, nil
	default:
		return nil, nil
	}
}

// GetWallet returns the current tenant wallet.
