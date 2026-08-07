package bootstrap

import (
	"github.com/Delpak/shahrah/services/control-plane/internal/admin"
	"github.com/Delpak/shahrah/services/control-plane/internal/supportfiles"
	"go.uber.org/zap"
)

func (b *builder) buildManagement() error {
	supportFileService, err := supportfiles.NewFromEnv()
	if err != nil {
		b.logger.Warn("Support attachment service is unavailable", zap.Error(err))
	}

	dependencies := admin.Dependencies{
		Keys:                b.store,
		Models:              b.store,
		Wallets:             b.store.WalletRepository,
		ProviderAccounts:    b.store.ProviderAccountRepository,
		ProviderSchemas:     b.credentialRuntime.schemas,
		ProviderCredentials: b.credentialRuntime.service,
		OrgUnits:            b.store,
		Memberships:         b.store,
		Users:               b.store,
		VirtualUsers:        b.store,
		Teams:               b.store,
		Onboarding:          b.store,
		LoginDiscovery:      b.store,
		Traces:              b.store,
		Logs:                b.store,
		Metrics:             b.store,
		TracePrivacy:        b.store,
		RequestLogging:      b.store,
		Prompts:             b.store,
		Roles:               b.store,
		Plans:               b.store.PlanRepository,
		Entitlements:        b.store.PlanRepository,
		SupportTenants:      b.store.SupportRepository,
		SupportTickets:      b.store.SupportRepository,
		SupportFiles:        supportFileService,
	}

	b.adminHandlers = admin.New(dependencies)
	b.authHandlers = admin.NewAuthHandlers(b.store, admin.NewTenantAdminOTPNotifierFromEnv())

	return nil
}
