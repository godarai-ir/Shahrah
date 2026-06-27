package admin

import (
	"errors"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/modelid"
)

func normalizeAndValidateModelFQNs(models []string) ([]string, error) {
	normalized := make([]string, 0, len(models))

	seen := make(map[string]struct{}, len(models))
	for _, model := range models {
		model = strings.TrimSpace(model)
		if model == "" {
			continue
		}

		if !isModelFQN(model) {
			return nil, errors.New("models must use provider:account_alias:model_id format")
		}

		if provider, accountAlias, officialModel, ok := modelid.SplitModelFQN(model); ok {
			model = modelid.BuildModelFQN(provider, accountAlias, officialModel)
		}

		if _, ok := seen[model]; ok {
			continue
		}

		seen[model] = struct{}{}
		normalized = append(normalized, model)
	}

	if len(normalized) == 0 {
		return nil, errors.New("at least one allowed model is required")
	}

	return normalized, nil
}

func isVirtualModelInScope(vm platform.VirtualModel, scopeAllowedSet map[string]struct{}) bool {
	if len(vm.Targets) == 0 {
		return false
	}

	for _, target := range vm.Targets {
		targetFQN := normalizeModelFQN(target.Model)
		if targetFQN == "" {
			return false
		}

		if _, allowed := scopeAllowedSet[targetFQN]; !allowed {
			return false
		}
	}

	return true
}

func buildVirtualModelSnapshotsForAliases(aliases []string, virtualModels []platform.VirtualModel, scopeAllowedSet map[string]struct{}) (map[string]platform.VirtualModelSnapshot, error) {
	if len(aliases) == 0 {
		return map[string]platform.VirtualModelSnapshot{}, nil
	}

	byAlias := make(map[string]platform.VirtualModel, len(virtualModels))
	for _, vm := range virtualModels {
		alias := strings.TrimSpace(vm.Alias)
		if alias == "" {
			continue
		}

		byAlias[alias] = vm
	}

	out := make(map[string]platform.VirtualModelSnapshot, len(aliases))
	for _, alias := range aliases {
		alias = strings.TrimSpace(alias)

		vm, ok := byAlias[alias]
		if !ok {
			return nil, errors.New("one or more requested models are not allowed in current tenant scope")
		}

		if !isVirtualModelInScope(vm, scopeAllowedSet) {
			return nil, errors.New("one or more requested virtual models include targets outside your provider scope")
		}

		targetFQNs := make([]string, 0, len(vm.Targets))

		seenTargets := make(map[string]struct{}, len(vm.Targets))
		for _, target := range vm.Targets {
			fqn := normalizeModelFQN(target.Model)
			if fqn == "" {
				continue
			}

			if _, exists := seenTargets[fqn]; exists {
				continue
			}

			seenTargets[fqn] = struct{}{}
			targetFQNs = append(targetFQNs, fqn)
		}

		if len(targetFQNs) == 0 {
			return nil, errors.New("one or more requested virtual models do not have eligible targets")
		}

		out[alias] = platform.VirtualModelSnapshot{
			Alias:          alias,
			VirtualModelID: vm.ID,
			Version:        vm.Version,
			TargetFQNs:     targetFQNs,
			ExposedModes:   append([]string(nil), vm.ExposedModes...),
		}
	}

	return out, nil
}

func isModelFQN(model string) bool {
	parts := strings.Split(model, ":")
	if len(parts) != 3 {
		return false
	}

	for _, part := range parts {
		if strings.TrimSpace(part) == "" {
			return false
		}
	}

	return true
}

func normalizeModelFQN(model string) string {
	model = strings.TrimSpace(model)
	if model == "" {
		return ""
	}

	if provider, accountAlias, officialModel, ok := modelid.SplitModelFQN(model); ok {
		return modelid.BuildModelFQN(provider, accountAlias, officialModel)
	}

	return model
}

func ensureRequestedModelsInScope(requested []string, allowed []string) error {
	if len(allowed) == 0 {
		return errors.New("no allowed models found in current tenant scope for this principal")
	}

	allowedSet := make(map[string]struct{}, len(allowed))
	for _, item := range allowed {
		allowedSet[strings.TrimSpace(item)] = struct{}{}
	}

	invalid := make([]string, 0)

	for _, model := range requested {
		if _, ok := allowedSet[strings.TrimSpace(model)]; !ok {
			invalid = append(invalid, model)
		}
	}

	if len(invalid) > 0 {
		return errors.New("one or more requested models are not allowed in current tenant scope")
	}

	return nil
}
