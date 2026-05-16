package admin

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	"github.com/Delpak/shahrah/services/control-plane/internal/storage"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type keyStore interface {
	CreateKey(ctx context.Context, params platform.CreateKeyParams) (*platform.CreatedKey, error)
	ListKeys(ctx context.Context, tenantID string) ([]platform.CreatedKey, error)
	ListAllowedModelFQNs(ctx context.Context, tenantID string, userID string, teamID string) ([]string, error)
	ListPATs(ctx context.Context, tenantID string, userID string, query string, limit int, offset int) ([]platform.CreatedKey, int, error)
	PATExists(ctx context.Context, tenantID string, userID string, name string) (bool, error)
	RevokePAT(ctx context.Context, tenantID string, userID string, keyID string, revokedByUserID string) (bool, error)
	RevokeAllPATs(ctx context.Context, tenantID string, userID string, revokedByUserID string) (int64, error)
}

type modelStore interface {
	CreateDeployment(ctx context.Context, dep platform.Deployment) (*platform.Deployment, error)
	ListDeployments(ctx context.Context, tenantID string) ([]platform.Deployment, error)
	CreateVirtualModel(ctx context.Context, tenantID string, req platform.CreateVirtualModelRequest, actor string) (*platform.VirtualModel, error)
	ListVirtualModels(ctx context.Context, tenantID string) ([]platform.VirtualModel, error)
	GetVirtualModel(ctx context.Context, tenantID string, id string) (*platform.VirtualModel, error)
	UpdateVirtualModel(ctx context.Context, tenantID string, id string, req platform.UpdateVirtualModelRequest, actor string) (*platform.VirtualModel, error)
	DeleteVirtualModel(ctx context.Context, tenantID string, id string) error
	ListVirtualModelsForPrincipal(ctx context.Context, tenantID string, userID string, teamID string) ([]platform.VirtualModel, error)
	CreateRateLimitRule(ctx context.Context, tenantID string, req platform.CreateRateLimitRuleRequest) (*platform.RateLimitRule, error)
	ListRateLimitRules(ctx context.Context, tenantID string, query string, limit int, offset int) ([]platform.RateLimitRule, int, error)
	GetRateLimitRuleByRuleID(ctx context.Context, tenantID string, ruleID string) (*platform.RateLimitRule, error)
	UpdateRateLimitRuleByRuleID(ctx context.Context, tenantID string, ruleID string, req platform.UpdateRateLimitRuleRequest) (*platform.RateLimitRule, error)
	UpdateRateLimitRuleStatusByRuleID(ctx context.Context, tenantID string, ruleID string, enabled bool) (*platform.RateLimitRule, error)
	DeleteRateLimitRuleByRuleID(ctx context.Context, tenantID string, ruleID string) error
	CreateBudgetLimitRule(ctx context.Context, tenantID string, req platform.CreateBudgetLimitRuleRequest) (*platform.BudgetLimitRule, error)
	ListBudgetLimitRules(ctx context.Context, tenantID string, query string, limit int, offset int) ([]platform.BudgetLimitRule, int, error)
	GetBudgetLimitRuleByRuleID(ctx context.Context, tenantID string, ruleID string) (*platform.BudgetLimitRule, error)
	UpdateBudgetLimitRuleByRuleID(ctx context.Context, tenantID string, ruleID string, req platform.UpdateBudgetLimitRuleRequest) (*platform.BudgetLimitRule, error)
	UpdateBudgetLimitRuleStatusByRuleID(ctx context.Context, tenantID string, ruleID string, enabled bool) (*platform.BudgetLimitRule, error)
	DeleteBudgetLimitRuleByRuleID(ctx context.Context, tenantID string, ruleID string) error
}

type walletStore interface {
	GetWallet(ctx context.Context, tenantID string) (*platform.Wallet, error)
	TopUpWallet(ctx context.Context, tenantID string, amountIRR float64, metadata map[string]interface{}) (*platform.Transaction, error)
	GetLatestExchangeRate(ctx context.Context) (*platform.ExchangeRate, error)
	GetWalletTopUpSettings(ctx context.Context) (*platform.WalletTopUpSettings, error)
	UpdateWalletTopUpSettings(ctx context.Context, minUSD, maxUSD, taxPercent, feePercent float64) (*platform.WalletTopUpSettings, error)
	CreateWalletTopUpIntent(ctx context.Context, p storage.CreateTopUpIntentParams) (*platform.WalletTopUpIntent, error)
	AttachWalletTopUpGatewayData(ctx context.Context, intentPublicID, authority, paymentURL string, raw map[string]interface{}) error
	GetWalletTopUpIntentByAuthority(ctx context.Context, authority string) (*platform.WalletTopUpIntent, error)
	CompleteWalletTopUpByAuthority(ctx context.Context, authority string, verifyRaw map[string]interface{}, refID, gatewayTxnID, cardMask, payerPhone string) (*platform.WalletTopUpIntent, bool, error)
	MarkWalletTopUpFailedByAuthority(ctx context.Context, authority string, reason string, raw map[string]interface{}) error
	ListWalletTransactions(ctx context.Context, tenantID string, limit int) ([]platform.Transaction, error)
}

type providerAccountStore interface {
	CreateTenantProviderAccount(ctx context.Context, tenantID string, req platform.CreateTenantProviderAccountRequest) (*platform.TenantProviderAccount, error)
	ListTenantProviderAccounts(ctx context.Context, tenantID string) ([]platform.TenantProviderAccount, error)
	GetTenantProviderAccountDetails(ctx context.Context, tenantID string, accountID string) (*platform.TenantProviderAccountDetailsResponse, error)
	UpdateTenantProviderAccountCredentials(ctx context.Context, tenantID string, accountID string, req platform.UpdateTenantProviderAccountCredentialsRequest) (*platform.TenantProviderAccount, error)
	UpsertTenantProviderAccountModels(ctx context.Context, tenantID string, accountID string, models []platform.TenantProviderAccountModel) ([]platform.TenantProviderAccountModel, error)
	UpdateTenantProviderAccountAccess(ctx context.Context, tenantID string, accountID string, access platform.TenantProviderAccountAccessControl) (*platform.TenantProviderAccountAccessControl, error)
	TestTenantProviderAccountCredentials(ctx context.Context, tenantID string, accountID string, req platform.TestTenantProviderAccountCredentialsRequest) (*platform.TestTenantProviderAccountCredentialsResponse, error)
	ListTenantProviderAccountEffectiveModels(ctx context.Context, tenantID string, accountID string, keyType string, userID string, teamID string) ([]platform.EffectiveModelView, error)
	IsTenantProviderAccountManager(ctx context.Context, tenantID string, accountID string, userID string, teamID string) (bool, error)
	DeleteTenantProviderAccount(ctx context.Context, tenantID string, accountID string) error
	SearchAccessPrincipals(ctx context.Context, tenantID string, kind string, query string, limit int) ([]platform.AccessPrincipalCandidate, error)
}

