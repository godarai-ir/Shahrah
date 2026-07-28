package proxy

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"mime/multipart"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	"github.com/Delpak/shahrah/services/control-plane/internal/router"
	"github.com/Delpak/shahrah/services/control-plane/internal/storage"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	videoruntime "github.com/Delpak/shahrah/services/control-plane/internal/video"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type videoJobStore interface {
	CreateVideoJob(context.Context, *platform.VideoGenerationJob) (*platform.VideoGenerationJob, error)
	GetVideoJob(context.Context, storage.VideoJobAccessScope, string) (*platform.VideoGenerationJob, error)
	GetVideoJobByIdempotency(context.Context, storage.VideoJobAccessScope, string) (*platform.VideoGenerationJob, error)
	GetVideoJobByProviderID(context.Context, string, string) (*platform.VideoGenerationJob, error)
	ListVideoJobs(context.Context, storage.VideoJobAccessScope, int, time.Time) ([]platform.VideoGenerationJob, error)
	SaveVideoJob(context.Context, *platform.VideoGenerationJob, int64) error
	MarkVideoSubmission(context.Context, string, int, string, map[string]interface{}, string) error
}

type videoBilling interface {
	budgeter
	budgetReservationManager
	deploymentUsageRater
}

type VideoHandler struct {
	router   *router.Router
	billing  videoBilling
	jobs     videoJobStore
	adapters *videoruntime.AdapterRegistry
	objects  videoruntime.ObjectStore
	logger   *zap.Logger
}

func NewVideoHandler(llmRouter *router.Router, billing videoBilling, jobs videoJobStore, adapters *videoruntime.AdapterRegistry, objects videoruntime.ObjectStore, logger *zap.Logger) *VideoHandler {
	return &VideoHandler{router: llmRouter, billing: billing, jobs: jobs, adapters: adapters, objects: objects, logger: logger}
}

func NewVideoDeploymentResolver(llmRouter *router.Router) videoruntime.DeploymentResolver {
	return func(ctx context.Context, tenantID, model, deploymentID string) (platform.Deployment, error) {
		var routeErr error
		for _, candidate := range modelRoutingCandidates(model) {
			var deployment *router.Deployment
			var err error
			if deploymentID == "" {
				deployment, err = llmRouter.Route(ctx, tenantID, candidate)
			} else {
				deployment, err = llmRouter.ResolveDeployment(ctx, tenantID, candidate, deploymentID)
			}
			if err != nil || deployment == nil {
				routeErr = err
				continue
			}
			return toContractDeployment(deployment), nil
		}
		return platform.Deployment{}, routeErr
	}
}

