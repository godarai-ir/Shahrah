package proxy

import (
	"context"
	"strconv"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/gofiber/fiber/v2"
)

func deploymentCredentialSource(dep platform.Deployment) string {
	source := strings.TrimSpace(dep.CredentialSource)
	switch source {
	case platform.CredentialSourceUserBYOK:
		return platform.CredentialSourceUserBYOK
	case platform.CredentialSourceGodarManagedKey, platform.CredentialSourcePlatformManaged, platform.CredentialSourceEnvBreakGlass:
		return source
	default:
		if strings.TrimSpace(dep.BillingMode) == platform.BillingModeExternalBilled {
			return platform.CredentialSourceUserBYOK
		}

		return platform.CredentialSourceGodarManagedKey
	}
}

func deploymentBillingMode(dep platform.Deployment) string {
	if strings.EqualFold(strings.TrimSpace(dep.Provider), "mock") {
		return platform.BillingModeExternalBilled
	}

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

func shouldDeferBudgetReservationUntilRoute(_ string) bool {
	// Effective pricing belongs to a concrete deployment, so wallet reservation
	// must never happen before routing has selected that deployment.
	return true
}

func withTokenUsageMeasures(breakdown platform.UsageCostBreakdown, inputTokens, outputTokens int) platform.UsageCostBreakdown {
	breakdown.InputTokens = inputTokens
	breakdown.OutputTokens = outputTokens
	breakdown.TotalTokens = inputTokens + outputTokens
	breakdown.UsageMeasures = nil
	if inputTokens > 0 {
		breakdown.UsageMeasures = append(breakdown.UsageMeasures, platform.UsageMeasure{
			Usage: "input", Unit: "token", Amount: strconv.Itoa(inputTokens), ServiceTier: "standard",
		})
	}
	if outputTokens > 0 {
		breakdown.UsageMeasures = append(breakdown.UsageMeasures, platform.UsageMeasure{
			Usage: "output", Unit: "token", Amount: strconv.Itoa(outputTokens), ServiceTier: "standard",
		})
	}
	return breakdown
}

type deploymentUsageRater interface {
	RateDeploymentUsage(context.Context, platform.Deployment, platform.UsageCostBreakdown) (platform.UsageCostBreakdown, error)
}

func rateDeploymentEstimate(c *fiber.Ctx, service budgeter, dep platform.Deployment, estimated platform.UsageCostBreakdown) (platform.UsageCostBreakdown, error) {
	rater, ok := service.(deploymentUsageRater)
	if !ok {
		// Test doubles and intentionally minimal embedders may not implement
		// deployment rating. The production billing service always does.
		return estimated, nil
	}

	rated, err := rater.RateDeploymentUsage(c.Context(), dep, estimated)
	if err != nil {
		return platform.UsageCostBreakdown{}, writeAPIError(
			c,
			fiber.StatusUnprocessableEntity,
			"Effective pricing is unavailable for the selected deployment",
			"pricing_unresolved",
		)
	}
	return rated, nil
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

	if value := strings.TrimSpace(dep.ConnectionID); value != "" {
		out["connection_id"] = value
	}

	if value := strings.TrimSpace(dep.CredentialVersionID); value != "" {
		out["credential_version_id"] = value
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
