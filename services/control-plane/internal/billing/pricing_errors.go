package billing

import (
	"errors"
	"fmt"
	"strings"
)

var (
	// ErrPricingUnavailable indicates that the provider pricing snapshot is missing or stale.
	ErrPricingUnavailable = errors.New("pricing_unavailable")
	// ErrPricingModelUnavailable indicates that the snapshot has no billable entry for the requested model.
	ErrPricingModelUnavailable = errors.New("pricing_model_unavailable")
	// ErrPricingRateUnavailable indicates that the model has no unambiguous rate for the requested configuration.
	ErrPricingRateUnavailable = errors.New("pricing_rate_unavailable")
	// ErrUnsupportedServiceTier indicates that the requested service tier is not supported by the model.
	ErrUnsupportedServiceTier = errors.New("unsupported_service_tier")
)

// UnsupportedServiceTierError describes a model-specific service-tier validation failure.
type UnsupportedServiceTierError struct {
	Provider string
	Model    string
	Tier     string
}

// Error returns the internal error text.
func (e *UnsupportedServiceTierError) Error() string {
	if e == nil {
		return ErrUnsupportedServiceTier.Error()
	}

	return fmt.Sprintf(
		"%s model %s does not support service_tier %q",
		strings.TrimSpace(e.Provider),
		strings.TrimSpace(e.Model),
		strings.TrimSpace(e.Tier),
	)
}

// Unwrap exposes the stable error category.
func (e *UnsupportedServiceTierError) Unwrap() error {
	return ErrUnsupportedServiceTier
}

// PublicMessage returns a client-safe validation message.
func (e *UnsupportedServiceTierError) PublicMessage() string {
	if e == nil {
		return "The requested service_tier is not supported"
	}

	return fmt.Sprintf(
		"Model %s does not support service_tier %q",
		strings.TrimSpace(e.Model),
		strings.TrimSpace(e.Tier),
	)
}
