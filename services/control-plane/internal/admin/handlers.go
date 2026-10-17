package admin

import (
	"context"
	"mime/multipart"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	billingdomain "github.com/Delpak/shahrah/services/control-plane/internal/domain/billing"
	"github.com/Delpak/shahrah/services/control-plane/internal/platformcredentials"
	"github.com/Delpak/shahrah/services/control-plane/internal/supportfiles"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenantadmin"
	"github.com/gofiber/fiber/v2"
)

type keyStore interface {
	CreateKey(ctx context.Context, params platform.CreateKeyParams) (*platform.CreatedKey, error)
	ListKeys(ctx context.Context, tenantID string) ([]platform.CreatedKey, error)
	ListAllowedModelFQNs(ctx context.Context, tenantID string, userID string, teamID string) ([]string, error)
	ListPATs(ctx context.Context, tenantID string, userID string, query string, limit int, offset int) ([]platform.CreatedKey, int, error)
	PATExists(ctx context.Context, tenantID string, userID string, name string) (bool, error)
	RevokePAT(ctx context.Context, tenantID string, userID string, keyID string, revokedByUserID string) (bool, error)
	RevokeAllPATs(ctx context.Context, tenantID string, userID string, revokedByUserID string) (int64, error)
	ListVATTokens(ctx context.Context, tenantID string, virtualUserID string, query string, limit int, offset int) ([]platform.CreatedKey, int, error)
	UpdateVATToken(ctx context.Context, tenantID string, virtualUserID string, keyID string, req platform.UpdateVATTokenRequest) (*platform.CreatedKey, error)
	RevokeVATToken(ctx context.Context, tenantID string, virtualUserID string, keyID string, revokedByUserID string, reason string) (bool, error)
	RotateVATToken(ctx context.Context, tenantID string, virtualUserID string, keyID string, rotatedByUserID string, gracePeriod string) (*platform.CreatedKey, error)
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
}

type policyStore interface {
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
	CreateTenantGuardrail(ctx context.Context, tenantID string, req platform.CreateTenantGuardrailRequest, actor string) (*platform.TenantGuardrail, error)
	ListTenantGuardrails(ctx context.Context, tenantID string, query string, limit int, offset int) ([]platform.TenantGuardrail, int, error)
	GetTenantGuardrailBySlug(ctx context.Context, tenantID string, slug string) (*platform.TenantGuardrail, error)
	UpdateTenantGuardrailBySlug(ctx context.Context, tenantID string, slug string, req platform.UpdateTenantGuardrailRequest, actor string) (*platform.TenantGuardrail, error)
	DeleteTenantGuardrailBySlug(ctx context.Context, tenantID string, slug string) error
}

