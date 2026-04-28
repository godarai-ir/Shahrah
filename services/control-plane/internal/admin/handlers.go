package admin

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

type keyStore interface {
	CreateKey(ctx context.Context, params platform.CreateKeyParams) (*platform.CreatedKey, error)
	ListKeys(ctx context.Context, tenantID string) ([]platform.CreatedKey, error)
}

type modelStore interface {
	CreateDeployment(ctx context.Context, dep platform.Deployment) (*platform.Deployment, error)
	ListDeployments(ctx context.Context, tenantID string) ([]platform.Deployment, error)
}

type walletStore interface {
	GetWallet(ctx context.Context, tenantID string) (*platform.Wallet, error)
	TopUpWallet(ctx context.Context, tenantID string, amountIRR float64, metadata map[string]interface{}) (*platform.Transaction, error)
}

type providerAccountStore interface {
	CreateTenantProviderAccount(ctx context.Context, tenantID string, req platform.CreateTenantProviderAccountRequest) (*platform.TenantProviderAccount, error)
	ListTenantProviderAccounts(ctx context.Context, tenantID string) ([]platform.TenantProviderAccount, error)
	GetTenantProviderAccountDetails(ctx context.Context, tenantID string, accountID string) (*platform.TenantProviderAccountDetailsResponse, error)
	UpsertTenantProviderAccountModels(ctx context.Context, tenantID string, accountID string, models []platform.TenantProviderAccountModel) ([]platform.TenantProviderAccountModel, error)
	UpdateTenantProviderAccountAccess(ctx context.Context, tenantID string, accountID string, access platform.TenantProviderAccountAccessControl) (*platform.TenantProviderAccountAccessControl, error)
	TestTenantProviderAccountCredentials(ctx context.Context, tenantID string, accountID string, req platform.TestTenantProviderAccountCredentialsRequest) (*platform.TestTenantProviderAccountCredentialsResponse, error)
	SearchAccessPrincipals(ctx context.Context, kind string, query string, limit int) ([]platform.AccessPrincipalCandidate, error)
}

// Handlers serves bootstrap management APIs.
type Handlers struct {
	keys             keyStore
	models           modelStore
	wallets          walletStore
	providerAccounts providerAccountStore
}

// New creates a management handlers bundle.
func New(keys keyStore, models modelStore, wallets walletStore) *Handlers {
	accounts, _ := models.(providerAccountStore)
	return &Handlers{
		keys:             keys,
		models:           models,
		wallets:          wallets,
		providerAccounts: accounts,
	}
}

// CreateKey creates a tenant-scoped virtual key.
func (h *Handlers) CreateKey(c *fiber.Ctx) error {
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}

	var req platform.CreateKeyParams
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	req.TenantID = currentTenant.ID
	if req.Role == "" {
		req.Role = "virtual_key"
	}

	created, err := h.keys.CreateKey(c.Context(), req)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(created)
}

// ListKeys lists keys for the current tenant.
func (h *Handlers) ListKeys(c *fiber.Ctx) error {
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	items, err := h.keys.ListKeys(c.Context(), currentTenant.ID)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"data": items})
}

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
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(dep)
}

// ListModels lists tenant-visible model aliases.
func (h *Handlers) ListModels(c *fiber.Ctx) error {
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
		byAlias[id] = model
	}

	data := make([]openai.Model, 0, len(byAlias))
	for _, item := range byAlias {
		data = append(data, openai.Model{
			ID:       item.ID,
			Object:   "model",
			OwnedBy:  item.OwnedBy,
			Provider: item.Provider,
			Aliases:  item.Aliases,
		})
	}
	return c.JSON(openai.ModelListResponse{Object: "list", Data: data})
}

// GetWallet returns the current tenant wallet.
func (h *Handlers) GetWallet(c *fiber.Ctx) error {
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
	items, err := discoveryProviders(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, err.Error())
	}
	return c.JSON(platform.ProviderCatalogListResponse{Data: items})
}

// ListProviderModelsCatalog returns provider model catalog for step-2 selection.
func (h *Handlers) ListProviderModelsCatalog(c *fiber.Ctx) error {
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
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(account)
}

// ListTenantProviderAccounts returns all provider accounts configured for the tenant.
func (h *Handlers) ListTenantProviderAccounts(c *fiber.Ctx) error {
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

// UpsertTenantProviderAccountModels replaces selected models for a provider account.
func (h *Handlers) UpsertTenantProviderAccountModels(c *fiber.Ctx) error {
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
		if req.Models[i].PricingMode == "" {
			req.Models[i].PricingMode = platform.PricingModePublic
		}
	}

	items, err := store.UpsertTenantProviderAccountModels(c.Context(), currentTenant.ID, accountID, req.Models)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "provider account not found")
		}
		return err
	}
	return c.JSON(fiber.Map{"data": items})
}

// UpdateTenantProviderAccountAccess replaces access control entries for one provider account.
func (h *Handlers) UpdateTenantProviderAccountAccess(c *fiber.Ctx) error {
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

	access, err := store.UpdateTenantProviderAccountAccess(c.Context(), currentTenant.ID, accountID, req.Access)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "provider account not found")
		}
		return err
	}
	return c.JSON(fiber.Map{"access": access})
}

// TestTenantProviderAccountCredentials validates configured credentials (or virtual-key-only mode).
func (h *Handlers) TestTenantProviderAccountCredentials(c *fiber.Ctx) error {
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

// ListAccessPrincipals returns user/team principals for access picker controls.
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
	case "user", "team":
	default:
		return fiber.NewError(fiber.StatusBadRequest, "kind must be user or team")
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

	items, err := store.SearchAccessPrincipals(c.Context(), kind, query, limit)
	if err != nil {
		return err
	}
	return c.JSON(platform.AccessPrincipalListResponse{Data: items})
}

func (h *Handlers) requireProviderAccountsStore() (providerAccountStore, error) {
	if h.providerAccounts == nil {
		return nil, fiber.NewError(fiber.StatusNotImplemented, "provider account store not configured")
	}
	return h.providerAccounts, nil
}

func appendUnique(values []string, candidate string) []string {
	candidate = strings.TrimSpace(candidate)
	if candidate == "" {
		return values
	}
	for _, existing := range values {
		if existing == candidate {
			return values
		}
	}
	return append(values, candidate)
}