// Webhook accepts provider callbacks only after timestamped HMAC validation.
// It persists normalized provider state and wakes the regular worker; delivery
// and settlement therefore share one idempotent state machine with polling.
func (h *VideoHandler) Webhook(c *fiber.Ctx) error {
	provider := strings.ToLower(strings.TrimSpace(c.Params("provider")))
	secret := videoWebhookSecret(provider)
	if secret == "" {
		return writeAPIError(c, fiber.StatusServiceUnavailable, "Video webhook is not configured", "webhook_unavailable")
	}
	timestampRaw := strings.TrimSpace(c.Get("X-Shahrah-Timestamp"))
	timestamp, err := strconv.ParseInt(timestampRaw, 10, 64)
	if err != nil || absDuration(time.Since(time.Unix(timestamp, 0))) > 5*time.Minute {
		return writeAPIError(c, fiber.StatusUnauthorized, "Video webhook timestamp is invalid", "invalid_signature")
	}
	signatureRaw := strings.TrimPrefix(strings.TrimSpace(c.Get("X-Shahrah-Signature")), "sha256=")
	signature, err := hex.DecodeString(signatureRaw)
	if err != nil {
		return writeAPIError(c, fiber.StatusUnauthorized, "Video webhook signature is invalid", "invalid_signature")
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(timestampRaw))
	_, _ = mac.Write([]byte("."))
	_, _ = mac.Write(c.Body())
	if !hmac.Equal(signature, mac.Sum(nil)) {
		return writeAPIError(c, fiber.StatusUnauthorized, "Video webhook signature is invalid", "invalid_signature")
	}

	result := videoruntime.ParseWebhookResult(c.Body())
	providerJobID := firstNonEmptyString(result.ProviderJobID, c.Query("provider_job_id"))
	if providerJobID == "" {
		return writeAPIError(c, fiber.StatusBadRequest, "Provider job ID is required", "invalid_request_error")
	}
	job, err := h.jobs.GetVideoJobByProviderID(c.Context(), provider, providerJobID)
	if errors.Is(err, pgx.ErrNoRows) {
		return writeAPIError(c, fiber.StatusNotFound, "Video job not found", "not_found")
	}
	if err != nil {
		return writeAPIError(c, fiber.StatusServiceUnavailable, "Unable to load video job", "storage_error")
	}
	if isTerminalVideoStatus(job.GenerationStatus) &&
		(job.SettlementStatus == platform.VideoSettlementCommitted || job.SettlementStatus == platform.VideoSettlementReleased) {
		return c.SendStatus(fiber.StatusNoContent)
	}

	expected := job.Version
	if result.Status != "" {
		job.GenerationStatus = result.Status
	}
	if len(result.ActualUsage) > 0 {
		job.ActualUsage = result.ActualUsage
		if job.SettlementStatus == platform.VideoSettlementReview {
			job.SettlementStatus = platform.VideoSettlementPending
		}
	}
	if len(result.Artifacts) > 0 {
		job.Artifacts = result.Artifacts
	}
	exact := false
	if job.ProviderResponse != nil {
		exact, _ = job.ProviderResponse["_shahrah_request_exact"].(bool)
	}
	job.ProviderResponse = result.Raw
	if job.ProviderResponse == nil {
		job.ProviderResponse = map[string]interface{}{}
	}
	if exact {
		job.ProviderResponse["_shahrah_request_exact"] = true
	}
	now := time.Now().UTC()
	job.NextPollAt = &now
	if err := h.jobs.SaveVideoJob(c.Context(), job, expected); err != nil {
		if errors.Is(err, storage.ErrVideoJobConflict) {
			return c.SendStatus(fiber.StatusNoContent)
		}
		return writeAPIError(c, fiber.StatusServiceUnavailable, "Unable to persist video webhook", "storage_error")
	}
	return c.SendStatus(fiber.StatusAccepted)
}

func videoWebhookSecret(provider string) string {
	key := strings.ToUpper(strings.NewReplacer("-", "_", ".", "_").Replace(provider))
	return firstNonEmptyString(os.Getenv("VIDEO_WEBHOOK_SECRET_"+key), os.Getenv("VIDEO_WEBHOOK_SECRET"))
}

func absDuration(value time.Duration) time.Duration {
	if value < 0 {
		return -value
	}
	return value
}

