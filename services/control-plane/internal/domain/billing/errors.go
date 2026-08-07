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
	ErrSubscriptionNotFound              = errors.New("tenant subscription not found")
	ErrSubscriptionPaymentIntentNotFound = errors.New("subscription payment intent not found")
	ErrTenantQuotaExceeded               = errors.New("tenant quota exceeded")
	ErrTopUpIntentNotFound               = errors.New("top-up intent not found")
	ErrBankTransferNotFound              = errors.New("bank transfer request not found")
	ErrFinancialCodeInvalid              = errors.New("financial code is invalid")
	ErrBillingProfileInvalid             = errors.New("billing profile is invalid")
)
