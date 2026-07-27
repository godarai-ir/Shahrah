package video

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/billing"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type JobStore interface {
	ClaimVideoJobs(context.Context, string, int, time.Duration) ([]platform.VideoGenerationJob, error)
	SaveVideoJob(context.Context, *platform.VideoGenerationJob, int64) error
	MarkVideoSubmission(context.Context, string, int, string, map[string]interface{}, string) error
}

type expiredJobStore interface {
	ClaimExpiredVideoJobs(context.Context, string, int, time.Duration) ([]platform.VideoGenerationJob, error)
}

type WorkerBilling interface {
	CommitBudget(context.Context, *platform.BudgetReservation, platform.UsageCostBreakdown) error
	ReleaseBudget(context.Context, *platform.BudgetReservation, string) error
}

type SpendLogger interface {
	LogSpend(context.Context, platform.SpendLog) error
}

type DeploymentResolver func(context.Context, string, string, string) (platform.Deployment, error)

type Worker struct {
	jobs      JobStore
	billing   WorkerBilling
	spend     SpendLogger
	adapters  *AdapterRegistry
	objects   ObjectStore
	resolve   DeploymentResolver
	logger    *zap.Logger
	workerID  string
	interval  time.Duration
	batchSize int
	lease     time.Duration
	cancel    context.CancelFunc
	wg        sync.WaitGroup
}

func NewWorker(jobs JobStore, billingService WorkerBilling, spend SpendLogger, adapters *AdapterRegistry, objects ObjectStore, resolve DeploymentResolver, logger *zap.Logger) *Worker {
	return &Worker{
		jobs: jobs, billing: billingService, spend: spend, adapters: adapters, objects: objects,
		resolve: resolve, logger: logger, workerID: fmt.Sprintf("video-worker-%d", time.Now().UnixNano()),
		interval: time.Second, batchSize: 10, lease: 30 * time.Second,
	}
}

