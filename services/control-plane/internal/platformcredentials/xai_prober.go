package platformcredentials

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
)

type xaiAPIKeyProbeResponse struct {
	ACLs           []string `json:"acls"`
	APIKeyDisabled bool     `json:"api_key_disabled"`
	APIKeyBlocked  bool     `json:"api_key_blocked"`
	TeamBlocked    bool     `json:"team_blocked"`
}

func classifyXAIAPIKeyProbe(status int, body []byte, modelHint string) ProbeResult {
	if status == http.StatusUnauthorized {
		return ProbeResult{Category: "invalid_credential", Summary: "xAI rejected the API key."}
	}

	if status == http.StatusForbidden {
		return ProbeResult{
			Authenticated: true,
			Category:      "forbidden",
			Summary:       "The xAI API key authenticated, but xAI denied access to its key-status endpoint.",
		}
	}

	if status == http.StatusTooManyRequests {
		return ProbeResult{
			Authenticated: true,
			Category:      "rate_limited",
			Summary:       "The xAI API key authenticated, but rate limiting prevented readiness validation.",
		}
	}

	if status < http.StatusOK || status >= http.StatusMultipleChoices {
		return ProbeResult{
			Category: "provider",
			Summary:  fmt.Sprintf("xAI API key validation returned HTTP %d.", status),
		}
	}

	var payload xaiAPIKeyProbeResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return ProbeResult{
			Authenticated: true,
			Category:      "provider",
			Summary:       "xAI authenticated the API key but returned an invalid key-status response.",
		}
	}

	payload.ACLs = normalizeXAIPermissions(payload.ACLs)
	metadata := ProbeMetadata{
		Permissions:    payload.ACLs,
		KeyDisabled:    boolPointer(payload.APIKeyDisabled),
		KeyBlocked:     boolPointer(payload.APIKeyBlocked),
		AccountBlocked: boolPointer(payload.TeamBlocked),
	}

	switch {
	case payload.APIKeyDisabled:
		return ProbeResult{
			Authenticated: true,
			Category:      "credential_disabled",
			Summary:       "The xAI API key is valid but disabled.",
			Metadata:      metadata,
		}
	case payload.APIKeyBlocked:
		return ProbeResult{
			Authenticated: true,
			Category:      "credential_blocked",
			Summary:       "The xAI API key is valid but blocked.",
			Metadata:      metadata,
		}
	case payload.TeamBlocked:
		return ProbeResult{
			Authenticated: true,
			Category:      "team_blocked",
			Summary:       "The xAI API key is valid, but its team is blocked. Verify that the team has active credits or licenses in the xAI Console.",
			Metadata:      metadata,
		}
	}

	missing := missingXAIRuntimePermissions(payload.ACLs, modelHint)
	if len(missing) > 0 {
		return ProbeResult{
			Authenticated: true,
			Category:      "missing_scope",
			Summary:       "The xAI API key is valid but lacks required runtime permissions: " + strings.Join(missing, ", ") + ".",
			Metadata:      metadata,
		}
	}

	return ProbeResult{
		Authenticated: true,
		Ready:         true,
		Category:      "ok",
		Summary:       "The xAI API key is valid, active, and has runtime permissions.",
		Metadata:      metadata,
	}
}

func normalizeXAIPermissions(permissions []string) []string {
	seen := make(map[string]struct{}, len(permissions))
	normalized := make([]string, 0, len(permissions))

	for _, permission := range permissions {
		permission = normalizedPermission(permission)
		if permission == "" {
			continue
		}

		if _, ok := seen[permission]; ok {
			continue
		}

		seen[permission] = struct{}{}
		normalized = append(normalized, permission)
	}

	sort.Strings(normalized)

	return normalized
}

func missingXAIRuntimePermissions(permissions []string, modelHint string) []string {
	permissionSet := make(map[string]struct{}, len(permissions))
	for _, permission := range permissions {
		permissionSet[normalizedPermission(permission)] = struct{}{}
	}

	missing := make([]string, 0, 2)
	if !hasXAIRuntimeEndpointPermission(permissionSet) {
		missing = append(missing, "api-key:endpoint:<runtime-endpoint>")
	}

	modelHint = canonicalXAIModelHint(modelHint)
	if modelHint != "" {
		if !hasXAIPermission(permissionSet, "api-key:model:*", "api-key:model:"+modelHint) {
			missing = append(missing, "api-key:model:"+modelHint)
		}
	} else if !hasXAIPermissionPrefix(permissionSet, "api-key:model:") {
		missing = append(missing, "api-key:model:<model>")
	}

	return missing
}

func hasXAIRuntimeEndpointPermission(permissions map[string]struct{}) bool {
	if _, ok := permissions["api-key:endpoint:*"]; ok {
		return true
	}

	for permission := range permissions {
		if strings.HasPrefix(permission, "api-key:endpoint:") && permission != "api-key:endpoint:models" {
			return true
		}
	}

	return false
}

func hasXAIPermission(permissions map[string]struct{}, candidates ...string) bool {
	for _, candidate := range candidates {
		if _, ok := permissions[normalizedPermission(candidate)]; ok {
			return true
		}
	}

	return false
}

func hasXAIPermissionPrefix(permissions map[string]struct{}, prefix string) bool {
	prefix = normalizedPermission(prefix)
	for permission := range permissions {
		if strings.HasPrefix(permission, prefix) {
			return true
		}
	}

	return false
}

func canonicalXAIModelHint(model string) string {
	model = strings.ToLower(strings.TrimSpace(model))
	if model == "" {
		return ""
	}

	if parts := strings.Split(model, ":"); len(parts) == 3 && parts[0] == "xai" {
		model = parts[2]
	}

	return strings.TrimPrefix(model, "xai/")
}
