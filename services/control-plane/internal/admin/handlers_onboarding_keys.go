package admin

import (
	"errors"
	"strconv"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/storage"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
)

func (h *Handlers) CheckTenantSlugAvailability(c *fiber.Ctx) error {
	store, err := h.requireOnboardingStore()
	if err != nil {
		return err
	}
	slug := strings.TrimSpace(c.Params("slug"))
	result, err := store.CheckTenantSlugAvailability(c.Context(), slug)
	if err != nil {
		return err
	}
	return c.JSON(result)
}

func (h *Handlers) RequestOnboardingOTP(c *fiber.Ctx) error {
	store, err := h.requireOnboardingStore()
	if err != nil {
		return err
	}
	var req storage.CreateOnboardingOTPParams
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	resp, err := store.CreateOnboardingOTP(c.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrTenantSlugInvalid):
			return fiber.NewError(fiber.StatusBadRequest, "tenant slug is invalid or reserved")
		case errors.Is(err, storage.ErrTenantSlugTaken):
			return fiber.NewError(fiber.StatusConflict, "tenant slug is already taken")
		default:
			return err
		}
	}
	if resp.AlreadyPending {
		return c.Status(fiber.StatusOK).JSON(resp)
	}
	return c.Status(fiber.StatusCreated).JSON(resp)
}

func (h *Handlers) VerifyOnboardingOTP(c *fiber.Ctx) error {
	store, err := h.requireOnboardingStore()
	if err != nil {
		return err
	}
	var req storage.VerifyOnboardingOTPParams
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	resp, err := store.VerifyOnboardingOTPAndProvision(c.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrOnboardingNotFound):
			return fiber.NewError(fiber.StatusNotFound, "onboarding request not found")
		case errors.Is(err, storage.ErrOnboardingExpired):
			return fiber.NewError(fiber.StatusGone, "onboarding request expired")
		case errors.Is(err, storage.ErrOnboardingOTPFailed):
			return fiber.NewError(fiber.StatusBadRequest, "invalid onboarding otp")
		default:
			return err
		}
	}
	return c.Status(fiber.StatusCreated).JSON(resp)
}

func (h *Handlers) createKey(c *fiber.Ctx, forcedKeyType string) error {
	if err := requireTeamAdminOrHigher(c); err != nil {
		return err
	}
	if h.keys == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "key store not configured")
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}

	var req platform.CreateKeyParams
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	req.Normalize()
	if req.KeyAlias == "" {
		return fiber.NewError(fiber.StatusBadRequest, "key_alias is required")
	}

	if forcedKeyType != "" {
		req.KeyType = forcedKeyType
	}

	req.TenantID = currentTenant.ID
	switch req.KeyType {
	case "", platform.KeyTypePAT:
		req.KeyType = platform.KeyTypePAT
		req.Role = platform.LegacyKeyRoleVirtualKey
	case platform.KeyTypeVAT:
		req.Role = platform.LegacyKeyRoleServiceAccount
	default:
		return fiber.NewError(fiber.StatusBadRequest, "key_type must be pat or vat")
	}
	if req.KeyType == platform.KeyTypeVAT && req.TeamID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "team_id is required for vat")
	}

	// PAT no longer carries model-level allowlist. Keep backward compatibility by
	// accepting empty model list for PAT while retaining explicit-model validation for VAT.
	if req.KeyType == platform.KeyTypePAT && len(req.AllowedModels) == 0 {
		req.Models = []string{}
		req.AllowedModels = []string{}
	} else {
		normalizedModels, err := normalizeAndValidateModelFQNs(req.AllowedModels)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		req.AllowedModels = normalizedModels
		req.Models = normalizedModels
	}

	scopeUserID := strings.TrimSpace(c.Get("x-shahrah-user-id"))
	scopeTeamID := ""
	if req.KeyType == platform.KeyTypeVAT {
		scopeUserID = ""
		scopeTeamID = strings.TrimSpace(req.TeamID)
	}
	if len(req.AllowedModels) > 0 {
		scopeAllowedModels, err := h.keys.ListAllowedModelFQNs(c.Context(), currentTenant.ID, scopeUserID, scopeTeamID)
		if err != nil {
			return err
		}
		if err := ensureRequestedModelsInScope(req.AllowedModels, scopeAllowedModels); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
	}

	created, err := h.keys.CreateKey(c.Context(), req)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(created)
}

