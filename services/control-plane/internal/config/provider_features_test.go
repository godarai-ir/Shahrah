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

// TestTokenBazaarChatAndImagesFlagsDefaultToDisabled asserts all three
// TokenBazaar kill switches default to false when unset, per
// plans/tokenbazaar-production-readiness.md's release gates.
func TestTokenBazaarChatAndImagesFlagsDefaultToDisabled(t *testing.T) {
	settings := LoadProviderFeatureSettings()

	require.False(t, settings.TokenBazaarEnabled)
	require.False(t, settings.TokenBazaarImagesEnabled)
	require.False(t, settings.TokenBazaarEmbeddingsEnabled)
}

// TestTokenBazaarEmbeddingsFlagIsIndependentOfChatAndImagesFlags asserts
// SHAHRAH_TOKENBAZAAR_EMBEDDINGS_ENABLED is read from its own distinct
// environment variable and does not influence, or get influenced by,
// SHAHRAH_TOKENBAZAAR_ENABLED (chat) or SHAHRAH_TOKENBAZAAR_IMAGES_ENABLED
// (images) in either direction, so all three capabilities can be
// canaried/rolled back independently, per the "one model's readiness never
// unlocks another" rule in plans/tokenbazaar-production-readiness.md.
func TestTokenBazaarEmbeddingsFlagIsIndependentOfChatAndImagesFlags(t *testing.T) {
	t.Run("embeddings on, chat and images off", func(t *testing.T) {
		t.Setenv("SHAHRAH_TOKENBAZAAR_EMBEDDINGS_ENABLED", "true")

		settings := LoadProviderFeatureSettings()
		require.True(t, settings.TokenBazaarEmbeddingsEnabled)
		require.False(t, settings.TokenBazaarEnabled)
		require.False(t, settings.TokenBazaarImagesEnabled)
	})

	t.Run("chat and images on, embeddings off", func(t *testing.T) {
		t.Setenv("SHAHRAH_TOKENBAZAAR_ENABLED", "true")
		t.Setenv("SHAHRAH_TOKENBAZAAR_IMAGES_ENABLED", "true")

		settings := LoadProviderFeatureSettings()
		require.True(t, settings.TokenBazaarEnabled)
		require.True(t, settings.TokenBazaarImagesEnabled)
		require.False(t, settings.TokenBazaarEmbeddingsEnabled)
	})

	t.Run("all three on independently", func(t *testing.T) {
		t.Setenv("SHAHRAH_TOKENBAZAAR_ENABLED", "true")
		t.Setenv("SHAHRAH_TOKENBAZAAR_IMAGES_ENABLED", "true")
		t.Setenv("SHAHRAH_TOKENBAZAAR_EMBEDDINGS_ENABLED", "true")

		settings := LoadProviderFeatureSettings()
		require.True(t, settings.TokenBazaarEnabled)
		require.True(t, settings.TokenBazaarImagesEnabled)
		require.True(t, settings.TokenBazaarEmbeddingsEnabled)
	})

	t.Run("invalid embeddings value fails closed without affecting chat or images flags", func(t *testing.T) {
		t.Setenv("SHAHRAH_TOKENBAZAAR_ENABLED", "true")
		t.Setenv("SHAHRAH_TOKENBAZAAR_IMAGES_ENABLED", "true")
		t.Setenv("SHAHRAH_TOKENBAZAAR_EMBEDDINGS_ENABLED", "not-a-bool")

		settings := LoadProviderFeatureSettings()
		require.False(t, settings.TokenBazaarEmbeddingsEnabled)
		require.True(t, settings.TokenBazaarEnabled)
		require.True(t, settings.TokenBazaarImagesEnabled)
	})
}

// TestTokenBazaarImagesFlagIsIndependentOfChatFlag asserts
// SHAHRAH_TOKENBAZAAR_IMAGES_ENABLED and SHAHRAH_TOKENBAZAAR_ENABLED are read
// from distinct environment variables and do not influence each other in
// either direction, so chat and image support can be canaried/rolled back
// independently.
func TestTokenBazaarImagesFlagIsIndependentOfChatFlag(t *testing.T) {
	t.Run("images on, chat off", func(t *testing.T) {
		t.Setenv("SHAHRAH_TOKENBAZAAR_IMAGES_ENABLED", "true")

		settings := LoadProviderFeatureSettings()
		require.True(t, settings.TokenBazaarImagesEnabled)
		require.False(t, settings.TokenBazaarEnabled)
	})

	t.Run("chat on, images off", func(t *testing.T) {
		t.Setenv("SHAHRAH_TOKENBAZAAR_ENABLED", "true")

		settings := LoadProviderFeatureSettings()
		require.True(t, settings.TokenBazaarEnabled)
		require.False(t, settings.TokenBazaarImagesEnabled)
	})

	t.Run("both on independently", func(t *testing.T) {
		t.Setenv("SHAHRAH_TOKENBAZAAR_ENABLED", "true")
		t.Setenv("SHAHRAH_TOKENBAZAAR_IMAGES_ENABLED", "true")

		settings := LoadProviderFeatureSettings()
		require.True(t, settings.TokenBazaarEnabled)
		require.True(t, settings.TokenBazaarImagesEnabled)
	})

	t.Run("invalid images value fails closed without affecting chat flag", func(t *testing.T) {
		t.Setenv("SHAHRAH_TOKENBAZAAR_ENABLED", "true")
		t.Setenv("SHAHRAH_TOKENBAZAAR_IMAGES_ENABLED", "not-a-bool")

		settings := LoadProviderFeatureSettings()
		require.False(t, settings.TokenBazaarImagesEnabled)
		require.True(t, settings.TokenBazaarEnabled)
	})
}
