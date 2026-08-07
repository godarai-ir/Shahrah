package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadAdminSettingsParsesOTPSnapshot(t *testing.T) {
	t.Setenv("NODE_ENV", "production")
	t.Setenv("EMAIL_SENDING_ENABLED", "false")
	t.Setenv("SMTP_PORT", "2525")
	t.Setenv("GODARAI_EMAIL_LOGO_URL", "https://cdn.example/logo.png")
	t.Setenv("BALE_SAFIR_API_ACCESS_KEY", "secret")
	t.Setenv("BALE_SAFIR_BOT_ID", "42")

	settings := LoadAdminSettings()

	require.True(t, settings.OTPEmail.Production)
	require.False(t, settings.OTPEmail.Enabled)
	require.Equal(t, 2525, settings.OTPEmail.SMTPPort)
	require.Equal(t, "https://cdn.example/logo.png", settings.OTPEmail.LogoURL)
	require.True(t, settings.OTPMobile.Enabled)
	require.Equal(t, int64(42), settings.OTPMobile.BotID)
}