type orgUnitStore interface {
	CreateOrgUnit(ctx context.Context, tenantID string, req platform.CreateOrgUnitRequest) (*platform.OrgUnit, error)
	ListOrgUnits(ctx context.Context, tenantID string, query string, limit int, offset int) ([]platform.OrgUnit, int, error)
	GetOrgUnit(ctx context.Context, tenantID string, orgUnitID string) (*platform.OrgUnit, error)
	UpdateOrgUnit(ctx context.Context, tenantID string, orgUnitID string, req platform.UpdateOrgUnitRequest) (*platform.OrgUnit, error)
	DeleteOrgUnit(ctx context.Context, tenantID string, orgUnitID string) error
	ListOrgUnitMemberships(ctx context.Context, tenantID string, orgUnitID string) ([]platform.OrgUnitMembership, error)
	AddOrgUnitMembership(ctx context.Context, tenantID string, orgUnitID string, req platform.CreateOrgUnitMembershipRequest) (*platform.OrgUnitMembership, error)
	UpdateOrgUnitMembership(ctx context.Context, tenantID string, orgUnitID string, userID string, req platform.UpdateOrgUnitMembershipRequest) (*platform.OrgUnitMembership, error)
	DeleteOrgUnitMembership(ctx context.Context, tenantID string, orgUnitID string, userID string) error
}

type virtualUserStore interface {
	CreateVirtualUser(ctx context.Context, tenantID string, req platform.CreateVirtualUserRequest) (*platform.VirtualUser, error)
	ListVirtualUsers(ctx context.Context, tenantID string, query string, limit int, offset int) ([]platform.VirtualUser, int, error)
	GetVirtualUser(ctx context.Context, tenantID string, virtualUserID string) (*platform.VirtualUser, error)
	UpdateVirtualUser(ctx context.Context, tenantID string, virtualUserID string, req platform.UpdateVirtualUserRequest) (*platform.VirtualUser, error)
	DeleteVirtualUser(ctx context.Context, tenantID string, virtualUserID string) error
}

type membershipStore interface {
	ListTeamMemberships(ctx context.Context, tenantID string, teamID string) ([]platform.TeamMembership, error)
	AddTeamMembership(ctx context.Context, tenantID string, teamID string, req platform.CreateTeamMembershipRequest) (*platform.TeamMembership, error)
	UpdateTeamMembership(ctx context.Context, tenantID string, teamID string, userID string, req platform.UpdateTeamMembershipRequest) (*platform.TeamMembership, error)
	DeleteTeamMembership(ctx context.Context, tenantID string, teamID string, userID string) error
}

type teamStore interface {
	CreateTeam(ctx context.Context, tenantID string, req platform.CreateTeamRequest) (*platform.Team, error)
	ListTeams(ctx context.Context, tenantID string, query string, limit int, offset int) ([]platform.Team, int, error)
	GetTeam(ctx context.Context, tenantID string, teamID string) (*platform.Team, error)
	UpdateTeam(ctx context.Context, tenantID string, teamID string, req platform.UpdateTeamRequest) (*platform.Team, error)
	DeleteTeam(ctx context.Context, tenantID string, teamID string) error
	GetTeamPermissions(ctx context.Context, tenantID string, teamID string) (*platform.TeamPermissionsResponse, error)
	UpdateTeamPermissions(ctx context.Context, tenantID string, teamID string, req platform.UpdateTeamPermissionsRequest) (*platform.TeamPermissionsResponse, error)
}

type userStore interface {
	ListTenantAdminUsers(ctx context.Context, tenantSlug string, query string, limit int, offset int) ([]platform.TenantAdminUser, int, error)
	InviteTenantAdminUser(ctx context.Context, tenantID string, tenantSlug string, req platform.InviteTenantAdminUserRequest) (*platform.TenantAdminInvitation, string, error)
	AcceptTenantAdminInvitation(ctx context.Context, tenantID string, tenantSlug string, token string) (*platform.TenantAdminInvitation, *platform.TenantAdminUser, error)
	SetTenantAdminUserActive(ctx context.Context, tenantSlug string, userID string, active bool) (*platform.TenantAdminUser, error)
}

type traceStore interface {
	ListRequestTraces(ctx context.Context, params platform.ListRequestTracesParams) ([]platform.RequestTrace, int, error)
	GetRequestTrace(ctx context.Context, tenantID string, traceID string) (*platform.RequestTrace, error)
	ListTraceSpans(ctx context.Context, tenantID string, traceID string) ([]platform.TraceSpan, error)
	ListRequestTracesV2(ctx context.Context, params platform.ListRequestTracesV2Params) ([]platform.RequestTraceV2, int, error)
	GetRequestTraceV2(ctx context.Context, tenantID string, traceID string) (*platform.RequestTraceDetailV2, error)
	ListTraceTimelineV2(ctx context.Context, tenantID string, traceID string) ([]platform.TraceSpan, error)
	ListTraceAttemptsV2(ctx context.Context, tenantID string, traceID string) ([]platform.TraceAttemptV2, error)
	GetTraceFacetsV2(ctx context.Context, params platform.ListRequestTracesV2Params) (*platform.TraceFacetsV2, error)
}

type metricsStore interface {
	QueryModelMetrics(ctx context.Context, tenantID string, req platform.MetricsQueryRequest) ([]map[string]interface{}, error)
}

type tracePrivacyStore interface {
	GetTenantTracePrivacySettings(ctx context.Context, tenantID string) (*platform.TenantTracePrivacySettings, error)
	UpsertTenantTracePrivacySettings(ctx context.Context, tenantID, capturePolicy, updatedByUserID string) (*platform.TenantTracePrivacySettings, error)
}

type onboardingStore interface {
	CheckTenantSlugAvailability(ctx context.Context, slug string) (*storage.TenantSlugAvailability, error)
	CreateOnboardingOTP(ctx context.Context, req storage.CreateOnboardingOTPParams) (*storage.OnboardingOTP, error)
	VerifyOnboardingOTPAndProvision(ctx context.Context, req storage.VerifyOnboardingOTPParams) (*storage.ProvisionedTenant, error)
}

// Handlers serves bootstrap management APIs.
type Handlers struct {
	keys             keyStore
	models           modelStore
	wallets          walletStore
	providerAccounts providerAccountStore
	orgUnits         orgUnitStore
	memberships      membershipStore
	users            userStore
	virtualUsers     virtualUserStore
	teams            teamStore
	onboarding       onboardingStore
	traces           traceStore
	metrics          metricsStore
	tracePrivacy     tracePrivacyStore
}

// New creates a management handlers bundle.
func New(keys keyStore, models modelStore, wallets walletStore) *Handlers {
	accounts, _ := models.(providerAccountStore)
	orgUnits, _ := keys.(orgUnitStore)
	memberships, _ := keys.(membershipStore)
	users, _ := models.(userStore)
	virtualUsers, _ := keys.(virtualUserStore)
	teams, _ := keys.(teamStore)
	onboarding, _ := keys.(onboardingStore)
	traces, _ := keys.(traceStore)
	metrics, _ := keys.(metricsStore)
	tracePrivacy, _ := keys.(tracePrivacyStore)
	if tracePrivacy == nil {
		tracePrivacy, _ = wallets.(tracePrivacyStore)
	}
	return &Handlers{
		keys:             keys,
		models:           models,
		wallets:          wallets,
		providerAccounts: accounts,
		orgUnits:         orgUnits,
		memberships:      memberships,
		users:            users,
		virtualUsers:     virtualUsers,
		teams:            teams,
		onboarding:       onboarding,
		traces:           traces,
		metrics:          metrics,
		tracePrivacy:     tracePrivacy,
	}
}

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

	normalizedModels, err := normalizeAndValidateModelFQNs(req.AllowedModels)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	req.AllowedModels = normalizedModels
	req.Models = normalizedModels

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
	scopeUserID := strings.TrimSpace(c.Get("x-shahrah-user-id"))
	scopeTeamID := ""
	if req.KeyType == platform.KeyTypeVAT {
		scopeUserID = ""
		scopeTeamID = strings.TrimSpace(req.TeamID)
	}
	scopeAllowedModels, err := h.keys.ListAllowedModelFQNs(c.Context(), currentTenant.ID, scopeUserID, scopeTeamID)
	if err != nil {
		return err
	}
	if err := ensureRequestedModelsInScope(req.AllowedModels, scopeAllowedModels); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	created, err := h.keys.CreateKey(c.Context(), req)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(created)
}

