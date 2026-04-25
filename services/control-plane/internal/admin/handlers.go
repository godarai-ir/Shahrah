package admin

import (
	"context"

	"github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
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

// Handlers serves bootstrap management APIs.
type Handlers struct {
	keys    keyStore
	models  modelStore
	wallets walletStore
}

// New creates a management handlers bundle.
func New(keys keyStore, models modelStore, wallets walletStore) *Handlers {
	return &Handlers{keys: keys, models: models, wallets: wallets}
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

	data := make([]openai.Model, 0, len(deployments))
	for _, dep := range deployments {
		data = append(data, openai.Model{
			ID:       dep.ID,
			Object:   "model",
			OwnedBy:  "shahrah",
			Provider: dep.Provider,
			Aliases:  []string{dep.ModelName},
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
