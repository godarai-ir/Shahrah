package proxy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	metricsobs "github.com/Delpak/shahrah/services/control-plane/internal/observability/metrics"
	"github.com/Delpak/shahrah/services/control-plane/internal/observability/tracing"
	"github.com/Delpak/shahrah/services/control-plane/internal/providernative"
	"github.com/Delpak/shahrah/services/control-plane/internal/router"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

const (
	videoResourceKind = "video"
	// Request shapes a job was created with; GET answers in the same shape.
	videoShapeOpenAI = "openai"
	videoShapeXAI    = "xai"
	videoRoute       = "/v1/videos"
	videoRequestType = "videos"
	// maxVideoRequestBytes bounds create bodies, which may carry inline images.
	maxVideoRequestBytes = 32 << 20
	// minVideoRefreshInterval caps how often client GETs reach the provider.
	minVideoRefreshInterval = 2 * time.Second
)

// Metadata keys persisted on a video job record.
const (
	videoMetaModel          = "model"
	videoMetaShape          = "api_shape"
	videoMetaOperation      = "operation"
	videoMetaCreatedAt      = "created_at"
	videoMetaCompletedAt    = "completed_at"
	videoMetaProgress       = "progress"
	videoMetaPolledAt       = "polled_at"
	videoMetaSeconds        = "seconds"
	videoMetaSize           = "size"
	videoMetaSourceID       = "source_video_id"
	videoMetaErrorCode      = "error_code"
	videoMetaErrorMessage   = "error_message"
	videoMetaDeletedAt      = "deleted_at"
	videoMetaRequestID      = "request_id"
	videoMetaTokenHash      = "token_hash"
	videoMetaTeamID         = "team_id"
	videoMetaUserID         = "user_id"
	videoMetaBillingMode    = "billing_mode"
	videoMetaRequestMeta    = "request_metadata"
	videoMetaSettledCostUSD = "settled_cost_usd"
	videoMetaSettlement     = "settlement"
)

var errVideoNotFound = errors.New("video not found")

type videoResourceStore interface {
	providerNativeResourceStore
	ClaimProviderNativeResourceStatus(ctx context.Context, tenantID, provider, resourceKind, publicID string, fromStatuses []string, toStatus string) (bool, error)
	ListPendingProviderNativeResourceMappings(ctx context.Context, provider, resourceKind string, statuses []string, limit int) ([]platform.ProviderNativeResourceMapping, error)
	FindTenantProviderNativeResourceMapping(ctx context.Context, tenantID, resourceKind, publicID string) (*platform.ProviderNativeResourceMapping, error)
	ListTenantProviderNativeResourceMappings(ctx context.Context, tenantID, resourceKind, after string, ascending bool, limit int) ([]platform.ProviderNativeResourceMapping, error)
	PatchProviderNativeResourceState(ctx context.Context, tenantID, provider, resourceKind, publicID, expectedStatus, status string, patch map[string]interface{}) (bool, error)
}

// VideosHandler serves the provider-neutral /v1/videos surface and the
// xAI-native video routes, and settles every job exactly once.
type VideosHandler struct {
	router        *router.Router
	pipeline      *RequestPipeline
	billing       budgeter
	resources     videoResourceStore
	logs          usageLogger
	logger        *zap.Logger
	emitMetric    func(input metricsobs.BuildInput)
	codecs        map[string]videoProviderCodec
	enabled       bool
	canaryTenants map[string]struct{}
	maxJobAge     time.Duration
	now           func() time.Time
}

