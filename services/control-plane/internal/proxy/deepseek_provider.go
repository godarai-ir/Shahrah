package proxy

import "strings"

func isDeepSeekProvider(provider string) bool {
	return strings.EqualFold(strings.TrimSpace(provider), deepSeekProviderName)
}
