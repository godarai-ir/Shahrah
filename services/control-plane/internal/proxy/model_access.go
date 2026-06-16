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
	return modelAccessForType(authInfo, model, expected).Allowed
}

type modelTypeAccessResult struct {
	Allowed           bool
	UnsupportedVMMode bool
	ExposedModes      []string
}

func modelAccessForType(authInfo *platform.AuthContext, model string, expected platform.ModelType) modelTypeAccessResult {
	if !isModelAllowed(authInfo, model) {
		return modelTypeAccessResult{}
	}
	if authInfo == nil || expected == "" {
		return modelTypeAccessResult{Allowed: true}
	}
	if len(authInfo.AllowedModelTypes) == 0 {
		// Backward-compatible fallback for keys that don't have model-type policy loaded yet.
		return modelTypeAccessResult{Allowed: true}
	}
	types, ok := authInfo.AllowedModelTypes[model]
	if !ok || len(types) == 0 {
		return modelTypeAccessResult{}
	}
	wanted := strings.TrimSpace(string(expected))
	for _, t := range types {
		if strings.TrimSpace(t) == wanted {
			return modelTypeAccessResult{Allowed: true, ExposedModes: append([]string(nil), types...)}
		}
	}
	if _, isVirtual := authInfo.VirtualModelSnapshots[model]; isVirtual {
		return modelTypeAccessResult{
			UnsupportedVMMode: true,
			ExposedModes:      append([]string(nil), types...),
		}
	}
	if authInfo.AllowedModelPricingMode != nil && strings.TrimSpace(authInfo.AllowedModelPricingMode[model]) == "virtual" {
		return modelTypeAccessResult{
			UnsupportedVMMode: true,
			ExposedModes:      append([]string(nil), types...),
		}
	}
	return modelTypeAccessResult{ExposedModes: append([]string(nil), types...)}
}