// NewVideosHandler creates the video handler. It serves only when the rollout
// flag is on, and only canary tenants when a canary list is configured.
func NewVideosHandler(deps BaseHandlerDependencies, resources videoResourceStore) *VideosHandler {
	canary := make(map[string]struct{}, len(deps.ProviderFeatures.VideoCanaryTenantIDs))
	for _, tenantID := range deps.ProviderFeatures.VideoCanaryTenantIDs {
		if trimmed := strings.TrimSpace(tenantID); trimmed != "" {
			canary[trimmed] = struct{}{}
		}
	}

	return &VideosHandler{
		router:        deps.Router,
		pipeline:      deps.requestPipeline(),
		billing:       deps.Billing,
		resources:     resources,
		logs:          deps.logger(),
		logger:        deps.Logger,
		emitMetric:    deps.metricEmitter(),
		codecs:        map[string]videoProviderCodec{"xai": xaiVideoCodec{}},
		enabled:       deps.ProviderFeatures.Configured && deps.ProviderFeatures.VideosEnabled,
		canaryTenants: canary,
		maxJobAge:     2 * time.Hour,
		now:           time.Now,
	}
}

func (h *VideosHandler) servesTenant(tenantID string) bool {
	if !h.enabled || h.resources == nil {
		return false
	}

	if len(h.canaryTenants) == 0 {
		return true
	}

	_, ok := h.canaryTenants[tenantID]

	return ok
}

// videoRequestContext authenticates the caller and enforces the rollout gate.
// Disabled tenants see 404 so the surface stays invisible outside the canary.
func (h *VideosHandler) videoRequestContext(c *fiber.Ctx) (*platform.AuthContext, *platform.Tenant, error) {
	authInfo := auth.GetAuthInfo(c)
	if authInfo == nil {
		return nil, nil, writeAPIError(c, fiber.StatusUnauthorized, "Authentication required", "authentication_error")
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return nil, nil, writeAPIError(c, fiber.StatusBadRequest, "Tenant resolution required", "invalid_request_error")
	}

	authInfo.TenantID = currentTenant.ID
	tracing.SetTenantForTrace(c, currentTenant.ID)

	if !h.servesTenant(currentTenant.ID) {
		return nil, nil, writeAPIErrorDetail(c, fiber.StatusNotFound, "Video generation is not available for this workspace", "invalid_request_error", "", "unsupported_endpoint")
	}

	return authInfo, currentTenant, nil
}

// HandleCreate serves OpenAI-style POST /v1/videos.
func (h *VideosHandler) HandleCreate(c *fiber.Ctx) error {
	return h.create(c, platform.VideoOperationGenerate, videoShapeOpenAI)
}

// HandleGenerations serves xAI-native POST /v1/videos/generations.
func (h *VideosHandler) HandleGenerations(c *fiber.Ctx) error {
	return h.create(c, platform.VideoOperationGenerate, videoShapeXAI)
}

// HandleEdits serves POST /v1/videos/edits in either request shape.
func (h *VideosHandler) HandleEdits(c *fiber.Ctx) error {
	return h.create(c, platform.VideoOperationEdit, videoSourceShape(c))
}

// HandleExtensions serves POST /v1/videos/extensions in either request shape.
func (h *VideosHandler) HandleExtensions(c *fiber.Ctx) error {
	return h.create(c, platform.VideoOperationExtend, videoSourceShape(c))
}

// videoSourceShape detects the OpenAI shape by its `video: {id}` reference.
func videoSourceShape(c *fiber.Ctx) string {
	var probe struct {
		Video *struct {
			ID string `json:"id"`
		} `json:"video"`
	}

	if json.Unmarshal(c.Body(), &probe) == nil && probe.Video != nil && strings.TrimSpace(probe.Video.ID) != "" {
		return videoShapeOpenAI
	}

	return videoShapeXAI
}

type videoCreatePlan struct {
	requestedModel string
	routed         *router.Deployment
	deployment     platform.Deployment
	codec          videoProviderCodec
	profile        platform.ModelRuntimeProfile
	body           map[string]interface{}
	size           string
	sourceID       string
}

