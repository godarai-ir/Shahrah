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
		{name: "xai", provider: "xai", raw: "https://api.x.ai/v1/", want: "https://api.x.ai/v1"},
		{name: "moonshot", provider: "moonshot", raw: "https://api.moonshot.ai/v1/", want: "https://api.moonshot.ai/v1"},
		{name: "deepseek", provider: "deepseek", raw: "https://api.deepseek.com/", want: "https://api.deepseek.com"},
		{name: "mistral canonical", provider: "mistral-ai", raw: "https://api.mistral.ai/v1/", want: "https://api.mistral.ai/v1"},
		{name: "mistral legacy alias", provider: "mistral", raw: "https://api.mistral.ai/v1", want: "https://api.mistral.ai/v1"},
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

func TestValidateProviderAPIBaseAllowsAnyHostForBYOKProviders(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		raw      string
		want     string
	}{
		{name: "zai custom gateway", provider: "zai", raw: "https://my-gateway.example.com/zai/v4/", want: "https://my-gateway.example.com/zai/v4"},
		{name: "zai canonical host", provider: "zai", raw: "https://api.z.ai/api/paas/v4", want: "https://api.z.ai/api/paas/v4"},
		{name: "deepseek custom gateway", provider: "deepseek", raw: "https://my-gateway.example.com/v1/", want: "https://my-gateway.example.com/v1"},
		{name: "deepseek with port and path", provider: "deepseek", raw: "https://api.deepseek.com:8443/v1", want: "https://api.deepseek.com:8443/v1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateProviderAPIBase(tt.provider, tt.raw)

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestValidateProviderAPIBaseRejectsSSRFTargetsForBYOKProviders(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		raw      string
	}{
		{name: "zai loopback", provider: "zai", raw: "https://127.0.0.1:8080/v1"},
		{name: "zai localhost", provider: "zai", raw: "https://localhost:8080/v1"},
		{name: "zai metadata service", provider: "zai", raw: "http://169.254.169.254/latest/meta-data"},
		{name: "deepseek loopback", provider: "deepseek", raw: "https://127.0.0.1:8080/v1"},
		{name: "deepseek localhost", provider: "deepseek", raw: "https://localhost:8080/v1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateProviderAPIBase(tt.provider, tt.raw)

			require.Error(t, err)
			assert.ErrorIs(t, err, ErrInvalidProviderAPIBase)
		})
	}
}
