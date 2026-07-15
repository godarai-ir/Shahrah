package proxy

import (
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

func deploymentCredentialSource(dep platform.Deployment) string {
	source := strings.TrimSpace(dep.CredentialSource)
	switch source {
	case platform.CredentialSourceUserBYOK:
		return platform.CredentialSourceUserBYOK
	case platform.CredentialSourceGodarManagedKey:
		return platform.CredentialSourceGodarManagedKey
	default:
		if strings.TrimSpace(dep.BillingMode) == platform.BillingModeExternalBilled {
			return platform.CredentialSourceUserBYOK
		}
		return platform.CredentialSourceGodarManagedKey
	}
}

func deploymentBillingMode(dep platform.Deployment) string {
	mode := strings.TrimSpace(dep.BillingMode)
	switch mode {
	case platform.BillingModeExternalBilled:
		return platform.BillingModeExternalBilled
	case platform.BillingModeWalletBilled:
		return platform.BillingModeWalletBilled
	default:
		if deploymentCredentialSource(dep) == platform.CredentialSourceUserBYOK {
			return platform.BillingModeExternalBilled
		}
		return platform.BillingModeWalletBilled
	}
}

func shouldDeferBudgetReservationUntilRoute(model string) bool {
	_, _, _, ok := splitModelFQN(model)
	return ok
}

func hasDeploymentBillingContext(dep platform.Deployment) bool {
	return strings.TrimSpace(dep.ID) != "" ||
		strings.TrimSpace(dep.Provider) != "" ||
		strings.TrimSpace(dep.Model) != "" ||
		strings.TrimSpace(dep.CredentialSource) != "" ||
		strings.TrimSpace(dep.BillingMode) != ""
}

func enrichDeploymentBillingMetadata(input map[string]interface{}, dep platform.Deployment) map[string]interface{} {
	out := map[string]interface{}{}
	for k, v := range input {
		out[k] = v
	}

	out["credential_source"] = deploymentCredentialSource(dep)
	out["billing_mode"] = deploymentBillingMode(dep)

	if value := strings.TrimSpace(dep.ProviderAccountID); value != "" {
		out["provider_account_id"] = value
	}
	if value := strings.TrimSpace(dep.ProviderAccountName); value != "" {
		out["provider_account_name"] = value
	}

	return out
}

func enrichRequestMetadataForDeployment(input map[string]interface{}, dep platform.Deployment) map[string]interface{} {
	return enrichDeploymentBillingMetadata(input, dep)
}

func enrichMetricMetadataForDeployment(input map[string]interface{}, dep platform.Deployment) map[string]interface{} {
	if !hasDeploymentBillingContext(dep) {
		out := map[string]interface{}{}
		for k, v := range input {
			out[k] = v
		}
		return out
	}

	return enrichDeploymentBillingMetadata(input, dep)
}
