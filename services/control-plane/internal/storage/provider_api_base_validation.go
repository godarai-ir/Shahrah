package storage

import (
	"errors"
	"fmt"

	"github.com/Delpak/shahrah/services/control-plane/internal/providervalidation"
)

var ErrInvalidProviderAPIBase = errors.New("invalid provider api_base")

func ValidateProviderAPIBase(provider string, raw string) (string, error) {
	normalized, err := providervalidation.ValidateBaseURL(provider, raw)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidProviderAPIBase, err)
	}

	return normalized, nil
}

func validateProviderConfigAPIBase(provider string, config map[string]interface{}) error {
	if err := providervalidation.NormalizeBaseURLConfig(provider, config); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidProviderAPIBase, err)
	}

	return nil
}

func safeProviderAPIBase(provider string, raw string) string {
	normalized, err := ValidateProviderAPIBase(provider, raw)
	if err != nil {
		return ""
	}

	return normalized
}