func (h *VideoHandler) Create(c *fiber.Ctx) error {
	authInfo, currentTenant, err := videoRequestContext(c)
	if err != nil {
		return err
	}
	ownerPrincipal, err := videoOwnerPrincipal(authInfo)
	if err != nil {
		return writeAPIError(c, fiber.StatusForbidden, err.Error(), "authorization_error")
	}
	ownerScope := storage.VideoJobAccessScope{
		TenantID:       currentTenant.ID,
		OwnerPrincipal: ownerPrincipal,
	}
	if h.objects == nil {
		return writeAPIError(c, fiber.StatusServiceUnavailable, "Video object storage is unavailable", "storage_error")
	}
	request, inputFile, err := parseVideoCreateRequest(c)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadRequest, err.Error(), "invalid_request_error")
	}
	if err := validateVideoCreateRequest(request); err != nil {
		return writeAPIError(c, fiber.StatusBadRequest, err.Error(), "invalid_request_error")
	}
	if request.InputMediaID != "" && !strings.Contains("/"+strings.TrimLeft(strings.ReplaceAll(request.InputMediaID, "\\", "/"), "/"), "/tenants/"+currentTenant.ID+"/videos/") {
		return writeAPIError(c, fiber.StatusForbidden, "input_media_id does not belong to this tenant", "authorization_error")
	}

	requestedModel := normalizeModelIdentifier(request.Model)
	access := modelAccessForType(authInfo, requestedModel, platform.ModelTypeVideoGeneration)
	if !access.Allowed {
		return writeAPIError(c, fiber.StatusForbidden, "Model access denied or unsupported for video generation", "authorization_error")
	}

	idempotencyKey := strings.TrimSpace(c.Get("Idempotency-Key"))
	if len(idempotencyKey) > 128 {
		return writeAPIError(c, fiber.StatusBadRequest, "Idempotency-Key is too long", "invalid_request_error")
	}
	if idempotencyKey != "" {
		existing, lookupErr := h.jobs.GetVideoJobByIdempotency(c.Context(), ownerScope, idempotencyKey)
		if lookupErr == nil {
			return c.Status(fiber.StatusOK).JSON(existing)
		}
		if !errors.Is(lookupErr, pgx.ErrNoRows) {
			return writeAPIError(c, fiber.StatusServiceUnavailable, "Unable to resolve idempotent video request", "storage_error")
		}
	}

	deployment, err := h.routeVideo(c.Context(), currentTenant.ID, requestedModel)
	if err != nil {
		return writeAPIError(c, fiber.StatusServiceUnavailable, "No available video deployment", "routing_error")
	}
	contractDeployment := toContractDeployment(deployment)

	request.Model = requestedModel
	requestedUsage := videoRequestedUsage(request)
	estimate := platform.UsageCostBreakdown{
		Model: requestedModel, RequestType: "video_generation", Endpoint: "/v1/videos",
		UsageMeasures: requestedUsage,
	}
	estimate, err = h.billing.RateDeploymentUsage(c.Context(), contractDeployment, estimate)
	if err != nil {
		return writeAPIError(c, fiber.StatusUnprocessableEntity, "Effective video pricing is unavailable", "pricing_unresolved")
	}

	requestID := strings.TrimSpace(c.Get("X-Request-ID"))
	if requestID == "" {
		requestID = "video_" + uuid.NewString()
	}
	reservation, handled, err := reserveBudgetForRequest(c, h.billing, authInfo, estimate, requestID, deploymentBillingMode(contractDeployment))
	if err != nil || handled {
		return err
	}

	jobID := uuid.NewString()
	if inputFile != nil {
		if h.objects == nil {
			releaseBudgetBestEffort(c.Context(), h.logger, h.billing, reservation, "object_store_unavailable")
			return writeAPIError(c, fiber.StatusServiceUnavailable, "Video input storage is unavailable", "storage_error")
		}
		inputArtifact, uploadErr := h.storeVideoInput(c.Context(), currentTenant.ID, jobID, inputFile)
		if uploadErr != nil {
			releaseBudgetBestEffort(c.Context(), h.logger, h.billing, reservation, "input_upload_failed")
			return writeAPIError(c, fiber.StatusBadGateway, "Unable to store video input", "storage_error")
		}
		request.InputMediaID = inputArtifact.ObjectKey
	}

	now := time.Now().UTC()
	job := &platform.VideoGenerationJob{
		ID: jobID, TenantID: currentTenant.ID, RequestID: requestID, IdempotencyKey: idempotencyKey,
		TokenHash: authInfo.TokenHash, ActorID: firstNonEmptyString(authInfo.UserID, authInfo.VirtualUserID), OwnerPrincipal: ownerPrincipal,
		Model: requestedModel, Provider: contractDeployment.Provider, ProviderModel: contractDeployment.Model,
		DeploymentID: contractDeployment.ID, ProviderAccountID: contractDeployment.ProviderAccountID,
		Region: contractDeployment.RegionName, BillingMode: deploymentBillingMode(contractDeployment),
		PricingSource: contractDeployment.PricingSource, GenerationStatus: platform.VideoGenerationSubmitting,
		DeliveryStatus: platform.VideoDeliveryPending, SettlementStatus: platform.VideoSettlementPending,
		Request: request, RequestedUsage: requestedUsage, PricingSnapshot: contractDeployment.EffectivePricing,
		EstimatedRating: ratedUsageFromBreakdown(estimate, contractDeployment), ExchangeRate: estimate.ExchangeRateExact,
		BudgetReservation: reservation, Version: 1, ExpiresAt: now.Add(videoRetention()),
		CreatedAt: now, UpdatedAt: now,
	}
	if _, err := h.jobs.CreateVideoJob(c.Context(), job); err != nil {
		releaseBudgetBestEffort(c.Context(), h.logger, h.billing, reservation, "job_persist_failed")
		return writeAPIError(c, fiber.StatusServiceUnavailable, "Unable to persist video job", "storage_error")
	}

	if err := h.submit(c.Context(), job, contractDeployment); err != nil {
		return writeAPIError(c, fiber.StatusBadGateway, "Video provider submission failed", "provider_error")
	}
	return c.Status(fiber.StatusAccepted).JSON(job)
}

