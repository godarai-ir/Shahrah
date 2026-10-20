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
	GetCreditBalance(*fiber.Ctx) error
	ListCreditPacks(*fiber.Ctx) error
	ListCreditPackOrders(*fiber.Ctx) error
	GetCreditPackOrder(*fiber.Ctx) error
	CreateCreditPackOrder(*fiber.Ctx) error
	CreditPackPaymentCallback(*fiber.Ctx) error
	GetCreditSettings(*fiber.Ctx) error
	UpdateCreditSettings(*fiber.Ctx) error
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
	GetPlatformCreditPricePolicy(*fiber.Ctx) error
	ApprovePlatformCreditPriceRevision(*fiber.Ctx) error
	RejectPlatformCreditPriceRevision(*fiber.Ctx) error
	ListPlatformCreditRateCards(*fiber.Ctx) error
	CreatePlatformCreditRateCardDraft(*fiber.Ctx) error
	UpsertPlatformCreditRateCardMultiplier(*fiber.Ctx) error
	PublishPlatformCreditRateCard(*fiber.Ctx) error
	ListPlatformCreditPacks(*fiber.Ctx) error
	CreatePlatformCreditPack(*fiber.Ctx) error
	PublishPlatformCreditPack(*fiber.Ctx) error
	GetPlatformTenantCredits(*fiber.Ctx) error
	ListPlatformTenantCreditLedger(*fiber.Ctx) error
	ListPlatformTenantCreditLots(*fiber.Ctx) error
	ListPlatformTenantCreditLotAllocations(*fiber.Ctx) error
	AdjustPlatformTenantCredits(*fiber.Ctx) error
	UpdatePlatformTenantCreditOverage(*fiber.Ctx) error
	GetPlatformTenantCreditConversion(*fiber.Ctx) error
	ConvertPlatformTenantToCredits(*fiber.Ctx) error
	RollbackPlatformTenantCredits(*fiber.Ctx) error
	ListPlatformCreditConversionPreviews(*fiber.Ctx) error
	ConvertPlatformCreditFleet(*fiber.Ctx) error
	GetSeatSummary(*fiber.Ctx) error
	ListSeatMembers(*fiber.Ctx) error
	SetMemberSeat(*fiber.Ctx) error
	CancelPendingSeatChange(*fiber.Ctx) error
	QuoteSeats(*fiber.Ctx) error
	QuoteSeatRenewal(*fiber.Ctx) error
	PreviewSeatRenewal(*fiber.Ctx) error
	UpdateSeatRenewal(*fiber.Ctx) error
	ListSeatOrders(*fiber.Ctx) error
	CreateSeatOrder(*fiber.Ctx) error
	SeatPaymentCallback(*fiber.Ctx) error
	ListSeatRequests(*fiber.Ctx) error
	ApproveSeatRequest(*fiber.Ctx) error
	DenySeatRequest(*fiber.Ctx) error
	ListPlatformSeatTypes(*fiber.Ctx) error
	UpdatePlatformSeatType(*fiber.Ctx) error
	GetPlatformTenantSeats(*fiber.Ctx) error
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
	router.Get("/api/v1/credits/packs/payment/callback/:provider", r.handler.CreditPackPaymentCallback)
	router.Post("/api/v1/credits/packs/payment/callback/:provider", r.handler.CreditPackPaymentCallback)
	router.Get("/api/v1/seats/payment/callback/:provider", r.handler.SeatPaymentCallback)
	router.Post("/api/v1/seats/payment/callback/:provider", r.handler.SeatPaymentCallback)
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
	router.Get("/billing/credits/balance", r.handler.GetCreditBalance)
	router.Get("/billing/credits/packs", r.handler.ListCreditPacks)
	router.Get("/billing/credits/orders", r.handler.ListCreditPackOrders)
	router.Post("/billing/credits/orders", r.handler.CreateCreditPackOrder)
	router.Get("/billing/credits/orders/:orderID", r.handler.GetCreditPackOrder)
	router.Get("/billing/credits/settings", r.handler.GetCreditSettings)
	router.Put("/billing/credits/settings", r.handler.UpdateCreditSettings)
	r.registerTenantSeats(router)
}