func (w *Worker) Start(parent context.Context) {
	if w == nil || w.jobs == nil {
		return
	}
	ctx, cancel := context.WithCancel(parent)
	w.cancel = cancel
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		ticker := time.NewTicker(w.interval)
		defer ticker.Stop()
		for {
			if err := w.runOnce(ctx); err != nil && !errors.Is(err, context.Canceled) && w.logger != nil {
				w.logger.Warn("video worker iteration failed", zap.Error(err))
			}
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
}

func (w *Worker) Stop() {
	if w == nil || w.cancel == nil {
		return
	}
	w.cancel()
	w.wg.Wait()
}

func (w *Worker) runOnce(ctx context.Context) error {
	jobs, err := w.jobs.ClaimVideoJobs(ctx, w.workerID, w.batchSize, w.lease)
	if err != nil {
		return err
	}
	for i := range jobs {
		if err := w.process(ctx, &jobs[i]); err != nil && w.logger != nil {
			w.logger.Warn("video job processing failed", zap.Error(err), zap.String("job_id", jobs[i].ID))
		}
	}
	if store, ok := w.jobs.(expiredJobStore); ok {
		expired, claimErr := store.ClaimExpiredVideoJobs(ctx, w.workerID+"-retention", w.batchSize, w.lease)
		if claimErr != nil {
			return claimErr
		}
		for i := range expired {
			if cleanupErr := w.deleteExpiredArtifacts(ctx, &expired[i]); cleanupErr != nil && w.logger != nil {
				w.logger.Warn("video artifact retention cleanup failed", zap.Error(cleanupErr), zap.String("job_id", expired[i].ID))
			}
		}
	}
	return nil
}

func (w *Worker) deleteExpiredArtifacts(ctx context.Context, job *platform.VideoGenerationJob) error {
	if w.objects == nil {
		return ErrObjectStoreUnavailable
	}
	for _, artifact := range job.Artifacts {
		if artifact.ObjectKey == "" {
			continue
		}
		if err := w.objects.Delete(ctx, artifact.ObjectKey); err != nil {
			return err
		}
	}
	expected := job.Version
	job.Artifacts = nil
	job.DeliveryStatus = platform.VideoDeliveryDeleted
	job.LastError = ""
	job.NextPollAt = nil
	return w.jobs.SaveVideoJob(ctx, job, expected)
}

func (w *Worker) process(ctx context.Context, job *platform.VideoGenerationJob) error {
	expected := job.Version
	dep, err := w.resolve(ctx, job.TenantID, job.Model, job.DeploymentID)
	if err != nil {
		return w.retry(job, expected, err)
	}
	adapter, err := w.adapters.Adapter(job.Provider)
	if err != nil {
		return w.review(job, expected, err)
	}

	if job.GenerationStatus == platform.VideoGenerationSubmitting && job.ProviderJobID == "" {
		inputURL := ""
		if job.Request.InputMediaID != "" && w.objects != nil {
			inputURL, err = w.objects.PresignGet(ctx, job.Request.InputMediaID, 30*time.Minute)
			if err != nil {
				return w.retry(job, expected, err)
			}
		}
		job.AttemptCount++
		_ = w.jobs.MarkVideoSubmission(ctx, job.ID, 1, "processing", nil, "")
		result, submitErr := adapter.Submit(ctx, job.RequestID, job.TenantID, dep, job.Request, inputURL)
		if submitErr != nil {
			if IsProviderRejection(submitErr) {
				_ = w.jobs.MarkVideoSubmission(ctx, job.ID, 1, "failed", nil, submitErr.Error())
				return w.rejectSubmission(ctx, job, expected, submitErr)
			}
			_ = w.jobs.MarkVideoSubmission(ctx, job.ID, 1, "processing", nil, submitErr.Error())
			return w.retry(job, expected, submitErr)
		}
		w.applyResult(job, result)
		_ = w.jobs.MarkVideoSubmission(ctx, job.ID, 1, "submitted", result.Raw, "")
	} else if job.GenerationStatus == platform.VideoGenerationQueued || job.GenerationStatus == platform.VideoGenerationRunning {
		result, getErr := adapter.GetStatus(ctx, job.RequestID, job.TenantID, dep, job.ProviderJobID)
		if getErr != nil {
			return w.retry(job, expected, getErr)
		}
		w.applyResult(job, result)
	}

	if job.GenerationStatus == platform.VideoGenerationFailed || job.GenerationStatus == platform.VideoGenerationCancelled {
		if job.SettlementStatus == platform.VideoSettlementPending {
			if len(job.ActualUsage) == 0 {
				return w.review(job, expected, errors.New("provider terminated video generation without provable actual usage"))
			}
			if requestUsageIsExact(job.ProviderResponse) {
				job.ActualUsage = completeUsageDimensions(job.ActualUsage, job.Request)
			}
			if err := w.settle(ctx, job); err != nil {
				return w.review(job, expected, err)
			}
		}
		now := time.Now().UTC()
		job.CompletedAt = &now
		job.NextPollAt = nil
		return w.jobs.SaveVideoJob(ctx, job, expected)
	}

	if job.GenerationStatus == platform.VideoGenerationSucceeded {
		if job.SettlementStatus == platform.VideoSettlementPending {
			if len(job.ActualUsage) > 0 && requestUsageIsExact(job.ProviderResponse) {
				job.ActualUsage = completeUsageDimensions(job.ActualUsage, job.Request)
			}
			if len(job.ActualUsage) == 0 {
				if err := w.copyOutputs(ctx, job); err != nil {
					return w.retry(job, expected, err)
				}
				job.ActualUsage = usageFromArtifactMetadata(job.Artifacts, job.Request.ServiceTier)
			}
			if err := w.settle(ctx, job); err != nil {
				return w.review(job, expected, err)
			}
		}
		if job.DeliveryStatus != platform.VideoDeliveryReady {
			if err := w.deliver(ctx, job); err != nil {
				job.DeliveryStatus = platform.VideoDeliveryFailed
				job.LastError = err.Error()
				next := time.Now().UTC().Add(backoff(job.AttemptCount))
				job.NextPollAt = &next
				job.AttemptCount++
				return w.jobs.SaveVideoJob(ctx, job, expected)
			}
		}
		now := time.Now().UTC()
		job.CompletedAt = &now
		job.NextPollAt = nil
	}

	return w.jobs.SaveVideoJob(ctx, job, expected)
}

func (w *Worker) rejectSubmission(ctx context.Context, job *platform.VideoGenerationJob, expected int64, rejection error) error {
	if err := w.billing.ReleaseBudget(ctx, job.BudgetReservation, "video_provider_rejected_submission"); err != nil {
		return w.review(job, expected, errors.Join(rejection, err))
	}
	now := time.Now().UTC()
	job.GenerationStatus = platform.VideoGenerationFailed
	job.SettlementStatus = platform.VideoSettlementReleased
	job.LastError = rejection.Error()
	job.CompletedAt = &now
	job.NextPollAt = nil
	if err := w.jobs.SaveVideoJob(ctx, job, expected); err != nil {
		return errors.Join(rejection, err)
	}
	return rejection
}

func completeUsageDimensions(measures []platform.UsageMeasure, request platform.VideoCreateRequest) []platform.UsageMeasure {
	out := make([]platform.UsageMeasure, len(measures))
	copy(out, measures)
	resolution := strings.TrimSpace(request.Resolution)
	if resolution == "" {
		parts := strings.Split(strings.ToLower(strings.TrimSpace(request.Size)), "x")
		if len(parts) == 2 {
			left, leftErr := strconv.Atoi(parts[0])
			right, rightErr := strconv.Atoi(parts[1])
			if leftErr == nil && rightErr == nil {
				height := left
				if right < height {
					height = right
				}
				if height > 0 {
					resolution = strconv.Itoa(height) + "p"
				}
			}
		}
	}
	serviceTier := strings.TrimSpace(request.ServiceTier)
	if serviceTier == "" {
		serviceTier = "standard"
	}
	for i := range out {
		if out[i].Resolution == "" {
			out[i].Resolution = resolution
		}
		if out[i].ServiceTier == "" {
			out[i].ServiceTier = serviceTier
		}
	}
	return out
}

func (w *Worker) applyResult(job *platform.VideoGenerationJob, result ProviderResult) {
	if result.ProviderJobID != "" {
		job.ProviderJobID = result.ProviderJobID
	}
	if result.Status != "" {
		job.GenerationStatus = result.Status
	}
	if len(result.ActualUsage) > 0 {
		job.ActualUsage = result.ActualUsage
	}
	if len(result.Artifacts) > 0 {
		job.Artifacts = result.Artifacts
	}
	if result.Raw != nil {
		exact := requestUsageIsExact(job.ProviderResponse) || result.RequestIsExact
		job.ProviderResponse = result.Raw
		if exact {
			job.ProviderResponse["_shahrah_request_exact"] = true
		}
	}
	next := time.Now().UTC().Add(result.PollAfter)
	job.NextPollAt = &next
}

func (w *Worker) settle(ctx context.Context, job *platform.VideoGenerationJob) error {
	measures := job.ActualUsage
	if len(measures) == 0 && requestUsageIsExact(job.ProviderResponse) {
		measures = job.RequestedUsage
	}
	if len(measures) == 0 {
		return errors.New("provider result has no provable actual video usage")
	}
	if job.BillingMode == platform.BillingModeExternalBilled &&
		(job.PricingSnapshot == nil || strings.EqualFold(strings.TrimSpace(job.PricingSnapshot.Status), "unknown")) {
		rated := platform.RatedUsage{
			Status: "unrated", Model: job.Model, Provider: job.Provider, Region: job.Region,
			PricingSource: job.PricingSource, Measures: measures, TotalCostUSD: "0",
			ExchangeRate: firstNonEmpty(job.ExchangeRate, "0"), TotalCostIRR: "0",
		}
		job.ActualUsage = measures
		job.FinalRating = &rated
		job.SettlementStatus = platform.VideoSettlementCommitted
		if w.spend != nil {
			if err := w.spend.LogSpend(ctx, videoSpendLog(job, rated)); err != nil {
				return err
			}
		}
		return nil
	}
	exchangeRate, err := decimal.NewFromString(job.ExchangeRate)
	if err != nil {
		return err
	}
	rated, err := billing.RateUsage(billing.RateInput{
		Model: job.Model, Provider: job.Provider, Region: job.Region, PricingSource: job.PricingSource,
		Pricing: job.PricingSnapshot, Measures: measures, ExchangeRate: exchangeRate,
	})
	if err != nil {
		return err
	}
	breakdown, err := billing.UsageCostBreakdownFromRated(rated, job.Model)
	if err != nil {
		return err
	}
	if job.BillingMode != platform.BillingModeExternalBilled {
		if err := w.billing.CommitBudget(ctx, job.BudgetReservation, breakdown); err != nil {
			return err
		}
	}

	job.ActualUsage = measures
	job.FinalRating = &rated
	job.SettlementStatus = platform.VideoSettlementCommitted
	if w.spend != nil {
		if err := w.spend.LogSpend(ctx, videoSpendLog(job, rated)); err != nil {
			return err
		}
	}
	return nil
}

func (w *Worker) deliver(ctx context.Context, job *platform.VideoGenerationJob) error {
	if err := w.copyOutputs(ctx, job); err != nil {
		return err
	}
	job.DeliveryStatus = platform.VideoDeliveryReady
	job.LastError = ""
	return nil
}

func (w *Worker) copyOutputs(ctx context.Context, job *platform.VideoGenerationJob) error {
	if w.objects == nil {
		return ErrObjectStoreUnavailable
	}
	job.DeliveryStatus = platform.VideoDeliveryCopying
	delivered := make([]platform.VideoArtifact, 0, len(job.Artifacts))
	for _, artifact := range job.Artifacts {
		if artifact.ObjectKey != "" {
			delivered = append(delivered, artifact)
			continue
		}
		if artifact.ProviderURL == "" {
			return errors.New("provider result is missing an output URL")
		}
		stored, err := w.objects.CopyFromURL(ctx, job.TenantID, job.ID, artifact)
		if err != nil {
			return err
		}
		delivered = append(delivered, stored)
	}
	if len(delivered) == 0 {
		return errors.New("provider result contains no video artifacts")
	}
	job.Artifacts = delivered
	return nil
}

func usageFromArtifactMetadata(artifacts []platform.VideoArtifact, serviceTier string) []platform.UsageMeasure {
	type aggregate struct {
		duration decimal.Decimal
		count    int
	}
	byResolution := map[string]aggregate{}
	for _, artifact := range artifacts {
		duration, err := decimal.NewFromString(strings.TrimSpace(artifact.Duration))
		if err != nil || !duration.IsPositive() {
			return nil
		}
		resolution := strings.TrimSpace(artifact.Resolution)
		current := byResolution[resolution]
		current.duration = current.duration.Add(duration)
		current.count++
		byResolution[resolution] = current
	}
	if len(byResolution) == 0 {
		return nil
	}
	resolutions := make([]string, 0, len(byResolution))
	for resolution := range byResolution {
		resolutions = append(resolutions, resolution)
	}
	sort.Strings(resolutions)
	if strings.TrimSpace(serviceTier) == "" {
		serviceTier = "standard"
	}
	measures := make([]platform.UsageMeasure, 0, len(resolutions))
	for _, resolution := range resolutions {
		value := byResolution[resolution]
		measures = append(measures, platform.UsageMeasure{
			Usage: "output", Unit: "second", Amount: value.duration.String(),
			Resolution: resolution, ServiceTier: serviceTier,
			Qualifiers: map[string]interface{}{"output_count": value.count, "source": "media_probe"},
		})
	}
	return measures
}

func (w *Worker) retry(job *platform.VideoGenerationJob, expected int64, processErr error) error {
	job.AttemptCount++
	job.LastError = processErr.Error()
	next := time.Now().UTC().Add(backoff(job.AttemptCount))
	job.NextPollAt = &next
	if job.AttemptCount >= 12 {
		return w.review(job, expected, processErr)
	}
	if err := w.jobs.SaveVideoJob(context.Background(), job, expected); err != nil {
		return errors.Join(processErr, err)
	}
	return processErr
}

func (w *Worker) review(job *platform.VideoGenerationJob, expected int64, processErr error) error {
	job.GenerationStatus = platform.VideoGenerationBillingReview
	job.SettlementStatus = platform.VideoSettlementReview
	job.LastError = processErr.Error()
	job.NextPollAt = nil
	if err := w.jobs.SaveVideoJob(context.Background(), job, expected); err != nil {
		return errors.Join(processErr, err)
	}
	return processErr
}

func videoSpendLog(job *platform.VideoGenerationJob, rated platform.RatedUsage) platform.SpendLog {
	costUSD, _ := decimal.NewFromString(rated.TotalCostUSD)
	costIRR, _ := decimal.NewFromString(rated.TotalCostIRR)
	costUSDFloat, _ := costUSD.Float64()
	costIRRFloat, _ := costIRR.Float64()
	seconds, resolution := videoUsageProjection(rated.Measures)
	now := time.Now().UTC()
	return platform.SpendLog{
		RequestID: job.RequestID, TokenHash: job.TokenHash, TenantID: job.TenantID, Model: job.Model,
		CostUSD: costUSDFloat, CostIRR: costIRRFloat, Status: "success",
		RequestMetadata: map[string]interface{}{"video_job_id": job.ID, "provider": job.Provider, "billing_mode": job.BillingMode},
		CompletionStart: job.CreatedAt, CompletionEnd: now,
		UsageMeasures: rated.Measures, RatingSnapshot: &rated, CostComponents: rated.Components,
		PricingSource: rated.PricingSource, RatingStatus: rated.Status,
		VideoSeconds: seconds, VideoResolution: resolution, OutputCount: videoOutputCount(rated.Measures),
	}
}

func videoUsageProjection(measures []platform.UsageMeasure) (string, string) {
	for _, measure := range measures {
		if measure.Usage == "output" && measure.Unit == "second" {
			return measure.Amount, measure.Resolution
		}
	}
	return "", ""
}

func videoOutputCount(measures []platform.UsageMeasure) int {
	for _, measure := range measures {
		if value, ok := measure.Qualifiers["output_count"]; ok {
			switch typed := value.(type) {
			case int:
				return typed
			case float64:
				return int(typed)
			case string:
				count, _ := strconv.Atoi(typed)
				return count
			}
		}
	}
	return 1
}

func requestUsageIsExact(response map[string]interface{}) bool {
	value, _ := response["_shahrah_request_exact"].(bool)
	return value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func backoff(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	if attempt > 6 {
		attempt = 6
	}
	base := time.Duration(1<<uint(attempt-1)) * time.Second
	// Add up to 25% jitter so retries from a provider outage do not synchronize.
	return base + time.Duration(rand.Int63n(int64(base/4)+1))
}