func (h *VideoHandler) List(c *fiber.Ctx) error {
	authInfo, currentTenant, err := videoRequestContext(c)
	if err != nil {
		return err
	}
	scope, err := videoJobAccessScope(authInfo, currentTenant.ID)
	if err != nil {
		return writeAPIError(c, fiber.StatusForbidden, err.Error(), "authorization_error")
	}
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	var before time.Time
	if raw := strings.TrimSpace(c.Query("before")); raw != "" {
		before, err = time.Parse(time.RFC3339Nano, raw)
		if err != nil {
			return writeAPIError(c, fiber.StatusBadRequest, "Invalid before cursor", "invalid_request_error")
		}
	}
	items, err := h.jobs.ListVideoJobs(c.Context(), scope, limit, before)
	if err != nil {
		return writeAPIError(c, fiber.StatusServiceUnavailable, "Unable to list video jobs", "storage_error")
	}
	response := platform.VideoJobList{Data: items}
	if len(items) > 0 {
		response.NextCursor = items[len(items)-1].CreatedAt.Format(time.RFC3339Nano)
	}
	return c.JSON(response)
}

func (h *VideoHandler) Get(c *fiber.Ctx) error {
	authInfo, currentTenant, err := videoRequestContext(c)
	if err != nil {
		return err
	}
	scope, err := videoJobAccessScope(authInfo, currentTenant.ID)
	if err != nil {
		return writeAPIError(c, fiber.StatusForbidden, err.Error(), "authorization_error")
	}
	job, err := h.jobs.GetVideoJob(c.Context(), scope, c.Params("id"))
	if errors.Is(err, pgx.ErrNoRows) {
		return writeAPIError(c, fiber.StatusNotFound, "Video job not found", "not_found")
	}
	if err != nil {
		return writeAPIError(c, fiber.StatusServiceUnavailable, "Unable to load video job", "storage_error")
	}
	return c.JSON(job)
}

func (h *VideoHandler) Content(c *fiber.Ctx) error {
	authInfo, currentTenant, err := videoRequestContext(c)
	if err != nil {
		return err
	}
	scope, err := videoJobAccessScope(authInfo, currentTenant.ID)
	if err != nil {
		return writeAPIError(c, fiber.StatusForbidden, err.Error(), "authorization_error")
	}
	job, err := h.jobs.GetVideoJob(c.Context(), scope, c.Params("id"))
	if errors.Is(err, pgx.ErrNoRows) {
		return writeAPIError(c, fiber.StatusNotFound, "Video job not found", "not_found")
	}
	if err != nil || job.DeliveryStatus != platform.VideoDeliveryReady || len(job.Artifacts) == 0 {
		return writeAPIError(c, fiber.StatusConflict, "Video content is not ready", "content_not_ready")
	}
	if h.objects == nil {
		return writeAPIError(c, fiber.StatusServiceUnavailable, "Video object store is unavailable", "storage_error")
	}
	artifact := job.Artifacts[0]
	if requestedID := strings.TrimSpace(c.Query("artifact_id")); requestedID != "" {
		found := false
		for _, candidate := range job.Artifacts {
			if candidate.ID == requestedID {
				artifact = candidate
				found = true
				break
			}
		}
		if !found {
			return writeAPIError(c, fiber.StatusNotFound, "Video artifact not found", "not_found")
		}
	}
	signedURL, err := h.objects.PresignGet(c.Context(), artifact.ObjectKey, 15*time.Minute)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadGateway, "Unable to create video download URL", "storage_error")
	}
	return c.Redirect(signedURL, fiber.StatusTemporaryRedirect)
}

