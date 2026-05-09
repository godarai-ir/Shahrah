package proxy

import "strings"

// splitModelFQN parses a model in provider:account_alias:official_model_id format.
func splitModelFQN(input string) (provider string, accountAlias string, officialModel string, ok bool) {
	parts := strings.Split(strings.TrimSpace(input), ":")
	if len(parts) != 3 {
		return "", "", "", false
	}
	provider = strings.TrimSpace(parts[0])
	accountAlias = strings.TrimSpace(parts[1])
	officialModel = strings.TrimSpace(parts[2])
	if provider == "" || accountAlias == "" || officialModel == "" {
		return "", "", "", false
	}
	return provider, accountAlias, officialModel, true
}

// modelRoutingCandidates returns route candidates for an inbound model id.
// For strict tenant/account isolation, FQN values are routed only as-is.
func modelRoutingCandidates(model string) []string {
	model = strings.TrimSpace(model)
	if model == "" {
		return nil
	}
	return []string{model}
}

func providerModelID(model string) string {
	model = strings.TrimSpace(model)
	_, _, officialModel, ok := splitModelFQN(model)
	if ok {
		return officialModel
	}
	return model
}
