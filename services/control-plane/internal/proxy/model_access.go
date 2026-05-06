package proxy

import "github.com/Delpak/shahrah/packages/contracts/platform"

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
