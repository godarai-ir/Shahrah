package admin

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/modelid"
	"github.com/Delpak/shahrah/services/control-plane/internal/platformcredentials"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

const providerCatalogAssetCacheControl = "public, max-age=31536000, immutable"

// ListProviderCatalog returns providers available for tenant model onboarding.
func (h *Handlers) ListProviderCatalog(c *fiber.Ctx) error {
	if err := requirePermissionAnyOrTenantAdmin(c, "provider account read permission required", []string{
		"provider_account:read",
		"provider_account:manage",
		"provider_account:create",
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
		"provider_account:create",
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
		"provider_account:create",
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

// GetProviderCatalogDisplaySettings returns model-catalog display settings through control-plane.
func (h *Handlers) GetProviderCatalogDisplaySettings(c *fiber.Ctx) error {
	if err := requirePermissionAnyOrTenantAdmin(c, "provider account read permission required", []string{
		"provider_account:read",
		"provider_account:manage",
		"provider_account:create",
	}); err != nil {
		return err
	}

	settings, err := discoveryDisplaySettings(c.Context())
	if err != nil {
		if errors.Is(err, errDiscoveryNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "provider catalog display settings not found")
		}

		return fiber.NewError(fiber.StatusBadGateway, err.Error())
	}

	return c.JSON(platform.ProviderCatalogDisplaySettingsResponse{Data: *settings})
}

// ListProviderModelsCatalog returns provider model catalog for step-2 selection.
func (h *Handlers) ListProviderModelsCatalog(c *fiber.Ctx) error {
	if err := requirePermissionAnyOrTenantAdmin(c, "provider account read permission required", []string{
		"provider_account:read",
		"provider_account:manage",
		"provider_account:create",
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

// ListTenantProviderConnectionSchemas returns the documented provider-specific
// connection and write-only credential fields used by tenant onboarding.
func (h *Handlers) ListTenantProviderConnectionSchemas(c *fiber.Ctx) error {
	if err := requirePermissionAnyOrTenantAdmin(c, "provider account read permission required", []string{
		"provider_account:read",
		"provider_account:manage",
		"provider_account:create",
	}); err != nil {
		return err
	}

	provider := normalizeProviderSlug(c.Params("provider"))
	if provider == "" {
		return fiber.NewError(fiber.StatusBadRequest, "provider is required")
	}

	if h.providerSchemas == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "provider connection schema catalog is unavailable")
	}

	items, err := h.providerSchemas.Schemas(c.Context(), provider)
	if err != nil {
		if errors.Is(err, platformcredentials.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "provider connection schema not found")
		}

		return fiber.NewError(fiber.StatusBadGateway, "provider connection schema is unavailable")
	}

	return c.JSON(fiber.Map{"data": items})
}

// GetProviderCatalogAsset proxies provider logo assets from model-discovery.
func (h *Handlers) GetProviderCatalogAsset(c *fiber.Ctx) error {
	if err := requirePermissionAnyOrTenantAdmin(c, "provider account read permission required", []string{
		"provider_account:read",
		"provider_account:manage",
		"provider_account:create",
	}); err != nil {
		return err
	}

	bucket := strings.TrimSpace(c.Params("bucket"))

	fileName := strings.TrimSpace(c.Params("fileName"))
	if bucket == "" || fileName == "" {
		return fiber.NewError(fiber.StatusBadRequest, "bucket and fileName are required")
	}

	if bucket != "provider-logos" && bucket != "model-series-logos" {
		return fiber.NewError(fiber.StatusBadRequest, "unsupported provider catalog asset bucket")
	}

	if strings.Contains(fileName, "..") {
		return fiber.NewError(fiber.StatusBadRequest, "invalid provider catalog asset name")
	}

	baseURL := getDiscoveryBaseURL()
	if baseURL == "" {
		return fiber.NewError(fiber.StatusBadGateway, "model-discovery service unavailable")
	}

	rawURL := baseURL + "/providers/assets/" + url.PathEscape(bucket) + "/" + url.PathEscape(fileName)

	body, status, contentType, err := discoveryFetch(c.Context(), rawURL)
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, err.Error())
	}

	if status == fiber.StatusNotFound {
		return fiber.NewError(fiber.StatusNotFound, "provider catalog asset not found")
	}

	if status >= fiber.StatusBadRequest {
		return fiber.NewError(fiber.StatusBadGateway, "failed to load provider catalog asset")
	}

	if strings.TrimSpace(contentType) == "" {
		contentType = http.DetectContentType(body)
	}

	c.Set(fiber.HeaderContentType, contentType)
	c.Set(fiber.HeaderCacheControl, providerCatalogAssetCacheControl)

	return c.Status(status).Send(body)
}

