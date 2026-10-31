package proxy

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	metricsobs "github.com/Delpak/shahrah/services/control-plane/internal/observability/metrics"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

const (
	videoOutputSecondsDimension = "output_video_seconds"
	// Transient worker states: a claimed poll, and a settlement in flight.
	videoStatusReconciling = "reconciling"
	videoStatusSettling    = "settling"
)

var (
	errVideoUpstreamGone        = errors.New("video job is no longer available upstream")
	errVideoUsagePricingUnwired = errors.New("video usage pricing is unavailable")
	videoWorkerStatuses         = []string{videoStatusQueued, videoStatusInProgress, videoStatusReconciling, videoStatusSettling}
)

type videoUsagePricer interface {
	PriceVideoUsage(ctx context.Context, model string, seconds int, providerCostUSDTicks *int64, estimated platform.UsageCostBreakdown) (platform.UsageCostBreakdown, error)
}

func (p *RequestPipeline) priceVideoUsage(
	ctx context.Context,
	model string,
	seconds int,
	providerCostUSDTicks *int64,
	estimated platform.UsageCostBreakdown,
) (platform.UsageCostBreakdown, error) {
	if p == nil || p.billingCapabilities.VideoUsage == nil {
		return platform.UsageCostBreakdown{}, errVideoUsagePricingUnwired
	}

	return p.billingCapabilities.VideoUsage.PriceVideoUsage(ctx, model, seconds, providerCostUSDTicks, estimated)
}

