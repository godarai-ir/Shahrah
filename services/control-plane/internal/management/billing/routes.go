// Package billing owns subscription, wallet, and commercial billing HTTP routes.
package billing

import "github.com/gofiber/fiber/v2"

// Handler defines billing operations consumed by this feature.
type Handler interface {
	ListPublicCommercialPlans(*fiber.Ctx) error
	SubscriptionPaymentCallback(*fiber.Ctx) error
	GetTenantSubscription(*fiber.Ctx) error
	ChangeTenantSubscription(*fiber.Ctx) error
	GetTenantEntitlements(*fiber.Ctx) error
	GetWallet(*fiber.Ctx) error
	TopUpWallet(*fiber.Ctx) error
	CreateWalletTopUpIntent(*fiber.Ctx) error
	ListWalletTransactions(*fiber.Ctx) error
	ListBankTransfers(*fiber.Ctx) error
	CreateBankTransfer(*fiber.Ctx) error
	DownloadBankTransferReceipt(*fiber.Ctx) error
	GetWalletTopUpSettings(*fiber.Ctx) error
	UpdateWalletTopUpSettings(*fiber.Ctx) error
	WalletTopUpCallback(*fiber.Ctx) error
	RedeemGiftCode(*fiber.Ctx) error
	RedeemDiscountCode(*fiber.Ctx) error
	GetBillingProfile(*fiber.Ctx) error
	UpsertBillingProfile(*fiber.Ctx) error
	GetBillingUsageSummary(*fiber.Ctx) error
	GetCreditUsageSummary(*fiber.Ctx) error
	GetBillingMonthlyUsage(*fiber.Ctx) error
	ListBillingInvoices(*fiber.Ctx) error
	ListBillingTransactions(*fiber.Ctx) error
	ListBillingDiscounts(*fiber.Ctx) error
	ListPlatformFeatureCatalog(*fiber.Ctx) error
	ListPlatformCommercialPlans(*fiber.Ctx) error
	GetPlatformCommercialSignupSettings(*fiber.Ctx) error
	UpdatePlatformCommercialSignupSettings(*fiber.Ctx) error
	CreatePlatformCommercialPlan(*fiber.Ctx) error
	UpdatePlatformCommercialPlan(*fiber.Ctx) error
	DeactivatePlatformCommercialPlan(*fiber.Ctx) error
	CreatePlatformPlanVersion(*fiber.Ctx) error
	PublishPlatformPlanVersion(*fiber.Ctx) error
	UpsertPlatformPlanEntitlement(*fiber.Ctx) error
	UpsertPlatformPlanPrice(*fiber.Ctx) error
	UpsertPlatformTenantSubscription(*fiber.Ctx) error
	UpsertPlatformTenantEntitlementOverride(*fiber.Ctx) error
	DeletePlatformTenantEntitlementOverride(*fiber.Ctx) error
	GetPlatformTenantWallet(*fiber.Ctx) error
	ListPlatformTenantWalletTransactions(*fiber.Ctx) error
	AdjustPlatformTenantWallet(*fiber.Ctx) error
	ApprovePlatformBankTransfer(*fiber.Ctx) error
	RejectPlatformBankTransfer(*fiber.Ctx) error
	SchedulePlatformTenantSubscriptionCancellation(*fiber.Ctx) error
	ReversePlatformTenantSubscriptionCancellation(*fiber.Ctx) error
	ListPlatformTenantSubscriptionPaymentIntents(*fiber.Ctx) error
}

// Routes owns billing route registration.
type Routes struct {
	handler Handler
}

// NewRoutes creates billing routes.
func NewRoutes(handler Handler) *Routes {
	return &Routes{handler: handler}
}

// RegisterPublic registers public commercial billing endpoints.
func (r *Routes) RegisterPublic(router fiber.Router) {
	router.Get("/api/v1/commercial-plans", r.handler.ListPublicCommercialPlans)
	router.Get("/api/v1/subscription/payment/callback/:provider", r.handler.SubscriptionPaymentCallback)
	router.Post("/api/v1/subscription/payment/callback/:provider", r.handler.SubscriptionPaymentCallback)
}

