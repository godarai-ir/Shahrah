package proxy

import (
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

func isModelAllowed(authInfo *platform.AuthContext, model string) bool {
	if authInfo == nil {
		return false
	}
	if authInfo.EnforceModelAllowlist {
		return containsModel(authInfo.AllowedModels, model) || containsModel(authInfo.AllowedModels, "*")
	}
	if len(authInfo.AllowedModels) == 0 {
		return true
	}
	return containsModel(authInfo.AllowedModels, model) || containsModel(authInfo.AllowedModels, "*")
}

func denyReasonForModel(authInfo *platform.AuthContext) string {
	if authInfo == nil {
		return denyReasonKeyAllowlist
	}
	if authInfo.EnforceModelAllowlist {
		return denyReasonProviderPolicy
	}
	return denyReasonKeyAllowlist
}

func isModelAllowedForType(authInfo *platform.AuthContext, model string, expected platform.ModelType) bool {
	if !isModelAllowed(authInfo, model) {
		return false
	}
	if authInfo == nil || expected == "" {
		return true
	}
	if len(authInfo.AllowedModelTypes) == 0 {
		// Backward-compatible fallback for keys that don't have model-type policy loaded yet.
		return true
	}
	types, ok := authInfo.AllowedModelTypes[model]
	if !ok || len(types) == 0 {
		return false
	}
	wanted := strings.TrimSpace(string(expected))
	for _, t := range types {
		if strings.TrimSpace(t) == wanted {
			return true
		}
	}
	return false
}