// CreateTenantProviderAccount creates a tenant-scoped provider account definition.
func (h *Handlers) CreateTenantProviderAccount(c *fiber.Ctx) error {
	if err := requirePermissionAnyOrTenantAdmin(c, "provider account create permission required", []string{
		"provider_account:create",
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

	parseErr := c.BodyParser(&req)
	if parseErr != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	req.Provider = normalizeProviderSlug(req.Provider)
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

	actor := delegatedUserID(c)
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
		"provider_account:create",
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

	records, recordErr := store.ListTenantProviderAccountCatalogRecords(c.Context(), currentTenant.ID)
	if recordErr != nil {
		return recordErr
	}

	catalog := loadTenantProviderCatalogModels(c.Context(), records)
	summaries := buildTenantProviderAccountSummaries(records, catalog)

	return c.JSON(platform.TenantProviderAccountListResponse{Data: items, Summaries: summaries})
}

// ListTenantProviderAccountModels returns the tenant model catalog as a
// read-only projection over provider accounts and model-discovery metadata.
func (h *Handlers) ListTenantProviderAccountModels(c *fiber.Ctx) error {
	if err := requirePermissionAnyOrTenantAdmin(c, "provider account read permission required", []string{
		"provider_account:read",
		"provider_account:manage",
		"provider_account:create",
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

	records, err := store.ListTenantProviderAccountCatalogRecords(c.Context(), currentTenant.ID)
	if err != nil {
		return err
	}

	query, err := parseTenantProviderModelCatalogQuery(c)
	if err != nil {
		return err
	}

	catalog := loadTenantProviderCatalogModels(c.Context(), records)
	items := buildTenantProviderAccountModelCatalog(records, catalog, query)
	total := len(items)

	start := query.Offset
	if start > total {
		start = total
	}

	end := start + query.Limit
	if end > total {
		end = total
	}

	return c.JSON(platform.TenantProviderAccountModelCatalogResponse{
		Data:   items[start:end],
		Total:  total,
		Limit:  query.Limit,
		Offset: query.Offset,
	})
}

// TestProviderConnection validates write-only connection material without
// creating or updating a provider account.
func (h *Handlers) TestProviderConnection(c *fiber.Ctx) error {
	if err := requirePermissionAnyOrTenantAdmin(c, "provider account create permission required", []string{
		"provider_account:create",
		"provider_account:manage",
	}); err != nil {
		return err
	}

	var req platform.TestProviderConnectionRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid provider connection test payload")
	}

	req.Provider = normalizeProviderSlug(req.Provider)

	req.ModelHint = strings.TrimSpace(req.ModelHint)
	if req.Provider == "" {
		return fiber.NewError(fiber.StatusBadRequest, "provider is required")
	}

	if req.AuthMode == "" {
		return fiber.NewError(fiber.StatusBadRequest, "auth_mode is required")
	}

	providers, err := discoveryProviders(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, "provider catalog is unavailable")
	}

	var selected *platform.ProviderCatalogProvider

	for index := range providers {
		if normalizeProviderSlug(providers[index].Provider) == req.Provider {
			selected = &providers[index]
			break
		}
	}

	if selected == nil || !selected.Enabled {
		return fiber.NewError(fiber.StatusBadRequest, "provider is not available")
	}

	if !providerCatalogAllowsAuthMode(*selected, req.AuthMode) {
		return fiber.NewError(fiber.StatusBadRequest, "auth_mode is not supported by this provider")
	}

	store, err := h.requireProviderAccountsStore()
	if err != nil {
		return err
	}

	result, err := store.TestProviderConnection(c.Context(), req)
	if err != nil {
		return err
	}

	return c.JSON(result)
}

// GetTenantProviderAccount returns full account details for wizard editing.
func (h *Handlers) GetTenantProviderAccount(c *fiber.Ctx) error {
	if err := requirePermissionAnyOrTenantAdmin(c, "provider account read permission required", []string{
		"provider_account:read",
		"provider_account:manage",
		"provider_account:create",
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

	if details != nil {
		if enriched, enrichErr := enrichTenantProviderAccountModels(c.Context(), details.Account.Provider, details.Models); enrichErr == nil {
			details.Models = enriched
		}
	}

	return c.JSON(details)
}

// DeleteTenantProviderAccount deletes one provider account for the tenant.
func (h *Handlers) DeleteTenantProviderAccount(c *fiber.Ctx) error {
	if err := requirePermissionAnyOrTenantAdmin(c, "provider account delete permission required", []string{
		"provider_account:delete",
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

		return mapProviderAccountStoreError(err)
	}

	if details == nil {
		return fiber.NewError(fiber.StatusNotFound, "provider account not found")
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
		"provider_account:create",
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
//
//nolint:govet // Short-lived branch errors are checked immediately and intentionally scoped with short declarations.
func (h *Handlers) UpsertTenantProviderAccountModels(c *fiber.Ctx) error {
	if err := requirePermissionAnyOrTenantAdmin(c, "provider account manage permission required", []string{
		"provider_account:manage",
		"provider_account:create",
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

	parseErr := c.BodyParser(&req)
	if parseErr != nil {
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

		req.Models[i].EffectiveCapabilities = normalizeModelTypes(req.Models[i].EffectiveCapabilities)
		if len(req.Models[i].EffectiveCapabilities) == 0 {
			return fiber.NewError(fiber.StatusBadRequest, "effective_capabilities is required for each enabled model")
		}

		if req.Models[i].PricingMode == "" {
			req.Models[i].PricingMode = platform.PricingModePublic
		}
	}

	details, err := store.GetTenantProviderAccountDetails(c.Context(), currentTenant.ID, accountID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "provider account not found")
		}

		return mapProviderAccountStoreError(err)
	}

	normalizedModels, invalidModels, invalidCapabilities, err := resolveProviderAccountModelSelections(c.Context(), details.Account.Provider, req.Models)
	if err != nil {
		if errors.Is(err, errDiscoveryNotFound) {
			return fiber.NewError(fiber.StatusBadRequest, "provider model catalog not found")
		}

		return fiber.NewError(fiber.StatusBadGateway, err.Error())
	}

	if len(invalidModels) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"type":    "validation_error",
				"message": "one or more models are not valid for this provider account",
				"details": fiber.Map{
					"invalid_models": invalidModels,
					"provider":       details.Account.Provider,
				},
			},
		})
	}

	if len(invalidCapabilities) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"type":    "validation_error",
				"message": "one or more model capabilities are not valid for this provider account",
				"details": fiber.Map{
					"invalid_capabilities": invalidCapabilities,
					"provider":             details.Account.Provider,
				},
			},
		})
	}

	allowsPrivatePricing := tenantProviderAccountAllowsPrivatePricing(details.Account)
	for _, item := range normalizedModels {
		if !allowsPrivatePricing && (item.PricingMode == platform.PricingModePrivate || item.PrivatePricing != nil) {
			return fiber.NewError(fiber.StatusBadRequest, "private pricing requires a provider account with user-owned credentials")
		}

		if item.PricingMode == platform.PricingModePrivate && !hasAnyModelPricingRate(item.PrivatePricing) {
			return fiber.NewError(fiber.StatusBadRequest, "private pricing mode requires at least one valid private pricing rate")
		}

		if item.PricingMode == platform.PricingModePrivate {
			if err := validatePrivatePricingOverride(item.PublicPricing, item.PrivatePricing); err != nil {
				return fiber.NewError(fiber.StatusBadRequest, err.Error())
			}
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

func validatePrivatePricingOverride(publicPrice, privatePrice *platform.ModelPrice) error {
	if publicPrice == nil || privatePrice == nil {
		return errors.New("private pricing requires a current public pricing revision")
	}

	if privatePrice.InputUSDPer1M != nil && publicPrice.InputUSDPer1M == nil {
		return errors.New("private input pricing cannot override an unpublished public rate")
	}

	if privatePrice.OutputUSDPer1M != nil && publicPrice.OutputUSDPer1M == nil {
		return errors.New("private output pricing cannot override an unpublished public rate")
	}

	publicRates := make(map[string]platform.ModelPricingRate, len(publicPrice.Rates))
	for _, rate := range publicPrice.Rates {
		publicRates[rate.Key] = rate
	}

	seen := make(map[string]struct{}, len(privatePrice.Rates))
	for _, override := range privatePrice.Rates {
		key := strings.TrimSpace(override.Key)

		publicRate, ok := publicRates[key]
		if !ok {
			return errors.New("private pricing contains an unknown or expired rate key")
		}

		if _, duplicate := seen[key]; duplicate {
			return errors.New("private pricing contains a duplicate rate key")
		}

		seen[key] = struct{}{}

		if override.Dimension != publicRate.Dimension || override.Unit != publicRate.Unit {
			return errors.New("private pricing cannot change a rate dimension or unit")
		}

		amount, err := decimal.NewFromString(strings.TrimSpace(override.UnitPriceUSD))
		if err != nil || amount.IsNegative() {
			return errors.New("private pricing contains an invalid USD decimal amount")
		}
	}

	return nil
}

func tenantProviderAccountAllowsPrivatePricing(account platform.TenantProviderAccount) bool {
	return account.HasCredential &&
		!account.UsePlatformManagedCredentials &&
		account.AuthMode != platform.ProviderAuthModeNoCredential
}

func hasAnyModelPricingRate(price *platform.ModelPrice) bool {
	if price == nil {
		return false
	}

	if price.InputUSDPer1M != nil || price.OutputUSDPer1M != nil {
		return true
	}

	for _, rate := range price.Rates {
		if strings.TrimSpace(rate.Key) != "" &&
			strings.TrimSpace(rate.Dimension) != "" &&
			strings.TrimSpace(rate.Unit) != "" &&
			strings.TrimSpace(rate.UnitPriceUSD) != "" {
			return true
		}
	}

	return false
}

// UpdateTenantProviderAccountAccess replaces access control entries for one provider account.
func (h *Handlers) UpdateTenantProviderAccountAccess(c *fiber.Ctx) error {
	if err := requirePermissionAnyOrTenantAdmin(c, "provider account manage permission required", []string{
		"provider_account:manage",
		"provider_account:create",
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

	parseErr := c.BodyParser(&req)
	if parseErr != nil {
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

func enrichTenantProviderAccountModels(ctx context.Context, provider string, models []platform.TenantProviderAccountModel) ([]platform.TenantProviderAccountModel, error) {
	catalog, err := discoveryProviderModels(ctx, provider)
	if err != nil {
		return nil, err
	}

	index := make(map[string]platform.ProviderCatalogModel, len(catalog))
	for _, item := range catalog {
		key := strings.TrimSpace(item.Model)
		if key == "" {
			continue
		}

		index[key] = item
	}

	out := make([]platform.TenantProviderAccountModel, 0, len(models))
	for _, item := range models {
		next := item

		catalogModel, ok := index[strings.TrimSpace(item.Model)]
		if ok {
			if len(next.Types) == 0 {
				next.Types = normalizeModelTypes(catalogModel.Types)
			}

			if len(next.InputModalities) == 0 {
				next.InputModalities = normalizeStringList(catalogModel.InputModalities)
			}

			if len(next.OutputModalities) == 0 {
				next.OutputModalities = normalizeStringList(catalogModel.OutputModalities)
			}

			if next.PublicPricing == nil {
				next.PublicPricing = cloneModelPrice(catalogModel.PublicPricing)
			}
		}

		out = append(out, next)
	}

	return out, nil
}

type tenantProviderCatalogLookup map[string]map[string]platform.ProviderCatalogModel

type tenantProviderModelCatalogQuery struct {
	Query      string
	Providers  map[string]struct{}
	AccountIDs map[string]struct{}
	Sort       string
	Order      string
	Limit      int
	Offset     int
}

func parseTenantProviderModelCatalogQuery(c *fiber.Ctx) (tenantProviderModelCatalogQuery, error) {
	out := tenantProviderModelCatalogQuery{
		Query:      strings.ToLower(strings.TrimSpace(c.Query("q"))),
		Providers:  csvValueSet(c.Query("providers"), true),
		AccountIDs: csvValueSet(c.Query("account_ids"), false),
		Sort:       strings.ToLower(strings.TrimSpace(c.Query("sort"))),
		Order:      strings.ToLower(strings.TrimSpace(c.Query("order"))),
		Limit:      25,
	}
	if out.Sort == "" {
		out.Sort = "name"
	}

	switch out.Sort {
	case "name", "provider", "availability", "context", "input_price", "output_price":
	default:
		return out, fiber.NewError(fiber.StatusBadRequest, "unsupported model catalog sort")
	}

	if out.Order == "" {
		out.Order = "asc"
	}

	if out.Order != "asc" && out.Order != "desc" {
		return out, fiber.NewError(fiber.StatusBadRequest, "order must be asc or desc")
	}

	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value <= 0 {
			return out, fiber.NewError(fiber.StatusBadRequest, "limit must be a positive number")
		}

		if value > 100 {
			value = 100
		}

		out.Limit = value
	}

	if raw := strings.TrimSpace(c.Query("offset")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 0 {
			return out, fiber.NewError(fiber.StatusBadRequest, "offset must be zero or greater")
		}

		out.Offset = value
	}

	return out, nil
}

func csvValueSet(raw string, normalizeProvider bool) map[string]struct{} {
	out := make(map[string]struct{})

	for _, part := range strings.Split(raw, ",") {
		value := strings.TrimSpace(part)
		if normalizeProvider {
			value = normalizeProviderSlug(value)
		}

		if value != "" {
			out[value] = struct{}{}
		}
	}

	return out
}

func loadTenantProviderCatalogModels(ctx context.Context, records []platform.TenantProviderAccountCatalogRecord) tenantProviderCatalogLookup {
	providers := make(map[string]struct{})

	for _, record := range records {
		if strings.TrimSpace(record.Model) == "" {
			continue
		}

		provider := normalizeProviderSlug(record.Account.Provider)
		if provider != "" {
			providers[provider] = struct{}{}
		}
	}

	out := make(tenantProviderCatalogLookup, len(providers))
	for provider := range providers {
		models, err := discoveryProviderModels(ctx, provider)
		if err != nil {
			continue
		}

		index := make(map[string]platform.ProviderCatalogModel, len(models))
		for _, item := range models {
			key := strings.ToLower(modelid.CanonicalOfficialModel(provider, item.Model))
			if key != "" {
				index[key] = item
			}
		}

		out[provider] = index
	}

	return out
}

func catalogModelForRecord(record platform.TenantProviderAccountCatalogRecord, catalog tenantProviderCatalogLookup) (platform.ProviderCatalogModel, bool) {
	provider := normalizeProviderSlug(record.Account.Provider)
	model := strings.ToLower(modelid.CanonicalOfficialModel(provider, record.Model))
	item, ok := catalog[provider][model]

	return item, ok
}

func resolvedRecordPrice(record platform.TenantProviderAccountCatalogRecord, catalogModel platform.ProviderCatalogModel, hasCatalog bool) *platform.ModelPrice {
	if record.PricingMode == platform.PricingModePrivate {
		return cloneModelPrice(record.PrivatePricing)
	}

	if hasCatalog && catalogModel.PublicPricing != nil {
		return cloneModelPrice(catalogModel.PublicPricing)
	}

	return cloneModelPrice(record.PublicPricing)
}

func modelPriceResolved(price *platform.ModelPrice) bool {
	return price != nil && (price.InputUSDPer1M != nil || price.OutputUSDPer1M != nil)
}

func accountReadiness(account platform.TenantProviderAccount, summary platform.TenantProviderAccountSummary) platform.ProviderAccountReadiness {
	if !account.HasCredential &&
		!account.UsePlatformManagedCredentials &&
		account.AuthMode != platform.ProviderAuthModeNoCredential {
		return platform.ProviderAccountReadinessNeedsCredentials
	}

	if summary.ModelCount == 0 {
		return platform.ProviderAccountReadinessNeedsModels
	}

	if summary.ManagerCount == 0 {
		return platform.ProviderAccountReadinessNeedsAccess
	}

	if summary.PricingUnresolvedCount > 0 {
		return platform.ProviderAccountReadinessPricingIncomplete
	}

	return platform.ProviderAccountReadinessReady
}

func buildTenantProviderAccountSummaries(records []platform.TenantProviderAccountCatalogRecord, catalog tenantProviderCatalogLookup) map[string]platform.TenantProviderAccountSummary {
	out := make(map[string]platform.TenantProviderAccountSummary)
	accounts := make(map[string]platform.TenantProviderAccount)

	for _, record := range records {
		summary := out[record.Account.ID]
		summary.ManagerCount = record.ManagerCount
		summary.UserCount = record.UserCount

		accounts[record.Account.ID] = record.Account
		if strings.TrimSpace(record.Model) != "" {
			summary.ModelCount++

			catalogModel, ok := catalogModelForRecord(record, catalog)
			if !modelPriceResolved(resolvedRecordPrice(record, catalogModel, ok)) {
				summary.PricingUnresolvedCount++
			}
		}

		out[record.Account.ID] = summary
	}

	for accountID, summary := range out {
		summary.Readiness = accountReadiness(accounts[accountID], summary)
		out[accountID] = summary
	}

	return out
}

func buildTenantProviderAccountModelCatalog(
	records []platform.TenantProviderAccountCatalogRecord,
	catalog tenantProviderCatalogLookup,
	query tenantProviderModelCatalogQuery,
) []platform.TenantProviderAccountModelCatalogItem {
	summaries := buildTenantProviderAccountSummaries(records, catalog)
	byKey := make(map[string]*platform.TenantProviderAccountModelCatalogItem)

	for _, record := range records {
		if strings.TrimSpace(record.Model) == "" {
			continue
		}

		provider := normalizeProviderSlug(record.Account.Provider)
		if len(query.Providers) > 0 {
			if _, ok := query.Providers[provider]; !ok {
				continue
			}
		}

		if len(query.AccountIDs) > 0 {
			if _, ok := query.AccountIDs[record.Account.ID]; !ok {
				continue
			}
		}

		officialModel := modelid.CanonicalOfficialModel(provider, record.Model)
		key := provider + "\x00" + strings.ToLower(officialModel)
		item := byKey[key]

		catalogModel, hasCatalog := catalogModelForRecord(record, catalog)
		if item == nil {
			item = &platform.TenantProviderAccountModelCatalogItem{
				Provider:         provider,
				Model:            officialModel,
				DisplayName:      officialModel,
				Types:            append([]platform.ModelType(nil), record.Types...),
				InputModalities:  append([]string(nil), record.InputModalities...),
				OutputModalities: append([]string(nil), record.OutputModalities...),
				Availability:     make([]platform.TenantProviderAccountModelAvailability, 0, 1),
			}
			if hasCatalog {
				item.DisplayName = strings.TrimSpace(catalogModel.DisplayName)
				if item.DisplayName == "" {
					item.DisplayName = officialModel
				}

				item.Modes = append([]string(nil), catalogModel.Modes...)
				item.Types = append([]platform.ModelType(nil), catalogModel.Types...)
				item.ContextWindow = catalogModel.ContextWindow
				item.InputModalities = append([]string(nil), catalogModel.InputModalities...)
				item.OutputModalities = append([]string(nil), catalogModel.OutputModalities...)
				item.Capabilities = append([]string(nil), catalogModel.Capabilities...)
				item.Features = append([]string(nil), catalogModel.Features...)
				item.PublicPricing = cloneModelPrice(catalogModel.PublicPricing)
				item.IsDeprecated = catalogModel.IsDeprecated
				item.ReplacementModel = catalogModel.ReplacementModel
			}

			byKey[key] = item
		}

		effectivePrice := resolvedRecordPrice(record, catalogModel, hasCatalog)

		credentialSource := platform.CredentialSourceGodarManagedKey
		if record.Account.HasCredential &&
			!record.Account.UsePlatformManagedCredentials &&
			record.Account.AuthMode != platform.ProviderAuthModeNoCredential {
			credentialSource = platform.CredentialSourceUserBYOK
		}

		item.Availability = append(item.Availability, platform.TenantProviderAccountModelAvailability{
			AccountID:             record.Account.ID,
			AccountName:           record.Account.Name,
			FQN:                   modelid.BuildModelFQN(provider, record.Account.Name, officialModel),
			CredentialSource:      credentialSource,
			EffectiveCapabilities: append([]platform.ModelType(nil), record.Types...),
			PricingMode:           record.PricingMode,
			EffectivePricing:      effectivePrice,
			PricingResolved:       modelPriceResolved(effectivePrice),
			Readiness:             summaries[record.Account.ID].Readiness,
		})
	}

	items := make([]platform.TenantProviderAccountModelCatalogItem, 0, len(byKey))
	for _, item := range byKey {
		sort.SliceStable(item.Availability, func(i, j int) bool {
			return strings.ToLower(item.Availability[i].AccountName) < strings.ToLower(item.Availability[j].AccountName)
		})

		if query.Query != "" && !tenantCatalogItemMatches(*item, query.Query) {
			continue
		}

		items = append(items, *item)
	}

	sort.SliceStable(items, func(i, j int) bool {
		comparison := compareTenantCatalogItems(items[i], items[j], query.Sort)
		if comparison == 0 {
			comparison = strings.Compare(strings.ToLower(items[i].Provider+":"+items[i].Model), strings.ToLower(items[j].Provider+":"+items[j].Model))
		}

		if query.Order == "desc" {
			return comparison > 0
		}

		return comparison < 0
	})

	return items
}

func tenantCatalogItemMatches(item platform.TenantProviderAccountModelCatalogItem, query string) bool {
	values := make([]string, 0, 3+2*len(item.Availability))
	values = append(values, item.Provider, item.Model, item.DisplayName)

	for _, availability := range item.Availability {
		values = append(values, availability.AccountName, availability.FQN)
	}

	for _, value := range values {
		if strings.Contains(strings.ToLower(value), query) {
			return true
		}
	}

	return false
}

func compareTenantCatalogItems(left, right platform.TenantProviderAccountModelCatalogItem, field string) int {
	switch field {
	case "provider":
		return strings.Compare(strings.ToLower(left.Provider), strings.ToLower(right.Provider))
	case "availability":
		return compareInt(len(left.Availability), len(right.Availability))
	case "context":
		return compareInt(intValue(left.ContextWindow), intValue(right.ContextWindow))
	case "input_price":
		return compareFloat(catalogSortPrice(left, true), catalogSortPrice(right, true))
	case "output_price":
		return compareFloat(catalogSortPrice(left, false), catalogSortPrice(right, false))
	default:
		return strings.Compare(strings.ToLower(left.DisplayName), strings.ToLower(right.DisplayName))
	}
}

func compareInt(left, right int) int {
	switch {
	case left < right:
		return -1
	case left > right:
		return 1
	default:
		return 0
	}
}

func compareFloat(left, right float64) int {
	switch {
	case left < right:
		return -1
	case left > right:
		return 1
	default:
		return 0
	}
}

func intValue(value *int) int {
	if value == nil {
		return -1
	}

	return *value
}

func catalogSortPrice(item platform.TenantProviderAccountModelCatalogItem, input bool) float64 {
	var value *float64

	if item.PublicPricing != nil {
		if input {
			value = item.PublicPricing.InputUSDPer1M
		} else {
			value = item.PublicPricing.OutputUSDPer1M
		}
	}

	if value == nil && len(item.Availability) > 0 && item.Availability[0].EffectivePricing != nil {
		if input {
			value = item.Availability[0].EffectivePricing.InputUSDPer1M
		} else {
			value = item.Availability[0].EffectivePricing.OutputUSDPer1M
		}
	}

	if value == nil {
		return 1.7976931348623157e+308
	}

	return *value
}

func providerCatalogAllowsAuthMode(provider platform.ProviderCatalogProvider, authMode platform.ProviderAuthMode) bool {
	for _, allowed := range provider.AuthModes {
		if allowed == authMode {
			return true
		}
	}

	return false
}

// TestTenantProviderAccountCredentials validates configured credentials (or platform-managed credentials).
func (h *Handlers) TestTenantProviderAccountCredentials(c *fiber.Ctx) error {
	if err := requirePermissionAnyOrTenantAdmin(c, "provider account manage permission required", []string{
		"provider_account:manage",
		"provider_account:create",
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
		parseErr := c.BodyParser(&req)
		if parseErr != nil {
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
		"provider_account:create",
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

	parseErr := c.BodyParser(&req)
	if parseErr != nil {
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

	permissionErr := requireAccessPrincipalSearchPermission(c, kind)
	if permissionErr != nil {
		return permissionErr
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