// CreateKey creates a tenant-scoped key using the transitional endpoint.
func (h *Handlers) CreateKey(c *fiber.Ctx) error {
	return h.createKey(c, "")
}

// CreatePATKey creates a tenant-scoped PAT using the canonical endpoint.
func (h *Handlers) CreatePATKey(c *fiber.Ctx) error {
	return h.createKey(c, platform.KeyTypePAT)
}

// CreateVATKey creates a tenant-scoped VAT using the canonical endpoint.
func (h *Handlers) CreateVATKey(c *fiber.Ctx) error {
	return h.createKey(c, platform.KeyTypeVAT)
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
	if len(req.AllowedModels) == 0 && len(req.Models) > 0 {
		req.AllowedModels = append([]string(nil), req.Models...)
	}
	models, err := normalizeRequestedModelIDs(req.AllowedModels)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	req.AllowedModels = models

	scopeAllowedModels, err := h.keys.ListAllowedModelFQNs(c.Context(), currentTenant.ID, actorUserID, "")
	if err != nil {
		return err
	}
	scopeAllowedSet := make(map[string]struct{}, len(scopeAllowedModels))
	for _, item := range scopeAllowedModels {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		scopeAllowedSet[item] = struct{}{}
	}
	requestedFQNs, requestedAliases := splitRequestedModels(req.AllowedModels)
	if len(requestedFQNs) > 0 {
		if err := ensureRequestedModelsInScope(requestedFQNs, scopeAllowedModels); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
	}
	virtualModels, err := h.models.ListVirtualModelsForPrincipal(c.Context(), currentTenant.ID, actorUserID, "")
	if err != nil {
		return err
	}
	vmSnapshots, err := buildVirtualModelSnapshotsForAliases(requestedAliases, virtualModels, scopeAllowedSet)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	exists, err := h.keys.PATExists(c.Context(), currentTenant.ID, actorUserID, req.Name)
	if err != nil {
		return err
	}
	if exists {
		return fiber.NewError(fiber.StatusConflict, "personal access token with this name already exists")
	}

	metadata := map[string]interface{}{}
	if len(vmSnapshots) > 0 {
		metadata["virtual_model_snapshots"] = vmSnapshots
	}

	created, err := h.keys.CreateKey(c.Context(), platform.CreateKeyParams{
		TenantID:      currentTenant.ID,
		KeyAlias:      req.Name,
		Role:          platform.LegacyKeyRoleVirtualKey,
		KeyType:       platform.KeyTypePAT,
		Models:        req.AllowedModels,
		AllowedModels: req.AllowedModels,
		UserID:        actorUserID,
		ExpiresIn:     strings.TrimSpace(req.ExpiresIn),
		Metadata:      metadata,
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
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(dep)
}

// ListModels lists tenant-visible model aliases.
func (h *Handlers) ListModels(c *fiber.Ctx) error {
	if err := requireTenantAdminOrHigher(c); err != nil {
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

func (h *Handlers) CreateVirtualModel(c *fiber.Ctx) error {
	if err := requireTeamAdminOrHigher(c); err != nil {
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
	actor := strings.TrimSpace(c.Get("x-shahrah-user-id"))
	item, err := h.models.CreateVirtualModel(c.Context(), currentTenant.ID, req, actor)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *Handlers) ListVirtualModels(c *fiber.Ctx) error {
	if err := requireTeamAdminOrHigher(c); err != nil {
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
	return c.JSON(fiber.Map{"data": items})
}

func (h *Handlers) GetVirtualModel(c *fiber.Ctx) error {
	if err := requireTeamAdminOrHigher(c); err != nil {
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
		if errors.Is(err, storage.ErrVirtualModelNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "virtual model not found")
		}
		return err
	}
	return c.JSON(item)
}

func (h *Handlers) UpdateVirtualModel(c *fiber.Ctx) error {
	if err := requireTeamAdminOrHigher(c); err != nil {
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
	if err := validateUpdateVirtualModelRequest(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	actor := strings.TrimSpace(c.Get("x-shahrah-user-id"))
	item, err := h.models.UpdateVirtualModel(c.Context(), currentTenant.ID, id, req, actor)
	if err != nil {
		if errors.Is(err, storage.ErrVirtualModelNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "virtual model not found")
		}
		return err
	}
	return c.JSON(item)
}

func (h *Handlers) DeleteVirtualModel(c *fiber.Ctx) error {
	if err := requireTenantAdminOrHigher(c); err != nil {
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
		if errors.Is(err, storage.ErrVirtualModelNotFound) || errors.Is(err, pgx.ErrNoRows) {
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
	if strings.ToLower(strings.TrimSpace(req.RoutingStrategy)) != "priority" {
		return errors.New("routing_strategy must be priority in MVP")
	}
	if len(req.Targets) == 0 {
		return errors.New("at least one target is required")
	}
	return validateVirtualModelTargets(req.Targets)
}

func validateUpdateVirtualModelRequest(req platform.UpdateVirtualModelRequest) error {
	if req.RoutingStrategy != nil && strings.ToLower(strings.TrimSpace(*req.RoutingStrategy)) != "priority" {
		return errors.New("routing_strategy must be priority in MVP")
	}
	if len(req.Targets) > 0 {
		return validateVirtualModelTargets(req.Targets)
	}
	return nil
}

func validateVirtualModelTargets(targets []platform.VirtualModelTarget) error {
	seenOrders := make(map[int]struct{}, len(targets))
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
		if t.Policy.Retry.Attempts < 0 || t.Policy.Retry.Attempts > 5 {
			return errors.New("target.retry.attempts must be between 0 and 5")
		}
		if t.Policy.Retry.DelayMS < 0 || t.Policy.Retry.DelayMS > 10000 {
			return errors.New("target.retry.delay_ms must be between 0 and 10000")
		}
	}
	return nil
}

func (h *Handlers) ListRateLimitRules(c *fiber.Ctx) error {
	if err := requireTenantAdminOrHigher(c); err != nil {
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
	items, total, err := h.models.ListRateLimitRules(c.Context(), currentTenant.ID, query, limit, offset)
	if err != nil {
		return err
	}
	var nextOffset *int
	if offset+len(items) < total {
		n := offset + len(items)
		nextOffset = &n
	}
	return c.JSON(platform.RateLimitRuleListResponse{
		Data:       items,
		Total:      total,
		Limit:      limit,
		Offset:     offset,
		NextOffset: nextOffset,
	})
}

func (h *Handlers) CreateRateLimitRule(c *fiber.Ctx) error {
	if err := requireTenantAdminOrHigher(c); err != nil {
		return err
	}
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	req, err := parseCreateRateLimitRuleRequest(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if err := validateCreateRateLimitRuleRequest(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	item, err := h.models.CreateRateLimitRule(c.Context(), currentTenant.ID, req)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusUnprocessableEntity, "ثبت قانون انجام نشد. لطفا فیلدهای قانون و شرایط را دوباره بررسی کنید.")
		}
		return mapProviderAccountStoreError(err)
	}
	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *Handlers) GetRateLimitRule(c *fiber.Ctx) error {
	if err := requireTenantAdminOrHigher(c); err != nil {
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
	item, err := h.models.GetRateLimitRuleByRuleID(c.Context(), currentTenant.ID, ruleID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "rate limit rule not found")
		}
		return err
	}
	return c.JSON(item)
}

func (h *Handlers) UpdateRateLimitRule(c *fiber.Ctx) error {
	if err := requireTenantAdminOrHigher(c); err != nil {
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
	if err := validateUpdateRateLimitRuleRequest(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	item, err := h.models.UpdateRateLimitRuleByRuleID(c.Context(), currentTenant.ID, ruleID, req)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "rate limit rule not found")
		}
		return mapProviderAccountStoreError(err)
	}
	return c.JSON(item)
}

func (h *Handlers) DeleteRateLimitRule(c *fiber.Ctx) error {
	if err := requireTenantAdminOrHigher(c); err != nil {
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
	if err := h.models.DeleteRateLimitRuleByRuleID(c.Context(), currentTenant.ID, ruleID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "rate limit rule not found")
		}
		return err
	}
	return c.JSON(fiber.Map{"success": true})
}

func (h *Handlers) UpdateRateLimitRuleStatus(c *fiber.Ctx) error {
	if err := requireTenantAdminOrHigher(c); err != nil {
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
	var req platform.UpdateRateLimitRuleStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	item, err := h.models.UpdateRateLimitRuleStatusByRuleID(c.Context(), currentTenant.ID, ruleID, req.Enabled)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "rate limit rule not found")
		}
		return err
	}
	return c.JSON(item)
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
	var raw map[string]interface{}
	if err := c.BodyParser(&raw); err != nil {
		return platform.CreateRateLimitRuleRequest{}, err
	}
	effectiveFrom, err := parseOptionalTimeValue(raw["effective_from"])
	if err != nil {
		return platform.CreateRateLimitRuleRequest{}, err
	}
	expiresAt, err := parseOptionalTimeValue(raw["expires_at"])
	if err != nil {
		return platform.CreateRateLimitRuleRequest{}, err
	}
	raw["effective_from"] = nil
	raw["expires_at"] = nil

	encoded, err := json.Marshal(raw)
	if err != nil {
		return platform.CreateRateLimitRuleRequest{}, err
	}
	var req platform.CreateRateLimitRuleRequest
	if err := json.Unmarshal(encoded, &req); err != nil {
		return platform.CreateRateLimitRuleRequest{}, err
	}
	req.EffectiveFrom = effectiveFrom
	req.ExpiresAt = expiresAt
	return req, nil
}

func parseUpdateRateLimitRuleRequest(c *fiber.Ctx) (platform.UpdateRateLimitRuleRequest, error) {
	var raw map[string]interface{}
	if err := c.BodyParser(&raw); err != nil {
		return platform.UpdateRateLimitRuleRequest{}, err
	}
	effectiveFrom, err := parseOptionalTimeValue(raw["effective_from"])
	if err != nil {
		return platform.UpdateRateLimitRuleRequest{}, err
	}
	expiresAt, err := parseOptionalTimeValue(raw["expires_at"])
	if err != nil {
		return platform.UpdateRateLimitRuleRequest{}, err
	}
	raw["effective_from"] = nil
	raw["expires_at"] = nil

	encoded, err := json.Marshal(raw)
	if err != nil {
		return platform.UpdateRateLimitRuleRequest{}, err
	}
	var req platform.UpdateRateLimitRuleRequest
	if err := json.Unmarshal(encoded, &req); err != nil {
		return platform.UpdateRateLimitRuleRequest{}, err
	}
	req.EffectiveFrom = effectiveFrom
	req.ExpiresAt = expiresAt
	return req, nil
}

func (h *Handlers) ListBudgetLimitRules(c *fiber.Ctx) error {
	if err := requireTenantAdminOrHigher(c); err != nil {
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
	items, total, err := h.models.ListBudgetLimitRules(c.Context(), currentTenant.ID, query, limit, offset)
	if err != nil {
		return err
	}
	var nextOffset *int
	if offset+len(items) < total {
		n := offset + len(items)
		nextOffset = &n
	}
	return c.JSON(platform.BudgetLimitRuleListResponse{
		Data:       items,
		Total:      total,
		Limit:      limit,
		Offset:     offset,
		NextOffset: nextOffset,
	})
}

func (h *Handlers) CreateBudgetLimitRule(c *fiber.Ctx) error {
	if err := requireTenantAdminOrHigher(c); err != nil {
		return err
	}
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	var req platform.CreateBudgetLimitRuleRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if err := validateCreateBudgetLimitRuleRequest(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	item, err := h.models.CreateBudgetLimitRule(c.Context(), currentTenant.ID, req)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusUnprocessableEntity, "failed to create budget limit rule")
		}
		return mapProviderAccountStoreError(err)
	}
	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *Handlers) GetBudgetLimitRule(c *fiber.Ctx) error {
	if err := requireTenantAdminOrHigher(c); err != nil {
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
	item, err := h.models.GetBudgetLimitRuleByRuleID(c.Context(), currentTenant.ID, ruleID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "budget limit rule not found")
		}
		return err
	}
	return c.JSON(item)
}

