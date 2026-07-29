package storage

import (
	"errors"
	"fmt"
	"strings"
)

var ErrProviderConfigContainsSecret = errors.New(
	"provider_config must not contain credentials",
)

var providerConfigSecretKeys = map[string]struct{}{
	"api_key":       {},
	"apikey":        {},
	"access_token":  {},
	"authorization": {},
	"client_secret": {},
	"credential":    {},
	"credentials":   {},
	"password":      {},
	"private_key":   {},
	"secret":        {},
	"secret_ref":    {},
	"token":         {},
}

func validateProviderConfigContainsNoSecrets(
	config map[string]interface{},
) error {
	return walkProviderConfigForSecrets(config, "provider_config")
}

func walkProviderConfigForSecrets(value interface{}, path string) error {
	switch typed := value.(type) {
	case map[string]interface{}:
		for key, child := range typed {
			normalized := strings.ToLower(strings.TrimSpace(key))
			normalized = strings.ReplaceAll(normalized, "-", "_")
			if _, forbidden := providerConfigSecretKeys[normalized]; forbidden ||
				strings.HasSuffix(normalized, "_api_key") ||
				strings.HasSuffix(normalized, "_access_token") ||
				strings.HasSuffix(normalized, "_client_secret") ||
				strings.HasSuffix(normalized, "_password") ||
				strings.HasSuffix(normalized, "_private_key") {
				return fmt.Errorf(
					"%w: %s.%s must be supplied through the write-only credentials field",
					ErrProviderConfigContainsSecret,
					path,
					key,
				)
			}
			if err := walkProviderConfigForSecrets(
				child,
				path+"."+key,
			); err != nil {
				return err
			}
		}
	case []interface{}:
		for index, child := range typed {
			if err := walkProviderConfigForSecrets(
				child,
				fmt.Sprintf("%s[%d]", path, index),
			); err != nil {
				return err
			}
		}
	}
	return nil
}
