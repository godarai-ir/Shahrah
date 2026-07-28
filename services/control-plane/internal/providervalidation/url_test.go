package providervalidation

import (
	"errors"
	"reflect"
	"testing"
)

func TestValidateBaseURLAllowsEverySupportedProviderHost(t *testing.T) {
	t.Parallel()

	tests := []struct {
		provider string
		raw      string
	}{
		{provider: "anthropic", raw: "https://api.anthropic.com"},
		{provider: "azure-open-ai", raw: "https://resource.openai.azure.com"},
		{provider: "cerebras", raw: "https://api.cerebras.ai/v1"},
		{provider: "cohere", raw: "https://api.cohere.com"},
		{provider: "databricks", raw: "https://workspace.cloud.databricks.com/serving-endpoints/model/invocations"},
		{provider: "deepinfra", raw: "https://api.deepinfra.com/v1/openai"},
		{provider: "google-gemini", raw: "https://generativelanguage.googleapis.com/v1beta"},
		{provider: "groq", raw: "https://api.groq.com/openai/v1"},
		{provider: "mistral-ai", raw: "https://api.mistral.ai/v1"},
		{provider: "openai", raw: "https://api.openai.com/v1"},
		{provider: "openrouter", raw: "https://openrouter.ai/api/v1"},
		{provider: "together-ai", raw: "https://api.together.xyz/v1"},
		{provider: "xai", raw: "https://api.x.ai/v1"},
	}

	for _, test := range tests {
		test := test
		t.Run(test.provider, func(t *testing.T) {
			t.Parallel()

			if _, err := ValidateBaseURL(test.provider, test.raw); err != nil {
				t.Fatalf("ValidateBaseURL(%q, %q): %v", test.provider, test.raw, err)
			}
		})
	}
}

func TestValidateBaseURLRejectsUnsafeTargets(t *testing.T) {
	t.Parallel()

	tests := []string{
		"http://api.openai.com/v1",
		"https://127.0.0.1/v1",
		"https://localhost/v1",
		"https://api.evil.example/v1",
		"https://token@api.openai.com/v1",
		"https://api.openai.com:8443/v1",
		"https://api.openai.com/v1?key=secret",
	}

	for _, raw := range tests {
		if _, err := ValidateBaseURL("openai", raw); !errors.Is(err, ErrInvalidBaseURL) {
			t.Fatalf("expected ErrInvalidBaseURL for %q, got %v", raw, err)
		}
	}
}

func TestNormalizeBaseURLConfigValidatesEveryAlias(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		provider string
		key      string
		raw      string
		want     string
	}{
		{
			name:     "Azure endpoint",
			provider: "azure-open-ai",
			key:      "endpoint",
			raw:      " HTTPS://RESOURCE.OPENAI.AZURE.COM/ ",
			want:     "https://resource.openai.azure.com",
		},
		{
			name:     "Databricks serving endpoint",
			provider: "databricks",
			key:      "serving_endpoint_url",
			raw:      "https://WORKSPACE.CLOUD.DATABRICKS.COM/serving-endpoints/model/invocations/",
			want:     "https://workspace.cloud.databricks.com/serving-endpoints/model/invocations",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			config := map[string]interface{}{test.key: test.raw}
			if err := NormalizeBaseURLConfig(test.provider, config); err != nil {
				t.Fatal(err)
			}
			if got := config[test.key]; got != test.want {
				t.Fatalf("unexpected normalized URL %#v, want %q", got, test.want)
			}
		})
	}
}

func TestNormalizeBaseURLConfigRejectsUnsafeOrInvalidAliasesAtomically(t *testing.T) {
	t.Parallel()

	tests := []map[string]interface{}{
		{"endpoint": "https://127.0.0.1"},
		{"serving_endpoint_url": "https://metadata.google.internal"},
		{"base_url": 42},
	}

	for _, config := range tests {
		before := make(map[string]interface{}, len(config)+1)
		for key, value := range config {
			before[key] = value
		}
		before["api_base"] = " HTTPS://API.OPENAI.COM/V1/ "
		config["api_base"] = before["api_base"]

		if err := NormalizeBaseURLConfig("openai", config); !errors.Is(err, ErrInvalidBaseURL) {
			t.Fatalf("expected ErrInvalidBaseURL for %#v, got %v", before, err)
		}
		if !reflect.DeepEqual(config, before) {
			t.Fatalf("rejected config was partially mutated: got %#v, want %#v", config, before)
		}
	}
}

func TestGeminiVersionIsNotDuplicated(t *testing.T) {
	t.Parallel()

	got, err := ensureGeminiVersionedBase(
		"https://generativelanguage.googleapis.com/v1beta",
		"v1beta",
	)
	if err != nil {
		t.Fatal(err)
	}
	if got != "https://generativelanguage.googleapis.com/v1beta" {
		t.Fatalf("unexpected Gemini base %q", got)
	}
}

func TestAzureDeploymentEndpointIsReducedToResourceModelsEndpoint(t *testing.T) {
	t.Parallel()

	got, err := azureOpenAIModelsURL(
		"https://resource.openai.azure.com/openai/deployments/my-model",
		"2024-10-21",
	)
	if err != nil {
		t.Fatal(err)
	}
	want := "https://resource.openai.azure.com/openai/models?api-version=2024-10-21"
	if got != want {
		t.Fatalf("unexpected Azure models URL %q, want %q", got, want)
	}
}
