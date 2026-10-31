package proxy

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

// Canonical video job states, matching the OpenAI Videos `status` enum.
const (
	videoStatusQueued     = "queued"
	videoStatusInProgress = "in_progress"
	videoStatusCompleted  = "completed"
	videoStatusFailed     = "failed"
)

// videoJobState is the canonical view of one upstream job poll.
type videoJobState struct {
	Status       string
	Progress     int
	Seconds      int
	CostUSDTicks *int64
	ErrorCode    string
	ErrorMessage string
	// ChargeEstimate settles at the reservation when the provider's own
	// outcome can no longer be read (timeout or expired job).
	ChargeEstimate bool
}

func (s videoJobState) terminal() bool {
	return s.Status == videoStatusCompleted || s.Status == videoStatusFailed
}

// videoProviderCodec maps the canonical video contract onto one provider's
// wire format. HTTP stays behind the provider egress; a codec only shapes
// bodies and parses responses, so adding a provider never touches the routes,
// the job store, or the settlement worker.
type videoProviderCodec interface {
	createOperation(operation string) platform.ProviderOperation
	pollOperation() platform.ProviderOperation
	contentOperation() platform.ProviderOperation
	// buildCreateBody translates a canonical request into the provider body.
	buildCreateBody(request canonicalVideoRequest, profile platform.ModelRuntimeProfile) (map[string]interface{}, error)
	// validateBody checks the final provider body against the pinned profile.
	validateBody(operation string, body []byte, profile platform.ModelRuntimeProfile) (platform.VideoRequestFacts, error)
	parseCreateResponse(body []byte) (string, error)
	parseJob(body []byte) (videoJobState, error)
	// sourceReference turns a completed job poll into the provider's source
	// video reference for edits and extensions.
	sourceReference(poll []byte) (map[string]interface{}, error)
}

// canonicalVideoRequest is an OpenAI Videos-shaped request after parsing.
type canonicalVideoRequest struct {
	Operation  string
	Model      string
	Prompt     string
	Seconds    int
	Size       string
	Image      map[string]interface{}
	Source     map[string]interface{}
	Extensions map[string]json.RawMessage
}

type xaiVideoCodec struct{}

func (xaiVideoCodec) createOperation(operation string) platform.ProviderOperation {
	switch operation {
	case platform.VideoOperationEdit:
		return platform.ProviderOperationXAIVideosEdit
	case platform.VideoOperationExtend:
		return platform.ProviderOperationXAIVideosExtend
	default:
		return platform.ProviderOperationXAIVideosGenerate
	}
}

func (xaiVideoCodec) pollOperation() platform.ProviderOperation {
	return platform.ProviderOperationXAIVideosGet
}

func (xaiVideoCodec) contentOperation() platform.ProviderOperation {
	return platform.ProviderOperationXAIVideosContent
}

func (xaiVideoCodec) buildCreateBody(request canonicalVideoRequest, profile platform.ModelRuntimeProfile) (map[string]interface{}, error) {
	body := map[string]interface{}{"model": request.Model}
	if request.Prompt != "" {
		body["prompt"] = request.Prompt
	}

	if request.Seconds > 0 {
		body["duration"] = request.Seconds
	}

	if request.Size != "" {
		format, ok := profile.VideoControls.Sizes[request.Size]
		if !ok {
			return nil, fmt.Errorf("size %q is not supported by this video model", request.Size)
		}

		body["aspect_ratio"], body["resolution"] = format.AspectRatio, format.Resolution
	}

	if request.Image != nil {
		body["image"] = request.Image
	}

	if request.Source != nil {
		body["video"] = request.Source
	}

	// Extension fields already use xAI names; the profile allow-list is
	// enforced before translation, and the final body is validated again.
	for field, raw := range request.Extensions {
		if _, overlaps := body[field]; overlaps {
			return nil, fmt.Errorf("parameter %q conflicts with a standard video parameter", field)
		}

		body[field] = raw
	}

	return body, nil
}