func (h *VideosHandler) create(c *fiber.Ctx, operation, shape string) error {
	startTime := h.now()

	authInfo, currentTenant, err := h.videoRequestContext(c)
	if authInfo == nil {
		return err
	}

	if len(c.Body()) > maxVideoRequestBytes {
		return writeAPIErrorDetail(c, fiber.StatusRequestEntityTooLarge, "Video request body is too large", "invalid_request_error", "", "request_too_large")
	}

	requestLog := newRequestLogLifecycle(h.logger, h.logs, c, authInfo, c.Path(), videoRequestType, startTime)
	defer requestLog.FinalizeFromFiber()

	canonical, native, err := parseVideoCreateRequest(c, operation, shape)
	if err != nil {
		return writeAPIErrorDetail(c, fiber.StatusBadRequest, err.Error(), "invalid_request_error", "", "invalid_video_request")
	}

	requestedModel := normalizeModelIdentifier(canonical.Model)
	if shape == videoShapeXAI {
		requestedModel = normalizeModelIdentifier(stringValue(native["model"]))
	}

	requestLog.SetModel(requestedModel)
	requestLog.SetRequestBody(map[string]interface{}{"model": requestedModel, "operation": operation, "api_shape": shape})

	plan, status, err := h.planVideoCreate(c, authInfo, currentTenant.ID, operation, shape, requestedModel, canonical, native)
	if err != nil {
		return writeAPIErrorDetail(c, status, err.Error(), videoErrorType(status), "", "invalid_video_request")
	}

	requestLog.SetDeployment(plan.deployment)

	return h.dispatchVideoCreate(c, authInfo, currentTenant.ID, operation, shape, plan, startTime, requestLog)
}

// planVideoCreate routes the model, translates and validates the request
// against the pinned profile. Nothing is reserved or dispatched yet.
func (h *VideosHandler) planVideoCreate(
	c *fiber.Ctx,
	authInfo *platform.AuthContext,
	tenantID, operation, shape, requestedModel string,
	canonical canonicalVideoRequest,
	native map[string]interface{},
) (videoCreatePlan, int, error) {
	plan := videoCreatePlan{requestedModel: requestedModel}
	if requestedModel == "" {
		return plan, fiber.StatusBadRequest, fmt.Errorf("model is required")
	}

	if access := modelAccessForType(authInfo, requestedModel, platform.ModelTypeVideoGeneration); !access.Allowed {
		emitDenyCounter(h.logger, denyReasonForModel(authInfo), requestedModel, tenantID)

		return plan, fiber.StatusForbidden, fmt.Errorf("model access denied or unsupported for video generation")
	}

	var dep *router.Deployment

	routeErr := fmt.Errorf("no available deployment for model: %s", requestedModel)
	for _, candidate := range modelRoutingCandidates(requestedModel) {
		if routed, err := h.router.Route(c.Context(), tenantID, candidate); err == nil && routed != nil {
			dep, routeErr = routed, nil

			break
		}
	}

	if routeErr != nil {
		return plan, fiber.StatusServiceUnavailable, routeErr
	}

	provider := strings.ToLower(strings.TrimSpace(dep.Provider))

	codec, ok := h.codecs[provider]
	if !ok || (shape == videoShapeXAI && provider != "xai") {
		return plan, fiber.StatusBadRequest, fmt.Errorf("model %s does not support video generation", requestedModel)
	}

	plan.codec, plan.routed, plan.deployment = codec, dep, toContractDeployment(dep)

	profile, err := h.videoProfile(c.Context(), billingModelForDeployment(requestedModel, plan.deployment), dep.Provider, providerOutboundModel(requestedModel, dep))
	if err != nil {
		return plan, fiber.StatusServiceUnavailable, err
	}

	plan.profile = profile
	if !profile.AllowsVideoOperation(operation) {
		return plan, fiber.StatusBadRequest, fmt.Errorf("model %s does not support video %s", requestedModel, operation)
	}

	if shape == videoShapeXAI {
		plan.body = native
	} else {
		if sourceErr := h.resolveCanonicalSource(c.Context(), tenantID, plan, &canonical); sourceErr != nil {
			return plan, fiber.StatusBadRequest, sourceErr
		}

		if fieldErr := checkVideoExtensionFields(canonical.Extensions, profile); fieldErr != nil {
			return plan, fiber.StatusBadRequest, fieldErr
		}

		plan.body, err = codec.buildCreateBody(canonical, profile)
		if err != nil {
			return plan, fiber.StatusBadRequest, err
		}

		plan.size, plan.sourceID = canonical.Size, stringValue(canonical.Source["id"])
	}

	plan.body["model"] = providerOutboundModel(requestedModel, dep)

	return plan, fiber.StatusOK, nil
}

