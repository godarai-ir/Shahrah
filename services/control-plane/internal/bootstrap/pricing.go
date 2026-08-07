package bootstrap

import (
	"context"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/billing"
	"github.com/Delpak/shahrah/services/control-plane/internal/providerregistry"
	"go.uber.org/zap"
)

func (b *builder) configureProviderPricingSnapshots() (*billing.AnthropicPricingSnapshotManager, error) {
	var anthropicPricing *billing.AnthropicPricingSnapshotManager

	for _, provider := range providerregistry.PricingProviders() {
		switch provider.PricingSnapshot {
		case providerregistry.PricingSnapshotAnthropic:
			manager, err := billing.NewAnthropicPricingSnapshotManager(b.cfg.Admin.Discovery.BaseURL, nil)
			if err != nil {
				b.logger.Warn(
					provider.DisplayName+" pricing is unavailable; requests will fail closed until configuration is restored",
					zap.Error(err),
				)

				continue
			}

			b.billing.SetAnthropicPricingSnapshotManager(manager)
			anthropicPricing = manager
			b.reloadPricing(provider, manager.Reload)

			if err := b.lifecycle.AddLoop(provider.Slug+" pricing refresh", func(ctx context.Context) {
				manager.Run(ctx, 5*time.Minute)
			}); err != nil {
				return nil, err
			}

		case providerregistry.PricingSnapshotCatalog:
			manager, err := billing.NewCatalogPricingSnapshotManager(
				provider.Slug,
				b.cfg.Admin.Discovery.BaseURL,
				nil,
			)
			if err != nil {
				b.logger.Warn(
					provider.DisplayName+" pricing is unavailable; requests will fail closed until configuration is restored",
					zap.Error(err),
				)

				continue
			}

			b.billing.SetCatalogPricingSnapshotManager(provider.Slug, manager)
			b.reloadPricing(provider, manager.Reload)

			if err := b.lifecycle.AddLoop(provider.Slug+" pricing refresh", func(ctx context.Context) {
				manager.Run(ctx, 5*time.Minute)
			}); err != nil {
				return nil, err
			}
		}
	}

	return anthropicPricing, nil
}

func (b *builder) reloadPricing(
	provider providerregistry.Definition,
	reload func(context.Context) error,
) {
	if err := reload(b.ctx); err != nil {
		b.logger.Warn(
			"Initial "+provider.DisplayName+" pricing snapshot is unavailable; background refresh will retry",
			zap.Error(err),
		)
	}
}
