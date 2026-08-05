package proxy

import (
	"fmt"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

// runtimeProfilesFromEstimate pins dispatch to the catalog profile used for pricing.
func runtimeProfilesFromEstimate(
	provider, model string,
	estimated platform.UsageCostBreakdown,
) ([]platform.ModelRuntimeProfileBinding, error) {
	provider = strings.ToLower(strings.TrimSpace(provider))
	if !providerRequiresRuntimeProfile(provider) {
		return nil, nil
	}

	providerLabel := runtimeProfileProviderLabel(provider)
	if estimated.ModelRuntimeProfile == nil {
		return nil, fmt.Errorf("%s model runtime profile is unavailable", providerLabel)
	}

	binding := *estimated.ModelRuntimeProfile
	if err := binding.Validate(); err != nil ||
		binding.CatalogRevision != estimated.PricingRevision ||
		!binding.MatchesModel(model) ||
		!runtimeProfileProviderMatches(provider, binding.Provider) {
		return nil, fmt.Errorf("%s model runtime profile is invalid", providerLabel)
	}

	return []platform.ModelRuntimeProfileBinding{binding}, nil
}

func providerRequiresRuntimeProfile(provider string) bool {
	return provider == "anthropic" ||
		provider == "xai" ||
		provider == "moonshot" ||
		provider == deepSeekProviderName ||
		isGoogleGeminiProvider(provider) ||
		isMistralProvider(provider)
}

func runtimeProfileProviderLabel(provider string) string {
	switch {
	case isGoogleGeminiProvider(provider):
		return "Google Gemini"
	case provider == "xai":
		return "xAI"
	case provider == "moonshot":
		return "Moonshot"
	case provider == deepSeekProviderName:
		return "DeepSeek"
	case isMistralProvider(provider):
		return "Mistral AI"
	default:
		return "Anthropic"
	}
}

func runtimeProfileProviderMatches(deploymentProvider, bindingProvider string) bool {
	if strings.EqualFold(
		strings.TrimSpace(deploymentProvider),
		strings.TrimSpace(bindingProvider),
	) {
		return true
	}

	return isGoogleGeminiProvider(deploymentProvider) &&
		isGoogleGeminiProvider(bindingProvider) ||
		isMistralProvider(deploymentProvider) && isMistralProvider(bindingProvider)
}