// CreateKey preserves the legacy PAT/VAT key creation endpoint.
func (h *Handlers) CreateKey(c *fiber.Ctx) error {
	return h.createKey(c, "")
}

// CreateVATKey creates a tenant-scoped VAT using the canonical endpoint.
func (h *Handlers) CreateVATKey(c *fiber.Ctx) error {
	return h.createKey(c, platform.KeyTypeVAT)
}

// ListKeys returns all tenant keys without exposing raw key material.
func (h *Handlers) ListKeys(c *fiber.Ctx) error {
	if err := requireTeamAdminOrHigher(c); err != nil {
		return err
	}
	if h.keys == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "key store not configured")
	}

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

func (h *Handlers) ListPersonalAccessTokens(c *fiber.Ctx) error {
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	actorUserID := strings.TrimSpace(c.Get("x-shahrah-user-id"))
	if actorUserID == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "x-shahrah-user-id header is required")
	}

	query := strings.TrimSpace(c.Query("query"))
	limit := 25
	offset := 0
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		parsed, parseErr := strconv.Atoi(raw)
		if parseErr != nil || parsed <= 0 {
			return fiber.NewError(fiber.StatusBadRequest, "limit must be a positive number")
		}
		if parsed > 200 {
			parsed = 200
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

	items, total, err := h.keys.ListPATs(c.Context(), currentTenant.ID, actorUserID, query, limit, offset)
	if err != nil {
		return err
	}
	var nextOffset *int
	if offset+len(items) < total {
		next := offset + len(items)
		nextOffset = &next
	}
	return c.JSON(platform.ListPATResponse{
		Data:       items,
		Total:      total,
		Limit:      limit,
		Offset:     offset,
		NextOffset: nextOffset,
	})
}

func (h *Handlers) PersonalAccessTokenExists(c *fiber.Ctx) error {
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	actorUserID := strings.TrimSpace(c.Get("x-shahrah-user-id"))
	if actorUserID == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "x-shahrah-user-id header is required")
	}
	name := strings.TrimSpace(c.Query("name"))
	if name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "name is required")
	}
	exists, err := h.keys.PATExists(c.Context(), currentTenant.ID, actorUserID, name)
	if err != nil {
		return err
	}
	return c.JSON(platform.PATExistsResponse{Exists: exists})
}

func (h *Handlers) CreatePersonalAccessToken(c *fiber.Ctx) error {
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	actorUserID := strings.TrimSpace(c.Get("x-shahrah-user-id"))
	if actorUserID == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "x-shahrah-user-id header is required")
	}

	var req platform.CreatePATRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "name is required")
	}
	var rawBody map[string]interface{}
	if err := c.BodyParser(&rawBody); err == nil {
		if _, ok := rawBody["models"]; ok {
			return fiber.NewError(fiber.StatusBadRequest, "models/allowed_models are no longer accepted for PAT creation; access is controlled by Provider Account and Virtual Model access settings")
		}
		if _, ok := rawBody["allowed_models"]; ok {
			return fiber.NewError(fiber.StatusBadRequest, "models/allowed_models are no longer accepted for PAT creation; access is controlled by Provider Account and Virtual Model access settings")
		}
	}

	exists, err := h.keys.PATExists(c.Context(), currentTenant.ID, actorUserID, req.Name)
	if err != nil {
		return err
	}
	if exists {
		return fiber.NewError(fiber.StatusConflict, "personal access token with this name already exists")
	}

	created, err := h.keys.CreateKey(c.Context(), platform.CreateKeyParams{
		TenantID:      currentTenant.ID,
		KeyAlias:      req.Name,
		Role:          platform.LegacyKeyRoleVirtualKey,
		KeyType:       platform.KeyTypePAT,
		Models:        []string{},
		AllowedModels: []string{},
		UserID:        actorUserID,
		ExpiresIn:     strings.TrimSpace(req.ExpiresIn),
		Metadata:      map[string]interface{}{},
	})
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(created)
}

