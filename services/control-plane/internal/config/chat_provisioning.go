package config

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// chatSlugPlaceholder marks where a tenant slug goes in ChatProvisioningSettings.ChatURLTemplate.
const chatSlugPlaceholder = "{slug}"

// minChatProvisioningTokenLength matches the minimum the chat provisioning API enforces.
const minChatProvisioningTokenLength = 32

// ChatProvisioningSettings configures registering every tenant with GodarChat (LibreChat).
type ChatProvisioningSettings struct {
	// Enabled turns on the reconciler that creates each tenant's chat client and registers it.
	Enabled bool
	// ProvisioningURL is the private base URL of the chat provisioning API.
	ProvisioningURL string
	// Token is the bearer token the chat provisioning API expects.
	Token string
	// ChatURLTemplate is a tenant's chat origin with a {slug} placeholder, for example
	// https://{slug}.chat.example.com.
	ChatURLTemplate string
	// Interval is how often the reconciler looks for tenants that need work.
	Interval time.Duration
	// RequestTimeout bounds each call to the provisioning API.
	RequestTimeout time.Duration
	// BatchSize is the most tenants handled per pass.
	BatchSize int
}

// LoadChatProvisioningSettings reads the chat provisioning settings once at the configuration edge.
func LoadChatProvisioningSettings() ChatProvisioningSettings {
	return ChatProvisioningSettings{
		Enabled:         booleanEnv("CHAT_PROVISIONING_ENABLED", false),
		ProvisioningURL: strings.TrimRight(strings.TrimSpace(stringEnv("CHAT_PROVISIONING_URL", "")), "/"),
		Token:           strings.TrimSpace(stringEnv("CHAT_PROVISIONING_TOKEN", "")),
		ChatURLTemplate: strings.TrimSpace(stringEnv("CHAT_URL_TEMPLATE", "")),
		Interval:        positiveDurationEnv("CHAT_PROVISIONING_INTERVAL", 15*time.Second),
		RequestTimeout:  positiveDurationEnv("CHAT_PROVISIONING_REQUEST_TIMEOUT", 15*time.Second),
		BatchSize:       positiveIntEnv("CHAT_PROVISIONING_BATCH_SIZE", 10),
	}
}

// Validate reports why enabled settings cannot work. Disabled settings are always valid.
func (s ChatProvisioningSettings) Validate() error {
	if !s.Enabled {
		return nil
	}

	if err := validateChatProvisioningURL(s.ProvisioningURL); err != nil {
		return err
	}

	if len(s.Token) < minChatProvisioningTokenLength {
		return fmt.Errorf("CHAT_PROVISIONING_TOKEN must be at least %d characters", minChatProvisioningTokenLength)
	}

	return validateChatURLTemplate(s.ChatURLTemplate)
}

func validateChatProvisioningURL(raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" ||
		parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return errors.New("CHAT_PROVISIONING_URL must be an http(s) URL without credentials, query or fragment")
	}

	return nil
}

func validateChatURLTemplate(template string) error {
	if strings.Count(template, chatSlugPlaceholder) != 1 {
		return fmt.Errorf("CHAT_URL_TEMPLATE must contain %s exactly once", chatSlugPlaceholder)
	}

	parsed, err := url.Parse(strings.Replace(template, chatSlugPlaceholder, "tenant", 1))
	if err != nil || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return errors.New("CHAT_URL_TEMPLATE must be an absolute URL without credentials, query or fragment")
	}

	if !strings.HasPrefix(parsed.Hostname(), "tenant.") {
		return fmt.Errorf("CHAT_URL_TEMPLATE must place %s as the whole leftmost host label", chatSlugPlaceholder)
	}

	hostname := parsed.Hostname()
	local := hostname == "localhost" || strings.HasSuffix(hostname, ".localhost")

	switch {
	case parsed.Scheme == "https":
		return nil
	case parsed.Scheme == "http" && local:
		return nil
	default:
		return errors.New("CHAT_URL_TEMPLATE must use https (http is accepted only on .localhost)")
	}
}
