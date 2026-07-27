package video

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

type workerTestStore struct {
	saved *platform.VideoGenerationJob
}

func (s *workerTestStore) ClaimVideoJobs(context.Context, string, int, time.Duration) ([]platform.VideoGenerationJob, error) {
	return nil, nil
}
func (s *workerTestStore) SaveVideoJob(_ context.Context, job *platform.VideoGenerationJob, _ int64) error {
	copy := *job
	s.saved = &copy
	return nil
}
func (s *workerTestStore) MarkVideoSubmission(context.Context, string, int, string, map[string]interface{}, string) error {
	return nil
}

type workerTestBilling struct {
	releases int
	commits  int
}

type workerTestSpend struct {
	logs []platform.SpendLog
}

func (s *workerTestSpend) LogSpend(_ context.Context, log platform.SpendLog) error {
	s.logs = append(s.logs, log)
	return nil
}

func (b *workerTestBilling) CommitBudget(context.Context, *platform.BudgetReservation, platform.UsageCostBreakdown) error {
	b.commits++
	return nil
}
func (b *workerTestBilling) ReleaseBudget(context.Context, *platform.BudgetReservation, string) error {
	b.releases++
	return nil
}

type workerTestTunnel struct{}

func (workerTestTunnel) Proxy(context.Context, platform.TunnelRequest) (*platform.TunnelResponse, error) {
	return &platform.TunnelResponse{StatusCode: 200, Body: []byte(`{}`)}, nil
}
func (workerTestTunnel) Stream(context.Context, platform.TunnelRequest) (platform.TunnelStream, error) {
	return nil, nil
}
func (workerTestTunnel) Healthy(context.Context) error { return nil }

type workerTestObjects struct {
	deleted []string
}

func (o *workerTestObjects) PutInput(context.Context, string, string, string, string, io.Reader) (platform.VideoArtifact, error) {
	return platform.VideoArtifact{}, nil
}
func (o *workerTestObjects) CopyFromURL(context.Context, string, string, platform.VideoArtifact) (platform.VideoArtifact, error) {
	return platform.VideoArtifact{}, nil
}
func (o *workerTestObjects) PresignGet(context.Context, string, time.Duration) (string, error) {
	return "", nil
}
func (o *workerTestObjects) Delete(_ context.Context, key string) error {
	o.deleted = append(o.deleted, key)
	return nil
}

func TestTerminalJobWithoutProvableUsageGoesToBillingReview(t *testing.T) {
	store := &workerTestStore{}
	billingService := &workerTestBilling{}
	worker := NewWorker(
		store,
		billingService,
		nil,
		NewAdapterRegistry(workerTestTunnel{}),
		nil,
		func(context.Context, string, string, string) (platform.Deployment, error) {
			return platform.Deployment{Provider: "xai"}, nil
		},
		nil,
	)
	job := &platform.VideoGenerationJob{
		ID: "job-1", TenantID: "tenant-1", Provider: "xai",
		GenerationStatus: platform.VideoGenerationFailed,
		SettlementStatus: platform.VideoSettlementPending,
		Version:          1,
	}

	if err := worker.process(context.Background(), job); err == nil {
		t.Fatal("expected billing review error")
	}
	if store.saved == nil || store.saved.GenerationStatus != platform.VideoGenerationBillingReview ||
		store.saved.SettlementStatus != platform.VideoSettlementReview {
		t.Fatalf("terminal usage-less job was not preserved for review: %#v", store.saved)
	}
	if billingService.releases != 0 || billingService.commits != 0 {
		t.Fatalf("unproven usage must not release or commit: %#v", billingService)
	}
}

func TestRetentionCleanupDeletesArtifactsAndMarksDeliveryDeleted(t *testing.T) {
	store := &workerTestStore{}
	objects := &workerTestObjects{}
	worker := NewWorker(store, &workerTestBilling{}, nil, nil, objects, nil, nil)
	job := &platform.VideoGenerationJob{
		ID: "job-2", TenantID: "tenant-1", Version: 4,
		DeliveryStatus: platform.VideoDeliveryReady,
		Artifacts: []platform.VideoArtifact{
			{ObjectKey: "tenants/tenant-1/videos/job-2/one.mp4"},
			{ObjectKey: "tenants/tenant-1/videos/job-2/two.mp4"},
		},
	}

	if err := worker.deleteExpiredArtifacts(context.Background(), job); err != nil {
		t.Fatal(err)
	}
	if len(objects.deleted) != 2 || store.saved == nil ||
		store.saved.DeliveryStatus != platform.VideoDeliveryDeleted || len(store.saved.Artifacts) != 0 {
		t.Fatalf("unexpected retention cleanup state: deleted=%v job=%#v", objects.deleted, store.saved)
	}
}

func TestSettlementUsesSubmitPricingSnapshotAfterDeploymentPriceChanges(t *testing.T) {
	store := &workerTestStore{}
	spend := &workerTestSpend{}
	changedPricing := &platform.ModelPrice{Status: "priced", Profiles: []platform.PricingProfile{{
		Region: "global", Currency: "USD",
		Rates: []platform.PricingRate{{Key: "changed", Usage: "output", Unit: "second", Amount: "9", Quantity: 1}},
	}}}
	worker := NewWorker(
		store,
		&workerTestBilling{},
		spend,
		NewAdapterRegistry(workerTestTunnel{}),
		nil,
		func(context.Context, string, string, string) (platform.Deployment, error) {
			return platform.Deployment{Provider: "xai", EffectivePricing: changedPricing}, nil
		},
		nil,
	)
	submitPricing := &platform.ModelPrice{Status: "priced", Profiles: []platform.PricingProfile{{
		Region: "global", Currency: "USD",
		Rates: []platform.PricingRate{{Key: "submit", Usage: "output", Unit: "second", Amount: "0.08", Quantity: 1}},
	}}}
	job := &platform.VideoGenerationJob{
		ID: "job-3", TenantID: "tenant-1", RequestID: "request-3", Model: "video", Provider: "xai",
		BillingMode: platform.BillingModeExternalBilled, GenerationStatus: platform.VideoGenerationSucceeded,
		DeliveryStatus: platform.VideoDeliveryReady, SettlementStatus: platform.VideoSettlementPending,
		ActualUsage:     []platform.UsageMeasure{{Usage: "output", Unit: "second", Amount: "10"}},
		PricingSnapshot: submitPricing, ExchangeRate: "1", Version: 1,
	}

	if err := worker.process(context.Background(), job); err != nil {
		t.Fatal(err)
	}
	if store.saved == nil || store.saved.FinalRating == nil || store.saved.FinalRating.TotalCostUSD != "0.8" {
		t.Fatalf("settlement did not use submit snapshot: %#v", store.saved)
	}
	if len(spend.logs) != 1 || spend.logs[0].RatingSnapshot.TotalCostUSD != "0.8" {
		t.Fatalf("unexpected spend log: %#v", spend.logs)
	}
}
