package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadProviderFeatureSettingsFailsClosedForInvalidExplicitFlags(t *testing.T) {
	t.Setenv("SHAHRAH_MISTRAL_OCR_ENABLED", "invalid")
	t.Setenv("ANTHROPIC_NATIVE_API_ENABLED", "invalid")

	settings := LoadProviderFeatureSettings()

	require.False(t, settings.MistralEndpoints["ocr"])
	require.False(t, settings.AnthropicNativeEnabled)
}