// videoProfile resolves the pinned runtime profile through the pricing
// snapshot, the same source billing settles against.
func (h *VideosHandler) videoProfile(ctx context.Context, billingModel, provider, outboundModel string) (platform.ModelRuntimeProfile, error) {
	probe, err := priceMeteredUsage(ctx, h.pipeline, billingModel, videoOutputSecondsDimension, "1", platform.UsageCostBreakdown{Model: billingModel})
	if err != nil || probe.ModelRuntimeProfile == nil {
		return platform.ModelRuntimeProfile{}, fmt.Errorf("video pricing is unavailable for this model")
	}

	if _, err := runtimeProfilesFromEstimate(provider, outboundModel, probe); err != nil {
		return platform.ModelRuntimeProfile{}, err
	}

	return probe.ModelRuntimeProfile.Profile, nil
}

// resolveCanonicalSource turns an OpenAI `video: {id}` into the provider's
// source reference. Only the caller's own completed videos can be used.
func (h *VideosHandler) resolveCanonicalSource(ctx context.Context, tenantID string, plan videoCreatePlan, canonical *canonicalVideoRequest) error {
	if canonical.Operation == platform.VideoOperationGenerate {
		return nil
	}

	sourceID := stringValue(canonical.Source["id"])

	source, err := h.findVideo(ctx, tenantID, sourceID)
	if err != nil || source.Status != videoStatusCompleted ||
		!strings.EqualFold(source.ProviderAccountID, plan.deployment.ProviderAccountID) {
		return fmt.Errorf("video %q is not a completed video of this provider account", sourceID)
	}

	poll, err := h.pollVideo(ctx, source, plan.deployment)
	if err != nil {
		return err
	}

	reference, err := plan.codec.sourceReference(poll)
	if err != nil {
		return err
	}

	canonical.Source = reference
	canonical.Source["id"] = sourceID

	return nil
}

