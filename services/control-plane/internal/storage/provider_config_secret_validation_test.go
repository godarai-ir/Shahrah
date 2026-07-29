package storage

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateProviderConfigContainsNoSecrets(t *testing.T) {
	t.Run("accepts connection identifiers and runtime settings", func(t *testing.T) {
		require.NoError(t, validateProviderConfigContainsNoSecrets(
			map[string]interface{}{
				"api_base":        "https://api.anthropic.com",
				"api_version":     "2023-06-01",
				"token_source_id": "workload-identity-source",
				"nested": map[string]interface{}{
					"organization_id": "org_test",
				},
			},
		))
	})

	for name, config := range map[string]map[string]interface{}{
		"top-level API key": {
			"api_key": "sk-plaintext",
		},
		"nested client secret": {
			"oauth": map[string]interface{}{
				"client_secret": "plaintext",
			},
		},
		"secret in array": {
			"connections": []interface{}{
				map[string]interface{}{"access-token": "plaintext"},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			err := validateProviderConfigContainsNoSecrets(config)
			require.ErrorIs(t, err, ErrProviderConfigContainsSecret)
			require.NotContains(t, err.Error(), "sk-plaintext")
			require.NotContains(t, err.Error(), "plaintext")
		})
	}
}
