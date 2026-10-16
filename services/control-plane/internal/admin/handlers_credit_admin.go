package admin

import (
	"context"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	billingdomain "github.com/Delpak/shahrah/services/control-plane/internal/domain/billing"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/credits"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// creditAdminStore is the support-admin side of credit billing: the price
// policy, rate cards, the pack catalog, and per-tenant ledgers.
type creditAdminStore interface {
	GetCreditPriceStatus(ctx context.Context, now time.Time) (*platform.CreditPriceStatus, error)
	ApproveCreditPriceRevision(ctx context.Context, revisionID string, actor platform.SupportActor, reason string, now time.Time) (*platform.CreditPriceRevision, error)
	RejectCreditPriceRevision(ctx context.Context, revisionID string, actor platform.SupportActor, reason string) (*platform.CreditPriceRevision, error)
	ListCreditRateCardVersions(ctx context.Context) ([]platform.CreditRateCardVersion, error)
	CreateCreditRateCardDraft(ctx context.Context, actor platform.SupportActor) (int, error)
	UpsertCreditRateCardMultiplier(ctx context.Context, version int, req platform.UpsertCreditMultiplierRequest, actor platform.SupportActor) error
	PublishCreditRateCard(ctx context.Context, version int, req platform.PublishCreditRateCardRequest, actor platform.SupportActor, now time.Time) (*platform.CreditRateCardVersion, error)
	ListCreditPackVersions(ctx context.Context) ([]platform.CreditPackVersion, error)
	CreateCreditPackVersion(ctx context.Context, req platform.CreateCreditPackVersionRequest, actor platform.SupportActor) (*platform.CreditPackVersion, error)
	PublishCreditPackVersion(ctx context.Context, slug string, version int, actor platform.SupportActor) error
	GetTenantCreditOverview(ctx context.Context, tenantID string) (*platform.TenantCreditOverview, error)
	ListTenantCreditLedger(ctx context.Context, tenantID string, limit int, before *time.Time) ([]platform.CreditLedgerEntry, error)
	ListTenantCreditLots(ctx context.Context, tenantID string) ([]platform.CreditLot, error)
	ListCreditLotAllocations(ctx context.Context, tenantID, lotID string) ([]platform.CreditLotAllocation, error)
	AdjustCredits(ctx context.Context, tenantID string, req platform.CreditAdjustmentRequest, actor platform.SupportActor) (*platform.TenantCreditOverview, error)
	SetTenantBillingUnitAudited(ctx context.Context, tenantID string, req platform.SetBillingUnitRequest, actor platform.SupportActor) error
}

func (h *Handlers) requireCreditAdminStore() (creditAdminStore, error) {
	if h.creditAdmin == nil {
		return nil, fiber.NewError(fiber.StatusNotImplemented, "credit administration is not configured")
	}

	return h.creditAdmin, nil
}

// GetPlatformCreditPricePolicy returns the active price, any proposal, and the 7-day average.
func (h *Handlers) GetPlatformCreditPricePolicy(c *fiber.Ctx) error {
	store, err := h.requireCreditAdminStore()
	if err != nil {
		return err
	}

	status, err := store.GetCreditPriceStatus(c.Context(), time.Now())
	if err != nil {
		return err
	}

	return c.JSON(status)
}

// ApprovePlatformCreditPriceRevision activates a proposed list price.
func (h *Handlers) ApprovePlatformCreditPriceRevision(c *fiber.Ctx) error {
	return h.decideCreditPriceRevision(c, true)
}

// RejectPlatformCreditPriceRevision closes a proposed list price.
func (h *Handlers) RejectPlatformCreditPriceRevision(c *fiber.Ctx) error {
	return h.decideCreditPriceRevision(c, false)
}

func (h *Handlers) decideCreditPriceRevision(c *fiber.Ctx, approve bool) error {
	store, err := h.requireCreditAdminStore()
	if err != nil {
		return err
	}

	var req platform.CreditPriceDecisionRequest
	if parseErr := c.BodyParser(&req); parseErr != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	revisionID := strings.TrimSpace(c.Params("revisionID"))
	if _, parseErr := uuid.Parse(revisionID); parseErr != nil {
		return creditPackError(c, billingdomain.ErrCreditPriceRevisionNotFound)
	}

	var revision *platform.CreditPriceRevision
	if approve {
		revision, err = store.ApproveCreditPriceRevision(c.Context(), revisionID, supportActor(c), req.Reason, time.Now())
	} else {
		revision, err = store.RejectCreditPriceRevision(c.Context(), revisionID, supportActor(c), req.Reason)
	}

	if err != nil {
		return creditPackError(c, err)
	}

	return c.JSON(revision)
}

// ListPlatformCreditRateCards returns every rate-card version with its multipliers.
func (h *Handlers) ListPlatformCreditRateCards(c *fiber.Ctx) error {
	store, err := h.requireCreditAdminStore()
	if err != nil {
		return err
	}

	versions, err := store.ListCreditRateCardVersions(c.Context())
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{"data": versions})
}

