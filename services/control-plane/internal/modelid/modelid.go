// Package modelid contains helpers for parsing and canonicalizing model identifiers.
package modelid

import "strings"

// SplitModelFQN parses provider:account_alias:model_id and canonicalizes
// the official model id by removing a redundant provider/ prefix when present.
func SplitModelFQN(input string) (provider, accountAlias, officialModel string, ok bool) {
	parts := strings.Split(strings.TrimSpace(input), ":")
	if len(parts) != 3 {
		return "", "", "", false
	}

	provider = strings.TrimSpace(parts[0])
	accountAlias = strings.TrimSpace(parts[1])

	officialModel = CanonicalOfficialModel(provider, parts[2])
	if provider == "" || accountAlias == "" || officialModel == "" {
		return "", "", "", false
	}

	return provider, accountAlias, officialModel, true
}

// CanonicalOfficialModel removes a redundant provider prefix from a model id.
func CanonicalOfficialModel(provider, officialModel string) string {
	provider = strings.TrimSpace(provider)

	officialModel = strings.TrimSpace(officialModel)
	if provider == "" || officialModel == "" {
		return officialModel
	}

	prefix := strings.ToLower(provider) + "/"

	lowerOfficial := strings.ToLower(officialModel)
	if strings.HasPrefix(lowerOfficial, prefix) {
		return strings.TrimSpace(officialModel[len(prefix):])
	}

	return officialModel
}

// BuildModelFQN builds the canonical provider:account:model identifier.
func BuildModelFQN(provider, accountAlias, officialModel string) string {
	provider = strings.TrimSpace(provider)
	accountAlias = strings.TrimSpace(accountAlias)

	officialModel = CanonicalOfficialModel(provider, officialModel)
	if provider == "" || accountAlias == "" || officialModel == "" {
		return ""
	}

	return provider + ":" + accountAlias + ":" + officialModel
}

// MatchOfficialModel compares two model identifiers after canonicalization.
func MatchOfficialModel(provider, left, right string) bool {
	return strings.EqualFold(CanonicalOfficialModel(provider, left), CanonicalOfficialModel(provider, right))
}