func (h *VideosHandler) dispatchVideoCreate(
	c *fiber.Ctx,
	authInfo *platform.AuthContext,
	tenantID, operation, shape string,
	plan videoCreatePlan,
	startTime time.Time,
	requestLog *requestLogLifecycle,
) error {
	sourceID := plan.sourceID
	delete(plan.body, "id")

	if source, ok := plan.body["video"].(map[string]interface{}); ok {
		delete(source, "id")
	}

	body, err := json.Marshal(plan.body)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadRequest, "Invalid video request body", "invalid_request_error")
	}

	if strings.EqualFold(plan.deployment.Provider, "xai") {
		body, err = rewriteXAIFileReferences(c.Context(), h.resources, tenantID, plan.deployment.ProviderAccountID, body)
		if err != nil {
			return writeXAIFileReferenceError(c, err)
		}
	}

	facts, err := plan.codec.validateBody(operation, body, plan.profile)
	if err != nil {
		return writeAPIErrorDetail(c, fiber.StatusBadRequest, err.Error(), "invalid_request_error", "", "invalid_video_request")
	}

	billingModel := billingModelForDeployment(plan.requestedModel, plan.deployment)
	seconds := videoReservationSeconds(facts, plan.profile.VideoControls)

	estimated, err := priceMeteredUsage(c.Context(), h.pipeline, billingModel, videoOutputSecondsDimension, strconv.Itoa(seconds),
		platform.UsageCostBreakdown{Model: billingModel, Endpoint: videoRoute, RequestType: videoRequestType})
	if err != nil || estimated.TotalCostUSD <= 0 {
		return writeAPIError(c, fiber.StatusServiceUnavailable, "Video pricing is unavailable for this model", "billing_error")
	}

	estimated, err = h.pipeline.applyDeploymentPricingOverride(c.Context(), estimated, plan.routed)
	if err != nil {
		return writeBillingEstimateError(c, err)
	}

	rateLimit, err := h.billing.CheckRateLimit(c.Context(), authInfo, plan.requestedModel)
	setRateLimitHeaders(c, rateLimit)

	if err != nil {
		if recordRateLimitRejection(c, h.logger, h.logs, h.emitMetric, authInfo, plan.requestedModel, c.Path(), videoRequestType, startTime, startTime, rateLimit, err) {
			requestLog.MarkFinalized()
		}

		return writeAPIError(c, rateLimitErrorStatus(err), err.Error(), "rate_limit_error")
	}

	requestID := tracing.TraceIDFromContext(c)
	if requestID == "" {
		requestID = fmt.Sprintf("req_%d", h.now().UnixNano())
	}

	reservation, err := h.pipeline.Reserve(c.Context(), budgetRequestFromHTTP(c, authInfo, estimated, requestID, deploymentBillingMode(plan.deployment)))
	if err != nil {
		return writeBudgetCheckError(c, err)
	}

	committed := false
	defer func() {
		if !committed {
			h.pipeline.Release(c.Context(), reservation, "request_not_committed")
		}
	}()

	if err = h.pipeline.EnsureReady(c.Context()); err != nil {
		return writeAPIError(c, fiber.StatusServiceUnavailable, "Tunnel path unavailable", "tunnel_unavailable")
	}

	profiles, err := runtimeProfilesFromEstimate(plan.deployment.Provider, stringValue(plan.body["model"]), estimated)
	if err != nil {
		return writeAPIError(c, fiber.StatusServiceUnavailable, err.Error(), "model_profile_error")
	}

	requestMetadata := enrichPricingMetadata(enrichRequestMetadataForDeployment(map[string]interface{}{
		"tenant_slug": strings.TrimSpace(tenant.FromContext(c).Slug), "model": plan.requestedModel,
		"operation": operation, "billing_mode": deploymentBillingMode(plan.deployment),
	}, plan.deployment), authInfo, plan.requestedModel)

	tunnelReq := platform.TunnelRequest{
		RequestID: requestID, TenantID: tenantID, TraceID: requestID,
		Operation:            plan.codec.createOperation(operation),
		Deployment:           plan.deployment,
		Headers:              map[string]string{"Content-Type": "application/json"},
		Body:                 body,
		RequestMetadata:      requestMetadata,
		ModelRuntimeProfiles: profiles,
	}
	requestLog.SetRequestIDs(requestID, requestID)

	response, err := h.pipeline.Dispatch(c.Context(), tunnelReq)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadGateway, fmt.Sprintf("Tunnel error: %v", err), "api_error")
	}

	if response.StatusCode >= httpErrorFloor {
		return writeCompatibleProviderError(c, plan.deployment.Provider, response.StatusCode, response.Body, response.Headers)
	}

	upstreamID, err := plan.codec.parseCreateResponse(response.Body)
	if err != nil {
		return writeAPIError(c, fiber.StatusBadGateway, err.Error(), "api_error")
	}

	estimated.Metadata = requestMetadata
	mapping := platform.ProviderNativeResourceMapping{
		PublicID: newVideoPublicID(), TenantID: tenantID, Provider: strings.ToLower(plan.deployment.Provider),
		ProviderAccountID: plan.deployment.ProviderAccountID, ResourceKind: videoResourceKind,
		UpstreamID: upstreamID, Status: videoStatusQueued, ReservationID: requestID, Reservation: reservation,
		PricingRevision: estimated.PricingRevision, EstimatedUsage: &estimated,
		Metadata: map[string]interface{}{
			videoMetaModel: plan.requestedModel, videoMetaShape: shape, videoMetaOperation: operation,
			videoMetaCreatedAt: h.now().Unix(), videoMetaProgress: 0, videoMetaSeconds: seconds,
			videoMetaSize: plan.size, videoMetaSourceID: sourceID, videoMetaRequestID: requestID,
			videoMetaTokenHash: authInfo.TokenHash, videoMetaTeamID: authInfo.TeamID, videoMetaUserID: authInfo.UserID,
			videoMetaBillingMode: deploymentBillingMode(plan.deployment), videoMetaRequestMeta: requestMetadata,
		},
	}

	committed = true

	stored, err := h.resources.UpsertProviderNativeResourceMapping(c.Context(), mapping)
	if err != nil {
		// The provider accepted a billable job we cannot track: charge the
		// estimate rather than leak free usage, and surface the failure.
		h.chargeVideo(c.Context(), &mapping, estimated, "untracked")
		h.logWarn("video job could not be recorded; charged at estimate", err, mapping.PublicID)

		return writeAPIError(c, fiber.StatusInternalServerError, "The video job could not be recorded", "api_error")
	}

	setTraceHeaders(c, requestID)
	setCostHeaders(c, estimated)

	if shape == videoShapeXAI {
		return c.JSON(fiber.Map{"request_id": stored.PublicID})
	}

	return c.JSON(renderOpenAIVideo(stored))
}