type walletStore interface {
	GetWallet(ctx context.Context, tenantID string) (*platform.Wallet, error)
	TopUpWallet(ctx context.Context, tenantID string, amountIRR float64, metadata map[string]interface{}) (*platform.Transaction, error)
	GetLatestExchangeRate(ctx context.Context) (*platform.ExchangeRate, error)
	GetWalletTopUpSettings(ctx context.Context) (*platform.WalletTopUpSettings, error)
	UpdateWalletTopUpSettings(ctx context.Context, minUSD, maxUSD, taxPercent, feePercent float64) (*platform.WalletTopUpSettings, error)
	CreateWalletTopUpIntent(ctx context.Context, p billingdomain.CreateTopUpIntentParams) (*platform.WalletTopUpIntent, error)
	AttachWalletTopUpGatewayData(ctx context.Context, intentPublicID, authority, paymentURL string, raw map[string]interface{}) error
	GetWalletTopUpIntentByAuthority(ctx context.Context, authority string) (*platform.WalletTopUpIntent, error)
	CompleteWalletTopUpByAuthority(ctx context.Context, authority string, verifyRaw map[string]interface{}, refID, gatewayTxnID, cardMask, payerPhone string) (*platform.WalletTopUpIntent, bool, error)
	MarkWalletTopUpFailedByAuthority(ctx context.Context, authority string, reason string, raw map[string]interface{}) error
	ListWalletTransactions(ctx context.Context, tenantID string, limit int) ([]platform.Transaction, error)
	CreateBankTransferRequest(ctx context.Context, tenantID string, req platform.CreateBankTransferRequest, actor string) (*platform.BankTransferRequest, error)
	ListBankTransferRequests(ctx context.Context, tenantID string, limit int) ([]platform.BankTransferRequest, error)
	BankTransferReceiptExists(ctx context.Context, tenantID string, receiptFileURL string) (bool, error)
	ApproveBankTransferRequest(ctx context.Context, tenantID string, requestID string, req platform.ReviewBankTransferRequest, actor platform.SupportActor) (*platform.BankTransferRequest, error)
	RejectBankTransferRequest(ctx context.Context, tenantID string, requestID string, req platform.ReviewBankTransferRequest, actor platform.SupportActor) (*platform.BankTransferRequest, error)
	RedeemGiftCode(ctx context.Context, tenantID string, code string, actor string) (*platform.GiftCodeRedemption, error)
	RedeemDiscountCode(ctx context.Context, tenantID string, code string, actor string) (*platform.DiscountRedemption, error)
	GetBillingProfile(ctx context.Context, tenantID string) (*platform.BillingProfile, error)
	UpsertBillingProfile(ctx context.Context, tenantID string, req platform.UpsertBillingProfileRequest, actor string) (*platform.BillingProfile, error)
	ListInvoices(ctx context.Context, tenantID string, limit int) ([]platform.Invoice, error)
	ListDiscountRedemptions(ctx context.Context, tenantID string, limit int) ([]platform.DiscountRedemption, error)
	GetBillingUsageSummary(ctx context.Context, tenantID string, preset string) (*platform.BillingUsageSummary, error)
	GetMonthlyUsageDetail(ctx context.Context, tenantID string, jalaliYear int, jalaliMonth int) (*platform.MonthlyUsageDetail, error)
}

type providerAccountStore interface {
	CreateTenantProviderAccount(ctx context.Context, tenantID string, req platform.CreateTenantProviderAccountRequest) (*platform.TenantProviderAccount, error)
	ListTenantProviderAccounts(ctx context.Context, tenantID string) ([]platform.TenantProviderAccount, error)
	ListTenantProviderAccountCatalogRecords(ctx context.Context, tenantID string) ([]platform.TenantProviderAccountCatalogRecord, error)
	GetTenantProviderAccountDetails(ctx context.Context, tenantID string, accountID string) (*platform.TenantProviderAccountDetailsResponse, error)
	UpdateTenantProviderAccountCredentials(ctx context.Context, tenantID string, accountID string, req platform.UpdateTenantProviderAccountCredentialsRequest) (*platform.TenantProviderAccount, error)
	UpsertTenantProviderAccountModels(ctx context.Context, tenantID string, accountID string, models []platform.TenantProviderAccountModel) ([]platform.TenantProviderAccountModel, error)
	UpdateTenantProviderAccountAccess(ctx context.Context, tenantID string, accountID string, access platform.TenantProviderAccountAccessControl) (*platform.TenantProviderAccountAccessControl, error)
	TestTenantProviderAccountCredentials(ctx context.Context, tenantID string, accountID string, req platform.TestTenantProviderAccountCredentialsRequest) (*platform.TestTenantProviderAccountCredentialsResponse, error)
	TestProviderConnection(ctx context.Context, req platform.TestProviderConnectionRequest) (*platform.TestTenantProviderAccountCredentialsResponse, error)
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
	GetTenantAdminUserRecord(ctx context.Context, tenantSlug string, userID string) (*platform.TenantAdminUser, error)
	InviteTenantAdminUser(ctx context.Context, tenantID string, tenantSlug string, req platform.InviteTenantAdminUserRequest) (*platform.TenantAdminInvitation, string, error)
	AcceptTenantAdminInvitation(ctx context.Context, tenantID string, tenantSlug string, token string) (*platform.TenantAdminInvitation, *platform.TenantAdminUser, error)
	SetTenantAdminUserActive(ctx context.Context, tenantSlug string, userID string, active bool) (*platform.TenantAdminUser, error)
	UpdateTenantAdminUserProfile(ctx context.Context, tenantSlug string, userID string, req platform.UpdateTenantAdminUserProfileRequest) (*platform.TenantAdminUser, error)
	GetTenantSeatUsage(ctx context.Context, tenantID string) (*platform.TenantSeatUsage, error)
}