func (h *Handlers) UpdateBudgetLimitRule(c *fiber.Ctx) error {
	if err := requireTenantAdminOrHigher(c); err != nil {
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
	item, err := h.models.UpdateBudgetLimitRuleByRuleID(c.Context(), currentTenant.ID, ruleID, req)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "budget limit rule not found")
		}
		return mapProviderAccountStoreError(err)
	}
	return c.JSON(item)
}

func (h *Handlers) DeleteBudgetLimitRule(c *fiber.Ctx) error {
	if err := requireTenantAdminOrHigher(c); err != nil {
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
	if err := h.models.DeleteBudgetLimitRuleByRuleID(c.Context(), currentTenant.ID, ruleID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "budget limit rule not found")
		}
		return err
	}
	return c.JSON(fiber.Map{"success": true})
}

func (h *Handlers) UpdateBudgetLimitRuleStatus(c *fiber.Ctx) error {
	if err := requireTenantAdminOrHigher(c); err != nil {
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
	var req platform.UpdateBudgetLimitRuleStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	item, err := h.models.UpdateBudgetLimitRuleStatusByRuleID(c.Context(), currentTenant.ID, ruleID, req.Enabled)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "budget limit rule not found")
		}
		return err
	}
	return c.JSON(item)
}

func validateCreateBudgetLimitRuleRequest(req platform.CreateBudgetLimitRuleRequest) error {
	if strings.TrimSpace(req.RuleID) == "" {
		return errors.New("rule_id is required")
	}
	if req.LimitTo <= 0 {
		return errors.New("limit_to must be greater than zero")
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

func validateBudgetAppliesPer(value string) error {
	normalized := strings.TrimSpace(value)
	switch normalized {
	case string(platform.BudgetAppliesPerUser), string(platform.BudgetAppliesPerModel), string(platform.BudgetAppliesPerVirtualAccount):
		return nil
	}
	if strings.HasPrefix(normalized, "metadata.") && len(strings.TrimPrefix(normalized, "metadata.")) > 0 {
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
	if err := requireTenantAdminOrHigher(c); err != nil {
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
	if err := requireTenantAdminOrHigher(c); err != nil {
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
	if err := requireTenantAdminOrHigher(c); err != nil {
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
	if err := requireTenantAdminOrHigher(c); err != nil {
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
	if err := requireTenantAdminOrHigher(c); err != nil {
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
	if err := requireTenantAdminOrHigher(c); err != nil {
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
	if err := requireTenantAdminOrHigher(c); err != nil {
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
	if err := requireTenantAdminOrHigher(c); err != nil {
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
	if err := requireTenantAdminOrHigher(c); err != nil {
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
	if err := requireTenantAdminOrHigher(c); err != nil {
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
	if err := requireTenantAdminOrHigher(c); err != nil {
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
	if err := requireTeamAdminOrHigher(c); err != nil {
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
	if err := requireTeamAdminOrHigher(c); err != nil {
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

	kind := strings.TrimSpace(strings.ToLower(c.Query("kind")))
	switch kind {
	case "user", "team", "org_unit":
	default:
		return fiber.NewError(fiber.StatusBadRequest, "kind must be user, team or org_unit")
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

// ListTenantUsers lists tenant-admin users for the current tenant.
func (h *Handlers) ListTenantUsers(c *fiber.Ctx) error {
	if err := requireTenantAdminOrHigher(c); err != nil {
		return err
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	if h.users == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "user store not configured")
	}

	query := strings.TrimSpace(c.Query("query"))
	limit := 50
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

	items, total, err := h.users.ListTenantAdminUsers(c.Context(), currentTenant.Slug, query, limit, offset)
	if err != nil {
		return err
	}

	var nextOffset *int
	if offset+len(items) < total {
		n := offset + len(items)
		nextOffset = &n
	}

	return c.JSON(platform.TenantAdminUserListResponse{
		Data:       items,
		Total:      total,
		Limit:      limit,
		Offset:     offset,
		NextOffset: nextOffset,
	})
}

// InviteTenantUser creates a tenant-admin invitation link.
func (h *Handlers) InviteTenantUser(c *fiber.Ctx) error {
	if err := requireTenantAdminOrHigher(c); err != nil {
		return err
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	if h.users == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "user store not configured")
	}

	var req platform.InviteTenantAdminUserRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Mobile = strings.TrimSpace(req.Mobile)
	req.Role = strings.TrimSpace(strings.ToLower(req.Role))
	req.AcceptInviteClientURL = strings.TrimSpace(req.AcceptInviteClientURL)

	if req.Email == "" && req.Mobile == "" {
		return fiber.NewError(fiber.StatusBadRequest, "either email or mobile is required")
	}
	if req.Email != "" && !storage.IsValidEmailAddress(req.Email) {
		return fiber.NewError(fiber.StatusBadRequest, "Please enter a valid email address.")
	}
	if req.Role == "" {
		req.Role = "tenant_member"
	}

	invitation, link, err := h.users.InviteTenantAdminUser(c.Context(), currentTenant.ID, currentTenant.Slug, req)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrInvalidInvitationEmail):
			return fiber.NewError(fiber.StatusBadRequest, "Please enter a valid email address.")
		case errors.Is(err, storage.ErrTenantMemberAlreadyExists):
			return fiber.NewError(fiber.StatusConflict, "This user is already a member of this tenant.")
		case errors.Is(err, storage.ErrTenantMemberDeactivated):
			return fiber.NewError(fiber.StatusConflict, "This user is deactivated and cannot receive an invitation.")
		case errors.Is(err, storage.ErrInvitationAlreadyPending):
			return fiber.NewError(fiber.StatusConflict, "An invitation has already been sent to this email.")
		default:
			return err
		}
	}
	return c.Status(fiber.StatusCreated).JSON(platform.InviteTenantAdminUserResponse{
		Invitation: *invitation,
		Link:       link,
	})
}

// AcceptTenantUserInvitation accepts invite token and activates/creates user.
func (h *Handlers) AcceptTenantUserInvitation(c *fiber.Ctx) error {
	if err := requireTenantAdminOrHigher(c); err != nil {
		return err
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	if h.users == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "user store not configured")
	}

	var req platform.AcceptTenantAdminInvitationRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	req.Token = strings.TrimSpace(req.Token)
	if req.Token == "" {
		return fiber.NewError(fiber.StatusBadRequest, "token is required")
	}

	invitation, user, err := h.users.AcceptTenantAdminInvitation(c.Context(), currentTenant.ID, currentTenant.Slug, req.Token)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrInvalidInvitationToken):
			return fiber.NewError(fiber.StatusBadRequest, "invalid invitation token")
		case errors.Is(err, storage.ErrInvitationExpired):
			return fiber.NewError(fiber.StatusGone, "invitation expired")
		case errors.Is(err, storage.ErrInvitationAlreadyUsed):
			return fiber.NewError(fiber.StatusConflict, "invitation already used")
		case errors.Is(err, storage.ErrInvitationContactTaken):
			return fiber.NewError(fiber.StatusConflict, "invitation contact already belongs to another account")
		case errors.Is(err, pgx.ErrNoRows):
			return fiber.NewError(fiber.StatusNotFound, "invitation not found")
		default:
			return err
		}
	}

	return c.JSON(platform.AcceptTenantAdminInvitationResponse{
		Invitation: *invitation,
		User:       *user,
	})
}

// ActivateTenantUser marks one tenant-admin user as active.
func (h *Handlers) ActivateTenantUser(c *fiber.Ctx) error {
	return h.setTenantUserActive(c, true)
}

// DeactivateTenantUser marks one tenant-admin user as deactivated.
func (h *Handlers) DeactivateTenantUser(c *fiber.Ctx) error {
	return h.setTenantUserActive(c, false)
}

func (h *Handlers) setTenantUserActive(c *fiber.Ctx, active bool) error {
	if err := requireTenantAdminOrHigher(c); err != nil {
		return err
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	if h.users == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "user store not configured")
	}

	userID := strings.TrimSpace(c.Params("userID"))
	if userID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "userID is required")
	}
	user, err := h.users.SetTenantAdminUserActive(c.Context(), currentTenant.Slug, userID, active)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "user not found")
		}
		return err
	}
	return c.JSON(user)
}

func (h *Handlers) requireProviderAccountsStore() (providerAccountStore, error) {
	if h.providerAccounts == nil {
		return nil, fiber.NewError(fiber.StatusNotImplemented, "provider account store not configured")
	}
	return h.providerAccounts, nil
}

func (h *Handlers) requireOnboardingStore() (onboardingStore, error) {
	if h.onboarding == nil {
		return nil, fiber.NewError(fiber.StatusNotImplemented, "onboarding store not configured")
	}
	return h.onboarding, nil
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

func (h *Handlers) ListOrgUnits(c *fiber.Ctx) error {
	if err := requireOrgAdminOrHigher(c); err != nil {
		return err
	}
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	if h.orgUnits == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "org unit store not configured")
	}

	query := strings.TrimSpace(c.Query("query"))
	limit := 50
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

	items, total, err := h.orgUnits.ListOrgUnits(c.Context(), currentTenant.ID, query, limit, offset)
	if err != nil {
		return err
	}
	var nextOffset *int
	if offset+len(items) < total {
		n := offset + len(items)
		nextOffset = &n
	}
	return c.JSON(platform.OrgUnitListResponse{Data: items, Total: total, Limit: limit, Offset: offset, NextOffset: nextOffset})
}

func (h *Handlers) ListTeams(c *fiber.Ctx) error {
	if err := requireOrgAdminOrHigher(c); err != nil {
		return err
	}
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	if h.teams == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "team store not configured")
	}
	query := strings.TrimSpace(c.Query("query"))
	limit := 50
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
	items, total, err := h.teams.ListTeams(c.Context(), currentTenant.ID, query, limit, offset)
	if err != nil {
		return err
	}
	var nextOffset *int
	if offset+len(items) < total {
		n := offset + len(items)
		nextOffset = &n
	}
	return c.JSON(platform.TeamListResponse{Data: items, Total: total, Limit: limit, Offset: offset, NextOffset: nextOffset})
}

