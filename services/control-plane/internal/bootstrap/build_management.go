package bootstrap

import (
	"github.com/Delpak/shahrah/services/control-plane/internal/admin"
	"github.com/Delpak/shahrah/services/control-plane/internal/management"
	"github.com/Delpak/shahrah/services/control-plane/internal/supportfiles"
	"go.uber.org/zap"
)

func (b *builder) buildManagement() error {
	supportFileService, err := supportfiles.NewFromEnv()
	if err != nil {
		b.logger.Warn("Support attachment service is unavailable", zap.Error(err))
	}

	dependencies := admin.Dependencies{
		IAM: admin.IAMDependencies{
			Keys:         b.store,
			OrgUnits:     b.store,
			Memberships:  b.store,
			Users:        b.store,
			VirtualUsers: b.store,
			Teams:        b.store,
			Roles:        b.store,
		},
		Catalog: admin.CatalogDependencies{
			Models:              b.store,
			ProviderAccounts:    b.store.ProviderAccountRepository,
			ProviderSchemas:     b.credentialRuntime.schemas,
			ProviderCredentials: b.credentialRuntime.service,
			Prompts:             b.store,
		},
		Billing: admin.BillingDependencies{
			Wallets:      b.store.WalletRepository,
			Plans:        b.store.PlanRepository,
			Entitlements: b.store.PlanRepository,
		},
		Onboarding: admin.OnboardingDependencies{
			Onboarding:     b.store,
			LoginDiscovery: b.store,
		},
		Observability: admin.ObservabilityDependencies{
			Traces:         b.store,
			Logs:           b.store,
			Metrics:        b.store,
			TracePrivacy:   b.store,
			RequestLogging: b.store,
		},
		Support: admin.SupportDependencies{
			Tenants: b.store.SupportRepository,
			Tickets: b.store.SupportRepository,
			Files:   supportFileService,
		},
	}

	adminHandlers := admin.New(dependencies)
	authHandlers := admin.NewAuthHandlers(b.store, admin.NewTenantAdminOTPNotifierFromEnv())

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