// HandleGet serves GET /v1/videos/{video_id} in the shape the job was made with.
func (h *VideosHandler) HandleGet(c *fiber.Ctx) error {
	authInfo, currentTenant, err := h.videoRequestContext(c)
	if authInfo == nil {
		return err
	}

	return h.respondWithVideo(c, currentTenant.ID, c.Params("video_id"))
}

func (h *VideosHandler) respondWithVideo(c *fiber.Ctx, tenantID, publicID string) error {
	mapping, err := h.findVideo(c.Context(), tenantID, publicID)
	if err != nil {
		return writeVideoLookupError(c, err)
	}

	if isPendingVideoStatus(mapping.Status) && h.refreshDue(mapping) {
		// Refresh through the shared, idempotent settlement path.
		h.reconcileVideoJob(c.Context(), mapping)

		if refreshed, refreshErr := h.findVideo(c.Context(), tenantID, publicID); refreshErr == nil {
			mapping = refreshed
		}
	}

	return c.JSON(renderVideo(c, mapping))
}

// refreshDue reports whether a client GET may poll the provider. Faster
// pollers get the stored state; the worker still polls on its own schedule.
func (h *VideosHandler) refreshDue(mapping *platform.ProviderNativeResourceMapping) bool {
	polledAt := int64(intValue(mapping.Metadata[videoMetaPolledAt]))

	return polledAt == 0 || h.now().Sub(time.UnixMilli(polledAt)) >= minVideoRefreshInterval
}

// HandleList serves GET /v1/videos from GodarAI's own job records, because
// providers such as xAI expose no list endpoint.
func (h *VideosHandler) HandleList(c *fiber.Ctx) error {
	authInfo, currentTenant, err := h.videoRequestContext(c)
	if authInfo == nil {
		return err
	}

	limit := 20

	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		value, parseErr := strconv.Atoi(raw)
		if parseErr != nil || value < 1 || value > 100 {
			return writeAPIErrorDetail(c, fiber.StatusBadRequest, "limit must be between 1 and 100", "invalid_request_error", "limit", "invalid_value")
		}

		limit = value
	}

	order := strings.ToLower(strings.TrimSpace(c.Query("order", "desc")))
	if order != "asc" && order != "desc" {
		return writeAPIErrorDetail(c, fiber.StatusBadRequest, "order must be asc or desc", "invalid_request_error", "order", "invalid_value")
	}

	mappings, err := h.resources.ListTenantProviderNativeResourceMappings(c.Context(), currentTenant.ID, videoResourceKind, strings.TrimSpace(c.Query("after")), order == "asc", limit+1)
	if err != nil {
		return writeAPIError(c, fiber.StatusServiceUnavailable, "Video jobs are unavailable", "api_error")
	}

	hasMore := len(mappings) > limit
	if hasMore {
		mappings = mappings[:limit]
	}

	data := make([]map[string]interface{}, 0, len(mappings))
	for index := range mappings {
		data = append(data, renderOpenAIVideo(&mappings[index]))
	}

	list := map[string]interface{}{"object": "list", "data": data, "first_id": nil, "last_id": nil, "has_more": hasMore}
	if len(mappings) > 0 {
		list["first_id"], list["last_id"] = mappings[0].PublicID, mappings[len(mappings)-1].PublicID
	}

	return c.JSON(list)
}