func (h *Handlers) CreateTeam(c *fiber.Ctx) error {
	if err := requireOrgAdminOrHigher(c); err != nil {
		return err
	}
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	if h.teams == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "team store not configured")
	}
	var req platform.CreateTeamRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	req.Alias = strings.TrimSpace(req.Alias)
	if req.Alias == "" {
		return fiber.NewError(fiber.StatusBadRequest, "alias is required")
	}
	item, err := h.teams.CreateTeam(c.Context(), currentTenant.ID, req)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *Handlers) GetTeam(c *fiber.Ctx) error {
	if err := requireOrgAdminOrHigher(c); err != nil {
		return err
	}
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	if h.teams == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "team store not configured")
	}
	teamID := strings.TrimSpace(c.Params("teamID"))
	if teamID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "teamID is required")
	}
	item, err := h.teams.GetTeam(c.Context(), currentTenant.ID, teamID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "team not found")
		}
		return err
	}
	return c.JSON(item)
}

func (h *Handlers) UpdateTeam(c *fiber.Ctx) error {
	if err := requireTeamAdminOrHigher(c); err != nil {
		return err
	}
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	if h.teams == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "team store not configured")
	}
	teamID := strings.TrimSpace(c.Params("teamID"))
	if teamID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "teamID is required")
	}
	var req platform.UpdateTeamRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	item, err := h.teams.UpdateTeam(c.Context(), currentTenant.ID, teamID, req)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "team not found")
		}
		return err
	}
	return c.JSON(item)
}

