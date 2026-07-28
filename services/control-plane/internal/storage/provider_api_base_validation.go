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

func validateProviderConfigAPIBase(provider string, config map[string]interface{}, keys ...string) error {
	for _, key := range keys {
		raw, ok := config[key].(string)
		if !ok {
			continue
		}

		normalized, err := ValidateProviderAPIBase(provider, raw)
		if err != nil {
			return err
		}

		if normalized == "" {
			delete(config, key)
		} else {
			config[key] = normalized
		}
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