// RegisterTenant registers authenticated tenant billing endpoints.
func (r *Routes) RegisterTenant(router fiber.Router) {
	r.registerTenantCredits(router)
	router.Get("/subscription", r.handler.GetTenantSubscription)
	router.Post("/subscription/change", r.handler.ChangeTenantSubscription)
	router.Get("/entitlements", r.handler.GetTenantEntitlements)
	router.Get("/wallet", r.handler.GetWallet)
	router.Post("/wallet/topup", r.handler.TopUpWallet)
	router.Post("/wallet/topup/intents", r.handler.CreateWalletTopUpIntent)
	router.Get("/wallet/transactions", r.handler.ListWalletTransactions)
	router.Get("/wallet/bank-transfers", r.handler.ListBankTransfers)
	router.Post("/wallet/bank-transfers", r.handler.CreateBankTransfer)
	router.Get("/wallet/bank-transfer-receipts/:fileName", r.handler.DownloadBankTransferReceipt)
	router.Get("/wallet/topup/settings", r.handler.GetWalletTopUpSettings)
	router.Put("/wallet/topup/settings", r.handler.UpdateWalletTopUpSettings)
	router.Get("/wallet/topup/callback/:provider", r.handler.WalletTopUpCallback)
	router.Post("/wallet/topup/callback/:provider", r.handler.WalletTopUpCallback)
	router.Post("/billing/gift-codes/redeem", r.handler.RedeemGiftCode)
	router.Post("/billing/discount-codes/redeem", r.handler.RedeemDiscountCode)
	router.Get("/billing/profile", r.handler.GetBillingProfile)
	router.Put("/billing/profile", r.handler.UpsertBillingProfile)
	router.Get("/billing/reports/summary", r.handler.GetBillingUsageSummary)
	router.Get("/billing/reports/monthly", r.handler.GetBillingMonthlyUsage)
	router.Get("/billing/documents/invoices", r.handler.ListBillingInvoices)
	router.Get("/billing/documents/transactions", r.handler.ListBillingTransactions)
	router.Get("/billing/documents/discounts", r.handler.ListBillingDiscounts)
}

// registerTenantCredits registers tenant credit endpoints.
func (r *Routes) registerTenantCredits(router fiber.Router) {
	router.Get("/billing/credits/summary", r.handler.GetCreditUsageSummary)
}

// RegisterPlatform registers platform billing endpoints.
func (r *Routes) RegisterPlatform(router fiber.Router) {
	router.Get("/features", r.handler.ListPlatformFeatureCatalog)
	router.Get("/commercial-plans", r.handler.ListPlatformCommercialPlans)
	router.Get("/commercial-signup-settings", r.handler.GetPlatformCommercialSignupSettings)
	router.Put("/commercial-signup-settings", r.handler.UpdatePlatformCommercialSignupSettings)
	router.Post("/commercial-plans", r.handler.CreatePlatformCommercialPlan)
	router.Patch("/commercial-plans/:slug", r.handler.UpdatePlatformCommercialPlan)
	router.Delete("/commercial-plans/:slug", r.handler.DeactivatePlatformCommercialPlan)
	router.Post("/commercial-plans/:slug/versions", r.handler.CreatePlatformPlanVersion)
	router.Post("/commercial-plans/:slug/versions/:version/publish", r.handler.PublishPlatformPlanVersion)
	router.Post("/commercial-plans/:slug/versions/:version/entitlements", r.handler.UpsertPlatformPlanEntitlement)
	router.Put("/commercial-plans/:slug/versions/:version/entitlements/:featureKey", r.handler.UpsertPlatformPlanEntitlement)
	router.Post("/commercial-plans/:slug/versions/:version/prices", r.handler.UpsertPlatformPlanPrice)
	router.Put("/commercial-plans/:slug/versions/:version/prices/:interval", r.handler.UpsertPlatformPlanPrice)
	router.Put("/tenant-subscriptions", r.handler.UpsertPlatformTenantSubscription)
	router.Put("/tenant-entitlement-overrides", r.handler.UpsertPlatformTenantEntitlementOverride)
	router.Delete("/tenant-entitlement-overrides/:tenantID/:featureKey", r.handler.DeletePlatformTenantEntitlementOverride)
	router.Get("/tenants/:tenantID/wallet", r.handler.GetPlatformTenantWallet)
	router.Get("/tenants/:tenantID/wallet/transactions", r.handler.ListPlatformTenantWalletTransactions)
	router.Post("/tenants/:tenantID/wallet/adjustments", r.handler.AdjustPlatformTenantWallet)
	router.Post("/tenants/:tenantID/wallet/bank-transfers/:requestID/approve", r.handler.ApprovePlatformBankTransfer)
	router.Post("/tenants/:tenantID/wallet/bank-transfers/:requestID/reject", r.handler.RejectPlatformBankTransfer)
	router.Post("/tenants/:tenantID/subscription/cancel-at-period-end", r.handler.SchedulePlatformTenantSubscriptionCancellation)
	router.Delete("/tenants/:tenantID/subscription/cancel-at-period-end", r.handler.ReversePlatformTenantSubscriptionCancellation)
	router.Get("/tenants/:tenantID/billing/payment-intents", r.handler.ListPlatformTenantSubscriptionPaymentIntents)
}
