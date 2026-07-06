package runtimeprofile

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

func formatVersion(version int64) string {
	return strconv.FormatInt(version, 10)
}

func parseVersion(value string) (int64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, fmt.Errorf("profile version is empty")
	}
	return strconv.ParseInt(value, 10, 64)
}

func cloneProfile(in *Profile) *Profile {
	if in == nil {
		return nil
	}
	out := *in
	out.AuthByTokenHash = make(map[string]platform.AuthContext, len(in.AuthByTokenHash))
	for key, value := range in.AuthByTokenHash {
		out.AuthByTokenHash[key] = cloneAuthContext(value)
	}
	out.VirtualModels = make(map[string]platform.VirtualModel, len(in.VirtualModels))
	for key, value := range in.VirtualModels {
		out.VirtualModels[key] = cloneVirtualModel(value)
	}
	out.Deployments = cloneDeployments(in.Deployments)
	out.Guardrails = make(map[string]platform.TenantGuardrail, len(in.Guardrails))
	for key, value := range in.Guardrails {
		out.Guardrails[key] = cloneGuardrail(value)
	}
	if in.Entitlements != nil {
		entitlements := *in.Entitlements
		entitlements.Features = cloneInterfaceMap(in.Entitlements.Features)
		if in.Entitlements.Subscription != nil {
			subscription := *in.Entitlements.Subscription
			subscription.PriceSnapshot = cloneInterfaceMap(subscription.PriceSnapshot)
			subscription.Metadata = cloneInterfaceMap(subscription.Metadata)
			entitlements.Subscription = &subscription
		}
		out.Entitlements = &entitlements
	}
	out.RateLimitRules = append([]platform.RateLimitRule(nil), in.RateLimitRules...)
	for i := range out.RateLimitRules {
		out.RateLimitRules[i].Conditions = cloneInterfaceMap(out.RateLimitRules[i].Conditions)
		out.RateLimitRules[i].Limit = cloneInterfaceMap(out.RateLimitRules[i].Limit)
	}
	out.BudgetRules = append([]platform.BudgetLimitRule(nil), in.BudgetRules...)
	for i := range out.BudgetRules {
		out.BudgetRules[i].Conditions = cloneInterfaceMap(out.BudgetRules[i].Conditions)
		out.BudgetRules[i].Alerts = cloneInterfaceMap(out.BudgetRules[i].Alerts)
	}
	out.ProviderAPIKeys = make(map[string]string, len(in.ProviderAPIKeys))
	for key, value := range in.ProviderAPIKeys {
		out.ProviderAPIKeys[key] = value
	}
	return &out
}

func cloneAuthContext(in platform.AuthContext) platform.AuthContext {
	out := in
	out.Permissions = append([]string(nil), in.Permissions...)
	out.TeamMemberPermissions = append([]string(nil), in.TeamMemberPermissions...)
	out.AllowedModels = append([]string(nil), in.AllowedModels...)
	out.AllowedModelTypes = cloneStringSliceMap(in.AllowedModelTypes)
	out.AllowedModelPricingMode = cloneStringMap(in.AllowedModelPricingMode)
	out.AllowedModelPricingResolved = cloneBoolMap(in.AllowedModelPricingResolved)
	out.VirtualModelSnapshots = make(map[string]platform.VirtualModelSnapshot, len(in.VirtualModelSnapshots))
	for key, value := range in.VirtualModelSnapshots {
		value.TargetFQNs = append([]string(nil), value.TargetFQNs...)
		value.ExposedModes = append([]string(nil), value.ExposedModes...)
		out.VirtualModelSnapshots[key] = value
	}
	return out
}

func cloneDeployments(in []platform.Deployment) []platform.Deployment {
	out := append([]platform.Deployment(nil), in...)
	for i := range out {
		out[i].ModelInfo = cloneInterfaceMap(out[i].ModelInfo)
	}
	return out
}

func cloneVirtualModel(in platform.VirtualModel) platform.VirtualModel {
	out := in
	out.ExposedModes = append([]string(nil), in.ExposedModes...)
	out.Targets = append([]platform.VirtualModelTarget(nil), in.Targets...)
	for i := range out.Targets {
		out.Targets[i].Policy.Retry.StatusCodes = append([]int(nil), in.Targets[i].Policy.Retry.StatusCodes...)
		out.Targets[i].Policy.Fallback.StatusCodes = append([]int(nil), in.Targets[i].Policy.Fallback.StatusCodes...)
	}
	out.Access.Managers = append([]platform.AccessPrincipal(nil), in.Access.Managers...)
	out.Access.Users = append([]platform.AccessPrincipal(nil), in.Access.Users...)
	out.Access.Rules = append([]platform.TenantProviderAccountAccessRule(nil), in.Access.Rules...)
	return out
}

func cloneGuardrail(in platform.TenantGuardrail) platform.TenantGuardrail {
	out := in
	out.Config = cloneInterfaceMap(in.Config)
	return out
}

func cloneStringSliceMap(in map[string][]string) map[string][]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string][]string, len(in))
	for key, value := range in {
		out[key] = append([]string(nil), value...)
	}
	return out
}

func cloneStringMap(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

func cloneBoolMap(in map[string]bool) map[string]bool {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]bool, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

func cloneInterfaceMap(in map[string]interface{}) map[string]interface{} {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]interface{}, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}
