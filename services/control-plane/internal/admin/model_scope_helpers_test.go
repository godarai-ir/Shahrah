package admin

import (
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

func TestIsVirtualModelInScope_NormalizesTargetFQN(t *testing.T) {
	vm := platform.VirtualModel{
		Alias: "mistral",
		Targets: []platform.VirtualModelTarget{
			{Model: "mistral-ai:mistral-ai-eager-harbor:mistral-ai/ministral-8b-2512"},
		},
	}
	scopeAllowedSet := map[string]struct{}{
		"mistral-ai:mistral-ai-eager-harbor:ministral-8b-2512": {},
	}

	if !isVirtualModelInScope(vm, scopeAllowedSet) {
		t.Fatal("expected virtual model target to match canonical provider-account scope")
	}
}

func TestBuildVirtualModelSnapshotsForAliases_NormalizesTargetFQNs(t *testing.T) {
	virtualModels := []platform.VirtualModel{
		{
			ID:      "vm-1",
			Alias:   "mistral",
			Version: 3,
			Targets: []platform.VirtualModelTarget{
				{Model: "mistral-ai:mistral-ai-eager-harbor:mistral-ai/ministral-8b-2512"},
			},
		},
	}
	scopeAllowedSet := map[string]struct{}{
		"mistral-ai:mistral-ai-eager-harbor:ministral-8b-2512": {},
	}

	snapshots, err := buildVirtualModelSnapshotsForAliases([]string{"mistral"}, virtualModels, scopeAllowedSet)
	if err != nil {
		t.Fatalf("expected snapshot build to succeed: %v", err)
	}

	targets := snapshots["mistral"].TargetFQNs
	if len(targets) != 1 || targets[0] != "mistral-ai:mistral-ai-eager-harbor:ministral-8b-2512" {
		t.Fatalf("expected canonical target FQN, got %#v", targets)
	}
}
