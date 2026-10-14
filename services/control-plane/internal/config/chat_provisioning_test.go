package config

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func validChatProvisioning() ChatProvisioningSettings {
	return ChatProvisioningSettings{
		Enabled:         true,
		ProvisioningURL: "http://librechat:3081",
		Token:           strings.Repeat("t", 32),
		ChatURLTemplate: "https://{slug}.chat.example.com",
		Interval:        time.Second,
		RequestTimeout:  time.Second,
		BatchSize:       1,
	}
}

func TestLoadChatProvisioningSettingsReadsTheEnvironment(t *testing.T) {
	t.Setenv("CHAT_PROVISIONING_ENABLED", "true")
	t.Setenv("CHAT_PROVISIONING_URL", " http://librechat:3081/ ")
	t.Setenv("CHAT_PROVISIONING_TOKEN", " secret-token ")
	t.Setenv("CHAT_URL_TEMPLATE", "https://{slug}.chat.example.com")
	t.Setenv("CHAT_PROVISIONING_INTERVAL", "30s")
	t.Setenv("CHAT_PROVISIONING_BATCH_SIZE", "4")

	settings := LoadChatProvisioningSettings()

	require.True(t, settings.Enabled)
	require.Equal(t, "http://librechat:3081", settings.ProvisioningURL)
	require.Equal(t, "secret-token", settings.Token)
	require.Equal(t, 30*time.Second, settings.Interval)
	require.Equal(t, 4, settings.BatchSize)
}

func TestLoadChatProvisioningSettingsIsOffByDefault(t *testing.T) {
	t.Setenv("CHAT_PROVISIONING_ENABLED", "")

	require.False(t, LoadChatProvisioningSettings().Enabled)
	require.NoError(t, LoadChatProvisioningSettings().Validate())
}

func TestChatProvisioningSettingsValidate(t *testing.T) {
	t.Parallel()

	require.NoError(t, validChatProvisioning().Validate())

	local := validChatProvisioning()
	local.ChatURLTemplate = "http://{slug}.chat.localhost:3090"
	require.NoError(t, local.Validate())

	cases := map[string]func(*ChatProvisioningSettings){
		"no url":               func(s *ChatProvisioningSettings) { s.ProvisioningURL = "" },
		"url with credentials": func(s *ChatProvisioningSettings) { s.ProvisioningURL = "http://u:p@librechat:3081" },
		"url with query":       func(s *ChatProvisioningSettings) { s.ProvisioningURL = "http://librechat:3081?x=1" },
		"url scheme":           func(s *ChatProvisioningSettings) { s.ProvisioningURL = "ftp://librechat" },
		"short token":          func(s *ChatProvisioningSettings) { s.Token = "short" },
		"no placeholder":       func(s *ChatProvisioningSettings) { s.ChatURLTemplate = "https://chat.example.com" },
		"two placeholders":     func(s *ChatProvisioningSettings) { s.ChatURLTemplate = "https://{slug}.{slug}.example.com" },
		"not the first label":  func(s *ChatProvisioningSettings) { s.ChatURLTemplate = "https://a-{slug}.example.com" },
		"in the path":          func(s *ChatProvisioningSettings) { s.ChatURLTemplate = "https://chat.example.com/{slug}" },
		"http on a public host": func(s *ChatProvisioningSettings) {
			s.ChatURLTemplate = "http://{slug}.chat.example.com"
		},
		"query": func(s *ChatProvisioningSettings) { s.ChatURLTemplate = "https://{slug}.chat.example.com/?a=1" },
	}

	for name, mutate := range cases {
		settings := validChatProvisioning()

		mutate(&settings)

		require.Error(t, settings.Validate(), name)
	}
}
