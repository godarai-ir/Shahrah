package storage

import (
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/stretchr/testify/assert"
)

func TestShouldClearProviderAccountCredentials(t *testing.T) {
	assert.True(t, shouldClearProviderAccountCredentials(string(platform.ProviderAuthModeNoCredential), false, nil))
	assert.True(t, shouldClearProviderAccountCredentials(string(platform.ProviderAuthModeAPIKey), true, nil))
	assert.True(t, shouldClearProviderAccountCredentials(string(platform.ProviderAuthModeAPIKey), false, &platform.ProviderAccountCredentials{SkipRequired: true}))

	assert.False(t, shouldClearProviderAccountCredentials(string(platform.ProviderAuthModeAPIKey), false, &platform.ProviderAccountCredentials{APIKey: "sk-test"}))
	assert.False(t, shouldClearProviderAccountCredentials(string(platform.ProviderAuthModeAPIKey), false, nil))
}

func TestProviderAccountCredentialMapHasCredential(t *testing.T) {
	assert.False(t, providerAccountCredentialMapHasCredential(map[string]interface{}{}))
	assert.False(t, providerAccountCredentialMapHasCredential(map[string]interface{}{"api_key": "  "}))

	assert.True(t, providerAccountCredentialMapHasCredential(map[string]interface{}{"api_key": "sk-test"}))
	assert.True(t, providerAccountCredentialMapHasCredential(map[string]interface{}{"secret_ref": "vault/provider"}))
	assert.True(t, providerAccountCredentialMapHasCredential(map[string]interface{}{"raw": map[string]interface{}{"client_email": "svc@example.com"}}))
}