func (xaiVideoCodec) validateBody(operation string, body []byte, profile platform.ModelRuntimeProfile) (platform.VideoRequestFacts, error) {
	facts, err := platform.ParseXAIVideoRequest(operation, body)
	if err != nil {
		return facts, err
	}

	return facts, profile.ValidateVideoRequest(facts)
}

func (xaiVideoCodec) parseCreateResponse(body []byte) (string, error) {
	var response struct {
		RequestID string `json:"request_id"`
	}

	if err := json.Unmarshal(body, &response); err != nil || strings.TrimSpace(response.RequestID) == "" {
		return "", fmt.Errorf("xAI video response omitted request_id")
	}

	return strings.TrimSpace(response.RequestID), nil
}

type xaiVideoPoll struct {
	Status   string `json:"status"`
	Progress *int   `json:"progress"`
	Video    *struct {
		URL               string `json:"url"`
		Duration          int    `json:"duration"`
		RespectModeration *bool  `json:"respect_moderation"`
	} `json:"video"`
	Usage *struct {
		CostInUSDTicks *int64 `json:"cost_in_usd_ticks"`
	} `json:"usage"`
	Error *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func (xaiVideoCodec) parseJob(body []byte) (videoJobState, error) {
	var poll xaiVideoPoll
	if err := json.Unmarshal(body, &poll); err != nil {
		return videoJobState{}, fmt.Errorf("xAI video status is unreadable")
	}

	state := videoJobState{}
	if poll.Progress != nil {
		state.Progress = *poll.Progress
	}

	if poll.Usage != nil {
		state.CostUSDTicks = poll.Usage.CostInUSDTicks
	}

	if poll.Video != nil {
		state.Seconds = poll.Video.Duration
	}

	switch poll.Status {
	case "pending":
		state.Status = videoStatusQueued
		if state.Progress > 0 {
			state.Status = videoStatusInProgress
		}
	case "done":
		state.Status, state.Progress = videoStatusCompleted, 100
		if poll.Video == nil || (poll.Video.RespectModeration != nil && !*poll.Video.RespectModeration) {
			state.Status = videoStatusFailed
			state.ErrorCode, state.ErrorMessage = "content_policy_violation", "The generated video was blocked by provider moderation."
		}
	case "failed":
		state.Status, state.ErrorCode, state.ErrorMessage = videoStatusFailed, "upstream_failed", "The provider failed to generate the video."
		if poll.Error != nil {
			state.ErrorCode = xaiVideoErrorCode(poll.Error.Code)
			state.ErrorMessage = strings.TrimSpace(poll.Error.Message)
		}
	case "expired":
		state.Status, state.ErrorCode, state.ErrorMessage = videoStatusFailed, "upstream_expired", "The provider expired the video job."
	default:
		state.Status, state.ErrorCode, state.ErrorMessage = videoStatusFailed, "upstream_unknown_status", "The provider returned an unknown job status."
	}

	return state, nil
}

func xaiVideoErrorCode(code string) string {
	switch code {
	case "invalid_argument", "permission_denied", "failed_precondition":
		return "invalid_request"
	case "service_unavailable":
		return "provider_unavailable"
	default:
		return "upstream_failed"
	}
}

func (xaiVideoCodec) sourceReference(poll []byte) (map[string]interface{}, error) {
	var response xaiVideoPoll
	if err := json.Unmarshal(poll, &response); err != nil || response.Status != "done" ||
		response.Video == nil || strings.TrimSpace(response.Video.URL) == "" {
		return nil, fmt.Errorf("the source video is no longer available upstream")
	}

	return map[string]interface{}{"url": strings.TrimSpace(response.Video.URL)}, nil
}

// videoReservationSeconds is the duration a request is reserved for: the
// requested value, else the profile default, else the longest edit source.
func videoReservationSeconds(facts platform.VideoRequestFacts, controls platform.ModelRuntimeVideoControls) int {
	if facts.DurationSeconds > 0 {
		return facts.DurationSeconds
	}

	switch facts.Operation {
	case platform.VideoOperationExtend:
		return controls.DefaultExtensionSeconds
	case platform.VideoOperationEdit:
		return int(math.Ceil(controls.MaxEditSourceSeconds))
	default:
		return controls.DefaultDurationSeconds
	}
}
