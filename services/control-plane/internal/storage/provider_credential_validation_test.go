package storage

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateProviderCredentialUsesSchemaBaseURL(t *testing.T) {
	var gotAuth string

	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		gotAuth = r.Header.Get("Authorization")
		assert.Equal(t, "https://api.openai.com/v1/models", r.URL.String())

		return jsonResponse(http.StatusOK, `{"data":[]}`), nil
	})}

	result := validateProviderCredential(
		context.Background(),
		"openai",
		"api_key",
		map[string]interface{}{"base_url": "https://api.openai.com/v1"},
		map[string]string{"api_key": "sk-real-key-1234"},
		client,
	)

	assert.True(t, result.Success)
	assert.Equal(t, "ok", result.Category)
	assert.Equal(t, "Bearer sk-real-key-1234", gotAuth)
}

func TestValidateProviderCredentialSupportsGemini(t *testing.T) {
	var gotGoogleAPIKey string

	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		gotGoogleAPIKey = r.Header.Get("x-goog-api-key")
		assert.Equal(t, "https://generativelanguage.googleapis.com/v1beta/models?pageSize=1", r.URL.String())

		return jsonResponse(http.StatusOK, `{"models":[]}`), nil
	})}

	result := validateProviderCredential(
		context.Background(),
		"google-gemini",
		"api_key",
		map[string]interface{}{
			"base_url":    "https://generativelanguage.googleapis.com/v1beta",
			"api_version": "v1beta",
		},
		map[string]string{"api_key": "gemini-real-key"},
		client,
	)

	assert.True(t, result.Success)
	assert.Equal(t, "gemini-real-key", gotGoogleAPIKey)
}

func TestValidateProviderCredentialSupportsCanonicalMistralSlug(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		assert.Equal(t, "https://api.mistral.ai/v1/models", r.URL.String())

		return jsonResponse(http.StatusOK, `{"data":[]}`), nil
	})}

	result := validateProviderCredential(
		context.Background(),
		"mistral-ai",
		"api_key",
		nil,
		map[string]string{"api_key": "mistral-real-key"},
		client,
	)

	assert.True(t, result.Success)
}

func TestValidateProviderCredentialSupportsDatabricksPAT(t *testing.T) {
	var gotAuth string

	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		gotAuth = r.Header.Get("Authorization")
		assert.Equal(t, "https://workspace.cloud.databricks.com/api/2.0/serving-endpoints", r.URL.String())

		return jsonResponse(http.StatusOK, `{"endpoints":[]}`), nil
	})}

	result := validateProviderCredential(
		context.Background(),
		"databricks",
		"personal_access_token",
		map[string]interface{}{
			"serving_endpoint_url": "https://workspace.cloud.databricks.com/serving-endpoints/model/invocations",
		},
		map[string]string{"access_token": "databricks-real-token"},
		client,
	)

	assert.True(t, result.Success)
	assert.Equal(t, "Bearer databricks-real-token", gotAuth)
}

func TestValidateProviderCredentialRejectsInvalidProviderKeyWithoutLeakingIt(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return jsonResponse(http.StatusUnauthorized, `{"error":{"message":"Incorrect API key provided: sk-real-key-1234"}}`), nil
	})}

	result := validateProviderCredential(
		context.Background(),
		"openai",
		"api_key",
		nil,
		map[string]string{"api_key": "sk-real-key-1234"},
		client,
	)

	assert.False(t, result.Success)
	assert.Equal(t, "invalid_credential", result.Category)
	assert.NotContains(t, result.Message, "sk-real-key-1234")
	assert.Contains(t, result.Message, "[redacted]")
}

func TestValidateProviderCredentialRejectsPrivateBaseBeforeNetwork(t *testing.T) {
	called := false
	client := &http.Client{
		Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			called = true

			return jsonResponse(http.StatusOK, `{}`), nil
		}),
	}

	result := validateProviderCredential(
		context.Background(),
		"openai",
		"api_key",
		map[string]interface{}{"base_url": "https://127.0.0.1:8080/v1"},
		map[string]string{"api_key": "sk-real-key-1234"},
		client,
	)

	assert.False(t, result.Success)
	assert.Equal(t, "configuration", result.Category)
	assert.False(t, called)
}

func TestValidateProviderCredentialTreatsQuotaAsAuthenticated(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return jsonResponse(http.StatusTooManyRequests, `{"message":"quota reached"}`), nil
	})}

	result := validateProviderCredential(
		context.Background(),
		"openai",
		"api_key",
		nil,
		map[string]string{"api_key": "sk-real-key-1234"},
		client,
	)

	assert.True(t, result.Success)
	assert.Equal(t, "quota_exhausted", result.Category)
}

func TestValidateProviderCredentialReturnsUnsupportedWithoutNetwork(t *testing.T) {
	called := false
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		called = true

		return jsonResponse(http.StatusOK, `{}`), nil
	})}

	result := validateProviderCredential(
		context.Background(),
		"deepgram",
		"api_key",
		nil,
		map[string]string{"api_key": "deepgram-key"},
		client,
	)

	assert.True(t, result.Unsupported)
	assert.Equal(t, "unsupported", result.Category)
	assert.False(t, called)
}

func TestProviderCredentialSecretsMergesRawSchemaFields(t *testing.T) {
	secrets := providerCredentialSecrets(map[string]interface{}{
		"api_key": "flattened",
		"raw": map[string]interface{}{
			"access_token": "databricks-token",
			"api_key":      "schema-key",
		},
	})

	assert.Equal(t, "schema-key", secrets["api_key"])
	assert.Equal(t, "databricks-token", secrets["access_token"])
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func jsonResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestValidateProviderCredentialRejectsMaskedPlaceholder(t *testing.T) {
	result := validateProviderCredential(
		context.Background(),
		"openai",
		"api_key",
		nil,
		map[string]string{"api_key": "sk-••••••••1234"},
		nil,
	)

	assert.False(t, result.Success)
	assert.Equal(t, "missing_credential", result.Category)
}

func TestMaskCredentialHintDoesNotExposeFullSecret(t *testing.T) {
	assert.Equal(t, "sk-••••••••abcd", maskCredentialHint("sk-test-secret-abcd"))
	assert.Equal(t, "••••", maskCredentialHint("short"))
	assert.Equal(t, "secr••••••••name", maskCredentialHint("secret/provider/name"))
}
