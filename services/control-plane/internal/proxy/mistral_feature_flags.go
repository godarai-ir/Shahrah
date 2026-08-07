package proxy

import (
	"strings"
)

func mistralEndpointEnabled(config ProviderFeatureConfig, endpoint string) bool {
	if !config.Configured {
		return true
	}

	if !config.MistralEnabled {
		return false
	}

	endpoint = strings.ToLower(strings.TrimSpace(endpoint))
	enabled, found := config.MistralEndpoints[endpoint]

	if !found {
		return true
	}

	return enabled
}