func (h *VideoHandler) Cancel(c *fiber.Ctx) error {
	authInfo, currentTenant, err := videoRequestContext(c)
	if err != nil {
		return err
	}
	scope, err := videoJobAccessScope(authInfo, currentTenant.ID)
	if err != nil {
		return writeAPIError(c, fiber.StatusForbidden, err.Error(), "authorization_error")
	}
	job, err := h.jobs.GetVideoJob(c.Context(), scope, c.Params("id"))
	if errors.Is(err, pgx.ErrNoRows) {
		return writeAPIError(c, fiber.StatusNotFound, "Video job not found", "not_found")
	}
	if err != nil {
		return writeAPIError(c, fiber.StatusServiceUnavailable, "Unable to load video job", "storage_error")
	}
	if isTerminalVideoStatus(job.GenerationStatus) {
		return c.JSON(job)
	}
	dep, err := NewVideoDeploymentResolver(h.router)(c.Context(), currentTenant.ID, job.Model, job.DeploymentID)
	if err != nil {
		return writeAPIError(c, fiber.StatusServiceUnavailable, "Unable to resolve video deployment", "routing_error")
	}
	if job.ProviderJobID != "" {
		adapter, adapterErr := h.adapters.Adapter(job.Provider)
		if adapterErr != nil {
			return writeAPIError(c, fiber.StatusBadGateway, adapterErr.Error(), "provider_error")
		}
		result, cancelErr := adapter.Cancel(c.Context(), job.RequestID, job.TenantID, dep, job.ProviderJobID)
		if cancelErr != nil {
			return writeAPIError(c, fiber.StatusBadGateway, "Provider cancellation failed", "provider_error")
		}
		if result.ProviderJobID != "" {
			job.ProviderJobID = result.ProviderJobID
		}
		if len(result.ActualUsage) > 0 {
			job.ActualUsage = result.ActualUsage
		}
		if len(result.Artifacts) > 0 {
			job.Artifacts = result.Artifacts
		}
		if result.Raw != nil {
			job.ProviderResponse = result.Raw
		}
	}
	expected := job.Version
	now := time.Now().UTC()
	job.GenerationStatus = platform.VideoGenerationCancelled
	job.CancelledAt = &now
	if job.ProviderJobID == "" {
		job.SettlementStatus = platform.VideoSettlementReleased
		if err := h.billing.ReleaseBudget(c.Context(), job.BudgetReservation, "video_cancelled_before_provider_submission"); err != nil {
			job.SettlementStatus = platform.VideoSettlementReview
			job.GenerationStatus = platform.VideoGenerationBillingReview
			job.LastError = err.Error()
		}
	} else if len(job.ActualUsage) == 0 {
		job.SettlementStatus = platform.VideoSettlementReview
		job.GenerationStatus = platform.VideoGenerationBillingReview
		job.LastError = "provider cancellation returned no provable actual usage"
	} else {
		job.SettlementStatus = platform.VideoSettlementPending
		job.NextPollAt = timePointer(now)
	}
	if err := h.jobs.SaveVideoJob(c.Context(), job, expected); err != nil {
		return writeAPIError(c, fiber.StatusConflict, "Video job changed concurrently", "conflict")
	}
	return c.JSON(job)
}