// HandleDelete hides a job from the tenant. Upstream jobs cannot be canceled,
// so a pending job keeps settling in the background.
func (h *VideosHandler) HandleDelete(c *fiber.Ctx) error {
	authInfo, currentTenant, err := h.videoRequestContext(c)
	if authInfo == nil {
		return err
	}

	mapping, err := h.findVideo(c.Context(), currentTenant.ID, c.Params("video_id"))
	if err != nil {
		return writeVideoLookupError(c, err)
	}

	// Only the metadata changes: a pending job's status stays with the worker.
	if _, err := h.resources.PatchProviderNativeResourceState(c.Context(), mapping.TenantID, mapping.Provider, videoResourceKind,
		mapping.PublicID, "", "", map[string]interface{}{videoMetaDeletedAt: h.now().Unix()}); err != nil {
		return writeAPIError(c, fiber.StatusServiceUnavailable, "The video could not be deleted", "api_error")
	}

	return c.JSON(fiber.Map{"id": mapping.PublicID, "object": "video.deleted", "deleted": true})
}

// HandleContent streams the finished MP4 through the provider egress.
func (h *VideosHandler) HandleContent(c *fiber.Ctx) error {
	authInfo, currentTenant, err := h.videoRequestContext(c)
	if authInfo == nil {
		return err
	}

	mapping, err := h.findVideo(c.Context(), currentTenant.ID, c.Params("video_id"))
	if err != nil {
		return writeVideoLookupError(c, err)
	}

	if mapping.Status != videoStatusCompleted {
		return writeAPIErrorDetail(c, fiber.StatusConflict, "The video is not completed", "invalid_request_error", "", "video_not_ready")
	}

	profile := videoMappingProfile(mapping)
	if variant := strings.TrimSpace(c.Query("variant", "video")); profile == nil ||
		!containsFoldValue(profile.Profile.VideoControls.ContentVariants, variant) {
		return writeAPIErrorDetail(c, fiber.StatusBadRequest, fmt.Sprintf("variant %q is not available for this video", variant), "invalid_request_error", "variant", "unsupported_variant")
	}

	if !h.pipeline.HasRawTransport() {
		return writeAPIError(c, fiber.StatusServiceUnavailable, "Tunnel path unavailable", "tunnel_unavailable")
	}

	deployment, codec, err := h.mappingDeployment(c.Context(), mapping)
	if err != nil {
		return writeAPIError(c, fiber.StatusServiceUnavailable, err.Error(), "routing_error")
	}

	result, err := h.pipeline.DispatchRaw(c.Context(), platform.TunnelRequest{
		RequestID: fmt.Sprintf("video_content_%d", h.now().UnixNano()), TenantID: mapping.TenantID,
		Operation:            codec.contentOperation(),
		OperationInput:       platform.ProviderOperationInput{ResourceID: mapping.UpstreamID},
		Deployment:           deployment,
		ModelRuntimeProfiles: []platform.ModelRuntimeProfileBinding{*profile},
	}, nil, 0)
	if err != nil || result == nil || result.Body == nil {
		return writeAPIError(c, fiber.StatusBadGateway, "The video content could not be fetched", "api_error")
	}

	if result.StatusCode >= httpErrorFloor {
		defer result.Body.Close() //nolint:errcheck // Cleanup cannot change the error response.

		errorBody, _ := io.ReadAll(io.LimitReader(result.Body, 64*1024)) //nolint:errcheck // A truncated error body still maps to a stable error.

		return writeCompatibleProviderError(c, deployment.Provider, result.StatusCode, errorBody, result.Headers)
	}

	c.Set(fiber.HeaderContentType, "video/mp4")

	if contentType := responseContentType(result.Headers); contentType != "" {
		c.Set(fiber.HeaderContentType, contentType)
	}

	// A known length is sent as Content-Length; otherwise the body is chunked.
	// fasthttp streams the body and closes it once the response is written.
	length := -1
	if parsed, parseErr := strconv.Atoi(headerValueFold(result.Headers, "content-length")); parseErr == nil && parsed >= 0 {
		length = parsed
	}

	c.Status(result.StatusCode)
	c.Context().SetBodyStream(result.Body, length)

	return nil
}

