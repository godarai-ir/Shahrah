package storage

import "strings"

func maskCredentialHint(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}

	if isMaskedCredentialValue(trimmed) {
		return trimmed
	}

	prefixLen := 4
	if strings.HasPrefix(trimmed, "sk-") || strings.HasPrefix(trimmed, "gsk_") || strings.HasPrefix(trimmed, "gsk-") {
		prefixLen = 3
	}

	if len(trimmed) <= prefixLen+4 {
		return "••••"
	}

	return trimmed[:prefixLen] + "••••••••" + trimmed[len(trimmed)-4:]
}

func isMaskedCredentialValue(value string) bool {
	trimmed := strings.TrimSpace(value)
	return strings.Contains(trimmed, "•") || strings.Contains(trimmed, "*") || strings.Contains(trimmed, "…")
}