func (h *VideoHandler) Delete(c *fiber.Ctx) error {
	authInfo, currentTenant, err := videoRequestContext(c)
	if err != nil {
		return err
	}
	scope, err := videoJobAccessScope(authInfo, currentTenant.ID)
	if err != nil {
		return writeAPIError(c, fiber.StatusForbidden, err.Error(), "authorization_error")
	}
	job, err := h.jobs.GetVideoJob(c.Context(), scope, c.Params("id"))
	if errors.Is(err, pgx.ErrNoRows) {
		return writeAPIError(c, fiber.StatusNotFound, "Video job not found", "not_found")
	}
	if err != nil {
		return writeAPIError(c, fiber.StatusServiceUnavailable, "Unable to load video job", "storage_error")
	}
	if !isTerminalVideoStatus(job.GenerationStatus) {
		return writeAPIError(c, fiber.StatusConflict, "Cancel the running video job before deleting it", "conflict")
	}
	for _, artifact := range job.Artifacts {
		if artifact.ObjectKey != "" && h.objects != nil {
			if err := h.objects.Delete(c.Context(), artifact.ObjectKey); err != nil {
				return writeAPIError(c, fiber.StatusBadGateway, "Unable to delete video content", "storage_error")
			}
		}
	}
	expected := job.Version
	job.Artifacts = nil
	job.DeliveryStatus = platform.VideoDeliveryDeleted
	if err := h.jobs.SaveVideoJob(c.Context(), job, expected); err != nil {
		return writeAPIError(c, fiber.StatusConflict, "Video job changed concurrently", "conflict")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *VideoHandler) submit(ctx context.Context, job *platform.VideoGenerationJob, dep platform.Deployment) error {
	adapter, err := h.adapters.Adapter(job.Provider)
	if err != nil {
		return h.failSubmission(ctx, job, err)
	}
	_ = h.jobs.MarkVideoSubmission(ctx, job.ID, 1, "processing", nil, "")

	inputURL := ""
	if job.Request.InputMediaID != "" && h.objects != nil {
		inputURL, err = h.objects.PresignGet(ctx, job.Request.InputMediaID, 30*time.Minute)
		if err != nil {
			return h.failSubmission(ctx, job, err)
		}
	}
	result, err := adapter.Submit(ctx, job.RequestID, job.TenantID, dep, job.Request, inputURL)
	if err != nil {
		if videoruntime.IsProviderRejection(err) {
			return h.failSubmission(ctx, job, err)
		}
		expected := job.Version
		job.LastError = err.Error()
		next := time.Now().UTC().Add(time.Second)
		job.NextPollAt = &next
		if saveErr := h.jobs.SaveVideoJob(ctx, job, expected); saveErr != nil {
			return errors.Join(err, saveErr)
		}
		_ = h.jobs.MarkVideoSubmission(ctx, job.ID, 1, "processing", nil, err.Error())
		// The provider may have accepted the request before the connection was
		// interrupted. Keep the reservation and let the outbox worker retry with
		// the same provider idempotency key.
		return nil
	}

	expected := job.Version
	job.ProviderJobID = result.ProviderJobID
	job.GenerationStatus = result.Status
	job.ActualUsage = result.ActualUsage
	job.Artifacts = result.Artifacts
	job.ProviderResponse = result.Raw
	if job.ProviderResponse == nil {
		job.ProviderResponse = map[string]interface{}{}
	}
	job.ProviderResponse["_shahrah_request_exact"] = result.RequestIsExact
	next := time.Now().UTC().Add(result.PollAfter)
	job.NextPollAt = &next
	if result.Status == platform.VideoGenerationSucceeded {
		job.NextPollAt = timePointer(time.Now().UTC())
	}
	if err := h.jobs.SaveVideoJob(ctx, job, expected); err != nil {
		return err
	}
	_ = h.jobs.MarkVideoSubmission(ctx, job.ID, 1, "submitted", result.Raw, "")
	return nil
}

func (h *VideoHandler) failSubmission(ctx context.Context, job *platform.VideoGenerationJob, submissionErr error) error {
	expected := job.Version
	now := time.Now().UTC()
	job.GenerationStatus = platform.VideoGenerationFailed
	job.SettlementStatus = platform.VideoSettlementReleased
	job.LastError = submissionErr.Error()
	job.CompletedAt = &now
	if err := h.billing.ReleaseBudget(ctx, job.BudgetReservation, "provider_submission_failed"); err != nil {
		job.GenerationStatus = platform.VideoGenerationBillingReview
		job.SettlementStatus = platform.VideoSettlementReview
		job.LastError = errors.Join(submissionErr, err).Error()
	}
	_ = h.jobs.SaveVideoJob(ctx, job, expected)
	_ = h.jobs.MarkVideoSubmission(ctx, job.ID, 1, "failed", nil, submissionErr.Error())
	return submissionErr
}

func (h *VideoHandler) routeVideo(ctx context.Context, tenantID, model string) (*router.Deployment, error) {
	var routeErr error
	for _, candidate := range modelRoutingCandidates(model) {
		deployment, err := h.router.Route(ctx, tenantID, candidate)
		if err == nil && deployment != nil {
			return deployment, nil
		}
		routeErr = err
	}
	return nil, routeErr
}

func videoRequestContext(c *fiber.Ctx) (*platform.AuthContext, *platform.Tenant, error) {
	authInfo := auth.GetAuthInfo(c)
	if authInfo == nil {
		return nil, nil, writeAPIError(c, fiber.StatusUnauthorized, "Authentication required", "authentication_error")
	}
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return nil, nil, writeAPIError(c, fiber.StatusBadRequest, "Tenant resolution required", "invalid_request_error")
	}
	authInfo.TenantID = currentTenant.ID
	return authInfo, currentTenant, nil
}

