package proxy

import (
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

func buildVirtualModelFromSnapshot(authInfo *platform.AuthContext, alias string) (*platform.VirtualModel, bool) {
	if authInfo == nil {
		return nil, false
	}

	alias = strings.TrimSpace(alias)
	if alias == "" || len(authInfo.VirtualModelSnapshots) == 0 {
		return nil, false
	}

	snapshot, ok := authInfo.VirtualModelSnapshots[alias]
	if !ok || len(snapshot.TargetFQNs) == 0 {
		return nil, false
	}

	targets := make([]platform.VirtualModelTarget, 0, len(snapshot.TargetFQNs))
	seen := make(map[string]struct{}, len(snapshot.TargetFQNs))
	order := 1

	for _, targetFQN := range snapshot.TargetFQNs {
		targetFQN = strings.TrimSpace(targetFQN)
		if targetFQN == "" {
			continue
		}

		if _, exists := seen[targetFQN]; exists {
			continue
		}

		seen[targetFQN] = struct{}{}

		provider, _, _, ok := splitModelFQN(targetFQN)
		if !ok {
			return nil, false
		}

		targets = append(targets, platform.VirtualModelTarget{
			Order:    order,
			Provider: provider,
			Model:    targetFQN,
			Policy: platform.VirtualModelPolicy{
				Retry: platform.VirtualModelRetryPolicy{
					Attempts:    2,
					DelayMS:     100,
					StatusCodes: []int{429, 500, 502, 503},
				},
				Fallback: platform.VirtualModelFallbackPolicy{
					StatusCodes: []int{429, 500, 502, 503, 504},
					Candidate:   true,
				},
			},
		})
		order++
	}

	if len(targets) == 0 {
		return nil, false
	}

	return &platform.VirtualModel{
		ID:              strings.TrimSpace(snapshot.VirtualModelID),
		TenantID:        authInfo.TenantID,
		Alias:           alias,
		RoutingStrategy: "priority",
		Enabled:         true,
		Version:         snapshot.Version,
		ExposedModes:    append([]string(nil), snapshot.ExposedModes...),
		Targets:         targets,
	}, true
}
