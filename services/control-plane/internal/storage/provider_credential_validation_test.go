package storage

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateProviderAPIKeyCallsProviderModelsEndpoint(t *testing.T) {
	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		assert.Equal(t, "/models", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	defer server.Close()

	result := validateProviderAPIKey(context.Background(), "openai", "sk-real-key-1234", map[string]interface{}{
		"api_base": server.URL,
	}, server.Client())

	assert.True(t, result.Success)
	assert.Equal(t, "Bearer sk-real-key-1234", gotAuth)
	assert.Contains(t, result.Message, "accepted")
}

func TestValidateProviderAPIKeySupportsSambaNova(t *testing.T) {
	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		assert.Equal(t, "/models", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	defer server.Close()

	result := validateProviderAPIKey(context.Background(), "sambanova", "sbnv-real-key-1234", map[string]interface{}{
		"api_base": server.URL,
	}, server.Client())

	assert.True(t, result.Success)
	assert.Equal(t, "Bearer sbnv-real-key-1234", gotAuth)
}

func TestValidateProviderAPIKeyRejectsInvalidProviderKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":{"message":"Incorrect API key provided: sk-real-key-1234"}}`))
	}))
	defer server.Close()

	result := validateProviderAPIKey(context.Background(), "openai", "sk-real-key-1234", map[string]interface{}{
		"api_base": server.URL,
	}, server.Client())

	assert.False(t, result.Success)
	assert.Contains(t, result.Message, "invalid")
	assert.NotContains(t, result.Message, "sk-real-key-1234")
	assert.Contains(t, result.Message, "[redacted]")
}

func TestValidateProviderAPIKeyClassifiesForbiddenAndQuota(t *testing.T) {
	for name, status := range map[string]int{
		"forbidden": http.StatusForbidden,
		"quota":     http.StatusTooManyRequests,
	} {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(status)
				_, _ = w.Write([]byte(`{"message":"provider detail"}`))
			}))
			defer server.Close()

			result := validateProviderAPIKey(context.Background(), "mistral", "mistral-key-1234", map[string]interface{}{
				"api_base": server.URL,
			}, server.Client())

			assert.False(t, result.Success)
			if status == http.StatusForbidden {
				assert.Contains(t, result.Message, "denied access")
			} else {
				assert.Contains(t, result.Message, "quota")
			}
			assert.Contains(t, result.Message, "provider detail")
		})
	}
}

func TestValidateProviderAPIKeyRejectsMaskedPlaceholder(t *testing.T) {
	result := validateProviderAPIKey(context.Background(), "openai", "sk-••••••••1234", nil, nil)

	assert.False(t, result.Success)
	assert.Contains(t, result.Message, "masked")
}

func TestMaskCredentialHintDoesNotExposeFullSecret(t *testing.T) {
	assert.Equal(t, "sk-••••••••abcd", maskCredentialHint("sk-test-secret-abcd"))
	assert.Equal(t, "••••", maskCredentialHint("short"))
	assert.Equal(t, "secr••••••••name", maskCredentialHint("secret/provider/name"))
}