// registerTenantSeats registers member seat endpoints.
func (r *Routes) registerTenantSeats(router fiber.Router) {
	router.Get("/billing/seats", r.handler.GetSeatSummary)
	router.Get("/billing/seats/members", r.handler.ListSeatMembers)
	router.Put("/billing/seats/members/:userID", r.handler.SetMemberSeat)
	router.Delete("/billing/seats/members/:userID/pending", r.handler.CancelPendingSeatChange)
	router.Post("/billing/seats/quote", r.handler.QuoteSeats)
	router.Get("/billing/seats/renewal", r.handler.QuoteSeatRenewal)
	router.Get("/billing/seats/renewal/preview", r.handler.PreviewSeatRenewal)
	router.Put("/billing/seats/renewal", r.handler.UpdateSeatRenewal)
	router.Get("/billing/seats/orders", r.handler.ListSeatOrders)
	router.Post("/billing/seats/orders", r.handler.CreateSeatOrder)
	router.Get("/billing/seats/requests", r.handler.ListSeatRequests)
	router.Post("/billing/seats/requests/:requestID/approve", r.handler.ApproveSeatRequest)
	router.Post("/billing/seats/requests/:requestID/deny", r.handler.DenySeatRequest)
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
	r.registerPlatformCredits(router)
}

// registerPlatformCredits registers support-admin credit billing endpoints.
func (r *Routes) registerPlatformCredits(router fiber.Router) {
	router.Get("/credits/price-policy", r.handler.GetPlatformCreditPricePolicy)
	router.Post("/credits/price-revisions/:revisionID/approve", r.handler.ApprovePlatformCreditPriceRevision)
	router.Post("/credits/price-revisions/:revisionID/reject", r.handler.RejectPlatformCreditPriceRevision)
	router.Get("/credits/rate-cards", r.handler.ListPlatformCreditRateCards)
	router.Post("/credits/rate-cards", r.handler.CreatePlatformCreditRateCardDraft)
	router.Put("/credits/rate-cards/:version/multipliers", r.handler.UpsertPlatformCreditRateCardMultiplier)
	router.Post("/credits/rate-cards/:version/publish", r.handler.PublishPlatformCreditRateCard)
	router.Get("/credits/packs", r.handler.ListPlatformCreditPacks)
	router.Post("/credits/packs", r.handler.CreatePlatformCreditPack)
	router.Post("/credits/packs/:slug/versions/:version/publish", r.handler.PublishPlatformCreditPack)
	router.Get("/tenants/:tenantID/credits", r.handler.GetPlatformTenantCredits)
	router.Get("/tenants/:tenantID/credits/ledger", r.handler.ListPlatformTenantCreditLedger)
	router.Get("/tenants/:tenantID/credits/lots", r.handler.ListPlatformTenantCreditLots)
	router.Get("/tenants/:tenantID/credits/lots/:lotID/allocations", r.handler.ListPlatformTenantCreditLotAllocations)
	router.Post("/tenants/:tenantID/credits/adjustments", r.handler.AdjustPlatformTenantCredits)
	router.Put("/tenants/:tenantID/credits/overage", r.handler.UpdatePlatformTenantCreditOverage)
	router.Get("/credits/conversions/preview", r.handler.ListPlatformCreditConversionPreviews)
	router.Post("/credits/conversions/fleet", r.handler.ConvertPlatformCreditFleet)
	router.Get("/tenants/:tenantID/credits/conversion", r.handler.GetPlatformTenantCreditConversion)
	router.Post("/tenants/:tenantID/credits/conversion", r.handler.ConvertPlatformTenantToCredits)
	router.Post("/tenants/:tenantID/credits/conversion/rollback", r.handler.RollbackPlatformTenantCredits)
	router.Get("/seat-types", r.handler.ListPlatformSeatTypes)
	router.Put("/seat-types/:seatType", r.handler.UpdatePlatformSeatType)
	router.Get("/tenants/:tenantID/seats", r.handler.GetPlatformTenantSeats)
}
