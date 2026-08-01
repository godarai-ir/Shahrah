package proxy

import (
	"os"
	"strconv"
	"strings"
)

func mistralEndpointEnabled(endpoint string) bool {
	if !mistralBooleanEnvironment("SHAHRAH_MISTRAL_ENABLED", true) {
		return false
	}

	key := "SHAHRAH_MISTRAL_" + strings.ToUpper(strings.NewReplacer("/", "_", "-", "_", " ", "_").Replace(strings.TrimSpace(endpoint))) + "_ENABLED"

	return mistralBooleanEnvironment(key, true)
}

func mistralBooleanEnvironment(key string, fallback bool) bool {
	raw, exists := os.LookupEnv(key)
	if !exists || strings.TrimSpace(raw) == "" {
		return fallback
	}

	value, err := strconv.ParseBool(strings.TrimSpace(raw))

	return err == nil && value
}