// RunVideoWorker polls pending video jobs and settles each reservation once.
// It keeps running while routes are disabled so in-flight jobs still settle.
func (h *VideosHandler) RunVideoWorker(ctx context.Context, interval time.Duration) {
	if h == nil || h.resources == nil || !h.pipeline.HasTransport() {
		return
	}

	if interval <= 0 {
		interval = 5 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		h.reconcileVideos(ctx)

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (h *VideosHandler) reconcileVideos(ctx context.Context) {
	for provider := range h.codecs {
		mappings, err := h.resources.ListPendingProviderNativeResourceMappings(ctx, provider, videoResourceKind, videoWorkerStatuses, 100)
		if err != nil {
			h.logWarn("video reconciliation listing failed", err, "")

			continue
		}

		for index := range mappings {
			h.reconcileVideoJob(ctx, &mappings[index])
		}
	}
}

// reconcileVideoJob claims one job, polls it, and either records progress or
// settles it. Every caller (worker or GET refresh) goes through this path.
func (h *VideosHandler) reconcileVideoJob(ctx context.Context, mapping *platform.ProviderNativeResourceMapping) {
	if mapping.Status == videoStatusSettling {
		// Only a settlement interrupted for over five minutes is reclaimed.
		if h.claimVideo(ctx, mapping, nil, videoStatusSettling) {
			h.abandonInterruptedSettlement(ctx, mapping)
		}

		return
	}

	from := []string{mapping.Status}
	if mapping.Status == videoStatusReconciling {
		from = nil
	}

	if !h.claimVideo(ctx, mapping, from, videoStatusReconciling) {
		return
	}

	state, err := h.currentVideoState(ctx, mapping)
	if err != nil {
		h.logWarn("video poll failed; retrying later", err, mapping.PublicID)
		h.patchVideo(ctx, mapping, videoStatusReconciling, retryVideoStatus(mapping), nil)

		return
	}

	if !state.terminal() {
		h.patchVideo(ctx, mapping, videoStatusReconciling, state.Status, map[string]interface{}{videoMetaProgress: state.Progress})

		return
	}

	h.settleVideo(ctx, mapping, state)
}

func (h *VideosHandler) claimVideo(ctx context.Context, mapping *platform.ProviderNativeResourceMapping, from []string, to string) bool {
	claimed, err := h.resources.ClaimProviderNativeResourceStatus(ctx, mapping.TenantID, mapping.Provider, videoResourceKind, mapping.PublicID, from, to)
	if err != nil {
		h.logWarn("video claim failed", err, mapping.PublicID)

		return false
	}

	return claimed
}

func retryVideoStatus(mapping *platform.ProviderNativeResourceMapping) string {
	if isPendingVideoStatus(mapping.Status) {
		return mapping.Status
	}

	return videoStatusQueued
}

func (h *VideosHandler) currentVideoState(ctx context.Context, mapping *platform.ProviderNativeResourceMapping) (videoJobState, error) {
	if created := intValue(mapping.Metadata[videoMetaCreatedAt]); created > 0 &&
		h.now().Sub(time.Unix(int64(created), 0)) > h.maxJobAge {
		return videoJobState{Status: videoStatusFailed, ErrorCode: "upstream_timeout",
			ErrorMessage: "The provider did not finish the video in time.", ChargeEstimate: true}, nil
	}

	deployment, codec, err := h.mappingDeployment(ctx, mapping)
	if err != nil {
		return videoJobState{}, err
	}

	poll, err := h.pollVideo(ctx, mapping, deployment)
	if errors.Is(err, errVideoUpstreamGone) {
		return videoJobState{Status: videoStatusFailed, ErrorCode: "upstream_expired",
			ErrorMessage: "The provider no longer has this video job.", ChargeEstimate: true}, nil
	}

	if err != nil {
		return videoJobState{}, err
	}

	return codec.parseJob(poll)
}

// settleVideo records the settlement intent before charging, so a crash
// between the charge and the final status can never charge twice.
func (h *VideosHandler) settleVideo(ctx context.Context, mapping *platform.ProviderNativeResourceMapping, state videoJobState) {
	actual, charge, note := h.videoSettlementCost(ctx, mapping, state)

	patch := map[string]interface{}{
		videoMetaProgress: state.Progress, videoMetaSettlement: note,
		videoMetaErrorCode: state.ErrorCode, videoMetaErrorMessage: state.ErrorMessage,
	}
	if state.Seconds > 0 {
		patch[videoMetaSeconds] = state.Seconds
	}

	if charge {
		patch[videoMetaSettledCostUSD] = actual.TotalCostUSD
	}

	if !h.patchVideo(ctx, mapping, videoStatusReconciling, videoStatusSettling, patch) {
		return
	}

	if charge {
		h.chargeVideo(ctx, mapping, actual, state.Status)
	} else {
		h.pipeline.Release(ctx, mapping.Reservation, "video_"+state.ErrorCode)
	}

	h.patchVideo(ctx, mapping, videoStatusSettling, state.Status, map[string]interface{}{videoMetaCompletedAt: h.now().Unix()})
}

// videoSettlementCost decides what one terminal job costs. The provider's own
// reported charge wins; catalog seconds are the fallback; failures the
// provider did not charge for are released.
func (h *VideosHandler) videoSettlementCost(
	ctx context.Context,
	mapping *platform.ProviderNativeResourceMapping,
	state videoJobState,
) (platform.UsageCostBreakdown, bool, string) {
	estimated := platform.UsageCostBreakdown{}
	if mapping.EstimatedUsage != nil {
		estimated = *mapping.EstimatedUsage
	}

	if state.ChargeEstimate {
		h.logWarn("video settled at its estimate", errors.New(state.ErrorCode), mapping.PublicID)

		return estimated, true, "estimate_charged"
	}

	reportedCharge := state.CostUSDTicks != nil && *state.CostUSDTicks > 0
	if state.Status == videoStatusFailed && !reportedCharge {
		return platform.UsageCostBreakdown{}, false, "released"
	}

	seconds := state.Seconds
	if mapping.Metadata[videoMetaOperation] == platform.VideoOperationExtend || seconds <= 0 {
		// An extension bills its new segment, never the stitched total.
		seconds = intValue(mapping.Metadata[videoMetaSeconds])
	}

	actual, err := h.pipeline.priceVideoUsage(ctx, estimated.Model, seconds, state.CostUSDTicks, estimated)
	if err != nil {
		h.logWarn("video pricing failed; settled at its estimate", err, mapping.PublicID)

		return estimated, true, "estimate_charged"
	}

	if !reportedCharge {
		h.logWarn("provider omitted the video charge; settled from catalog seconds", errors.New("missing cost_in_usd_ticks"), mapping.PublicID)

		return actual, true, "estimated_settlement"
	}

	if drift := actual.TotalCostUSD - estimated.TotalCostUSD; drift > 0 && h.logger != nil {
		h.logger.Warn("video charge exceeded its reservation", zap.String("video_id", mapping.PublicID), zap.Float64("drift_usd", drift))
	}

	return actual, true, "provider_reported_charge"
}

func (h *VideosHandler) chargeVideo(ctx context.Context, mapping *platform.ProviderNativeResourceMapping, actual platform.UsageCostBreakdown, status string) {
	billingMode := stringValue(mapping.Metadata[videoMetaBillingMode])
	if err := h.pipeline.Settle(ctx, SettlementRequest{
		Reservation: mapping.Reservation, TenantID: mapping.TenantID, Actual: actual,
		RequestID: mapping.ReservationID, BillingMode: billingMode,
	}); err != nil {
		h.logWarn("video budget settlement failed", err, mapping.PublicID)
	}

	requestMetadata, ok := mapping.Metadata[videoMetaRequestMeta].(map[string]interface{})
	if !ok || requestMetadata == nil {
		requestMetadata = map[string]interface{}{"billing_mode": billingMode}
	}

	requestMetadata["video_id"], requestMetadata["video_status"] = mapping.PublicID, status
	actual.Endpoint, actual.RequestType = videoRoute, videoRequestType
	createdAt := time.Unix(int64(intValue(mapping.Metadata[videoMetaCreatedAt])), 0).UTC()
	model := stringValue(mapping.Metadata[videoMetaModel])

	logSpendBestEffort(h.logger, h.logs, ctx, platform.SpendLog{
		RequestID: mapping.ReservationID, TokenHash: stringValue(mapping.Metadata[videoMetaTokenHash]),
		TenantID: mapping.TenantID, TeamID: stringValue(mapping.Metadata[videoMetaTeamID]),
		UserID: stringValue(mapping.Metadata[videoMetaUserID]), Model: model,
		CostUSD: actual.TotalCostUSD, CostIRR: actual.TotalCostIRR, Status: "success",
		ResponseTimeMS: int(h.now().Sub(createdAt).Milliseconds()), RequestMetadata: requestMetadata,
		UsageBreakdown: &actual, PricingSnapshot: actual.PricingSnapshot,
		CompletionStart: createdAt, CompletionEnd: h.now().UTC(),
	})

	if h.emitMetric != nil {
		h.emitMetric(metricsobs.BuildInput{
			TS: h.now().UTC(), TenantID: mapping.TenantID, RequestID: mapping.ReservationID,
			ModelName: model, ProviderName: mapping.Provider, RequestType: videoRequestType,
			UserID: stringValue(mapping.Metadata[videoMetaUserID]), StatusCode: fiber.StatusOK,
			CostUSD: actual.TotalCostUSD,
			Metadata: map[string]interface{}{"endpoint": videoRoute, "video_status": status,
				"settlement": stringValue(mapping.Metadata[videoMetaSettlement])},
		})
	}
}

// abandonInterruptedSettlement finishes a job whose settlement was cut off.
// Whether the charge landed is unknown, so it is never applied twice.
func (h *VideosHandler) abandonInterruptedSettlement(ctx context.Context, mapping *platform.ProviderNativeResourceMapping) {
	if h.logger != nil {
		h.logger.Error("video settlement was interrupted; charge not re-applied", zap.String("video_id", mapping.PublicID))
	}

	final := videoStatusCompleted
	if stringValue(mapping.Metadata[videoMetaErrorCode]) != "" {
		final = videoStatusFailed
	}

	h.patchVideo(ctx, mapping, videoStatusSettling, final, map[string]interface{}{
		videoMetaCompletedAt: h.now().Unix(), videoMetaSettlement: "interrupted",
	})
}

func (h *VideosHandler) patchVideo(ctx context.Context, mapping *platform.ProviderNativeResourceMapping, expected, status string, patch map[string]interface{}) bool {
	updated, err := h.resources.PatchProviderNativeResourceState(ctx, mapping.TenantID, mapping.Provider, videoResourceKind, mapping.PublicID, expected, status, patch)
	if err != nil || !updated {
		if err == nil {
			err = fmt.Errorf("video %s is no longer %s", mapping.PublicID, expected)
		}

		h.logWarn("video state update failed", err, mapping.PublicID)

		return false
	}

	mapping.Status = status
	for key, value := range patch {
		mapping.Metadata[key] = value
	}

	return true
}

// renderOpenAIVideo renders the OpenAI Videos `video` object. Prompts are not
// stored, so `prompt` is always null.
func renderOpenAIVideo(mapping *platform.ProviderNativeResourceMapping) map[string]interface{} {
	status := mapping.Status
	if status == videoStatusReconciling || status == videoStatusSettling {
		status = videoStatusInProgress
	}

	video := map[string]interface{}{
		"id": mapping.PublicID, "object": "video", "model": stringValue(mapping.Metadata[videoMetaModel]),
		"status": status, "progress": intValue(mapping.Metadata[videoMetaProgress]),
		"created_at": intValue(mapping.Metadata[videoMetaCreatedAt]), "completed_at": nil, "expires_at": nil,
		"prompt": nil, "size": nil, "seconds": strconv.Itoa(intValue(mapping.Metadata[videoMetaSeconds])),
		"remixed_from_video_id": nil, "error": nil,
	}

	if size := stringValue(mapping.Metadata[videoMetaSize]); size != "" {
		video["size"] = size
	}

	if completedAt := intValue(mapping.Metadata[videoMetaCompletedAt]); completedAt > 0 {
		video["completed_at"] = completedAt
	}

	if status == videoStatusFailed {
		video["error"] = map[string]interface{}{
			"code": stringValue(mapping.Metadata[videoMetaErrorCode]), "message": stringValue(mapping.Metadata[videoMetaErrorMessage]),
		}
	}

	return video
}

// renderXAIVideo renders the xAI poll shape. The upstream temporary URL is
// replaced with GodarAI's authenticated content route.
func renderXAIVideo(c *fiber.Ctx, mapping *platform.ProviderNativeResourceMapping) map[string]interface{} {
	result := map[string]interface{}{"model": stringValue(mapping.Metadata[videoMetaModel])}

	switch mapping.Status {
	case videoStatusCompleted:
		result["status"], result["progress"] = "done", 100
		result["video"] = map[string]interface{}{
			"url":      strings.TrimSuffix(c.BaseURL(), "/") + videoRoute + "/" + mapping.PublicID + "/content",
			"duration": intValue(mapping.Metadata[videoMetaSeconds]), "respect_moderation": true,
		}
	case videoStatusFailed:
		result["status"] = "failed"
		result["error"] = map[string]interface{}{
			"code": stringValue(mapping.Metadata[videoMetaErrorCode]), "message": stringValue(mapping.Metadata[videoMetaErrorMessage]),
		}
	default:
		result["status"], result["progress"] = "pending", intValue(mapping.Metadata[videoMetaProgress])
	}

	return result
}

func renderVideo(c *fiber.Ctx, mapping *platform.ProviderNativeResourceMapping) map[string]interface{} {
	if stringValue(mapping.Metadata[videoMetaShape]) == videoShapeXAI {
		return renderXAIVideo(c, mapping)
	}

	return renderOpenAIVideo(mapping)
}

// intValue reads a JSON number stored in job metadata.
func intValue(value interface{}) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	default:
		return 0
	}
}

func headerValueFold(headers map[string]string, name string) string {
	for key, value := range headers {
		if strings.EqualFold(key, name) {
			return strings.TrimSpace(value)
		}
	}

	return ""
}
