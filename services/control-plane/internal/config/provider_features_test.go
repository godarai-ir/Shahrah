package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadProviderFeatureSettingsFailsClosedForInvalidExplicitFlags(t *testing.T) {
	t.Setenv("SHAHRAH_MISTRAL_FILES_ENABLED", "invalid")

	settings := LoadProviderFeatureSettings()

	require.False(t, settings.MistralFilesEnabled)
}
