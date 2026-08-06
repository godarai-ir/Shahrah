package billing

import (
	"testing"

	"github.com/Delpak/shahrah/services/control-plane/internal/providerregistry"
)

func TestProviderPricingStrategiesCoverRegistry(t *testing.T) {
	declared := map[providerregistry.PricingStrategyKind]bool{
		providerregistry.PricingStrategyStandard: true,
	}

	for _, definition := range providerregistry.Definitions() {
		declared[definition.PricingStrategy] = true

		strategy, ok := providerPricingStrategies[definition.PricingStrategy]
		if !ok || strategy == nil {
			t.Errorf(
				"provider %q has no billing strategy for %q",
				definition.Slug,
				definition.PricingStrategy,
			)
		}
	}

	for kind := range providerPricingStrategies {
		if !declared[kind] {
			t.Errorf("billing strategy %q has no provider definition", kind)
		}
	}
}
