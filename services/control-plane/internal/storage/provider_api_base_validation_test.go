package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateProviderAPIBaseAllowsOfficialProviderHosts(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		raw      string
		want     string
	}{
		{name: "openai", provider: "openai", raw: "https://api.openai.com/v1/", want: "https://api.openai.com/v1"},
		{name: "groq", provider: "groq", raw: "https://api.groq.com/openai/v1", want: "https://api.groq.com/openai/v1"},
		{name: "azure", provider: "azure", raw: "https://tenant.openai.azure.com/openai/deployments/gpt", want: "https://tenant.openai.azure.com/openai/deployments/gpt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateProviderAPIBase(tt.provider, tt.raw)

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestValidateProviderAPIBaseRejectsSSRFTargets(t *testing.T) {
	tests := []struct {
		name string
		raw  string
	}{
		{name: "loopback", raw: "https://127.0.0.1:8080/v1"},
		{name: "metadata service", raw: "http://169.254.169.254/latest/meta-data"},
		{name: "private hostname", raw: "https://localhost:8080/v1"},
		{name: "wrong provider host", raw: "https://api.evil.example/v1"},
		{name: "userinfo", raw: "https://token@api.openai.com/v1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateProviderAPIBase("openai", tt.raw)

			require.Error(t, err)
			assert.ErrorIs(t, err, ErrInvalidProviderAPIBase)
		})
	}
}

func TestValidateProviderConfigAPIBaseCoversEveryPersistedAlias(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		key      string
		value    interface{}
	}{
		{name: "Azure endpoint", provider: "azure-open-ai", key: "endpoint", value: "https://127.0.0.1"},
		{name: "Databricks serving endpoint", provider: "databricks", key: "serving_endpoint_url", value: "https://metadata.google.internal"},
		{name: "non-string base URL", provider: "openai", key: "base_url", value: 42},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := map[string]interface{}{tt.key: tt.value}

			err := validateProviderConfigAPIBase(tt.provider, config)

			require.Error(t, err)
			assert.ErrorIs(t, err, ErrInvalidProviderAPIBase)
			assert.Equal(t, tt.value, config[tt.key])
		})
	}
}
