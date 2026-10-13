// Package billing contains billing-domain rules, errors, and value objects.
package billing

import "errors"

var (
	ErrFeatureNotFound                   = errors.New("feature not found")
	ErrFeatureOverrideForbidden          = errors.New("feature override is not allowed")
	ErrInsufficientWalletBalance         = errors.New("insufficient wallet balance")
	ErrInvalidSignupCredit               = errors.New("signup credit must be a non-negative whole IRR amount")
	ErrPlanNotFound                      = errors.New("commercial plan not found")
	ErrPlanPriceUnavailable              = errors.New("plan price is not configured")
	ErrPlanVersionNotFound               = errors.New("plan version not found")
	ErrPlanVersionPublished              = errors.New("published plan versions are immutable")
	ErrPlanUserLimitReached              = errors.New("plan user limit reached")
	ErrCreditBillingNotEnabled           = errors.New("tenant is not billed in credits")
	ErrCreditPackNotFound                = errors.New("credit pack not found")
	ErrCreditPackOrderNotFound           = errors.New("credit pack order not found")
	ErrCreditPackOrderNotPayable         = errors.New("credit pack order cannot be paid in its current state")
	ErrCreditPackTransferMismatch        = errors.New("bank transfer amount does not match the credit pack order")
	ErrCreditPaymentProviderUnavailable  = errors.New("payment provider is not available for credit packs")
	ErrOverageNotAllowed                 = errors.New("overage is not allowed on this plan")
	ErrInvalidOverageCap                 = errors.New("overage cap must be a positive amount of credits with at most two decimals")
	ErrSubscriptionNotFound              = errors.New("tenant subscription not found")
	ErrSubscriptionPaymentIntentNotFound = errors.New("subscription payment intent not found")
	ErrTenantQuotaExceeded               = errors.New("tenant quota exceeded")
	ErrTopUpIntentNotFound               = errors.New("top-up intent not found")
	ErrBankTransferNotFound              = errors.New("bank transfer request not found")
	ErrFinancialCodeInvalid              = errors.New("financial code is invalid")
	ErrBillingProfileInvalid             = errors.New("billing profile is invalid")
)
