package cache

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestMemoryBudgetWindowReserveCommitReleaseIdempotent(t *testing.T) {
	mem, err := NewMemoryCache(zap.NewNop())
	if err != nil {
		t.Fatalf("NewMemoryCache returned error: %v", err)
	}
	ctx := context.Background()
	windowEnd := time.Now().UTC().Add(time.Hour)

	if _, err := mem.SeedBudgetWindow(ctx, BudgetWindowSeed{
		WindowKey:    "budget:tenant:t1:window:202607",
		AllowanceIRR: 100,
		ExpiresAt:    windowEnd,
		TTL:          time.Hour,
		Version:      1,
	}); err != nil {
		t.Fatalf("SeedBudgetWindow returned error: %v", err)
	}

	reserved, err := mem.ReserveBudget(ctx, BudgetReservationRequest{
		WindowKey:      "budget:tenant:t1:window:202607",
		ReservationKey: "budget:tenant:t1:reservation:req-1",
		ReservationID:  "req-1",
		EstimatedIRR:   60,
		TTL:            time.Hour,
	})
	if err != nil {
		t.Fatalf("ReserveBudget returned error: %v", err)
	}
	if !reserved.Allowed || reserved.ReservedIRR != 60 || reserved.RemainingIRR != 40 {
		t.Fatalf("unexpected reserve result: %+v", reserved)
	}

	again, err := mem.ReserveBudget(ctx, BudgetReservationRequest{
		WindowKey:      "budget:tenant:t1:window:202607",
		ReservationKey: "budget:tenant:t1:reservation:req-1",
		ReservationID:  "req-1",
		EstimatedIRR:   60,
		TTL:            time.Hour,
	})
	if err != nil {
		t.Fatalf("second ReserveBudget returned error: %v", err)
	}
	if again.ReservedIRR != 60 {
		t.Fatalf("expected idempotent reserve to keep reserved at 60, got %+v", again)
	}

	exhausted, err := mem.ReserveBudget(ctx, BudgetReservationRequest{
		WindowKey:      "budget:tenant:t1:window:202607",
		ReservationKey: "budget:tenant:t1:reservation:req-2",
		ReservationID:  "req-2",
		EstimatedIRR:   50,
		TTL:            time.Hour,
	})
	if err != nil {
		t.Fatalf("exhausting ReserveBudget returned error: %v", err)
	}
	if exhausted.Allowed || exhausted.State != "exhausted" {
		t.Fatalf("expected exhausted reservation, got %+v", exhausted)
	}

	committed, err := mem.CommitBudget(ctx, BudgetCommitRequest{
		WindowKey:      "budget:tenant:t1:window:202607",
		ReservationKey: "budget:tenant:t1:reservation:req-1",
		ActualIRR:      40,
	})
	if err != nil {
		t.Fatalf("CommitBudget returned error: %v", err)
	}
	if !committed.Allowed || committed.ReservedIRR != 0 || committed.CommittedIRR != 40 || committed.RemainingIRR != 60 {
		t.Fatalf("unexpected commit result: %+v", committed)
	}

	committedAgain, err := mem.CommitBudget(ctx, BudgetCommitRequest{
		WindowKey:      "budget:tenant:t1:window:202607",
		ReservationKey: "budget:tenant:t1:reservation:req-1",
		ActualIRR:      40,
	})
	if err != nil {
		t.Fatalf("second CommitBudget returned error: %v", err)
	}
	if committedAgain.CommittedIRR != 40 {
		t.Fatalf("expected idempotent commit to keep committed at 40, got %+v", committedAgain)
	}

	released, err := mem.ReleaseBudget(ctx, BudgetReleaseRequest{
		WindowKey:      "budget:tenant:t1:window:202607",
		ReservationKey: "budget:tenant:t1:reservation:req-2",
	})
	if err != nil {
		t.Fatalf("ReleaseBudget returned error: %v", err)
	}
	if !released.Allowed || released.CommittedIRR != 40 {
		t.Fatalf("unexpected release result: %+v", released)
	}

	snapshot, err := mem.GetBudgetWindow(ctx, "budget:tenant:t1:window:202607")
	if err != nil {
		t.Fatalf("GetBudgetWindow returned error: %v", err)
	}
	if snapshot == nil || snapshot.CommittedIRR != 40 || snapshot.ReservedIRR != 0 {
		t.Fatalf("unexpected snapshot: %+v", snapshot)
	}

	active, err := mem.ListActiveBudgetWindows(ctx, 10)
	if err != nil {
		t.Fatalf("ListActiveBudgetWindows returned error: %v", err)
	}
	if len(active) != 1 || active[0].WindowKey != "budget:tenant:t1:window:202607" {
		t.Fatalf("unexpected active windows: %+v", active)
	}
}