func videoJobAccessScope(authInfo *platform.AuthContext, tenantID string) (storage.VideoJobAccessScope, error) {
	if authInfo == nil {
		return storage.VideoJobAccessScope{}, errors.New("authenticated video owner is required")
	}
	if strings.TrimSpace(authInfo.UserID) != "" && strings.TrimSpace(authInfo.VirtualUserID) != "" {
		return storage.VideoJobAccessScope{}, errors.New("ambiguous video owner identity")
	}
	if auth.IsTenantAdminOrHigher(authInfo) {
		return storage.VideoJobAccessScope{
			TenantID:        tenantID,
			AllowTenantWide: true,
		}, nil
	}
	ownerPrincipal, err := videoOwnerPrincipal(authInfo)
	if err != nil {
		return storage.VideoJobAccessScope{}, err
	}
	return storage.VideoJobAccessScope{
		TenantID:       tenantID,
		OwnerPrincipal: ownerPrincipal,
	}, nil
}

func videoOwnerPrincipal(authInfo *platform.AuthContext) (string, error) {
	if authInfo == nil {
		return "", errors.New("authenticated video owner is required")
	}
	userID := strings.TrimSpace(authInfo.UserID)
	virtualUserID := strings.TrimSpace(authInfo.VirtualUserID)
	tokenHash := strings.TrimSpace(authInfo.TokenHash)
	if userID != "" && virtualUserID != "" {
		return "", errors.New("ambiguous video owner identity")
	}
	if userID != "" {
		return "user:" + userID, nil
	}
	if virtualUserID != "" {
		return "virtual_user:" + virtualUserID, nil
	}
	if tokenHash != "" {
		return "token:" + tokenHash, nil
	}
	return "", errors.New("authenticated video owner is required")
}

func parseVideoCreateRequest(c *fiber.Ctx) (platform.VideoCreateRequest, *multipart.FileHeader, error) {
	var request platform.VideoCreateRequest
	contentType := strings.ToLower(c.Get("Content-Type"))
	if !strings.HasPrefix(contentType, "multipart/form-data") {
		if err := json.Unmarshal(c.Body(), &request); err != nil {
			return request, nil, errors.New("invalid request body")
		}
		return request, nil, nil
	}

	form, err := c.MultipartForm()
	if err != nil {
		return request, nil, errors.New("invalid multipart request")
	}
	value := func(key string) string {
		if values := form.Value[key]; len(values) > 0 {
			return strings.TrimSpace(values[0])
		}
		return ""
	}
	request.Model = value("model")
	request.Prompt = value("prompt")
	request.Size = value("size")
	request.Resolution = value("resolution")
	request.AspectRatio = value("aspect_ratio")
	request.ServiceTier = value("service_tier")
	request.Seconds, _ = strconv.Atoi(value("seconds"))
	request.Count, _ = strconv.Atoi(value("n"))
	if rawSeed := value("seed"); rawSeed != "" {
		seed, parseErr := strconv.ParseInt(rawSeed, 10, 64)
		if parseErr != nil {
			return request, nil, errors.New("seed must be an integer")
		}
		request.Seed = &seed
	}
	var file *multipart.FileHeader
	if files := form.File["input_reference"]; len(files) > 0 {
		file = files[0]
	}
	return request, file, nil
}