func (h *Handlers) DeleteTeam(c *fiber.Ctx) error {
	if err := requireOrgAdminOrHigher(c); err != nil {
		return err
	}
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	if h.teams == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "team store not configured")
	}
	teamID := strings.TrimSpace(c.Params("teamID"))
	if teamID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "teamID is required")
	}
	if err := h.teams.DeleteTeam(c.Context(), currentTenant.ID, teamID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "team not found")
		}
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handlers) GetTeamPermissions(c *fiber.Ctx) error {
	if err := requireTeamAdminOrHigher(c); err != nil {
		return err
	}
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	if h.teams == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "team store not configured")
	}
	teamID := strings.TrimSpace(c.Params("teamID"))
	if teamID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "teamID is required")
	}
	item, err := h.teams.GetTeamPermissions(c.Context(), currentTenant.ID, teamID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "team not found")
		}
		return err
	}
	return c.JSON(item)
}

func (h *Handlers) UpdateTeamPermissions(c *fiber.Ctx) error {
	if err := requireTeamAdminOrHigher(c); err != nil {
		return err
	}
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	if h.teams == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "team store not configured")
	}
	teamID := strings.TrimSpace(c.Params("teamID"))
	if teamID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "teamID is required")
	}
	var req platform.UpdateTeamPermissionsRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	item, err := h.teams.UpdateTeamPermissions(c.Context(), currentTenant.ID, teamID, req)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "team not found")
		}
		return err
	}
	return c.JSON(item)
}

func (h *Handlers) CreateOrgUnit(c *fiber.Ctx) error {
	if err := requireOrgAdminOrHigher(c); err != nil {
		return err
	}
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	if h.orgUnits == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "org unit store not configured")
	}
	var req platform.CreateOrgUnitRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	req.Alias = strings.TrimSpace(req.Alias)
	req.DisplayName = strings.TrimSpace(req.DisplayName)
	if req.Alias == "" {
		return fiber.NewError(fiber.StatusBadRequest, "alias is required")
	}
	item, err := h.orgUnits.CreateOrgUnit(c.Context(), currentTenant.ID, req)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *Handlers) GetOrgUnit(c *fiber.Ctx) error {
	if err := requireOrgAdminOrHigher(c); err != nil {
		return err
	}
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	if h.orgUnits == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "org unit store not configured")
	}
	orgUnitID := strings.TrimSpace(c.Params("orgUnitID"))
	if orgUnitID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "orgUnitID is required")
	}
	item, err := h.orgUnits.GetOrgUnit(c.Context(), currentTenant.ID, orgUnitID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "org unit not found")
		}
		return err
	}
	return c.JSON(item)
}

func (h *Handlers) UpdateOrgUnit(c *fiber.Ctx) error {
	if err := requireOrgAdminOrHigher(c); err != nil {
		return err
	}
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	if h.orgUnits == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "org unit store not configured")
	}
	orgUnitID := strings.TrimSpace(c.Params("orgUnitID"))
	if orgUnitID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "orgUnitID is required")
	}
	var req platform.UpdateOrgUnitRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	item, err := h.orgUnits.UpdateOrgUnit(c.Context(), currentTenant.ID, orgUnitID, req)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "org unit not found")
		}
		return err
	}
	return c.JSON(item)
}

func (h *Handlers) DeleteOrgUnit(c *fiber.Ctx) error {
	if err := requireOrgAdminOrHigher(c); err != nil {
		return err
	}
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	if h.orgUnits == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "org unit store not configured")
	}
	orgUnitID := strings.TrimSpace(c.Params("orgUnitID"))
	if orgUnitID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "orgUnitID is required")
	}
	if err := h.orgUnits.DeleteOrgUnit(c.Context(), currentTenant.ID, orgUnitID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "org unit not found")
		}
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handlers) ListOrgUnitMembers(c *fiber.Ctx) error {
	if err := requireOrgAdminOrHigher(c); err != nil {
		return err
	}
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	if h.orgUnits == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "org unit store not configured")
	}
	orgUnitID := strings.TrimSpace(c.Params("orgUnitID"))
	if orgUnitID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "orgUnitID is required")
	}
	items, err := h.orgUnits.ListOrgUnitMemberships(c.Context(), currentTenant.ID, orgUnitID)
	if err != nil {
		return err
	}
	return c.JSON(platform.OrgUnitMembershipListResponse{Data: items})
}

