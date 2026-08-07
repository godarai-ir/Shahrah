package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// AdminSettings contains immutable tenant-admin delivery and presentation settings.
type AdminSettings struct {
	OTPEmail       AdminOTPEmailSettings
	OTPMobile      AdminOTPMobileSettings
	Payment        AdminPaymentSettings
	SupportLimits  AdminSupportLimitsSettings
	ExposeDebugOTP bool
	Discovery      AdminDiscoverySettings
}

// AdminOTPEmailSettings configures tenant-admin OTP email delivery.
type AdminOTPEmailSettings struct {
	Production       bool
	Enabled          bool
	SMTPHost         string
	SMTPPort         int
	SMTPFrom         string
	SMTPUser         string
	SMTPPassword     string
	ResendAPIKey     string
	ResendFromEmail  string
	ResendFromName   string
	ResendReplyTo    string
	LogoURL          string
	WorkspaceURL     string
	TenantBaseDomain string
}

// AdminOTPMobileSettings configures tenant-admin OTP delivery through Bale Safir.
type AdminOTPMobileSettings struct {
	Enabled      bool
	APIAccessKey string
	BotID        int64
}

// AdminPaymentSettings configures management payment gateway adapters.
type AdminPaymentSettings struct {
	DefaultSubscriptionProvider string
	ZarinpalMerchantID          string
	ZarinpalBaseURL             string
	IDPayAPIKey                 string
	IDPayBaseURL                string
	IDPaySandbox                bool
}

// AdminSupportLimitsSettings configures support abuse-control limits.
type AdminSupportLimitsSettings struct {
	TicketsPerUserPerHour           int
	MessagesPerUserPerTicketPerHour int
}

// AdminDiscoverySettings configures the model-discovery management client.
type AdminDiscoverySettings struct {
	BaseURL  string
	Token    string
	CacheTTL time.Duration
}

// LoadAdminSettings reads tenant-admin settings once at the configuration edge.
func LoadAdminSettings() AdminSettings {
	apiAccessKey := strings.TrimSpace(os.Getenv("BALE_SAFIR_API_ACCESS_KEY"))
	botID := int64Env("BALE_SAFIR_BOT_ID", 0)
	mobileEnabled := apiAccessKey != "" || botID > 0

	if _, found := os.LookupEnv("BALE_OTP_ENABLED"); found {
		mobileEnabled = booleanEnv("BALE_OTP_ENABLED", mobileEnabled)
	}

	nodeEnvironment := strings.ToLower(strings.TrimSpace(os.Getenv("NODE_ENV")))

	exposeDebugOTP := false
	if nodeEnvironment == "development" || nodeEnvironment == "test" {
		exposeDebugOTP = booleanEnv("SHAHRAH_EXPOSE_DEBUG_OTP", false)
	}

	return AdminSettings{
		ExposeDebugOTP: exposeDebugOTP,
		OTPEmail: AdminOTPEmailSettings{
			Production:       strings.EqualFold(strings.TrimSpace(os.Getenv("NODE_ENV")), "production"),
			Enabled:          booleanEnv("EMAIL_SENDING_ENABLED", true),
			SMTPHost:         stringEnv("SMTP_HOST", "127.0.0.1"),
			SMTPPort:         positiveIntEnv("SMTP_PORT", 1025),
			SMTPFrom:         stringEnv("SMTP_FROM", "no-reply@shahrah.local"),
			SMTPUser:         strings.TrimSpace(os.Getenv("SMTP_USER")),
			SMTPPassword:     strings.TrimSpace(os.Getenv("SMTP_PASS")),
			ResendAPIKey:     strings.TrimSpace(os.Getenv("RESEND_API_KEY")),
			ResendFromEmail:  strings.TrimSpace(os.Getenv("RESEND_FROM_EMAIL")),
			ResendFromName:   strings.TrimSpace(os.Getenv("RESEND_FROM_NAME")),
			ResendReplyTo:    strings.TrimSpace(os.Getenv("RESEND_REPLY_TO")),
			LogoURL:          firstStringEnv("GODARAI_EMAIL_LOGO_URL", "SHAHRAH_EMAIL_LOGO_URL"),
			WorkspaceURL:     firstStringEnv("NEXT_PUBLIC_TENANT_ADMIN_URL", "SHAHRAH_TENANT_ADMIN_URL"),
			TenantBaseDomain: firstStringEnv("SHAHRAH_TENANT_BASE_DOMAIN", "TENANT_BASE_DOMAIN"),
		},
		OTPMobile: AdminOTPMobileSettings{
			Enabled:      mobileEnabled,
			APIAccessKey: apiAccessKey,
			BotID:        botID,
		},
		Payment: AdminPaymentSettings{
			DefaultSubscriptionProvider: strings.ToLower(stringEnv("SUBSCRIPTION_PAYMENT_PROVIDER", "mock")),
			ZarinpalMerchantID:          strings.TrimSpace(os.Getenv("ZARINPAL_MERCHANT_ID")),
			ZarinpalBaseURL:             stringEnv("ZARINPAL_BASE_URL", "https://api.zarinpal.com"),
			IDPayAPIKey:                 strings.TrimSpace(os.Getenv("IDPAY_API_KEY")),
			IDPayBaseURL:                stringEnv("IDPAY_BASE_URL", "https://api.idpay.ir/v1.1"),
			IDPaySandbox:                booleanEnv("IDPAY_SANDBOX", false),
		},
		SupportLimits: AdminSupportLimitsSettings{
			TicketsPerUserPerHour:           positiveIntEnv("SUPPORT_TICKETS_PER_USER_PER_HOUR", 10),
			MessagesPerUserPerTicketPerHour: positiveIntEnv("SUPPORT_MESSAGES_PER_USER_PER_TICKET_PER_HOUR", 30),
		},
		Discovery: AdminDiscoverySettings{
			BaseURL:  firstStringEnv("SHAHRAH_MODEL_DISCOVERY_URL", "SHAHRAH_DISCOVERY_URL"),
			Token:    strings.TrimSpace(os.Getenv("SHAHRAH_DISCOVERY_TOKEN")),
			CacheTTL: nonNegativeSecondsEnv("SHAHRAH_DISCOVERY_CACHE_TTL_SEC", 5*time.Minute),
		},
	}
}

func firstStringEnv(keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}

	return ""
}

func int64Env(key string, fallback int64) int64 {
	raw := strings.TrimSpace(os.Getenv(key))

	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return fallback
	}

	return value
}
