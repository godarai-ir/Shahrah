package storage

import (
	"context"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

// applyRuntimeModelPolicy fills the model allowlist of a user- or team-scoped principal.
// When the tenant has enabled provider-account models, the allowlist is the provider policy
// plus the virtual models whose targets it covers; otherwise it is fallbackModels.
func (s *AuthRepository) applyRuntimeModelPolicy(ctx context.Context, auth *platform.AuthContext, fallbackModels []string) error {
	runtimeModels, enforceProviderPolicy, err := s.listRuntimeAllowedModelsByProviderAccount(ctx, auth.TenantID, auth.TeamID, auth.UserID)
	if err != nil {
		return err
	}

	auth.EnforceModelAllowlist = enforceProviderPolicy
	auth.AllowedModelTypes = make(map[string][]string)
	auth.AllowedModelPricingMode = make(map[string]string)

	auth.AllowedModelPricingResolved = make(map[string]bool)
	if enforceProviderPolicy {
		allowedSet := make(map[string]struct{}, len(runtimeModels))

		auth.AllowedModels = make([]string, 0, len(runtimeModels))
		for _, item := range dedupeRuntimePolicies(runtimeModels) {
			auth.AllowedModels = append(auth.AllowedModels, item.Model)
			allowedSet[item.Model] = struct{}{}
			auth.AllowedModelTypes[item.Model] = append([]string(nil), item.Types...)
			auth.AllowedModelPricingMode[item.Model] = item.PricingMode
			auth.AllowedModelPricingResolved[item.Model] = item.PricingResolved
		}

		virtualModels, err := s.ListVirtualModelsForPrincipal(ctx, auth.TenantID, auth.UserID, auth.TeamID)
		if err != nil {
			return err
		}

		for i := range virtualModels {
			vm := &virtualModels[i]

			alias := strings.TrimSpace(vm.Alias)
			if alias == "" {
				continue
			}

			valid := true

			for i := range vm.Targets {
				target := &vm.Targets[i]

				fqn := normalizeModelFQN(target.Model)
				if _, ok := runtimePolicyByModel(runtimeModels, fqn); !ok {
					valid = false
					break
				}
			}

			if !valid {
				continue
			}

			aliasTypes := normalizeStringList(vm.ExposedModes)
			if len(aliasTypes) == 0 {
				continue
			}

			if _, exists := allowedSet[alias]; exists {
				continue
			}

			allowedSet[alias] = struct{}{}
			auth.AllowedModels = append(auth.AllowedModels, alias)
			auth.AllowedModelTypes[alias] = aliasTypes
			auth.AllowedModelPricingMode[alias] = "virtual"
			auth.AllowedModelPricingResolved[alias] = true
		}
	} else {
		auth.AllowedModels = fallbackModels
	}

	return nil
}
