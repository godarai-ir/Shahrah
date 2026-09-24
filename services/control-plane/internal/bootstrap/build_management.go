package bootstrap

import (
	"github.com/Delpak/shahrah/services/control-plane/internal/admin"
	"github.com/Delpak/shahrah/services/control-plane/internal/management"
	"github.com/Delpak/shahrah/services/control-plane/internal/supportfiles"
	"go.uber.org/zap"
)

func (b *builder) buildManagement() error {
	supportFileService, err := supportfiles.NewFromConfig(supportfiles.Config{
		Endpoint:       b.cfg.SupportFiles.Endpoint,
		Bucket:         b.cfg.SupportFiles.Bucket,
		AccessKey:      b.cfg.SupportFiles.AccessKey,
		SecretKey:      b.cfg.SupportFiles.SecretKey,
		Region:         b.cfg.SupportFiles.Region,
		ForcePathStyle: b.cfg.SupportFiles.ForcePathStyle,
		ClamAVAddress:  b.cfg.SupportFiles.ClamAVAddress,
		ClamAVTimeout:  b.cfg.SupportFiles.ClamAVTimeout,
		AllowUnscanned: b.cfg.SupportFiles.AllowUnscanned,
	})
	if err != nil {
		b.logger.Warn("Support attachment service is unavailable", zap.Error(err))
	}

	dependencies := admin.Dependencies{
		IAM: admin.IAMDependencies{
			Keys:         b.iamApplication,
			OrgUnits:     b.iamApplication,
			Memberships:  b.iamApplication,
			Users:        b.tenantAdminApplication,
			VirtualUsers: b.iamApplication,
			Teams:        b.iamApplication,
			Roles:        b.iamApplication,
		},
		Catalog: admin.CatalogDependencies{
			Models:              b.modelApplication,
			Policies:            b.policyApplication,
			ProviderAccounts:    b.providerAccountApplication,
			ProviderSchemas:     b.credentialRuntime.schemas,
			ProviderCredentials: b.credentialRuntime.service,
			Prompts:             b.promptApplication,
			SemanticGuardrails:  b.guardrailEngine(),
			Discovery: admin.DiscoveryConfig{
				BaseURL:  b.cfg.Admin.Discovery.BaseURL,
				Token:    b.cfg.Admin.Discovery.Token,
				CacheTTL: b.cfg.Admin.Discovery.CacheTTL,
			},
		},
		Billing: admin.BillingDependencies{
			Wallets:      b.store.WalletRepository,
			Plans:        b.store.PlanRepository,
			Entitlements: b.store.PlanRepository,
			Payment: admin.PaymentConfig{
				FallbackUSDtoIRR:            b.cfg.BillingSettings.DefaultUSDtoIRR,
				DefaultSubscriptionProvider: b.cfg.Admin.Payment.DefaultSubscriptionProvider,
				ZarinpalMerchantID:          b.cfg.Admin.Payment.ZarinpalMerchantID,
				ZarinpalBaseURL:             b.cfg.Admin.Payment.ZarinpalBaseURL,
				IDPayAPIKey:                 b.cfg.Admin.Payment.IDPayAPIKey,
				IDPayBaseURL:                b.cfg.Admin.Payment.IDPayBaseURL,
				IDPaySandbox:                b.cfg.Admin.Payment.IDPaySandbox,
			},
		},
		Onboarding: admin.OnboardingDependencies{
			Onboarding:     b.tenantAdminApplication,
			LoginDiscovery: b.store.TenantRepository,
		},
		Observability: admin.ObservabilityDependencies{
			Traces:         b.store.TraceRepository,
			Logs:           b.store.ObservabilityRepository,
			Metrics:        b.store.MetricsRepository,
			TracePrivacy:   b.store.ObservabilitySettingsRepository,
			RequestLogging: b.store.ObservabilitySettingsRepository,
		},
		Support: admin.SupportDependencies{
			Tenants: b.store.SupportRepository,
			Tickets: b.store.SupportRepository,
			Files:   supportFileService,
			Limits: admin.SupportLimitsConfig{
				TicketsPerUserPerHour:           b.cfg.Admin.SupportLimits.TicketsPerUserPerHour,
				MessagesPerUserPerTicketPerHour: b.cfg.Admin.SupportLimits.MessagesPerUserPerTicketPerHour,
			},
		},
	}

	adminHandlers := admin.New(dependencies)
	authHandlers := admin.NewAuthHandlersWithConfig(b.tenantAdminApplication, admin.NewTenantAdminOTPNotifier(
		admin.OTPEmailConfig{
			Production:      b.cfg.Admin.OTPEmail.Production,
			Enabled:         b.cfg.Admin.OTPEmail.Enabled,
			SMTPHost:        b.cfg.Admin.OTPEmail.SMTPHost,
			SMTPPort:        b.cfg.Admin.OTPEmail.SMTPPort,
			SMTPFrom:        b.cfg.Admin.OTPEmail.SMTPFrom,
			SMTPUser:        b.cfg.Admin.OTPEmail.SMTPUser,
			SMTPPassword:    b.cfg.Admin.OTPEmail.SMTPPassword,
			ResendAPIKey:    b.cfg.Admin.OTPEmail.ResendAPIKey,
			ResendFromEmail: b.cfg.Admin.OTPEmail.ResendFromEmail,
			ResendFromName:  b.cfg.Admin.OTPEmail.ResendFromName,
			ResendReplyTo:   b.cfg.Admin.OTPEmail.ResendReplyTo,
			Brand: admin.OTPBrandConfig{
				LogoURL:          b.cfg.Admin.OTPEmail.LogoURL,
				WorkspaceURL:     b.cfg.Admin.OTPEmail.WorkspaceURL,
				TenantBaseDomain: b.cfg.Admin.OTPEmail.TenantBaseDomain,
			},
		},
		admin.OTPMobileConfig{
			Enabled:      b.cfg.Admin.OTPMobile.Enabled,
			APIAccessKey: b.cfg.Admin.OTPMobile.APIAccessKey,
			BotID:        b.cfg.Admin.OTPMobile.BotID,
		},
	), admin.AuthHandlersConfig{ExposeDebugOTP: b.cfg.Admin.ExposeDebugOTP})

	managementRoutes, err := management.New(management.Dependencies{
		Auth:          authHandlers,
		Billing:       adminHandlers,
		Catalog:       adminHandlers,
		IAM:           adminHandlers,
		Observability: adminHandlers,
		Onboarding:    adminHandlers,
		Support:       adminHandlers,
	})
	if err != nil {
		return err
	}

	b.managementRoutes = managementRoutes

	return nil
}
