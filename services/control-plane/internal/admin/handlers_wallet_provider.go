package admin

import (
	"errors"
	"strconv"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

func (h *Handlers) GetWallet(c *fiber.Ctx) error {
	if err := requireTenantAdminOrHigher(c); err != nil {
		return err
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	wallet, err := h.wallets.GetWallet(c.Context(), currentTenant.ID)
	if err != nil {
		return err
	}
	return c.JSON(wallet)
}

type topUpRequest struct {
	AmountIRR float64                `json:"amount_irr"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// TopUpWallet performs a stubbed immediate top-up.
func (h *Handlers) TopUpWallet(c *fiber.Ctx) error {
	if err := requireTenantAdminOrHigher(c); err != nil {
		return err
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}

	var req topUpRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	txn, err := h.wallets.TopUpWallet(c.Context(), currentTenant.ID, req.AmountIRR, req.Metadata)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(txn)
}

// ListProviderCatalog returns providers available for tenant model onboarding.
func (h *Handlers) ListProviderCatalog(c *fiber.Ctx) error {
	if err := requirePermissionAnyOrTenantAdmin(c, "provider account read permission required", []string{
		"provider_account:read",
		"provider_account:manage",
		"provider_account:use_integrations",
	}); err != nil {
		return err
	}

	items, err := discoveryProviders(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, err.Error())
	}
	return c.JSON(platform.ProviderCatalogListResponse{Data: items})
}

// RefreshProviderCatalogCache invalidates in-memory discovery cache.
func (h *Handlers) RefreshProviderCatalogCache(c *fiber.Ctx) error {
	if err := requirePermissionAnyOrTenantAdmin(c, "provider account manage permission required", []string{
		"provider_account:manage",
		"settings:manage",
	}); err != nil {
		return err
	}
	invalidateDiscoveryCache()
	return c.JSON(fiber.Map{
		"ok":      true,
		"message": "provider catalog cache invalidated",
	})
}

// GetProviderCatalogStats returns discovery/cache telemetry counters.
func (h *Handlers) GetProviderCatalogStats(c *fiber.Ctx) error {
	if err := requirePermissionAnyOrTenantAdmin(c, "provider account read permission required", []string{
		"provider_account:read",
		"provider_account:manage",
		"provider_account:use_integrations",
	}); err != nil {
		return err
	}
	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"cache_ttl_sec": int64(discoveryCacheTTL().Seconds()),
			"counters":      discoveryTelemetrySnapshot(),
		},
	})
}

// ListProviderModelsCatalog returns provider model catalog for step-2 selection.
func (h *Handlers) ListProviderModelsCatalog(c *fiber.Ctx) error {
	if err := requirePermissionAnyOrTenantAdmin(c, "provider account read permission required", []string{
		"provider_account:read",
		"provider_account:manage",
		"provider_account:use_integrations",
	}); err != nil {
		return err
	}

	provider := strings.TrimSpace(strings.ToLower(c.Params("provider")))
	if provider == "" {
		return fiber.NewError(fiber.StatusBadRequest, "provider is required")
	}
	items, err := discoveryProviderModels(c.Context(), provider)
	if err != nil {
		if errors.Is(err, errDiscoveryNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "provider model catalog not found")
		}
		return fiber.NewError(fiber.StatusBadGateway, err.Error())
	}
	return c.JSON(platform.ProviderModelCatalogListResponse{Data: items})
}

// CreateTenantProviderAccount creates a tenant-scoped provider account definition.
func (h *Handlers) CreateTenantProviderAccount(c *fiber.Ctx) error {
	if err := requirePermissionAnyOrTenantAdmin(c, "provider account create permission required", []string{
		"provider_account:create",
		"provider_account:manage",
	}); err != nil {
		return err
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	store, err := h.requireProviderAccountsStore()
	if err != nil {
		return err
	}

	var req platform.CreateTenantProviderAccountRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	req.Provider = strings.TrimSpace(strings.ToLower(req.Provider))
	req.Name = strings.TrimSpace(req.Name)

	if req.Provider == "" || req.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "provider and name are required")
	}
	if req.AuthMode == "" {
		return fiber.NewError(fiber.StatusBadRequest, "auth_mode is required")
	}
	exists, err := discoveryProviderExists(c.Context(), req.Provider)
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, err.Error())
	}
	if !exists {
		return fiber.NewError(fiber.StatusBadRequest, "unsupported provider from discovery service")
	}

	account, err := store.CreateTenantProviderAccount(c.Context(), currentTenant.ID, req)
	if err != nil {
		return mapProviderAccountStoreError(err)
	}
	actor := strings.TrimSpace(c.Get("x-shahrah-user-id"))
	if actor == "" {
		actor = "unknown"
	}
	c.Context().SetUserValue("audit.provider_account_event", map[string]interface{}{
		"event":      "provider_account_created",
		"tenant_id":  currentTenant.ID,
		"account_id": account.ID,
		"provider":   account.Provider,
		"actor":      actor,
	})
	return c.Status(fiber.StatusCreated).JSON(account)
}

// ListTenantProviderAccounts returns all provider accounts configured for the tenant.
func (h *Handlers) ListTenantProviderAccounts(c *fiber.Ctx) error {
	if err := requirePermissionAnyOrTenantAdmin(c, "provider account read permission required", []string{
		"provider_account:read",
		"provider_account:manage",
		"provider_account:use_integrations",
	}); err != nil {
		return err
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	store, err := h.requireProviderAccountsStore()
	if err != nil {
		return err
	}

	items, err := store.ListTenantProviderAccounts(c.Context(), currentTenant.ID)
	if err != nil {
		return err
	}
	return c.JSON(platform.TenantProviderAccountListResponse{Data: items})
}

// GetTenantProviderAccount returns full account details for wizard editing.
func (h *Handlers) GetTenantProviderAccount(c *fiber.Ctx) error {
	if err := requirePermissionAnyOrTenantAdmin(c, "provider account read permission required", []string{
		"provider_account:read",
		"provider_account:manage",
		"provider_account:use_integrations",
	}); err != nil {
		return err
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	store, err := h.requireProviderAccountsStore()
	if err != nil {
		return err
	}

	accountID := strings.TrimSpace(c.Params("accountID"))
	if accountID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "accountID is required")
	}

	details, err := store.GetTenantProviderAccountDetails(c.Context(), currentTenant.ID, accountID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "provider account not found")
		}
		return err
	}
	return c.JSON(details)
}

// DeleteTenantProviderAccount deletes one provider account for the tenant.
func (h *Handlers) DeleteTenantProviderAccount(c *fiber.Ctx) error {
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	store, err := h.requireProviderAccountsStore()
	if err != nil {
		return err
	}

	accountID := strings.TrimSpace(c.Params("accountID"))
	if accountID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "accountID is required")
	}

	details, err := store.GetTenantProviderAccountDetails(c.Context(), currentTenant.ID, accountID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "provider account not found")
		}
		return mapProviderAccountStoreError(err)
	}
	if details == nil {
		return fiber.NewError(fiber.StatusNotFound, "provider account not found")
	}

	if err := h.requireTenantAdminOrProviderAccountManager(c, store, currentTenant.ID, accountID); err != nil {
		return err
	}

	if err := store.DeleteTenantProviderAccount(c.Context(), currentTenant.ID, accountID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "provider account not found")
		}
		return mapProviderAccountStoreError(err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// GetTenantProviderAccountEffectiveModels returns effective model access view for debugging/observability.
func (h *Handlers) GetTenantProviderAccountEffectiveModels(c *fiber.Ctx) error {
	if err := requirePermissionAnyOrTenantAdmin(c, "provider account read permission required", []string{
		"provider_account:read",
		"provider_account:manage",
		"provider_account:use_integrations",
	}); err != nil {
		return err
	}
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	store, err := h.requireProviderAccountsStore()
	if err != nil {
		return err
	}
	accountID := strings.TrimSpace(c.Params("accountID"))
	if accountID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "accountID is required")
	}
	keyType := strings.TrimSpace(strings.ToLower(c.Query("key_type")))
	if keyType == "" {
		keyType = platform.KeyTypePAT
	}
	if keyType != platform.KeyTypePAT && keyType != platform.KeyTypeVAT {
		return fiber.NewError(fiber.StatusBadRequest, "key_type must be pat or vat")
	}
	userID := strings.TrimSpace(c.Query("user_id"))
	teamID := strings.TrimSpace(c.Query("team_id"))

	items, err := store.ListTenantProviderAccountEffectiveModels(c.Context(), currentTenant.ID, accountID, keyType, userID, teamID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "provider account not found")
		}
		return err
	}
	return c.JSON(platform.EffectiveModelsResponse{Data: items})
}

// UpsertTenantProviderAccountModels replaces selected models for a provider account.
func (h *Handlers) UpsertTenantProviderAccountModels(c *fiber.Ctx) error {
	if err := requirePermissionAnyOrTenantAdmin(c, "provider account manage permission required", []string{
		"provider_account:manage",
	}); err != nil {
		return err
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	store, err := h.requireProviderAccountsStore()
	if err != nil {
		return err
	}

	accountID := strings.TrimSpace(c.Params("accountID"))
	if accountID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "accountID is required")
	}

	var req platform.UpsertTenantProviderAccountModelsRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if len(req.Models) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "at least one model must be selected")
	}
	for i := range req.Models {
		req.Models[i].Model = strings.TrimSpace(req.Models[i].Model)
		if req.Models[i].Model == "" {
			return fiber.NewError(fiber.StatusBadRequest, "model is required for each item")
		}
		normalizedTypes := make([]platform.ModelType, 0, len(req.Models[i].Types))
		for _, mt := range req.Models[i].Types {
			trimmed := platform.ModelType(strings.TrimSpace(strings.ToLower(string(mt))))
			if trimmed == "" {
				continue
			}
			normalizedTypes = append(normalizedTypes, trimmed)
		}
		req.Models[i].Types = normalizedTypes
		if req.Models[i].PricingMode == "" {
			req.Models[i].PricingMode = platform.PricingModePublic
		}
	}
	// Normalize duplicate model entries while preserving input order.
	seenModels := make(map[string]struct{}, len(req.Models))
	normalizedModels := make([]platform.TenantProviderAccountModel, 0, len(req.Models))
	for _, item := range req.Models {
		if _, ok := seenModels[item.Model]; ok {
			continue
		}
		seenModels[item.Model] = struct{}{}
		normalizedModels = append(normalizedModels, item)
	}

	details, err := store.GetTenantProviderAccountDetails(c.Context(), currentTenant.ID, accountID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "provider account not found")
		}
		return mapProviderAccountStoreError(err)
	}
	invalidModels, invalidTypeModels, err := validateProviderModelsFromCatalog(c.Context(), details.Account.Provider, normalizedModels)
	if err != nil {
		if errors.Is(err, errDiscoveryNotFound) {
			return fiber.NewError(fiber.StatusBadRequest, "provider model catalog not found")
		}
		return fiber.NewError(fiber.StatusBadGateway, err.Error())
	}
	if len(invalidModels) > 0 || len(invalidTypeModels) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"type":    "validation_error",
				"message": "one or more models are not valid for this provider account or requested capabilities",
				"details": fiber.Map{
					"invalid_models":      invalidModels,
					"invalid_model_types": invalidTypeModels,
					"provider":            details.Account.Provider,
				},
			},
		})
	}

	for _, item := range normalizedModels {
		if item.PricingMode == platform.PricingModePrivate && (item.PrivatePricing == nil || (item.PrivatePricing.InputUSDPer1M == nil && item.PrivatePricing.OutputUSDPer1M == nil)) {
			return fiber.NewError(fiber.StatusBadRequest, "private pricing mode requires private_pricing input/output values")
		}
	}

	items, err := store.UpsertTenantProviderAccountModels(c.Context(), currentTenant.ID, accountID, normalizedModels)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "provider account not found")
		}
		return mapProviderAccountStoreError(err)
	}
	out := make([]platform.TenantProviderAccountModelEffective, 0, len(items))
	for _, item := range items {
		resolved := item.PricingMode == platform.PricingModePublic || item.PrivatePricing != nil
		out = append(out, platform.TenantProviderAccountModelEffective{
			Model:                 item.Model,
			Enabled:               item.Enabled,
			EffectiveCapabilities: item.Types,
			PricingMode:           item.PricingMode,
			PricingResolved:       resolved,
		})
	}
	return c.JSON(platform.TenantProviderAccountModelsUpsertResponse{Data: out})
}

// UpdateTenantProviderAccountAccess replaces access control entries for one provider account.
func (h *Handlers) UpdateTenantProviderAccountAccess(c *fiber.Ctx) error {
	if err := requirePermissionAnyOrTenantAdmin(c, "provider account manage permission required", []string{
		"provider_account:manage",
	}); err != nil {
		return err
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	store, err := h.requireProviderAccountsStore()
	if err != nil {
		return err
	}

	accountID := strings.TrimSpace(c.Params("accountID"))
	if accountID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "accountID is required")
	}

	var req platform.UpdateTenantProviderAccountAccessRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if len(req.Access.Managers) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "at least one manager is required")
	}
	for _, item := range req.Access.Managers {
		if item.Type == platform.PrincipalTypeEveryone {
			return fiber.NewError(fiber.StatusBadRequest, "manager role only supports user/team principals")
		}
	}
	for _, rule := range req.Access.Rules {
		if rule.Role == platform.AccessRoleManager && rule.Principal.Type == platform.PrincipalTypeEveryone {
			return fiber.NewError(fiber.StatusBadRequest, "manager role only supports user/team principals")
		}
		if rule.AccessScope == platform.AccessScopeModel && strings.TrimSpace(rule.Model) == "" {
			return fiber.NewError(fiber.StatusBadRequest, "model is required when access_scope=model")
		}
	}

	access, err := store.UpdateTenantProviderAccountAccess(c.Context(), currentTenant.ID, accountID, req.Access)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "provider account not found")
		}
		return mapProviderAccountStoreError(err)
	}
	return c.JSON(fiber.Map{"access": access})
}

// TestTenantProviderAccountCredentials validates configured credentials (or virtual-key-only mode).
func (h *Handlers) TestTenantProviderAccountCredentials(c *fiber.Ctx) error {
	if err := requirePermissionAnyOrTenantAdmin(c, "provider account manage permission required", []string{
		"provider_account:manage",
	}); err != nil {
		return err
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	store, err := h.requireProviderAccountsStore()
	if err != nil {
		return err
	}

	accountID := strings.TrimSpace(c.Params("accountID"))
	if accountID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "accountID is required")
	}

	var req platform.TestTenantProviderAccountCredentialsRequest
	if len(c.Body()) > 0 {
		if err := c.BodyParser(&req); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
		}
	}

	result, err := store.TestTenantProviderAccountCredentials(c.Context(), currentTenant.ID, accountID, req)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "provider account not found")
		}
		return err
	}
	return c.JSON(result)
}

// UpdateTenantProviderAccountCredentials updates auth mode / credential settings for a provider account.
func (h *Handlers) UpdateTenantProviderAccountCredentials(c *fiber.Ctx) error {
	if err := requirePermissionAnyOrTenantAdmin(c, "provider account manage permission required", []string{
		"provider_account:manage",
	}); err != nil {
		return err
	}
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	store, err := h.requireProviderAccountsStore()
	if err != nil {
		return err
	}

	accountID := strings.TrimSpace(c.Params("accountID"))
	if accountID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "accountID is required")
	}

	var req platform.UpdateTenantProviderAccountCredentialsRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if req.AuthMode != "" {
		req.AuthMode = platform.ProviderAuthMode(strings.TrimSpace(strings.ToLower(string(req.AuthMode))))
	}
	if req.Credentials != nil {
		req.Credentials.APIKey = strings.TrimSpace(req.Credentials.APIKey)
		req.Credentials.SecretRef = strings.TrimSpace(req.Credentials.SecretRef)
	}

	account, err := store.UpdateTenantProviderAccountCredentials(c.Context(), currentTenant.ID, accountID, req)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "provider account not found")
		}
		return mapProviderAccountStoreError(err)
	}
	return c.JSON(account)
}

// ListAccessPrincipals returns user/team/org-unit principals for access picker controls.
func (h *Handlers) ListAccessPrincipals(c *fiber.Ctx) error {
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	store, err := h.requireProviderAccountsStore()
	if err != nil {
		return err
	}

	kind := strings.TrimSpace(strings.ToLower(c.Query("kind")))
	switch kind {
	case "user", "team", "org_unit":
	default:
		return fiber.NewError(fiber.StatusBadRequest, "kind must be user, team or org_unit")
	}
	if err := requireAccessPrincipalSearchPermission(c, kind); err != nil {
		return err
	}
	query := strings.TrimSpace(c.Query("q"))
	limit := 20
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		parsed, parseErr := strconv.Atoi(raw)
		if parseErr != nil || parsed <= 0 {
			return fiber.NewError(fiber.StatusBadRequest, "limit must be a positive number")
		}
		if parsed > 100 {
			parsed = 100
		}
		limit = parsed
	}

	items, err := store.SearchAccessPrincipals(c.Context(), currentTenant.ID, kind, query, limit)
	if err != nil {
		return err
	}
	return c.JSON(platform.AccessPrincipalListResponse{Data: items})
}