type roleStore interface {
	ListRoles(ctx context.Context, tenantID string, query string, limit int, offset int) ([]platform.RoleDefinition, int, error)
	CreateRole(ctx context.Context, tenantID string, req platform.CreateRoleRequest) (*platform.RoleDefinition, error)
	UpdateRole(ctx context.Context, tenantID string, roleID string, req platform.UpdateRoleRequest) (*platform.RoleDefinition, error)
	UpdateUserRoles(ctx context.Context, tenantID string, req platform.UpdateUserRolesRequest) (*platform.UpdateUserRolesResponse, error)
	GetEffectivePermissionsForUser(ctx context.Context, tenantID string, userID string, fallbackRole string) ([]string, []string, error)
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

type logsStore interface {
	ListRequestLogsV2(ctx context.Context, params platform.ListRequestLogsV2Params) ([]platform.RequestLogV2, int, error)
	GetRequestLogV2(ctx context.Context, tenantID string, logID string, includeContent bool) (*platform.RequestLogDetailV2, error)
	GetRequestLogFacetsV2(ctx context.Context, params platform.ListRequestLogsV2Params) (*platform.LogFacetsV2, error)
	GetRequestLogSummaryV2(ctx context.Context, params platform.ListRequestLogsV2Params) (*platform.LogSummaryV2, error)
	GetRequestLogRangeStatsV2(ctx context.Context, params platform.ListRequestLogsV2Params) (*platform.LogRangeStatsV2, error)
	GetRequestLogAnalyticsV2(ctx context.Context, params platform.LogAnalyticsV2Params) ([]platform.LogAnalyticsPointV2, error)
	ListSavedLogViewsV2(ctx context.Context, tenantID string, userID string) ([]platform.SavedLogViewV2, error)
	CreateSavedLogViewV2(ctx context.Context, tenantID string, userID string, req platform.CreateSavedLogViewV2Request) (*platform.SavedLogViewV2, error)
	UpdateSavedLogViewV2(ctx context.Context, tenantID string, savedViewID string, req platform.UpdateSavedLogViewV2Request) (*platform.SavedLogViewV2, error)
	DeleteSavedLogViewV2(ctx context.Context, tenantID string, savedViewID string) error
	ListLogExportsV2(ctx context.Context, tenantID string, limit int) ([]platform.LogExportJobV2, error)
	CreateLogExportV2(ctx context.Context, tenantID string, userID string, req platform.CreateLogExportV2Request) (*platform.LogExportJobV2, error)
	GetLogExportV2(ctx context.Context, tenantID string, exportID string) (*platform.LogExportJobV2, error)
	CancelLogExportV2(ctx context.Context, tenantID string, exportID string) (*platform.LogExportJobV2, error)
}

type creditUsageStore interface {
	GetCreditUsageSummary(ctx context.Context, tenantID string, start time.Time, end time.Time) (*platform.CreditUsageSummary, error)
}

type metricsStore interface {
	QueryModelMetrics(ctx context.Context, tenantID string, req platform.MetricsQueryRequest) ([]map[string]interface{}, error)
	ListModelMetricsFacets(ctx context.Context, tenantID string) (*platform.ModelMetricsFacets, error)
}

type tracePrivacyStore interface {
	GetTenantTracePrivacySettings(ctx context.Context, tenantID string) (*platform.TenantTracePrivacySettings, error)
	UpsertTenantTracePrivacySettings(ctx context.Context, tenantID, capturePolicy, updatedByUserID string) (*platform.TenantTracePrivacySettings, error)
}

type requestLoggingSettingsStore interface {
	GetTenantRequestLoggingSettings(ctx context.Context, tenantID string) (*platform.TenantRequestLoggingSettings, error)
	UpsertTenantRequestLoggingSettings(ctx context.Context, tenantID, loggingMode, updatedByUserID string) (*platform.TenantRequestLoggingSettings, error)
}

type promptStore interface {
	CreatePrompt(ctx context.Context, tenantID string, tenantSlug string, req platform.CreatePromptRequest, actor string) (*platform.PromptVersion, error)
	CreatePromptVersion(ctx context.Context, tenantID string, tenantSlug string, promptID string, req platform.CreatePromptVersionRequest, actor string) (*platform.PromptVersion, error)
	ListPrompts(ctx context.Context, tenantID string, query string, limit int, offset int) ([]platform.Prompt, int, error)
	GetPrompt(ctx context.Context, tenantID string, promptID string) (*platform.Prompt, error)
	DeletePrompt(ctx context.Context, tenantID string, promptID string) error
	ListPromptVersions(ctx context.Context, tenantID string, promptID string, limit int, offset int) ([]platform.PromptVersion, int, error)
	GetPromptVersion(ctx context.Context, tenantID string, versionID string) (*platform.PromptVersion, error)
	GetPromptVersionByFQN(ctx context.Context, tenantID string, fqn string) (*platform.PromptVersion, error)
	RenderPromptVersion(ctx context.Context, tenantID string, versionID string, req platform.RenderPromptVersionRequest) (*platform.RenderPromptVersionResponse, error)
}

type onboardingStore interface {
	CheckTenantSlugAvailability(ctx context.Context, slug string) (*tenantadmin.TenantSlugAvailability, error)
	CreateOnboardingOTP(ctx context.Context, req tenantadmin.CreateOnboardingOTPParams) (*tenantadmin.OnboardingOTP, error)
	VerifyOnboardingOTPAndProvision(ctx context.Context, req tenantadmin.VerifyOnboardingOTPParams) (*tenantadmin.ProvisionedTenant, error)
}

type loginDiscoveryStore interface {
	GetTenantBySlug(ctx context.Context, slug string) (*platform.Tenant, error)
	ListTenantsByAdminIdentifier(ctx context.Context, channel string, identifier string) ([]platform.Tenant, error)
}

type planStore interface {
	GetCommercialSignupSettings(ctx context.Context) (*platform.CommercialSignupSettings, error)
	UpdateCommercialSignupSettings(ctx context.Context, credit int64, actor string) (*platform.CommercialSignupSettings, error)
	ListFeatureCatalog(ctx context.Context) ([]platform.FeatureDefinition, error)
	ListPublicCommercialPlans(ctx context.Context) ([]platform.CommercialPlan, error)
	ListPlatformCommercialPlans(ctx context.Context) ([]platform.CommercialPlan, error)
	CreateCommercialPlan(ctx context.Context, req platform.CreateCommercialPlanRequest) (*platform.CommercialPlan, error)
	UpdateCommercialPlan(ctx context.Context, planSlug string, req platform.UpdateCommercialPlanRequest) (*platform.CommercialPlan, error)
	DeactivateCommercialPlan(ctx context.Context, planSlug string) error
	CreatePlanVersion(ctx context.Context, planSlug string, req platform.CreatePlanVersionRequest) (*platform.PlanVersion, error)
	PublishPlanVersion(ctx context.Context, planSlug string, version int) (*platform.PlanVersion, error)
	UpsertPlanEntitlement(ctx context.Context, planSlug string, version int, req platform.UpsertPlanEntitlementRequest) (*platform.PlanEntitlement, error)
	UpsertPlanPrice(ctx context.Context, planSlug string, version int, req platform.UpsertPlanPriceRequest) (*platform.PlanPrice, error)
	GetTenantSubscription(ctx context.Context, tenantID string) (*platform.TenantSubscription, error)
	ChangeTenantSubscription(ctx context.Context, tenantID string, req platform.ChangeTenantSubscriptionRequest) (*platform.TenantSubscription, error)
	PrepareTenantSubscriptionChange(ctx context.Context, tenantID string, req platform.ChangeTenantSubscriptionRequest, callbackURL string, returnURL string) (*platform.SubscriptionChangeResult, error)
	AttachSubscriptionPaymentGatewayData(ctx context.Context, publicID string, authority string, paymentURL string, raw map[string]interface{}) error
	GetSubscriptionPaymentIntentByAuthority(ctx context.Context, authority string) (*platform.SubscriptionPaymentIntent, error)
	CompleteSubscriptionPaymentByAuthority(ctx context.Context, authority string, verifyRaw map[string]interface{}, refID string, gatewayTxnID string, cardMask string, payerPhone string) (*platform.SubscriptionPaymentIntent, *platform.TenantSubscription, bool, error)
	MarkSubscriptionPaymentFailedByAuthority(ctx context.Context, authority string, reason string, raw map[string]interface{}) error
	ResolveEffectiveEntitlements(ctx context.Context, tenantID string) (*platform.EffectiveEntitlementsResponse, error)
	UpsertTenantSubscription(ctx context.Context, req platform.UpsertTenantSubscriptionRequest, actor string) (*platform.TenantSubscription, error)
	UpsertTenantEntitlementOverride(ctx context.Context, req platform.UpsertTenantEntitlementOverrideRequest) (*platform.TenantEntitlementOverride, error)
	DeleteTenantEntitlementOverride(ctx context.Context, tenantID string, featureKey string, actor string) error
}

type entitlementResolverStore interface {
	ResolveEffectiveEntitlements(ctx context.Context, tenantID string) (*platform.EffectiveEntitlementsResponse, error)
}

type supportTenantOperationsStore interface {
	ListSupportTenants(ctx context.Context, query, status, plan string, limit, offset int) ([]platform.SupportTenantSummary, error)
	GetSupportTenantSummary(ctx context.Context, tenantID string) (*platform.SupportTenantSummary, error)
	GetTenantOwnerContact(ctx context.Context, tenantID string) (*platform.TenantOwnerContact, error)
	ListTenantEvents(ctx context.Context, tenantID string, limit, offset int) ([]platform.TenantEvent, error)
	AdjustTenantWallet(ctx context.Context, tenantID string, req platform.WalletAdjustmentRequest, actor platform.SupportActor) (*platform.WalletAdjustmentResult, error)
	ChangeTenantStatus(ctx context.Context, tenantID, target, reason string, actor platform.SupportActor) (*platform.SupportTenantSummary, error)
	SetTenantSubscriptionCancellation(ctx context.Context, tenantID string, cancel bool, reason string, actor platform.SupportActor) (*platform.TenantSubscription, error)
	ListTenantSubscriptionPaymentIntents(ctx context.Context, tenantID string, limit, offset int) ([]platform.SubscriptionPaymentIntent, error)
}

type supportTicketStore interface {
	ListSupportTickets(context.Context, platform.SupportTicketListParams) ([]platform.SupportTicket, int, error)
	CountSupportTickets(context.Context, platform.SupportTicketListParams) (*platform.SupportTicketCounts, error)
	GetSupportTicket(context.Context, string, string, string, bool) (*platform.SupportTicket, error)
	GetSupportTicketDetail(context.Context, string, string, string, bool) (*platform.SupportTicketDetail, error)
	CreateSupportTicket(context.Context, string, string, platform.CreateSupportTicketRequest, []platform.SupportTicketAttachmentUpload, platform.SupportTicketActor, int) (*platform.SupportTicketDetail, bool, error)
	ListSupportTicketMessages(context.Context, string, int, int) ([]platform.SupportTicketMessage, error)
	AddSupportTicketMessage(context.Context, string, string, platform.AddSupportTicketMessageRequest, []platform.SupportTicketAttachmentUpload, platform.SupportTicketActor, bool, int) (*platform.SupportTicketDetail, bool, error)
	TriageSupportTicket(context.Context, string, platform.TriageSupportTicketRequest, platform.SupportTicketActor) (*platform.SupportTicket, error)
	ReopenSupportTicket(context.Context, string, platform.ReopenSupportTicketRequest, platform.SupportTicketActor) (*platform.SupportTicket, error)
	RateSupportTicket(context.Context, string, string, string, platform.RateSupportTicketRequest, platform.SupportTicketActor) (*platform.SupportTicket, error)
	ListSupportTicketEvents(context.Context, string, int, int) ([]platform.SupportTicketEvent, error)
	GetSupportTicketAttachment(context.Context, string, string, string, bool) (*platform.SupportTicketAttachmentObject, error)
	RecordSupportTicketAttachmentDownload(context.Context, *platform.SupportTicketAttachmentObject, platform.SupportTicketActor, string) error
}

type supportAttachmentService interface {
	UploadFiles(context.Context, string, string, []*multipart.FileHeader) ([]platform.SupportTicketAttachmentUpload, error)
	UploadReceiptImage(context.Context, string, *multipart.FileHeader) (*supportfiles.ReceiptUpload, error)
	Get(context.Context, string) (*supportfiles.ObjectReader, error)
	DeleteUploads(context.Context, []platform.SupportTicketAttachmentUpload)
	DeleteObject(context.Context, string)
}

type platformProviderCredentialService interface {
	Schemas(context.Context, string) ([]platform.ProviderConnectionSchema, error)
	CreateConnection(context.Context, platform.CreatePlatformProviderConnectionRequest, platformcredentials.Actor, string) (*platform.PlatformProviderConnection, error)
	UpdateConnection(context.Context, string, platform.UpdatePlatformProviderConnectionRequest, platformcredentials.Actor) (*platform.PlatformProviderConnection, error)
	ListConnections(context.Context, string, string) ([]platform.PlatformProviderConnection, error)
	GetConnection(context.Context, string) (*platform.PlatformProviderConnection, error)
	CreateCredentialVersion(context.Context, string, platform.CreatePlatformProviderCredentialVersionRequest, platformcredentials.Actor, string) (*platform.PlatformProviderCredentialVersion, error)
	TestCredential(context.Context, string, int64, platformcredentials.Actor) (*platform.PlatformProviderCredentialTestResult, error)
	Activate(context.Context, string, int64, int64, platformcredentials.Actor, string, string) (*platform.PlatformProviderConnection, error)
	Rollback(context.Context, string, int64, int64, platformcredentials.Actor, string) (*platform.PlatformProviderConnection, error)
	Disable(context.Context, string, int64, platformcredentials.Actor, string) (*platform.PlatformProviderConnection, error)
	Revoke(context.Context, string, int64, int64, platformcredentials.Actor, string) (*platform.PlatformProviderConnection, error)
}

// IAMDependencies contains identity and access management ports.
type IAMDependencies struct {
	Keys         keyStore
	OrgUnits     orgUnitStore
	Memberships  membershipStore
	Users        userStore
	VirtualUsers virtualUserStore
	Teams        teamStore
	Roles        roleStore
}

// CatalogDependencies contains model, prompt, and provider catalog ports.
type CatalogDependencies struct {
	Models              modelStore
	Policies            policyStore
	ProviderAccounts    providerAccountStore
	ProviderSchemas     platformcredentials.SchemaCatalog
	ProviderCredentials platformProviderCredentialService
	Prompts             promptStore
	Discovery           DiscoveryConfig
	SemanticGuardrails  semanticPolicyTester
}

// BillingDependencies contains wallet, plan, and entitlement ports.
type BillingDependencies struct {
	Wallets      walletStore
	Plans        planStore
	Entitlements entitlementResolverStore
	Payment      PaymentConfig
	// CreditUsage reads shadow credit usage; nil disables the credit summary.
	CreditUsage creditUsageStore
	// CreditPacks sells add-on packs; nil disables the credit pack endpoints.
	CreditPacks creditPackStore
	// CreditAdmin backs the support-admin credit endpoints.
	CreditAdmin creditAdminStore
}

// OnboardingDependencies contains onboarding and login discovery ports.
type OnboardingDependencies struct {
	Onboarding     onboardingStore
	LoginDiscovery loginDiscoveryStore
}

// ObservabilityDependencies contains logs, traces, metrics, and settings ports.
type ObservabilityDependencies struct {
	Traces         traceStore
	Logs           logsStore
	Metrics        metricsStore
	TracePrivacy   tracePrivacyStore
	RequestLogging requestLoggingSettingsStore
}

// SupportDependencies contains tenant support and attachment ports.
type SupportDependencies struct {
	Tenants supportTenantOperationsStore
	Tickets supportTicketStore
	Files   supportAttachmentService
	Limits  SupportLimitsConfig
}

// Dependencies groups management ports by feature boundary.
type Dependencies struct {
	IAM           IAMDependencies
	Catalog       CatalogDependencies
	Billing       BillingDependencies
	Onboarding    OnboardingDependencies
	Observability ObservabilityDependencies
	Support       SupportDependencies
}

type iamState struct {
	keys         keyStore
	orgUnits     orgUnitStore
	memberships  membershipStore
	users        userStore
	virtualUsers virtualUserStore
	teams        teamStore
	roles        roleStore
}

type catalogState struct {
	models              modelStore
	policies            policyStore
	providerAccounts    providerAccountStore
	providerSchemas     platformcredentials.SchemaCatalog
	providerCredentials platformProviderCredentialService
	prompts             promptStore
	discovery           DiscoveryConfig
	semanticGuardrails  semanticPolicyTester
	semanticTestLimiter *authRateLimiter
}

type billingState struct {
	wallets      walletStore
	plans        planStore
	entitlements entitlementResolverStore
	payment      PaymentConfig
	creditUsage  creditUsageStore
	creditPacks  creditPackStore
	creditAdmin  creditAdminStore
}

type onboardingState struct {
	onboarding     onboardingStore
	loginDiscovery loginDiscoveryStore
}

type observabilityState struct {
	traces         traceStore
	logs           logsStore
	metrics        metricsStore
	tracePrivacy   tracePrivacyStore
	requestLogging requestLoggingSettingsStore
}

type supportState struct {
	supportTenants supportTenantOperationsStore
	supportTickets supportTicketStore
	supportFiles   supportAttachmentService
	limits         SupportLimitsConfig
}

// Handlers serves management APIs while retaining feature-owned state.
type Handlers struct {
	iamState
	catalogState
	billingState
	onboardingState
	observabilityState
	supportState
}

// New creates a management handlers bundle.
func New(deps Dependencies) *Handlers {
	return &Handlers{
		iamState: iamState{
			keys:         deps.IAM.Keys,
			orgUnits:     deps.IAM.OrgUnits,
			memberships:  deps.IAM.Memberships,
			users:        deps.IAM.Users,
			virtualUsers: deps.IAM.VirtualUsers,
			teams:        deps.IAM.Teams,
			roles:        deps.IAM.Roles,
		},
		catalogState: catalogState{
			models:              deps.Catalog.Models,
			policies:            deps.Catalog.Policies,
			providerAccounts:    deps.Catalog.ProviderAccounts,
			providerSchemas:     deps.Catalog.ProviderSchemas,
			providerCredentials: deps.Catalog.ProviderCredentials,
			prompts:             deps.Catalog.Prompts,
			discovery:           deps.Catalog.Discovery,
			semanticGuardrails:  deps.Catalog.SemanticGuardrails,
			semanticTestLimiter: newAuthRateLimiter(),
		},
		billingState: billingState{
			wallets:      deps.Billing.Wallets,
			plans:        deps.Billing.Plans,
			entitlements: deps.Billing.Entitlements,
			payment:      deps.Billing.Payment.withDefaults(),
			creditUsage:  deps.Billing.CreditUsage,
			creditPacks:  deps.Billing.CreditPacks,
			creditAdmin:  deps.Billing.CreditAdmin,
		},
		onboardingState: onboardingState{
			onboarding:     deps.Onboarding.Onboarding,
			loginDiscovery: deps.Onboarding.LoginDiscovery,
		},
		observabilityState: observabilityState{
			traces:         deps.Observability.Traces,
			logs:           deps.Observability.Logs,
			metrics:        deps.Observability.Metrics,
			tracePrivacy:   deps.Observability.TracePrivacy,
			requestLogging: deps.Observability.RequestLogging,
		},
		supportState: supportState{
			supportTenants: deps.Support.Tenants,
			supportTickets: deps.Support.Tickets,
			supportFiles:   deps.Support.Files,
			limits:         deps.Support.Limits.withDefaults(),
		},
	}
}

// ListRoles lists tenant-scoped role definitions.
func isTenantAdminUser(user *platform.TenantAdminUser) bool {
	if user == nil {
		return false
	}

	if strings.EqualFold(strings.TrimSpace(user.Role), "tenant_admin") {
		return true
	}

	for _, role := range user.Roles {
		if strings.EqualFold(strings.TrimSpace(role), "tenant_admin") {
			return true
		}
	}

	return false
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

func (h *Handlers) requireLoginDiscoveryStore() (loginDiscoveryStore, error) {
	if h.loginDiscovery == nil {
		return nil, fiber.NewError(fiber.StatusNotImplemented, "login discovery store not configured")
	}

	return h.loginDiscovery, nil
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
