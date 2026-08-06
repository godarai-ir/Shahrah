package proxy

import (
	"fmt"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/providerregistry"
)

// runtimeProfilesFromEstimate pins dispatch to the catalog profile used for pricing.
func runtimeProfilesFromEstimate(
	provider, model string,
	estimated platform.UsageCostBreakdown,
) ([]platform.ModelRuntimeProfileBinding, error) {
	definition, ok := providerregistry.Lookup(provider)
	if !ok || !definition.DispatchRuntimeProfileRequired {
		return nil, nil
	}

	if estimated.ModelRuntimeProfile == nil {
		return nil, fmt.Errorf("%s model runtime profile is unavailable", definition.DisplayName)
	}

	binding := *estimated.ModelRuntimeProfile
	if err := binding.Validate(); err != nil ||
		binding.CatalogRevision != estimated.PricingRevision ||
		!binding.MatchesModel(model) ||
		!providerregistry.Matches(definition.Slug, binding.Provider) {
		return nil, fmt.Errorf("%s model runtime profile is invalid", definition.DisplayName)
	}

	return []platform.ModelRuntimeProfileBinding{binding}, nil
}
