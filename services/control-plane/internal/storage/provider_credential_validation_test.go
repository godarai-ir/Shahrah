package storage

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateProviderAPIKeyCallsProviderModelsEndpoint(t *testing.T) {
	var gotAuth string

	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		gotAuth = r.Header.Get("Authorization")
		assert.Equal(t, "/v1/models", r.URL.Path)

		return jsonResponse(http.StatusOK, `{"data":[]}`), nil
	})}

	result := validateProviderAPIKey(context.Background(), "openai", "sk-real-key-1234", map[string]interface{}{
		"api_base": "https://api.openai.com/v1",
	}, client)

	assert.True(t, result.Success)
	assert.Equal(t, "Bearer sk-real-key-1234", gotAuth)
	assert.Contains(t, result.Message, "accepted")
}

func TestValidateProviderAPIKeySupportsSambaNova(t *testing.T) {
	var gotAuth string

	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		gotAuth = r.Header.Get("Authorization")
		assert.Equal(t, "/v1/models", r.URL.Path)

		return jsonResponse(http.StatusOK, `{"data":[]}`), nil
	})}

	result := validateProviderAPIKey(context.Background(), "sambanova", "sbnv-real-key-1234", map[string]interface{}{
		"api_base": "https://api.sambanova.ai/v1",
	}, client)

	assert.True(t, result.Success)
	assert.Equal(t, "Bearer sbnv-real-key-1234", gotAuth)
}

func TestValidateProviderAPIKeyRejectsInvalidProviderKey(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return jsonResponse(http.StatusUnauthorized, `{"error":{"message":"Incorrect API key provided: sk-real-key-1234"}}`), nil
	})}

	result := validateProviderAPIKey(context.Background(), "openai", "sk-real-key-1234", map[string]interface{}{
		"api_base": "https://api.openai.com/v1",
	}, client)

	assert.False(t, result.Success)
	assert.Contains(t, result.Message, "invalid")
	assert.NotContains(t, result.Message, "sk-real-key-1234")
	assert.Contains(t, result.Message, "[redacted]")
}

func TestValidateProviderAPIKeyRejectsPrivateAPIBaseBeforeNetwork(t *testing.T) {
	called := false
	client := &http.Client{
		Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			called = true
			return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}, nil
		}),
	}

	result := validateProviderAPIKey(context.Background(), "openai", "sk-real-key-1234", map[string]interface{}{
		"api_base": "http://127.0.0.1:8080/v1",
	}, client)

	assert.False(t, result.Success)
	assert.Contains(t, result.Message, "api_base")
	assert.False(t, called)
}

func TestValidateProviderAPIKeyClassifiesForbiddenAndQuota(t *testing.T) {
	for name, status := range map[string]int{
		"forbidden": http.StatusForbidden,
		"quota":     http.StatusTooManyRequests,
	} {
		t.Run(name, func(t *testing.T) {
			client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return jsonResponse(status, `{"message":"provider detail"}`), nil
			})}

			result := validateProviderAPIKey(context.Background(), "mistral", "mistral-key-1234", map[string]interface{}{
				"api_base": "https://api.mistral.ai/v1",
			}, client)

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
