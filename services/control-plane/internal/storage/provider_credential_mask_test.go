package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaskedCredentialPlaceholderIsRejected(t *testing.T) {
	assert.True(t, isMaskedCredentialValue("sk-••••••••1234"))
	assert.True(t, isMaskedCredentialValue("********"))
	assert.False(t, isMaskedCredentialValue("sk-real-key-1234"))
}

func TestMaskCredentialHintDoesNotExposeFullSecret(t *testing.T) {
	assert.Equal(t, "sk-••••••••abcd", maskCredentialHint("sk-test-secret-abcd"))
	assert.Equal(t, "••••", maskCredentialHint("short"))
	assert.Equal(t, "secr••••••••name", maskCredentialHint("secret/provider/name"))
}