func validateVideoCreateRequest(request platform.VideoCreateRequest) error {
	if strings.TrimSpace(request.Model) == "" {
		return errors.New("model is required")
	}
	if strings.TrimSpace(request.Prompt) == "" && strings.TrimSpace(request.InputMediaID) == "" {
		return errors.New("prompt or input_reference is required")
	}
	if request.Seconds <= 0 || request.Seconds > 120 {
		return errors.New("seconds must be between 1 and 120")
	}
	if request.Count == 0 {
		request.Count = 1
	}
	if request.Count < 0 || request.Count > 4 {
		return errors.New("n must be between 1 and 4")
	}
	if strings.Contains(request.InputMediaID, "://") {
		return errors.New("arbitrary input URLs are not supported")
	}
	return nil
}

func (h *VideoHandler) storeVideoInput(ctx context.Context, tenantID, jobID string, header *multipart.FileHeader) (platform.VideoArtifact, error) {
	if header.Size > 20*1024*1024 {
		return platform.VideoArtifact{}, errors.New("input image exceeds 20 MiB")
	}
	contentType := header.Header.Get("Content-Type")
	if !strings.HasPrefix(strings.ToLower(contentType), "image/") {
		return platform.VideoArtifact{}, errors.New("input_reference must be an image")
	}
	file, err := header.Open()
	if err != nil {
		return platform.VideoArtifact{}, err
	}
	defer file.Close()
	return h.objects.PutInput(ctx, tenantID, jobID, header.Filename, contentType, file)
}

func videoRequestedUsage(request platform.VideoCreateRequest) []platform.UsageMeasure {
	count := request.Count
	if count <= 0 {
		count = 1
	}
	tier := strings.TrimSpace(request.ServiceTier)
	if tier == "" {
		tier = "standard"
	}
	resolution := strings.TrimSpace(request.Resolution)
	if resolution == "" {
		resolution = resolutionFromSize(request.Size)
	}
	return []platform.UsageMeasure{{
		Usage: "output", Unit: "second", Amount: strconv.Itoa(request.Seconds * count),
		Resolution: resolution, ServiceTier: tier, Qualifiers: map[string]interface{}{"output_count": count},
	}}
}

func resolutionFromSize(size string) string {
	switch strings.ToLower(strings.TrimSpace(size)) {
	case "960x540", "540x960":
		return "540p"
	case "1280x720", "720x1280":
		return "720p"
	case "1920x1080", "1080x1920":
		return "1080p"
	default:
		return ""
	}
}

func ratedUsageFromBreakdown(breakdown platform.UsageCostBreakdown, dep platform.Deployment) *platform.RatedUsage {
	totalUSD := firstNonEmptyString(breakdown.TotalCostUSDExact, "0")
	exchangeRate := firstNonEmptyString(breakdown.ExchangeRateExact, "0")
	totalIRR := firstNonEmptyString(breakdown.TotalCostIRRExact, "0")
	return &platform.RatedUsage{
		Status: breakdown.RatingStatus, Model: breakdown.Model, Provider: dep.Provider,
		Region: dep.RegionName, PricingSource: breakdown.PricingSource,
		Measures: breakdown.UsageMeasures, Components: breakdown.AppliedRates,
		TotalCostUSD: totalUSD, ExchangeRate: exchangeRate,
		TotalCostIRR: totalIRR, PricingSnapshot: breakdown.PricingSnapshot,
		SnapshotHash: breakdown.PricingSnapshotHash,
	}
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func isTerminalVideoStatus(status string) bool {
	switch status {
	case platform.VideoGenerationSucceeded, platform.VideoGenerationFailed,
		platform.VideoGenerationCancelled, platform.VideoGenerationBillingReview:
		return true
	default:
		return false
	}
}

func videoRetention() time.Duration {
	return time.Duration(videoruntime.RetentionDays()) * 24 * time.Hour
}

func timePointer(value time.Time) *time.Time {
	return &value
}