// CreatePlatformCreditRateCardDraft starts (or returns) the rate-card draft.
func (h *Handlers) CreatePlatformCreditRateCardDraft(c *fiber.Ctx) error {
	store, err := h.requireCreditAdminStore()
	if err != nil {
		return err
	}

	version, err := store.CreateCreditRateCardDraft(c.Context(), supportActor(c))
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"version": version})
}

// UpsertPlatformCreditRateCardMultiplier sets a model multiplier on the draft.
func (h *Handlers) UpsertPlatformCreditRateCardMultiplier(c *fiber.Ctx) error {
	store, err := h.requireCreditAdminStore()
	if err != nil {
		return err
	}

	version, err := c.ParamsInt("version")
	if err != nil || version <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "version must be a positive integer")
	}

	var req platform.UpsertCreditMultiplierRequest
	if parseErr := c.BodyParser(&req); parseErr != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	if err := store.UpsertCreditRateCardMultiplier(c.Context(), version, req, supportActor(c)); err != nil {
		return creditPackError(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// PublishPlatformCreditRateCard publishes the draft, or schedules it when it raises a rate.
func (h *Handlers) PublishPlatformCreditRateCard(c *fiber.Ctx) error {
	store, err := h.requireCreditAdminStore()
	if err != nil {
		return err
	}

	version, err := c.ParamsInt("version")
	if err != nil || version <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "version must be a positive integer")
	}

	var req platform.PublishCreditRateCardRequest
	if parseErr := c.BodyParser(&req); parseErr != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	published, err := store.PublishCreditRateCard(c.Context(), version, req, supportActor(c), time.Now())
	if err != nil {
		return creditPackError(c, err)
	}

	return c.JSON(published)
}

// ListPlatformCreditPacks returns every version of every add-on pack.
func (h *Handlers) ListPlatformCreditPacks(c *fiber.Ctx) error {
	store, err := h.requireCreditAdminStore()
	if err != nil {
		return err
	}

	packs, err := store.ListCreditPackVersions(c.Context())
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{"data": packs})
}

// CreatePlatformCreditPack drafts a new pack version.
func (h *Handlers) CreatePlatformCreditPack(c *fiber.Ctx) error {
	store, err := h.requireCreditAdminStore()
	if err != nil {
		return err
	}

	var req platform.CreateCreditPackVersionRequest
	if parseErr := c.BodyParser(&req); parseErr != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	pack, err := store.CreateCreditPackVersion(c.Context(), req, supportActor(c))
	if err != nil {
		return creditPackError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(pack)
}

// PublishPlatformCreditPack makes a pack version purchasable.
func (h *Handlers) PublishPlatformCreditPack(c *fiber.Ctx) error {
	store, err := h.requireCreditAdminStore()
	if err != nil {
		return err
	}

	version, err := c.ParamsInt("version")
	if err != nil || version <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "version must be a positive integer")
	}

	if err := store.PublishCreditPackVersion(c.Context(), strings.TrimSpace(c.Params("slug")), version, supportActor(c)); err != nil {
		return creditPackError(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// GetPlatformTenantCredits summarizes a tenant's credit balance and settings.
func (h *Handlers) GetPlatformTenantCredits(c *fiber.Ctx) error {
	store, tenantID, err := h.creditAdminTenant(c)
	if err != nil {
		return err
	}

	overview, err := store.GetTenantCreditOverview(c.Context(), tenantID)
	if err != nil {
		return creditPackError(c, err)
	}

	return c.JSON(overview)
}

// ListPlatformTenantCreditLedger pages a tenant's credit ledger, newest first.
func (h *Handlers) ListPlatformTenantCreditLedger(c *fiber.Ctx) error {
	store, tenantID, err := h.creditAdminTenant(c)
	if err != nil {
		return err
	}

	var before *time.Time

	if raw := strings.TrimSpace(c.Query("before")); raw != "" {
		parsed, parseErr := time.Parse(time.RFC3339Nano, raw)
		if parseErr != nil {
			return fiber.NewError(fiber.StatusBadRequest, "before must be an RFC 3339 timestamp")
		}

		before = &parsed
	}

	entries, err := store.ListTenantCreditLedger(c.Context(), tenantID, c.QueryInt("limit", 50), before)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{"data": entries})
}

// ListPlatformTenantCreditLots returns a tenant's lots.
func (h *Handlers) ListPlatformTenantCreditLots(c *fiber.Ctx) error {
	store, tenantID, err := h.creditAdminTenant(c)
	if err != nil {
		return err
	}

	lots, err := store.ListTenantCreditLots(c.Context(), tenantID)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{"data": lots})
}