func (h *VideosHandler) findVideo(ctx context.Context, tenantID, publicID string) (*platform.ProviderNativeResourceMapping, error) {
	publicID = strings.TrimSpace(publicID)
	if publicID == "" {
		return nil, errVideoNotFound
	}

	mapping, err := h.resources.FindTenantProviderNativeResourceMapping(ctx, tenantID, videoResourceKind, publicID)
	if errors.Is(err, providernative.ErrResourceNotFound) || (err == nil && mapping == nil) {
		return nil, errVideoNotFound
	}

	if err != nil {
		return nil, err
	}

	if mapping.Metadata == nil {
		mapping.Metadata = map[string]interface{}{}
	}

	if _, deleted := mapping.Metadata[videoMetaDeletedAt]; deleted {
		return nil, errVideoNotFound
	}

	return mapping, nil
}

// mappingDeployment re-resolves the job's deployment, keeping provider-account
// affinity: polls and downloads must use the account that created the job.
func (h *VideosHandler) mappingDeployment(ctx context.Context, mapping *platform.ProviderNativeResourceMapping) (platform.Deployment, videoProviderCodec, error) {
	codec, ok := h.codecs[mapping.Provider]
	if !ok {
		return platform.Deployment{}, nil, fmt.Errorf("video provider %q is not supported", mapping.Provider)
	}

	routed, err := h.router.Route(ctx, mapping.TenantID, stringValue(mapping.Metadata[videoMetaModel]))
	if err != nil || routed == nil || !strings.EqualFold(routed.ProviderAccountID, mapping.ProviderAccountID) {
		return platform.Deployment{}, nil, fmt.Errorf("the provider account that created this video is unavailable")
	}

	return toContractDeployment(routed), codec, nil
}

func (h *VideosHandler) pollVideo(ctx context.Context, mapping *platform.ProviderNativeResourceMapping, deployment platform.Deployment) ([]byte, error) {
	codec := h.codecs[mapping.Provider]
	profile := videoMappingProfile(mapping)

	if codec == nil || profile == nil {
		return nil, fmt.Errorf("video job is missing its provider profile")
	}

	response, err := h.pipeline.Dispatch(ctx, platform.TunnelRequest{
		RequestID: fmt.Sprintf("video_poll_%d", h.now().UnixNano()), TenantID: mapping.TenantID,
		Operation:            codec.pollOperation(),
		OperationInput:       platform.ProviderOperationInput{ResourceID: mapping.UpstreamID},
		Deployment:           deployment,
		ModelRuntimeProfiles: []platform.ModelRuntimeProfileBinding{*profile},
	})
	if err != nil {
		return nil, err
	}

	if response.StatusCode == http.StatusNotFound {
		return nil, errVideoUpstreamGone
	}

	if response.StatusCode >= httpErrorFloor {
		return nil, fmt.Errorf("video status returned HTTP %d", response.StatusCode)
	}

	return response.Body, nil
}

func videoMappingProfile(mapping *platform.ProviderNativeResourceMapping) *platform.ModelRuntimeProfileBinding {
	if mapping == nil || mapping.EstimatedUsage == nil {
		return nil
	}

	return mapping.EstimatedUsage.ModelRuntimeProfile
}