func (h *Handlers) AddOrgUnitMember(c *fiber.Ctx) error {
	if err := requireOrgAdminOrHigher(c); err != nil {
		return err
	}
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	if h.orgUnits == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "org unit store not configured")
	}
	orgUnitID := strings.TrimSpace(c.Params("orgUnitID"))
	if orgUnitID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "orgUnitID is required")
	}
	var req platform.CreateOrgUnitMembershipRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	req.UserID = strings.TrimSpace(req.UserID)
	req.Role = strings.TrimSpace(req.Role)
	if req.UserID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "user_id is required")
	}
	if req.Role == "" {
		req.Role = "member"
	}
	item, err := h.orgUnits.AddOrgUnitMembership(c.Context(), currentTenant.ID, orgUnitID, req)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "user not found for this tenant")
		}
		if strings.Contains(strings.ToLower(err.Error()), "invalid uuid") {
			return fiber.NewError(fiber.StatusBadRequest, "user_id must be a UUID platform user id")
		}
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *Handlers) UpdateOrgUnitMember(c *fiber.Ctx) error {
	if err := requireOrgAdminOrHigher(c); err != nil {
		return err
	}
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	if h.orgUnits == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "org unit store not configured")
	}
	orgUnitID := strings.TrimSpace(c.Params("orgUnitID"))
	userID := strings.TrimSpace(c.Params("userID"))
	if orgUnitID == "" || userID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "orgUnitID and userID are required")
	}
	var req platform.UpdateOrgUnitMembershipRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	req.Role = strings.TrimSpace(req.Role)
	if req.Role == "" {
		return fiber.NewError(fiber.StatusBadRequest, "role is required")
	}
	item, err := h.orgUnits.UpdateOrgUnitMembership(c.Context(), currentTenant.ID, orgUnitID, userID, req)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "org unit membership not found")
		}
		return err
	}
	return c.JSON(item)
}

func (h *Handlers) DeleteOrgUnitMember(c *fiber.Ctx) error {
	if err := requireOrgAdminOrHigher(c); err != nil {
		return err
	}
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	if h.orgUnits == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "org unit store not configured")
	}
	orgUnitID := strings.TrimSpace(c.Params("orgUnitID"))
	userID := strings.TrimSpace(c.Params("userID"))
	if orgUnitID == "" || userID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "orgUnitID and userID are required")
	}
	if err := h.orgUnits.DeleteOrgUnitMembership(c.Context(), currentTenant.ID, orgUnitID, userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "org unit membership not found")
		}
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handlers) ListTeamMembers(c *fiber.Ctx) error {
	if err := requireTeamAdminOrHigher(c); err != nil {
		return err
	}
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	if h.memberships == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "membership store not configured")
	}
	teamID := strings.TrimSpace(c.Params("teamID"))
	if teamID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "teamID is required")
	}
	items, err := h.memberships.ListTeamMemberships(c.Context(), currentTenant.ID, teamID)
	if err != nil {
		return err
	}
	return c.JSON(platform.TeamMembershipListResponse{Data: items})
}

func (h *Handlers) AddTeamMember(c *fiber.Ctx) error {
	if err := requireTeamAdminOrHigher(c); err != nil {
		return err
	}
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	if h.memberships == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "membership store not configured")
	}
	teamID := strings.TrimSpace(c.Params("teamID"))
	if teamID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "teamID is required")
	}
	var req platform.CreateTeamMembershipRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	req.UserID = strings.TrimSpace(req.UserID)
	req.Role = strings.TrimSpace(req.Role)
	if req.UserID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "user_id is required")
	}
	if req.Role == "" {
		req.Role = "team_member"
	}
	item, err := h.memberships.AddTeamMembership(c.Context(), currentTenant.ID, teamID, req)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *Handlers) UpdateTeamMember(c *fiber.Ctx) error {
	if err := requireTeamAdminOrHigher(c); err != nil {
		return err
	}
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	if h.memberships == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "membership store not configured")
	}
	teamID := strings.TrimSpace(c.Params("teamID"))
	userID := strings.TrimSpace(c.Params("userID"))
	if teamID == "" || userID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "teamID and userID are required")
	}
	var req platform.UpdateTeamMembershipRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	req.Role = strings.TrimSpace(req.Role)
	item, err := h.memberships.UpdateTeamMembership(c.Context(), currentTenant.ID, teamID, userID, req)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "team membership not found")
		}
		return err
	}
	return c.JSON(item)
}

func (h *Handlers) DeleteTeamMember(c *fiber.Ctx) error {
	if err := requireTeamAdminOrHigher(c); err != nil {
		return err
	}
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	if h.memberships == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "membership store not configured")
	}
	teamID := strings.TrimSpace(c.Params("teamID"))
	userID := strings.TrimSpace(c.Params("userID"))
	if teamID == "" || userID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "teamID and userID are required")
	}
	if err := h.memberships.DeleteTeamMembership(c.Context(), currentTenant.ID, teamID, userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "team membership not found")
		}
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handlers) ListVirtualUsers(c *fiber.Ctx) error {
	if err := requireTeamAdminOrHigher(c); err != nil {
		return err
	}
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	if h.virtualUsers == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "virtual user store not configured")
	}
	query := strings.TrimSpace(c.Query("query"))
	limit := 50
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
	items, total, err := h.virtualUsers.ListVirtualUsers(c.Context(), currentTenant.ID, query, limit, offset)
	if err != nil {
		return err
	}
	var nextOffset *int
	if offset+len(items) < total {
		n := offset + len(items)
		nextOffset = &n
	}
	return c.JSON(platform.VirtualUserListResponse{Data: items, Total: total, Limit: limit, Offset: offset, NextOffset: nextOffset})
}

func (h *Handlers) CreateVirtualUser(c *fiber.Ctx) error {
	if err := requireTeamAdminOrHigher(c); err != nil {
		return err
	}
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	if h.virtualUsers == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "virtual user store not configured")
	}
	var req platform.CreateVirtualUserRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	req.Alias = strings.TrimSpace(req.Alias)
	req.DisplayName = strings.TrimSpace(req.DisplayName)
	req.TeamID = strings.TrimSpace(req.TeamID)
	req.OrgUnitID = strings.TrimSpace(req.OrgUnitID)
	if req.Alias == "" {
		return fiber.NewError(fiber.StatusBadRequest, "alias is required")
	}
	item, err := h.virtualUsers.CreateVirtualUser(c.Context(), currentTenant.ID, req)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *Handlers) GetVirtualUser(c *fiber.Ctx) error {
	if err := requireTeamAdminOrHigher(c); err != nil {
		return err
	}
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	if h.virtualUsers == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "virtual user store not configured")
	}
	virtualUserID := strings.TrimSpace(c.Params("virtualUserID"))
	if virtualUserID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "virtualUserID is required")
	}
	item, err := h.virtualUsers.GetVirtualUser(c.Context(), currentTenant.ID, virtualUserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "virtual user not found")
		}
		return err
	}
	return c.JSON(item)
}

func (h *Handlers) UpdateVirtualUser(c *fiber.Ctx) error {
	if err := requireTeamAdminOrHigher(c); err != nil {
		return err
	}
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	if h.virtualUsers == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "virtual user store not configured")
	}
	virtualUserID := strings.TrimSpace(c.Params("virtualUserID"))
	if virtualUserID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "virtualUserID is required")
	}
	var req platform.UpdateVirtualUserRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	item, err := h.virtualUsers.UpdateVirtualUser(c.Context(), currentTenant.ID, virtualUserID, req)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "virtual user not found")
		}
		return err
	}
	return c.JSON(item)
}