// ListPlatformTenantCreditLotAllocations drills into one lot.
func (h *Handlers) ListPlatformTenantCreditLotAllocations(c *fiber.Ctx) error {
	store, tenantID, err := h.creditAdminTenant(c)
	if err != nil {
		return err
	}

	lotID := strings.TrimSpace(c.Params("lotID"))
	if _, parseErr := uuid.Parse(lotID); parseErr != nil {
		return fiber.NewError(fiber.StatusBadRequest, "lotID must be a UUID")
	}

	allocations, err := store.ListCreditLotAllocations(c.Context(), tenantID, lotID)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{"data": allocations})
}

// AdjustPlatformTenantCredits writes an audited manual credit adjustment.
func (h *Handlers) AdjustPlatformTenantCredits(c *fiber.Ctx) error {
	store, tenantID, err := h.creditAdminTenant(c)
	if err != nil {
		return err
	}

	var req platform.CreditAdjustmentRequest
	if parseErr := c.BodyParser(&req); parseErr != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	if len(req.Reason) > 1000 || len(req.IdempotencyKey) > 128 {
		return fiber.NewError(fiber.StatusBadRequest, "request field is too long")
	}

	overview, err := store.AdjustCredits(c.Context(), tenantID, req, supportActor(c))
	if err != nil {
		return creditPackError(c, err)
	}

	return c.JSON(overview)
}

// SetPlatformTenantBillingUnit moves a tenant between IRR and credit billing.
// Moving to credit is refused unless the credit ledger is enabled, because the
// ledger's settlement is what debits a credit tenant. Moving back to IRR is
// always allowed.
func (h *Handlers) SetPlatformTenantBillingUnit(c *fiber.Ctx) error {
	store, tenantID, err := h.creditAdminTenant(c)
	if err != nil {
		return err
	}

	var req platform.SetBillingUnitRequest
	if parseErr := c.BodyParser(&req); parseErr != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	if !h.creditLedgerEnabled && strings.TrimSpace(req.BillingUnit) == credits.BillingUnitCredit {
		return creditPackError(c, billingdomain.ErrBillingUnitChangeDisabled)
	}

	if changeErr := store.SetTenantBillingUnitAudited(c.Context(), tenantID, req, supportActor(c)); changeErr != nil {
		return creditPackError(c, changeErr)
	}

	overview, err := store.GetTenantCreditOverview(c.Context(), tenantID)
	if err != nil {
		return creditPackError(c, err)
	}

	return c.JSON(overview)
}

func (h *Handlers) creditAdminTenant(c *fiber.Ctx) (creditAdminStore, string, error) {
	store, err := h.requireCreditAdminStore()
	if err != nil {
		return nil, "", err
	}

	tenantID := strings.TrimSpace(c.Params("tenantID"))
	if _, parseErr := uuid.Parse(tenantID); parseErr != nil {
		return nil, "", fiber.NewError(fiber.StatusBadRequest, "tenantID must be a UUID")
	}

	return store, tenantID, nil
}
