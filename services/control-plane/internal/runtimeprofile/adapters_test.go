package runtimeprofile

import (
	"context"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"go.uber.org/zap"
)

func TestDeploymentStoreListVirtualModelsUsesProfileSortedDefensiveCopy(t *testing.T) {
	createdFirst := time.Date(2026, 7, 7, 8, 0, 0, 0, time.UTC)
	createdSecond := createdFirst.Add(time.Minute)
	manager := NewManager(nil, nil, nil, zap.NewNop(), Config{StaleGrace: time.Minute, VersionCheckInterval: time.Minute})
	manager.setL1("tenant-1", &Profile{
		TenantID: "tenant-1",
		Version:  1,
		VirtualModels: map[string]platform.VirtualModel{
			"zeta": {
				ID:           "vm-2",
				TenantID:     "tenant-1",
				Alias:        "zeta",
				CreatedAt:    createdSecond,
				ExposedModes: []string{string(platform.ModelTypeChat)},
				Targets: []platform.VirtualModelTarget{{
					Model: "openai:main:gpt-4o",
					Policy: platform.VirtualModelPolicy{
						Retry: platform.VirtualModelRetryPolicy{StatusCodes: []int{500}},
					},
				}},
			},
			"alpha": {
				ID:           "vm-1",
				TenantID:     "tenant-1",
				Alias:        "alpha",
				CreatedAt:    createdFirst,
				ExposedModes: []string{string(platform.ModelTypeEmbedding)},
				Targets: []platform.VirtualModelTarget{{
					Model: "openai:main:text-embedding-3-small",
					Policy: platform.VirtualModelPolicy{
						Retry: platform.VirtualModelRetryPolicy{StatusCodes: []int{429}},
					},
				}},
			},
		},
	})

	store := NewDeploymentStore(manager)
	models, err := store.ListVirtualModels(context.Background(), "tenant-1")
	if err != nil {
		t.Fatalf("ListVirtualModels returned error: %v", err)
	}
	if len(models) != 2 {
		t.Fatalf("expected 2 virtual models, got %d", len(models))
	}
	if models[0].Alias != "alpha" || models[1].Alias != "zeta" {
		t.Fatalf("expected stable created_at ordering, got %q then %q", models[0].Alias, models[1].Alias)
	}

	models[0].Alias = "mutated"
	models[0].ExposedModes[0] = "mutated"
	models[0].Targets[0].Policy.Retry.StatusCodes[0] = 418

	again, err := store.ListVirtualModels(context.Background(), "tenant-1")
	if err != nil {
		t.Fatalf("second ListVirtualModels returned error: %v", err)
	}
	if again[0].Alias != "alpha" {
		t.Fatalf("expected defensive copy alias, got %q", again[0].Alias)
	}
	if again[0].ExposedModes[0] != string(platform.ModelTypeEmbedding) {
		t.Fatalf("expected defensive copy exposed mode, got %q", again[0].ExposedModes[0])
	}
	if again[0].Targets[0].Policy.Retry.StatusCodes[0] != 429 {
		t.Fatalf("expected defensive copy retry status, got %d", again[0].Targets[0].Policy.Retry.StatusCodes[0])
	}
}