func (h *Handlers) DeleteVirtualUser(c *fiber.Ctx) error {
	if err := requireTeamAdminOrHigher(c); err != nil {
		return err
	}
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	if h.virtualUsers == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "virtual user store not configured")
	}
	virtualUserID := strings.TrimSpace(c.Params("virtualUserID"))
	if virtualUserID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "virtualUserID is required")
	}
	if err := h.virtualUsers.DeleteVirtualUser(c.Context(), currentTenant.ID, virtualUserID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "virtual user not found")
		}
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func requireTenantAdminOrHigher(c *fiber.Ctx) error {
	authInfo := auth.GetAuthInfo(c)
	if auth.IsTenantAdminOrHigher(authInfo) {
		return nil
	}
	return fiber.NewError(fiber.StatusForbidden, "tenant admin privileges required")
}

func requireTeamAdminOrHigher(c *fiber.Ctx) error {
	authInfo := auth.GetAuthInfo(c)
	if auth.IsTeamAdminOrHigher(authInfo) {
		return nil
	}
	return fiber.NewError(fiber.StatusForbidden, "team admin privileges required")
}

// requireOrgAdminOrHigher checks if the caller is org_admin or higher role.
func requireOrgAdminOrHigher(c *fiber.Ctx) error {
	authInfo := auth.GetAuthInfo(c)
	if auth.IsOrgAdminOrHigher(authInfo) {
		return nil
	}
	return fiber.NewError(fiber.StatusForbidden, "organization admin privileges required")
}

func (h *Handlers) requireTenantAdminOrProviderAccountManager(c *fiber.Ctx, store providerAccountStore, tenantID string, accountID string) error {
	authInfo := auth.GetAuthInfo(c)
	if auth.IsTenantAdminOrHigher(authInfo) {
		return nil
	}
	if authInfo == nil {
		return fiber.NewError(fiber.StatusForbidden, "insufficient permissions to delete provider account")
	}
	allowed, err := store.IsTenantProviderAccountManager(c.Context(), tenantID, accountID, authInfo.UserID, authInfo.TeamID)
	if err != nil {
		return err
	}
	if !allowed {
		return fiber.NewError(fiber.StatusForbidden, "insufficient permissions to delete provider account")
	}
	return nil
}

func mapProviderAccountStoreError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, storage.ErrInvalidAccessPrincipal) {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if errors.Is(err, storage.ErrProviderAccountDependencyConflict) {
		return fiber.NewError(fiber.StatusConflict, "provider account is referenced by active keys; revoke or update keys first")
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return fiber.NewError(fiber.StatusConflict, "resource already exists")
		case "23514", "23502", "22P02":
			return fiber.NewError(fiber.StatusBadRequest, "invalid provider account payload")
		}
	}
	return err
}

func normalizeAndValidateModelFQNs(models []string) ([]string, error) {
	normalized := make([]string, 0, len(models))
	seen := make(map[string]struct{}, len(models))
	for _, model := range models {
		model = strings.TrimSpace(model)
		if model == "" {
			continue
		}
		if !isModelFQN(model) {
			return nil, errors.New("models must use provider:account_alias:official_model_id format")
		}
		if _, ok := seen[model]; ok {
			continue
		}
		seen[model] = struct{}{}
		normalized = append(normalized, model)
	}
	if len(normalized) == 0 {
		return nil, errors.New("at least one allowed model is required")
	}
	return normalized, nil
}

func normalizeRequestedModelIDs(models []string) ([]string, error) {
	normalized := make([]string, 0, len(models))
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
		normalized = append(normalized, model)
	}
	if len(normalized) == 0 {
		return nil, errors.New("at least one allowed model is required")
	}
	return normalized, nil
}

func splitRequestedModels(models []string) ([]string, []string) {
	fqns := make([]string, 0, len(models))
	aliases := make([]string, 0, len(models))
	for _, model := range models {
		if isModelFQN(model) {
			fqns = append(fqns, model)
			continue
		}
		aliases = append(aliases, model)
	}
	return fqns, aliases
}

func isVirtualModelInScope(vm platform.VirtualModel, scopeAllowedSet map[string]struct{}) bool {
	if len(vm.Targets) == 0 {
		return false
	}
	for _, target := range vm.Targets {
		targetFQN := strings.TrimSpace(target.Model)
		if targetFQN == "" {
			return false
		}
		if _, allowed := scopeAllowedSet[targetFQN]; !allowed {
			return false
		}
	}
	return true
}

func buildVirtualModelSnapshotsForAliases(aliases []string, virtualModels []platform.VirtualModel, scopeAllowedSet map[string]struct{}) (map[string]platform.VirtualModelSnapshot, error) {
	if len(aliases) == 0 {
		return map[string]platform.VirtualModelSnapshot{}, nil
	}
	byAlias := make(map[string]platform.VirtualModel, len(virtualModels))
	for _, vm := range virtualModels {
		alias := strings.TrimSpace(vm.Alias)
		if alias == "" {
			continue
		}
		byAlias[alias] = vm
	}
	out := make(map[string]platform.VirtualModelSnapshot, len(aliases))
	for _, alias := range aliases {
		alias = strings.TrimSpace(alias)
		vm, ok := byAlias[alias]
		if !ok {
			return nil, errors.New("one or more requested models are not allowed in current tenant scope")
		}
		if !isVirtualModelInScope(vm, scopeAllowedSet) {
			return nil, errors.New("one or more requested virtual models include targets outside your provider scope")
		}
		targetFQNs := make([]string, 0, len(vm.Targets))
		seenTargets := make(map[string]struct{}, len(vm.Targets))
		for _, target := range vm.Targets {
			fqn := strings.TrimSpace(target.Model)
			if fqn == "" {
				continue
			}
			if _, exists := seenTargets[fqn]; exists {
				continue
			}
			seenTargets[fqn] = struct{}{}
			targetFQNs = append(targetFQNs, fqn)
		}
		if len(targetFQNs) == 0 {
			return nil, errors.New("one or more requested virtual models do not have eligible targets")
		}
		out[alias] = platform.VirtualModelSnapshot{
			Alias:          alias,
			VirtualModelID: vm.ID,
			Version:        vm.Version,
			TargetFQNs:     targetFQNs,
		}
	}
	return out, nil
}

func isModelFQN(model string) bool {
	parts := strings.Split(model, ":")
	if len(parts) != 3 {
		return false
	}
	for _, part := range parts {
		if strings.TrimSpace(part) == "" {
			return false
		}
	}
	return true
}

func ensureRequestedModelsInScope(requested []string, allowed []string) error {
	if len(allowed) == 0 {
		return errors.New("no allowed models found in current tenant scope for this principal")
	}
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, item := range allowed {
		allowedSet[strings.TrimSpace(item)] = struct{}{}
	}
	invalid := make([]string, 0)
	for _, model := range requested {
		if _, ok := allowedSet[strings.TrimSpace(model)]; !ok {
			invalid = append(invalid, model)
		}
	}
	if len(invalid) > 0 {
		return errors.New("one or more requested models are not allowed in current tenant scope")
	}
	return nil
}