func (h *Handlers) ListAllowedModels(c *fiber.Ctx) error {
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	actorUserID := strings.TrimSpace(c.Get("x-shahrah-user-id"))
	if actorUserID == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "x-shahrah-user-id header is required")
	}

	keyType := strings.TrimSpace(strings.ToLower(c.Query("key_type")))
	if keyType == "" {
		keyType = platform.KeyTypePAT
	}
	if keyType != platform.KeyTypePAT && keyType != platform.KeyTypeVAT {
		return fiber.NewError(fiber.StatusBadRequest, "key_type must be pat or vat")
	}

	teamID := strings.TrimSpace(c.Query("team_id"))
	scopeUserID := actorUserID
	scopeTeamID := ""
	if keyType == platform.KeyTypeVAT {
		scopeUserID = ""
		scopeTeamID = teamID
		if scopeTeamID == "" {
			return c.JSON(fiber.Map{"data": []fiber.Map{}})
		}
	}

	models, err := h.keys.ListAllowedModelFQNs(c.Context(), currentTenant.ID, scopeUserID, scopeTeamID)
	if err != nil {
		return err
	}
	items := make([]fiber.Map, 0, len(models))
	seen := make(map[string]struct{}, len(models))
	for _, model := range models {
		model = strings.TrimSpace(model)
		if model == "" {
			continue
		}
		if _, exists := seen[model]; exists {
			continue
		}
		seen[model] = struct{}{}
		items = append(items, fiber.Map{"id": model})
	}
	virtualModels, err := h.models.ListVirtualModelsForPrincipal(c.Context(), currentTenant.ID, scopeUserID, scopeTeamID)
	if err != nil {
		return err
	}
	scopeAllowedSet := make(map[string]struct{}, len(models))
	for _, model := range models {
		model = strings.TrimSpace(model)
		if model == "" {
			continue
		}
		scopeAllowedSet[model] = struct{}{}
	}
	for _, vm := range virtualModels {
		alias := strings.TrimSpace(vm.Alias)
		if alias == "" {
			continue
		}
		if _, exists := seen[alias]; exists {
			continue
		}
		if !isVirtualModelInScope(vm, scopeAllowedSet) {
			continue
		}
		seen[alias] = struct{}{}
		items = append(items, fiber.Map{"id": alias})
	}
	return c.JSON(fiber.Map{"data": items})
}

func (h *Handlers) RevokePersonalAccessToken(c *fiber.Ctx) error {
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	actorUserID := strings.TrimSpace(c.Get("x-shahrah-user-id"))
	if actorUserID == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "x-shahrah-user-id header is required")
	}
	keyID := strings.TrimSpace(c.Params("id"))
	if keyID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "id is required")
	}
	revoked, err := h.keys.RevokePAT(c.Context(), currentTenant.ID, actorUserID, keyID, actorUserID)
	if err != nil {
		return err
	}
	if !revoked {
		return fiber.NewError(fiber.StatusNotFound, "personal access token not found")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handlers) RevokeAllPersonalAccessTokens(c *fiber.Ctx) error {
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	actorUserID := strings.TrimSpace(c.Get("x-shahrah-user-id"))
	if actorUserID == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "x-shahrah-user-id header is required")
	}
	revoked, err := h.keys.RevokeAllPATs(c.Context(), currentTenant.ID, actorUserID, actorUserID)
	if err != nil {
		return err
	}
	return c.JSON(platform.RevokeAllPATResponse{Revoked: revoked})
}
