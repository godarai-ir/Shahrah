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
