package bootstrap

import (
	"os"
	"strings"
)

func splitCSVEnv(key string) []string {
	parts := strings.Split(os.Getenv(key), ",")
	values := make([]string, 0, len(parts))

	for _, part := range parts {
		if value := strings.TrimSpace(part); value != "" {
			values = append(values, value)
		}
	}

	return values
}

func stringEnvOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}

	return fallback
}
